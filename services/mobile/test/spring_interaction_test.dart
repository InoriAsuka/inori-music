// spring_interaction_test.dart
//
// Covers the v5.25.0 spring hover/press wrapper. The property that matters
// most is the one that's easy to break: SpringInteraction must *observe*
// pointer events without joining the gesture arena, so wrapping an existing
// button doesn't cost that button its taps.
//
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/shared/widgets/spring_interaction.dart';

/// The wrapper builds exactly two nested Transforms: translate outside,
/// scale inside. Scoped to the wrapper's own subtree — a bare
/// `find.byType(Transform)` also picks up whatever MaterialApp and Scaffold
/// put in the tree.
List<Transform> _transforms(WidgetTester tester) => tester
    .widgetList<Transform>(
      find.descendant(
        of: find.byType(SpringInteraction),
        matching: find.byType(Transform),
      ),
    )
    .toList();

/// Current visual scale applied by the wrapper, read off the rendered
/// transform rather than any internal state.
///
/// Reads m00 directly rather than going through `getMaxScaleOnAxis()`: that
/// helper takes the largest scale across all three axes, and Transform.scale
/// leaves z at 1, so it reports 1.0 for every shrink.
double scaleOf(WidgetTester tester) =>
    _transforms(tester).last.transform.storage[0];

double liftOf(WidgetTester tester) =>
    -_transforms(tester).first.transform.getTranslation().y;

const _targetKey = Key('spring-target');

Widget _app({VoidCallback? onPressed}) => MaterialApp(
  home: Scaffold(
    body: Center(
      child: SpringInteraction(
        child: onPressed == null
            ? const SizedBox(key: _targetKey, width: 48, height: 48)
            : IconButton(
                key: _targetKey,
                icon: const Icon(Icons.play_arrow),
                onPressed: onPressed,
              ),
      ),
    ),
  ),
);

Offset _targetCenter(WidgetTester tester) =>
    tester.getCenter(find.byKey(_targetKey));

void main() {
  testWidgets('the wrapped button still receives taps', (tester) async {
    var taps = 0;
    await tester.pumpWidget(_app(onPressed: () => taps++));

    await tester.tap(find.byKey(_targetKey));
    await tester.pumpAndSettle();

    expect(taps, 1, reason: 'The wrapper must not swallow the tap');
  });

  testWidgets('pressing scales the child down and releasing springs it back', (
    tester,
  ) async {
    await tester.pumpWidget(_app());

    expect(scaleOf(tester), closeTo(1.0, 0.001));

    final gesture = await tester.startGesture(_targetCenter(tester));
    // Part-way through the spring, not settled: an under-damped spring
    // overshoots, so asserting on a specific value here would be brittle —
    // what matters is that it is moving toward the pressed scale.
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 60));
    expect(scaleOf(tester), lessThan(1.0));

    await gesture.up();
    await tester.pumpAndSettle();
    expect(
      scaleOf(tester),
      closeTo(1.0, 0.01),
      reason: 'Release must settle back to rest',
    );
  });

  testWidgets('hovering lifts the child and leaving lowers it', (tester) async {
    await tester.pumpWidget(_app());

    expect(liftOf(tester), closeTo(0, 0.001));

    final mouse = await tester.createGesture(kind: PointerDeviceKind.mouse);
    await mouse.addPointer(location: Offset.zero);
    addTearDown(mouse.removePointer);
    await mouse.moveTo(_targetCenter(tester));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 60));

    expect(liftOf(tester), greaterThan(0));

    await mouse.moveTo(const Offset(-200, -200));
    await tester.pumpAndSettle();
    expect(
      liftOf(tester),
      closeTo(0, 0.05),
      reason: 'Leaving must settle back to rest',
    );
  });

  testWidgets('a cancelled press settles back to rest', (tester) async {
    // Scrolling away mid-press cancels the pointer; without handling
    // onPointerCancel the control would stay visually stuck at pressed scale.
    await tester.pumpWidget(_app());

    final gesture = await tester.startGesture(_targetCenter(tester));
    await tester.pump(const Duration(milliseconds: 60));
    await gesture.cancel();
    await tester.pumpAndSettle();

    expect(scaleOf(tester), closeTo(1.0, 0.01));
  });
}
