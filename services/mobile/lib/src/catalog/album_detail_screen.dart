// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/catalog_album.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/shared/widgets/track_list_tile.dart';

final _albumDetailProvider = FutureProvider.family<CatalogAlbum, String>((ref, id) {
  return ref.watch(catalogRepositoryProvider).getAlbum(id);
});

final _albumTracksProvider = FutureProvider.family<List<CatalogTrack>, String>((ref, id) {
  return ref.watch(catalogRepositoryProvider).tracksByAlbum(id);
});

class AlbumDetailScreen extends ConsumerWidget {
  const AlbumDetailScreen({super.key, required this.id});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final albumState = ref.watch(_albumDetailProvider(id));
    final tracksState = ref.watch(_albumTracksProvider(id));

    final albumTitle = albumState.valueOrNull?.title ?? 'Album';

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 200,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              title: Text(albumTitle),
              background: Container(
                color: NeonShrineColors.surfaceContainer,
                child: const Icon(Icons.album, size: 80, color: NeonShrineColors.outlineVariant),
              ),
            ),
          ),
          albumState.when(
            loading: () => const SliverToBoxAdapter(child: SizedBox()),
            error: (e, _) => SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Text('$e', style: const TextStyle(color: NeonShrineColors.error)),
              ),
            ),
            data: (album) => SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (album.releaseYear != null)
                      Text(
                        '${album.releaseYear}',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                  ],
                ),
              ),
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
            data: (tracks) => SliverList(
              delegate: SliverChildBuilderDelegate(
                (context, i) => TrackListTile(
                  track: tracks[i],
                  isFavorite: tracks[i].isFavorite ?? false,
                ),
                childCount: tracks.length,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
