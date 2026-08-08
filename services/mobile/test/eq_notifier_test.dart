import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/audio/eq_notifier.dart';
import 'package:inori_music/src/audio/eq_settings.dart';
import 'package:inori_music/src/playback/playback_engine.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';

import 'support/fake_playback_engine.dart';

// ---------------------------------------------------------------------------
// EqNotifier tests.
//
// Since v5.27.0 EqNotifier asks the playback engine for an equalizer instead
// of testing Platform.isAndroid, so these inject a fake engine. That is what
// makes the "an equalizer exists" path testable at all — previously it could
// only run on a real Android device, and every test here was really asserting
// that the test host is not Android.
// ---------------------------------------------------------------------------

/// Container whose engine has no equalizer — the desktop/iOS shape.
ProviderContainer _containerWithoutEq() => ProviderContainer(
  overrides: [playbackEngineProvider.overrideWithValue(FakePlaybackEngine())],
);

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  group('EqNotifier custom presets', () {
    test(
      'saveCurrentAsPreset stores bands under the given name and selects it',
      () async {
        final container = _containerWithoutEq();
        addTearDown(container.dispose);
        final notifier = container.read(eqNotifierProvider.notifier);

        await notifier.setBand(0, 5.0);
        await notifier.saveCurrentAsPreset('My Preset');

        final state = container.read(eqNotifierProvider);
        expect(state.customPresets.containsKey('My Preset'), isTrue);
        expect(state.customPresets['My Preset']![0], 5.0);
        expect(state.preset, 'My Preset');
      },
    );

    test(
      'saveCurrentAsPreset trims whitespace and ignores empty names',
      () async {
        final container = _containerWithoutEq();
        addTearDown(container.dispose);
        final notifier = container.read(eqNotifierProvider.notifier);

        await notifier.saveCurrentAsPreset('  Spaced  ');
        expect(
          container
              .read(eqNotifierProvider)
              .customPresets
              .containsKey('Spaced'),
          isTrue,
        );

        final before = container.read(eqNotifierProvider).customPresets.length;
        await notifier.saveCurrentAsPreset('   ');
        expect(container.read(eqNotifierProvider).customPresets.length, before);
      },
    );

    test(
      'selectCustomPreset switches active bands to the saved preset',
      () async {
        final container = _containerWithoutEq();
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
      },
    );

    test('selectCustomPreset is a no-op for an unknown name', () async {
      final container = _containerWithoutEq();
      addTearDown(container.dispose);
      final notifier = container.read(eqNotifierProvider.notifier);

      final before = container.read(eqNotifierProvider);
      await notifier.selectCustomPreset('does-not-exist');
      final after = container.read(eqNotifierProvider);
      expect(after.preset, before.preset);
      expect(after.bands, before.bands);
    });

    test(
      'deleteCustomPreset removes the preset and falls back to flat when selected',
      () async {
        final container = _containerWithoutEq();
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
      },
    );

    test(
      'deleteCustomPreset does not disturb selection when a different preset is active',
      () async {
        final container = _containerWithoutEq();
        addTearDown(container.dispose);
        final notifier = container.read(eqNotifierProvider.notifier);

        await notifier.saveCurrentAsPreset('A');
        await notifier.setPreset('vocal');
        await notifier.deleteCustomPreset('A');

        final state = container.read(eqNotifierProvider);
        expect(state.customPresets.containsKey('A'), isFalse);
        expect(state.preset, 'vocal');
      },
    );

    test(
      'setEnabled(true) is a no-op when the engine has no equalizer',
      () async {
        final container = _containerWithoutEq();
        addTearDown(container.dispose);
        final notifier = container.read(eqNotifierProvider.notifier);

        await notifier.setEnabled(true);
        expect(container.read(eqNotifierProvider).enabled, isFalse);
      },
    );

    test(
      'setEnabled(false) is also a no-op with no equalizer, not just enable',
      () async {
        // Regression test: the guard used to be `if (enabled && !isAndroid)
        // return`, which only blocked the *enable* path — a disable call fell
        // through to the platform channel regardless, which is exactly the
        // unconditional-call pattern that threw MissingPluginException.
        final container = _containerWithoutEq();
        addTearDown(container.dispose);
        final notifier = container.read(eqNotifierProvider.notifier);

        await notifier.setEnabled(false);
        expect(container.read(eqNotifierProvider).enabled, isFalse);
      },
    );

    test(
      'persists custom presets across a fresh restore from SharedPreferences',
      () async {
        final container1 = _containerWithoutEq();
        final notifier1 = container1.read(eqNotifierProvider.notifier);
        await notifier1.setBand(3, 4.5);
        await notifier1.saveCurrentAsPreset('Saved');
        container1.dispose();

        final container2 = _containerWithoutEq();
        addTearDown(container2.dispose);
        // Reading the provider triggers build(), which calls _restore()
        // asynchronously; await a microtask turn for it to complete.
        container2.read(eqNotifierProvider);
        await Future<void>.delayed(Duration.zero);

        final restored = container2.read(eqNotifierProvider);
        expect(restored.customPresets.containsKey('Saved'), isTrue);
        expect(restored.customPresets['Saved']![3], 4.5);
      },
    );
  });

  // -------------------------------------------------------------------------
  // The path that used to be unreachable in tests
  // -------------------------------------------------------------------------

  group('EqNotifier with an equalizer present', () {
    late FakeEngineEqualizer eq;
    late ProviderContainer container;

    setUp(() {
      // Five device bands against the UI's ten: the real Android shape, and
      // the reason band application is a mapping rather than a copy.
      eq = FakeEngineEqualizer(bands: 5, min: -12, max: 12);
      container = ProviderContainer(
        overrides: [
          playbackEngineProvider.overrideWithValue(
            FakePlaybackEngine(
              capabilities: const PlaybackCapabilities(equalizer: true),
              equalizer: eq,
            ),
          ),
        ],
      );
      addTearDown(container.dispose);
    });

    test('enabling turns the effect on and pushes gains', () async {
      final notifier = container.read(eqNotifierProvider.notifier);
      await notifier.setPreset('bassBoost');
      await notifier.setEnabled(true);

      expect(container.read(eqNotifierProvider).enabled, isTrue);
      expect(eq.enabled, isTrue);
      expect(
        eq.gains.length,
        5,
        reason: 'One gain per device band, not per UI band',
      );
      expect(
        eq.gains[0],
        greaterThan(0),
        reason: 'bassBoost lifts the low bands',
      );
    });

    test('disabling flattens every band back to zero', () async {
      final notifier = container.read(eqNotifierProvider.notifier);
      await notifier.setPreset('bassBoost');
      await notifier.setEnabled(true);
      await notifier.setEnabled(false);

      expect(eq.enabled, isFalse);
      expect(eq.gains.values.every((g) => g == 0), isTrue);
    });

    test('gains are clamped to what the device accepts', () async {
      // 'bassBoost' asks for +6; a device with a narrower range must not be
      // handed a value outside it.
      final notifier = container.read(eqNotifierProvider.notifier);
      await notifier.setBand(0, 40);
      await notifier.setEnabled(true);

      expect(eq.gains.values.every((g) => g >= -12 && g <= 12), isTrue);
    });
  });
}
