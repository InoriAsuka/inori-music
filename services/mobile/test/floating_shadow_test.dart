// floating_shadow_test.dart
//
// Covers floatingShadow() (lib/src/shared/widgets/floating_shadow.dart),
// the v5.30.6 replacement for Material's `elevation` on the mini player bar
// and every GlassPanel (sidebar, full player control/side panels) — the
// field report's "跟苹果相差一节" complaint about Material's single tight,
// always-black shadow. Plus a light integration check that GlassPanel
// actually wires this in ahead of its ClipRRect rather than behind it,
// where it would be clipped away.
//
// mini_player_bar_test.dart separately covers the mini player bar's own
// use of this function (its Material.elevation is 0, and its own
// DecoratedBox carries the resulting shadow) — that check needs the
// existing PlayerNotifier-stubbing setup already defined there.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/shared/widgets/floating_shadow.dart';
import 'package:inori_music/src/shared/widgets/glass_panel.dart';

void main() {
  group('floatingShadow', () {
    test('returns exactly two layers — a tight one and an ambient spread', () {
      final shadows = floatingShadow(const Color(0x263B2A3F));
      expect(shadows, hasLength(2));
    });

    test('the ambient (second) layer uses a larger blur than the tight '
        '(first) one — that gap is what reads as "soft" rather than '
        '"Material elevation" tight', () {
      final shadows = floatingShadow(const Color(0x263B2A3F));
      expect(shadows[1].blurRadius, greaterThan(shadows[0].blurRadius));
    });

    test('both layers fall inside the Apple-reference 24-32px ambient / '
        'tight-closer-than-that blur band', () {
      final shadows = floatingShadow(const Color(0x263B2A3F));
      expect(shadows[1].blurRadius, inInclusiveRange(24.0, 32.0));
      expect(shadows[0].blurRadius, lessThan(shadows[1].blurRadius));
    });

    test('a higher-alpha base (a dark-skin token) produces a stronger '
        'shadow than a lower-alpha one (a light-skin token) in both '
        'layers, so each skin gets a shadow tuned to its own ground', () {
      const lightSkinToken = Color(0x263B2A3F); // Sakura Dusk's own token
      const darkSkinToken = Color(0x66000000); // Moonlit Indigo's own token

      final lightShadows = floatingShadow(lightSkinToken);
      final darkShadows = floatingShadow(darkSkinToken);

      expect(darkShadows[0].color.a, greaterThan(lightShadows[0].color.a));
      expect(darkShadows[1].color.a, greaterThan(lightShadows[1].color.a));
    });

    test('never produces an out-of-range alpha even for an already-opaque '
        'base', () {
      final shadows = floatingShadow(const Color(0xFF000000));
      for (final shadow in shadows) {
        expect(shadow.color.a, inInclusiveRange(0.0, 1.0));
      }
    });

    test('both layers offset downward, matching a light source above the '
        'panel rather than beside or below it', () {
      final shadows = floatingShadow(const Color(0x263B2A3F));
      for (final shadow in shadows) {
        expect(shadow.offset.dy, greaterThan(0));
      }
    });
  });

  group('GlassPanel', () {
    testWidgets(
      'carries its shadow on a DecoratedBox ahead of its ClipRRect, not '
      'behind it (a shadow painted inside the clip would be cut off at the '
      'panel\'s own edge instead of spreading past it)',
      (tester) async {
        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: GlassPanel(child: SizedBox(width: 100, height: 100)),
            ),
          ),
        );
        await tester.pump();

        expect(find.byType(ClipRRect), findsWidgets);
        expect(find.byType(DecoratedBox), findsWidgets);

        // The DecoratedBox carrying the shadow must be an ancestor of the
        // ClipRRect, i.e. outside it — ancestor(of: ClipRRect, matching:
        // DecoratedBox) only matches widgets *above* the ClipRRect in the
        // tree.
        final shadowBox = find.ancestor(
          of: find.byType(ClipRRect).first,
          matching: find.byWidgetPredicate((widget) {
            if (widget is! DecoratedBox) return false;
            final decoration = widget.decoration;
            return decoration is BoxDecoration &&
                (decoration.boxShadow?.isNotEmpty ?? false);
          }),
        );
        expect(shadowBox, findsOneWidget);
      },
    );
  });
}
