// cover_palette_test.dart
//
// Covers v5.24.0's cover-art colour extraction: the swatch fallback order,
// the provider's null paths (no artwork / unreadable file), and the fact
// that LyricsBackground degrades to its skin colour rather than rendering an
// unscrimmed image whenever extraction yields nothing.
//
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/lyrics/lyrics_background.dart';
import 'package:inori_music/src/shared/theme/skin_definition.dart';

void main() {
  group('CoverPalette.backdropFor', () {
    const dominant = Color(0xFF102030);
    const darkMuted = Color(0xFF201020);
    const lightMuted = Color(0xFFE0D0E0);
    const darkVibrant = Color(0xFF400020);
    const lightVibrant = Color(0xFFFF80C0);

    test('prefers the muted swatch matching the theme brightness', () {
      const palette = CoverPalette(
        dominant: dominant,
        darkMuted: darkMuted,
        lightMuted: lightMuted,
        darkVibrant: darkVibrant,
        lightVibrant: lightVibrant,
      );

      expect(palette.backdropFor(Brightness.dark), darkMuted);
      expect(palette.backdropFor(Brightness.light), lightMuted);
    });

    test('falls back to the vibrant swatch when muted is missing', () {
      const palette = CoverPalette(
        dominant: dominant,
        darkVibrant: darkVibrant,
        lightVibrant: lightVibrant,
      );

      expect(palette.backdropFor(Brightness.dark), darkVibrant);
      expect(palette.backdropFor(Brightness.light), lightVibrant);
    });

    test('falls back to dominant when the image yields nothing else', () {
      const palette = CoverPalette(dominant: dominant);

      expect(palette.backdropFor(Brightness.dark), dominant);
      expect(palette.backdropFor(Brightness.light), dominant);
    });

    test('a dark-only palette never hands a light theme a dark swatch it '
        'does not have', () {
      // Guards the asymmetry in the fallback chain: brightness-specific
      // swatches must not cross over, only degrade to dominant.
      const palette = CoverPalette(
        dominant: dominant,
        darkMuted: darkMuted,
        darkVibrant: darkVibrant,
      );

      expect(palette.backdropFor(Brightness.light), dominant);
    });
  });

  group('coverPaletteProvider', () {
    test('returns null when there is no artwork to sample', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final result = await container.read(
        coverPaletteProvider((albumId: null, localArtUri: null)).future,
      );
      expect(result, isNull);
    });

    test('returns null for an empty album id', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final result = await container.read(
        coverPaletteProvider((albumId: '', localArtUri: null)).future,
      );
      expect(result, isNull);
    });

    test(
      'returns null rather than throwing for a missing local cover file',
      () async {
        final container = ProviderContainer();
        addTearDown(container.dispose);

        final result = await container.read(
          coverPaletteProvider((
            albumId: null,
            localArtUri: Uri.file('/definitely/not/a/real/cover.jpg'),
          )).future,
        );
        expect(result, isNull);
      },
    );
  });

  group('LyricsBackground scrim', () {
    LinearGradient gradientOf(WidgetTester tester) {
      final container = tester.widget<AnimatedContainer>(
        find.byType(AnimatedContainer),
      );
      return (container.decoration! as BoxDecoration).gradient!
          as LinearGradient;
    }

    Widget buildApp({CoverPalette? palette}) => ProviderScope(
      overrides: [
        coverPaletteProvider.overrideWith((ref, source) async => palette),
      ],
      child: const MaterialApp(
        home: LyricsBackground(
          albumId: 'album-1',
          localArtUri: null,
          child: SizedBox.shrink(),
        ),
      ),
    );

    testWidgets('uses the extracted colour once a palette resolves', (
      tester,
    ) async {
      const extracted = Color(0xFF3A1F2B);
      await tester.pumpWidget(
        buildApp(palette: const CoverPalette(dominant: extracted)),
      );
      await tester.pumpAndSettle();

      expect(gradientOf(tester).colors, [
        extracted.withValues(alpha: 0.55),
        extracted.withValues(alpha: 0.82),
      ]);
    });

    testWidgets('falls back to the skin background when extraction yields '
        'nothing', (tester) async {
      await tester.pumpWidget(buildApp(palette: null));
      await tester.pumpAndSettle();

      final skinBackground = SkinDefinition.sakuraDusk.colors.background;
      expect(gradientOf(tester).colors, [
        skinBackground.withValues(alpha: 0.55),
        skinBackground.withValues(alpha: 0.82),
      ]);
    });

    testWidgets('scrims while the palette is still resolving', (tester) async {
      // The first frame must already be scrimmed — otherwise lyrics would sit
      // on raw artwork until extraction lands.
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            coverPaletteProvider.overrideWith((ref, source) {
              return Future<CoverPalette?>.delayed(
                const Duration(seconds: 1),
                () => const CoverPalette(dominant: Color(0xFF3A1F2B)),
              );
            }),
          ],
          child: const MaterialApp(
            home: LyricsBackground(
              albumId: 'album-1',
              localArtUri: null,
              child: SizedBox.shrink(),
            ),
          ),
        ),
      );
      await tester.pump();

      final skinBackground = SkinDefinition.sakuraDusk.colors.background;
      expect(
        gradientOf(tester).colors.first,
        skinBackground.withValues(alpha: 0.55),
      );

      // Let the delayed future complete so the test doesn't leave a pending timer.
      await tester.pump(const Duration(seconds: 1));
      await tester.pumpAndSettle();
    });
  });
}
