import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:window_manager/window_manager.dart';

import 'package:inori_music/src/shared/system_titlebar_provider.dart';
import 'package:inori_music/src/shared/widgets/desktop_app_bar.dart';

// ---------------------------------------------------------------------------
// DesktopAppBar reads DesktopIntegration.isDesktop (via defaultTargetPlatform)
// and systemTitleBarProvider to decide whether to add drag/window-button
// chrome around a plain AppBar. As in app_title_bar_test.dart (superseded by
// this file), debugDefaultTargetPlatformOverride must be reset synchronously
// at the end of each test body — the binding's end-of-test invariant check
// runs before addTearDown/tearDown callbacks would fire.
// ---------------------------------------------------------------------------

void main() {
  testWidgets('mobile: behaves like a plain AppBar, no window buttons', (
    tester,
  ) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.android;

    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(
          home: Scaffold(
            appBar: DesktopAppBar(title: Text('Title')),
            body: SizedBox(),
          ),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('Title'), findsOneWidget);
    expect(find.byType(WindowCaptionButton), findsNothing);
    expect(find.byType(DragToMoveArea), findsNothing);

    debugDefaultTargetPlatformOverride = null;
  });

  testWidgets(
    'desktop (Windows), custom title bar: drag area + three window buttons',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.windows;

      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              appBar: DesktopAppBar(title: Text('Title')),
              body: SizedBox(),
            ),
          ),
        ),
      );
      await tester.pump();

      expect(find.text('Title'), findsOneWidget);
      expect(find.byType(DragToMoveArea), findsOneWidget);
      expect(find.byType(WindowCaptionButton), findsNWidgets(3));

      debugDefaultTargetPlatformOverride = null;
    },
  );

  testWidgets(
    'desktop (macOS), custom title bar: drag area, no self-drawn buttons',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;

      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              appBar: DesktopAppBar(title: Text('Title')),
              body: SizedBox(),
            ),
          ),
        ),
      );
      await tester.pump();

      expect(find.byType(DragToMoveArea), findsOneWidget);
      expect(find.byType(WindowCaptionButton), findsNothing);

      debugDefaultTargetPlatformOverride = null;
    },
  );

  testWidgets(
    'desktop with systemTitleBarProvider enabled: falls back to a plain AppBar',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.windows;

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            systemTitleBarProvider.overrideWith(_AlwaysTrueSystemTitleBar.new),
          ],
          child: const MaterialApp(
            home: Scaffold(
              appBar: DesktopAppBar(title: Text('Title')),
              body: SizedBox(),
            ),
          ),
        ),
      );
      await tester.pump();

      expect(find.text('Title'), findsOneWidget);
      expect(find.byType(WindowCaptionButton), findsNothing);
      expect(find.byType(DragToMoveArea), findsNothing);

      debugDefaultTargetPlatformOverride = null;
    },
  );

  // -------------------------------------------------------------------------
  // DesktopSliverAppBar — same drag/window-button chrome as DesktopAppBar,
  // but as a SliverAppBar for collapsing-header screens (album/playlist/
  // artist detail). Critically, DragToMoveArea must wrap only `background`
  // (a box widget slot inside FlexibleSpaceBar), not the SliverAppBar itself
  // — GestureDetector (what DragToMoveArea builds) can't wrap a sliver and
  // still work as a CustomScrollView.slivers entry.
  // -------------------------------------------------------------------------

  Widget sliverHarness({List<Override> overrides = const []}) => ProviderScope(
    overrides: overrides,
    child: MaterialApp(
      home: Scaffold(
        body: CustomScrollView(
          slivers: [
            DesktopSliverAppBar(
              title: const Text('Album'),
              background: Container(color: Colors.blue),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 400)),
          ],
        ),
      ),
    ),
  );

  testWidgets(
    'DesktopSliverAppBar mobile: plain SliverAppBar, no window buttons',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.android;

      await tester.pumpWidget(sliverHarness());
      await tester.pump();

      expect(find.text('Album'), findsOneWidget);
      expect(find.byType(WindowCaptionButton), findsNothing);
      expect(find.byType(DragToMoveArea), findsNothing);

      debugDefaultTargetPlatformOverride = null;
    },
  );

  testWidgets(
    'DesktopSliverAppBar desktop (Windows): drag area on background + three window buttons',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.windows;

      await tester.pumpWidget(sliverHarness());
      await tester.pump();

      expect(find.text('Album'), findsOneWidget);
      expect(find.byType(DragToMoveArea), findsOneWidget);
      expect(find.byType(WindowCaptionButton), findsNWidgets(3));

      debugDefaultTargetPlatformOverride = null;
    },
  );

  testWidgets(
    'DesktopSliverAppBar desktop (macOS): drag area, no self-drawn buttons',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;

      await tester.pumpWidget(sliverHarness());
      await tester.pump();

      expect(find.byType(DragToMoveArea), findsOneWidget);
      expect(find.byType(WindowCaptionButton), findsNothing);

      debugDefaultTargetPlatformOverride = null;
    },
  );

  testWidgets(
    'DesktopSliverAppBar with systemTitleBarProvider enabled: falls back to a plain SliverAppBar',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.windows;

      await tester.pumpWidget(
        sliverHarness(
          overrides: [
            systemTitleBarProvider.overrideWith(_AlwaysTrueSystemTitleBar.new),
          ],
        ),
      );
      await tester.pump();

      expect(find.text('Album'), findsOneWidget);
      expect(find.byType(WindowCaptionButton), findsNothing);
      expect(find.byType(DragToMoveArea), findsNothing);

      debugDefaultTargetPlatformOverride = null;
    },
  );
}

class _AlwaysTrueSystemTitleBar extends SystemTitleBarNotifier {
  @override
  bool build() => true;
}
