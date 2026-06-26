// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/catalog_album.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/favorites/track_favorite_notifier.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/shared/widgets/track_list_tile.dart';

final _artistDetailProvider = FutureProvider.family<CatalogArtist, String>((ref, id) {
  return ref.watch(catalogRepositoryProvider).getArtist(id);
});

final _artistAlbumsProvider = FutureProvider.family<List<CatalogAlbum>, String>((ref, id) {
  return ref.watch(catalogRepositoryProvider).albumsByArtist(id);
});

final _artistTracksProvider = FutureProvider.family<List<CatalogTrack>, String>((ref, id) {
  return ref.watch(catalogRepositoryProvider).tracksByArtist(id);
});

class ArtistDetailScreen extends ConsumerWidget {
  const ArtistDetailScreen({super.key, required this.id});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final artistState = ref.watch(_artistDetailProvider(id));
    final albumsState = ref.watch(_artistAlbumsProvider(id));
    final tracksState = ref.watch(_artistTracksProvider(id));

    final artistName = artistState.valueOrNull?.name ?? 'Artist';

    return Scaffold(
      appBar: AppBar(title: Text(artistName)),
      body: CustomScrollView(
        slivers: [
          // Albums section
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Text('Albums', style: Theme.of(context).textTheme.titleLarge),
            ),
          ),
          albumsState.when(
            loading: () => const SliverToBoxAdapter(
              child: SizedBox(height: 80, child: Center(child: CircularProgressIndicator())),
            ),
            error: (e, _) => SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Text('Error: $e', style: const TextStyle(color: NeonShrineColors.error)),
              ),
            ),
            data: (albums) => SliverToBoxAdapter(
              child: SizedBox(
                height: 160,
                child: albums.isEmpty
                    ? const Center(child: Text('No albums'))
                    : ListView.builder(
                        scrollDirection: Axis.horizontal,
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                        itemCount: albums.length,
                        itemBuilder: (context, i) {
                          final album = albums[i];
                          return GestureDetector(
                            onTap: () => context.go(AppRoutes.albumDetailPath(album.id)),
                            child: Container(
                              width: 120,
                              margin: const EdgeInsets.symmetric(horizontal: 4),
                              child: Column(
                                children: [
                                  Container(
                                    width: 100,
                                    height: 100,
                                    decoration: BoxDecoration(
                                      color: NeonShrineColors.surfaceContainer,
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: const Icon(Icons.album, color: NeonShrineColors.outlineVariant, size: 40),
                                  ),
                                  const SizedBox(height: 6),
                                  Text(
                                    album.title,
                                    style: Theme.of(context).textTheme.labelMedium,
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                    textAlign: TextAlign.center,
                                  ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
              ),
            ),
          ),

          // Tracks section
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Text('Tracks', style: Theme.of(context).textTheme.titleLarge),
            ),
          ),
          tracksState.when(
            loading: () => const SliverToBoxAdapter(
              child: SizedBox(height: 80, child: Center(child: CircularProgressIndicator())),
            ),
            error: (e, _) => SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Text('Error: $e', style: const TextStyle(color: NeonShrineColors.error)),
              ),
            ),
            data: (tracks) => tracks.isEmpty
                ? const SliverToBoxAdapter(child: Center(child: Text('No tracks')))
                : SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, i) => _TrackTile(track: tracks[i]),
                      childCount: tracks.length,
                    ),
                  ),
          ),
        ],
      ),
    );
  }
}

class _TrackTile extends ConsumerStatefulWidget {
  const _TrackTile({required this.track});
  final CatalogTrack track;

  @override
  ConsumerState<_TrackTile> createState() => _TrackTileState();
}

class _TrackTileState extends ConsumerState<_TrackTile> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      ref
          .read(trackFavoriteProvider(widget.track.id).notifier)
          .init(widget.track.isFavorite ?? false);
    });
  }

  @override
  Widget build(BuildContext context) {
    final isFav = ref.watch(trackFavoriteProvider(widget.track.id));
    return TrackListTile(
      track: widget.track,
      isFavorite: isFav,
      onFavoriteTap: () =>
          ref.read(trackFavoriteProvider(widget.track.id).notifier).toggle(),
    );
  }
}
