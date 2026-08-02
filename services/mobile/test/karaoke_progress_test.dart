import 'package:flutter_test/flutter_test.dart';
import 'package:inori_music/src/lyrics/karaoke_progress.dart';
import 'package:inori_music/src/lyrics/lyric_line.dart';

LyricLine _line(int ms, String text, [List<LyricWord>? words]) => LyricLine(
      timestamp: Duration(milliseconds: ms),
      text: text,
      words: words,
    );

LyricWord _word(int ms, String text) =>
    LyricWord(offset: Duration(milliseconds: ms), text: text);

void main() {
  group('activeLineIndex', () {
    final lines = [_line(0, 'a'), _line(1000, 'b'), _line(2000, 'c')];

    test('returns -1 before the first line starts', () {
      expect(activeLineIndex([_line(500, 'a')], Duration.zero), -1);
    });

    test('returns the last reached line', () {
      expect(activeLineIndex(lines, const Duration(milliseconds: 1500)), 1);
    });

    test('is inclusive of the exact timestamp', () {
      expect(activeLineIndex(lines, const Duration(milliseconds: 1000)), 1);
    });

    test('stays on the final line past the end', () {
      expect(activeLineIndex(lines, const Duration(seconds: 99)), 2);
    });

    test('handles an empty list', () {
      expect(activeLineIndex([], const Duration(seconds: 1)), -1);
    });
  });

  group('wordProgress', () {
    final words = [_word(1000, 'one'), _word(2000, 'two')];

    test('is 0 before the word starts', () {
      expect(wordProgress(words, 0, const Duration(milliseconds: 500), null), 0);
    });

    test('is 0 exactly at the word start', () {
      expect(wordProgress(words, 0, const Duration(milliseconds: 1000), null), 0);
    });

    test('interpolates linearly to the next word', () {
      expect(
        wordProgress(words, 0, const Duration(milliseconds: 1500), null),
        closeTo(0.5, 1e-9),
      );
    });

    test('is 1 once the next word has started', () {
      expect(wordProgress(words, 0, const Duration(milliseconds: 2500), null), 1);
    });

    test('last word uses the next line start as its end', () {
      expect(
        wordProgress(words, 1, const Duration(milliseconds: 2500),
            const Duration(milliseconds: 3000)),
        closeTo(0.5, 1e-9),
      );
    });

    test('last word falls back to a fixed tail without a next line', () {
      // Tail is 800ms, so +400ms is half way.
      expect(
        wordProgress(words, 1, const Duration(milliseconds: 2400), null),
        closeTo(0.5, 1e-9),
      );
    });

    test('never exceeds 1 past the tail', () {
      expect(wordProgress(words, 1, const Duration(seconds: 60), null), 1);
    });

    test('returns 1 for a zero-length span rather than dividing by zero', () {
      final same = [_word(1000, 'a'), _word(1000, 'b')];
      expect(wordProgress(same, 0, const Duration(milliseconds: 1200), null), 1);
    });

    test('out-of-range index is 0', () {
      expect(wordProgress(words, 9, const Duration(seconds: 5), null), 0);
      expect(wordProgress(words, -1, const Duration(seconds: 5), null), 0);
    });
  });
}
