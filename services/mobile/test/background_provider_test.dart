import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/shared/background_provider.dart';

// ---------------------------------------------------------------------------
// Tests for BackgroundSettings — the pure value object. BackgroundNotifier's
// methods (pickImage/clearImage/setOpacity) all touch SharedPreferences/File
// I/O with no platform-channel mock available in this test harness (same
// constraint documented in download_notifier_test.dart), so the opacity
// clamp is verified as a mirrored pure function instead of through the real
// method — see setOpacity's `value.clamp(0.1, 0.8)`.
// ---------------------------------------------------------------------------

void main() {
  group('BackgroundSettings defaults', () {
    test('no image and default opacity when constructed with no args', () {
      const settings = BackgroundSettings();
      expect(settings.imagePath, isNull);
      expect(settings.opacity, 0.45);
    });
  });

  group('BackgroundSettings.copyWith', () {
    test('copyWith sets a new imagePath', () {
      const settings = BackgroundSettings();
      final updated = settings.copyWith(imagePath: '/tmp/bg.png');
      expect(updated.imagePath, '/tmp/bg.png');
      expect(updated.opacity, settings.opacity);
    });

    test('copyWith opacity leaves imagePath untouched', () {
      const settings = BackgroundSettings(imagePath: '/tmp/bg.png');
      final updated = settings.copyWith(opacity: 0.6);
      expect(updated.imagePath, '/tmp/bg.png');
      expect(updated.opacity, 0.6);
    });

    test('clearImage: true removes imagePath even if a new one is also passed', () {
      const settings = BackgroundSettings(imagePath: '/tmp/bg.png', opacity: 0.5);
      final cleared = settings.copyWith(imagePath: '/tmp/other.png', clearImage: true);
      expect(cleared.imagePath, isNull);
      expect(cleared.opacity, 0.5);
    });

    test('omitting clearImage preserves the existing imagePath', () {
      const settings = BackgroundSettings(imagePath: '/tmp/bg.png');
      final updated = settings.copyWith(opacity: 0.3);
      expect(updated.imagePath, '/tmp/bg.png');
    });
  });

  // Mirrors BackgroundNotifier.setOpacity's `value.clamp(0.1, 0.8)`.
  group('opacity clamp (mirrors BackgroundNotifier.setOpacity)', () {
    double clampOpacity(double value) => value.clamp(0.1, 0.8);

    test('values inside range pass through unchanged', () {
      expect(clampOpacity(0.45), 0.45);
    });

    test('values below the floor clamp to 0.1', () {
      expect(clampOpacity(0.0), 0.1);
    });

    test('values above the ceiling clamp to 0.8', () {
      expect(clampOpacity(1.0), 0.8);
    });
  });
}
