// full_player_layout_test.dart
//
// Covers the v5.28.0 player-page layout: the macOS traffic-light gutter, the
// grouped transport row, and the centred/split behaviour that replaces the
// artwork-vs-lyrics PageView on wide windows.
//
import 'package:audio_service/audio_service.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/full_player_screen.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/lyrics/lyric_line.dart';
import 'package:inori_music/src/lyrics/lyrics_provider.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';

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

Widget _app() => ProviderScope(
  overrides: [
    playerProvider.overrideWith(() => _StubPlayerNotifier(_playing())),
    playbackEngineProvider.overrideWithValue(FakePlaybackEngine()),
    lyricsProvider.overrideWith(_StubLyricsNotifier.new),
  ],
  child: MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: const FullPlayerScreen(),
  ),
);

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
      closeTo((1400 - 380) / 2, 40),
      reason: 'And it should centre in the width that is left, not just move',
    );
  });
}
