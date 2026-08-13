// engine_selection_test.dart
//
// choosePlaybackEngineKind is the pure function main() uses to decide which
// PlaybackEngine implementation to construct. Extracted specifically so this
// decision is testable without booting a real audio stack or faking
// dart:io's Platform.isWindows — pass any operating-system string in, assert
// the engine choice out.
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/playback/engine_selection.dart';

void main() {
  test('windows picks media_kit', () {
    expect(choosePlaybackEngineKind('windows'), EngineKind.mediaKit);
  });

  test('every other platform keeps just_audio', () {
    for (final os in ['macos', 'linux', 'android', 'ios', 'fuchsia']) {
      expect(
        choosePlaybackEngineKind(os),
        EngineKind.justAudio,
        reason:
            '$os must not be switched to media_kit — just_audio only '
            'has zero implementation on Windows; every other platform '
            'already works and macOS in particular was only just repaired '
            '(v5.37.0–v5.37.2).',
      );
    }
  });

  test(
    'an unrecognised platform string falls back to just_audio, not a throw',
    () {
      // Defends the shape of the function itself: it is a total function over
      // any String, not a switch that could hit an unmatched-case error if a
      // future dart:io ever reports something this list didn't anticipate.
      expect(choosePlaybackEngineKind('plan9'), EngineKind.justAudio);
    },
  );
}
