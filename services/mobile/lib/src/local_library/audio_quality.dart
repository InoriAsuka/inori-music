/// Sample rate at or above which a file is in Hi-Res territory. 48 kHz is the
/// conventional floor — anything at or under 44.1 kHz is CD quality at best.
const kHiResSampleRateHz = 48000;

/// Bitrate floor, in bits per second, used as a stand-in for "not a lossy
/// codec". Practical ceilings for lossy encoders sit well below this (MP3
/// tops out at 320 kbps, AAC and Opus rarely exceed ~500 kbps even at their
/// most extravagant), while even 16-bit/48 kHz FLAC lands comfortably above
/// it. That makes 700 kbps a clean separator without needing to know the
/// codec.
const kLosslessBitrateFloorBps = 700000;

/// Whether a track's parameters put it in Hi-Res territory, for the "HQ"
/// badge on the track detail panel.
///
/// OriginalSound HQ Player's own rule is sample rate >= 48 kHz **and** bit
/// depth >= 24. This project reads local file metadata through
/// `audio_metadata_reader`, whose public `AudioMetadata` model exposes
/// `sampleRate` and `bitrate` but **no bit depth at all** — getting one would
/// mean forking or replacing that package, which is wildly out of proportion
/// to a badge. So bitrate stands in for bit depth here: it is what separates
/// a 24-bit lossless file from a high-sample-rate lossy one, which is the
/// case the depth check exists to exclude.
///
/// Both values must be known. A track imported before v5.19.0 started
/// capturing them has nulls, and "unknown" must not read as "Hi-Res".
bool isHiResAudio({int? sampleRate, int? bitrate}) {
  if (sampleRate == null || bitrate == null) return false;
  return sampleRate >= kHiResSampleRateHz &&
      bitrate >= kLosslessBitrateFloorBps;
}
