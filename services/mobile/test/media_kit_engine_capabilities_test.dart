// media_kit_engine_capabilities_test.dart
//
// Asserts what the media_kit engine declares through PlaybackCapabilities,
// without constructing a real MediaKitEngine — that would construct a real
// Player, which reaches for libmpv itself. This repo deliberately bundles
// libmpv for Windows only (see pubspec.yaml: media_kit_libs_windows_audio),
// so a real Player cannot be built on the macOS/Linux machines this test
// suite actually runs on. mediaKitCapabilities is a top-level constant for
// exactly this reason — see its doc comment in media_kit_engine.dart.
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/playback/media_kit_engine.dart';

void main() {
  test('media_kit capabilities are reported honestly for what is wired end '
      'to end', () {
    // Actually wired.
    expect(mediaKitCapabilities.speedControl, isTrue);
    expect(mediaKitCapabilities.gapless, isTrue);

    // Deliberately not implemented in this transitional engine — see the
    // doc comment on mediaKitCapabilities for why (most likely to be
    // discarded once the in-house engine lands).
    expect(mediaKitCapabilities.equalizer, isFalse);
    expect(mediaKitCapabilities.crossfade, isFalse);

    // Interface gap, not an engine limitation: libmpv can do all three
    // (audio-device, --audio-exclusive, --audio-format/--audio-samplerate),
    // but PlaybackEngine has no enumerate/select-device method yet for an
    // engine to implement them against. See mediaKitCapabilities' own doc
    // comment.
    expect(mediaKitCapabilities.outputDeviceSelection, isFalse);
    expect(mediaKitCapabilities.exclusiveOutput, isFalse);
    expect(mediaKitCapabilities.outputFormatControl, isFalse);
  });
}
