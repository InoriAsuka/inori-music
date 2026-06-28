import 'dart:io';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as p;

import 'package:inori_music/src/api/api_client.dart';
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
  @override
  Map<String, DownloadStatus> build() {
    // Restore already-downloaded tracks from DB on startup.
    _restoreFromDb();
    return {};
  }

  Future<void> _restoreFromDb() async {
    final tracks = await OfflineDb.instance.queryAll();
    final restored = <String, DownloadStatus>{};
    for (final t in tracks) {
      if (File(t.localPath).existsSync()) {
        restored[t.trackId] = const DownloadDone();
      }
    }
    if (restored.isNotEmpty) {
      state = {...state, ...restored};
    }
  }

  Future<void> startDownload(String trackId) async {
    if (state[trackId] is DownloadInProgress) return;
    state = {...state, trackId: const DownloadInProgress(0)};

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
      final localPath = p.join(dir.path, 'offline', '$trackId.audio');
      await Directory(p.dirname(localPath)).create(recursive: true);

      // 3. Stream download with progress.
      final dio = ref.read(dioProvider);
      await dio.download(
        url,
        localPath,
        onReceiveProgress: (received, total) {
          if (total > 0) {
            state = {...state, trackId: DownloadInProgress(received / total)};
          }
        },
      );

      // 4. Fetch track metadata for DB record.
      final track = await catalog.getTrack(trackId);
      final sizeBytes = await File(localPath).length();

      // 5. Persist to OfflineDb.
      await OfflineDb.instance.insert(OfflineTrack(
        trackId: trackId,
        title: track.title,
        artistName: track.artistId, // resolved display name via cache elsewhere
        albumTitle: track.albumId ?? '',
        albumId: track.albumId,
        localPath: localPath,
        sizeBytes: sizeBytes,
        downloadedAt: DateTime.now(),
      ));

      state = {...state, trackId: const DownloadDone()};
    } catch (e) {
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
