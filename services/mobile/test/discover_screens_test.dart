// ignore_for_file: implementation_imports
//
// discover_screens_test.dart
//
// Smoke coverage for the two new v5.33.0 "发现音乐" destinations
// (shell_scaffold.dart's _desktopDiscoverItems): ForYouScreen (a guiding
// empty state — there is no recommendation engine behind it) and
// ExploreScreen (real content — recently-added and randomly-sampled
// catalog albums via CatalogRepository.listAlbums).
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:inori_api/src/model/catalog_album.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/discover/explore_screen.dart';
import 'package:inori_music/src/discover/for_you_screen.dart';
import 'package:inori_music/src/shared/widgets/album_card.dart';

/// Stubs artwork lookups to a bare `null` with no network round trip and no
/// 300s keep-alive Timer (see ArtworkUrlNotifier.build) — the same fix
/// mini_player_bar_test.dart's own `_StubArtworkNotifier` exists for.
/// AlbumCard renders fine with no artwork (its own placeholder branch); what
/// it cannot survive is a real Dio call with nothing behind it, which is
/// what every AlbumCard in ExploreScreen would otherwise attempt.
class _NullArtworkNotifier extends ArtworkUrlNotifier {
  @override
  Future<String?> build(String albumId) async => null;
}

Widget _app(Widget child, {List<Override> overrides = const []}) =>
    ProviderScope(
      overrides: [
        artworkUrlProvider.overrideWith(_NullArtworkNotifier.new),
        ...overrides,
      ],
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: child,
      ),
    );

CatalogAlbum _album(String id, String title) => CatalogAlbum(
  id: id,
  title: title,
  artistId: 'artist-1',
  createdAt: DateTime(2026, 1, 1),
  updatedAt: DateTime(2026, 1, 1),
);

void main() {
  group('ForYouScreen', () {
    testWidgets('renders a guiding empty state, not a dead blank screen', (
      tester,
    ) async {
      await tester.pumpWidget(_app(const ForYouScreen()));
      await tester.pump();

      final t = TestAppLocalizations.of(tester);
      expect(find.text(t.forYou), findsAtLeastNWidgets(1));
      expect(find.text(t.forYouComingSoon), findsOneWidget);
      expect(find.byIcon(Icons.auto_awesome_outlined), findsOneWidget);
    });
  });

  group('ExploreScreen', () {
    testWidgets(
      'renders both section headings and the fetched albums as AlbumCards',
      (tester) async {
        final recent = [_album('a1', 'Recent Album')];
        final random = [_album('a2', 'Random Album')];
        await tester.pumpWidget(
          _app(
            const ExploreScreen(),
            overrides: [
              recentlyAddedAlbumsProvider.overrideWith((ref) async => recent),
              randomAlbumsProvider.overrideWith((ref) async => random),
            ],
          ),
        );
        // Two pumps: both providers resolve via a microtask even though the
        // override itself has no real async gap.
        await tester.pump();
        await tester.pump();

        final t = TestAppLocalizations.of(tester);
        expect(find.text(t.recentlyAdded), findsOneWidget);
        expect(find.text(t.randomPicks), findsOneWidget);
        expect(find.text('Recent Album'), findsOneWidget);
        expect(find.text('Random Album'), findsOneWidget);
        expect(find.byType(AlbumCard), findsNWidgets(2));
      },
    );

    testWidgets('an empty catalog renders the shared "no data" state, not '
        'an error or a crash', (tester) async {
      await tester.pumpWidget(
        _app(
          const ExploreScreen(),
          overrides: [
            recentlyAddedAlbumsProvider.overrideWith((ref) async => const []),
            randomAlbumsProvider.overrideWith((ref) async => const []),
          ],
        ),
      );
      await tester.pump();
      await tester.pump();

      expect(tester.takeException(), isNull);
      final t = TestAppLocalizations.of(tester);
      expect(find.text(t.noData), findsNWidgets(2));
    });
  });
}

/// Small helper so this file's assertions read off the same localized
/// strings the widgets themselves use, rather than hardcoded English copies
/// that would silently stop matching if a translation changed.
extension TestAppLocalizations on AppLocalizations {
  static AppLocalizations of(WidgetTester tester) =>
      AppLocalizations.of(tester.element(find.byType(Directionality).first));
}
