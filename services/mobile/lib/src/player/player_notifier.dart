// ignore_for_file: implementation_imports
import 'dart:async';
import 'dart:io';

import 'package:audio_service/audio_service.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_music/src/offline/offline_db.dart';
import 'package:inori_api/src/api/catalog_api.dart';
import 'package:inori_api/src/api/history_api.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:inori_api/src/model/record_play_event_request.dart';
import 'package:just_audio/just_audio.dart';

import 'package:inori_music/main.dart' show audioHandler;
import 'package:inori_music/src/api/api_client.dart';
import 'package:inori_music/src/catalog/catalog_cache_providers.dart';
import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

final playerProvider = NotifierProvider<PlayerNotifier, pstate.PlayerState>(
  PlayerNotifier.new,
);

final historyApiProvider = Provider<HistoryApi>((ref) {
  return HistoryApi(ref.read(dioProvider));
});

// ---------------------------------------------------------------------------
// Player notifier — owns the just_audio AudioPlayer + queue logic
// ---------------------------------------------------------------------------

class PlayerNotifier extends Notifier<pstate.PlayerState> {
  late final AudioPlayer _audioPlayer;
  late final CatalogApi _catalog;
  late final HistoryApi _history;

  // In-memory track metadata cache to avoid redundant catalog API calls.
  final Map<String, CatalogTrack> _trackCache = {};

  // Resolved display names — keyed by artistId / albumId.
  final Map<String, String> _artistNameCache = {};
  final Map<String, String> _albumTitleCache = {};

  // Store subscriptions so they can be cancelled on dispose.
  late final List<StreamSubscription> _subscriptions;

  @override
  pstate.PlayerState build() {
    // Use the AudioPlayer instance owned by the AudioHandler so that
    // audio_service (notifications, MediaSession, lock screen) stays in sync.
    _audioPlayer = audioHandler.audioPlayer;
    _catalog = ref.read(catalogApiProvider);
    _history = ref.read(historyApiProvider);
    _subscriptions = _setupPlayerListeners();
    // Cancel all stream subscriptions when the provider is disposed.
    ref.onDispose(() {
      for (final sub in _subscriptions) {
        sub.cancel();
      }
    });
    return pstate.PlayerState();
  }

  // ---- Public playback API ----

  /// Resolve the playback URL for a track and prepare the audio source.
  /// Returns the resolved URL or null if unavailable.
  Future<String?> resolvePlaybackUrl(String trackId) async {
    // Check local offline cache first.
    final offline = await OfflineDb.instance.query(trackId);
    if (offline != null && File(offline.localPath).existsSync()) {
      return 'file://${offline.localPath}';
    }
    try {
      final resp = await _catalog.getTrackPlaybackDescriptor(id: trackId);
      final descriptor = resp.data;
      if (descriptor == null) return null;

      if (descriptor.presignedUrl != null && descriptor.presignedUrl!.isNotEmpty) {
        return descriptor.presignedUrl;
      }
      if (descriptor.streamUrl != null && descriptor.streamUrl!.isNotEmpty) {
        final token = await ref.read(tokenProvider.future);
        if (token != null) {
          final uri = Uri.parse(descriptor.streamUrl!);
          return uri.replace(queryParameters: {...uri.queryParameters, 'token': token}).toString();
        }
        return descriptor.streamUrl;
      }
      // Fallback: viewer stream endpoint
      final token = await ref.read(tokenProvider.future);
      if (token != null) {
        return '/api/v1/catalog/tracks/$trackId/stream?token=$token';
      }
      return null;
    } catch (e) {
      debugPrint('PlayerNotifier: failed to resolve playback URL for $trackId: $e');
      return null;
    }
  }

  /// Play a single track by ID, optionally building the full queue.
  Future<void> playTrack(String trackId, {List<String>? queueIds, int index = 0}) async {
    final url = await resolvePlaybackUrl(trackId);
    if (url == null) return;

    // Build queue with stub items immediately so the UI has something to render,
    // then update the current item with real metadata once resolved.
    if (queueIds != null && queueIds.isNotEmpty) {
      final clampedIndex = index < 0 ? 0 : (index > queueIds.length - 1 ? queueIds.length - 1 : index);
      state = state.copyWith(
        queue: queueIds.map((id) => _stubMediaItem(id)).toList(),
        currentIndex: clampedIndex,
      );
      // Resolve all URLs for gapless playback via ConcatenatingAudioSource.
      _buildConcatQueue(queueIds, index, url, trackId);
    }

    // Resolve real track metadata (title, artist, album, duration) from catalog.
    final track = await _resolveTrack(trackId);
    final mediaItem = _makeMediaItem(trackId, track);

    // If no queue was supplied, fall back to single ProgressiveAudioSource.
    if (queueIds == null || queueIds.isEmpty) {
      final source = ProgressiveAudioSource(Uri.parse(url), tag: trackId);
      await _audioPlayer.setAudioSource(source);
    }
    // Push to AudioHandler so the OS notification shows the current track.
    audioHandler.mediaItem.add(mediaItem);
    state = state.copyWith(
      mediaItem: mediaItem,
      currentIndex: state.queue.isNotEmpty ? (state.currentIndex >= 0 ? state.currentIndex : index) : 0,
    );
    await _audioPlayer.play();
  }

  /// Asynchronously pre-resolves all queue URLs and calls [audioHandler.updateQueue]
  /// so ConcatenatingAudioSource enables gapless transitions.
  void _buildConcatQueue(List<String> queueIds, int startIndex, String resolvedUrl, String resolvedTrackId) {
    Future<void>(() async {
      final urls = <String>[];
      for (final id in queueIds) {
        if (id == resolvedTrackId) {
          urls.add(resolvedUrl);
        } else {
          final u = await resolvePlaybackUrl(id);
          urls.add(u ?? '');
        }
      }
      // Filter out empty URLs but rebuild index accordingly.
      final validPairs = <MapEntry<int, String>>[];
      for (var i = 0; i < urls.length; i++) {
        if (urls[i].isNotEmpty) validPairs.add(MapEntry(i, urls[i]));
      }
      if (validPairs.isEmpty) return;
      await audioHandler.updateConcatQueue(validPairs.map((e) => e.value).toList());
      // Seek to the correct position in the ConcatenatingAudioSource.
      final concatIndex = validPairs.indexWhere((e) => e.key == startIndex);
      if (concatIndex > 0) {
        await _audioPlayer.seek(Duration.zero, index: concatIndex);
      }
    });
  }

  /// Play a list of track IDs starting at [initialIndex].
  Future<void> playQueue(List<String> trackIds, {int initialIndex = 0}) async {
    if (trackIds.isEmpty) return;
    final idx = initialIndex < 0 ? 0 : (initialIndex > trackIds.length - 1 ? trackIds.length - 1 : initialIndex);
    state = state.copyWith(queue: trackIds.map((id) => _stubMediaItem(id)).toList());
    await playTrack(trackIds[idx], queueIds: trackIds, index: idx);
  }

  /// Enqueue tracks after the current position.
  Future<void> enqueue(List<String> trackIds) async {
    final items = trackIds.map((id) => _stubMediaItem(id)).toList();
    final newQueue = [...state.queue, ...items];
    state = state.copyWith(queue: newQueue);
    // Refresh gapless source with all current queue URLs.
    _refreshConcatSource(newQueue.map((m) => m.id).toList());
  }

  /// Enqueue a single track immediately after the current one.
  Future<void> enqueueNext(String trackId) async {
    final item = _stubMediaItem(trackId);
    final newQueue = [...state.queue];
    final insertAt = state.currentIndex + 1;
    if (insertAt < newQueue.length) {
      newQueue.insert(insertAt, item);
    } else {
      newQueue.add(item);
    }
    state = state.copyWith(queue: newQueue);
    _refreshConcatSource(newQueue.map((m) => m.id).toList());
  }

  /// Rebuild the ConcatenatingAudioSource to match the current queue.
  void _refreshConcatSource(List<String> trackIds) {
    Future<void>(() async {
      final urls = <String>[];
      for (final id in trackIds) {
        final u = await resolvePlaybackUrl(id);
        urls.add(u ?? '');
      }
      final valid = urls.where((u) => u.isNotEmpty).toList();
      if (valid.isNotEmpty) {
        await audioHandler.updateConcatQueue(valid);
      }
    });
  }

  Future<void> play() async => _audioPlayer.play();
  Future<void> pause() async => _audioPlayer.pause();

  Future<void> togglePlayPause() async {
    if (state.isPlaying) {
      await pause();
    } else if (!state.isIdle) {
      await play();
    }
  }

  Future<void> seekTo(Duration position) async => _audioPlayer.seek(position);

  Future<void> seekRelative(Duration delta) async {
    final raw = state.position + delta;
    final newPos = raw < Duration.zero ? Duration.zero : (raw > state.duration ? state.duration : raw);
    await seekTo(newPos);
  }

  Future<void> next() async {
    if (state.queue.isEmpty || state.currentIndex < 0) return;
    final nextIdx = state.repeat == pstate.RepeatMode.one
        ? state.currentIndex
        : (state.currentIndex + 1) % state.queue.length;
    await _playAtIndex(nextIdx);
  }

  Future<void> previous() async {
    if (state.queue.isEmpty || state.currentIndex < 0) return;
    if (state.position.inSeconds > 3) {
      await seekTo(Duration.zero);
    } else {
      final prevIdx = (state.currentIndex - 1 + state.queue.length) % state.queue.length;
      await _playAtIndex(prevIdx);
    }
  }

  Future<void> reorderQueue(int oldIndex, int newIndex) async {
    final queue = List<MediaItem>.from(state.queue);
    if (newIndex > oldIndex) newIndex--;
    final item = queue.removeAt(oldIndex);
    queue.insert(newIndex, item);
    state = state.copyWith(queue: queue);
  }

  Future<void> removeFromQueue(int index) async {
    final queue = List<MediaItem>.from(state.queue);
    if (index < 0 || index >= queue.length) return;
    queue.removeAt(index);
    if (index == state.currentIndex) {
      if (queue.isEmpty) {
        await _audioPlayer.stop();
        state = pstate.PlayerState();
        return;
      }
      final newCurrent = index < queue.length ? index : queue.length - 1;
      state = state.copyWith(queue: queue, currentIndex: newCurrent, clearMediaItem: true);
      _refreshConcatSource(queue.map((m) => m.id).toList());
      await _playAtIndex(newCurrent);
    } else {
      final newCurrent = state.currentIndex > index ? state.currentIndex - 1 : state.currentIndex;
      state = state.copyWith(queue: queue, currentIndex: newCurrent);
      _refreshConcatSource(queue.map((m) => m.id).toList());
    }
  }

  Future<void> setVolume(double volume) async {
    await _audioPlayer.setVolume(volume);
    state = state.copyWith(volume: volume);
  }

  Future<void> setRepeat(pstate.RepeatMode repeat) async {
    state = state.copyWith(repeat: repeat);
  }

  Future<void> setShuffle(bool shuffle) async {
    state = state.copyWith(shuffle: shuffle);
  }

  Future<void> stop() async {
    await _audioPlayer.stop();
    state = pstate.PlayerState();
  }

  // ---- OS media button bridge ----
  // audioHandler.customEvent carries skipToNext / skipToPrevious from
  // lock-screen / notification / Bluetooth controls.

  void _handleCustomEvent(dynamic payload) {
    if (payload is Map<String, dynamic> && payload['action'] == 'next') {
      next();
    } else if (payload is Map<String, dynamic> && payload['action'] == 'previous') {
      previous();
    }
  }

  // ---- Private helpers ----

  /// Fetch and cache CatalogTrack metadata. Returns null on failure.
  Future<CatalogTrack?> _resolveTrack(String trackId) async {
    if (_trackCache.containsKey(trackId)) return _trackCache[trackId];
    try {
      final resp = await _catalog.getCatalogTrack(id: trackId);
      final track = resp.data;
      if (track != null) _trackCache[trackId] = track;
      return track;
    } catch (_) {
      return null;
    }
  }

  /// Stub MediaItem used to populate the queue immediately before metadata resolves.
  MediaItem _stubMediaItem(String trackId) => MediaItem(
        id: trackId,
        title: _trackCache[trackId]?.title ?? trackId,
        artist: _artistNameCache[_trackCache[trackId]?.artistId ?? ''] ?? '',
        extras: {
          'trackId': trackId,
          'albumId': _trackCache[trackId]?.albumId,
        },
      );

  /// Full MediaItem populated from resolved CatalogTrack metadata.
  /// Artist name and album title are filled from the in-memory cache when
  /// available; otherwise an async backfill updates the state after lookup.
  MediaItem _makeMediaItem(String trackId, CatalogTrack? track) {
    final artistId = track?.artistId ?? '';
    final albumId = track?.albumId ?? '';
    final artistName = artistId.isNotEmpty ? (_artistNameCache[artistId] ?? artistId) : '';
    final albumTitle = albumId.isNotEmpty ? (_albumTitleCache[albumId] ?? '') : '';

    // Trigger background lookups when names aren't cached yet.
    if (artistId.isNotEmpty && !_artistNameCache.containsKey(artistId)) {
      _backfillArtistName(trackId, artistId);
    }
    if (albumId.isNotEmpty && !_albumTitleCache.containsKey(albumId)) {
      _backfillAlbumTitle(trackId, albumId);
    }

    return MediaItem(
      id: trackId,
      title: track?.title ?? trackId,
      artist: artistName,
      album: albumTitle,
      duration: track?.durationMs != null
          ? Duration(milliseconds: track!.durationMs!)
          : Duration.zero,
      artUri: null,
      extras: {
        'trackId': trackId,
        'mediaObjectId': track?.mediaObjectId,
        'albumId': albumId.isNotEmpty ? albumId : null,
      },
    );
  }

  /// Fetch artist name in the background and update the current MediaItem if
  /// it is the track being displayed.
  Future<void> _backfillArtistName(String trackId, String artistId) async {
    try {
      final name = await ref.read(artistNameProvider(artistId).future);
      _artistNameCache[artistId] = name;
      if (state.mediaItem?.id == trackId) {
        final updated = state.mediaItem!.copyWith(artist: name);
        state = state.copyWith(mediaItem: updated);
        audioHandler.mediaItem.add(updated);
      }
    } catch (_) {
      // Non-fatal: UUID shown as fallback.
    }
  }

  /// Fetch album title in the background and update the current MediaItem.
  Future<void> _backfillAlbumTitle(String trackId, String albumId) async {
    try {
      final title = await ref.read(albumTitleProvider(albumId).future);
      _albumTitleCache[albumId] = title;
      if (state.mediaItem?.id == trackId) {
        final updated = state.mediaItem!.copyWith(album: title);
        state = state.copyWith(mediaItem: updated);
        audioHandler.mediaItem.add(updated);
      }
    } catch (_) {
      // Non-fatal.
    }
  }

  Future<void> _playAtIndex(int index) async {
    if (index < 0 || index >= state.queue.length) return;
    final trackId = state.queue[index].id;
    state = state.copyWith(currentIndex: index);
    final queueIds = state.queue.map((m) => m.id).toList();
    await playTrack(trackId, queueIds: queueIds, index: index);
  }

  List<StreamSubscription> _setupPlayerListeners() {
    final subs = <StreamSubscription>[];

    // Position
    subs.add(_audioPlayer.positionStream.listen((pos) {
      state = state.copyWith(position: pos);
    }));

    // Duration
    subs.add(_audioPlayer.durationStream.listen((dur) {
      if (dur != null) state = state.copyWith(duration: dur);
    }));

    // Processing state — auto-advance on completion
    subs.add(_audioPlayer.processingStateStream.listen((ps) {
      if (ps == ProcessingState.completed) {
        _postHistory();
        if (state.repeat == pstate.RepeatMode.one) {
          _audioPlayer.seek(Duration.zero);
          _audioPlayer.play();
        } else {
          next();
        }
      }
    }));

    // Player state — playing/paused/buffering
    subs.add(_audioPlayer.playerStateStream.listen((ps) {
      state = state.copyWith(
        playbackState: PlaybackState(
          playing: ps.playing,
          processingState: _toAudioProcessingState(ps.processingState),
          controls: [
            MediaControl.skipToPrevious,
            if (ps.playing) MediaControl.pause else MediaControl.play,
            MediaControl.skipToNext,
          ],
        ),
      );
    }));

    // OS media button events (lock screen, notification, Bluetooth)
    subs.add(audioHandler.customEvent.listen(_handleCustomEvent));

    return subs;
  }

  AudioProcessingState _toAudioProcessingState(ProcessingState ps) {
    switch (ps) {
      case ProcessingState.idle:
        return AudioProcessingState.idle;
      case ProcessingState.loading:
        return AudioProcessingState.loading;
      case ProcessingState.buffering:
        return AudioProcessingState.buffering;
      case ProcessingState.ready:
        return AudioProcessingState.ready;
      case ProcessingState.completed:
        return AudioProcessingState.completed;
    }
  }

  Future<void> _postHistory() async {
    try {
      if (state.queue.isEmpty || state.currentIndex < 0 || state.currentIndex >= state.queue.length) return;
      final trackId = state.queue[state.currentIndex].id;
      if (trackId.isNotEmpty) {
        await _history.recordPlayEvent(
          recordPlayEventRequest: RecordPlayEventRequest(trackId: trackId),
        );
      }
    } catch (e) {
      debugPrint('PlayerNotifier: failed to post history: $e');
    }
  }
}
