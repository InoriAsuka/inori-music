import 'package:sqflite/sqflite.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as p;

// OfflineTrack model
class OfflineTrack {
  const OfflineTrack({
    required this.trackId,
    required this.title,
    required this.artistName,
    required this.albumTitle,
    this.albumId,
    required this.localPath,
    required this.sizeBytes,
    required this.downloadedAt,
  });

  final String trackId;
  final String title;
  final String artistName;
  final String albumTitle;
  final String? albumId;
  final String localPath;
  final int sizeBytes;
  final DateTime downloadedAt;

  Map<String, dynamic> toMap() => {
    'track_id': trackId,
    'title': title,
    'artist_name': artistName,
    'album_title': albumTitle,
    'album_id': albumId,
    'local_path': localPath,
    'size_bytes': sizeBytes,
    'downloaded_at': downloadedAt.millisecondsSinceEpoch,
  };

  static OfflineTrack fromMap(Map<String, dynamic> m) => OfflineTrack(
    trackId: m['track_id'] as String,
    title: m['title'] as String,
    artistName: m['artist_name'] as String,
    albumTitle: m['album_title'] as String,
    albumId: m['album_id'] as String?,
    localPath: m['local_path'] as String,
    sizeBytes: m['size_bytes'] as int,
    downloadedAt: DateTime.fromMillisecondsSinceEpoch(m['downloaded_at'] as int),
  );
}

// OfflineDb singleton helper
class OfflineDb {
  OfflineDb._();
  static final OfflineDb instance = OfflineDb._();

  Database? _db;

  Future<Database> get db async {
    _db ??= await _open();
    return _db!;
  }

  Future<Database> _open() async {
    final dir = await getApplicationDocumentsDirectory();
    final path = p.join(dir.path, 'inori_offline.db');
    return openDatabase(
      path,
      version: 1,
      onCreate: (db, version) => db.execute('''
        CREATE TABLE IF NOT EXISTS offline_tracks (
          track_id TEXT PRIMARY KEY,
          title TEXT NOT NULL,
          artist_name TEXT NOT NULL DEFAULT '',
          album_title TEXT NOT NULL DEFAULT '',
          album_id TEXT,
          local_path TEXT NOT NULL,
          size_bytes INTEGER NOT NULL DEFAULT 0,
          downloaded_at INTEGER NOT NULL
        )
      '''),
    );
  }

  Future<void> insert(OfflineTrack track) async {
    final d = await db;
    await d.insert('offline_tracks', track.toMap(),
        conflictAlgorithm: ConflictAlgorithm.replace);
  }

  Future<OfflineTrack?> query(String trackId) async {
    final d = await db;
    final rows = await d.query('offline_tracks',
        where: 'track_id = ?', whereArgs: [trackId]);
    return rows.isEmpty ? null : OfflineTrack.fromMap(rows.first);
  }

  Future<List<OfflineTrack>> queryAll() async {
    final d = await db;
    final rows = await d.query('offline_tracks', orderBy: 'downloaded_at DESC');
    return rows.map(OfflineTrack.fromMap).toList();
  }

  Future<void> delete(String trackId) async {
    final d = await db;
    await d.delete('offline_tracks', where: 'track_id = ?', whereArgs: [trackId]);
  }

  Future<void> deleteAll() async {
    final d = await db;
    await d.delete('offline_tracks');
  }
}
