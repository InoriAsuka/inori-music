import 'package:inori_music/src/lyrics/lyric_line.dart';

/// Parses LRC-format lyrics into a list of [LyricLine]s sorted by timestamp.
///
/// Handles standard [MM:SS.xx] and [MM:SS.xxx] timestamps.
/// Metadata tags like [ti:], [ar:], [al:] are ignored.
class LrcParser {
  static List<LyricLine> parse(String content) {
    final lines = <LyricLine>[];
    final tagRegex = RegExp(r'^\[(\d{2}):(\d{2})[.:](\d{2,3})\](.*)$');
    for (final raw in content.split('\n')) {
      final line = raw.trim();
      final m = tagRegex.firstMatch(line);
      if (m == null) continue;
      final min = int.parse(m.group(1)!);
      final sec = int.parse(m.group(2)!);
      final csStr = m.group(3)!;
      // normalise to milliseconds: 2 digits = centiseconds, 3 = milliseconds
      final ms = csStr.length == 2
          ? int.parse(csStr) * 10
          : int.parse(csStr);
      final text = (m.group(4) ?? '').trim();
      if (text.isEmpty) continue; // skip metadata-only tags
      lines.add(LyricLine(
        timestamp: Duration(minutes: min, seconds: sec, milliseconds: ms),
        text: text,
      ));
    }
    lines.sort((a, b) => a.timestamp.compareTo(b.timestamp));
    return lines;
  }
}
