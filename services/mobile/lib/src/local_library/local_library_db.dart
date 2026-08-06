import 'dart:async';

import 'package:sqflite/sqflite.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as p;

/// A locally-imported audio file — no server track ID, no account required.
/// Kept in its own table (not [OfflineTrack]'s `offline_tracks`) because
/// every consumer of that table assumes `track_id` is a server catalog UUID
/// usable for re-download/dedup-by-server-id; a local file has no server
/// identity at all and conflating the two would permanently tangle both.
class LocalLibraryTrack {
  const LocalLibraryTrack({
    required this.id,
    required this.title,
    required this.artistName,
    required this.albumTitle,
    required this.localPath,
    this.durationMs,
    this.coverArtPath,
    required this.sizeBytes,
    required this.importedAt,
    this.sampleRate,
    this.bitrate,
    this.fileFormat,
    this.genre,
    this.trackNumber,
    this.embeddedLyrics,
  });

  final String id;
  final String title;
  final String artistName;
  final String albumTitle;
  final String localPath;
  final int? durationMs;
  final String? coverArtPath;
  final int sizeBytes;
  final DateTime importedAt;

  // v5.19.0 — captured from audio_metadata_reader's AudioMetadata at import
  // time (previously parsed then discarded). Null on any track imported
  // before this version; the DB migration that added these columns doesn't
  // retroactively re-scan existing files, so UI must treat null as "unknown"
  // rather than assuming re-import.
  final int? sampleRate;
  final int? bitrate;

  /// File extension without the leading dot, uppercased (e.g. "FLAC") —
  /// derived from the source path at import time, not from tag data.
  final String? fileFormat;
  final String? genre;
  final int? trackNumber;

  /// Raw embedded lyrics text (LRC or plain) from the file's own tags, if
  /// present. Consumed by `localLyricsProvider` (v5.20.0).
  final String? embeddedLyrics;

  Map<String, dynamic> toMap() => {
    'id': id,
    'title': title,
    'artist_name': artistName,
    'album_title': albumTitle,
    'local_path': localPath,
    'duration_ms': durationMs,
    'cover_art_path': coverArtPath,
    'size_bytes': sizeBytes,
    'imported_at': importedAt.millisecondsSinceEpoch,
    'sample_rate': sampleRate,
    'bitrate': bitrate,
    'file_format': fileFormat,
    'genre': genre,
    'track_number': trackNumber,
    'embedded_lyrics': embeddedLyrics,
  };

  static LocalLibraryTrack fromMap(Map<String, dynamic> m) => LocalLibraryTrack(
    id: m['id'] as String,
    title: m['title'] as String,
    artistName: m['artist_name'] as String,
    albumTitle: m['album_title'] as String,
    localPath: m['local_path'] as String,
    durationMs: m['duration_ms'] as int?,
    coverArtPath: m['cover_art_path'] as String?,
    sizeBytes: m['size_bytes'] as int,
    importedAt: DateTime.fromMillisecondsSinceEpoch(m['imported_at'] as int),
    sampleRate: m['sample_rate'] as int?,
    bitrate: m['bitrate'] as int?,
    fileFormat: m['file_format'] as String?,
    genre: m['genre'] as String?,
    trackNumber: m['track_number'] as int?,
    embeddedLyrics: m['embedded_lyrics'] as String?,
  );
}

/// Flat-list sort dimensions for [LocalLibraryDb.queryAll] — the screen
/// itself stays a flat list (see local_library_screen.dart's v1 scope note),
/// this only changes the ordering within it.
enum LocalLibrarySortOrder {
  artistAlbumTitle,
  recentlyAdded,
  titleAZ,
  duration,
}

/// SQL `ORDER BY` clause for [sort] — a top-level pure function (rather than
/// inlined in [LocalLibraryDb.queryAll]) so it's testable without touching
/// sqflite/platform channels.
String localLibraryOrderByClause(LocalLibrarySortOrder sort) => switch (sort) {
  LocalLibrarySortOrder.artistAlbumTitle =>
    'artist_name COLLATE NOCASE, album_title COLLATE NOCASE, title COLLATE NOCASE',
  LocalLibrarySortOrder.recentlyAdded => 'imported_at DESC',
  LocalLibrarySortOrder.titleAZ => 'title COLLATE NOCASE',
  LocalLibrarySortOrder.duration => 'duration_ms DESC',
};

/// LocalLibraryDb singleton helper — mirrors [OfflineDb]'s completer-guarded
/// singleton pattern so concurrent first-callers never race-open the DB twice.
class LocalLibraryDb {
  LocalLibraryDb._();
  static final LocalLibraryDb instance = LocalLibraryDb._();

  Completer<Database>? _initCompleter;

  Future<Database> get db {
    final existing = _initCompleter;
    if (existing != null) return existing.future;

    final completer = Completer<Database>();
    _initCompleter = completer;
    _open().then(
      completer.complete,
      onError: (Object e, StackTrace st) {
        _initCompleter = null;
        completer.completeError(e, st);
      },
    );
    return completer.future;
  }

  static const _createTableSql = '''
    CREATE TABLE IF NOT EXISTS local_library_tracks (
      id TEXT PRIMARY KEY,
      title TEXT NOT NULL,
      artist_name TEXT NOT NULL DEFAULT '',
      album_title TEXT NOT NULL DEFAULT '',
      local_path TEXT NOT NULL,
      duration_ms INTEGER,
      cover_art_path TEXT,
      size_bytes INTEGER NOT NULL DEFAULT 0,
      imported_at INTEGER NOT NULL,
      sample_rate INTEGER,
      bitrate INTEGER,
      file_format TEXT,
      genre TEXT,
      track_number INTEGER,
      embedded_lyrics TEXT
    )
  ''';

  static const _v2Columns = [
    'ALTER TABLE local_library_tracks ADD COLUMN sample_rate INTEGER',
    'ALTER TABLE local_library_tracks ADD COLUMN bitrate INTEGER',
    'ALTER TABLE local_library_tracks ADD COLUMN file_format TEXT',
    'ALTER TABLE local_library_tracks ADD COLUMN genre TEXT',
    'ALTER TABLE local_library_tracks ADD COLUMN track_number INTEGER',
    'ALTER TABLE local_library_tracks ADD COLUMN embedded_lyrics TEXT',
  ];

  Future<Database> _open() async {
    final dir = await getApplicationDocumentsDirectory();
    final path = p.join(dir.path, 'inori_local_library.db');
    return openDatabase(
      path,
      version: 2,
      onCreate: (db, version) => db.execute(_createTableSql),
      onUpgrade: (db, oldVersion, newVersion) async {
        if (oldVersion < 2) {
          for (final stmt in _v2Columns) {
            await db.execute(stmt);
          }
        }
      },
    );
  }

  Future<void> insert(LocalLibraryTrack track) async {
    final d = await db;
    await d.insert(
      'local_library_tracks',
      track.toMap(),
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }

  Future<LocalLibraryTrack?> query(String id) async {
    final d = await db;
    final rows = await d.query(
      'local_library_tracks',
      where: 'id = ?',
      whereArgs: [id],
    );
    return rows.isEmpty ? null : LocalLibraryTrack.fromMap(rows.first);
  }

  /// All imported tracks as a flat list, ordered by [sort].
  Future<List<LocalLibraryTrack>> queryAll({
    LocalLibrarySortOrder sort = LocalLibrarySortOrder.artistAlbumTitle,
  }) async {
    final d = await db;
    final rows = await d.query(
      'local_library_tracks',
      orderBy: localLibraryOrderByClause(sort),
    );
    return rows.map(LocalLibraryTrack.fromMap).toList();
  }

  Future<void> delete(String id) async {
    final d = await db;
    await d.delete('local_library_tracks', where: 'id = ?', whereArgs: [id]);
  }

  Future<void> deleteAll() async {
    final d = await db;
    await d.delete('local_library_tracks');
  }
}
