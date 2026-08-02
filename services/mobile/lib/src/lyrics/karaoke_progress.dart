import 'package:inori_music/src/lyrics/lyric_line.dart';

/// Karaoke progress math, mirrored from
/// `services/web/lib/karaoke/progress.ts` so both clients highlight
/// identically for the same lyrics and playback position.

/// Index of the last line whose timestamp has been reached, or -1 before
/// the first line starts.
int activeLineIndex(List<LyricLine> lines, Duration position) {
  var idx = -1;
  for (var i = 0; i < lines.length; i++) {
    if (lines[i].timestamp <= position) {
      idx = i;
    } else {
      break;
    }
  }
  return idx;
}

/// Fill ratio (0..1) for the word at [index], used to paint the sung portion.
///
/// A word runs until the next word starts, or — for the last word of a line —
/// until [lineEnd] (the next line's timestamp). With no following line, fall
/// back to a fixed tail so the final word still animates instead of snapping.
double wordProgress(
  List<LyricWord> words,
  int index,
  Duration position,
  Duration? lineEnd,
) {
  if (index < 0 || index >= words.length) return 0;
  final word = words[index];
  if (position <= word.offset) return 0;

  final next = index + 1 < words.length ? words[index + 1] : null;
  final end = next?.offset ?? lineEnd ?? (word.offset + _finalWordTail);
  final span = end - word.offset;
  if (span <= Duration.zero) return 1;

  final elapsed = position - word.offset;
  final ratio = elapsed.inMicroseconds / span.inMicroseconds;
  return ratio.clamp(0.0, 1.0);
}

const _finalWordTail = Duration(milliseconds: 800);
