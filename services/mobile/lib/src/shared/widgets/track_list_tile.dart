// ignore_for_file: implementation_imports
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

class TrackListTile extends ConsumerWidget {
  const TrackListTile({
    super.key,
    required this.track,
    this.isFavorite = false,
    this.onFavoriteTap,
    this.onTap,
    this.artworkUrl,
  });

  final CatalogTrack track;
  final bool isFavorite;
  final VoidCallback? onFavoriteTap;
  final VoidCallback? onTap;

  /// Optional artwork URL. When provided, a thumbnail is shown instead of
  /// the track number / music-note icon.
  final String? artworkUrl;

  static String _formatDurationMs(int? ms) {
    if (ms == null) return '';
    final totalSec = ms ~/ 1000;
    final m = totalSec ~/ 60;
    final s = totalSec % 60;
    return '$m:${s.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final durationStr = _formatDurationMs(track.durationMs);

    Widget leading;
    if (artworkUrl != null && artworkUrl!.isNotEmpty) {
      leading = ClipRRect(
        borderRadius: BorderRadius.circular(4),
        child: CachedNetworkImage(
          imageUrl: artworkUrl!,
          width: 40,
          height: 40,
          fit: BoxFit.cover,
          placeholder: (ctx, url) => const _ArtworkPlaceholder(),
          errorWidget: (ctx, url, err) => const _ArtworkPlaceholder(),
        ),
      );
    } else {
      leading = SizedBox(
        width: 40,
        child: Center(
          child: track.trackNumber != null
              ? Text(
                  '${track.trackNumber}',
                  style: const TextStyle(
                    color: NeonShrineColors.onSurfaceVariant,
                    fontSize: 13,
                  ),
                )
              : const Icon(
                  Icons.music_note,
                  size: 18,
                  color: NeonShrineColors.onSurfaceVariant,
                ),
        ),
      );
    }

    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 2),
      leading: leading,
      title: Text(
        track.title,
        style: const TextStyle(
          color: NeonShrineColors.onSurface,
          fontSize: 14,
          fontWeight: FontWeight.w500,
        ),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: track.genre != null
          ? Text(
              track.genre!,
              style: const TextStyle(
                color: NeonShrineColors.onSurfaceVariant,
                fontSize: 12,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            )
          : null,
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (durationStr.isNotEmpty)
            Text(
              durationStr,
              style: const TextStyle(
                color: NeonShrineColors.onSurfaceVariant,
                fontSize: 12,
              ),
            ),
          const SizedBox(width: 8),
          GestureDetector(
            onTap: onFavoriteTap,
            child: Icon(
              isFavorite ? Icons.favorite : Icons.favorite_border,
              size: 20,
              color: isFavorite
                  ? NeonShrineColors.accentPink
                  : NeonShrineColors.onSurfaceVariant,
            ),
          ),
        ],
      ),
      onTap: onTap ??
          () {
            ref.read(playerProvider.notifier).playTrack(track.id);
          },
    );
  }
}

class _ArtworkPlaceholder extends StatelessWidget {
  const _ArtworkPlaceholder();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 40,
      height: 40,
      color: NeonShrineColors.surfaceContainer,
      child: const Icon(Icons.music_note, size: 18, color: NeonShrineColors.onSurfaceVariant),
    );
  }
}
