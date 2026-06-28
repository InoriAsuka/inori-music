// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/playlist.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

final _playlistsProvider = FutureProvider<List<Playlist>>((ref) {
  return ref.watch(catalogRepositoryProvider).listPlaylists();
});

class PlaylistsScreen extends ConsumerWidget {
  const PlaylistsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final state = ref.watch(_playlistsProvider);
    return Scaffold(
      appBar: AppBar(title: Text(t.playlists)),
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
                onPressed: () => ref.refresh(_playlistsProvider),
                child: Text(t.retry),
              ),
            ],
          ),
        ),
        data: (playlists) => playlists.isEmpty
            ? Center(child: Text(t.noData))
            : ListView.builder(
                itemCount: playlists.length,
                itemBuilder: (context, i) {
                  final pl = playlists[i];
                  return ListTile(
                    leading: Container(
                      width: 44,
                      height: 44,
                      decoration: BoxDecoration(
                        color: NeonShrineColors.surfaceContainer,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: const Icon(Icons.playlist_play, color: NeonShrineColors.primaryViolet),
                    ),
                    title: Text(pl.name),
                    subtitle: Text('${pl.trackIds.length} tracks'),
                    onTap: () => context.go(AppRoutes.playlistDetailPath(pl.id)),
                  );
                },
              ),
      ),
    );
  }
}
