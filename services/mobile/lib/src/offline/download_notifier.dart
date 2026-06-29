import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as p;

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/offline/offline_db.dart';

// DownloadStatus sealed hierarchy
sealed class DownloadStatus {
  const DownloadStatus();
}

class DownloadIdle extends DownloadStatus {
  const DownloadIdle();
}

class DownloadInProgress extends DownloadStatus {
  const DownloadInProgress(this.progress);
  final double progress; // 0.0 - 1.0
}

class DownloadDone extends DownloadStatus {
  const DownloadDone();
}

class DownloadError extends DownloadStatus {
  const DownloadError(this.message);
  final String message;
}

// State: Map<trackId, DownloadStatus>
class DownloadNotifier extends Notifier<Map<String, DownloadStatus>> {
  bool _disposed = false;

  @override
  Map<String, DownloadStatus> build() {
    ref.onDispose(() => _disposed = true);
    // Restore already-downloaded tracks from DB on startup.
    // Fire-and-forget with an explicit guard so errors are not silently lost
    // and a state write on a disposed notifier does not throw.
    _restoreFromDb();
    return {};
  }

  Future<void> _restoreFromDb() async {
    try {
      final tracks = await OfflineDb.instance.queryAll();
      final restored = <String, DownloadStatus>{};
      for (final t in tracks) {
        if (File(t.localPath).existsSync()) {
          restored[t.trackId] = const DownloadDone();
        }
      }
      // Guard against disposal between the await and the state write.
      if (restored.isNotEmpty && !_disposed) {
        state = {...state, ...restored};
      }
    } catch (e) {
      // Non-fatal: start with empty state. The user can re-download.
      // ignore: avoid_print
      print('[DownloadNotifier] _restoreFromDb failed: $e');
    }
  }

  Future<void> startDownload(String trackId) async {
    if (state[trackId] is DownloadInProgress) return;
    state = {...state, trackId: const DownloadInProgress(0)};

    // localPath is declared here so the catch block can clean it up.
    String? localPath;

    try {
      // 1. Resolve playback URL via catalog repository.
      final catalog = ref.read(catalogRepositoryProvider);
      final descriptor = await catalog.getPlaybackDescriptor(trackId);
      final url = (descriptor.presignedUrl != null && descriptor.presignedUrl!.isNotEmpty)
          ? descriptor.presignedUrl!
          : descriptor.streamUrl;
      if (url == null || url.isEmpty) throw Exception('no playback url for $trackId');

      // 2. Determine local file path.
      final dir = await getApplicationDocumentsDirectory();
      localPath = p.join(dir.path, 'offline', '$trackId.audio');
      await Directory(p.dirname(localPath)).create(recursive: true);

      // 3. Stream download with progress.
      // Use a bare Dio instance (no auth interceptors) so that presigned
      // S3/MinIO URLs are not contaminated with an Authorization header.
      // Sending both query-string credentials and an Authorization header
      // causes SignatureDoesNotMatch (403) on every presigned request.
      final bareDio = Dio(BaseOptions(
        connectTimeout: const Duration(seconds: 15),
        receiveTimeout: const Duration(minutes: 30),
      ));
      await bareDio.download(
        url,
        localPath,
        onReceiveProgress: (received, total) {
          if (total > 0) {
            state = {...state, trackId: DownloadInProgress(received / total)};
          }
        },
      );

      // 4. Fetch track metadata and resolve display names for DB record.
      final track = await catalog.getTrack(trackId);
      final sizeBytes = await File(localPath).length();

      // Resolve artist display name; fall back to artistId on error.
      String artistName = track.artistId;
      try {
        final artist = await catalog.getArtist(track.artistId);
        artistName = artist.name;
      } catch (_) {/* keep UUID fallback */}

      // Resolve album title; fall back to albumId on error.
      String albumTitle = track.albumId ?? '';
      if (track.albumId != null && track.albumId!.isNotEmpty) {
        try {
          final album = await catalog.getAlbum(track.albumId!);
          albumTitle = album.title;
        } catch (_) {/* keep UUID fallback */}
      }

      // 5. Persist to OfflineDb.
      await OfflineDb.instance.insert(OfflineTrack(
        trackId: trackId,
        title: track.title,
        artistName: artistName,
        albumTitle: albumTitle,
        albumId: track.albumId,
        localPath: localPath,
        sizeBytes: sizeBytes,
        downloadedAt: DateTime.now(),
      ));

      state = {...state, trackId: const DownloadDone()};
    } catch (e) {
      // Clean up any partial or orphaned file before marking as error.
      if (localPath != null) {
        try {
          final partial = File(localPath);
          if (await partial.exists()) await partial.delete();
        } catch (_) {/* ignore cleanup failure */}
      }
      state = {...state, trackId: DownloadError(e.toString())};
    }
  }

  Future<void> deleteDownload(String trackId) async {
    final offlineTrack = await OfflineDb.instance.query(trackId);
    if (offlineTrack != null) {
      final file = File(offlineTrack.localPath);
      if (file.existsSync()) await file.delete();
      await OfflineDb.instance.delete(trackId);
    }
    final updated = Map<String, DownloadStatus>.from(state)..remove(trackId);
    state = updated;
  }

  Future<void> deleteAllDownloads() async {
    final all = await OfflineDb.instance.queryAll();
    for (final t in all) {
      final f = File(t.localPath);
      if (f.existsSync()) await f.delete();
    }
    await OfflineDb.instance.deleteAll();
    state = {};
  }

  bool isDownloaded(String trackId) => state[trackId] is DownloadDone;
  bool isDownloading(String trackId) => state[trackId] is DownloadInProgress;
}

final downloadProvider =
    NotifierProvider<DownloadNotifier, Map<String, DownloadStatus>>(
  DownloadNotifier.new,
);
