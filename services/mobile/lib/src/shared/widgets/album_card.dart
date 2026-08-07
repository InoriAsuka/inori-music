// ignore_for_file: implementation_imports
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/catalog_album.dart';

import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Album cover-art grid card — shared by the Albums browse grid and the
/// artist detail page's albums section, which previously each hand-rolled
/// their own version (one with only a placeholder icon, one a smaller
/// one-off horizontal-carousel variant, neither showing real artwork even
/// though [artworkUrlProvider] was already available).
///
/// Sizing is left to the caller (grid cell constraints / an outer SizedBox
/// in a horizontal list) — this only lays out its own internal column.
class AlbumCard extends ConsumerWidget {
  const AlbumCard({super.key, required this.album, required this.onTap});

  final CatalogAlbum album;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final artworkUrl = ref.watch(artworkUrlProvider(album.id)).valueOrNull;

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
          crossAxisAlignment: CrossAxisAlignment.start,
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
              padding: const EdgeInsets.fromLTRB(8, 8, 8, 2),
              child: Text(
                album.title,
                style: Theme.of(context).textTheme.titleSmall,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(8, 0, 8, 8),
              child: Text(
                album.releaseYear?.toString() ?? '',
                style: Theme.of(context).textTheme.bodySmall,
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
      Icons.album,
      size: 56,
      color: context.skinColors.outlineVariant,
    ),
  );
}
