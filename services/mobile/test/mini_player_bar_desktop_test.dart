// mini_player_bar_desktop_test.dart
//
// Covers MiniPlayerBar(showNowPlaying: false) — the desktop shell's variant
// (see shell_scaffold.dart), which drops the cover+title/artist section in
// favour of shuffle/repeat on the left and volume/sleep-timer/queue on the
// right. mini_player_bar_test.dart keeps covering the default
// (showNowPlaying: true) mobile/tablet shape; this file is the desktop one's
// counterpart, plus the width-driven compact/expanded volume switch that
// only exists on this side of the split.
//
// shell_scaffold_nav_test.dart separately proves this variant renders
// correctly *inside* the real four-region desktop shell (including a
// 1200dp-window overflow check against the actual sidebar-adjacent width).
// The isolated-width tests here exist so a future regression in this
// specific compact/expanded switch points straight at MiniPlayerBar instead
// of requiring a trip through the whole shell to localise it.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/mini_player_bar.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;

class _StubPlayerNotifier extends PlayerNotifier {
  @override
  pstate.PlayerState build() => pstate.PlayerState();
}

/// Pumps the desktop bar constrained to exactly [width] logical pixels —
/// the same technique as giving it a fixed-width parent inside the real
/// shell's content column, without needing the whole shell (sidebar, router,
/// auth) just to control one number.
Widget _appAtWidth(double width) => ProviderScope(
  overrides: [playerProvider.overrideWith(_StubPlayerNotifier.new)],
  child: MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(
      body: SizedBox(
        width: width,
        child: const MiniPlayerBar(showNowPlaying: false),
      ),
    ),
  ),
);

void main() {
  testWidgets('drops now-playing info in favour of shuffle/repeat/volume/'
      'queue', (tester) async {
    await tester.pumpWidget(_appAtWidth(900));
    await tester.pump();

    expect(find.byType(MiniPlayerArtwork), findsNothing);
    expect(find.byIcon(Icons.shuffle), findsOneWidget);
    expect(find.byIcon(Icons.repeat), findsOneWidget);
    expect(find.byIcon(Icons.queue_music), findsOneWidget);
    // Volume renders as *some* speaker icon regardless of compact/expanded
    // mode — the default PlayerState().volume is 1.0.
    expect(find.byIcon(Icons.volume_up), findsOneWidget);
  });

  testWidgets('a spacious bar shows the volume control inline (icon + '
      'slider)', (tester) async {
    await tester.pumpWidget(_appAtWidth(900));
    await tester.pump();

    // Two sliders: the progress strip above the content row, and volume's
    // own inline track — if volume had collapsed to its icon-only popover
    // trigger there would only be the one (progress).
    expect(find.byType(Slider), findsNWidgets(2));
    expect(tester.takeException(), isNull);
  });

  testWidgets('a cramped bar collapses volume to an icon-only trigger '
      'instead of overflowing', (tester) async {
    // Well under _volumeCompactThreshold once the shuffle/repeat pair and
    // the transport trio have taken their share — chosen to force the
    // collapse deterministically rather than approaching it from a realistic
    // window size, which is what the 1200dp shell-level test already covers.
    await tester.pumpWidget(_appAtWidth(600));
    await tester.pump();

    expect(
      find.byType(Slider),
      findsNWidgets(1),
      reason:
          'Only the progress strip\'s slider should remain — volume\'s own '
          'track must have collapsed away',
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets('the 1200dp desktop-window breakpoint floor does not overflow '
      'the bar in isolation either', (tester) async {
    // Mirrors shell_scaffold_nav_test.dart's real-shell 1200dp check, but
    // pins the exact width the content column hands the bar at that window
    // size: 1200 - 220 (sidebar) - 8*3 (sidebar's own left/right/inter
    // margins) = 964.
    await tester.pumpWidget(_appAtWidth(964));
    await tester.pump();

    expect(tester.takeException(), isNull);
  });
}
