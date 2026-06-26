import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/favorites/track_favorite_notifier.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as ps;
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

/// Full-screen player overlay with progress bar, controls, and queue sheet.
class FullPlayerScreen extends ConsumerWidget {
  const FullPlayerScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(playerProvider);
    final isPlaying = state.isPlaying;
    final isBuffering = state.isBuffering;

    return Scaffold(
      backgroundColor: NeonShrineColors.background,
      body: SafeArea(
        child: Column(
          children: [
            // Top bar
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Row(
                children: [
                  IconButton(
                    icon: const Icon(Icons.keyboard_arrow_down, size: 32, color: NeonShrineColors.onBackground),
                    tooltip: 'Close player',
                    onPressed: () => Navigator.of(context).maybePop(),
                  ),
                  const Expanded(
                    child: Text(
                      'Now Playing',
                      textAlign: TextAlign.center,
                      style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: NeonShrineColors.onSurfaceVariant),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.queue_music, color: NeonShrineColors.onSurfaceVariant),
                    tooltip: 'Queue',
                    onPressed: () => _showQueueSheet(context, ref),
                  ),
                ],
              ),
            ),

            const Spacer(),

            // Large artwork placeholder
            Container(
              width: 280,
              height: 280,
              decoration: BoxDecoration(
                color: NeonShrineColors.surfaceVariant,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [
                  BoxShadow(
                    color: NeonShrineColors.primaryViolet.withValues(alpha: 0.15),
                    blurRadius: 32,
                    offset: const Offset(0, 8),
                  ),
                ],
              ),
              child: const Icon(Icons.music_note_rounded, size: 80, color: NeonShrineColors.primaryViolet),
            ),

            const Spacer(),

            // Title / artist
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Column(
                children: [
                  Text(
                    state.mediaItem?.title ?? 'Unknown Track',
                    style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: NeonShrineColors.onBackground),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 4),
                  Text(
                    state.mediaItem?.artist ?? '',
                    style: const TextStyle(fontSize: 15, color: NeonShrineColors.onSurfaceVariant),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.center,
                  ),
                ],
              ),
            ),

            const SizedBox(height: 24),

            // Seek bar
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Column(
                children: [
                  SliderTheme(
                    data: SliderTheme.of(context).copyWith(
                      trackHeight: 3,
                      thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 6),
                      overlayShape: const RoundSliderOverlayShape(overlayRadius: 14),
                    ),
                    child: Slider(
                      value: isBuffering
                          ? 0
                          : state.position.inMilliseconds.toDouble().clamp(
                                0,
                                state.duration.inMilliseconds.toDouble() > 0
                                    ? state.duration.inMilliseconds.toDouble()
                                    : 1,
                              ),
                      max: state.duration.inMilliseconds.toDouble() > 0
                          ? state.duration.inMilliseconds.toDouble()
                          : 1,
                      onChanged: isBuffering
                          ? null
                          : (v) => ref.read(playerProvider.notifier).seekTo(Duration(milliseconds: v.toInt())),
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(_formatDuration(state.position), style: const TextStyle(fontSize: 12, color: NeonShrineColors.onSurfaceVariant)),
                        Text(_formatDuration(state.duration), style: const TextStyle(fontSize: 12, color: NeonShrineColors.onSurfaceVariant)),
                      ],
                    ),
                  ),
                ],
              ),
            ),

            // Controls row
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  IconButton(
                    icon: Icon(
                      Icons.repeat,
                      color: state.repeat != ps.RepeatMode.none ? NeonShrineColors.primaryVioletLight : NeonShrineColors.onSurfaceVariant,
                    ),
                    onPressed: () {
                      final notifier = ref.read(playerProvider.notifier);
                      switch (state.repeat) {
                        case ps.RepeatMode.none:
                          notifier.setRepeat(ps.RepeatMode.all);
                          break;
                        case ps.RepeatMode.all:
                          notifier.setRepeat(ps.RepeatMode.one);
                          break;
                        case ps.RepeatMode.one:
                          notifier.setRepeat(ps.RepeatMode.none);
                          break;
                      }
                    },
                    tooltip: 'Repeat: ${state.repeat.name}',
                  ),
                  IconButton(
                    icon: const Icon(Icons.skip_previous, size: 36, color: NeonShrineColors.onSurface),
                    onPressed: () => ref.read(playerProvider.notifier).previous(),
                  ),
                  // Play / Pause button
                  Container(
                    decoration: const BoxDecoration(
                      color: NeonShrineColors.primaryViolet,
                      shape: BoxShape.circle,
                    ),
                    child: IconButton(
                      icon: Icon(
                        isBuffering ? Icons.play_arrow_rounded : (isPlaying ? Icons.pause_rounded : Icons.play_arrow_rounded),
                        size: 36,
                        color: Colors.white,
                      ),
                      onPressed: isBuffering ? null : () => ref.read(playerProvider.notifier).togglePlayPause(),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.skip_next, size: 36, color: NeonShrineColors.onSurface),
                    onPressed: () => ref.read(playerProvider.notifier).next(),
                  ),
                  IconButton(
                    icon: Consumer(builder: (context2, ref2, child2) {
                      final trackId = ref2.watch(playerProvider).mediaItem?.id;
                      if (trackId == null) {
                        return const Icon(Icons.favorite_border, color: NeonShrineColors.onSurfaceVariant);
                      }
                      final isFav = ref2.watch(trackFavoriteProvider(trackId));
                      return Icon(
                        isFav ? Icons.favorite : Icons.favorite_border,
                        color: isFav ? NeonShrineColors.accentPink : NeonShrineColors.onSurfaceVariant,
                      );
                    }),
                    onPressed: () {
                      final trackId = ref.read(playerProvider).mediaItem?.id;
                      if (trackId != null) {
                        ref.read(trackFavoriteProvider(trackId).notifier).toggle();
                      }
                    },
                    tooltip: 'Favorite',
                  ),
                ],
              ),
            ),

            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }

  void _showQueueSheet(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => DraggableScrollableSheet(
        initialChildSize: 0.6,
        maxChildSize: 0.9,
        minChildSize: 0.3,
        builder: (_, controller) => Container(
          decoration: const BoxDecoration(
            color: NeonShrineColors.surface,
            borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
          ),
          child: Column(
            children: [
              const SizedBox(height: 12),
              Container(
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: NeonShrineColors.outlineVariant,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const Padding(
                padding: EdgeInsets.all(12),
                child: Text(
                  'Queue',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: NeonShrineColors.onBackground),
                ),
              ),
              Expanded(
                child: Consumer(
                  builder: (context2, ref2, child2) {
                    final playerState = ref2.watch(playerProvider);
                    final queue = playerState.queue;
                    final currentIndex = playerState.currentIndex;
                    return ReorderableListView.builder(
                      scrollController: controller,
                      itemCount: queue.length,
                      onReorderItem: (oldIdx, newIdx) {
                        ref2.read(playerProvider.notifier).reorderQueue(oldIdx, newIdx);
                      },
                      itemBuilder: (_, i) {
                        final item = queue[i];
                        final isCurrent = i == currentIndex;
                        return ListTile(
                          key: ValueKey(item.id),
                          leading: Icon(
                            Icons.music_note,
                            color: isCurrent ? NeonShrineColors.primaryVioletLight : NeonShrineColors.onSurfaceVariant,
                          ),
                          title: Text(
                            item.title,
                            style: TextStyle(
                              color: isCurrent ? NeonShrineColors.primaryVioletLight : NeonShrineColors.onSurface,
                              fontWeight: isCurrent ? FontWeight.w600 : FontWeight.normal,
                            ),
                          ),
                          subtitle: Text(item.artist ?? '', style: const TextStyle(color: NeonShrineColors.onSurfaceVariant)),
                          trailing: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              if (isCurrent && playerState.isPlaying)
                                const Icon(Icons.equalizer, color: NeonShrineColors.primaryVioletLight, size: 20),
                              const Icon(Icons.drag_handle, color: NeonShrineColors.onSurfaceVariant, size: 20),
                            ],
                          ),
                          onTap: () {
                            ref2.read(playerProvider.notifier).playQueue(
                              queue.map((m) => m.id).toList(),
                              initialIndex: i,
                            );
                          },
                        );
                      },
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  static String _formatDuration(Duration d) {
    final mins = d.inMinutes;
    final secs = d.inSeconds % 60;
    return '$mins:${secs.toString().padLeft(2, '0')}';
  }
}
