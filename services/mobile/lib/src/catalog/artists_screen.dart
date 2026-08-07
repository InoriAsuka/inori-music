// ignore_for_file: implementation_imports, unnecessary_non_null_assertion
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_artist.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/desktop_app_bar.dart';

final _artistsProvider = FutureProvider<List<CatalogArtist>>((ref) {
  return ref.watch(catalogRepositoryProvider).listArtists(limit: 200);
});

/// artistId → one of that artist's album IDs, used purely as a stand-in cover
/// source: [CatalogArtist] carries no artwork field of its own (see the
/// generated model), and there is no artist-artwork endpoint to add one from.
///
/// Resolved with a single catalog-wide album list call rather than an
/// `albumsByArtist` request per grid cell — the grid holds up to 200 artists,
/// so per-cell resolution would mean 200 extra round trips for what is only a
/// thumbnail.
final _artistCoverAlbumIdsProvider = FutureProvider<Map<String, String>>((
  ref,
) async {
  final albums = await ref
      .watch(catalogRepositoryProvider)
      .listAlbums(limit: 500);
  final byArtist = <String, String>{};
  for (final album in albums) {
    byArtist.putIfAbsent(album.artistId, () => album.id);
  }
  return byArtist;
});

class ArtistsScreen extends ConsumerWidget {
  const ArtistsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final state = ref.watch(_artistsProvider);
    // Not awaited alongside the artist list: a failure here (or a slow
    // response) must only cost the thumbnails, never block the grid itself.
    final coverAlbumIds =
        ref.watch(_artistCoverAlbumIdsProvider).valueOrNull ?? const {};
    return Scaffold(
      appBar: DesktopAppBar(title: Text(t.artists)),
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
                onPressed: () => ref.refresh(_artistsProvider),
                child: Text(t.retry),
              ),
            ],
          ),
        ),
        data: (artists) => artists.isEmpty
            ? Center(child: Text(t.noData))
            : GridView.builder(
                padding: const EdgeInsets.all(12),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2,
                  crossAxisSpacing: 12,
                  mainAxisSpacing: 12,
                  childAspectRatio: 0.85,
                ),
                itemCount: artists.length,
                itemBuilder: (context, i) {
                  final artist = artists[i];
                  return _ArtistCard(
                    artist: artist,
                    coverAlbumId: coverAlbumIds[artist.id],
                    onTap: () =>
                        context.go(AppRoutes.artistDetailPath(artist.id)),
                  );
                },
              ),
      ),
    );
  }
}

class _ArtistCard extends ConsumerWidget {
  const _ArtistCard({
    required this.artist,
    required this.coverAlbumId,
    required this.onTap,
  });

  final CatalogArtist artist;

  /// Album whose cover stands in for this artist, or null when the artist has
  /// no albums (or the mapping hasn't resolved yet) — see
  /// [_artistCoverAlbumIdsProvider].
  final String? coverAlbumId;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final albumId = coverAlbumId;
    final artworkUrl = albumId == null
        ? null
        : ref.watch(artworkUrlProvider(albumId)).valueOrNull;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        decoration: BoxDecoration(
          color: context.skinColors.surfaceVariant,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: context.skinColors.outlineVariant,
            width: 0.5,
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Expanded(
              child: ClipRRect(
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(12),
                ),
                child: artworkUrl != null && artworkUrl.isNotEmpty
                    ? CachedNetworkImage(
                        imageUrl: artworkUrl,
                        fit: BoxFit.cover,
                        width: double.infinity,
                        placeholder: (context, _) => _placeholder(context),
                        errorWidget: (context, _, _) => _placeholder(context),
                      )
                    : _placeholder(context),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(8),
              child: Text(
                artist.name,
                style: Theme.of(context).textTheme.titleSmall,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                textAlign: TextAlign.center,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _placeholder(BuildContext context) => Container(
    width: double.infinity,
    color: context.skinColors.surfaceContainer,
    child: Icon(
      Icons.person,
      size: 56,
      color: context.skinColors.outlineVariant,
    ),
  );
}
