// ignore_for_file: implementation_imports
import 'dart:async';
import 'dart:io';
import 'dart:math' show pow;

import 'package:audio_service/audio_service.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_music/src/local_library/local_library_db.dart';
import 'package:inori_music/src/local_library/local_library_notifier.dart'
    show localTrackIdPrefix;
import 'package:inori_music/src/offline/offline_db.dart';
import 'package:inori_api/src/api/catalog_api.dart';
import 'package:inori_api/src/api/history_api.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:inori_api/src/model/record_play_event_request.dart';
import 'package:just_audio/just_audio.dart';

import 'package:inori_music/main.dart' show audioHandler;
import 'package:inori_music/src/api/api_client.dart';
import 'package:inori_music/src/api/me_api.dart';
import 'package:inori_music/src/audio/replay_gain_notifier.dart';
import 'package:inori_music/src/catalog/catalog_cache_providers.dart';
import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/player/player_state_reporter.dart';

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
  late final MeApi _me;

  // In-memory track metadata cache to avoid redundant catalog API calls.
  final Map<String, CatalogTrack> _trackCache = {};

  // Same idea for guest-mode local-file tracks (id prefix `local:`) — see
  // resolvePlaybackUrl / _localMediaItem / _backfillLocalTrack below.
  final Map<String, LocalLibraryTrack> _localTrackCache = {};

  // Resolved display names — keyed by artistId / albumId.
  final Map<String, String> _artistNameCache = {};
  final Map<String, String> _albumTitleCache = {};

  // Store subscriptions so they can be cancelled on dispose.
  late final List<StreamSubscription> _subscriptions;

  // Cross-device player-state reporter (v5.4.0): 30s throttle while playing,
  // immediate PUT on track change / pause / app background.
  late final PlayerStateReporter _reporter;
  // Last playing flag observed, so playerStateStream only triggers an immediate
  // report on an actual play↔pause transition (not on every buffering tick).
  bool _lastReportedPlaying = false;

  @override
  pstate.PlayerState build() {
    // Use the AudioPlayer instance owned by the AudioHandler so that
    // audio_service (notifications, MediaSession, lock screen) stays in sync.
    _audioPlayer = audioHandler.audioPlayer;
    _catalog = ref.read(catalogApiProvider);
    _history = ref.read(historyApiProvider);
    _me = ref.read(meApiProvider);
    _reporter = PlayerStateReporter(onReport: _reportPlayerState);
    _subscriptions = _setupPlayerListeners();
    // Cancel all stream subscriptions when the provider is disposed.
    ref.onDispose(() {
      _reporter.dispose();
      for (final sub in _subscriptions) {
        sub.cancel();
      }
    });
    // Recompute effective volume immediately when the ReplayGain toggle flips.
    ref.listen(
      replayGainEnabledProvider,
      (_, _) => _applyVolumeWithGain(state.volume),
    );
    return pstate.PlayerState();
  }

  // ---- Public playback API ----

  /// Resolve the playback URL for a track and prepare the audio source.
  /// Returns the resolved URL or null if unavailable.
  Future<String?> resolvePlaybackUrl(String trackId) async {
    // Guest-mode local file — never touches the server or OfflineDb.
    if (trackId.startsWith(localTrackIdPrefix)) {
      final local = await LocalLibraryDb.instance.query(trackId);
      if (local == null) return null;
      // Mirrors the OfflineDb existsSync() check just below — without it, a
      // missing file (moved/deleted outside the app, or a container reset)
      // would still return a file:// URL, leaving just_audio to fail on
      // whatever it does with a nonexistent path rather than failing here
      // with a clear reason.
      if (!File(local.localPath).existsSync()) {
        debugPrint(
          'PlayerNotifier: local file missing for $trackId at ${local.localPath}',
        );
        return null;
      }
      return 'file://${local.localPath}';
    }
    // Check local offline cache first.
    final offline = await OfflineDb.instance.query(trackId);
    if (offline != null && File(offline.localPath).existsSync()) {
      return 'file://${offline.localPath}';
    }
    try {
      final resp = await _catalog.getTrackPlaybackDescriptor(id: trackId);
      final descriptor = resp.data;
      if (descriptor == null) return null;

      if (descriptor.presignedUrl != null &&
          descriptor.presignedUrl!.isNotEmpty) {
        return descriptor.presignedUrl;
      }
      if (descriptor.streamUrl != null && descriptor.streamUrl!.isNotEmpty) {
        // streamUrl already carries HMAC signature from the server — use as-is.
        return descriptor.streamUrl;
      }
      // Fallback: construct stream URL; Flutter sends Authorization: Bearer <token>.
      final base = await ref.read(baseUrlProvider.future);
      return '$base/api/v1/catalog/tracks/$trackId/stream';
    } catch (e) {
      debugPrint(
        'PlayerNotifier: failed to resolve playback URL for $trackId: $e',
      );
      return null;
    }
  }

  /// Play a single track by ID, optionally building the full queue.
  ///
  /// Everything from URL resolution through `play()` is wrapped in a single
  /// try/catch that surfaces failures via [PlayerState.playbackError] —
  /// previously an exception anywhere in this chain (e.g. just_audio
  /// rejecting a local file) was an unhandled Future error: invisible in a
  /// release build, so a broken track looked exactly like "nothing happens."
  Future<void> playTrack(
    String trackId, {
    List<String>? queueIds,
    int index = 0,
  }) async {
    state = state.copyWith(clearPlaybackError: true);
    try {
      final url = await resolvePlaybackUrl(trackId);
      if (url == null) {
        state = state.copyWith(
          playbackError: pstate.PlaybackFailure(
            'Could not resolve a playback source for this track ($trackId)',
          ),
        );
        return;
      }

      // Build queue with stub items immediately so the UI has something to render,
      // then update the current item with real metadata once resolved.
      if (queueIds != null && queueIds.isNotEmpty) {
        final clampedIndex = index < 0
            ? 0
            : (index > queueIds.length - 1 ? queueIds.length - 1 : index);
        state = state.copyWith(
          queue: queueIds.map((id) => _stubMediaItem(id)).toList(),
          currentIndex: clampedIndex,
        );
        // Resolve all URLs for gapless playback via ConcatenatingAudioSource.
        // Must be awaited — this is what actually calls _audioPlayer.setAudioSource
        // for the queued-playback path (the single-track branch below only runs
        // when there's no queue). This used to be fire-and-forget, racing against
        // play() below: server tracks had enough other network latency in this
        // function to usually mask the race, but local files resolve near-
        // instantly (plain SQLite lookups, no network), so play() routinely won
        // the race and fired with no audio source set at all — correct metadata/
        // artwork (from _resolveTrack/_makeMediaItem below, unrelated to this),
        // but silent playback stuck at 0:00. _buildConcatQueue resolves the
        // queue's URLs in parallel so this await doesn't reintroduce a
        // perceptible delay for long server playlists.
        await _buildConcatQueue(queueIds, clampedIndex, url, trackId);
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
        currentIndex: state.queue.isNotEmpty
            ? (state.currentIndex >= 0 ? state.currentIndex : index)
            : 0,
      );
      await _audioPlayer.play();
    } catch (e, st) {
      debugPrint('PlayerNotifier: playback failed for $trackId: $e\n$st');
      state = state.copyWith(
        playbackError: pstate.PlaybackFailure('Playback failed: $e'),
      );
    }
  }

  /// Pre-resolves all queue URLs (in parallel — this is awaited by [playTrack]
  /// before it calls play(), so a long queue must not turn into N sequential
  /// round trips) and calls [audioHandler.updateConcatQueue] so
  /// ConcatenatingAudioSource enables gapless transitions.
  Future<void> _buildConcatQueue(
    List<String> queueIds,
    int startIndex,
    String resolvedUrl,
    String resolvedTrackId,
  ) async {
    final urls = await Future.wait(
      queueIds.map((id) async {
        if (id == resolvedTrackId) return resolvedUrl;
        return await resolvePlaybackUrl(id) ?? '';
      }),
    );
    // Filter out empty URLs but rebuild index accordingly.
    final validPairs = <MapEntry<int, String>>[];
    for (var i = 0; i < urls.length; i++) {
      if (urls[i].isNotEmpty) validPairs.add(MapEntry(i, urls[i]));
    }
    if (validPairs.isEmpty) return;
    await audioHandler.updateConcatQueue(
      validPairs.map((e) => e.value).toList(),
    );
    // Seek to the correct position in the ConcatenatingAudioSource.
    final concatIndex = validPairs.indexWhere((e) => e.key == startIndex);
    if (concatIndex > 0) {
      await _audioPlayer.seek(Duration.zero, index: concatIndex);
    }
  }

  /// Play a list of track IDs starting at [initialIndex].
  Future<void> playQueue(List<String> trackIds, {int initialIndex = 0}) async {
    if (trackIds.isEmpty) return;
    final idx = initialIndex < 0
        ? 0
        : (initialIndex > trackIds.length - 1
              ? trackIds.length - 1
              : initialIndex);
    state = state.copyWith(
      queue: trackIds.map((id) => _stubMediaItem(id)).toList(),
    );
    await playTrack(trackIds[idx], queueIds: trackIds, index: idx);
  }

  /// Enqueue tracks after the current position.
  Future<void> enqueue(List<String> trackIds) async {
    final items = trackIds.map((id) => _stubMediaItem(id)).toList();
    final newQueue = [...state.queue, ...items];
    state = state.copyWith(queue: newQueue);
    for (final id in trackIds) {
      final url = await resolvePlaybackUrl(id);
      if (url != null) await audioHandler.addSource(url);
    }
  }

  /// Enqueue a single track immediately after the current one.
  Future<void> enqueueNext(String trackId) async {
    final item = _stubMediaItem(trackId);
    final newQueue = [...state.queue];
    final insertAt = state.currentIndex + 1;
    final append = insertAt >= newQueue.length;
    if (append) {
      newQueue.add(item);
    } else {
      newQueue.insert(insertAt, item);
    }
    state = state.copyWith(queue: newQueue);
    final url = await resolvePlaybackUrl(trackId);
    if (url != null) {
      if (append) {
        await audioHandler.addSource(url);
      } else {
        await audioHandler.insertSource(insertAt, url);
      }
    }
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
    final newPos = raw < Duration.zero
        ? Duration.zero
        : (raw > state.duration ? state.duration : raw);
    await seekTo(newPos);
  }

  Future<void> next() async {
    if (state.queue.isEmpty || state.currentIndex < 0) return;
    if (state.repeat == pstate.RepeatMode.one) {
      await _audioPlayer.seek(Duration.zero);
      await _audioPlayer.play();
      return;
    }
    final nextIdx = state.currentIndex + 1;
    if (nextIdx >= state.queue.length) {
      if (state.repeat == pstate.RepeatMode.all) {
        await _playAtIndex(0);
      }
      // RepeatMode.none 播完自然停止 — 什么也不做，concat 源会自己播完停
      return;
    }
    await _audioPlayer.seekToNext();
  }

  Future<void> previous() async {
    if (state.queue.isEmpty || state.currentIndex < 0) return;
    if (state.position.inSeconds > 3) {
      await seekTo(Duration.zero);
    } else {
      await _audioPlayer.seekToPrevious();
    }
  }

  Future<void> reorderQueue(int oldIndex, int newIndex) async {
    final queue = List<MediaItem>.from(state.queue);
    if (newIndex > oldIndex) newIndex--;
    final item = queue.removeAt(oldIndex);
    queue.insert(newIndex, item);
    final oldCurrent = state.currentIndex;
    int newCurrent = oldCurrent;
    if (oldCurrent == oldIndex) {
      newCurrent = newIndex;
    } else if (oldIndex < oldCurrent && newIndex >= oldCurrent) {
      newCurrent = oldCurrent - 1;
    } else if (oldIndex > oldCurrent && newIndex <= oldCurrent) {
      newCurrent = oldCurrent + 1;
    }
    state = state.copyWith(queue: queue, currentIndex: newCurrent);
    await audioHandler.moveSource(oldIndex, newIndex);
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
      state = state.copyWith(
        queue: queue,
        currentIndex: newCurrent,
        clearMediaItem: true,
      );
      await audioHandler.removeSourceAt(index);
      await _playAtIndex(newCurrent);
    } else {
      final newCurrent = state.currentIndex > index
          ? state.currentIndex - 1
          : state.currentIndex;
      state = state.copyWith(queue: queue, currentIndex: newCurrent);
      await audioHandler.removeSourceAt(index);
    }
  }

  Future<void> setVolume(double volume) async {
    state = state.copyWith(volume: volume); // 只存用户意图
    await _applyVolumeWithGain(volume);
  }

  Future<void> _applyVolumeWithGain(double userVol) async {
    var gain = 1.0;
    if (ref.read(replayGainEnabledProvider)) {
      final id =
          state.currentIndex >= 0 && state.currentIndex < state.queue.length
          ? state.queue[state.currentIndex].id
          : null;
      final db = id != null ? _trackCache[id]?.replayGainDb : null;
      if (db != null) gain = pow(10, db / 20).toDouble().clamp(0.1, 2.0);
    }
    final effective = (userVol * gain).clamp(0.0, 1.0);
    audioHandler.targetVolume = effective; // Step 5 的 fade 基准同步
    await _audioPlayer.setVolume(effective);
  }

  Future<void> setRepeat(pstate.RepeatMode repeat) async {
    state = state.copyWith(repeat: repeat);
    switch (repeat) {
      case pstate.RepeatMode.one:
        await _audioPlayer.setLoopMode(LoopMode.one);
        break;
      case pstate.RepeatMode.all:
        await _audioPlayer.setLoopMode(LoopMode.all);
        break;
      case pstate.RepeatMode.none:
        await _audioPlayer.setLoopMode(LoopMode.off);
        break;
    }
  }

  Future<void> setShuffle(bool shuffle) async {
    state = state.copyWith(shuffle: shuffle);
    await _audioPlayer.setShuffleModeEnabled(shuffle);
    if (shuffle) await _audioPlayer.shuffle();
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
    } else if (payload is Map<String, dynamic> &&
        payload['action'] == 'previous') {
      previous();
    }
  }

  // ---- Private helpers ----

  /// Fetch and cache CatalogTrack metadata. Returns null on failure.
  Future<CatalogTrack?> _resolveTrack(String trackId) async {
    // Local tracks never have a CatalogTrack — _makeMediaItem branches to
    // _localMediaItem for these regardless of what this returns.
    if (trackId.startsWith(localTrackIdPrefix)) return null;
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
  MediaItem _stubMediaItem(String trackId) {
    if (trackId.startsWith(localTrackIdPrefix)) return _localMediaItem(trackId);
    return MediaItem(
      id: trackId,
      title: _trackCache[trackId]?.title ?? trackId,
      artist: _artistNameCache[_trackCache[trackId]?.artistId ?? ''] ?? '',
      extras: {'trackId': trackId, 'albumId': _trackCache[trackId]?.albumId},
    );
  }

  /// MediaItem for a guest-mode local file. Reads the in-memory cache
  /// populated from [LocalLibraryDb]; on a cache miss it returns an
  /// id-only stub and kicks off an async backfill (same pattern as
  /// [_backfillArtistName]/[_backfillAlbumTitle] below) rather than making
  /// this function async — SQLite lookups are fast but queue construction
  /// needs a synchronous MediaItem immediately.
  MediaItem _localMediaItem(String trackId) {
    final cached = _localTrackCache[trackId];
    if (cached == null) {
      _backfillLocalTrack(trackId);
      return MediaItem(id: trackId, title: trackId, artist: '', album: '');
    }
    return MediaItem(
      id: trackId,
      title: cached.title,
      artist: cached.artistName,
      album: cached.albumTitle,
      duration: cached.durationMs != null
          ? Duration(milliseconds: cached.durationMs!)
          : Duration.zero,
      artUri: cached.coverArtPath != null
          ? Uri.file(cached.coverArtPath!)
          : null,
      extras: {'trackId': trackId},
    );
  }

  /// Fetch local track metadata in the background and update the current
  /// MediaItem if it is the track being displayed.
  Future<void> _backfillLocalTrack(String trackId) async {
    final row = await LocalLibraryDb.instance.query(trackId);
    if (row == null) return;
    _localTrackCache[trackId] = row;
    if (state.mediaItem?.id == trackId) {
      final updated = _localMediaItem(trackId);
      state = state.copyWith(mediaItem: updated);
      audioHandler.mediaItem.add(updated);
    }
  }

  /// Full MediaItem populated from resolved CatalogTrack metadata.
  /// Artist name and album title are filled from the in-memory cache when
  /// available; otherwise an async backfill updates the state after lookup.
  MediaItem _makeMediaItem(String trackId, CatalogTrack? track) {
    if (trackId.startsWith(localTrackIdPrefix)) return _localMediaItem(trackId);
    final artistId = track?.artistId ?? '';
    final albumId = track?.albumId ?? '';
    final artistName = artistId.isNotEmpty
        ? (_artistNameCache[artistId] ?? artistId)
        : '';
    final albumTitle = albumId.isNotEmpty
        ? (_albumTitleCache[albumId] ?? '')
        : '';

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
    subs.add(
      _audioPlayer.positionStream.listen((pos) {
        state = state.copyWith(position: pos);
      }),
    );

    // Duration
    subs.add(
      _audioPlayer.durationStream.listen((dur) {
        if (dur != null) state = state.copyWith(duration: dur);
      }),
    );

    // Concat source index — gapless auto-advance and manual seekToNext/Previous
    // both surface here. Keeps state.currentIndex, notification metadata, and
    // per-track history in sync with what just_audio is actually playing.
    subs.add(
      (() {
        int? lastIndex;
        return _audioPlayer.currentIndexStream.listen((idx) async {
          if (idx == null || idx < 0 || idx >= state.queue.length) return;
          if (idx == state.currentIndex) {
            lastIndex = idx;
            return;
          }
          // 上报"刚离开"的曲目（gapless 自动前进时逐曲触发）
          if (lastIndex != null &&
              lastIndex! >= 0 &&
              lastIndex! < state.queue.length) {
            await _postHistoryFor(state.queue[lastIndex!].id);
          }
          lastIndex = idx;
          final trackId = state.queue[idx].id;
          final track = await _resolveTrack(trackId);
          final resolved = _makeMediaItem(trackId, track);
          state = state.copyWith(currentIndex: idx, mediaItem: resolved);
          audioHandler.mediaItem.add(resolved);
          await _applyVolumeWithGain(state.volume);
          // Cross-device sync: report the new track immediately (v5.4.0).
          _reporter.reportNow();
        });
      })(),
    );

    // Processing state — fires only when the whole concat queue has finished
    // playing (native LoopMode already handles one/all looping internally
    // without ever reaching `completed`).
    subs.add(
      _audioPlayer.processingStateStream.listen((ps) {
        if (ps == ProcessingState.completed) {
          if (state.currentIndex >= 0 &&
              state.currentIndex < state.queue.length) {
            _postHistoryFor(state.queue[state.currentIndex].id);
          }
          if (state.repeat == pstate.RepeatMode.one) {
            _audioPlayer.seek(Duration.zero);
            _audioPlayer.play();
          }
          // RepeatMode.all/none: native loopMode already handled wrap/stop.
        }
      }),
    );

    // Player state — playing/paused/buffering
    subs.add(
      _audioPlayer.playerStateStream.listen((ps) {
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
        // Cross-device sync: gate the 30s throttle on playback, and PUT
        // immediately on a play↔pause transition (v5.4.0).
        _reporter.setPlaying(ps.playing);
        if (ps.playing != _lastReportedPlaying) {
          _lastReportedPlaying = ps.playing;
          _reporter.reportNow();
        }
      }),
    );

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

  Future<void> _postHistoryFor(String trackId) async {
    // Local files have no server-side history to report against.
    if (trackId.startsWith(localTrackIdPrefix)) return;
    try {
      if (trackId.isNotEmpty) {
        await _history.recordPlayEvent(
          recordPlayEventRequest: RecordPlayEventRequest(trackId: trackId),
        );
      }
    } catch (e) {
      debugPrint('PlayerNotifier: failed to post history: $e');
    }
  }

  // ---- Cross-device player-state sync (v5.4.0) ----

  static String _repeatToWire(pstate.RepeatMode mode) {
    switch (mode) {
      case pstate.RepeatMode.one:
        return 'one';
      case pstate.RepeatMode.all:
        return 'all';
      case pstate.RepeatMode.none:
        return 'off';
    }
  }

  static pstate.RepeatMode _repeatFromWire(String wire) {
    switch (wire) {
      case 'one':
        return pstate.RepeatMode.one;
      case 'all':
        return pstate.RepeatMode.all;
      default:
        return pstate.RepeatMode.none;
    }
  }

  /// Serialize the current playback state into the cross-device wire DTO.
  PlayerStateDto _serializeState() {
    return PlayerStateDto(
      queue: state.queue.map((m) => m.id).toList(),
      currentIndex: state.currentIndex,
      positionSeconds: state.position.inMilliseconds / 1000.0,
      repeat: _repeatToWire(state.repeat),
      shuffle: state.shuffle,
      volume: state.volume,
      speed: _audioPlayer.speed,
      status: state.isPlaying
          ? 'playing'
          : (state.isIdle ? 'stopped' : 'paused'),
    );
  }

  /// PUT the current player state to the server. Skips empty/idle queues so a
  /// fresh install never overwrites another device's state with nothing.
  /// Swallows all errors — reporting must never break playback.
  Future<void> _reportPlayerState() async {
    if (state.queue.isEmpty || state.currentIndex < 0) return;
    // Guest-mode local files don't exist on other devices — reporting them
    // would tell another device to resume a queue it can never actually
    // play. Skip the whole report while a local track is current rather
    // than trying to filter-and-reindex the queue.
    if (state.currentIndex < state.queue.length &&
        state.queue[state.currentIndex].id.startsWith(localTrackIdPrefix)) {
      return;
    }
    try {
      await _me.putPlayerState(_serializeState());
    } catch (e) {
      debugPrint('PlayerNotifier: failed to report player state: $e');
    }
  }

  /// Force an immediate player-state report. Called when the app is backgrounded
  /// (see [InoriMusicApp]'s lifecycle listener) so progress survives a kill.
  void reportStateOnBackground() => _reporter.reportNow();

  /// Fetch the last cross-device player state, or null if none / on error.
  Future<PlayerStateDto?> fetchRemoteState() async {
    try {
      return await _me.getPlayerState();
    } catch (e) {
      debugPrint('PlayerNotifier: failed to fetch remote player state: $e');
      return null;
    }
  }

  /// Rebuild the queue from a remote snapshot and seek to its position without
  /// auto-playing (waits for a user gesture, mirroring the web resume flow).
  Future<void> resumeFromRemote(PlayerStateDto remote) async {
    if (remote.queue.isEmpty) return;
    final idx = remote.currentIndex < 0
        ? 0
        : (remote.currentIndex >= remote.queue.length
              ? remote.queue.length - 1
              : remote.currentIndex);
    await setRepeat(_repeatFromWire(remote.repeat));
    await setShuffle(remote.shuffle);
    await playQueue(remote.queue, initialIndex: idx);
    await pause();
    await seekTo(
      Duration(milliseconds: (remote.positionSeconds * 1000).round()),
    );
  }
}
