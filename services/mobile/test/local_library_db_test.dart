import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/local_library/local_library_db.dart';

// LocalLibraryDb itself touches sqflite/platform channels (same constraint
// noted in background_provider_test.dart), so this only covers the two
// things that don't: the ORDER BY clause per sort dimension, and the
// LocalLibraryTrack toMap/fromMap round-trip — including the v5.19.0 columns
// being genuinely optional (pre-migration rows have them as null).

void main() {
  group('localLibraryOrderByClause', () {
    test('artistAlbumTitle sorts by artist, then album, then title', () {
      expect(
        localLibraryOrderByClause(LocalLibrarySortOrder.artistAlbumTitle),
        'artist_name COLLATE NOCASE, album_title COLLATE NOCASE, title COLLATE NOCASE',
      );
    });

    test('recentlyAdded sorts by import time, newest first', () {
      expect(
        localLibraryOrderByClause(LocalLibrarySortOrder.recentlyAdded),
        'imported_at DESC',
      );
    });

    test('titleAZ sorts case-insensitively by title', () {
      expect(
        localLibraryOrderByClause(LocalLibrarySortOrder.titleAZ),
        'title COLLATE NOCASE',
      );
    });

    test('duration sorts longest first', () {
      expect(
        localLibraryOrderByClause(LocalLibrarySortOrder.duration),
        'duration_ms DESC',
      );
    });
  });

  group('LocalLibraryTrack.toMap / fromMap', () {
    LocalLibraryTrack fullTrack() => LocalLibraryTrack(
      id: 'local:abc',
      title: 'Emmanuel',
      artistName: 'Chris Botti',
      albumTitle: 'To Love Again',
      localPath: '/app/local_library_audio/local:abc.flac',
      durationMs: 245000,
      coverArtPath: '/app/local_library_covers/local:abc.jpg',
      sizeBytes: 42 * 1024 * 1024,
      importedAt: DateTime.utc(2026, 8, 6),
      sampleRate: 44100,
      bitrate: 1411000,
      fileFormat: 'FLAC',
      genre: 'Jazz',
      trackNumber: 3,
      embeddedLyrics: '[00:01.00]La la la',
    );

    test('round-trips every field, including the v5.19.0 additions', () {
      final track = fullTrack();
      final restored = LocalLibraryTrack.fromMap(track.toMap());

      expect(restored.id, track.id);
      expect(restored.title, track.title);
      expect(restored.artistName, track.artistName);
      expect(restored.albumTitle, track.albumTitle);
      expect(restored.localPath, track.localPath);
      expect(restored.durationMs, track.durationMs);
      expect(restored.coverArtPath, track.coverArtPath);
      expect(restored.sizeBytes, track.sizeBytes);
      // Compare the instant, not DateTime equality directly — fromMap()
      // reconstructs via DateTime.fromMillisecondsSinceEpoch() without
      // isUtc:true, so the restored value is "the same moment, as a local
      // DateTime" rather than being `==` to a UTC one (pre-existing
      // behavior from v5.12.0, unrelated to this phase's changes).
      expect(
        restored.importedAt.millisecondsSinceEpoch,
        track.importedAt.millisecondsSinceEpoch,
      );
      expect(restored.sampleRate, track.sampleRate);
      expect(restored.bitrate, track.bitrate);
      expect(restored.fileFormat, track.fileFormat);
      expect(restored.genre, track.genre);
      expect(restored.trackNumber, track.trackNumber);
      expect(restored.embeddedLyrics, track.embeddedLyrics);
    });

    test(
      'a pre-v5.19.0 row (new columns absent from the map) restores with nulls',
      () {
        // Simulates a row inserted before the v2 migration: toMap() always
        // writes every key, but a raw sqflite result row from before the
        // ALTER TABLEs simply wouldn't have these keys at all.
        final legacyMap = {
          'id': 'local:old',
          'title': 'Old Track',
          'artist_name': '',
          'album_title': '',
          'local_path': '/app/local_library_audio/local:old.mp3',
          'duration_ms': null,
          'cover_art_path': null,
          'size_bytes': 1024,
          'imported_at': DateTime.utc(2026, 1, 1).millisecondsSinceEpoch,
        };

        final restored = LocalLibraryTrack.fromMap(legacyMap);

        expect(restored.sampleRate, isNull);
        expect(restored.bitrate, isNull);
        expect(restored.fileFormat, isNull);
        expect(restored.genre, isNull);
        expect(restored.trackNumber, isNull);
        expect(restored.embeddedLyrics, isNull);
      },
    );
  });
}
