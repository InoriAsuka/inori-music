/// A single line of lyrics with its start timestamp.
class LyricLine {
  const LyricLine({required this.timestamp, required this.text});
  final Duration timestamp;
  final String text;
}
