// playback_boundary_test.dart
//
// Architecture test for the v5.27.0 playback seam. The seam's whole value is
// that it holds: the moment a screen or notifier imports just_audio again, or
// reaches into main.dart for a singleton, swapping the engine stops being a
// contained change — which is the situation this refactor existed to end.
//
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// Files allowed to know which audio package is in use. Adding an engine
/// means adding its file here, not loosening the rule.
const _engineImplementations = {'lib/src/playback/just_audio_engine.dart'};

Iterable<File> _dartFilesUnder(String path) => Directory(path)
    .listSync(recursive: true)
    .whereType<File>()
    .where((f) => f.path.endsWith('.dart'));

void main() {
  test('only the engine implementation imports just_audio', () {
    final offenders = <String>[];
    for (final file in _dartFilesUnder('lib')) {
      final relative = file.path.replaceFirst(RegExp(r'^.*/mobile/'), '');
      if (_engineImplementations.contains(relative)) continue;
      // The generated API client is vendored and irrelevant here.
      if (relative.startsWith('lib/src/api/generated/')) continue;
      if (file.readAsStringSync().contains('package:just_audio/')) {
        offenders.add(relative);
      }
    }

    expect(
      offenders,
      isEmpty,
      reason:
          'These reach past PlaybackEngine to just_audio directly. Add the '
          'capability to PlaybackEngine instead — otherwise the next engine '
          'has to satisfy an interface nobody is actually using.',
    );
  });

  test('nothing imports main.dart for a singleton', () {
    final offenders = <String>[];
    for (final file in _dartFilesUnder('lib')) {
      final relative = file.path.replaceFirst(RegExp(r'^.*/mobile/'), '');
      if (relative == 'lib/main.dart') continue;
      // Match the import, not prose about it — the doc comment on
      // playbackEngineProvider names the old pattern deliberately.
      if (RegExp(
        "^import 'package:inori_music/main.dart'",
        multiLine: true,
      ).hasMatch(file.readAsStringSync())) {
        offenders.add(relative);
      }
    }

    expect(
      offenders,
      isEmpty,
      reason:
          'main.dart is the composition root; importing it back down makes '
          'the importer untestable without booting the whole app.',
    );
  });
}
