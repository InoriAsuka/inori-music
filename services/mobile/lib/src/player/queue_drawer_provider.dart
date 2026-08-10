import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Whether the desktop shell's docked queue drawer (`queue_drawer.dart`) is
/// currently open.
///
/// Not persisted — unlike `coverFlowModeProvider`/`systemTitleBarProvider`,
/// this is transient UI state scoped to the current session (closing the
/// drawer and relaunching the app should not reopen it), so a plain
/// in-memory [Notifier] is the whole story here; there is no SharedPreferences
/// round trip to wire up.
final queueDrawerOpenProvider = NotifierProvider<QueueDrawerNotifier, bool>(
  QueueDrawerNotifier.new,
);

/// Toggled by [MiniPlayerBar]'s own queue button (desktop shell only — see
/// that button's `_openQueue` doc comment) and watched by `QueueDrawer`
/// (shell_scaffold.dart's desktop content column) to decide whether it is
/// slid into view.
class QueueDrawerNotifier extends Notifier<bool> {
  @override
  bool build() => false;

  void open() => state = true;

  void close() => state = false;

  void toggle() => state = !state;
}
