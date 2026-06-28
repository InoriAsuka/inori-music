// ignore_for_file: implementation_imports
import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/api/api_client.dart';

// ---------------------------------------------------------------------------
// Artwork URL provider
//   • Family key: albumId (String)
//   • Returns the presigned artwork URL on success, null on 404 / error.
//   • Keeps the resolved URL alive for 300 s so rapid rebuilds don't re-fetch.
// ---------------------------------------------------------------------------

final artworkUrlProvider = AsyncNotifierProvider.autoDispose
    .family<ArtworkUrlNotifier, String?, String>(ArtworkUrlNotifier.new);

class ArtworkUrlNotifier extends AutoDisposeFamilyAsyncNotifier<String?, String> {
  @override
  Future<String?> build(String albumId) async {
    if (albumId.isEmpty) return null;

    // Keep alive for 300 s even when no widgets are watching.
    final link = ref.keepAlive();
    final timer = Timer(const Duration(seconds: 300), link.close);
    ref.onDispose(timer.cancel);

    final dio = ref.read(dioProvider);
    try {
      final resp = await dio.get<Map<String, dynamic>>(
        '/api/v1/catalog/albums/$albumId/artwork',
      );
      final url = resp.data?['url'];
      if (url is String && url.isNotEmpty) return url;
      return null;
    } on DioException catch (e) {
      // 404 = no artwork set — return null instead of throwing.
      if (e.response?.statusCode == 404) return null;
      return null;
    } catch (_) {
      return null;
    }
  }
}
