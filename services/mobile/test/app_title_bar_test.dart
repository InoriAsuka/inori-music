import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:window_manager/window_manager.dart';

import 'package:inori_music/src/shared/widgets/app_title_bar.dart';

// ---------------------------------------------------------------------------
// AppTitleBar reads defaultTargetPlatform to decide macOS (native traffic
// lights, no self-drawn buttons) vs. everything else (three window_manager
// WindowCaptionButtons). debugDefaultTargetPlatformOverride simulates each
// platform without needing a real desktop window — the widget never taps a
// button in these tests, so it never reaches window_manager's platform
// channel (unavailable in the test harness; isMaximized() is wrapped in a
// catchError in the widget itself for exactly this scenario).
//
// The binding's end-of-test invariant check (debugAssertAllFoundationVarsUnset)
// runs immediately after the test body returns, before addTearDown/tearDown
// callbacks fire — so the override must be reset as the last synchronous
// step of the test body itself, not in a tearDown block.
// ---------------------------------------------------------------------------

void main() {
  testWidgets('macOS: no caption buttons, left space reserved for traffic lights', (tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.macOS;

    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(home: Scaffold(body: AppTitleBar())),
      ),
    );
    await tester.pump();

    expect(find.byType(WindowCaptionButton), findsNothing);
    expect(find.byType(DragToMoveArea), findsOneWidget);
    expect(tester.getSize(find.byType(AppTitleBar)).height, AppTitleBar.macHeight);

    debugDefaultTargetPlatformOverride = null;
  });

  testWidgets('Windows: three caption buttons (minimize/maximize/close)', (tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.windows;

    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(home: Scaffold(body: AppTitleBar())),
      ),
    );
    await tester.pump();

    expect(find.byType(WindowCaptionButton), findsNWidgets(3));
    expect(find.byType(DragToMoveArea), findsOneWidget);

    debugDefaultTargetPlatformOverride = null;
  });

  testWidgets('shows a well-formed widget even when isMaximized() has no platform channel', (tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.windows;

    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(home: Scaffold(body: AppTitleBar())),
      ),
    );
    await tester.pump();
    await tester.pump();

    // isMaximized() has no platform channel in the test harness and fails
    // silently (see catchError in AppTitleBar) — _isMaximized stays at its
    // default (false), so the widget should still be in a well-formed state
    // rather than stuck mid-async or throwing.
    expect(find.byType(AppTitleBar), findsOneWidget);
    expect(tester.takeException(), isNull);

    debugDefaultTargetPlatformOverride = null;
  });
}
