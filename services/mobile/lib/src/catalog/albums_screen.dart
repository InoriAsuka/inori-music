// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_album.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/album_card.dart';
import 'package:inori_music/src/shared/widgets/desktop_app_bar.dart';

final _albumsProvider = FutureProvider<List<CatalogAlbum>>((ref) {
  return ref.watch(catalogRepositoryProvider).listAlbums(limit: 200);
});

class AlbumsScreen extends ConsumerWidget {
  const AlbumsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final state = ref.watch(_albumsProvider);
    return Scaffold(
      appBar: DesktopAppBar(title: Text(t.albums)),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.error_outline,
                color: context.skinColors.error,
                size: 48,
              ),
              const SizedBox(height: 12),
              Text('$e', textAlign: TextAlign.center),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: () => ref.refresh(_albumsProvider),
                child: Text(t.retry),
              ),
            ],
          ),
        ),
        data: (albums) => albums.isEmpty
            ? Center(child: Text(t.noData))
            : GridView.builder(
                padding: const EdgeInsets.all(12),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2,
                  crossAxisSpacing: 12,
                  mainAxisSpacing: 12,
                  childAspectRatio: 0.78,
                ),
                itemCount: albums.length,
                itemBuilder: (context, i) {
                  final album = albums[i];
                  return AlbumCard(
                    album: album,
                    onTap: () =>
                        context.go(AppRoutes.albumDetailPath(album.id)),
                  );
                },
              ),
      ),
    );
  }
}
