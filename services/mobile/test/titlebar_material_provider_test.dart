import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/shared/titlebar_material_provider.dart';

void main() {
  group('TitlebarMaterial', () {
    test('has exactly the three values flutter_acrylic maps to', () {
      expect(TitlebarMaterial.values, [
        TitlebarMaterial.none,
        TitlebarMaterial.acrylic,
        TitlebarMaterial.mica,
      ]);
    });

    test(
      '.name round-trips for persistence (SharedPreferences stores by name)',
      () {
        for (final m in TitlebarMaterial.values) {
          final restored = TitlebarMaterial.values.firstWhere(
            (v) => v.name == m.name,
            orElse: () => TitlebarMaterial.none,
          );
          expect(restored, m);
        }
      },
    );

    test(
      'an unrecognized stored value falls back to none rather than throwing',
      () {
        final restored = TitlebarMaterial.values.firstWhere(
          (v) => v.name == 'some-old-removed-value',
          orElse: () => TitlebarMaterial.none,
        );
        expect(restored, TitlebarMaterial.none);
      },
    );
  });
}
