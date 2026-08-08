// artwork_overlay_skin_test.dart
//
// Regression guard for the v5.29.0 contrast fix. artworkOverlaySkin() used to
// assume the CoverFluidBackground scrim was always dark by construction and
// hardcoded white ink for everything painted over it. That assumption does
// not hold: the backdrop only darkens a *brightened* copy of the cover with a
// 24% black scrim, so a light cover still yields a light backdrop, and white
// text/icons on it are unreadable.
//
// These tests exercise the pure functions directly rather than pumping
// CoverFluidBackground/LyricsBackground: both run repeat() animation
// controllers that never settle under pumpAndSettle, and LyricsBackground's
// palette provider would otherwise need network/file-system stubbing for no
// extra coverage — backdropLuminance()/artworkOverlaySkin() already cover the
// same decision logic.
//
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/shared/theme/artwork_overlay_skin.dart';
import 'package:inori_music/src/shared/theme/skin_definition.dart';

CoverPalette _palette({required Color dominant, Color? lightVibrant}) =>
    CoverPalette(dominant: dominant, lightVibrant: lightVibrant);

void main() {
  group('backdropLuminance', () {
    test('a light dominant colour yields a light backdrop', () {
      const lightCover = Color(0xFFF2E9E4);
      expect(
        backdropLuminance(_palette(dominant: lightCover)),
        greaterThanOrEqualTo(0.5),
      );
    });

    test('a dark dominant colour yields a dark backdrop', () {
      const darkCover = Color(0xFF10101A);
      expect(backdropLuminance(_palette(dominant: darkCover)), lessThan(0.5));
    });
  });

  group('artworkOverlaySkin foreground', () {
    test('a light backdrop flips ink to dark, not white-on-white', () {
      final skin = artworkOverlaySkin(
        SkinDefinition.sakuraDusk,
        palette: _palette(dominant: const Color(0xFFF2E9E4)),
      );

      expect(skin.brightness, Brightness.light);
      expect(skin.colors.onBackground.computeLuminance(), lessThan(0.05));
      expect(skin.colors.onSurface.computeLuminance(), lessThan(0.05));
      expect(skin.colors.surface, Colors.black.withValues(alpha: 0.10));
      expect(
        skin.colors.surfaceContainer,
        Colors.black.withValues(alpha: 0.14),
      );
    });

    test('a dark, warm backdrop keeps the pre-v5.29.0 white treatment', () {
      // Regression guard: this was the only branch that existed before
      // v5.29.0, and its output must not change.
      final skin = artworkOverlaySkin(
        SkinDefinition.sakuraDusk,
        palette: _palette(dominant: const Color(0xFF1A1008)),
      );

      expect(skin.brightness, Brightness.dark);
      expect(skin.colors.onBackground, Colors.white);
      expect(skin.colors.onSurface, Colors.white);
      expect(
        skin.colors.onSurfaceVariant,
        Colors.white.withValues(alpha: 0.72),
      );
      expect(skin.colors.outline, Colors.white.withValues(alpha: 0.34));
      expect(skin.colors.outlineVariant, Colors.white.withValues(alpha: 0.18));
      expect(skin.colors.surface, Colors.white.withValues(alpha: 0.10));
      expect(skin.colors.surfaceVariant, Colors.white.withValues(alpha: 0.08));
    });

    test('no palette yet keeps the historical dark-backdrop default', () {
      final skin = artworkOverlaySkin(SkinDefinition.sakuraDusk, palette: null);

      expect(skin.brightness, Brightness.dark);
      expect(skin.colors.onBackground, Colors.white);
    });
  });

  group('accent independence', () {
    test('the accent does not flip with backdrop brightness', () {
      const sharedVibrant = Color(0xFFEE8888);
      final lightBackdrop = artworkOverlaySkin(
        SkinDefinition.sakuraDusk,
        palette: _palette(
          dominant: const Color(0xFFF2E9E4),
          lightVibrant: sharedVibrant,
        ),
      );
      final darkBackdrop = artworkOverlaySkin(
        SkinDefinition.sakuraDusk,
        palette: _palette(
          dominant: const Color(0xFF10101A),
          lightVibrant: sharedVibrant,
        ),
      );

      // The play/pause button and every accent-coloured control must stay on
      // the cover's colour regardless of which ink branch the rest of the
      // foreground took.
      expect(lightBackdrop.colors.sakuraPink, sharedVibrant);
      expect(darkBackdrop.colors.sakuraPink, sharedVibrant);
    });
  });

  group('cool-hue warmth nudge', () {
    test('a dark, cool-hued backdrop nudges white ink warm', () {
      final skin = artworkOverlaySkin(
        SkinDefinition.sakuraDusk,
        palette: _palette(dominant: const Color(0xFF10101A)),
      );

      expect(skin.colors.onBackground, isNot(Colors.white));
      expect(skin.colors.onBackground.b, lessThan(skin.colors.onBackground.r));
    });

    test('a dark, warm-hued backdrop is left as pure white', () {
      // Only the cool branch gets nudged — this is the asymmetric "brightness
      // only, plus one warmth nudge" scope decision, not a full complementary
      // colour system.
      final skin = artworkOverlaySkin(
        SkinDefinition.sakuraDusk,
        palette: _palette(dominant: const Color(0xFF1A1008)),
      );

      expect(skin.colors.onBackground, Colors.white);
    });
  });

  group('id encodes the brightness tier', () {
    test('the same accent on a light vs dark backdrop yields distinct ids', () {
      const sharedVibrant = Color(0xFFEE8888);
      final lightBackdrop = artworkOverlaySkin(
        SkinDefinition.sakuraDusk,
        palette: _palette(
          dominant: const Color(0xFFF2E9E4),
          lightVibrant: sharedVibrant,
        ),
      );
      final darkBackdrop = artworkOverlaySkin(
        SkinDefinition.sakuraDusk,
        palette: _palette(
          dominant: const Color(0xFF10101A),
          lightVibrant: sharedVibrant,
        ),
      );

      // Otherwise SkinScope.updateShouldNotify (id-based) would not
      // propagate the flip between them.
      expect(lightBackdrop.id, isNot(darkBackdrop.id));
    });
  });
}
