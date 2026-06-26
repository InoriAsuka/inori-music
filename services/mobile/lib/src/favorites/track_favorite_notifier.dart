// ignore_for_file: implementation_imports
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/player/player_notifier.dart';

/// Per-track favorite-state notifier.
/// Observes the current favorite state (from the track's `isFavorite` field)
/// and can toggle it via POST / DELETE on the API.
class TrackFavoriteNotifier extends FamilyNotifier<bool, String> {
  @override
  bool build(String trackId) => false; // initialised externally via [init]

  /// Seed the initial state from the catalog's isFavorite value.
  void init(bool value) => state = value;

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
    NotifierProviderFamily<TrackFavoriteNotifier, bool, String>(
  TrackFavoriteNotifier.new,
);
