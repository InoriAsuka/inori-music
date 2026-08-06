import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Exported so [DesktopIntegration] can read the persisted value directly
/// (via SharedPreferences, not this provider) during the synchronous-feeling
/// window bootstrap in `_initWindow()`, which runs before this Notifier's own
/// async `_restore()` would otherwise resolve.
const kSystemTitleBarKey = 'appearance.systemTitleBar';

final systemTitleBarProvider = NotifierProvider<SystemTitleBarNotifier, bool>(
  SystemTitleBarNotifier.new,
);

/// Whether the OS's native title bar is used instead of [DesktopAppBar]'s
/// custom drag region + window buttons. Desktop-only; defaults to the
/// custom chrome (false).
class SystemTitleBarNotifier extends Notifier<bool> {
  @override
  bool build() {
    _restore();
    return false;
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    state = prefs.getBool(kSystemTitleBarKey) ?? false;
  }

  Future<void> setEnabled(bool enabled) async {
    state = enabled;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(kSystemTitleBarKey, enabled);
  }
}
