// ignore_for_file: implementation_imports
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter/widgets.dart' show WidgetsBinding;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tray_manager/tray_manager.dart';
import 'package:hotkey_manager/hotkey_manager.dart';
import 'package:window_manager/window_manager.dart';

import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/player/player_notifier.dart';

/// Initialises system-tray, global hotkeys, and window sizing on
/// macOS / Windows / Linux.
///
/// Usage — call [DesktopIntegration.init] once after [ProviderScope] is ready,
/// e.g. inside an [AppLifecycleListener] or a [ConsumerStatefulWidget.initState].
/// Call [dispose] on app detach to release OS resources.
class DesktopIntegration with TrayListener {
  DesktopIntegration(this._ref);
  final WidgetRef _ref;

  static bool get isDesktop =>
      defaultTargetPlatform == TargetPlatform.macOS ||
      defaultTargetPlatform == TargetPlatform.windows ||
      defaultTargetPlatform == TargetPlatform.linux;

  // Gate screens (login / future splash) are a small fixed-proportion window,
  // matching how most desktop music players treat auth as a compact dialog
  // rather than a full browsing canvas; the main shell gets a wide default
  // that suits multi-column library browsing. Ratios mirror the reference
  // mockups from v5.12.2 (443x727 gate, 2042x1191 main).
  static const _gateWindowSize = Size(440, 720);
  static const _mainWindowSize = Size(1440, 840);
  static const _mainMinimumSize = Size(960, 600);

  Future<void> init() async {
    if (!isDesktop) return;
    await _initWindow();
    await _initTray();
    await _initHotkeys();
  }

  Future<void> _initWindow() async {
    await windowManager.ensureInitialized();
    final isPastGate = _ref.read(authProvider).valueOrNull?.isPastGate ?? false;
    await windowManager.waitUntilReadyToShow(
      WindowOptions(
        size: isPastGate ? _mainWindowSize : _gateWindowSize,
        minimumSize: isPastGate ? _mainMinimumSize : _gateWindowSize,
        center: true,
      ),
      () async {
        await windowManager.setResizable(isPastGate);
        await windowManager.show();
        await windowManager.focus();
      },
    );
  }

  /// Resizes the window when crossing the login gate in either direction.
  /// Called from a `ref.listen(authProvider, ...)` registered in
  /// [InoriMusicApp]'s build method — [WidgetRef.listen] isn't valid outside
  /// a widget build, so the subscription itself can't live here.
  Future<void> applyWindowForAuthState(bool isPastGate) async {
    if (!isDesktop) return;
    await windowManager.setResizable(isPastGate);
    await windowManager.setMinimumSize(isPastGate ? _mainMinimumSize : _gateWindowSize);
    await windowManager.setSize(isPastGate ? _mainWindowSize : _gateWindowSize, animate: true);
    await windowManager.center();
  }

  Future<void> dispose() async {
    if (!isDesktop) return;
    trayManager.removeListener(this);
    await hotKeyManager.unregisterAll();
  }

  // ---------------------------------------------------------------------------
  // Tray
  // ---------------------------------------------------------------------------

  Future<void> _initTray() async {
    await trayManager.setIcon('assets/images/tray_icon.png');
    await trayManager.setContextMenu(Menu(items: [
      MenuItem(key: 'play_pause', label: 'Play / Pause'),
      MenuItem(key: 'next', label: 'Next'),
      MenuItem(key: 'previous', label: 'Previous'),
      MenuItem.separator(),
      MenuItem(key: 'quit', label: 'Quit'),
    ]));
    trayManager.addListener(this);
  }

  @override
  void onTrayMenuItemClick(MenuItem menuItem) {
    final notifier = _ref.read(playerProvider.notifier);
    switch (menuItem.key) {
      case 'play_pause':
        notifier.togglePlayPause();
      case 'next':
        notifier.next();
      case 'previous':
        notifier.previous();
      case 'quit':
        // Use Flutter's graceful exit path so that dispose() callbacks,
        // SQLite WAL checkpoints, and audio-session teardown all run before
        // the process terminates.  dart:io exit(0) would bypass all of this.
        WidgetsBinding.instance.handleRequestAppExit();
    }
  }

  // ---------------------------------------------------------------------------
  // Global hotkeys
  // ---------------------------------------------------------------------------

  Future<void> _initHotkeys() async {
    await hotKeyManager.unregisterAll();

    // Alt+Space → togglePlayPause
    await hotKeyManager.register(
      HotKey(
        key: PhysicalKeyboardKey.space,
        modifiers: [HotKeyModifier.alt],
        scope: HotKeyScope.system,
      ),
      keyDownHandler: (_) => _ref.read(playerProvider.notifier).togglePlayPause(),
    );

    // Alt+Right → next
    await hotKeyManager.register(
      HotKey(
        key: PhysicalKeyboardKey.arrowRight,
        modifiers: [HotKeyModifier.alt],
        scope: HotKeyScope.system,
      ),
      keyDownHandler: (_) => _ref.read(playerProvider.notifier).next(),
    );

    // Alt+Left → previous
    await hotKeyManager.register(
      HotKey(
        key: PhysicalKeyboardKey.arrowLeft,
        modifiers: [HotKeyModifier.alt],
        scope: HotKeyScope.system,
      ),
      keyDownHandler: (_) => _ref.read(playerProvider.notifier).previous(),
    );
  }
}
