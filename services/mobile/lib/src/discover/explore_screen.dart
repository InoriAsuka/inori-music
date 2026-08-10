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

/// "最近添加" — a real, already-existing server capability
/// (`sortBy=createdAt&sortOrder=desc`, confirmed against
/// `handler_catalog.go`'s `AlbumSortByCreatedAt`/`sortCatalogAlbums`), not a
/// speculative endpoint invented for this screen.
final recentlyAddedAlbumsProvider = FutureProvider<List<CatalogAlbum>>((ref) {
  return ref
      .watch(catalogRepositoryProvider)
      .listAlbums(limit: 20, sortBy: 'createdAt', sortOrder: 'desc');
});

/// "随机发现" — the catalog API has no server-side random-sample endpoint
/// (confirmed: `listCatalogAlbums` only takes `limit`/`offset`/`sortBy`/
/// `sortOrder`/`releaseYearMin`/`releaseYearMax`, no shuffle/seed
/// parameter), so this fetches one ordinary page and shuffles it
/// client-side rather than inventing a server capability that doesn't
/// exist. `FutureProvider` caches the result, so the shuffle only runs once
/// per screen visit rather than re-randomizing on every rebuild.
final randomAlbumsProvider = FutureProvider<List<CatalogAlbum>>((ref) async {
  final albums = await ref
      .watch(catalogRepositoryProvider)
      .listAlbums(limit: 60);
  final shuffled = List<CatalogAlbum>.of(albums)..shuffle();
  return shuffled.take(12).toList();
});

/// "探索发现" — the sidebar's "发现音乐" group's other destination (see
/// `for_you_screen.dart`'s doc comment for the contrast): this one has real
/// catalog data behind it, so it renders real content instead of a guiding
/// empty state.
class ExploreScreen extends ConsumerWidget {
  const ExploreScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    return Scaffold(
      appBar: DesktopAppBar(title: Text(t.sidebarExplore)),
      body: ListView(
        padding: const EdgeInsets.symmetric(vertical: 16),
        children: [
          _AlbumRailSection(
            title: t.recentlyAdded,
            provider: recentlyAddedAlbumsProvider,
          ),
          const SizedBox(height: 24),
          _AlbumRailSection(
            title: t.randomPicks,
            provider: randomAlbumsProvider,
          ),
        ],
      ),
    );
  }
}

class _AlbumRailSection extends ConsumerWidget {
  const _AlbumRailSection({required this.title, required this.provider});

  final String title;
  final FutureProvider<List<CatalogAlbum>> provider;

  static const _cardWidth = 150.0;
  static const _cardHeight = 220.0;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(provider);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          child: Text(
            title,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w700,
              color: context.skinColors.onSurface,
            ),
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: _cardHeight,
          child: state.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Center(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Text(
                  '$e',
                  style: TextStyle(color: context.skinColors.onSurfaceVariant),
                ),
              ),
            ),
            data: (albums) => albums.isEmpty
                ? Center(
                    child: Text(
                      AppLocalizations.of(context).noData,
                      style: TextStyle(
                        color: context.skinColors.onSurfaceVariant,
                      ),
                    ),
                  )
                : ListView.separated(
                    scrollDirection: Axis.horizontal,
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    itemCount: albums.length,
                    separatorBuilder: (_, _) => const SizedBox(width: 12),
                    itemBuilder: (context, i) {
                      final album = albums[i];
                      return SizedBox(
                        width: _cardWidth,
                        child: AlbumCard(
                          album: album,
                          onTap: () =>
                              context.go(AppRoutes.albumDetailPath(album.id)),
                        ),
                      );
                    },
                  ),
          ),
        ),
      ],
    );
  }
}
