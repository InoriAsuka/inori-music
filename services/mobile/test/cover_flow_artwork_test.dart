// cover_flow_artwork_test.dart
//
// Covers the v5.30.0 Cover Flow artwork display mode: the pure side-count
// geometry (testable without pumping a widget, same pattern
// playerArtworkSize/playerControlWidth use in full_player_layout_test.dart)
// and the widget's actual rendered card count at different widths, since the
// plan's own validation explicitly calls out "rendered" card count, not just
// the number the geometry function returns.
//
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/player/cover_flow_artwork.dart';

/// The tap tests position side covers beyond the default 800x600 test
/// viewport (by design — Cover Flow needs real width to have neighbours to
/// tap), so tester.tap()'s synthetic pointer event needs a large enough
/// viewport to actually land on them.
void _sizeWindow(WidgetTester tester, Size size) {
  tester.view.devicePixelRatio = 1.0;
  tester.view.physicalSize = size;
  addTearDown(tester.view.reset);
}

Widget _harness({
  required int itemCount,
  required int currentIndex,
  required double centerSize,
  required double width,
  ValueChanged<int>? onSelect,
}) => MaterialApp(
  home: Scaffold(
    body: Center(
      child: CoverFlowArtwork(
        itemCount: itemCount,
        currentIndex: currentIndex,
        centerSize: centerSize,
        width: width,
        onSelect: onSelect,
        itemBuilder: (context, index, size) => SizedBox(
          key: ValueKey('cover-$index'),
          width: size,
          height: size,
          child: Text('$index'),
        ),
      ),
    ),
  ),
);

void main() {
  group('CoverFlowArtwork.visibleSideCount', () {
    test('a narrow width leaves no room for side covers', () {
      expect(CoverFlowArtwork.visibleSideCount(width: 400, centerSize: 300), 0);
    });

    test('a wider window fits more side covers than a narrower one', () {
      final narrow = CoverFlowArtwork.visibleSideCount(
        width: 700,
        centerSize: 300,
      );
      final wide = CoverFlowArtwork.visibleSideCount(
        width: 1200,
        centerSize: 300,
      );
      expect(
        wide,
        greaterThan(narrow),
        reason:
            'The whole point of Cover Flow is that width, not a fixed '
            'count, decides how many neighbours show',
      );
    });

    test('side count is capped even on an extremely wide window', () {
      expect(
        CoverFlowArtwork.visibleSideCount(width: 20000, centerSize: 300),
        lessThanOrEqualTo(3),
        reason:
            'An unbounded count would read as a filmstrip, not a "flow" '
            'with one clear focal point',
      );
    });

    test('a zero or negative centre size never divides by zero', () {
      expect(
        () => CoverFlowArtwork.visibleSideCount(width: 1000, centerSize: 0),
        returnsNormally,
      );
      expect(CoverFlowArtwork.visibleSideCount(width: 1000, centerSize: 0), 0);
    });
  });

  group('CoverFlowArtwork widget', () {
    testWidgets('renders more covers at a wider width than a narrower one', (
      tester,
    ) async {
      await tester.pumpWidget(
        _harness(itemCount: 9, currentIndex: 4, centerSize: 200, width: 500),
      );
      final narrowCount = tester
          .widgetList(
            find.byWidgetPredicate((w) => w is SizedBox && w.key is ValueKey),
          )
          .length;

      await tester.pumpWidget(
        _harness(itemCount: 9, currentIndex: 4, centerSize: 200, width: 1400),
      );
      final wideCount = tester
          .widgetList(
            find.byWidgetPredicate((w) => w is SizedBox && w.key is ValueKey),
          )
          .length;

      expect(
        wideCount,
        greaterThan(narrowCount),
        reason:
            'Rendered card count must actually track width, not just '
            'the geometry function in isolation',
      );
    });

    testWidgets('the current index is always among the rendered covers', (
      tester,
    ) async {
      await tester.pumpWidget(
        _harness(itemCount: 9, currentIndex: 4, centerSize: 200, width: 1400),
      );

      expect(find.byKey(const ValueKey('cover-4')), findsOneWidget);
    });

    testWidgets('never renders past the ends of a short queue', (tester) async {
      // itemCount 3, centred on the first track: only indices 0-2 exist, so
      // even an enormous width must not conjure negative or out-of-range
      // slots.
      await tester.pumpWidget(
        _harness(itemCount: 3, currentIndex: 0, centerSize: 200, width: 3000),
      );
      await tester.pump();

      expect(find.byKey(const ValueKey('cover-0')), findsOneWidget);
      expect(find.byKey(const ValueKey('cover-1')), findsOneWidget);
      expect(find.byKey(const ValueKey('cover-2')), findsOneWidget);
      expect(find.byKey(const ValueKey('cover--1')), findsNothing);
      expect(find.byKey(const ValueKey('cover-3')), findsNothing);
      expect(tester.takeException(), isNull);
    });

    testWidgets('tapping a side cover reports its index, not the centre\'s', (
      tester,
    ) async {
      _sizeWindow(tester, const Size(1600, 800));
      int? selected;
      await tester.pumpWidget(
        _harness(
          itemCount: 9,
          currentIndex: 4,
          centerSize: 200,
          width: 1400,
          onSelect: (i) => selected = i,
        ),
      );

      await tester.tap(find.byKey(const ValueKey('cover-5')));
      expect(selected, 5);
    });

    testWidgets('tapping the centre cover does not invoke onSelect', (
      tester,
    ) async {
      _sizeWindow(tester, const Size(1600, 800));
      var tapped = false;
      await tester.pumpWidget(
        _harness(
          itemCount: 9,
          currentIndex: 4,
          centerSize: 200,
          width: 1400,
          onSelect: (_) => tapped = true,
        ),
      );

      await tester.tap(find.byKey(const ValueKey('cover-4')));
      expect(
        tapped,
        isFalse,
        reason:
            'The centre is already the current track; selecting it '
            'again is a no-op the caller should not be asked to handle',
      );
    });
  });
}
