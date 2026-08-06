import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/lyrics/local_lyrics_provider.dart';

// LocalLyricsNotifier itself touches LocalLibraryDb (sqflite/platform
// channels — same constraint noted in local_library_db_test.dart), so this
// only covers parseEmbeddedLyrics, the pure LRC-vs-plain-text branch pulled
// out specifically to be testable without a real database.

void main() {
  group('parseEmbeddedLyrics', () {
    test('returns null for null input', () {
      expect(parseEmbeddedLyrics(null), isNull);
    });

    test('returns null for blank/whitespace-only input', () {
      expect(parseEmbeddedLyrics('   \n  '), isNull);
    });

    test('parses standard LRC timestamps into timed lines', () {
      const raw = '[00:01.00]La la la\n[00:05.50]Da da da';
      final lines = parseEmbeddedLyrics(raw);
      expect(lines, isNotNull);
      expect(lines!.length, 2);
      expect(lines[0].timestamp, const Duration(seconds: 1));
      expect(lines[0].text, 'La la la');
      expect(lines[1].timestamp, const Duration(seconds: 5, milliseconds: 500));
    });

    test(
      'falls back to a single untimed block when no LRC timestamps match',
      () {
        const raw = 'Just some plain lyrics\nwith no timing tags at all';
        final lines = parseEmbeddedLyrics(raw);
        expect(lines, isNotNull);
        expect(lines!.length, 1);
        expect(lines[0].timestamp, Duration.zero);
        expect(lines[0].text, raw);
      },
    );
  });
}
