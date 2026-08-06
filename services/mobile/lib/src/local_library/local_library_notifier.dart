import 'dart:io';

import 'package:audio_metadata_reader/audio_metadata_reader.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:uuid/uuid.dart';

import 'package:inori_music/src/local_library/local_library_db.dart';

/// Prefix marking a track id as locally-imported (guest mode / no account).
/// [PlayerNotifier] branches on this prefix at the handful of trackId-keyed
/// lookups that would otherwise hit the server catalog.
const localTrackIdPrefix = 'local:';

/// Audio container extensions [audio_metadata_reader] (and just_audio) can
/// handle. Lowercase, no leading dot.
const supportedLocalAudioExtensions = [
  'mp3', 'flac', 'm4a', 'mp4', 'ogg', 'oga', 'opus', 'wav', 'aiff', 'aif', 'ape',
];

final localLibraryProvider = AsyncNotifierProvider<LocalLibraryNotifier, List<LocalLibraryTrack>>(
  LocalLibraryNotifier.new,
);

/// Owns the "guest local library": importing files/folders, extracting
/// embedded metadata, and persisting to [LocalLibraryDb]. Pure client-side,
/// no server/account involvement — this is what makes guest mode usable.
class LocalLibraryNotifier extends AsyncNotifier<List<LocalLibraryTrack>> {
  @override
  Future<List<LocalLibraryTrack>> build() => LocalLibraryDb.instance.queryAll();

  /// Opens a multi-file picker filtered to supported audio extensions.
  Future<void> importFiles() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: supportedLocalAudioExtensions,
      allowMultiple: true,
    );
    final paths = result?.files.map((f) => f.path).whereType<String>().toList();
    if (paths == null || paths.isEmpty) return;
    await _importPaths(paths);
  }

  /// Opens a directory picker and recursively imports every supported audio
  /// file found underneath it.
  Future<void> importFolder() async {
    final dirPath = await FilePicker.platform.getDirectoryPath();
    if (dirPath == null) return;
    final paths = <String>[];
    await for (final entity in Directory(dirPath).list(recursive: true, followLinks: false)) {
      if (entity is! File) continue;
      final ext = p.extension(entity.path).replaceFirst('.', '').toLowerCase();
      if (supportedLocalAudioExtensions.contains(ext)) paths.add(entity.path);
    }
    await _importPaths(paths);
  }

  Future<void> _importPaths(List<String> paths) async {
    final audioDir = await _localLibraryDir('local_library_audio');
    final coverDir = await _localLibraryDir('local_library_covers');
    for (final path in paths) {
      try {
        await _importOne(path, audioDir, coverDir);
      } catch (e) {
        // One unreadable/corrupt file must not abort the rest of the batch.
        debugPrint('LocalLibrary: failed to import $path: $e');
      }
    }
    state = AsyncData(await LocalLibraryDb.instance.queryAll());
  }

  Future<void> _importOne(String path, Directory audioDir, Directory coverDir) async {
    final sourceFile = File(path);
    if (!sourceFile.existsSync()) return;
    // Read tags from the original picked file first (that access is only
    // guaranteed to be valid for the duration of this import call — see the
    // copy step below for why playback can't rely on the original path).
    final meta = readMetadata(sourceFile, getImage: true);
    final id = '$localTrackIdPrefix${const Uuid().v4()}';

    // Copy into the app's own storage rather than keeping the picked path.
    // On macOS, file_picker's NSOpenPanel grants sandbox read access that
    // isn't guaranteed to outlive this import call — a later playback
    // attempt reopening the *original* path can silently fail (metadata
    // reads fine here, but the track never actually plays). Owning a copy
    // also survives the user later moving/renaming/deleting the source file.
    final copiedFile = await sourceFile.copy(p.join(audioDir.path, '$id${p.extension(path)}'));

    String? coverPath;
    if (meta.pictures.isNotEmpty) {
      final pic = meta.pictures.first;
      final ext = pic.mimetype.split('/').last;
      final coverFile = File(p.join(coverDir.path, '$id.$ext'));
      await coverFile.writeAsBytes(pic.bytes);
      coverPath = coverFile.path;
    }

    final title = meta.title?.trim();
    await LocalLibraryDb.instance.insert(LocalLibraryTrack(
      id: id,
      title: (title == null || title.isEmpty) ? p.basenameWithoutExtension(path) : title,
      artistName: meta.artist?.trim() ?? '',
      albumTitle: meta.album?.trim() ?? '',
      localPath: copiedFile.path,
      durationMs: meta.duration?.inMilliseconds,
      coverArtPath: coverPath,
      sizeBytes: await copiedFile.length(),
      importedAt: DateTime.now(),
    ));
  }

  Future<Directory> _localLibraryDir(String subdir) async {
    final dir = await getApplicationSupportDirectory();
    final target = Directory(p.join(dir.path, subdir));
    if (!target.existsSync()) target.createSync(recursive: true);
    return target;
  }

  Future<void> remove(String id) async {
    final track = await LocalLibraryDb.instance.query(id);
    await LocalLibraryDb.instance.delete(id);
    if (track != null) {
      final audioFile = File(track.localPath);
      if (audioFile.existsSync()) await audioFile.delete();
      final coverPath = track.coverArtPath;
      if (coverPath != null) {
        final coverFile = File(coverPath);
        if (coverFile.existsSync()) await coverFile.delete();
      }
    }
    state = AsyncData(await LocalLibraryDb.instance.queryAll());
  }
}
