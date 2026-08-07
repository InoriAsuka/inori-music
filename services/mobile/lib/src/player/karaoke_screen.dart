import 'dart:ui' show ImageFilter;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/local_library/local_library_notifier.dart'
    show localTrackIdPrefix;
import 'package:inori_music/src/lyrics/karaoke_progress.dart';
import 'package:inori_music/src/lyrics/local_lyrics_provider.dart';
import 'package:inori_music/src/lyrics/lyric_line.dart';
import 'package:inori_music/src/lyrics/lyrics_background.dart';
import 'package:inori_music/src/lyrics/lyrics_provider.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Full-screen karaoke view with per-word progressive fill.
///
/// Mirrors the web `KaraokePanel`: the active line is shown large and fully
/// opaque, surrounding lines shrink and fade, and the word currently being
/// sung fills left-to-right via a [ShaderMask] gradient rather than flipping
/// colour in one step.
///
/// Lines without inline `<mm:ss.xx>` timings fall back to whole-line highlight.
class KaraokeScreen extends ConsumerStatefulWidget {
  const KaraokeScreen({super.key});

  @override
  ConsumerState<KaraokeScreen> createState() => _KaraokeScreenState();
}

class _KaraokeScreenState extends ConsumerState<KaraokeScreen> {
  final _scrollController = ScrollController();
  final _lineKeys = <int, GlobalKey>{};
  int _lastScrolledIndex = -1;

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _scrollToActive(int index) {
    if (index < 0 || index == _lastScrolledIndex) return;
    _lastScrolledIndex = index;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final ctx = _lineKeys[index]?.currentContext;
      if (ctx == null) return;
      Scrollable.ensureVisible(
        ctx,
        duration: const Duration(milliseconds: 350),
        curve: Curves.easeOut,
        alignment: 0.5,
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(playerProvider);
    final position = ref.watch(playerProvider.select((s) => s.position));
    final trackId = state.mediaItem?.id ?? '';
    final isLocal = trackId.startsWith(localTrackIdPrefix);
    final lyricsAsync = isLocal
        ? ref.watch(localLyricsProvider(trackId))
        : ref.watch(lyricsProvider(trackId));

    return Scaffold(
      body: LyricsBackground(
        albumId: state.mediaItem?.extras?['albumId'] as String?,
        localArtUri: state.mediaItem?.artUri,
        child: SafeArea(
          child: Column(
            children: [
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'KARAOKE',
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                              letterSpacing: 2,
                              color: context.skinColors.sakuraPink,
                            ),
                          ),
                          if (state.mediaItem != null)
                            Text(
                              state.mediaItem!.artist?.isNotEmpty == true
                                  ? '${state.mediaItem!.title} — ${state.mediaItem!.artist}'
                                  : state.mediaItem!.title,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: TextStyle(
                                fontSize: 12,
                                color: context.skinColors.onSurfaceVariant,
                              ),
                            ),
                        ],
                      ),
                    ),
                    IconButton(
                      icon: Icon(
                        Icons.close,
                        color: context.skinColors.onSurfaceVariant,
                      ),
                      tooltip: 'Close karaoke',
                      onPressed: () => Navigator.of(context).maybePop(),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: lyricsAsync.when(
                  loading: () => Center(
                    child: Text(
                      'Loading lyrics…',
                      style: TextStyle(
                        color: context.skinColors.onSurfaceVariant,
                      ),
                    ),
                  ),
                  error: (_, _) => Center(
                    child: Text(
                      'Could not load lyrics.',
                      style: TextStyle(
                        color: context.skinColors.onSurfaceVariant,
                      ),
                    ),
                  ),
                  data: (lines) {
                    if (lines == null || lines.isEmpty) {
                      return Center(
                        child: Text(
                          'No lyrics available.',
                          style: TextStyle(
                            color: context.skinColors.onSurfaceVariant,
                          ),
                        ),
                      );
                    }
                    final activeIndex = activeLineIndex(lines, position);
                    _scrollToActive(activeIndex);
                    final viewportPad =
                        MediaQuery.of(context).size.height * 0.35;
                    // Fades lines out as they approach the top and bottom
                    // edges instead of letting them clip against a hard
                    // boundary — the one piece of OriginalSound's lyrics
                    // treatment that translates directly, since its own
                    // implementation is the same gradient-into-dstIn trick.
                    return ShaderMask(
                      shaderCallback: (bounds) => const LinearGradient(
                        begin: Alignment.topCenter,
                        end: Alignment.bottomCenter,
                        colors: [
                          Colors.transparent,
                          Colors.black,
                          Colors.black,
                          Colors.transparent,
                        ],
                        stops: [0.0, 0.16, 0.84, 1.0],
                      ).createShader(bounds),
                      blendMode: BlendMode.dstIn,
                      child: ListView.builder(
                        controller: _scrollController,
                        padding: EdgeInsets.symmetric(
                          horizontal: 24,
                          vertical: viewportPad,
                        ),
                        itemCount: lines.length,
                        itemBuilder: (ctx, i) {
                          final key = _lineKeys.putIfAbsent(i, GlobalKey.new);
                          final active = i == activeIndex;
                          return Padding(
                            key: key,
                            padding: const EdgeInsets.symmetric(vertical: 14),
                            child: AnimatedOpacity(
                              duration: const Duration(milliseconds: 250),
                              opacity: active ? 1 : 0.4,
                              child: AnimatedScale(
                                duration: const Duration(milliseconds: 250),
                                scale: active ? 1 : 0.95,
                                child: _DepthBlur(
                                  distance: (i - activeIndex).abs(),
                                  child: _KaraokeLine(
                                    line: lines[i],
                                    active: active,
                                    position: position,
                                    nextLineStart: i + 1 < lines.length
                                        ? lines[i + 1].timestamp
                                        : null,
                                  ),
                                ),
                              ),
                            ),
                          );
                        },
                      ),
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
}

/// Softens a lyric line in proportion to how far it sits from the one being
/// sung, so the sung line reads as the focal plane.
///
/// A real gaussian blur, not a fake — but a deliberately shallow one, and
/// capped: every blurred line costs a `saveLayer`, and there are as many of
/// those as there are lines on screen. The blur is skipped outright for the
/// active line so the text the user is actually reading is never resampled.
class _DepthBlur extends StatelessWidget {
  const _DepthBlur({required this.distance, required this.child});

  /// Line distance from the active line, in rows.
  final int distance;
  final Widget child;

  static const _maxSigma = 2.0;
  static const _sigmaPerLine = 0.6;

  @override
  Widget build(BuildContext context) {
    if (distance == 0) return child;
    final sigma = (distance * _sigmaPerLine).clamp(0.0, _maxSigma);
    return ImageFiltered(
      imageFilter: ImageFilter.blur(sigmaX: sigma, sigmaY: sigma),
      child: child,
    );
  }
}

class _KaraokeLine extends StatelessWidget {
  const _KaraokeLine({
    required this.line,
    required this.active,
    required this.position,
    required this.nextLineStart,
  });

  final LyricLine line;
  final bool active;
  final Duration position;
  final Duration? nextLineStart;

  static const _style = TextStyle(
    fontSize: 28,
    fontWeight: FontWeight.w700,
    height: 1.35,
  );

  @override
  Widget build(BuildContext context) {
    final words = line.words;
    if (words == null || words.isEmpty) {
      return Text(
        line.text,
        textAlign: TextAlign.center,
        style: _style.copyWith(
          color: active
              ? context.skinColors.sakuraPink
              : context.skinColors.onSurfaceVariant,
        ),
      );
    }
    return RichText(
      textAlign: TextAlign.center,
      text: TextSpan(
        style: _style.copyWith(color: context.skinColors.onSurfaceVariant),
        children: [
          for (var i = 0; i < words.length; i++)
            WidgetSpan(
              child: _KaraokeWord(
                text: words[i].text,
                fill: active
                    ? wordProgress(words, i, position, nextLineStart)
                    : 0,
              ),
            ),
        ],
      ),
    );
  }
}

/// Paints the leading [fill] fraction of [text] in the primary colour.
class _KaraokeWord extends StatelessWidget {
  const _KaraokeWord({required this.text, required this.fill});

  final String text;
  final double fill;

  @override
  Widget build(BuildContext context) {
    if (fill <= 0) {
      return Text(
        text,
        style: _KaraokeLine._style.copyWith(
          color: context.skinColors.onSurfaceVariant,
        ),
      );
    }
    if (fill >= 1) {
      return Text(
        text,
        style: _KaraokeLine._style.copyWith(
          color: context.skinColors.sakuraPink,
        ),
      );
    }
    return ShaderMask(
      blendMode: BlendMode.srcIn,
      shaderCallback: (bounds) => LinearGradient(
        begin: Alignment.centerLeft,
        end: Alignment.centerRight,
        stops: [fill, fill],
        colors: [
          context.skinColors.sakuraPink,
          context.skinColors.onSurfaceVariant,
        ],
      ).createShader(bounds),
      child: Text(
        text,
        style: _KaraokeLine._style.copyWith(color: Colors.white),
      ),
    );
  }
}
