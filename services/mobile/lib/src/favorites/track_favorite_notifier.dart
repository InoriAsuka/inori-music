// ignore_for_file: implementation_imports
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/player/player_notifier.dart';

/// Per-track favorite-state notifier.
///
/// Uses [autoDispose] so notifier instances are released when no widget is
/// watching them (important for large track lists that create one per tile).
class TrackFavoriteNotifier extends AutoDisposeFamilyNotifier<bool, String> {
  @override
  bool build(String trackId) => false; // seeded on first watch via [init]

  /// Seed the initial state from the catalog model's isFavorite field.
  /// Called once from initState via addPostFrameCallback.
  void init(bool value) {
    // Only update if state actually differs to avoid spurious rebuilds.
    if (state != value) state = value;
  }

  Future<void> toggle() async {
    final api = ref.read(historyApiProvider);
    final current = state;
    // Optimistic update
    state = !current;
    try {
      if (current) {
        await api.removeFavoriteTrack(trackId: arg);
      } else {
        await api.addFavoriteTrack(trackId: arg);
      }
    } catch (_) {
      // Roll back on failure
      state = current;
    }
  }
}

final trackFavoriteProvider =
    NotifierProvider.autoDispose.family<TrackFavoriteNotifier, bool, String>(
  TrackFavoriteNotifier.new,
);
