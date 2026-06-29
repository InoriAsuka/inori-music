import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _kReplayGainKey = 'audio.replayGain';

final replayGainEnabledProvider = NotifierProvider<ReplayGainNotifier, bool>(
  ReplayGainNotifier.new,
);

/// Persists the ReplayGain enabled/disabled toggle.
/// Actual gain application happens in PlayerNotifier.playTrack when it
/// reads this provider and the track's replayGainDb field.
class ReplayGainNotifier extends Notifier<bool> {
  @override
  bool build() {
    _restore();
    return false;
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    state = prefs.getBool(_kReplayGainKey) ?? false;
  }

  Future<void> setEnabled(bool enabled) async {
    state = enabled;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_kReplayGainKey, enabled);
  }
}
