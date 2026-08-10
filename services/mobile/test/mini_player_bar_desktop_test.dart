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
import 'package:flutter/rendering.dart' show RenderParagraph;
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
    // Flutter's default test surface is only 800x600 logical px — a plain
    // SizedBox(width: 900) (as every other test in this file uses) gets
    // silently clamped to 800 by the Scaffold body's own incoming
    // constraints (BoxConstraints.tighten clamps to the incoming max, it
    // never exceeds it). That went unnoticed pre-v5.32.0 because the old
    // fixed 248px middle column left section 3 comfortably clear of
    // _volumeCompactThreshold even at the clamped 800 (section 3 = (800 -
    // 16 - 24 - 248) / 2 = 256). v5.32.0's wider, row-proportional middle
    // column needs more than 800 to leave section 3 the same room — at the
    // clamped 800 it would measure (800-16-24)*0.46 ≈ 350, leaving section 3
    // only (760-350)/2 = 205, under the 240 threshold no matter how this
    // test's *requested* width is raised, since raising it past 800 has no
    // effect while the request is silently capped. Sizing the test surface
    // itself (mirroring full_player_layout_test.dart's _sizeWindow) is the
    // actual fix, not a wider number: 1200 was the bar width the v5.32.0
    // field report itself measured against ("现在桌面播放条约 1200px 宽"), and
    // it makes the request genuinely take effect — row 1160, column ≈ 534,
    // section 3 ≈ 313, clearing 240 with real margin.
    tester.view.devicePixelRatio = 1.0;
    tester.view.physicalSize = const Size(1200, 600);
    addTearDown(tester.view.reset);

    await tester.pumpWidget(_appAtWidth(1200));
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
    // cramped. Since v5.32.0, section 3 shares equal Expanded flex with
    // section 1 around miniPlayerMiddleColumnWidth's own computed column
    // (not the old fixed 248px transport block) — out of a (700 - 16 - 24) =
    // 660px row that column measures 660*0.46 ≈ 304px, leaving section 3
    // only (660 - 304) / 2 ≈ 178px, still well under
    // _volumeCompactThreshold's 240 (in fact more cramped than the pre-
    // v5.32.0 206px, so this case is if anything more clearly "compact"
    // now, not less).
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

  // ---------------------------------------------------------------------
  // v5.32.0 — hovering must not shrink the track
  // ---------------------------------------------------------------------

  testWidgets(
    'the wide seek row\'s track reserves the same width whether the thumb '
    'is hovered or resting',
    (tester) async {
      // Field report: "鼠标放上去出现控制点会导致整个进度条收缩一部分". Root
      // cause was BaseSliderTrackShape.getPreferredRect insetting the track
      // by max(thumbShape, overlayShape).getPreferredSize — both of which
      // used to vary with hover (see GlowingSliderThumbShape.getPreferredSize
      // and the overlayShape wiring in mini_player_bar.dart, both fixed in
      // v5.32.0). This reads the *actual* SliderThemeData in effect at rest
      // and on hover — via SliderTheme.of, the same lookup Slider itself
      // uses — and calls the real trackShape.getPreferredRect the framework
      // calls during layout, rather than re-deriving the geometry by hand;
      // that's what makes this a regression test for the mechanism, not just
      // for GlowingSliderThumbShape's own getPreferredSize in isolation.
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

      double trackWidth() {
        final sliderFinder = find.byType(Slider).first;
        final theme = SliderTheme.of(tester.element(sliderFinder));
        final renderBox = tester.renderObject<RenderBox>(sliderFinder);
        return theme.trackShape!
            .getPreferredRect(
              parentBox: renderBox,
              sliderTheme: theme,
              isEnabled: true,
            )
            .width;
      }

      final restWidth = trackWidth();

      final mouse = await tester.createGesture(kind: PointerDeviceKind.mouse);
      await mouse.addPointer(location: Offset.zero);
      addTearDown(mouse.removePointer);
      await mouse.moveTo(tester.getCenter(find.byType(Slider).first));
      await tester.pump();

      expect(
        trackWidth(),
        restWidth,
        reason:
            'Hovering must only change what paint() draws inside the '
            'track, never the reserved track rect itself',
      );
    },
  );

  // ---------------------------------------------------------------------
  // v5.32.0 — the seek row's width proportion
  // ---------------------------------------------------------------------

  group('miniPlayerMiddleColumnWidth', () {
    test('scales at 0.46 of the row width in the middle of its range', () {
      // 800 * 0.46 = 368, comfortably inside [290, 640] on both sides.
      expect(miniPlayerMiddleColumnWidth(800), 368.0);
    });

    test('clamps to the floor for a narrow row', () {
      // At MiniPlayerBar._wideBreakpoint (640) itself, 640*0.46 = 294.4,
      // already above the 290 floor — this asserts the floor still holds
      // for anything narrower, which the wide shape never actually renders
      // at but the pure function must still behave sanely for.
      expect(miniPlayerMiddleColumnWidth(200), 290.0);
    });

    test('clamps to the ceiling on an ultrawide bar', () {
      expect(miniPlayerMiddleColumnWidth(3000), 640.0);
    });
  });

  group('miniPlayerSeekRowWidth', () {
    test('is 0.7 of the middle column in the middle of its range', () {
      // 400 * 0.7 = 280, inside [220, 460].
      expect(miniPlayerSeekRowWidth(400), 280.0);
    });

    test('clamps to the floor at the middle column\'s own floor', () {
      // 290 * 0.7 = 203, below the 220 floor.
      expect(miniPlayerSeekRowWidth(290), 220.0);
    });

    test('clamps to the ceiling at the middle column\'s own ceiling', () {
      // 640 * 0.7 = 448, below the 460 ceiling — the seek row's own ceiling
      // never actually binds at the middle column's clamp bounds, only on
      // some hypothetical wider column; asserted anyway so the function's
      // contract is pinned down independent of what miniPlayerMiddleColumnWidth
      // happens to produce today.
      expect(miniPlayerSeekRowWidth(700), 460.0);
    });
  });

  testWidgets(
    'the wide seek row\'s track width does not change when elapsed time '
    'crosses a digit-count boundary (9:59 -> 10:00)',
    (tester) async {
      // v5.32.0 field report: "进度条太短，与整条比例失调", one contributing
      // cause being that the elapsed-time label's own width used to grow the
      // instant it needed a second minute digit, stealing that width from
      // the slider. Duration is 15 minutes so both "9:59" and "10:00" are
      // valid elapsed positions within the same track — this isolates the
      // digit-count change from a track change, which the fixed-width label
      // (sized off the *duration*, see _timeLabelWidth) is specifically
      // meant to make invisible to the slider's own width.
      final mediaItem = MediaItem(id: 'track-1', title: 'Long Track');
      Widget appAt(Duration position) => _appAtWidth(
        900,
        playerState: pstate.PlayerState(
          queue: [mediaItem],
          currentIndex: 0,
          mediaItem: mediaItem,
          playbackState: PlaybackState(playing: true),
          position: position,
          duration: const Duration(minutes: 15),
        ),
      );

      await tester.pumpWidget(appAt(const Duration(minutes: 9, seconds: 59)));
      await tester.pump();
      final widthBefore = tester.getSize(find.byType(Slider).first).width;

      await tester.pumpWidget(appAt(const Duration(minutes: 10)));
      await tester.pump();
      final widthAfter = tester.getSize(find.byType(Slider).first).width;

      expect(
        widthAfter,
        widthBefore,
        reason:
            'The elapsed label crossing from 4 characters ("9:59") to 5 '
            '("10:00") must not nudge how much width is left for the slider',
      );
    },
  );

  // ---------------------------------------------------------------------
  // v5.33.0 — time labels must not wrap under a boosted textScaler
  // ---------------------------------------------------------------------

  testWidgets(
    'the wide seek row\'s time labels stay single-line under a boosted '
    'textScaler',
    (tester) async {
      // Field report: "0:43" rendered as two lines ("0:4" / "3") and "3:05"
      // as ("3:0" / "5") once the system font scale was increased.
      // _timeLabelWidth measured its reserved slot against the *default*
      // TextScaler (1.0) while the real Text below always painted at
      // whatever MediaQuery.textScalerOf(context) reported — so any scale
      // above 1.0 measured a narrower slot than the real glyphs needed.
      // 1.3 is inside the field report's own repro range.
      final mediaItem = MediaItem(id: 'track-1', title: 'Idol');
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            playerProvider.overrideWith(
              () => _StubPlayerNotifier(
                pstate.PlayerState(
                  queue: [mediaItem],
                  currentIndex: 0,
                  mediaItem: mediaItem,
                  playbackState: PlaybackState(playing: true),
                  position: const Duration(seconds: 43),
                  duration: const Duration(minutes: 3, seconds: 5),
                ),
              ),
            ),
          ],
          child: MaterialApp(
            localizationsDelegates: AppLocalizations.localizationsDelegates,
            supportedLocales: AppLocalizations.supportedLocales,
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(
                context,
              ).copyWith(textScaler: const TextScaler.linear(1.3)),
              child: child!,
            ),
            home: Scaffold(
              body: SizedBox(width: 900, child: const MiniPlayerBar()),
            ),
          ),
        ),
      );
      await tester.pump();

      // Measures the *actual rendered height* against an independent
      // single-line reference — not just "the widget exists" (which a
      // silently-two-line label would still satisfy) and not a hand-picked
      // pixel constant (which would go stale the moment the label's font
      // size changes). The reference painter reuses the render object's own
      // resolved span and textScaler, laid out at TextPainter's default
      // unbounded width, so it can never itself wrap — that makes its
      // height a trustworthy "one line" yardstick to compare the real,
      // width-constrained label against. A label that wrapped to two lines
      // renders at roughly double this height; framework line-height
      // rounding never gets remotely close to doubling it.
      RenderParagraph paragraphFor(String text) =>
          tester.renderObject<RenderParagraph>(find.text(text));

      double singleLineReferenceHeight(RenderParagraph paragraph) {
        final reference = TextPainter(
          text: paragraph.text,
          textDirection: TextDirection.ltr,
          textScaler: paragraph.textScaler,
        )..layout();
        return reference.height;
      }

      for (final label in ['0:43', '3:05']) {
        final paragraph = paragraphFor(label);
        final actualHeight = paragraph.size.height;
        final oneLineHeight = singleLineReferenceHeight(paragraph);

        expect(
          actualHeight,
          lessThan(oneLineHeight * 1.5),
          reason:
              '"$label" rendered at $actualHeight logical px tall against a '
              'single-line reference of $oneLineHeight — that only happens '
              'if it wrapped onto a second line',
        );
      }
    },
  );
}
