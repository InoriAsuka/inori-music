// user_playlist_detail_screen_test.dart
//
// v5.30.7: this screen had no test coverage at all before this phase. It
// earns a file now specifically because its "Play All" FilledButton.icon
// sits as a plain Row sibling next to a Spacer (`Row(children: [Text(...),
// const Spacer(), FilledButton.icon(...)])`) — the exact shape that trips a
// hard "BoxConstraints forces an infinite width" layout assertion under the
// app's pre-v5.30.7 global FilledButtonTheme (`minimumSize:
// Size(double.infinity, 48)`), confirmed live while investigating that fix.
// skin_definition_test.dart's own dedicated group guards the theme itself;
// this proves the fix actually reaches this real, previously-untested call
// site rather than only the synthetic repros used to find the bug.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_definition.dart';
import 'package:inori_music/src/user_playlist/user_playlist_detail_screen.dart';
import 'package:inori_music/src/user_playlist/user_playlist_notifier.dart';

class _StubUserPlaylistNotifier extends UserPlaylistNotifier {
  _StubUserPlaylistNotifier(this._trackIds);
  final List<String> _trackIds;

  @override
  Future<List<UserPlaylist>> build() async => const [];

  @override
  Future<List<String>> getTrackIds(String playlistId) async => _trackIds;
}

class _StubPlayerNotifier extends PlayerNotifier {
  List<String>? lastQueue;

  @override
  pstate.PlayerState build() => pstate.PlayerState();

  @override
  Future<void> playQueue(List<String> trackIds, {int initialIndex = 0}) async {
    lastQueue = List.of(trackIds);
  }
}

const _detailPath = '/detail';

Widget _buildApp(
  _StubUserPlaylistNotifier playlists,
  _StubPlayerNotifier player,
) {
  final router = GoRouter(
    initialLocation: _detailPath,
    routes: [
      GoRoute(
        path: _detailPath,
        builder: (_, _) => const UserPlaylistDetailScreen(playlistId: 'pl-1'),
      ),
      // "Play All" navigates here once playback starts (see
      // user_playlist_detail_screen.dart) — needs somewhere to land.
      GoRoute(
        path: AppRoutes.player,
        builder: (_, _) => const Scaffold(body: Text('body:player')),
      ),
    ],
  );
  return ProviderScope(
    overrides: [
      userPlaylistProvider.overrideWith(() => playlists),
      playerProvider.overrideWith(() => player),
    ],
    child: MaterialApp.router(
      // Real skin theme, not Flutter's own MaterialApp default — see
      // play_actions_row_test.dart's own _buildApp for why this matters.
      theme: buildThemeFromSkin(SkinDefinition.sakuraDusk),
      routerConfig: router,
    ),
  );
}

void main() {
  testWidgets(
    'a non-empty playlist renders its Play All toolbar and enqueues every '
    'track, under the real app theme',
    (tester) async {
      final playlists = _StubUserPlaylistNotifier(['t1', 't2', 't3']);
      final player = _StubPlayerNotifier();
      await tester.pumpWidget(_buildApp(playlists, player));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Play All'), findsOneWidget);

      await tester.tap(find.text('Play All'));
      await tester.pumpAndSettle();

      expect(player.lastQueue, ['t1', 't2', 't3']);
      expect(find.text('body:player'), findsOneWidget);
    },
  );

  testWidgets('an empty playlist shows the empty-state message instead of the '
      'toolbar (so there is nothing for the Play All button crash to hide '
      'behind)', (tester) async {
    final playlists = _StubUserPlaylistNotifier(const []);
    await tester.pumpWidget(_buildApp(playlists, _StubPlayerNotifier()));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.text('No tracks in this playlist'), findsOneWidget);
    expect(find.text('Play All'), findsNothing);
  });
}
