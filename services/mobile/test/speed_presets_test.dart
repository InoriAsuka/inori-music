import 'package:flutter_test/flutter_test.dart';
import 'package:inori_music/src/audio/speed_notifier.dart';

/// Mirrors the `SPEED_PRESETS` block in
/// `services/web/lib/player/controls.test.ts`. Both suites assert the same
/// literal tier list, so a change on either client fails its own tests until
/// the other is updated too — the EQ presets drifted precisely because no
/// such assertion existed.
void main() {
  group('speedPresets', () {
    test('is exactly the agreed tier list, web-aligned', () {
      expect(speedPresets, [0.5, 0.75, 1.0, 1.25, 1.5, 2.0]);
    });

    test('stays within the clamp bounds and includes the neutral rate', () {
      for (final preset in speedPresets) {
        expect(preset, greaterThanOrEqualTo(minSpeed));
        expect(preset, lessThanOrEqualTo(maxSpeed));
      }
      expect(speedPresets, contains(1.0));
    });

    test('is sorted ascending with no duplicates', () {
      final sorted = [...speedPresets]..sort();
      expect(speedPresets, sorted);
      expect(speedPresets.toSet().length, speedPresets.length);
    });
  });
}
