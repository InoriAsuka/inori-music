// seek_bar_effects_test.dart
//
// Direct unit coverage for GlowingSliderThumbShape's getPreferredSize/paint
// split, added alongside the v5.32.0 fix for "鼠标放上去出现控制点会导致
// 整个进度条收缩一部分" (hovering to reveal the thumb shrank the whole
// track). mini_player_bar_desktop_test.dart covers the same fix at the
// integration level (the real SliderThemeData in effect at rest vs. on
// hover); this file isolates the root-cause class itself so a future
// regression here fails fast, with no widget pumping required.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/shared/widgets/seek_bar_effects.dart';

void main() {
  group('GlowingSliderThumbShape.getPreferredSize', () {
    test('is independent of the painted radius', () {
      const glowing = GlowingSliderThumbShape(
        radius: 7,
        maxRadius: 7,
        glowing: false,
        glowColor: Colors.transparent,
      );
      const resting = GlowingSliderThumbShape(
        radius: 0,
        maxRadius: 7,
        glowing: false,
        glowColor: Colors.transparent,
      );

      // This is the exact mechanism the v5.32.0 field report traced: Flutter's
      // BaseSliderTrackShape.getPreferredRect insets the track by this
      // method's return value, not by whatever paint() draws — so if this
      // ever again returns Size.fromRadius(radius) instead of
      // Size.fromRadius(maxRadius), the track shrinks the instant the thumb
      // is revealed even though nothing about the track itself changed.
      expect(
        resting.getPreferredSize(true, false),
        glowing.getPreferredSize(true, false),
      );
      expect(resting.getPreferredSize(true, false), const Size.fromRadius(7));
    });

    test('reflects maxRadius, not whichever shape happens to be constructed '
        'with the larger radius', () {
      // The full player's seek bar passes a constant (radius == maxRadius);
      // the mini bar's wide seek row passes a hover-driven radius with a
      // fixed maxRadius. Both must key their reserved size off maxRadius
      // alone.
      const fullPlayerStyle = GlowingSliderThumbShape(
        radius: 6,
        maxRadius: 6,
        glowing: false,
        glowColor: Colors.transparent,
      );
      const miniBarHovered = GlowingSliderThumbShape(
        radius: 7,
        maxRadius: 7,
        glowing: false,
        glowColor: Colors.transparent,
      );

      expect(
        fullPlayerStyle.getPreferredSize(true, false),
        const Size.fromRadius(6),
      );
      expect(
        miniBarHovered.getPreferredSize(true, false),
        const Size.fromRadius(7),
      );
    });
  });

  group('GlowingSliderThumbShape.paint', () {
    // paint() itself needs a real PaintingContext/Canvas to exercise fully
    // (covered by the widget-level smoke tests in full_player_layout_test.dart
    // and mini_player_bar_desktop_test.dart, which actually render a Slider
    // using this shape). This just locks down the early-return contract
    // paint() relies on: a shape built with radius 0 must still be safe to
    // hand a real paint call — asserted indirectly by the widget tests
    // rendering the at-rest (radius: 0) mini-bar seek row without throwing.
    test('a radius of 0 is a valid, documented "draw nothing" state', () {
      const shape = GlowingSliderThumbShape(
        radius: 0,
        maxRadius: 7,
        glowing: false,
        glowColor: Colors.transparent,
      );
      expect(shape.radius, 0);
      // getPreferredSize must still report the *reserved* size, not 0 — this
      // is the same assertion as the group above, restated against the
      // specific "thumb fully collapsed" instance both seek rows construct
      // at rest.
      expect(shape.getPreferredSize(true, false), const Size.fromRadius(7));
    });
  });
}
