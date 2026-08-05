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
  );
}

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
    _open().then(completer.complete, onError: (Object e, StackTrace st) {
      _initCompleter = null;
      completer.completeError(e, st);
    });
    return completer.future;
  }

  Future<Database> _open() async {
    final dir = await getApplicationDocumentsDirectory();
    final path = p.join(dir.path, 'inori_local_library.db');
    return openDatabase(
      path,
      version: 1,
      onCreate: (db, version) => db.execute('''
        CREATE TABLE IF NOT EXISTS local_library_tracks (
          id TEXT PRIMARY KEY,
          title TEXT NOT NULL,
          artist_name TEXT NOT NULL DEFAULT '',
          album_title TEXT NOT NULL DEFAULT '',
          local_path TEXT NOT NULL,
          duration_ms INTEGER,
          cover_art_path TEXT,
          size_bytes INTEGER NOT NULL DEFAULT 0,
          imported_at INTEGER NOT NULL
        )
      '''),
    );
  }

  Future<void> insert(LocalLibraryTrack track) async {
    final d = await db;
    await d.insert('local_library_tracks', track.toMap(),
        conflictAlgorithm: ConflictAlgorithm.replace);
  }

  Future<LocalLibraryTrack?> query(String id) async {
    final d = await db;
    final rows = await d.query('local_library_tracks', where: 'id = ?', whereArgs: [id]);
    return rows.isEmpty ? null : LocalLibraryTrack.fromMap(rows.first);
  }

  /// All imported tracks, sorted for flat-list browsing (artist → album → title).
  Future<List<LocalLibraryTrack>> queryAll() async {
    final d = await db;
    final rows = await d.query(
      'local_library_tracks',
      orderBy: 'artist_name COLLATE NOCASE, album_title COLLATE NOCASE, title COLLATE NOCASE',
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
