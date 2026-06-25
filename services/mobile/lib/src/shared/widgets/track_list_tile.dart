// ignore_for_file: implementation_imports
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
  });

  final CatalogTrack track;
  final bool isFavorite;
  final VoidCallback? onFavoriteTap;
  final VoidCallback? onTap;

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

    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 2),
      leading: SizedBox(
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
      ),
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
