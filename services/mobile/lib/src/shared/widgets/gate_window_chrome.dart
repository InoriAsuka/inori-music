import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:window_manager/window_manager.dart';

import 'package:inori_music/src/shared/desktop_integration.dart';
import 'package:inori_music/src/shared/system_titlebar_provider.dart';

/// Minimal drag handle + (Windows/Linux) close button for [SplashScreen] and
/// [LoginScreen] — neither goes through [DesktopAppBar] (they're deliberately
/// chrome-free, full-bleed brand screens per v5.13.0), but with the native
/// title bar hidden they'd otherwise give a desktop user no way to move or
/// close the window while stuck on the gate. Transparent, thin (32px), only
/// a drag region plus a close button — no minimize/maximize (these windows
/// are fixed-size and non-resizable, so maximize is meaningless here).
///
/// Always uses the light-brand button styling — [LoginScreen]/[SplashScreen]
/// intentionally stay on the fixed Sakura Dusk presentation regardless of
/// the active skin (see background_provider.dart's scope note), so this
/// mirrors that rather than reading [skinProvider].
class GateWindowChrome extends ConsumerWidget {
  const GateWindowChrome({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (!DesktopIntegration.isDesktop) return const SizedBox.shrink();
    final useSystemTitleBar = ref.watch(systemTitleBarProvider);
    if (useSystemTitleBar) return const SizedBox.shrink();

    final isMac = defaultTargetPlatform == TargetPlatform.macOS;
    return SizedBox(
      height: 32,
      child: DragToMoveArea(
        child: isMac
            ? const SizedBox.shrink()
            : Align(
                alignment: Alignment.topRight,
                child: WindowCaptionButton.close(
                  brightness: Brightness.light,
                  onPressed: () => windowManager.close(),
                ),
              ),
      ),
    );
  }
}
