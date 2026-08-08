import 'dart:io';

import 'package:audio_metadata_reader/audio_metadata_reader.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

import 'package:inori_music/src/local_library/local_library_db.dart';
import 'package:inori_music/src/shared/safe_file_name.dart';

const _kLocalLibrarySortKey = 'localLibrary.sortOrder';

final localLibrarySortProvider =
    NotifierProvider<LocalLibrarySortNotifier, LocalLibrarySortOrder>(
      LocalLibrarySortNotifier.new,
    );

/// Persists the flat-list sort dimension shown in Settings' local library
/// sort sheet (see local_library_screen.dart). [LocalLibraryNotifier] watches
/// this so changing it automatically re-queries in the new order.
class LocalLibrarySortNotifier extends Notifier<LocalLibrarySortOrder> {
  @override
  LocalLibrarySortOrder build() {
    _restore();
    return LocalLibrarySortOrder.artistAlbumTitle;
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_kLocalLibrarySortKey);
    state = LocalLibrarySortOrder.values.firstWhere(
      (s) => s.name == raw,
      orElse: () => LocalLibrarySortOrder.artistAlbumTitle,
    );
  }

  Future<void> setSort(LocalLibrarySortOrder sort) async {
    state = sort;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kLocalLibrarySortKey, sort.name);
  }
}

/// Prefix marking a track id as locally-imported (guest mode / no account).
/// [PlayerNotifier] branches on this prefix at the handful of trackId-keyed
/// lookups that would otherwise hit the server catalog.
const localTrackIdPrefix = 'local:';

/// Audio container extensions [audio_metadata_reader] (and just_audio) can
/// handle. Lowercase, no leading dot.
const supportedLocalAudioExtensions = [
  'mp3',
  'flac',
  'm4a',
  'mp4',
  'ogg',
  'oga',
  'opus',
  'wav',
  'aiff',
  'aif',
  'ape',
];

final localLibraryProvider =
    AsyncNotifierProvider<LocalLibraryNotifier, List<LocalLibraryTrack>>(
      LocalLibraryNotifier.new,
    );

/// Result of an import attempt, so the UI can tell the three outcomes apart:
/// the user backed out of the picker, some files landed, or the whole thing
/// blew up. Previously every one of these looked identical on screen —
/// nothing happened — which is how a completely broken Windows import went
/// unnoticed.
class ImportOutcome {
  const ImportOutcome({
    required this.imported,
    required this.failed,
    this.firstError,
  }) : cancelled = false;

  const ImportOutcome.cancelled()
    : imported = 0,
      failed = 0,
      firstError = null,
      cancelled = true;

  ImportOutcome.failed(Object error)
    : imported = 0,
      failed = 0,
      firstError = error,
      cancelled = false;

  final int imported;
  final int failed;

  /// First per-file error, or the fatal one when the whole attempt failed.
  final Object? firstError;

  /// The user dismissed the picker — not an error, and not worth reporting.
  final bool cancelled;

  bool get hasError => firstError != null;
}

/// Owns the "guest local library": importing files/folders, extracting
/// embedded metadata, and persisting to [LocalLibraryDb]. Pure client-side,
/// no server/account involvement — this is what makes guest mode usable.
class LocalLibraryNotifier extends AsyncNotifier<List<LocalLibraryTrack>> {
  @override
  Future<List<LocalLibraryTrack>> build() {
    // Watched (not read) so changing the sort preference elsewhere
    // automatically re-triggers this build and re-queries in the new order.
    final sort = ref.watch(localLibrarySortProvider);
    return LocalLibraryDb.instance.queryAll(sort: sort);
  }

  /// Opens a multi-file picker filtered to supported audio extensions.
  Future<ImportOutcome> importFiles() async {
    final FilePickerResult? result;
    try {
      result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: supportedLocalAudioExtensions,
        allowMultiple: true,
      );
    } catch (e) {
      return ImportOutcome.failed(e);
    }
    final paths = result?.files.map((f) => f.path).whereType<String>().toList();
    if (paths == null || paths.isEmpty) return const ImportOutcome.cancelled();
    return _importPaths(paths);
  }

  /// Opens a directory picker and recursively imports every supported audio
  /// file found underneath it.
  Future<ImportOutcome> importFolder() async {
    final paths = <String>[];
    try {
      final dirPath = await FilePicker.platform.getDirectoryPath();
      if (dirPath == null) return const ImportOutcome.cancelled();
      await for (final entity in Directory(
        dirPath,
      ).list(recursive: true, followLinks: false)) {
        if (entity is! File) continue;
        final ext = p
            .extension(entity.path)
            .replaceFirst('.', '')
            .toLowerCase();
        if (supportedLocalAudioExtensions.contains(ext)) paths.add(entity.path);
      }
    } catch (e) {
      return ImportOutcome.failed(e);
    }
    return _importPaths(paths);
  }

  Future<ImportOutcome> _importPaths(List<String> paths) async {
    final Directory audioDir;
    final Directory coverDir;
    try {
      audioDir = await _localLibraryDir('local_library_audio');
      coverDir = await _localLibraryDir('local_library_covers');
    } catch (e) {
      return ImportOutcome.failed(e);
    }

    var imported = 0;
    Object? firstError;
    for (final path in paths) {
      try {
        await _importOne(path, audioDir, coverDir);
        imported++;
      } catch (e) {
        // One unreadable/corrupt file must not abort the rest of the batch —
        // but the first failure is kept so the caller can say *why* nothing
        // showed up, instead of the import appearing to do nothing at all.
        debugPrint('LocalLibrary: failed to import $path: $e');
        firstError ??= e;
      }
    }

    try {
      await _refresh();
    } catch (e) {
      return ImportOutcome.failed(e);
    }
    return ImportOutcome(
      imported: imported,
      failed: paths.length - imported,
      firstError: firstError,
    );
  }

  Future<void> _importOne(
    String path,
    Directory audioDir,
    Directory coverDir,
  ) async {
    final sourceFile = File(path);
    if (!sourceFile.existsSync()) return;
    // Read tags from the original picked file first (that access is only
    // guaranteed to be valid for the duration of this import call — see the
    // copy step below for why playback can't rely on the original path).
    final meta = readMetadata(sourceFile, getImage: true);
    final id = '$localTrackIdPrefix${const Uuid().v4()}';
    // The id itself keeps its `local:` prefix (PlayerNotifier branches on it),
    // but the colon can't go into a file name — see [safeFileName].
    final stem = safeFileName(id);

    // Copy into the app's own storage rather than keeping the picked path.
    // On macOS, file_picker's NSOpenPanel grants sandbox read access that
    // isn't guaranteed to outlive this import call — a later playback
    // attempt reopening the *original* path can silently fail (metadata
    // reads fine here, but the track never actually plays). Owning a copy
    // also survives the user later moving/renaming/deleting the source file.
    final copiedFile = await sourceFile.copy(
      p.join(audioDir.path, '$stem${p.extension(path)}'),
    );

    String? coverPath;
    if (meta.pictures.isNotEmpty) {
      final pic = meta.pictures.first;
      final ext = pic.mimetype.split('/').last;
      final coverFile = File(p.join(coverDir.path, '$stem.$ext'));
      await coverFile.writeAsBytes(pic.bytes);
      coverPath = coverFile.path;
    }

    final title = meta.title?.trim();
    final genre = meta.genres.isNotEmpty ? meta.genres.first : null;
    final ext = p.extension(path).replaceFirst('.', '');
    await LocalLibraryDb.instance.insert(
      LocalLibraryTrack(
        id: id,
        title: (title == null || title.isEmpty)
            ? p.basenameWithoutExtension(path)
            : title,
        artistName: meta.artist?.trim() ?? '',
        albumTitle: meta.album?.trim() ?? '',
        localPath: copiedFile.path,
        durationMs: meta.duration?.inMilliseconds,
        coverArtPath: coverPath,
        sizeBytes: await copiedFile.length(),
        importedAt: DateTime.now(),
        // These were already sitting on `meta` before v5.19.0 — readMetadata()
        // parses them regardless, they just weren't being kept.
        sampleRate: meta.sampleRate,
        bitrate: meta.bitrate,
        fileFormat: ext.isEmpty ? null : ext.toUpperCase(),
        genre: (genre != null && genre.trim().isNotEmpty) ? genre.trim() : null,
        trackNumber: meta.trackNumber,
        embeddedLyrics: (meta.lyrics != null && meta.lyrics!.trim().isNotEmpty)
            ? meta.lyrics
            : null,
      ),
    );
  }

  Future<Directory> _localLibraryDir(String subdir) async {
    final dir = await getApplicationSupportDirectory();
    final target = Directory(p.join(dir.path, subdir));
    if (!target.existsSync()) target.createSync(recursive: true);
    return target;
  }

  Future<void> remove(String id) async {
    await _deleteOne(id);
    await _refresh();
  }

  /// Batch counterpart to [remove], for the list's multi-select mode. Deletes
  /// every id before re-querying once, rather than re-reading the whole table
  /// per track. One unreadable file must not abort the rest of the batch —
  /// same rule the import path follows.
  Future<void> removeAll(Iterable<String> ids) async {
    for (final id in ids) {
      try {
        await _deleteOne(id);
      } catch (e) {
        debugPrint('LocalLibrary: failed to remove $id: $e');
      }
    }
    await _refresh();
  }

  Future<void> _deleteOne(String id) async {
    final track = await LocalLibraryDb.instance.query(id);
    await LocalLibraryDb.instance.delete(id);
    if (track == null) return;
    final audioFile = File(track.localPath);
    if (audioFile.existsSync()) await audioFile.delete();
    final coverPath = track.coverArtPath;
    if (coverPath != null) {
      final coverFile = File(coverPath);
      if (coverFile.existsSync()) await coverFile.delete();
    }
  }

  Future<void> _refresh() async {
    state = AsyncData(
      await LocalLibraryDb.instance.queryAll(
        sort: ref.read(localLibrarySortProvider),
      ),
    );
  }
}
