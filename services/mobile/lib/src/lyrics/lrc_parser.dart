import 'package:inori_music/src/lyrics/lyric_line.dart';

/// Parses LRC-format lyrics into a list of [LyricLine]s sorted by timestamp.
///
/// Handles standard [MM:SS.xx] and [MM:SS.xxx] timestamps.
/// Also parses enhanced (word-level karaoke) LRC inline tags of the form
/// `<mm:ss.xx>word`, producing per-word timing fragments in [LyricLine.words].
/// Lines without inline tags keep [LyricLine.words] as null.
/// Metadata tags like [ti:], [ar:], [al:] are ignored.
class LrcParser {
  static final _lineTagRegex =
      RegExp(r'^\[(\d{2}):(\d{2})[.:](\d{2,3})\](.*)$');
  static final _wordTagRegex = RegExp(r'<(\d{2}):(\d{2})[.:](\d{2,3})>');

  static List<LyricLine> parse(String content) {
    final lines = <LyricLine>[];
    for (final raw in content.split('\n')) {
      final line = raw.trim();
      final m = _lineTagRegex.firstMatch(line);
      if (m == null) continue;
      final timestamp = _parseDuration(m.group(1)!, m.group(2)!, m.group(3)!);
      final rawText = (m.group(4) ?? '').trim();
      if (rawText.isEmpty) continue; // skip metadata-only tags

      final words = _parseWords(rawText, timestamp);
      final text = words != null ? words.map((w) => w.text).join().trim() : rawText;
      if (text.isEmpty) continue;
      lines.add(LyricLine(timestamp: timestamp, text: text, words: words));
    }
    lines.sort((a, b) => a.timestamp.compareTo(b.timestamp));
    return lines;
  }

  /// Splits a line's raw text on inline <mm:ss.xx> tags into word fragments.
  /// Returns null when the line has no inline tags (plain line-level lyrics).
  /// Text preceding the first inline tag, if any, is kept as a fragment
  /// timed at the line's own [lineTimestamp].
  static List<LyricWord>? _parseWords(String rawText, Duration lineTimestamp) {
    final matches = _wordTagRegex.allMatches(rawText).toList();
    if (matches.isEmpty) return null;
    final words = <LyricWord>[];
    if (matches.first.start > 0) {
      words.add(LyricWord(
        offset: lineTimestamp,
        text: rawText.substring(0, matches.first.start),
      ));
    }
    for (var i = 0; i < matches.length; i++) {
      final m = matches[i];
      final offset = _parseDuration(m.group(1)!, m.group(2)!, m.group(3)!);
      final start = m.end;
      final end = i + 1 < matches.length ? matches[i + 1].start : rawText.length;
      if (end <= start) continue;
      words.add(LyricWord(offset: offset, text: rawText.substring(start, end)));
    }
    return words.isEmpty ? null : words;
  }

  static Duration _parseDuration(String minStr, String secStr, String csStr) {
    final min = int.parse(minStr);
    final sec = int.parse(secStr);
    // normalise to milliseconds: 2 digits = centiseconds, 3 = milliseconds
    final ms = csStr.length == 2 ? int.parse(csStr) * 10 : int.parse(csStr);
    return Duration(minutes: min, seconds: sec, milliseconds: ms);
  }
}

