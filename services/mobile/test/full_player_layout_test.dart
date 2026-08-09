// full_player_layout_test.dart
//
// Covers the v5.28.0 player-page layout (the macOS traffic-light gutter, the
// grouped transport row, and the centred/split behaviour that replaces the
// artwork-vs-lyrics PageView on wide windows) plus the v5.29.0 follow-up:
// the player block's width is now derived from the cover instead of
// stretched across the region, the side panel is a proportional half
// instead of a fixed 380px, and the compact-controls switch is judged
// against the control block's own width rather than the whole window.
//
// v5.30.0 adds: the top bar's title now comes from l10n instead of a
// hardcoded string, a gradient scrim behind the top bar for contrast
// robustness against the fluid backdrop's local colour variance, a
// falsification check that the play/pause button really does follow the
// extracted accent, and Cover Flow as an alternate artwork display mode.
//
import 'package:audio_service/audio_service.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/player/cover_flow_artwork.dart';
import 'package:inori_music/src/player/cover_flow_mode_provider.dart';
import 'package:inori_music/src/player/full_player_screen.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/lyrics/lyric_line.dart';
import 'package:inori_music/src/lyrics/lyrics_provider.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';
import 'package:inori_music/src/shared/widgets/glass_panel.dart';

import 'support/fake_playback_engine.dart';

/// Stops the lyrics panel from making a real HTTP request, which would leave
/// a pending timer and fail the test on teardown.
class _StubLyricsNotifier extends LyricsNotifier {
  @override
  Future<List<LyricLine>?> build(String trackId) async => null;
}

class _StubPlayerNotifier extends PlayerNotifier {
  _StubPlayerNotifier(this._state);
  final pstate.PlayerState _state;

  @override
  pstate.PlayerState build() => _state;
}

/// Stubs the artwork URL lookup so a test can drive the "has a cover" branch
/// without a network round trip — same shape as cover_palette_test.dart's.
class _StubArtworkNotifier extends ArtworkUrlNotifier {
  _StubArtworkNotifier(this._url);
  final String? _url;

  @override
  Future<String?> build(String albumId) async => _url;
}

pstate.PlayerState _playing() {
  const item = MediaItem(
    id: 'track-1',
    title: 'Emmanuel',
    artist: 'Chris Botti',
  );
  return pstate.PlayerState(
    queue: [item],
    currentIndex: 0,
    mediaItem: item,
    playbackState: PlaybackState(playing: true),
    duration: const Duration(minutes: 5, seconds: 55),
  );
}

/// A queue of [count] items, all carrying [albumId] so a single artwork
/// stub covers every slot Cover Flow might render.
pstate.PlayerState _playingWithArtwork(String albumId, {int count = 1}) {
  final items = List.generate(
    count,
    (i) => MediaItem(
      id: 'track-$i',
      title: 'Track $i',
      artist: 'Chris Botti',
      extras: {'albumId': albumId},
    ),
  );
  return pstate.PlayerState(
    queue: items,
    currentIndex: 0,
    mediaItem: items.first,
    playbackState: PlaybackState(playing: true),
    duration: const Duration(minutes: 5, seconds: 55),
  );
}

Widget _app({Locale? locale}) => ProviderScope(
  overrides: [
    playerProvider.overrideWith(() => _StubPlayerNotifier(_playing())),
    playbackEngineProvider.overrideWithValue(FakePlaybackEngine()),
    lyricsProvider.overrideWith(_StubLyricsNotifier.new),
  ],
  child: MaterialApp(
    locale: locale,
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: const FullPlayerScreen(),
  ),
);

/// Like [_app], but with a resolvable cover so [LyricsBackground] switches
/// content to the artwork-derived overlay skin — needed to test anything
/// that depends on [CoverPalette]-driven colour or Cover Flow, both of which
/// only ever activate once there's a cover to derive them from.
Widget _appWithArtwork({
  required String albumId,
  CoverPalette? palette,
  int queueLength = 1,
  bool coverFlow = false,
}) => ProviderScope(
  overrides: [
    playerProvider.overrideWith(
      () =>
          _StubPlayerNotifier(_playingWithArtwork(albumId, count: queueLength)),
    ),
    playbackEngineProvider.overrideWithValue(FakePlaybackEngine()),
    lyricsProvider.overrideWith(_StubLyricsNotifier.new),
    artworkUrlProvider.overrideWith(
      () => _StubArtworkNotifier('https://example/a.jpg'),
    ),
    coverPaletteProvider.overrideWith((ref, source) async => palette),
    if (coverFlow) coverFlowModeProvider.overrideWith(() => _AlwaysOn()),
  ],
  child: MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: const FullPlayerScreen(),
  ),
);

class _AlwaysOn extends CoverFlowModeNotifier {
  @override
  bool build() => true;
}

void _sizeWindow(WidgetTester tester, Size size) {
  tester.view.devicePixelRatio = 1.0;
  tester.view.physicalSize = size;
  addTearDown(tester.view.reset);
}

/// pump() rather than pumpAndSettle(): the cover backdrop runs repeat()
/// controllers that never settle. Two frames is enough to lay out.
Future<void> _settle(WidgetTester tester) async {
  await tester.pump();
  await tester.pump();
}

void main() {
  testWidgets('the transport trio is grouped, not spread across the bar', (
    tester,
  ) async {
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    final prev = tester.getCenter(find.byIcon(Icons.skip_previous));
    final next = tester.getCenter(find.byIcon(Icons.skip_next));
    final repeat = tester.getCenter(find.byIcon(Icons.repeat));
    final favourite = tester.getCenter(find.byIcon(Icons.favorite_border));

    final trioSpan = next.dx - prev.dx;
    final barSpan = favourite.dx - repeat.dx;
    expect(
      trioSpan,
      lessThan(barSpan / 2),
      reason:
          'spaceEvenly gave the transport trio the same spacing as the '
          'secondary controls; grouped, it should occupy a tight middle band',
    );

    // And that band is centred on the bar, not merely narrow.
    final trioCentre = (prev.dx + next.dx) / 2;
    final barCentre = (repeat.dx + favourite.dx) / 2;
    expect((trioCentre - barCentre).abs(), lessThan(24));
  });

  testWidgets('a wide window shows the cover alone, with no page dots', (
    tester,
  ) async {
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    expect(
      find.byType(PageView),
      findsNothing,
      reason: 'Lyrics dock beside the player here, so there is nothing to page',
    );
  });

  testWidgets('a narrow window keeps the artwork/lyrics pager', (tester) async {
    _sizeWindow(tester, const Size(420, 900));
    await tester.pumpWidget(_app());
    await _settle(tester);

    expect(find.byType(PageView), findsOneWidget);
  });

  testWidgets('the queue button docks a panel on a wide window', (
    tester,
  ) async {
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    expect(find.text('播放队列'), findsNothing);

    await tester.tap(find.byIcon(Icons.queue_music));
    await _settle(tester);

    expect(find.text('播放队列'), findsOneWidget);
    expect(
      find.byType(BottomSheet),
      findsNothing,
      reason: 'Wide windows dock the queue rather than covering the player',
    );
  });

  testWidgets('the same button closes the panel again', (tester) async {
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    await tester.tap(find.byIcon(Icons.queue_music));
    await _settle(tester);
    await tester.tap(find.byIcon(Icons.queue_music));
    await _settle(tester);

    expect(find.text('播放队列'), findsNothing);
  });

  testWidgets('opening lyrics replaces the queue rather than stacking', (
    tester,
  ) async {
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    await tester.tap(find.byIcon(Icons.queue_music));
    await _settle(tester);
    await tester.tap(find.byIcon(Icons.mic_external_on));
    await _settle(tester);

    expect(find.text('歌词'), findsOneWidget);
    expect(find.text('播放队列'), findsNothing);
  });

  testWidgets('the player stays centred in whatever width is left', (
    tester,
  ) async {
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    // Measured off the transport rather than the title: once the queue is
    // docked the track title appears twice (player heading and queue row).
    // Midpoint of prev/next, since skip_next alone sits right of centre.
    double transportCentre() =>
        (tester.getCenter(find.byIcon(Icons.skip_previous)).dx +
            tester.getCenter(find.byIcon(Icons.skip_next)).dx) /
        2;

    final centredX = transportCentre();
    expect(
      centredX,
      closeTo(700, 40),
      reason: 'With no panel the player occupies the whole 1400px window',
    );

    await tester.tap(find.byIcon(Icons.queue_music));
    await _settle(tester);

    final splitX = transportCentre();
    expect(
      splitX,
      lessThan(centredX),
      reason: 'Docking a panel must shift the player left, not overlap it',
    );
    expect(
      splitX,
      // The side panel is a proportional half (flex 1:1) since v5.29.0, not
      // a fixed 380px sidebar, so the player's half is 1400/2 and it centres
      // at a quarter of the full window width — expressed as a ratio rather
      // than a hardcoded pixel figure so this doesn't go stale the next time
      // the split ratio changes.
      closeTo(1400 / 4, 40),
      reason: 'And it should centre in the width that is left, not just move',
    );
  });

  group('playerArtworkSize', () {
    test('a comfortably large region clamps to the 420 ceiling', () {
      expect(playerArtworkSize(regionWidth: 2000, regionHeight: 1400), 420.0);
    });

    test('a cramped region clamps to the 160 floor', () {
      expect(playerArtworkSize(regionWidth: 200, regionHeight: 200), 160.0);
    });

    test('otherwise scales with the tighter of width and height', () {
      // width*0.42 = 210, height*0.40 = 280 — width is the tighter bound.
      expect(playerArtworkSize(regionWidth: 500, regionHeight: 700), 210.0);
    });
  });

  group('playerControlWidth', () {
    test('sits within the 1.4-1.5x cover-width band Apple Music used', () {
      const artworkSize = 300.0;
      final width = playerControlWidth(
        artworkSize: artworkSize,
        regionWidth: 2000,
      );
      expect(width / artworkSize, inInclusiveRange(1.4, 1.5));
    });

    test('never throws RangeError when the region is narrower than the '
        'floor plus its margin', () {
      // regionWidth - 48 is far below the usability floor here, which would
      // invert a plain clamp(floor, regionWidth - 48) — the v5.29.0 fix is
      // the math.max floor on the ceiling itself.
      expect(
        () => playerControlWidth(artworkSize: 160, regionWidth: 200),
        returnsNormally,
      );
      // Clamped up to the floor (400, not the original plan's estimated
      // 280 — see playerControlWidth's doc comment for why 280 measured
      // short of what the compact transport bar actually needs).
      expect(playerControlWidth(artworkSize: 160, regionWidth: 200), 400.0);
    });
  });

  testWidgets(
    'a spacious unsplit window clamps the cover and scales the control panel '
    'off it, not the full window width',
    (tester) async {
      // Both dimensions are large enough to hit the 420 ceiling regardless of
      // the exact top-bar height, so the expectation doesn't depend on
      // predicting that height precisely.
      _sizeWindow(tester, const Size(1400, 1400));
      await tester.pumpWidget(_app());
      await _settle(tester);

      expect(find.byType(GlassPanel), findsOneWidget);
      expect(
        tester.getSize(find.byType(GlassPanel)).width,
        closeTo(420 * 1.45, 2),
      );
    },
  );

  testWidgets('a narrow window keeps the pre-v5.29.0 fixed 280 cover size', (
    tester,
  ) async {
    _sizeWindow(tester, const Size(420, 900));
    await tester.pumpWidget(_app());
    await _settle(tester);

    expect(tester.getSize(find.byType(PageView)), const Size(280, 280));
  });

  testWidgets(
    'a mid-width window does not overflow the transport row once a panel '
    'is docked',
    (tester) async {
      // This is the case that made the pre-v5.29.0 compactControls switch
      // (judged off the whole window) stale: 950 clears the old 600px
      // window-width threshold comfortably, but splitting the window in
      // half for the queue panel leaves the control block under 300px wide
      // — this test is what _compactControlsBreakpoint is calibrated
      // against, not a number picked by inspection.
      _sizeWindow(tester, const Size(950, 700));
      await tester.pumpWidget(_app());
      await _settle(tester);

      await tester.tap(find.byIcon(Icons.queue_music));
      await _settle(tester);

      expect(find.text('播放队列'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'a wide unsplit window is not misjudged as needing compact controls',
    (tester) async {
      // The other half of the D.1 calibration: raising the threshold to fix
      // the mid-width overflow above must not accidentally push a genuinely
      // spacious layout into compact mode too. controlWidth here is ~609,
      // comfortably past _compactControlsBreakpoint (480), so buttons should
      // render at their ambient (non-themed) size rather than
      // _ControlDensity's compact footprint (26-36px, asserted above).
      _sizeWindow(tester, const Size(1400, 1400));
      await tester.pumpWidget(_app());
      await _settle(tester);

      final prevButton = find.ancestor(
        of: find.byIcon(Icons.skip_previous),
        matching: find.byType(IconButton),
      );
      expect(tester.getSize(prevButton).width, greaterThan(44));
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('an ultra-wide split panel is capped rather than growing '
      'unbounded', (tester) async {
    _sizeWindow(tester, const Size(2200, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    await tester.tap(find.byIcon(Icons.queue_music));
    await _settle(tester);

    final panelGlass = find.ancestor(
      of: find.text('播放队列'),
      matching: find.byType(GlassPanel),
    );
    expect(panelGlass, findsOneWidget);
    expect(tester.getSize(panelGlass).width, lessThanOrEqualTo(562));
  });

  // ---------------------------------------------------------------------
  // v5.29.0 field-report follow-ups (three small fixes carried into
  // v5.30.0's batch — see requirement.md's v5.29.0 entry).
  // ---------------------------------------------------------------------

  testWidgets('the top bar title comes from l10n, not a hardcoded string', (
    tester,
  ) async {
    // English and the hardcoded string used to be identical ("Now Playing"),
    // which would make find.text('Now Playing') pass either way and prove
    // nothing. Japanese's translation ("再生中") is what actually
    // discriminates between "reads AppLocalizations" and "still hardcoded".
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app(locale: const Locale('ja')));
    await _settle(tester);

    expect(find.text('再生中'), findsOneWidget);
    expect(find.text('Now Playing'), findsNothing);
  });

  testWidgets('the top bar carries its own contrast scrim', (tester) async {
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    final scrim = find.byWidgetPredicate((widget) {
      if (widget is! DecoratedBox) return false;
      final decoration = widget.decoration;
      return decoration is BoxDecoration &&
          decoration.gradient is LinearGradient;
    });
    expect(scrim, findsOneWidget);
    // It must actually sit above the title, not merely exist somewhere in
    // the tree, or it does nothing for the contrast it's meant to guarantee.
    expect(
      find.ancestor(of: find.text('Now Playing'), matching: scrim),
      findsOneWidget,
    );
  });

  testWidgets('falsification check: the play/pause button colour tracks the '
      'extracted accent, not a fixed pink', (tester) async {
    // Per the v5.30.0 plan: write this assertion and run it before
    // changing any code. artworkOverlaySkin() already wires sakuraPink to
    // CoverPalette.accentOverArtwork, and the button already reads
    // context.skinColors.sakuraPink — if that wiring is intact this
    // should simply pass.
    const vibrant = Color(0xFF34C77B);
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(
      _appWithArtwork(
        albumId: 'album-x',
        palette: const CoverPalette(
          dominant: Color(0xFF102030),
          lightVibrant: vibrant,
        ),
      ),
    );
    await _settle(tester);

    final playToggle = tester.widget<Container>(
      find.ancestor(
        of: find.byIcon(Icons.pause_rounded),
        matching: find.byType(Container),
      ),
    );
    final decoration = playToggle.decoration! as BoxDecoration;
    expect(
      decoration.color,
      vibrant,
      reason:
          'If this fails, the button is not actually reading the '
          'extracted accent — a real bug, not the field report reproduced '
          'by this test.',
    );
  });

  // ---------------------------------------------------------------------
  // v5.30.5 — volume control parity with the desktop mini player bar. The
  // bar's own volume/mute behaviour is covered directly in
  // volume_control_test.dart against the shared VolumeControl widget; this
  // just proves the full player screen actually mounts one, always in its
  // compact (icon + popover) shape — see full_player_screen.dart's inline
  // comment for why an inline slider isn't safe in this row.
  // ---------------------------------------------------------------------

  testWidgets('the transport row carries a volume control', (tester) async {
    _sizeWindow(tester, const Size(1400, 1000));
    await tester.pumpWidget(_app());
    await _settle(tester);

    expect(find.byIcon(Icons.volume_up), findsOneWidget);
  });

  // ---------------------------------------------------------------------
  // v5.30.0 — Cover Flow artwork display mode
  // ---------------------------------------------------------------------

  group('Cover Flow wiring in FullPlayerScreen', () {
    testWidgets(
      'renders Cover Flow instead of the plain tile when enabled with a '
      'multi-track queue',
      (tester) async {
        _sizeWindow(tester, const Size(1400, 1000));
        await tester.pumpWidget(
          _appWithArtwork(albumId: 'album-x', queueLength: 4, coverFlow: true),
        );
        await _settle(tester);

        expect(find.byType(CoverFlowArtwork), findsOneWidget);
      },
    );

    testWidgets('falls back to the plain tile for a single-track queue even '
        'when Cover Flow is enabled', (tester) async {
      _sizeWindow(tester, const Size(1400, 1000));
      await tester.pumpWidget(
        _appWithArtwork(albumId: 'album-x', queueLength: 1, coverFlow: true),
      );
      await _settle(tester);

      expect(
        find.byType(CoverFlowArtwork),
        findsNothing,
        reason: 'A Cover Flow of one track has no neighbours to flow through',
      );
    });

    testWidgets('stays off the plain tile when the setting is off', (
      tester,
    ) async {
      _sizeWindow(tester, const Size(1400, 1000));
      await tester.pumpWidget(
        _appWithArtwork(albumId: 'album-x', queueLength: 4),
      );
      await _settle(tester);

      expect(find.byType(CoverFlowArtwork), findsNothing);
    });

    // -----------------------------------------------------------------
    // v5.30.5 — the gate above (canSplit && coverFlowModeProvider &&
    // queue.length > 1) was already correct per the three tests above; the
    // field report's "Cover Flow doesn't show up" turned out to be that
    // Settings -> Appearance was the *only* way to flip the provider, three
    // navigations from the screen it affects. This is that second, more
    // discoverable entry point living in the top bar itself.
    // -----------------------------------------------------------------

    testWidgets(
      'the top bar\'s Cover Flow toggle flips the same provider the Settings '
      'switch does',
      (tester) async {
        _sizeWindow(tester, const Size(1400, 1000));
        // Deliberately not passing coverFlow: true — this exercises the real
        // CoverFlowModeNotifier (off by default), not the always-on test
        // double the other three cases above use.
        await tester.pumpWidget(
          _appWithArtwork(albumId: 'album-x', queueLength: 4),
        );
        await _settle(tester);

        expect(find.byType(CoverFlowArtwork), findsNothing);

        await tester.tap(find.byIcon(Icons.view_carousel_outlined));
        await _settle(tester);

        expect(
          find.byType(CoverFlowArtwork),
          findsOneWidget,
          reason:
              'The button must actually toggle coverFlowModeProvider, not '
              'just sit there',
        );

        await tester.tap(find.byIcon(Icons.view_carousel_outlined));
        await _settle(tester);

        expect(
          find.byType(CoverFlowArtwork),
          findsNothing,
          reason: 'It must toggle back off again, not only ever turn on',
        );
      },
    );
  });
}
