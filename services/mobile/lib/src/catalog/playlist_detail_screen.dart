// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/playlist.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/shared/widgets/track_list_tile.dart';

final _playlistDetailProvider = FutureProvider.family<Playlist, String>((ref, id) {
  return ref.watch(catalogRepositoryProvider).getPlaylist(id);
});

// Loads each track by ID from the playlist's trackIds list
final _playlistTracksProvider = FutureProvider.family<List<CatalogTrack>, String>((ref, id) async {
  final repo = ref.watch(catalogRepositoryProvider);
  final playlist = await repo.getPlaylist(id);
  final futures = playlist.trackIds.map((tid) => repo.getTrack(tid));
  final results = await Future.wait(futures, eagerError: false);
  return results;
});

class PlaylistDetailScreen extends ConsumerWidget {
  const PlaylistDetailScreen({super.key, required this.id});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final playlistState = ref.watch(_playlistDetailProvider(id));
    final tracksState = ref.watch(_playlistTracksProvider(id));

    final playlistName = playlistState.valueOrNull?.name ?? 'Playlist';

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            pinned: true,
            expandedHeight: 180,
            flexibleSpace: FlexibleSpaceBar(
              title: Text(playlistName),
              background: Container(
                color: NeonShrineColors.surfaceContainer,
                child: const Icon(Icons.playlist_play, size: 80, color: NeonShrineColors.primaryViolet),
              ),
            ),
          ),
          playlistState.when(
            loading: () => const SliverToBoxAdapter(child: SizedBox()),
            error: (e, _) => SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Text('$e', style: const TextStyle(color: NeonShrineColors.error)),
              ),
            ),
            data: (playlist) => SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                child: Text(
                  '${playlist.trackIds.length} tracks',
                  style: Theme.of(context).textTheme.bodySmall,
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
                child: Text('Error loading tracks: $e',
                    style: const TextStyle(color: NeonShrineColors.error)),
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
