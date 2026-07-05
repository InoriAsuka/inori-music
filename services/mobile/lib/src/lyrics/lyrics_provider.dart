import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/api/api_client.dart';
import 'package:inori_music/src/lyrics/lyric_line.dart';
import 'package:inori_music/src/lyrics/lrc_parser.dart';
import 'package:inori_music/src/lyrics/srt_parser.dart';

/// Fetches and parses lyrics for [trackId].
/// Returns null when the track has no lyrics (404).
/// Throws on other errors.
final lyricsProvider = AsyncNotifierProvider.autoDispose
    .family<LyricsNotifier, List<LyricLine>?, String>(
  LyricsNotifier.new,
);

class LyricsNotifier
    extends AutoDisposeFamilyAsyncNotifier<List<LyricLine>?, String> {
  @override
  Future<List<LyricLine>?> build(String trackId) async {
    if (trackId.isEmpty) return null;
    final dio = ref.watch(dioProvider);
    try {
      final resp = await dio.get<Map<String, dynamic>>(
        '/api/v1/catalog/tracks/$trackId/lyrics',
      );
      final data = resp.data!;
      final format = data['format'] as String? ?? '';
      final content = data['content'] as String? ?? '';
      if (content.isEmpty) return null;
      List<LyricLine>? lines;
      if (format == 'lrc') {
        lines = LrcParser.parse(content);
      } else if (format == 'srt') {
        lines = SrtParser.parse(content);
      } else {
        return null;
      }
      final translation = data['translation'] as String?;
      if (translation != null && translation.isNotEmpty) {
        lines = _mergeTranslation(lines, translation);
      }
      return lines;
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return null;
      rethrow;
    }
  }

  /// Aligns translation text to [lines] by splitting on newlines and
  /// assigning one translation line per lyric line, in order. When the
  /// line counts don't match, the translation is dropped (to avoid
  /// misaligned pairing) and only logged in debug builds.
  List<LyricLine> _mergeTranslation(List<LyricLine> lines, String translation) {
    final translationLines = translation
        .split('\n')
        .map((l) => l.trim())
        .where((l) => l.isNotEmpty)
        .toList();
    if (translationLines.length != lines.length) {
      debugPrint(
        'LyricsNotifier: translation line count (${translationLines.length}) '
        'does not match lyrics line count (${lines.length}); skipping bilingual merge',
      );
      return lines;
    }
    return [
      for (var i = 0; i < lines.length; i++)
        LyricLine(
          timestamp: lines[i].timestamp,
          text: lines[i].text,
          words: lines[i].words,
          translation: translationLines[i],
        ),
    ];
  }
}
