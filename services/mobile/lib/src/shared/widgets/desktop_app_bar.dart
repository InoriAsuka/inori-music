import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:window_manager/window_manager.dart';

import 'package:inori_music/src/shared/desktop_integration.dart';
import 'package:inori_music/src/shared/system_titlebar_provider.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Drop-in [AppBar] replacement that doubles as the window's draggable
/// title-bar region on desktop, replacing the standalone strip that shipped
/// in v5.15.0 — that approach stacked a second, content-free toolbar above
/// every screen's own AppBar. Spotube (github.com/KRTirtho/spotube) folds
/// the two into one component instead; this mirrors that structure.
///
/// On mobile, or when the user has opted into [systemTitleBarProvider]'s
/// "use the OS title bar" setting, this behaves exactly like a plain
/// [AppBar] — no drag handling, no extra buttons.
class DesktopAppBar extends StatelessWidget implements PreferredSizeWidget {
  const DesktopAppBar({
    super.key,
    this.title,
    this.actions,
    this.leading,
    this.automaticallyImplyLeading = true,
    this.bottom,
  });

  final Widget? title;
  final List<Widget>? actions;
  final Widget? leading;
  final bool automaticallyImplyLeading;
  final PreferredSizeWidget? bottom;

  @override
  Size get preferredSize =>
      Size.fromHeight(kToolbarHeight + (bottom?.preferredSize.height ?? 0));

  AppBar _plainAppBar() => AppBar(
    title: title,
    actions: actions,
    leading: leading,
    automaticallyImplyLeading: automaticallyImplyLeading,
    bottom: bottom,
  );

  @override
  Widget build(BuildContext context) {
    if (!DesktopIntegration.isDesktop) return _plainAppBar();

    return Consumer(
      builder: (context, ref, _) {
        final useSystemTitleBar = ref.watch(systemTitleBarProvider);
        if (useSystemTitleBar) return _plainAppBar();

        final skin = ref.watch(skinProvider).active;
        final isMac = defaultTargetPlatform == TargetPlatform.macOS;

        final appBar = AppBar(
          title: title,
          actions: [
            ...?actions,
            if (!isMac) ...[
              const SizedBox(width: 8),
              WindowCaptionButton.minimize(
                brightness: skin.brightness,
                onPressed: () => windowManager.minimize(),
              ),
              _MaximizeButton(brightness: skin.brightness),
              WindowCaptionButton.close(
                brightness: skin.brightness,
                onPressed: () => windowManager.close(),
              ),
            ],
          ],
          leading: leading,
          automaticallyImplyLeading: automaticallyImplyLeading,
          bottom: bottom,
        );

        // Reserve space for the native traffic lights (macOS draws them
        // above whatever Flutter renders, at a fixed offset from the
        // window's top-left) rather than letting the title/leading content
        // start flush against them.
        final withMacGutter = isMac
            ? Padding(padding: const EdgeInsets.only(left: 70), child: appBar)
            : appBar;

        return DragToMoveArea(child: withMacGutter);
      },
    );
  }
}

/// Toggles between maximize/restore icons based on live window state —
/// window_manager has no synchronous "is maximized" getter, so this tracks
/// it via [WindowListener] the same way the old standalone title bar did.
class _MaximizeButton extends StatefulWidget {
  const _MaximizeButton({required this.brightness});
  final Brightness brightness;

  @override
  State<_MaximizeButton> createState() => _MaximizeButtonState();
}

class _MaximizeButtonState extends State<_MaximizeButton> with WindowListener {
  bool _isMaximized = false;

  @override
  void initState() {
    super.initState();
    windowManager.addListener(this);
    windowManager
        .isMaximized()
        .then((v) {
          if (mounted) setState(() => _isMaximized = v);
        })
        .catchError((_) {
          // No native window (e.g. widget pumped directly in a test harness) —
          // keep the default.
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
    return _isMaximized
        ? WindowCaptionButton.unmaximize(
            brightness: widget.brightness,
            onPressed: () => windowManager.unmaximize(),
          )
        : WindowCaptionButton.maximize(
            brightness: widget.brightness,
            onPressed: () => windowManager.maximize(),
          );
  }
}
