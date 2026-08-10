// SidebarGroupCollapseNotifier tests — exercises the real production
// notifier against SharedPreferences.setMockInitialValues({}), the same
// pattern search_history_notifier_test.dart already established for a
// SharedPreferences-backed Notifier: the restore path is verified by
// reading a *second*, fresh ProviderContainer after the first one persists,
// which is what actually proves persistence rather than just in-memory
// state.

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/shared/widgets/sidebar_group_collapse_provider.dart';

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  group('SidebarGroupCollapseNotifier', () {
    test('initial state is empty (no group collapsed)', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      expect(container.read(sidebarGroupCollapseProvider), isEmpty);
    });

    test('setCollapsed(true) adds the key; isCollapsed reflects it', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(sidebarGroupCollapseProvider.notifier);

      await notifier.setCollapsed('discover', true);

      expect(container.read(sidebarGroupCollapseProvider), {'discover'});
      expect(notifier.isCollapsed('discover'), isTrue);
      expect(notifier.isCollapsed('library'), isFalse);
    });

    test('setCollapsed(false) removes the key again', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(sidebarGroupCollapseProvider.notifier);

      await notifier.setCollapsed('discover', true);
      await notifier.setCollapsed('discover', false);

      expect(container.read(sidebarGroupCollapseProvider), isEmpty);
    });

    test('two groups collapse independently', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(sidebarGroupCollapseProvider.notifier);

      await notifier.setCollapsed('discover', true);
      await notifier.setCollapsed('library', true);
      await notifier.setCollapsed('discover', false);

      expect(container.read(sidebarGroupCollapseProvider), {'library'});
    });

    test('collapsed state survives a fresh container — the actual '
        'persistence contract, not just in-memory state', () async {
      final first = ProviderContainer();
      await first
          .read(sidebarGroupCollapseProvider.notifier)
          .setCollapsed('discover', true);
      first.dispose();

      // A brand-new container has no shared in-memory state with the
      // first one at all — the only way it can know 'discover' was
      // collapsed is by actually reading it back from SharedPreferences,
      // which is exactly what this proves.
      final second = ProviderContainer();
      addTearDown(second.dispose);
      // Providers are lazy — build() (and the _restore() it kicks off as
      // a fire-and-forget Future) only runs once something actually
      // reads the provider, which is why this read comes *before* the
      // delays rather than after. Two delays, not one: matches
      // search_history_notifier_test.dart's own "state persists and
      // restores from SharedPreferences" case exactly — a single
      // Duration.zero was not reliably enough turns of the event loop
      // for SharedPreferences.getInstance() plus the state write after
      // it to both resolve.
      second.read(sidebarGroupCollapseProvider);
      await Future<void>.delayed(Duration.zero);
      await Future<void>.delayed(Duration.zero);

      expect(second.read(sidebarGroupCollapseProvider), {'discover'});
    });
  });
}
