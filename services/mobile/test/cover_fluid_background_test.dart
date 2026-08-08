// cover_fluid_background_test.dart
//
// Covers the v5.26.0 cover-derived backdrop. The properties worth locking
// down are the structural ones — four quadrant tiles, a blur over them, and a
// clean degrade to a flat colour when there is no artwork — since the visual
// result itself can only be judged on a real device.
//
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/shared/widgets/cover_fluid_background.dart';

/// A 1x1 transparent PNG, so the tiles have something real to lay out without
/// touching the network or the file system.
final _pixel = MemoryImage(
  Uint8List.fromList(const [
    0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, //
    0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
    0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
    0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
    0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
    0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
    0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
    0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
    0x42, 0x60, 0x82,
  ]),
);

const _fallback = Color(0xFF123456);

Widget _app({required ImageProvider? image}) => MaterialApp(
  home: CoverFluidBackground(
    image: image,
    fallbackColor: _fallback,
    child: const Text('content', textDirection: TextDirection.ltr),
  ),
);

void main() {
  testWidgets('with no artwork it is a flat colour and nothing animates', (
    tester,
  ) async {
    await tester.pumpWidget(_app(image: null));
    // pumpAndSettle would hang if any controller were driving something, so
    // reaching this line at all is part of the assertion.
    await tester.pumpAndSettle();

    expect(find.byType(Image), findsNothing);
    expect(
      find.byType(BackdropFilter),
      findsNothing,
      reason: 'No point paying for a 64-sigma blur over a solid colour',
    );
    expect(find.text('content'), findsOneWidget);
  });

  testWidgets('a cover produces four quadrant tiles under one blur', (
    tester,
  ) async {
    await tester.pumpWidget(_app(image: _pixel));
    await tester.pump();

    expect(
      find.byType(Image),
      findsNWidgets(4),
      reason: "One tile per quadrant of the cover — that's the whole technique",
    );
    expect(find.byType(BackdropFilter), findsOneWidget);
    expect(find.text('content'), findsOneWidget);
  });

  testWidgets('each tile is offset and rotated differently', (tester) async {
    await tester.pumpWidget(_app(image: _pixel));
    await tester.pump(const Duration(seconds: 5));

    final rotations = tester
        .widgetList<Transform>(find.byType(Transform))
        .map((t) => t.transform.storage[0]) // cos(angle) of each rotation
        .toSet();
    expect(
      rotations.length,
      greaterThan(1),
      reason: 'In lockstep the four tiles read as one rigid shape turning',
    );

    // Tiles are laid out around the centre rather than stacked on it.
    final centres = tester
        .widgetList<Image>(find.byType(Image))
        .map((image) => tester.getCenter(find.byWidget(image)))
        .toSet();
    expect(centres.length, 4);
  });

  testWidgets('swapping the cover away tears the animation down cleanly', (
    tester,
  ) async {
    await tester.pumpWidget(_app(image: _pixel));
    await tester.pump();
    await tester.pumpWidget(_app(image: null));
    // A leaked controller would surface here as a pending-timer failure.
    await tester.pumpAndSettle();

    expect(find.byType(Image), findsNothing);
  });
}
