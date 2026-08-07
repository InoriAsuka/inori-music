import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:window_manager/window_manager.dart';

import 'package:inori_music/src/shared/desktop_integration.dart';
import 'package:inori_music/src/shared/system_titlebar_provider.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/titlebar_material_provider.dart';

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
        // flutter_acrylic has no per-region API — the Mica/acrylic backdrop
        // is a window-level effect (see DesktopIntegration.applyTitlebarMaterial).
        // It only becomes visible here, where the AppBar's own background is
        // made transparent; everywhere else keeps its normal opaque skin
        // surface painted on top, which fully hides the native backdrop.
        final hasMaterial =
            Platform.isWindows &&
            ref.watch(titlebarMaterialProvider) != TitlebarMaterial.none;

        final appBar = AppBar(
          title: title,
          backgroundColor: hasMaterial ? Colors.transparent : null,
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

/// Collapsing-header counterpart to [DesktopAppBar], for screens that need a
/// [SliverAppBar] (large cover art that shrinks into a slim bar on scroll —
/// see album/playlist/artist detail screens) while keeping the same desktop
/// drag-region/window-caption-button integration.
///
/// Lives in this file rather than its own — it needs [_MaximizeButton] and
/// the same window-caption-button trio [DesktopAppBar] builds, and Dart's
/// visibility is per-library (per-file), not per-class.
///
/// [background] is [FlexibleSpaceBar]'s `background` (the expanded-state
/// content, e.g. cover art) — kept separate from [title] because the title
/// text needs to stay in the persistent toolbar row so it (and the window
/// buttons alongside it) remain visible at every scroll position, not just
/// when expanded.
class DesktopSliverAppBar extends StatelessWidget {
  const DesktopSliverAppBar({
    super.key,
    this.title,
    this.actions,
    this.leading,
    this.automaticallyImplyLeading = true,
    this.background,
    this.expandedHeight = 200,
    this.pinned = true,
  });

  final Widget? title;
  final List<Widget>? actions;
  final Widget? leading;
  final bool automaticallyImplyLeading;
  final Widget? background;
  final double expandedHeight;
  final bool pinned;

  // Title goes to *either* SliverAppBar.title (plain, no flexible space) or
  // FlexibleSpaceBar.title (which owns the whole large-to-small collapse
  // animation) — never both. Setting both is a real Flutter anti-pattern:
  // FlexibleSpaceBar's own crossfade already handles showing/hiding the
  // title as the header collapses, so a second copy on SliverAppBar.title
  // renders as a literal on-screen duplicate rather than a no-op.
  SliverAppBar _plainSliverAppBar() => SliverAppBar(
    title: background == null ? title : null,
    actions: actions,
    leading: leading,
    automaticallyImplyLeading: automaticallyImplyLeading,
    expandedHeight: expandedHeight,
    pinned: pinned,
    flexibleSpace: background == null
        ? null
        : FlexibleSpaceBar(title: title, background: background),
  );

  @override
  Widget build(BuildContext context) {
    if (!DesktopIntegration.isDesktop) return _plainSliverAppBar();

    return Consumer(
      builder: (context, ref, _) {
        final useSystemTitleBar = ref.watch(systemTitleBarProvider);
        if (useSystemTitleBar) return _plainSliverAppBar();

        final skin = ref.watch(skinProvider).active;
        final isMac = defaultTargetPlatform == TargetPlatform.macOS;
        final hasMaterial =
            Platform.isWindows &&
            ref.watch(titlebarMaterialProvider) != TitlebarMaterial.none;

        // Only the always-visible toolbar row (leading/title/actions) needs
        // to dodge the macOS traffic lights — unlike DesktopAppBar's flat
        // bar, padding the whole widget here would also shift the expanded
        // cover-art background, which is meant to stay edge-to-edge.
        final leadingContent =
            leading ??
            (automaticallyImplyLeading && Navigator.canPop(context)
                ? const BackButton()
                : null);
        final resolvedLeading = isMac && leadingContent != null
            ? Padding(
                padding: const EdgeInsets.only(left: 70),
                child: leadingContent,
              )
            : leadingContent;

        // DragToMoveArea builds a GestureDetector — a box widget — so unlike
        // DesktopAppBar it can't wrap the SliverAppBar itself (a sliver
        // widget can't be nested inside a box widget and still work as a
        // CustomScrollView.slivers entry; that combination throws at
        // runtime). Instead it wraps only `background`, which is an
        // ordinary box-widget slot inside FlexibleSpaceBar — that covers the
        // large hero-image area (the natural drag target, matching EchoMusic/
        // Spotube's header). The persistent toolbar row (title/actions) is
        // left undragged; window-caption buttons there stay simple taps with
        // no gesture-arena interaction to reason about.
        final draggableBackground = background == null
            ? null
            : DragToMoveArea(child: background!);

        return SliverAppBar(
          title: draggableBackground == null ? title : null,
          backgroundColor: hasMaterial ? Colors.transparent : null,
          leading: resolvedLeading,
          leadingWidth: isMac && leadingContent != null ? 126 : null,
          automaticallyImplyLeading: false,
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
          expandedHeight: expandedHeight,
          pinned: pinned,
          flexibleSpace: draggableBackground == null
              ? null
              : FlexibleSpaceBar(title: title, background: draggableBackground),
        );
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
