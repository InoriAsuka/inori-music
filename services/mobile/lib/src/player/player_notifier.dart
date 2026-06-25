// ignore_for_file: implementation_imports
import 'package:audio_service/audio_service.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:just_audio/just_audio.dart';
import 'package:inori_api/src/api/catalog_api.dart';
import 'package:inori_api/src/api/history_api.dart';
import 'package:inori_api/src/model/record_play_event_request.dart';

import 'package:inori_music/src/api/api_client.dart';
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

  @override
  pstate.PlayerState build() {
    _audioPlayer = AudioPlayer();
    _catalog = ref.read(catalogApiProvider);
    _history = ref.read(historyApiProvider);
    _setupPlayerListeners();
    return pstate.PlayerState();
  }

  // ---- Public playback API ----

  /// Resolve the playback URL for a track and prepare the audio source.
  /// Returns the resolved URL or null if unavailable.
  Future<String?> resolvePlaybackUrl(String trackId) async {
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

    // Build queue if provided
    if (queueIds != null && queueIds.isNotEmpty) {
      final clampedIndex = index < 0 ? 0 : (index > queueIds.length - 1 ? queueIds.length - 1 : index);
      state = state.copyWith(
        queue: queueIds.map((id) => _makeMediaItem(id)).toList(),
        currentIndex: clampedIndex,
      );
    }

    final source = ProgressiveAudioSource(Uri.parse(url), tag: trackId);
    await _audioPlayer.setAudioSource(source);
    state = state.copyWith(
      mediaItem: _makeMediaItem(trackId),
      currentIndex: state.queue.isNotEmpty ? (state.currentIndex >= 0 ? state.currentIndex : index) : 0,
    );
    await _audioPlayer.play();
  }

  /// Play a list of track IDs starting at [initialIndex].
  Future<void> playQueue(List<String> trackIds, {int initialIndex = 0}) async {
    if (trackIds.isEmpty) return;
    final idx = initialIndex < 0 ? 0 : (initialIndex > trackIds.length - 1 ? trackIds.length - 1 : initialIndex);
    state = state.copyWith(queue: trackIds.map((id) => _makeMediaItem(id)).toList());
    await playTrack(trackIds[idx], queueIds: trackIds, index: idx);
  }

  /// Enqueue tracks after the current position.
  Future<void> enqueue(List<String> trackIds) async {
    final items = trackIds.map((id) => _makeMediaItem(id)).toList();
    final newQueue = [...state.queue, ...items];
    state = state.copyWith(queue: newQueue);
  }

  /// Enqueue a single track immediately after the current one.
  Future<void> enqueueNext(String trackId) async {
    final item = _makeMediaItem(trackId);
    final newQueue = [...state.queue];
    final insertAt = state.currentIndex + 1;
    if (insertAt < newQueue.length) {
      newQueue.insert(insertAt, item);
    } else {
      newQueue.add(item);
    }
    state = state.copyWith(queue: newQueue);
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
      await _playAtIndex(newCurrent);
    } else {
      final newCurrent = state.currentIndex > index ? state.currentIndex - 1 : state.currentIndex;
      state = state.copyWith(queue: queue, currentIndex: newCurrent);
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

  // ---- Private helpers ----

  Future<void> _playAtIndex(int index) async {
    if (index < 0 || index >= state.queue.length) return;
    final trackId = state.queue[index].id;
    state = state.copyWith(currentIndex: index);
    final queueIds = state.queue.map((m) => m.id).toList();
    await playTrack(trackId, queueIds: queueIds, index: index);
  }

  MediaItem _makeMediaItem(String trackId) => MediaItem(
        id: trackId,
        title: trackId,
        artist: '',
        album: '',
        duration: const Duration(seconds: 0),
        artUri: null,
        extras: {'trackId': trackId},
      );

  void _setupPlayerListeners() {
    // Position
    _audioPlayer.positionStream.listen((pos) {
      state = state.copyWith(position: pos);
    });

    // Duration
    _audioPlayer.durationStream.listen((dur) {
      if (dur != null) state = state.copyWith(duration: dur);
    });

    // Processing state — auto-advance on completion
    _audioPlayer.processingStateStream.listen((ps) {
      if (ps == ProcessingState.completed) {
        _postHistory();
        if (state.repeat == pstate.RepeatMode.one) {
          _audioPlayer.seek(Duration.zero);
          _audioPlayer.play();
        } else {
          next();
        }
      }
    });

    // Player state — playing/paused/buffering
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
    });
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
