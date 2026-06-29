import 'package:inori_music/src/lyrics/lyric_line.dart';

/// Parses SRT-format subtitles into a list of [LyricLine]s (start-time only).
class SrtParser {
  static List<LyricLine> parse(String content) {
    final lines = <LyricLine>[];
    // SRT blocks are separated by blank lines
    final blocks = content.split(RegExp(r'\r?\n\r?\n'));
    final timeRegex = RegExp(
      r'(\d{2}):(\d{2}):(\d{2})[,.](\d{3})\s*-->\s*',
    );
    for (final block in blocks) {
      final blockLines = block.trim().split(RegExp(r'\r?\n'));
      if (blockLines.length < 2) continue;
      // First non-empty line is the sequence number — skip it
      // Second line is the time range
      String? timeLine;
      int textStart = 1;
      for (int i = 0; i < blockLines.length; i++) {
        final m = timeRegex.firstMatch(blockLines[i]);
        if (m != null) {
          timeLine = blockLines[i];
          textStart = i + 1;
          break;
        }
      }
      if (timeLine == null) continue;
      final m = timeRegex.firstMatch(timeLine)!;
      final h = int.parse(m.group(1)!);
      final min = int.parse(m.group(2)!);
      final sec = int.parse(m.group(3)!);
      final ms = int.parse(m.group(4)!);
      final text = blockLines.sublist(textStart).join(' ').trim();
      if (text.isEmpty) continue;
      lines.add(LyricLine(
        timestamp: Duration(
            hours: h, minutes: min, seconds: sec, milliseconds: ms),
        text: text,
      ));
    }
    lines.sort((a, b) => a.timestamp.compareTo(b.timestamp));
    return lines;
  }
}
