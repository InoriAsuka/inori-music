// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/favorites/track_favorite_notifier.dart';
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

// ---------------------------------------------------------------------------
// Favorites Screen
// ---------------------------------------------------------------------------

class FavoritesScreen extends ConsumerWidget {
  const FavoritesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final state = ref.watch(favoritesProvider);

    return Scaffold(
      appBar: AppBar(title: Text(t.favorites)),
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
                child: Text(t.retry),
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
                itemBuilder: (context, i) => _FavTile(
                  key: ValueKey(tracks[i].id),
                  track: tracks[i],
                  onRemoved: () => ref.refresh(favoritesProvider),
                ),
              ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Per-tile widget — uses trackFavoriteProvider for optimistic toggle.
// When a track is removed from favorites it triggers a full list refresh
// so the tile disappears from the screen.
// ---------------------------------------------------------------------------

class _FavTile extends ConsumerStatefulWidget {
  const _FavTile({super.key, required this.track, required this.onRemoved});
  final CatalogTrack track;
  final VoidCallback onRemoved;

  @override
  ConsumerState<_FavTile> createState() => _FavTileState();
}

class _FavTileState extends ConsumerState<_FavTile> {
  @override
  void initState() {
    super.initState();
    // Seed per-track notifier: all tracks in FavoritesScreen are favorites.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      ref.read(trackFavoriteProvider(widget.track.id).notifier).init(true);
    });
  }

  @override
  Widget build(BuildContext context) {
    final isFav = ref.watch(trackFavoriteProvider(widget.track.id));

    // If the optimistic toggle moved this track out of favorites, refresh
    // the full list so it disappears without waiting for the user to navigate.
    if (!isFav) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        widget.onRemoved();
      });
    }

    return TrackListTile(
      track: widget.track,
      isFavorite: isFav,
      onFavoriteTap: () =>
          ref.read(trackFavoriteProvider(widget.track.id).notifier).toggle(),
    );
  }
}
