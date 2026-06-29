// ignore_for_file: unnecessary_non_null_assertion
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/audio/sleep_timer_notifier.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

/// Persistent mini-player bar displayed at the bottom of the shell scaffold.
class MiniPlayerBar extends ConsumerWidget {
  const MiniPlayerBar({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final playerState = ref.watch(playerProvider);
    final mediaItem = playerState.mediaItem;
    final isPlaying = playerState.isPlaying;
    final isBuffering = playerState.isBuffering;
    final t = AppLocalizations.of(context);

    final title = mediaItem?.title ?? t.nothingPlaying;
    final artist = mediaItem?.artist ?? '';

    return Material(
      color: NeonShrineColors.playerBar,
      elevation: 8,
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
          child: InkWell(
            onTap: () => context.push(AppRoutes.player),
            borderRadius: BorderRadius.circular(8),
            child: Row(
              children: [
                // Artwork
                _MiniPlayerArtwork(
                  albumId: mediaItem?.extras?['albumId'] as String?,
                ),
                const SizedBox(width: 12),

                // Title / artist
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title,
                        style: const TextStyle(
                          color: NeonShrineColors.onSurface,
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      if (artist.isNotEmpty)
                        Text(
                          artist,
                          style: const TextStyle(
                            color: NeonShrineColors.onSurfaceVariant,
                            fontSize: 11,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                    ],
                  ),
                ),

                // Previous
                IconButton(
                  icon: const Icon(Icons.skip_previous, size: 24),
                  color: NeonShrineColors.onSurfaceVariant,
                  onPressed: () => ref.read(playerProvider.notifier).previous(),
                  tooltip: 'Previous',
                ),

                // Play / Pause
                IconButton(
                  icon: Icon(
                    isPlaying ? Icons.pause_rounded : Icons.play_arrow_rounded,
                    size: 28,
                    color: NeonShrineColors.onBackground,
                  ),
                  tooltip: isPlaying ? 'Pause' : 'Play',
                  onPressed: isBuffering ? null : () => ref.read(playerProvider.notifier).togglePlayPause(),
                ),

                // Next
                IconButton(
                  icon: const Icon(Icons.skip_next, size: 24),
                  color: NeonShrineColors.onSurfaceVariant,
                  onPressed: () => ref.read(playerProvider.notifier).next(),
                  tooltip: 'Next',
                ),

                // Sleep timer
                Consumer(
                  builder: (context, ref, _) {
                    final timerState = ref.watch(sleepTimerProvider);
                    final active = timerState.active;
                    final remaining = timerState.remaining;
                    final label = timerState.stopAfterTrack
                        ? '♪'
                        : (remaining != null
                            ? _formatRemaining(remaining)
                            : null);
                    return active
                        ? TextButton.icon(
                            icon: Icon(Icons.alarm,
                                size: 18,
                                color:
                                    Theme.of(context).colorScheme.primary),
                            label: Text(
                              label ?? '',
                              style: TextStyle(
                                  fontSize: 10,
                                  color: Theme.of(context)
                                      .colorScheme
                                      .primary),
                            ),
                            onPressed: () =>
                                _showSleepTimerSheet(context, ref),
                          )
                        : IconButton(
                            icon: const Icon(Icons.alarm, size: 22),
                            color: NeonShrineColors.onSurfaceVariant,
                            tooltip: 'Sleep timer',
                            onPressed: () =>
                                _showSleepTimerSheet(context, ref),
                          );
                  },
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  static String _formatRemaining(Duration d) {
    final m = d.inMinutes.remainder(60).toString().padLeft(2, '0');
    final s = d.inSeconds.remainder(60).toString().padLeft(2, '0');
    return '$m:$s';
  }
}

void _showSleepTimerSheet(BuildContext context, WidgetRef ref) {
  final timerActive = ref.read(sleepTimerProvider).active;
  showModalBottomSheet<void>(
    context: context,
    builder: (_) => SafeArea(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Padding(
            padding: EdgeInsets.all(16),
            child: Text('睡眠定时器',
                style: TextStyle(
                    fontSize: 16, fontWeight: FontWeight.bold)),
          ),
          for (final mins in [15, 30, 45, 60])
            ListTile(
              title: Text('$mins 分钟'),
              onTap: () {
                ref
                    .read(sleepTimerProvider.notifier)
                    .startFixed(Duration(minutes: mins));
                Navigator.pop(context);
              },
            ),
          ListTile(
            title: const Text('当前曲目结束后停止'),
            onTap: () {
              ref.read(sleepTimerProvider.notifier).startAfterTrack();
              Navigator.pop(context);
            },
          ),
          if (timerActive)
            ListTile(
              title: const Text('取消定时器',
                  style: TextStyle(color: Colors.red)),
              onTap: () {
                ref.read(sleepTimerProvider.notifier).cancel();
                Navigator.pop(context);
              },
            ),
        ],
      ),
    ),
  );
}

/// Mini player artwork thumbnail — shows the album cover or a fallback icon.
class _MiniPlayerArtwork extends ConsumerWidget {
  const _MiniPlayerArtwork({this.albumId});

  final String? albumId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final artworkAsync = albumId != null && albumId!.isNotEmpty
        ? ref.watch(artworkUrlProvider(albumId!))
        : null;

    Widget child;
    if (artworkAsync == null) {
      child = const Icon(Icons.music_note, color: NeonShrineColors.onSurfaceVariant, size: 22);
    } else {
      child = artworkAsync.when(
        data: (url) {
          if (url == null || url.isEmpty) {
            return const Icon(Icons.music_note, color: NeonShrineColors.onSurfaceVariant, size: 22);
          }
          return CachedNetworkImage(
            imageUrl: url,
            width: 44,
            height: 44,
            fit: BoxFit.cover,
            placeholder: (context, _) =>
                const Icon(Icons.music_note, color: NeonShrineColors.onSurfaceVariant, size: 22),
            errorWidget: (context, _, error) =>
                const Icon(Icons.music_note, color: NeonShrineColors.onSurfaceVariant, size: 22),
          );
        },
        loading: () => const Icon(Icons.music_note, color: NeonShrineColors.onSurfaceVariant, size: 22),
        error: (error, _) => const Icon(Icons.music_note, color: NeonShrineColors.onSurfaceVariant, size: 22),
      );
    }

    return Container(
      width: 44,
      height: 44,
      decoration: BoxDecoration(
        color: NeonShrineColors.surfaceContainer,
        borderRadius: BorderRadius.circular(6),
      ),
      clipBehavior: Clip.antiAlias,
      child: child,
    );
  }
}
