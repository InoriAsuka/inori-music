import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/audio/speed_notifier.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/favorites/track_favorite_notifier.dart';
import 'package:inori_music/src/lyrics/lyric_line.dart';
import 'package:inori_music/src/lyrics/lyrics_provider.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as ps;
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

/// Full-screen player overlay with progress bar, controls, and queue sheet.
class FullPlayerScreen extends ConsumerStatefulWidget {
  const FullPlayerScreen({super.key});

  @override
  ConsumerState<FullPlayerScreen> createState() => _FullPlayerScreenState();
}

class _FullPlayerScreenState extends ConsumerState<FullPlayerScreen> {
  late final PageController _pageController;
  int _pageIndex = 0;

  @override
  void initState() {
    super.initState();
    _pageController = PageController();
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(playerProvider);
    final isPlaying = state.isPlaying;
    final isBuffering = state.isBuffering;
    final trackId = state.mediaItem?.id ?? '';
    final position = ref.watch(playerProvider.select((s) => s.position));

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

            // Artwork / Lyrics PageView
            SizedBox(
              width: 280,
              height: 280,
              child: PageView(
                controller: _pageController,
                onPageChanged: (i) => setState(() => _pageIndex = i),
                children: [
                  // Page 0: Artwork
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
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(16),
                      child: _FullPlayerArtwork(
                        albumId: state.mediaItem?.extras?['albumId'] as String?,
                      ),
                    ),
                  ),
                  // Page 1: Lyrics
                  _LyricsPage(trackId: trackId, position: position),
                ],
              ),
            ),

            // Page indicator
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: List.generate(2, (i) {
                return Container(
                  margin: const EdgeInsets.symmetric(horizontal: 3),
                  width: _pageIndex == i ? 10 : 6,
                  height: 6,
                  decoration: BoxDecoration(
                    color: _pageIndex == i
                        ? NeonShrineColors.primaryViolet
                        : NeonShrineColors.onSurfaceVariant.withValues(alpha: 0.4),
                    borderRadius: BorderRadius.circular(3),
                  ),
                );
              }),
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
                  Consumer(
                    builder: (context2, ref2, child2) {
                      final isShuffle = ref2.watch(playerProvider).shuffle;
                      return IconButton(
                        icon: Icon(
                          Icons.shuffle,
                          color: isShuffle ? NeonShrineColors.primaryVioletLight : NeonShrineColors.onSurfaceVariant,
                        ),
                        onPressed: () => ref2.read(playerProvider.notifier).setShuffle(!isShuffle),
                        tooltip: 'Shuffle',
                      );
                    },
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
                  // Speed control button
                  Consumer(
                    builder: (context, ref, _) {
                      final speed = ref.watch(speedNotifierProvider);
                      return TextButton(
                        onPressed: () => _showSpeedSheet(context, ref),
                        child: Text('${speed}×', style: const TextStyle(fontSize: 14)),
                      );
                    },
                  ),
                  // Favorite button — wrapped in Consumer so icon and onPressed
                  // always use the same live trackId from the reactive ref.
                  Consumer(builder: (context2, ref2, child2) {
                    final trackId = ref2.watch(playerProvider).mediaItem?.id;
                    final isFav = trackId != null
                        ? ref2.watch(trackFavoriteProvider(trackId))
                        : false;
                    return IconButton(
                      icon: Icon(
                        isFav ? Icons.favorite : Icons.favorite_border,
                        color: isFav
                            ? NeonShrineColors.accentPink
                            : (trackId != null
                                ? NeonShrineColors.onSurface
                                : NeonShrineColors.onSurfaceVariant),
                      ),
                      onPressed: trackId == null
                          ? null
                          : () => ref2.read(trackFavoriteProvider(trackId).notifier).toggle(),
                      tooltip: 'Favorite',
                    );
                  }),
                ],
              ),
            ),

            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }

  void _showSpeedSheet(BuildContext context, WidgetRef ref) {
    const speeds = [0.5, 0.75, 1.0, 1.25, 1.5, 2.0];
    final current = ref.read(speedNotifierProvider);
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Padding(
              padding: EdgeInsets.all(16),
              child: Text('播放速度', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            ),
            for (final s in speeds)
              ListTile(
                title: Text('${s}×'),
                trailing: s == current ? const Icon(Icons.check) : null,
                onTap: () {
                  ref.read(speedNotifierProvider.notifier).setSpeed(s);
                  Navigator.pop(context);
                },
              ),
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

/// Large artwork widget for the full player screen.
/// Watches [artworkUrlProvider] for the album and shows CachedNetworkImage when
/// a URL is available; falls back to a music-note icon otherwise.
class _FullPlayerArtwork extends ConsumerWidget {
  const _FullPlayerArtwork({this.albumId});

  final String? albumId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (albumId == null || albumId!.isEmpty) {
      return const _ArtworkFallback();
    }
    final artworkAsync = ref.watch(artworkUrlProvider(albumId!));
    return artworkAsync.when(
      data: (url) {
        if (url == null || url.isEmpty) return const _ArtworkFallback();
        return CachedNetworkImage(
          imageUrl: url,
          width: 280,
          height: 280,
          fit: BoxFit.cover,
          placeholder: (context, _) => const _ArtworkFallback(),
          errorWidget: (context, _, error) => const _ArtworkFallback(),
        );
      },
      loading: () => const _ArtworkFallback(),
      error: (error, _) => const _ArtworkFallback(),
    );
  }
}

class _ArtworkFallback extends StatelessWidget {
  const _ArtworkFallback();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Icon(
        Icons.music_note_rounded,
        size: 80,
        color: NeonShrineColors.primaryViolet,
      ),
    );
  }
}

/// Lyrics page widget shown in the second page of the FullPlayerScreen PageView.
class _LyricsPage extends ConsumerWidget {
  const _LyricsPage({required this.trackId, required this.position});

  final String trackId;
  final Duration position;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (trackId.isEmpty) {
      return Center(
        child: Text(
          '暂无歌词',
          style: TextStyle(
            fontSize: 15,
            color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.5),
          ),
        ),
      );
    }
    final lyricsAsync = ref.watch(lyricsProvider(trackId));
    if (lyricsAsync.isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    final lines = lyricsAsync.valueOrNull;
    if (lines == null || lines.isEmpty) {
      return Center(
        child: Text(
          '暂无歌词',
          style: TextStyle(
            fontSize: 15,
            color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.5),
          ),
        ),
      );
    }
    final currentIndex =
        lines.lastIndexWhere((l) => l.timestamp <= position);
    return Container(
      decoration: BoxDecoration(
        color: NeonShrineColors.surfaceVariant,
        borderRadius: BorderRadius.circular(16),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: _LyricsList(
          lines: lines,
          currentIndex: currentIndex,
        ),
      ),
    );
  }
}

class _LyricsList extends StatefulWidget {
  const _LyricsList({required this.lines, required this.currentIndex});
  final List<LyricLine> lines;
  final int currentIndex;

  @override
  State<_LyricsList> createState() => _LyricsListState();
}

class _LyricsListState extends State<_LyricsList> {
  final ScrollController _scrollController = ScrollController();

  @override
  void didUpdateWidget(_LyricsList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.currentIndex != widget.currentIndex &&
        widget.currentIndex >= 0) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!_scrollController.hasClients) return;
        final itemHeight = 48.0;
        final offset = (widget.currentIndex * itemHeight -
                _scrollController.position.viewportDimension / 2 +
                itemHeight / 2)
            .clamp(
          _scrollController.position.minScrollExtent,
          _scrollController.position.maxScrollExtent,
        );
        _scrollController.animateTo(
          offset,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
        );
      });
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
      itemCount: widget.lines.length,
      itemBuilder: (context, i) {
        final isCurrent = i == widget.currentIndex;
        return SizedBox(
          height: 48,
          child: Center(
            child: Text(
              widget.lines[i].text,
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: isCurrent ? 18 : 15,
                fontWeight:
                    isCurrent ? FontWeight.w600 : FontWeight.normal,
                color: isCurrent
                    ? Theme.of(context).colorScheme.primary
                    : Theme.of(context)
                        .colorScheme
                        .onSurface
                        .withValues(alpha: 0.5),
              ),
            ),
          ),
        );
      },
    );
  }
}
