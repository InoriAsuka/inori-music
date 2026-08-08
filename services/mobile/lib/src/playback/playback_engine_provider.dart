import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/playback/playback_engine.dart';

/// The app's playback engine.
///
/// Replaces `import 'package:inori_music/main.dart' show audioHandler` — four
/// notifiers reached into `main.dart` for a global singleton, which made them
/// untestable without booting the real audio stack and tied every one of them
/// to `just_audio` by transitive import.
///
/// Deliberately throws instead of constructing a default: the engine has to
/// be built before `runApp` (it initialises `audio_service`), so a missing
/// override is a wiring bug that should fail loudly at startup rather than
/// quietly spin up a second audio player nobody is listening to. Tests
/// override it with a fake.
final playbackEngineProvider = Provider<PlaybackEngine>((ref) {
  throw UnimplementedError(
    'playbackEngineProvider must be overridden — see main().',
  );
});

/// What the current engine supports, for UI that needs to hide controls it
/// cannot honour. Reads better at call sites than
/// `ref.watch(playbackEngineProvider).capabilities`, and gives that lookup one
/// place to change.
final playbackCapabilitiesProvider = Provider<PlaybackCapabilities>(
  (ref) => ref.watch(playbackEngineProvider).capabilities,
);
