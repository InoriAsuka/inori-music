// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/shared/widgets/track_list_tile.dart';

// ---------------------------------------------------------------------------
// Favorites provider
// ---------------------------------------------------------------------------

final favoritesProvider = FutureProvider<List<CatalogTrack>>((ref) async {
  final api = ref.read(historyApiProvider);
  final resp = await api.listFavoriteTracks(limit: 500);
  final page = resp.data;
  if (page == null || page.trackIds.isEmpty) return [];

  // Fetch full track details for each favorite
  final catalogApi = ref.read(catalogApiProvider);
  final tracks = <CatalogTrack>[];
  for (final id in page.trackIds) {
    try {
      final trackResp = await catalogApi.getCatalogTrack(id: id);
      if (trackResp.data != null) {
        final t = trackResp.data!;
        tracks.add(CatalogTrack(
          id: t.id,
          artistId: t.artistId,
          createdAt: t.createdAt,
          mediaObjectId: t.mediaObjectId,
          title: t.title,
          updatedAt: t.updatedAt,
          albumId: t.albumId,
          discNumber: t.discNumber,
          durationMs: t.durationMs,
          genre: t.genre,
          isFavorite: true,
          sortTitle: t.sortTitle,
          trackNumber: t.trackNumber,
        ));
      }
    } catch (_) {
      // Skip tracks that can't be resolved
    }
  }
  return tracks;
});

/// Toggle a track's favorite status and refresh the list.
Future<void> _toggleFavorite(WidgetRef ref, CatalogTrack track) async {
  final api = ref.read(historyApiProvider);
  final isFav = track.isFavorite ?? false;

  try {
    if (isFav) {
      await api.removeFavoriteTrack(trackId: track.id);
    } else {
      await api.addFavoriteTrack(trackId: track.id);
    }
    // Refresh the full list — result is the future, which we ignore intentionally
    // ignore: unused_result
    ref.refresh(favoritesProvider);
  } catch (_) {
    // Silently fail — user can retry
  }
}

// ---------------------------------------------------------------------------
// Favorites Screen
// ---------------------------------------------------------------------------

class FavoritesScreen extends ConsumerWidget {
  const FavoritesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(favoritesProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Favorites')),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.error_outline, color: NeonShrineColors.error, size: 48),
              const SizedBox(height: 12),
              Text('$e', textAlign: TextAlign.center),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: () => ref.refresh(favoritesProvider),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (tracks) => tracks.isEmpty
            ? const Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.favorite_border, size: 64, color: NeonShrineColors.onSurfaceVariant),
                    SizedBox(height: 16),
                    Text('No favorites yet', style: TextStyle(fontSize: 18, color: NeonShrineColors.onSurfaceVariant)),
                  ],
                ),
              )
            : ListView.builder(
                itemCount: tracks.length,
                itemBuilder: (context, i) {
                  final track = tracks[i];
                  return TrackListTile(
                    track: track,
                    isFavorite: true,
                    onFavoriteTap: () => _toggleFavorite(ref, track),
                  );
                },
              ),
      ),
    );
  }
}
