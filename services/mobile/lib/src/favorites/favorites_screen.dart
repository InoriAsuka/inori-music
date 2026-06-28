// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/favorites/track_favorite_notifier.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/shared/widgets/track_list_tile.dart';
import 'package:inori_music/src/user_playlist/user_playlist_notifier.dart';

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
    final state = ref.watch(favoritesProvider);
    final playlistsState = ref.watch(userPlaylistProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Library')),
      body: CustomScrollView(
        slivers: [
          // ---- My Playlists section ----
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 8, 4),
              child: Row(
                children: [
                  const Text(
                    'My Playlists',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                      color: NeonShrineColors.onSurface,
                    ),
                  ),
                  const Spacer(),
                  IconButton(
                    icon: const Icon(Icons.add, color: NeonShrineColors.primaryViolet),
                    tooltip: 'New Playlist',
                    onPressed: () => _showCreateDialog(context, ref),
                  ),
                ],
              ),
            ),
          ),
          playlistsState.when(
            loading: () => const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.all(16),
                child: Center(child: CircularProgressIndicator()),
              ),
            ),
            error: (e, _) => SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Text('$e',
                    style: const TextStyle(color: NeonShrineColors.error)),
              ),
            ),
            data: (playlists) => playlists.isEmpty
                ? const SliverToBoxAdapter(
                    child: Padding(
                      padding: EdgeInsets.fromLTRB(16, 0, 16, 8),
                      child: Text(
                        'No playlists yet. Tap + to create one.',
                        style: TextStyle(color: NeonShrineColors.onSurfaceVariant),
                      ),
                    ),
                  )
                : SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, i) {
                        final pl = playlists[i];
                        return ListTile(
                          leading: Container(
                            width: 44,
                            height: 44,
                            decoration: BoxDecoration(
                              color: NeonShrineColors.surfaceContainer,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: const Icon(Icons.queue_music,
                                color: NeonShrineColors.primaryViolet),
                          ),
                          title: Text(pl.name),
                          subtitle: Text('${pl.trackIds.length} tracks'),
                          trailing: IconButton(
                            icon: const Icon(Icons.delete_outline,
                                size: 20,
                                color: NeonShrineColors.onSurfaceVariant),
                            onPressed: () async {
                              await ref
                                  .read(userPlaylistProvider.notifier)
                                  .delete(pl.id);
                            },
                          ),
                          onTap: () => context.go(AppRoutes.myPlaylistDetailPath(pl.id)),
                        );
                      },
                      childCount: playlists.length,
                    ),
                  ),
          ),
          // ---- Favorites section ----
          const SliverToBoxAdapter(
            child: Padding(
              padding: EdgeInsets.fromLTRB(16, 16, 16, 4),
              child: Text(
                'Favorites',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: NeonShrineColors.onSurface,
                ),
              ),
            ),
          ),
          state.when(
            loading: () => const SliverToBoxAdapter(
              child: Center(child: CircularProgressIndicator()),
            ),
            error: (e, _) => SliverToBoxAdapter(
              child: Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.error_outline,
                        color: NeonShrineColors.error, size: 48),
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
            ),
            data: (tracks) => tracks.isEmpty
                ? const SliverToBoxAdapter(
                    child: Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.favorite_border,
                              size: 64,
                              color: NeonShrineColors.onSurfaceVariant),
                          SizedBox(height: 16),
                          Text('No favorites yet',
                              style: TextStyle(
                                  fontSize: 18,
                                  color: NeonShrineColors.onSurfaceVariant)),
                        ],
                      ),
                    ),
                  )
                : SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, i) => _FavTile(
                        key: ValueKey(tracks[i].id),
                        track: tracks[i],
                        onRemoved: () => ref.refresh(favoritesProvider),
                      ),
                      childCount: tracks.length,
                    ),
                  ),
          ),
        ],
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref) async {
    final controller = TextEditingController();
    final name = await showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('New Playlist'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Playlist name'),
        ),
        actions: [
          TextButton(
            onPressed: () => ctx.pop(),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => ctx.pop(controller.text.trim()),
            child: const Text('Create'),
          ),
        ],
      ),
    );
    if (name != null && name.isNotEmpty) {
      await ref.read(userPlaylistProvider.notifier).create(name);
    }
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
