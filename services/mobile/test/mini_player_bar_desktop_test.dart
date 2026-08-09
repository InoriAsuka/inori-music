// mini_player_bar_desktop_test.dart
//
// Covers MiniPlayerBar's wide shape — the shuffle/repeat-flanked transport
// group, its own seek row with time labels, and the volume/timer/queue
// action group that appear once the bar's own measured width crosses
// MiniPlayerBar._wideBreakpoint (see mini_player_bar.dart).
//
// Before v5.30.6 this shape only ever appeared behind a `showNowPlaying:
// false` constructor flag the desktop shell always passed, and it replaced
// the cover+title section rather than sitting alongside it. That flag is
// gone now — see MiniPlayerBar's doc comment for why both call sites ended
// up wanting the same shape once the cover moved back into the bar — so
// this file drives the wide shape by pumping the bar at a width wide enough
// to trigger it, and mini_player_bar_test.dart's existing (unpinned-width)
// cases keep covering the narrow shape's own behaviour.
//
// shell_scaffold_nav_test.dart separately proves the wide shape renders
// correctly *inside* the real four-region desktop shell (including a
// 1200dp-window overflow check against the actual sidebar-adjacent width).
// The isolated-width tests here exist so a future regression in this
// specific narrow/wide switch points straight at MiniPlayerBar instead of
// requiring a trip through the whole shell to localise it.
import 'package:audio_service/audio_service.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/mini_player_bar.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;

class _StubPlayerNotifier extends PlayerNotifier {
  // Optional so the existing `_StubPlayerNotifier.new` tear-off (a zero-arg
  // `PlayerNotifier Function()`) below stays valid unchanged.
  _StubPlayerNotifier([pstate.PlayerState? state])
    : _state = state ?? pstate.PlayerState();
  final pstate.PlayerState _state;

  @override
  pstate.PlayerState build() => _state;
}

/// Pumps the bar constrained to exactly [width] logical pixels — the same
/// technique as giving it a fixed-width parent inside the real shell's
/// content column, without needing the whole shell (sidebar, router, auth)
/// just to control one number.
Widget _appAtWidth(double width, {pstate.PlayerState? playerState}) =>
    ProviderScope(
      overrides: [
        playerProvider.overrideWith(() => _StubPlayerNotifier(playerState)),
      ],
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Scaffold(
          body: SizedBox(width: width, child: const MiniPlayerBar()),
        ),
      ),
    );

void main() {
  testWidgets(
    'a wide bar keeps the cover/title and adds shuffle/repeat flanking the '
    'transport group plus volume/timer/queue on the right',
    (tester) async {
      await tester.pumpWidget(_appAtWidth(900));
      await tester.pump();

      // The cover+title block never leaves the bar (see MiniPlayerBar's doc
      // comment on why the v5.30.5 SidebarNowPlaying detour was reverted) —
      // the wide shape only *adds* controls around it, it never removes it.
      expect(find.byType(MiniPlayerArtwork), findsOneWidget);
      expect(find.byIcon(Icons.shuffle), findsOneWidget);
      expect(find.byIcon(Icons.repeat), findsOneWidget);
      expect(find.byIcon(Icons.queue_music), findsOneWidget);
      // Volume renders as *some* speaker icon regardless of compact/expanded
      // mode — the default PlayerState().volume is 1.0.
      expect(find.byIcon(Icons.volume_up), findsOneWidget);
    },
  );

  testWidgets(
    'a narrow bar keeps the pre-v5.30.6 shape: no shuffle/repeat/queue',
    (tester) async {
      await tester.pumpWidget(_appAtWidth(400));
      await tester.pump();

      expect(find.byType(MiniPlayerArtwork), findsOneWidget);
      expect(find.byIcon(Icons.shuffle), findsNothing);
      expect(find.byIcon(Icons.repeat), findsNothing);
      expect(find.byIcon(Icons.queue_music), findsNothing);
    },
  );

  testWidgets('a spacious wide bar shows the volume control inline (icon + '
      'slider)', (tester) async {
    await tester.pumpWidget(_appAtWidth(900));
    await tester.pump();

    // Two sliders: the wide shape's own seek row, and volume's own inline
    // track — if volume had collapsed to its icon-only popover trigger there
    // would only be the one (seek).
    expect(find.byType(Slider), findsNWidgets(2));
    expect(tester.takeException(), isNull);
  });

  testWidgets('a wide-but-cramped bar collapses volume to an icon-only trigger '
      'instead of overflowing', (tester) async {
    // Clears MiniPlayerBar._wideBreakpoint (640, measured against the
    // Material's own width, i.e. this SizedBox's width minus the bar's
    // 16px total horizontal margin: 700 - 16 = 684) but leaves section 3
    // — which shares equal Expanded flex with section 1 around the fixed
    // 248px transport block, out of a (700 - 16 - 24) = 660px row — only
    // (660 - 248) / 2 = 206px, under _volumeCompactThreshold's 240.
    await tester.pumpWidget(_appAtWidth(700));
    await tester.pump();

    expect(
      find.byType(Slider),
      findsNWidgets(1),
      reason:
          'Only the seek row\'s slider should remain — volume\'s own '
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

  testWidgets('the bar does not overflow at a 375dp phone width', (
    tester,
  ) async {
    await tester.pumpWidget(_appAtWidth(375));
    await tester.pump();

    expect(tester.takeException(), isNull);
    expect(
      find.byIcon(Icons.shuffle),
      findsNothing,
      reason: 'A 375dp phone is well below _wideBreakpoint',
    );
  });

  // ---------------------------------------------------------------------
  // v5.30.7 — seek row time-label spacing and effects
  // ---------------------------------------------------------------------

  testWidgets(
    'the wide seek row leaves breathing room between each time label and '
    'the track, instead of sitting flush against it',
    (tester) async {
      // Distinct position/duration text (rather than the default 0:00/0:00)
      // so both labels are independently findable.
      final mediaItem = MediaItem(id: 'track-1', title: 'Idol');
      await tester.pumpWidget(
        _appAtWidth(
          900,
          playerState: pstate.PlayerState(
            queue: [mediaItem],
            currentIndex: 0,
            mediaItem: mediaItem,
            playbackState: PlaybackState(playing: true),
            position: Duration.zero,
            duration: const Duration(minutes: 3),
          ),
        ),
      );
      await tester.pump();

      final seekSlider = find.byType(Slider).first;
      final sliderLeft = tester.getTopLeft(seekSlider).dx;
      final sliderRight = tester.getTopRight(seekSlider).dx;
      final positionLabelRight = tester.getTopRight(find.text('0:00')).dx;
      final durationLabelLeft = tester.getTopLeft(find.text('3:00')).dx;

      expect(
        sliderLeft - positionLabelRight,
        greaterThanOrEqualTo(9),
        reason: 'v5.30.7 field report: "进度条前后的时间贴的太紧了"',
      );
      expect(durationLabelLeft - sliderRight, greaterThanOrEqualTo(9));
    },
  );

  testWidgets(
    'the seek row\'s gradient/glow slider theming renders without error '
    'while playing',
    (tester) async {
      final mediaItem = MediaItem(id: 'track-1', title: 'Idol');
      await tester.pumpWidget(
        _appAtWidth(
          900,
          playerState: pstate.PlayerState(
            queue: [mediaItem],
            currentIndex: 0,
            mediaItem: mediaItem,
            playbackState: PlaybackState(playing: true),
            position: const Duration(seconds: 30),
            duration: const Duration(minutes: 3),
          ),
        ),
      );
      await tester.pump();

      // Hover reveals the thumb (and therefore the glow, gated on
      // hover/drag activity — see GlowingSliderThumbShape's `glowing`
      // wiring in mini_player_bar.dart) — without it the thumb radius is 0
      // and there is nothing to glow around.
      final mouse = await tester.createGesture(kind: PointerDeviceKind.mouse);
      await mouse.addPointer(location: Offset.zero);
      addTearDown(mouse.removePointer);
      await mouse.moveTo(tester.getCenter(find.byType(Slider).first));
      await tester.pump();

      expect(tester.takeException(), isNull);
    },
  );
}
