import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:window_manager/window_manager.dart';

import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Custom desktop window chrome — replaces the OS title bar (hidden via
/// `windowManager.setTitleBarStyle` in [DesktopIntegration]) with a thin
/// draggable strip whose background follows the active skin.
///
/// Per the platform split the user chose: Windows gets three self-drawn
/// caption buttons (window_manager's own [WindowCaptionButton] — custom-drawn
/// by the package, not OS chrome, so this still satisfies "no system Windows
/// frame"); macOS keeps the native traffic-light buttons (Apple HIG
/// convention) and this strip only supplies the draggable, skin-colored
/// backdrop behind them. No logo/wordmark here — `ShellScaffold`'s sidebar
/// and the login/splash screens already show the brand mark; repeating it
/// in a 28–32px strip above everything would just be visual noise.
class AppTitleBar extends ConsumerStatefulWidget {
  const AppTitleBar({super.key});

  static const double windowsHeight = 32;
  static const double macHeight = 28;

  @override
  ConsumerState<AppTitleBar> createState() => _AppTitleBarState();
}

class _AppTitleBarState extends ConsumerState<AppTitleBar> with WindowListener {
  bool _isMaximized = false;

  @override
  void initState() {
    super.initState();
    windowManager.addListener(this);
    windowManager.isMaximized().then((v) {
      if (mounted) setState(() => _isMaximized = v);
    }).catchError((_) {
      // No native window (e.g. widget pumped directly in a test harness with
      // no window_manager platform channel available) — keep the default.
    });
  }

  @override
  void dispose() {
    windowManager.removeListener(this);
    super.dispose();
  }

  @override
  void onWindowMaximize() => setState(() => _isMaximized = true);

  @override
  void onWindowUnmaximize() => setState(() => _isMaximized = false);

  @override
  Widget build(BuildContext context) {
    final skin = ref.watch(skinProvider).active;
    final isMac = defaultTargetPlatform == TargetPlatform.macOS;

    return Container(
      height: isMac ? AppTitleBar.macHeight : AppTitleBar.windowsHeight,
      color: skin.colors.playerBar,
      child: Row(
        children: [
          // On macOS this reserves space for the native traffic lights
          // (drawn by the OS above everything Flutter renders, but leaving
          // the area visually empty here avoids the drag strip looking like
          // it's competing with them).
          if (isMac) const SizedBox(width: 78),
          const Expanded(child: DragToMoveArea(child: SizedBox.expand())),
          if (!isMac) ...[
            WindowCaptionButton.minimize(
              brightness: skin.brightness,
              onPressed: () => windowManager.minimize(),
            ),
            _isMaximized
                ? WindowCaptionButton.unmaximize(
                    brightness: skin.brightness,
                    onPressed: () => windowManager.unmaximize(),
                  )
                : WindowCaptionButton.maximize(
                    brightness: skin.brightness,
                    onPressed: () => windowManager.maximize(),
                  ),
            WindowCaptionButton.close(
              brightness: skin.brightness,
              onPressed: () => windowManager.close(),
            ),
          ],
        ],
      ),
    );
  }
}
