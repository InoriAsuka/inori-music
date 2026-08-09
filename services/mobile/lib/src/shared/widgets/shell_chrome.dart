import 'package:flutter/widgets.dart';

/// Tells a [DesktopAppBar]/[DesktopSliverAppBar] somewhere below it whether
/// something to its left has already reserved room for macOS's native
/// traffic-light buttons, so it doesn't reserve that space a second time.
///
/// Provided by `_DesktopLayout` (`shell_scaffold.dart`), whose floating
/// sidebar spans the window's full height and therefore owns the window's
/// top-left corner — the lights land on the sidebar, not on the content
/// column to its right. That's the v5.30.0 field-report bug in one sentence:
/// the sidebar became the thing sitting at the corner, but the gutter
/// reservation stayed on `DesktopAppBar`, which by then started at x >= 236
/// and was never actually under the lights.
///
/// Every other host of a [DesktopAppBar] has no such ancestor and keeps the
/// pre-v5.30.5 fixed-70px behaviour — the login/splash gate is chrome-free
/// (`gate_window_chrome.dart`) and the full-bleed player route
/// (`full_player_screen.dart`) has no sidebar behind it when it's showing
/// (it's a sibling top-level route, not nested in the shell), so both still
/// sit directly under the corner themselves. That makes "no ancestor" the
/// correct default rather than a gap to fill in.
///
/// Deliberately just one flag rather than something platform- or
/// setting-aware: [DesktopAppBar] already re-derives "is this macOS, and is
/// the custom title bar even in use" on its own before ever consulting this —
/// this widget only answers "is a sidebar already handling it", not
/// "is handling it necessary right now".
class ShellChrome extends InheritedWidget {
  const ShellChrome({
    super.key,
    required this.reservesTrafficLightGutter,
    required super.child,
  });

  final bool reservesTrafficLightGutter;

  static ShellChrome? of(BuildContext context) =>
      context.dependOnInheritedWidgetOfExactType<ShellChrome>();

  @override
  bool updateShouldNotify(ShellChrome oldWidget) =>
      reservesTrafficLightGutter != oldWidget.reservesTrafficLightGutter;
}
