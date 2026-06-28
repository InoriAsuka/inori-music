// ignore_for_file: unnecessary_non_null_assertion
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/mini_player_bar.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

/// Adaptive shell scaffold:
/// - Mobile (<600dp): BottomNavigationBar + MiniPlayerBar
/// - Tablet (600–1199dp): NavigationRail + MiniPlayerBar
/// - Desktop (≥1200dp): Permanent NavigationDrawer + MiniPlayerBar
///
/// Desktop keyboard shortcuts:
/// - Space — toggle play/pause
/// - ← / MediaTrackPrevious — previous
/// - → / MediaTrackNext — next
class ShellScaffold extends ConsumerStatefulWidget {
  const ShellScaffold({super.key, required this.child});

  final Widget child;

  @override
  ConsumerState<ShellScaffold> createState() => _ShellScaffoldState();
}

class _ShellScaffoldState extends ConsumerState<ShellScaffold> {
  // Routes are stable constants; labels are resolved at build-time from l10n.
  static const _navRoutes = [
    (icon: Icons.people_outline, route: AppRoutes.artists),
    (icon: Icons.album_outlined, route: AppRoutes.albums),
    (icon: Icons.search, route: AppRoutes.search),
    (icon: Icons.favorite_outline, route: AppRoutes.favorites),
    (icon: Icons.history, route: AppRoutes.history),
  ];

  List<_NavItem> _navItems(AppLocalizations t) => [
    _NavItem(label: t.artists, icon: _navRoutes[0].icon, route: _navRoutes[0].route),
    _NavItem(label: t.albums, icon: _navRoutes[1].icon, route: _navRoutes[1].route),
    _NavItem(label: t.search, icon: _navRoutes[2].icon, route: _navRoutes[2].route),
    _NavItem(label: t.favorites, icon: _navRoutes[3].icon, route: _navRoutes[3].route),
    _NavItem(label: t.history, icon: _navRoutes[4].icon, route: _navRoutes[4].route),
  ];

  late final HardwareKeyboard _keyboard;

  @override
  void initState() {
    super.initState();
    _keyboard = HardwareKeyboard.instance;
    _keyboard.addHandler(_handleKey);
  }

  @override
  void dispose() {
    _keyboard.removeHandler(_handleKey);
    super.dispose();
  }

  bool _handleKey(KeyEvent event) {
    if (event is! KeyDownEvent) return false;
    final notifier = ref.read(playerProvider.notifier);
    switch (event.logicalKey) {
      case LogicalKeyboardKey.space:
        notifier.togglePlayPause();
        return true;
      case LogicalKeyboardKey.arrowLeft:
      case LogicalKeyboardKey.mediaTrackPrevious:
        notifier.previous();
        return true;
      case LogicalKeyboardKey.arrowRight:
      case LogicalKeyboardKey.mediaTrackNext:
        notifier.next();
        return true;
      default:
        return false;
    }
  }

  int _selectedIndex(BuildContext context, List<_NavItem> items) {
    final location = GoRouterState.of(context).matchedLocation;
    for (var i = 0; i < items.length; i++) {
      if (location.startsWith(items[i].route)) return i;
    }
    return 0;
  }

  void _onItemTapped(BuildContext context, List<_NavItem> items, int index) {
    context.go(items[index].route);
  }

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    final items = _navItems(t);
    final width = MediaQuery.sizeOf(context).width;
    final selectedIndex = _selectedIndex(context, items);
    const bottomBar = MiniPlayerBar();

    if (width >= 1200) {
      return _DesktopLayout(
        navItems: items,
        selectedIndex: selectedIndex,
        onItemTapped: (i) => _onItemTapped(context, items, i),
        bottomBar: bottomBar,
        child: widget.child,
      );
    } else if (width >= 600) {
      return _TabletLayout(
        navItems: items,
        selectedIndex: selectedIndex,
        onItemTapped: (i) => _onItemTapped(context, items, i),
        bottomBar: bottomBar,
        child: widget.child,
      );
    } else {
      return _MobileLayout(
        navItems: items,
        selectedIndex: selectedIndex,
        onItemTapped: (i) => _onItemTapped(context, items, i),
        bottomBar: bottomBar,
        child: widget.child,
      );
    }
  }
}

// ---------------------------------------------------------------------------
// Mobile layout
// ---------------------------------------------------------------------------

class _MobileLayout extends StatelessWidget {
  const _MobileLayout({
    required this.navItems,
    required this.selectedIndex,
    required this.onItemTapped,
    required this.child,
    required this.bottomBar,
  });

  final List<_NavItem> navItems;
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;
  final Widget child;
  final Widget bottomBar;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          Expanded(child: child),
          bottomBar,
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: selectedIndex,
        onDestinationSelected: onItemTapped,
        destinations: navItems
            .map((item) => NavigationDestination(
                  icon: Icon(item.icon),
                  label: item.label,
                ))
            .toList(),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Tablet layout
// ---------------------------------------------------------------------------

class _TabletLayout extends StatelessWidget {
  const _TabletLayout({
    required this.navItems,
    required this.selectedIndex,
    required this.onItemTapped,
    required this.child,
    required this.bottomBar,
  });

  final List<_NavItem> navItems;
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;
  final Widget child;
  final Widget bottomBar;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          Expanded(
            child: Row(
              children: [
                NavigationRail(
                  selectedIndex: selectedIndex,
                  onDestinationSelected: onItemTapped,
                  labelType: NavigationRailLabelType.all,
                  destinations: navItems
                      .map((item) => NavigationRailDestination(
                            icon: Icon(item.icon),
                            label: Text(item.label),
                          ))
                      .toList(),
                ),
                const VerticalDivider(thickness: 0.5, width: 0.5),
                Expanded(child: child),
              ],
            ),
          ),
          bottomBar,
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Desktop layout
// ---------------------------------------------------------------------------

class _DesktopLayout extends StatelessWidget {
  const _DesktopLayout({
    required this.navItems,
    required this.selectedIndex,
    required this.onItemTapped,
    required this.child,
    required this.bottomBar,
  });

  final List<_NavItem> navItems;
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;
  final Widget child;
  final Widget bottomBar;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          Expanded(
            child: Row(
              children: [
                SizedBox(
                  width: 220,
                  child: _DesktopSidebar(
                    navItems: navItems,
                    selectedIndex: selectedIndex,
                    onItemTapped: onItemTapped,
                  ),
                ),
                const VerticalDivider(thickness: 0.5, width: 0.5),
                Expanded(child: child),
              ],
            ),
          ),
          bottomBar,
        ],
      ),
    );
  }
}

class _DesktopSidebar extends StatelessWidget {
  const _DesktopSidebar({
    required this.navItems,
    required this.selectedIndex,
    required this.onItemTapped,
  });

  final List<_NavItem> navItems;
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: NeonShrineColors.surface,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 24, 16, 12),
            child: Row(
              children: [
                const Icon(Icons.music_note_rounded, color: NeonShrineColors.primaryViolet, size: 22),
                const SizedBox(width: 8),
                Text(
                  'Inori Music',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        color: NeonShrineColors.onSurface,
                        fontWeight: FontWeight.w700,
                      ),
                ),
              ],
            ),
          ),
          const Divider(),
          Expanded(
            child: ListView.builder(
              itemCount: navItems.length,
              itemBuilder: (context, i) {
                final item = navItems[i];
                final isSelected = i == selectedIndex;
                return ListTile(
                  leading: Icon(
                    item.icon,
                    color: isSelected ? NeonShrineColors.primaryVioletLight : NeonShrineColors.onSurfaceVariant,
                  ),
                  title: Text(
                    item.label,
                    style: TextStyle(
                      color: isSelected ? NeonShrineColors.onSurface : NeonShrineColors.onSurfaceVariant,
                      fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
                    ),
                  ),
                  selected: isSelected,
                  selectedTileColor: NeonShrineColors.primaryVioletDark.withValues(alpha: 0.3),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                  onTap: () => onItemTapped(i),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------

class _NavItem {
  const _NavItem({required this.label, required this.icon, required this.route});
  final String label;
  final IconData icon;
  final String route;
}
