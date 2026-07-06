import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/audio/eq_notifier.dart';
import 'package:inori_music/src/audio/eq_settings.dart';

// ---------------------------------------------------------------------------
// EqNotifier custom preset tests.
//
// EqNotifier reads `audioHandler` (a global set in main.dart) only when
// `state.enabled` is true. These tests keep EQ disabled throughout, so the
// custom-preset save/select/delete/persist logic is exercised without ever
// touching the audio_service-backed handler.
// ---------------------------------------------------------------------------

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  group('EqNotifier custom presets', () {
    test('saveCurrentAsPreset stores bands under the given name and selects it', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(eqNotifierProvider.notifier);

      await notifier.setBand(0, 5.0);
      await notifier.saveCurrentAsPreset('My Preset');

      final state = container.read(eqNotifierProvider);
      expect(state.customPresets.containsKey('My Preset'), isTrue);
      expect(state.customPresets['My Preset']![0], 5.0);
      expect(state.preset, 'My Preset');
    });

    test('saveCurrentAsPreset trims whitespace and ignores empty names', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(eqNotifierProvider.notifier);

      await notifier.saveCurrentAsPreset('  Spaced  ');
      expect(container.read(eqNotifierProvider).customPresets.containsKey('Spaced'), isTrue);

      final before = container.read(eqNotifierProvider).customPresets.length;
      await notifier.saveCurrentAsPreset('   ');
      expect(container.read(eqNotifierProvider).customPresets.length, before);
    });

    test('selectCustomPreset switches active bands to the saved preset', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(eqNotifierProvider.notifier);

      await notifier.setBand(1, 7.0);
      await notifier.saveCurrentAsPreset('Loud');
      await notifier.setPreset('flat');
      expect(container.read(eqNotifierProvider).bands[1], 0.0);

      await notifier.selectCustomPreset('Loud');
      final state = container.read(eqNotifierProvider);
      expect(state.preset, 'Loud');
      expect(state.bands[1], 7.0);
    });

    test('selectCustomPreset is a no-op for an unknown name', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(eqNotifierProvider.notifier);

      final before = container.read(eqNotifierProvider);
      await notifier.selectCustomPreset('does-not-exist');
      final after = container.read(eqNotifierProvider);
      expect(after.preset, before.preset);
      expect(after.bands, before.bands);
    });

    test('deleteCustomPreset removes the preset and falls back to flat when selected', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(eqNotifierProvider.notifier);

      await notifier.setBand(2, 3.0);
      await notifier.saveCurrentAsPreset('Temp');
      expect(container.read(eqNotifierProvider).preset, 'Temp');

      await notifier.deleteCustomPreset('Temp');
      final state = container.read(eqNotifierProvider);
      expect(state.customPresets.containsKey('Temp'), isFalse);
      expect(state.preset, 'flat');
      expect(state.bands, List<double>.from(eqPresets['flat']!));
    });

    test('deleteCustomPreset does not disturb selection when a different preset is active', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(eqNotifierProvider.notifier);

      await notifier.saveCurrentAsPreset('A');
      await notifier.setPreset('vocal');
      await notifier.deleteCustomPreset('A');

      final state = container.read(eqNotifierProvider);
      expect(state.customPresets.containsKey('A'), isFalse);
      expect(state.preset, 'vocal');
    });

    test('persists custom presets across a fresh restore from SharedPreferences', () async {
      final container1 = ProviderContainer();
      final notifier1 = container1.read(eqNotifierProvider.notifier);
      await notifier1.setBand(3, 4.5);
      await notifier1.saveCurrentAsPreset('Saved');
      container1.dispose();

      final container2 = ProviderContainer();
      addTearDown(container2.dispose);
      // Reading the provider triggers build(), which calls _restore()
      // asynchronously; await a microtask turn for it to complete.
      container2.read(eqNotifierProvider);
      await Future<void>.delayed(Duration.zero);

      final restored = container2.read(eqNotifierProvider);
      expect(restored.customPresets.containsKey('Saved'), isTrue);
      expect(restored.customPresets['Saved']![3], 4.5);
    });
  });
}
