import 'package:flutter/material.dart';

/// A rounded slider track whose *active* segment paints as a two-stop linear
/// gradient (the theme's [SliderThemeData.activeTrackColor] fading toward a
/// lightened version of itself) instead of Flutter's default flat fill.
///
/// Modelled on [RoundedRectSliderTrackShape] (Flutter's own default shape)
/// but deliberately simplified for how this app actually uses a [Slider]:
/// no secondary/buffered track (the [PlaybackEngine] interface this app is
/// built on exposes no buffered-position stream to paint one from — adding
/// one is a bigger change than a v5.30.7 visual-polish pass, see that
/// class's own doc comment on the decode/output split this project can't
/// fully make yet) and no RTL locale (only en/zh/ja ship, all LTR).
///
/// v5.30.7 field report: "进度条有点特效会更好" ("the progress bar could use
/// a bit of flair") — the flat single-colour bar read as flatter than the
/// rest of the player's frosted-glass, gradient-accented look elsewhere.
class GradientSliderTrackShape extends SliderTrackShape
    with BaseSliderTrackShape {
  const GradientSliderTrackShape({this.lightenAmount = 0.22});

  /// How much lighter the gradient's far (thumb) end is than
  /// [SliderThemeData.activeTrackColor] itself, in HSL lightness (0-1).
  final double lightenAmount;

  @override
  void paint(
    PaintingContext context,
    Offset offset, {
    required RenderBox parentBox,
    required SliderThemeData sliderTheme,
    required Animation<double> enableAnimation,
    required TextDirection textDirection,
    required Offset thumbCenter,
    Offset? secondaryOffset,
    bool isDiscrete = false,
    bool isEnabled = false,
    double additionalActiveTrackHeight = 2,
  }) {
    // Matches RoundedRectSliderTrackShape's own no-op guard: a track with no
    // height paints nothing regardless of colour.
    if (sliderTheme.trackHeight == null || sliderTheme.trackHeight! <= 0) {
      return;
    }

    final trackRect = getPreferredRect(
      parentBox: parentBox,
      offset: offset,
      sliderTheme: sliderTheme,
      isEnabled: isEnabled,
      isDiscrete: isDiscrete,
    );
    final trackRadius = Radius.circular(trackRect.height / 2);
    final activeTrackRadius = Radius.circular(
      (trackRect.height + additionalActiveTrackHeight) / 2,
    );

    final inactiveColor = (isEnabled
        ? sliderTheme.inactiveTrackColor
        : sliderTheme.disabledInactiveTrackColor)!;
    if (thumbCenter.dx < trackRect.right - (sliderTheme.trackHeight! / 2)) {
      context.canvas.drawRRect(
        RRect.fromLTRBR(
          thumbCenter.dx - (sliderTheme.trackHeight! / 2),
          trackRect.top,
          trackRect.right,
          trackRect.bottom,
          trackRadius,
        ),
        Paint()..color = inactiveColor,
      );
    }

    if (thumbCenter.dx > trackRect.left + (sliderTheme.trackHeight! / 2)) {
      final activeRect = Rect.fromLTRB(
        trackRect.left,
        trackRect.top - (additionalActiveTrackHeight / 2),
        thumbCenter.dx + (sliderTheme.trackHeight! / 2),
        trackRect.bottom + (additionalActiveTrackHeight / 2),
      );
      final baseColor = (isEnabled
          ? sliderTheme.activeTrackColor
          : sliderTheme.disabledActiveTrackColor)!;
      final gradientPaint = Paint()
        ..shader = LinearGradient(
          colors: [baseColor, _lighten(baseColor, lightenAmount)],
        ).createShader(activeRect);
      context.canvas.drawRRect(
        RRect.fromLTRBR(
          activeRect.left,
          activeRect.top,
          activeRect.right,
          activeRect.bottom,
          activeTrackRadius,
        ),
        gradientPaint,
      );
    }
  }

  @override
  bool get isRounded => true;
}

Color _lighten(Color color, double amount) {
  final hsl = HSLColor.fromColor(color);
  return hsl.withLightness((hsl.lightness + amount).clamp(0.0, 1.0)).toColor();
}

/// A round slider thumb that optionally paints a soft, blurred halo behind
/// itself — the "still playing" ambient glow the v5.30.7 field report asked
/// for ("进度条有点特效会更好"), gated by [glowing] rather than always on so
/// it collapses away the moment playback pauses.
///
/// Static, not pulsing, and deliberately not built on a repeating
/// [AnimationController]: it is repainted fresh every time this shape's
/// [paint] runs, which already happens on every position tick (the seek
/// row rebuilds continuously while playing regardless of this effect). A
/// pulsing glow would need its own Ticker to keep running independently of
/// that — exactly the kind of always-on animation
/// `cover_fluid_background.dart`'s own doc comment warns costs real CPU if
/// left running while paused, and exactly the kind of controller lifecycle
/// that same file's `late final` mistake (constructed lazily, then disposed
/// against an already-defunct element) came from. Skipping a controller
/// entirely for this effect sidesteps both problems rather than needing to
/// carefully avoid them.
class GlowingSliderThumbShape extends SliderComponentShape {
  const GlowingSliderThumbShape({
    required this.radius,
    required this.maxRadius,
    required this.glowing,
    required this.glowColor,
  });

  /// Radius to actually *paint* this frame. Matches RoundSliderThumbShape's
  /// own contract — 0 means "draw nothing" — and is what the mini bar's wide
  /// seek row varies between 0 (at rest) and its revealed size (on hover or
  /// drag). The full player's seek bar just passes a constant here, since its
  /// thumb never collapses.
  final double radius;

  /// Radius Flutter reserves *track space* for, via [getPreferredSize] — see
  /// that override's doc comment. Deliberately a separate field from
  /// [radius] rather than reusing it: this must stay the same value across
  /// every rebuild of a given seek bar, no matter what [radius] does from
  /// frame to frame, so callers should pass the largest [radius] this shape
  /// will ever be asked to paint (for the full player's constant-radius
  /// thumb, that is just the same value as [radius] itself).
  final double maxRadius;
  final bool glowing;
  final Color glowColor;

  /// The halo's radius relative to the thumb's own — wide enough to read as
  /// a glow rather than a second, slightly bigger dot.
  static const _glowRadiusMultiplier = 2.75;

  static const _glowBlurSigma = 8.0;

  /// v5.32.0 field report: "鼠标放上去出现控制点会导致整个进度条收缩一部分"
  /// ("hovering to reveal the thumb shrinks the whole track"). Root cause:
  /// [BaseSliderTrackShape.getPreferredRect] insets the track at each end by
  /// `max(thumbWidth, overlayWidth) / 2`, using *this method's* return value
  /// — not whatever [paint] happens to draw that frame. The previous
  /// implementation returned `Size.fromRadius(radius)` here, so the instant
  /// a hover swapped in a bigger `radius` the reserved inset grew with it,
  /// visibly shrinking the track by exactly that many pixels on both ends —
  /// the track was never actually shrinking, it was being asked to lay out
  /// inside a smaller rectangle every time the thumb grew. Returning
  /// [maxRadius] — a value that never changes across a given seek bar's
  /// lifetime — keeps that reserved inset, and therefore the track's own
  /// geometry, fixed; only what [paint] draws inside that fixed reservation
  /// is allowed to react to hover.
  @override
  Size getPreferredSize(bool isEnabled, bool isDiscrete) =>
      Size.fromRadius(maxRadius);

  @override
  void paint(
    PaintingContext context,
    Offset center, {
    required Animation<double> activationAnimation,
    required Animation<double> enableAnimation,
    required bool isDiscrete,
    required TextPainter labelPainter,
    required RenderBox parentBox,
    required SliderThemeData sliderTheme,
    required TextDirection textDirection,
    required double value,
    required double textScaleFactor,
    required Size sizeWithOverflow,
  }) {
    if (radius <= 0) return;
    final canvas = context.canvas;
    if (glowing) {
      canvas.drawCircle(
        center,
        radius * _glowRadiusMultiplier,
        Paint()
          ..color = glowColor
          ..maskFilter = const MaskFilter.blur(
            BlurStyle.normal,
            _glowBlurSigma,
          ),
      );
    }
    final colorTween = ColorTween(
      begin: sliderTheme.disabledThumbColor,
      end: sliderTheme.thumbColor,
    );
    canvas.drawCircle(
      center,
      radius,
      Paint()..color = colorTween.evaluate(enableAnimation)!,
    );
  }
}

/// How long a seek bar's displayed position takes to glide from one
/// [PlaybackEngine.positionStream] tick to the next, via a
/// [TweenAnimationBuilder] wrapping the raw position value — rather than
/// snapping straight to it as every seek row did through v5.30.6.
///
/// [TweenAnimationBuilder] rather than a hand-rolled [AnimationController]
/// deliberately: retargeting it mid-flight (a new tick arriving before the
/// previous 260ms leg finishes) is handled by the framework itself — the new
/// animation starts from wherever the current one is, not from the old
/// tick's value — and its controller lives and dies with the widget, so
/// there is no dispose-ordering pitfall to get wrong here at all (see
/// [GlowingSliderThumbShape]'s own doc comment on the controller-lifecycle
/// mistake this sidesteps by construction, not by care).
///
/// 260ms: short enough that the displayed position never trails the real one
/// by an amount a listener would notice, long enough to visibly smooth over
/// the ~200ms-1s gaps just_audio's own positionStream typically ticks at —
/// below that gap the animation keeps chasing an ever-advancing target
/// (continuous glide); above it, it finishes early and sits still until the
/// next tick, which still reads as far smoother than a hard jump every tick.
const positionTweenDuration = Duration(milliseconds: 260);
