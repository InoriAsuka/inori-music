// playback_mode_buttons_test.dart
//
// Covers RepeatModeButton/ShuffleButton — pulled out of
// full_player_screen.dart in v5.30.5 so the desktop mini player bar could
// reuse the exact same repeat-cycle/shuffle-toggle semantics instead of
// growing a second implementation that could drift from the full player's.
// These tests exercise the shared widgets directly against a stub notifier
// that overrides setRepeat/setShuffle to update state without touching
// PlaybackEngine (the real implementations also forward to the engine's
// setRepeatMode/setShuffleEnabled, which is a separate concern already
// covered at the notifier level in player_notifier_test.dart) — the point
// here is proving the widget extraction preserved the *cycle order and
// tinting* behaviour, not re-verifying the engine call.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/player/playback_mode_buttons.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/shared/theme/skin_definition.dart';

class _StubPlayerNotifier extends PlayerNotifier {
  @override
  pstate.PlayerState build() => pstate.PlayerState();

  @override
  Future<void> setRepeat(pstate.RepeatMode repeat) async {
    state = state.copyWith(repeat: repeat);
  }

  @override
  Future<void> setShuffle(bool shuffle) async {
    state = state.copyWith(shuffle: shuffle);
  }
}

Widget _app(Widget child) => ProviderScope(
  overrides: [playerProvider.overrideWith(_StubPlayerNotifier.new)],
  child: MaterialApp(home: Scaffold(body: child)),
);

// No SkinScope ancestor in this harness, so context.skinColors falls back to
// the default skin (see SkinScope.of) — these are the exact colours the
// buttons resolve to in that fallback.
final _active = SkinDefinition.sakuraDusk.colors.sakuraPinkLight;
final _inactive = SkinDefinition.sakuraDusk.colors.onSurfaceVariant;

void main() {
  group('RepeatModeButton', () {
    testWidgets('cycles none -> all -> one -> none on successive taps', (
      tester,
    ) async {
      await tester.pumpWidget(_app(const RepeatModeButton()));
      await tester.pump();

      expect(find.byTooltip('Repeat: none'), findsOneWidget);

      await tester.tap(find.byIcon(Icons.repeat));
      await tester.pump();
      expect(find.byTooltip('Repeat: all'), findsOneWidget);

      await tester.tap(find.byIcon(Icons.repeat));
      await tester.pump();
      expect(find.byTooltip('Repeat: one'), findsOneWidget);

      await tester.tap(find.byIcon(Icons.repeat));
      await tester.pump();
      expect(
        find.byTooltip('Repeat: none'),
        findsOneWidget,
        reason: 'The cycle must wrap back to none, not stop at one',
      );
    });

    testWidgets('icon tints with the accent whenever repeat is not none', (
      tester,
    ) async {
      await tester.pumpWidget(_app(const RepeatModeButton()));
      await tester.pump();

      expect(tester.widget<Icon>(find.byIcon(Icons.repeat)).color, _inactive);

      await tester.tap(find.byIcon(Icons.repeat));
      await tester.pump();
      expect(tester.widget<Icon>(find.byIcon(Icons.repeat)).color, _active);

      await tester.tap(find.byIcon(Icons.repeat));
      await tester.pump();
      expect(
        tester.widget<Icon>(find.byIcon(Icons.repeat)).color,
        _active,
        reason: 'RepeatMode.one is still "repeat active", same as .all',
      );
    });
  });

  group('ShuffleButton', () {
    testWidgets('toggles on and off, tinting the icon while active', (
      tester,
    ) async {
      await tester.pumpWidget(_app(const ShuffleButton()));
      await tester.pump();

      expect(tester.widget<Icon>(find.byIcon(Icons.shuffle)).color, _inactive);

      await tester.tap(find.byIcon(Icons.shuffle));
      await tester.pump();
      expect(tester.widget<Icon>(find.byIcon(Icons.shuffle)).color, _active);

      await tester.tap(find.byIcon(Icons.shuffle));
      await tester.pump();
      expect(tester.widget<Icon>(find.byIcon(Icons.shuffle)).color, _inactive);
    });
  });

  testWidgets('iconSize is applied to both buttons when provided', (
    tester,
  ) async {
    await tester.pumpWidget(
      _app(
        const Row(
          children: [
            RepeatModeButton(iconSize: 22),
            ShuffleButton(iconSize: 22),
          ],
        ),
      ),
    );
    await tester.pump();

    expect(tester.widget<Icon>(find.byIcon(Icons.repeat)).size, 22);
    expect(tester.widget<Icon>(find.byIcon(Icons.shuffle)).size, 22);
  });
}
