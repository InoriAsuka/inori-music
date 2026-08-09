import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:window_manager/window_manager.dart';

import 'package:inori_music/src/shared/system_titlebar_provider.dart';
import 'package:inori_music/src/shared/widgets/desktop_app_bar.dart';
import 'package:inori_music/src/shared/widgets/shell_chrome.dart';

// ---------------------------------------------------------------------------
// DesktopAppBar reads DesktopIntegration.isDesktop (via defaultTargetPlatform)
// and systemTitleBarProvider to decide whether to add drag/window-button
// chrome around a plain AppBar. As in app_title_bar_test.dart (superseded by
// this file), debugDefaultTargetPlatformOverride must be reset synchronously
// at the end of each test body — the binding's end-of-test invariant check
// runs before addTearDown/tearDown callbacks would fire.
//
// v5.30.5 adds: the ShellChrome handoff that lets a full-height desktop
// sidebar take over macOS traffic-light gutter duty from this bar (see
// shell_chrome.dart's doc comment for the field-report bug this fixes).
// ---------------------------------------------------------------------------

/// Finds the exact Padding the pre-v5.30.5 unconditional gutter logic added
/// — a stand-in for "is the 70px reservation actually present", since the
/// harnesses below have no other Padding that happens to use this value.
final _gutterPadding = find.byWidgetPredicate(
  (w) => w is Padding && w.padding == const EdgeInsets.only(left: 70),
);

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
      // No ShellChrome ancestor here — same as the login gate or any other
      // chrome-free screen — so the pre-v5.30.5 fixed gutter still applies.
      expect(_gutterPadding, findsOneWidget);

      debugDefaultTargetPlatformOverride = null;
    },
  );

  testWidgets(
    'desktop (macOS) under a ShellChrome that already reserves the gutter: '
    'no double reservation',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;

      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ShellChrome(
              reservesTrafficLightGutter: true,
              child: Scaffold(
                appBar: DesktopAppBar(title: Text('Title')),
                body: SizedBox(),
              ),
            ),
          ),
        ),
      );
      await tester.pump();

      // Still draggable, still no self-drawn caption buttons on macOS — only
      // the gutter itself is what a full-height sidebar ancestor removes.
      expect(find.byType(DragToMoveArea), findsOneWidget);
      expect(find.byType(WindowCaptionButton), findsNothing);
      expect(
        _gutterPadding,
        findsNothing,
        reason:
            'The ShellChrome ancestor (the desktop shell\'s sidebar) already '
            'reserves this space; double-reserving it is the v5.30.0 field '
            'report bug this fixes',
      );

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

  // [leading] defaults to null — the harness's home route can't pop, and
  // with no explicit leading widget DesktopSliverAppBar's own
  // automaticallyImplyLeading resolves leadingContent to null, meaning the
  // gutter logic has nothing to pad in the first place. The v5.30.5 gutter
  // tests below pass an explicit leading precisely so there is a widget for
  // the gutter to wrap (or not wrap).
  Widget sliverHarness({
    List<Override> overrides = const [],
    Widget? leading,
  }) => ProviderScope(
    overrides: overrides,
    child: MaterialApp(
      home: Scaffold(
        body: CustomScrollView(
          slivers: [
            DesktopSliverAppBar(
              title: const Text('Album'),
              leading: leading,
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
    'DesktopSliverAppBar macOS with explicit leading: gutter padding present '
    'without a ShellChrome ancestor',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;

      await tester.pumpWidget(sliverHarness(leading: const Icon(Icons.close)));
      await tester.pump();

      expect(_gutterPadding, findsOneWidget);

      debugDefaultTargetPlatformOverride = null;
    },
  );

  testWidgets(
    'DesktopSliverAppBar macOS with explicit leading, under a ShellChrome '
    'that reserves the gutter: no double reservation',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;

      await tester.pumpWidget(
        ShellChrome(
          reservesTrafficLightGutter: true,
          child: sliverHarness(leading: const Icon(Icons.close)),
        ),
      );
      await tester.pump();

      expect(_gutterPadding, findsNothing);

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
