// volume_control_test.dart
//
// Covers VolumeControl — PlayerNotifier.setVolume has existed since before
// v5.30.0 but nothing in the UI ever called it; this is that first UI. Two
// shapes are exercised: the inline icon+slider (mini player bar's default)
// and the compact icon-only trigger that opens the same slider in a popover
// (mini bar under width pressure, and the full player screen always).
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/player/volume_control.dart';

// ---------------------------------------------------------------------------
// Stub PlayerNotifier — overrides setVolume to record calls and update state
// directly, the same way mini_player_bar_test.dart's stub handles
// togglePlayPause. Avoids touching PlaybackEngine entirely (setVolume's real
// implementation calls through to it for ReplayGain-adjusted output), which
// isn't the concern of these widget-level tests.
// ---------------------------------------------------------------------------
class _StubPlayerNotifier extends PlayerNotifier {
  _StubPlayerNotifier([double initialVolume = 1.0])
    : _initialVolume = initialVolume;
  final double _initialVolume;
  final volumeCalls = <double>[];

  @override
  pstate.PlayerState build() => pstate.PlayerState(volume: _initialVolume);

  @override
  Future<void> setVolume(double volume) async {
    volumeCalls.add(volume);
    state = state.copyWith(volume: volume);
  }
}

Widget _app(_StubPlayerNotifier stub, Widget child) => ProviderScope(
  overrides: [playerProvider.overrideWith(() => stub)],
  child: MaterialApp(home: Scaffold(body: child)),
);

void main() {
  group('expanded (inline) mode', () {
    testWidgets('renders an inline slider, no popover needed', (tester) async {
      final stub = _StubPlayerNotifier();
      await tester.pumpWidget(_app(stub, const VolumeControl()));
      await tester.pump();

      expect(find.byType(Slider), findsOneWidget);
    });

    testWidgets('dragging the slider calls setVolume with the new value', (
      tester,
    ) async {
      final stub = _StubPlayerNotifier();
      await tester.pumpWidget(_app(stub, const VolumeControl()));
      await tester.pump();

      tester.widget<Slider>(find.byType(Slider)).onChanged!(0.3);
      await tester.pump();

      expect(stub.volumeCalls, contains(0.3));
    });

    for (final (volume, icon) in [
      (0.0, Icons.volume_off),
      (0.3, Icons.volume_down),
      (0.8, Icons.volume_up),
    ]) {
      testWidgets('shows ${icon.toString()} at volume $volume', (tester) async {
        final stub = _StubPlayerNotifier(volume);
        await tester.pumpWidget(_app(stub, const VolumeControl()));
        await tester.pump();

        expect(find.byIcon(icon), findsOneWidget);
      });
    }

    testWidgets('tapping the icon mutes, remembering the current volume', (
      tester,
    ) async {
      final stub = _StubPlayerNotifier(0.4);
      await tester.pumpWidget(_app(stub, const VolumeControl()));
      await tester.pump();

      await tester.tap(find.byIcon(Icons.volume_down));
      await tester.pump();

      expect(stub.volumeCalls.last, 0.0);
      expect(find.byIcon(Icons.volume_off), findsOneWidget);
    });

    testWidgets(
      'tapping again restores the exact remembered volume, not a hardcoded '
      'default',
      (tester) async {
        // 0.4 rather than 1.0 is the point: a naive "unmute -> full volume"
        // implementation would pass this test's mute step but fail the
        // restore, landing on 1.0 instead of 0.4.
        final stub = _StubPlayerNotifier(0.4);
        await tester.pumpWidget(_app(stub, const VolumeControl()));
        await tester.pump();

        await tester.tap(find.byIcon(Icons.volume_down));
        await tester.pump();
        await tester.tap(find.byIcon(Icons.volume_off));
        await tester.pump();

        expect(stub.volumeCalls.last, 0.4);
      },
    );

    testWidgets(
      'unmuting from a session that never muted restores full volume',
      (tester) async {
        final stub = _StubPlayerNotifier(0.0);
        await tester.pumpWidget(_app(stub, const VolumeControl()));
        await tester.pump();

        await tester.tap(find.byIcon(Icons.volume_off));
        await tester.pump();

        expect(stub.volumeCalls.last, 1.0);
      },
    );

    testWidgets('mute/restore round-trips across two independent VolumeControl '
        'instances sharing the same ProviderScope', (tester) async {
      final stub = _StubPlayerNotifier(0.6);
      await tester.pumpWidget(
        _app(
          stub,
          const Column(
            children: [
              VolumeControl(key: ValueKey('a')),
              VolumeControl(key: ValueKey('b')),
            ],
          ),
        ),
      );
      await tester.pump();

      // Mute via the first instance's icon (0.6 renders volume_up — see
      // _iconFor's 0.5 threshold).
      await tester.tap(
        find.descendant(
          of: find.byKey(const ValueKey('a')),
          matching: find.byIcon(Icons.volume_up),
        ),
      );
      await tester.pump();
      expect(stub.volumeCalls.last, 0.0);

      // Unmute via the *second* instance — it must restore 0.6, proving
      // the "remembered" value lives in shared provider state rather than
      // each widget's own local memory (see volumeBeforeMuteProvider's
      // doc comment for why that sharing matters).
      await tester.tap(
        find.descendant(
          of: find.byKey(const ValueKey('b')),
          matching: find.byIcon(Icons.volume_off),
        ),
      );
      await tester.pump();
      expect(stub.volumeCalls.last, 0.6);
    });
  });

  group('compact mode', () {
    testWidgets('renders only an icon button, no inline slider', (
      tester,
    ) async {
      final stub = _StubPlayerNotifier();
      await tester.pumpWidget(_app(stub, const VolumeControl(compact: true)));
      await tester.pump();

      expect(find.byType(Slider), findsNothing);
      expect(find.byType(IconButton), findsOneWidget);
    });

    testWidgets('tapping the icon opens a popover with the slider', (
      tester,
    ) async {
      final stub = _StubPlayerNotifier();
      await tester.pumpWidget(_app(stub, const VolumeControl(compact: true)));
      await tester.pump();

      await tester.tap(find.byType(IconButton));
      await tester.pumpAndSettle();

      expect(find.byType(Slider), findsOneWidget);
    });

    testWidgets('the popover slider still drives setVolume', (tester) async {
      final stub = _StubPlayerNotifier();
      await tester.pumpWidget(_app(stub, const VolumeControl(compact: true)));
      await tester.pump();

      await tester.tap(find.byType(IconButton));
      await tester.pumpAndSettle();

      tester.widget<Slider>(find.byType(Slider)).onChanged!(0.2);
      await tester.pump();

      expect(stub.volumeCalls, contains(0.2));
    });
  });
}
