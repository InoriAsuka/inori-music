// ignore_for_file: implementation_imports
import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/api/api_client.dart';
import 'package:inori_music/src/shared/absolute_url.dart' show absoluteApiUrl;

// ---------------------------------------------------------------------------
// Artwork URL provider
//   • Family key: albumId (String)
//   • Returns the artwork URL on success, null on 404 / error.
//   • Keeps the resolved URL alive for 300 s so rapid rebuilds don't re-fetch.
// ---------------------------------------------------------------------------

final artworkUrlProvider = AsyncNotifierProvider.autoDispose
    .family<ArtworkUrlNotifier, String?, String>(ArtworkUrlNotifier.new);

class ArtworkUrlNotifier
    extends AutoDisposeFamilyAsyncNotifier<String?, String> {
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
      if (url is! String || url.isEmpty) return null;
      // v5.39.0: for storage backends that can't presign (local/NFS/SMB —
      // most deployments), this is now a *relative* signed path, exactly
      // like track streaming's streamUrl (see absoluteApiUrl's doc comment
      // for why "as-is" broke playback in v5.37.1 and would break this the
      // same way). An S3-style presigned `url` is already absolute and
      // passes through unchanged. resp.requestOptions.baseUrl — rather than
      // reading baseUrlProvider separately — is the base this exact request
      // actually used (set by dioProvider's auth interceptor per-request),
      // so it can never disagree with where the response came from.
      return absoluteApiUrl(url, resp.requestOptions.baseUrl);
    } on DioException catch (e) {
      // 404 = no artwork set — return null instead of throwing.
      if (e.response?.statusCode == 404) return null;
      return null;
    } catch (_) {
      return null;
    }
  }
}
