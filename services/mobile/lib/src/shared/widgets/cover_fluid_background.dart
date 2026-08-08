import 'dart:math' as math;
import 'dart:ui' show ImageFilter;

import 'package:flutter/material.dart';

/// Slow-drifting colour field generated from the current cover art — the
/// backdrop EchoMusic uses behind its player and lyrics pages, ported to
/// Flutter.
///
/// The technique is not a gradient between two extracted colours. EchoMusic
/// (`src/renderer/views/lyric/LyricFluidBackground.vue`) draws the cover's
/// four quadrants into four small canvases, scatters them around the centre,
/// spins each one slowly while counter-spinning the group, boosts saturation
/// and brightness, then buries the whole thing under a heavy backdrop blur.
/// What survives is the cover's own colour distribution, moving — which is
/// why it always agrees with the artwork instead of approximating it with a
/// two-stop gradient.
///
/// Deliberately omitted from the port: EchoMusic additionally warps the group
/// through an SVG `feTurbulence` + `feDisplacementMap` filter. Flutter has no
/// equivalent short of a fragment shader, and the counter-rotation plus the
/// blur already carry the effect.
class CoverFluidBackground extends StatefulWidget {
  const CoverFluidBackground({
    super.key,
    required this.image,
    required this.fallbackColor,
    required this.child,
  });

  /// Cover art to derive the field from. Null (no artwork, or it failed to
  /// load) renders [fallbackColor] flat — this is decoration, and it must
  /// degrade to something plain rather than to nothing.
  final ImageProvider? image;

  final Color fallbackColor;
  final Widget child;

  /// Side of each quadrant tile relative to the larger viewport dimension.
  /// 0.707 ≈ 1/√2, EchoMusic's value: at this size a tile's corners never
  /// swing inside its own inscribed circle as it rotates, so no gap opens up.
  static const _tileFactor = 0.707;

  /// How far each tile's centre sits from the viewport centre, as a fraction
  /// of the tile's own side.
  static const _spread = 0.35;

  @override
  State<CoverFluidBackground> createState() => _CoverFluidBackgroundState();
}

class _CoverFluidBackgroundState extends State<CoverFluidBackground>
    with TickerProviderStateMixin {
  // Two loops at very different periods, turning opposite ways. Neither is
  // fast enough to read as "spinning"; together they read as the field slowly
  // reorganising itself.
  //
  // Built in initState rather than lazily: a `late final` initialiser that
  // only the has-artwork branch of build() touches never runs when there is
  // no cover, and dispose() then *creates* the controllers on its way out —
  // which means constructing a Ticker against an already-deactivated element.
  late final AnimationController _tiles;
  late final AnimationController _group;

  @override
  void initState() {
    super.initState();
    _tiles = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 60),
    );
    _group = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 150),
    );
    _syncAnimations();
  }

  @override
  void didUpdateWidget(CoverFluidBackground oldWidget) {
    super.didUpdateWidget(oldWidget);
    if ((oldWidget.image == null) != (widget.image == null)) _syncAnimations();
  }

  /// Nothing rotates when there is no cover, so nothing should be ticking
  /// either — an idle screen must not hold a vsync callback open for a
  /// backdrop it isn't drawing.
  void _syncAnimations() {
    if (widget.image != null) {
      if (!_tiles.isAnimating) _tiles.repeat();
      if (!_group.isAnimating) _group.repeat();
    } else {
      _tiles.stop();
      _group.stop();
    }
  }

  /// saturate(1.3) then brightness(1.5), as one 5x4 matrix. Both are
  /// EchoMusic's values, and the order matters — lifting brightness after the
  /// blur instead would wash the field out rather than make it glow.
  static const _boost = ColorFilter.matrix(<double>[
    1.85433, -0.32184, -0.03249, 0, 0, //
    -0.09567, 1.62816, -0.03249, 0, 0, //
    -0.09567, -0.32184, 1.91751, 0, 0, //
    0, 0, 0, 1, 0, //
  ]);

  @override
  void dispose() {
    _tiles.dispose();
    _group.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final image = widget.image;

    return Stack(
      fit: StackFit.expand,
      children: [
        ColoredBox(color: widget.fallbackColor),
        if (image != null) ...[
          // Isolated so the two always-running controllers repaint only this
          // subtree, never the lyrics and controls stacked above it.
          RepaintBoundary(
            child: ClipRect(
              child: LayoutBuilder(
                builder: (context, constraints) {
                  final size = constraints.biggest;
                  final side =
                      size.longestSide * CoverFluidBackground._tileFactor;
                  return ColorFiltered(
                    colorFilter: _boost,
                    child: AnimatedBuilder(
                      animation: _group,
                      builder: (context, child) => Transform.rotate(
                        // Against the tiles.
                        angle: -_group.value * 2 * math.pi,
                        // Oversized so the group's own rotation can't swing an
                        // empty corner into the viewport.
                        child: Transform.scale(scale: 1.2, child: child),
                      ),
                      child: Stack(
                        children: [
                          for (var i = 0; i < 4; i++)
                            _positionedTile(i, size, side, image),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
          ),
          // The blur that turns four recognisable thumbnails into a colour
          // field, plus the scrim that keeps foreground text legible over it.
          Positioned.fill(
            child: IgnorePointer(
              child: BackdropFilter(
                filter: ImageFilter.blur(sigmaX: 64, sigmaY: 64),
                child: ColoredBox(color: Colors.black.withValues(alpha: 0.24)),
              ),
            ),
          ),
        ],
        widget.child,
      ],
    );
  }

  /// Mirrors EchoMusic's `canvasStyle()`: tile centres offset from the
  /// viewport centre by ±`_spread` of a tile side, one per quadrant.
  Widget _positionedTile(int i, Size size, double side, ImageProvider image) {
    final signX = i.isEven ? -1 : 1;
    final signY = i < 2 ? -1 : 1;
    return Positioned(
      left:
          size.width / 2 +
          signX * side * CoverFluidBackground._spread -
          side / 2,
      top:
          size.height / 2 +
          signY * side * CoverFluidBackground._spread -
          side / 2,
      width: side,
      height: side,
      child: _QuadrantTile(
        image: image,
        quadrant: i,
        side: side,
        spin: _tiles,
        // Staggered like EchoMusic's -5s/-10s/-15s animation-delays: in
        // lockstep the four tiles read as one rigid shape turning.
        phase: i * 0.0833,
      ),
    );
  }
}

class _QuadrantTile extends StatelessWidget {
  const _QuadrantTile({
    required this.image,
    required this.quadrant,
    required this.side,
    required this.spin,
    required this.phase,
  });

  final ImageProvider image;

  /// 0=top-left, 1=top-right, 2=bottom-left, 3=bottom-right.
  final int quadrant;
  final double side;
  final Animation<double> spin;
  final double phase;

  static const _alignments = [
    Alignment.topLeft,
    Alignment.topRight,
    Alignment.bottomLeft,
    Alignment.bottomRight,
  ];

  @override
  Widget build(BuildContext context) {
    final alignment = _alignments[quadrant];

    return AnimatedBuilder(
      animation: spin,
      builder: (context, child) => Transform.rotate(
        angle: (spin.value + phase) * 2 * math.pi,
        child: child,
      ),
      child: ClipRect(
        child: OverflowBox(
          // Twice the tile, pinned to a corner, then clipped — that shows
          // exactly one quadrant of the cover without having to decode it to
          // a dart:ui.Image and drawImageRect it by hand.
          alignment: alignment,
          maxWidth: side * 2,
          maxHeight: side * 2,
          child: Image(
            image: image,
            width: side * 2,
            height: side * 2,
            fit: BoxFit.cover,
            // Nothing here is ever read as an image, only as colour, so low
            // filter quality is free savings on every frame.
            filterQuality: FilterQuality.low,
            errorBuilder: (_, _, _) => const SizedBox.shrink(),
          ),
        ),
      ),
    );
  }
}
