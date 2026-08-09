import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _kCoverFlowModeKey = 'player.coverFlowArtwork';

final coverFlowModeProvider = NotifierProvider<CoverFlowModeNotifier, bool>(
  CoverFlowModeNotifier.new,
);

/// Persists the Cover Flow artwork display mode toggle (Settings →
/// Appearance). Off by default — [FullPlayerScreen] then keeps rendering the
/// single-cover tile it always has. Mirrors [BilingualLyricsNotifier]'s exact
/// shape: this is a plain persisted boolean, not a whole settings subsystem.
class CoverFlowModeNotifier extends Notifier<bool> {
  @override
  bool build() {
    _restore();
    return false;
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    state = prefs.getBool(_kCoverFlowModeKey) ?? false;
  }

  Future<void> setEnabled(bool enabled) async {
    state = enabled;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_kCoverFlowModeKey, enabled);
  }
}
