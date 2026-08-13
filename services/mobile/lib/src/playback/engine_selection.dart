/// Picks which concrete `PlaybackEngine` implementation `main()` should
/// construct for the current platform.
///
/// A pure function of the platform name (`Platform.operatingSystem`) rather
/// than a place that reads `Platform.isWindows` inline — the decision itself
/// is then unit-testable without booting a real audio stack or faking
/// `dart:io` (see engine_selection_test.dart). `main()` is the only caller
/// and the only place allowed to ask what platform this actually is; this
/// file deliberately has no imports at all, not even `playback_engine.dart`,
/// so nothing about testing it ever needs more than a string in, an enum
/// out.
enum EngineKind {
  /// `just_audio` — macOS, Linux, Android, iOS. Unchanged by this file.
  justAudio,

  /// `media_kit` (libmpv) — Windows only. See media_kit_engine.dart's own
  /// doc comment for why: `just_audio` registers no Windows implementation
  /// at all, so Windows is the one platform where the choice differs.
  mediaKit,
}

EngineKind choosePlaybackEngineKind(String operatingSystem) =>
    operatingSystem == 'windows' ? EngineKind.mediaKit : EngineKind.justAudio;
