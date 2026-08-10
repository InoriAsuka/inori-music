import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _kCollapsedSidebarGroupsKey = 'sidebar.collapsedGroups';

final sidebarGroupCollapseProvider =
    NotifierProvider<SidebarGroupCollapseNotifier, Set<String>>(
      SidebarGroupCollapseNotifier.new,
    );

/// Which of the desktop sidebar's collapsible nav groups (identified by a
/// stable key — see `shell_scaffold.dart`'s `_NavGroup.collapseKey` — not
/// by their translated header text, so persistence survives a language
/// switch) are currently folded shut.
///
/// One [Notifier] holding a set of keys rather than one bool-Notifier per
/// group: `coverFlowModeProvider`/`systemTitleBarProvider` are this
/// codebase's established "SharedPreferences + Notifier" shape for a single
/// persisted flag, but the sidebar has two collapsible groups today and a
/// third would be a plausible future addition — a set keyed by a stable
/// per-group id extends to that without adding a second provider (and a
/// second SharedPreferences key) for every group this shell ever grows,
/// which is what "别造设置子系统" (v5.33.0's own instruction) is warning
/// against: not a full settings subsystem, but not four copies of the same
/// six lines either.
class SidebarGroupCollapseNotifier extends Notifier<Set<String>> {
  @override
  Set<String> build() {
    _restore();
    return const {};
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    state = (prefs.getStringList(_kCollapsedSidebarGroupsKey) ?? const [])
        .toSet();
  }

  bool isCollapsed(String groupKey) => state.contains(groupKey);

  Future<void> setCollapsed(String groupKey, bool collapsed) async {
    final next = Set<String>.of(state);
    if (collapsed) {
      next.add(groupKey);
    } else {
      next.remove(groupKey);
    }
    state = next;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setStringList(_kCollapsedSidebarGroupsKey, next.toList());
  }
}
