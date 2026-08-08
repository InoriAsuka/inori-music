// cover_palette_test.dart
//
// Covers cover-art colour extraction (v5.24.0) and what v5.26.0 does with it:
// the two opposite swatch preferences (backdrop vs accent), the provider's
// null paths, and the fact that LyricsBackground only takes over the skin
// when there is actually a cover to derive it from.
//
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/lyrics/lyrics_background.dart';
import 'package:inori_music/src/shared/theme/artwork_overlay_skin.dart';
import 'package:inori_music/src/shared/theme/skin_definition.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/cover_fluid_background.dart';

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

  group('CoverPalette.accentOverArtwork', () {
    const dominant = Color(0xFF102030);
    const lightMuted = Color(0xFFE0D0E0);
    const darkVibrant = Color(0xFF400020);
    const lightVibrant = Color(0xFFFF80C0);

    test('prefers a vibrant swatch, unlike the backdrop colour', () {
      // The mirror image of backdropFor: this colour lands on controls, where
      // popping off the artwork is the whole point.
      const palette = CoverPalette(
        dominant: dominant,
        lightMuted: lightMuted,
        darkVibrant: darkVibrant,
        lightVibrant: lightVibrant,
      );

      expect(palette.accentOverArtwork, lightVibrant);
      expect(
        palette.backdropFor(Brightness.dark),
        isNot(palette.accentOverArtwork),
        reason: 'Backdrop and accent must not converge on the same swatch',
      );
    });

    test('degrades vibrant -> muted -> dominant', () {
      expect(
        const CoverPalette(
          dominant: dominant,
          darkVibrant: darkVibrant,
        ).accentOverArtwork,
        darkVibrant,
      );
      expect(
        const CoverPalette(
          dominant: dominant,
          lightMuted: lightMuted,
        ).accentOverArtwork,
        lightMuted,
      );
      expect(
        const CoverPalette(dominant: dominant).accentOverArtwork,
        dominant,
      );
    });
  });

  group('LyricsBackground', () {
    // pump() rather than pumpAndSettle() throughout: once a cover resolves,
    // CoverFluidBackground runs two repeat() controllers that never settle,
    // so pumpAndSettle would simply time out.
    Widget buildApp({String? artworkUrl, CoverPalette? palette}) =>
        ProviderScope(
          overrides: [
            artworkUrlProvider.overrideWith(
              () => _StubArtworkNotifier(artworkUrl),
            ),
            coverPaletteProvider.overrideWith((ref, source) async => palette),
          ],
          child: const MaterialApp(
            home: LyricsBackground(
              albumId: 'album-1',
              localArtUri: null,
              child: _SkinProbe(),
            ),
          ),
        );

    testWidgets('renders no artwork backdrop when there is no cover', (
      tester,
    ) async {
      await tester.pumpWidget(buildApp(artworkUrl: null));
      await tester.pumpAndSettle();

      expect(find.byType(CoverFluidBackground), findsNothing);
      expect(
        _SkinProbe.captured!.onSurface,
        SkinDefinition.sakuraDusk.colors.onSurface,
        reason: "With no artwork the user's real skin must be left alone",
      );
    });

    testWidgets('a cover switches content to the overlay skin', (tester) async {
      await tester.pumpWidget(buildApp(artworkUrl: 'https://example/a.jpg'));
      await tester.pump();
      await tester.pump();

      expect(find.byType(CoverFluidBackground), findsOneWidget);
      expect(
        _SkinProbe.captured!.onSurface,
        Colors.white,
        reason: 'Dark ink over a saturated moving field is unreadable',
      );
      expect(
        _SkinProbe.captured!.background,
        Colors.transparent,
        reason: 'An opaque background would hide the backdrop underneath',
      );
    });

    testWidgets('the accent follows the extracted palette', (tester) async {
      const vibrant = Color(0xFFFF80C0);
      await tester.pumpWidget(
        buildApp(
          artworkUrl: 'https://example/a.jpg',
          palette: const CoverPalette(
            dominant: Color(0xFF102030),
            lightVibrant: vibrant,
          ),
        ),
      );
      await tester.pump();
      await tester.pump();

      expect(_SkinProbe.captured!.sakuraPink, vibrant);
    });

    testWidgets('the accent holds at the skin colour until a palette lands', (
      tester,
    ) async {
      // Nothing may flash through an uncoloured frame while extraction runs.
      await tester.pumpWidget(
        buildApp(artworkUrl: 'https://example/a.jpg', palette: null),
      );
      await tester.pump();
      await tester.pump();

      expect(
        _SkinProbe.captured!.sakuraPink,
        SkinDefinition.sakuraDusk.colors.sakuraPink,
      );
    });
  });
}

/// Stubs the artwork URL lookup so tests can drive the "has a cover" branch
/// without a network round trip.
class _StubArtworkNotifier extends ArtworkUrlNotifier {
  _StubArtworkNotifier(this._url);
  final String? _url;

  @override
  Future<String?> build(String albumId) async => _url;
}

/// Captures whatever SkinColors its position in the tree resolves to, so a
/// test can assert on the skin LyricsBackground handed down rather than on
/// rendered pixels.
class _SkinProbe extends StatelessWidget {
  const _SkinProbe();

  static SkinColors? captured;

  @override
  Widget build(BuildContext context) {
    captured = context.skinColors;
    return const SizedBox.shrink();
  }
}
