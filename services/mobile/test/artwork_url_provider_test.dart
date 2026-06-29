// ignore_for_file: implementation_imports
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/api/api_client.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';

// ---------------------------------------------------------------------------
// Stub Dio that returns canned JSON (or throws) for the artwork endpoint.
// ---------------------------------------------------------------------------
Dio _stubDio({required bool hasArtwork, bool throwError = false}) {
  final dio = Dio(BaseOptions(baseUrl: 'http://localhost'));
  dio.interceptors.add(InterceptorsWrapper(
    onRequest: (options, handler) {
      if (throwError) {
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'network error',
            type: DioExceptionType.connectionError,
          ),
        );
        return;
      }
      if (options.path.contains('/artwork')) {
        if (hasArtwork) {
          handler.resolve(
            Response<Map<String, dynamic>>(
              requestOptions: options,
              statusCode: 200,
              data: {'url': 'https://cdn.example.com/cover.jpg', 'expiresIn': 900},
            ),
          );
        } else {
          handler.reject(
            DioException(
              requestOptions: options,
              response: Response(requestOptions: options, statusCode: 404),
              type: DioExceptionType.badResponse,
            ),
          );
        }
      } else {
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unexpected path',
            type: DioExceptionType.badResponse,
          ),
        );
      }
    },
  ));
  return dio;
}

void main() {
  group('artworkUrlProvider', () {
    test('returns URL when server responds 200', () async {
      final container = ProviderContainer(
        overrides: [
          dioProvider.overrideWithValue(_stubDio(hasArtwork: true)),
        ],
      );
      addTearDown(container.dispose);

      final url = await container.read(artworkUrlProvider('album-001').future);
      expect(url, equals('https://cdn.example.com/cover.jpg'));
    });

    test('returns null on 404 (no artwork set)', () async {
      final container = ProviderContainer(
        overrides: [
          dioProvider.overrideWithValue(_stubDio(hasArtwork: false)),
        ],
      );
      addTearDown(container.dispose);

      final url = await container.read(artworkUrlProvider('album-no-art').future);
      expect(url, isNull);
    });

    test('returns null on network error (non-fatal)', () async {
      final container = ProviderContainer(
        overrides: [
          dioProvider.overrideWithValue(_stubDio(hasArtwork: false, throwError: true)),
        ],
      );
      addTearDown(container.dispose);

      final url = await container.read(artworkUrlProvider('album-err').future);
      expect(url, isNull);
    });

    test('returns null for empty albumId without making a request', () async {
      // Empty ID is short-circuited in ArtworkUrlNotifier.build().
      final container = ProviderContainer(
        overrides: [
          dioProvider.overrideWithValue(_stubDio(hasArtwork: true)),
        ],
      );
      addTearDown(container.dispose);

      final url = await container.read(artworkUrlProvider('').future);
      expect(url, isNull);
    });
  });
}
