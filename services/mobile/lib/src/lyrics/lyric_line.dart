/// A single word-level timing fragment within a [LyricLine], used for
/// karaoke-style progressive highlighting.
class LyricWord {
  const LyricWord({required this.offset, required this.text});
  final Duration offset;
  final String text;
}

/// A single line of lyrics with its start timestamp.
class LyricLine {
  const LyricLine({
    required this.timestamp,
    required this.text,
    this.words,
    this.translation,
  });
  final Duration timestamp;
  final String text;
  /// Word-level timing fragments for karaoke highlighting, or null/empty
  /// when the source line has no inline `<mm:ss.xx>` tags.
  final List<LyricWord>? words;
  /// Aligned translation text for this line, or null when no translation
  /// was uploaded or the translation line count didn't match.
  final String? translation;
}

