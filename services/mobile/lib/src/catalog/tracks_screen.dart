// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/shared/widgets/track_list_tile.dart';

final _tracksProvider = FutureProvider<List<CatalogTrack>>((ref) {
  return ref.watch(catalogRepositoryProvider).listTracks(limit: 500);
});

class TracksScreen extends ConsumerWidget {
  const TracksScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(_tracksProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Tracks')),
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
                onPressed: () => ref.refresh(_tracksProvider),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (tracks) => tracks.isEmpty
            ? const Center(child: Text('No tracks found'))
            : ListView.builder(
                itemCount: tracks.length,
                itemBuilder: (context, i) => TrackListTile(
                  track: tracks[i],
                  isFavorite: tracks[i].isFavorite ?? false,
                ),
              ),
      ),
    );
  }
}
