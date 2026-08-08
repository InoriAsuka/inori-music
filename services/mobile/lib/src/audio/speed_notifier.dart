import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/playback/playback_engine_provider.dart';

const _kSpeedKey = 'audio.speed';

/// Playback-speed presets offered in the UI.
///
/// Must stay in sync with `SPEED_PRESETS` in
/// `services/web/lib/player/controls.ts` so both clients offer the same tiers.
/// Declared once here rather than inline at each sheet — it was previously
/// duplicated across the full player and settings screens, which is exactly
/// how the EQ presets drifted apart from the web ones.
const speedPresets = <double>[0.5, 0.75, 1.0, 1.25, 1.5, 2.0];

/// Inclusive bounds any speed is clamped to, matching the web store's
/// MIN/MAX_PLAYBACK_SPEED.
const minSpeed = 0.5;
const maxSpeed = 2.0;

final speedNotifierProvider = NotifierProvider<SpeedNotifier, double>(
  SpeedNotifier.new,
);

/// Persists and applies the playback speed [0.5–2.0].
///
/// Pitch is preserved at every rate: just_audio does so by default, and the
/// web client sets `preservesPitch` explicitly to match.
class SpeedNotifier extends Notifier<double> {
  @override
  double build() {
    _restore();
    return 1.0;
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    final saved = prefs.getDouble(_kSpeedKey) ?? 1.0;
    if (saved != state) {
      state = saved;
      _apply(saved);
    }
  }

  Future<void> setSpeed(double speed) async {
    final clamped = speed.clamp(minSpeed, maxSpeed);
    state = clamped;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setDouble(_kSpeedKey, clamped);
    _apply(clamped);
  }

  void _apply(double speed) {
    ref.read(playbackEngineProvider).setSpeed(speed);
  }
}
