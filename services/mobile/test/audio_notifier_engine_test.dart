// audio_notifier_engine_test.dart
//
// The three small audio notifiers used to reach a global in main.dart, which
// meant none of them could be tested without booting the real audio stack.
// Since v5.27.0 they take the engine from a provider, so these assert the
// wiring that replaced the singleton.
//
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/audio/crossfade_notifier.dart';
import 'package:inori_music/src/audio/speed_notifier.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';

import 'support/fake_playback_engine.dart';

void main() {
  late FakePlaybackEngine engine;
  late ProviderContainer container;

  setUp(() {
    SharedPreferences.setMockInitialValues({});
    engine = FakePlaybackEngine();
    container = ProviderContainer(
      overrides: [playbackEngineProvider.overrideWithValue(engine)],
    );
    addTearDown(container.dispose);
  });

  test('speed changes reach the engine', () async {
    await container.read(speedNotifierProvider.notifier).setSpeed(1.5);

    expect(container.read(speedNotifierProvider), 1.5);
    expect(engine.lastSpeed, 1.5);
  });

  test(
    'an out-of-range speed is clamped before it reaches the engine',
    () async {
      await container.read(speedNotifierProvider.notifier).setSpeed(3.7);

      expect(engine.lastSpeed, maxSpeed);
      expect(
        container.read(speedNotifierProvider),
        maxSpeed,
        reason: 'State and engine must not disagree about the applied speed',
      );
    },
  );

  test('crossfade duration reaches the engine', () async {
    await container.read(crossfadeProvider.notifier).setSeconds(5);

    expect(container.read(crossfadeProvider), 5);
    expect(engine.crossfadeSecondsSet, 5);
  });

  test('crossfade duration is clamped before it reaches the engine', () async {
    await container.read(crossfadeProvider.notifier).setSeconds(99);

    expect(engine.crossfadeSecondsSet, 8);
    expect(container.read(crossfadeProvider), 8);
  });
}
