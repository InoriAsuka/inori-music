import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/main.dart' show audioHandler;

const _kCrossfadeKey = 'audio.crossfade';

final crossfadeProvider = NotifierProvider<CrossfadeNotifier, int>(
  CrossfadeNotifier.new,
);

/// Persists and exposes the crossfade duration in seconds (0–8).
/// 0 means crossfade is disabled.
class CrossfadeNotifier extends Notifier<int> {
  @override
  int build() {
    _restore();
    return 0;
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    final saved = prefs.getInt(_kCrossfadeKey) ?? 0;
    final clamped = saved.clamp(0, 8);
    state = clamped;
    audioHandler.crossfadeSeconds = clamped;
  }

  Future<void> setSeconds(int seconds) async {
    final clamped = seconds.clamp(0, 8);
    state = clamped;
    audioHandler.crossfadeSeconds = clamped;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setInt(_kCrossfadeKey, clamped);
  }
}
