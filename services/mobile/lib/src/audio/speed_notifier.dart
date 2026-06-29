import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/main.dart' show audioHandler;

const _kSpeedKey = 'audio.speed';

final speedNotifierProvider = NotifierProvider<SpeedNotifier, double>(
  SpeedNotifier.new,
);

/// Persists and applies the playback speed [0.5–2.0].
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
    final clamped = speed.clamp(0.5, 2.0);
    state = clamped;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setDouble(_kSpeedKey, clamped);
    _apply(clamped);
  }

  void _apply(double speed) {
    audioHandler.audioPlayer.setSpeed(speed);
  }
}
