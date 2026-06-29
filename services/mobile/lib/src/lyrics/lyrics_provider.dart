import 'package:dio/dio.dart';
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
      if (format == 'lrc') return LrcParser.parse(content);
      if (format == 'srt') return SrtParser.parse(content);
      return null;
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return null;
      rethrow;
    }
  }
}
