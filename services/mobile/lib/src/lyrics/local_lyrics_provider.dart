import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/local_library/local_library_db.dart';
import 'package:inori_music/src/lyrics/lrc_parser.dart';
import 'package:inori_music/src/lyrics/lyric_line.dart';

/// Reads [LocalLibraryTrack.embeddedLyrics] for a `local:`-prefixed
/// [trackId] and parses it — the local-import counterpart to
/// [lyricsProvider], which fetches from the server instead. Returns null
/// when the track has no embedded lyrics tag at all (nothing to show).
final localLyricsProvider = AsyncNotifierProvider.autoDispose
    .family<LocalLyricsNotifier, List<LyricLine>?, String>(
      LocalLyricsNotifier.new,
    );

class LocalLyricsNotifier
    extends AutoDisposeFamilyAsyncNotifier<List<LyricLine>?, String> {
  @override
  Future<List<LyricLine>?> build(String trackId) async {
    if (trackId.isEmpty) return null;
    final track = await LocalLibraryDb.instance.query(trackId);
    final raw = track?.embeddedLyrics;
    return parseEmbeddedLyrics(raw);
  }
}

/// Parses a [LocalLibraryTrack.embeddedLyrics] value. A top-level pure
/// function (rather than inlined in [LocalLyricsNotifier.build]) so the
/// LRC-vs-plain-text branch is testable without touching sqflite — same
/// rationale as `localLibraryOrderByClause` in local_library_db.dart.
///
/// Returns null when there's no tag content at all. When the content has no
/// recognisable `[mm:ss.xx]` timestamps (most likely a plain-text lyrics
/// tag), falls back to a single untimed block rather than "no lyrics" —
/// Duration.zero is <= any playback position, so it always reads as the
/// "current" line, and there's nothing to progressively highlight without
/// timestamps anyway.
List<LyricLine>? parseEmbeddedLyrics(String? raw) {
  final trimmed = raw?.trim();
  if (trimmed == null || trimmed.isEmpty) return null;
  final timed = LrcParser.parse(trimmed);
  if (timed.isNotEmpty) return timed;
  return [LyricLine(timestamp: Duration.zero, text: trimmed)];
}
