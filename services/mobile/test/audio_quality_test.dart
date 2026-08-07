// audio_quality_test.dart
//
// Covers the v5.25.0 "HQ" badge predicate. The interesting cases are the two
// kinds of near-miss: a high-sample-rate *lossy* file (which the bitrate
// floor exists to exclude, standing in for the bit-depth check the metadata
// reader can't give us) and a track imported before v5.19.0 started
// capturing these fields at all.
//
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/local_library/audio_quality.dart';

void main() {
  test('24-bit/96 kHz FLAC territory qualifies', () {
    expect(isHiResAudio(sampleRate: 96000, bitrate: 3500000), isTrue);
  });

  test('16-bit/48 kHz lossless qualifies', () {
    expect(isHiResAudio(sampleRate: 48000, bitrate: 1400000), isTrue);
  });

  test('320 kbps MP3 at 48 kHz does not qualify', () {
    // The case the bit-depth check exists to exclude: high sample rate, but
    // lossy. Bitrate is what separates them here.
    expect(isHiResAudio(sampleRate: 48000, bitrate: 320000), isFalse);
  });

  test('CD-quality lossless does not qualify', () {
    expect(isHiResAudio(sampleRate: 44100, bitrate: 900000), isFalse);
  });

  test('exactly at both thresholds qualifies', () {
    expect(
      isHiResAudio(
        sampleRate: kHiResSampleRateHz,
        bitrate: kLosslessBitrateFloorBps,
      ),
      isTrue,
    );
  });

  test('one below either threshold does not', () {
    expect(
      isHiResAudio(
        sampleRate: kHiResSampleRateHz - 1,
        bitrate: kLosslessBitrateFloorBps,
      ),
      isFalse,
    );
    expect(
      isHiResAudio(
        sampleRate: kHiResSampleRateHz,
        bitrate: kLosslessBitrateFloorBps - 1,
      ),
      isFalse,
    );
  });

  test('unknown metadata never reads as Hi-Res', () {
    // Tracks imported before v5.19.0 have nulls here; "unknown" must not be
    // shown to the user as a quality claim.
    expect(isHiResAudio(sampleRate: null, bitrate: 3500000), isFalse);
    expect(isHiResAudio(sampleRate: 96000, bitrate: null), isFalse);
    expect(isHiResAudio(sampleRate: null, bitrate: null), isFalse);
  });
}
