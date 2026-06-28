// ignore_for_file: implementation_imports
import 'dart:io' show exit;

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tray_manager/tray_manager.dart';
import 'package:hotkey_manager/hotkey_manager.dart';

import 'package:inori_music/src/player/player_notifier.dart';

/// Initialises system-tray and global hotkeys on macOS / Windows / Linux.
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

  Future<void> init() async {
    if (!isDesktop) return;
    await _initTray();
    await _initHotkeys();
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
        exit(0);
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
