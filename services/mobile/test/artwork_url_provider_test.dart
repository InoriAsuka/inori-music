// ignore_for_file: implementation_imports
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/api/api_client.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';

// ---------------------------------------------------------------------------
// Stub Dio that returns canned JSON (or throws) for the artwork endpoint.
//
// [artworkUrl] defaults to an already-absolute (S3-presigned-shaped) URL so
// the pre-existing "happy path" tests below stay exactly as they were. The
// v5.39.0-specific test overrides it with a *relative* signed path — the
// shape getAlbumArtwork now returns for local/NFS/SMB backends (see
// requirement.md v5.39.0) — to prove the provider absolutises it instead of
// handing it to an Image widget verbatim.
// ---------------------------------------------------------------------------
Dio _stubDio({
  required bool hasArtwork,
  bool throwError = false,
  String artworkUrl = 'https://cdn.example.com/cover.jpg',
  String baseUrl = 'http://localhost',
}) {
  final dio = Dio(BaseOptions(baseUrl: baseUrl));
  dio.interceptors.add(
    InterceptorsWrapper(
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
                data: {'url': artworkUrl, 'expiresIn': 900},
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
    ),
  );
  return dio;
}

void main() {
  group('artworkUrlProvider', () {
    test('returns URL when server responds 200', () async {
      final container = ProviderContainer(
        overrides: [dioProvider.overrideWithValue(_stubDio(hasArtwork: true))],
      );
      addTearDown(container.dispose);

      final url = await container.read(artworkUrlProvider('album-001').future);
      expect(url, equals('https://cdn.example.com/cover.jpg'));
    });

    test('returns null on 404 (no artwork set)', () async {
      final container = ProviderContainer(
        overrides: [dioProvider.overrideWithValue(_stubDio(hasArtwork: false))],
      );
      addTearDown(container.dispose);

      final url = await container.read(
        artworkUrlProvider('album-no-art').future,
      );
      expect(url, isNull);
    });

    test('returns null on network error (non-fatal)', () async {
      final container = ProviderContainer(
        overrides: [
          dioProvider.overrideWithValue(
            _stubDio(hasArtwork: false, throwError: true),
          ),
        ],
      );
      addTearDown(container.dispose);

      final url = await container.read(artworkUrlProvider('album-err').future);
      expect(url, isNull);
    });

    test('returns null for empty albumId without making a request', () async {
      // Empty ID is short-circuited in ArtworkUrlNotifier.build().
      final container = ProviderContainer(
        overrides: [dioProvider.overrideWithValue(_stubDio(hasArtwork: true))],
      );
      addTearDown(container.dispose);

      final url = await container.read(artworkUrlProvider('').future);
      expect(url, isNull);
    });

    // v5.39.0 guard: getAlbumArtwork now returns a *relative* signed path for
    // storage backends that cannot presign (local/NFS/SMB — most real
    // deployments; see requirement.md v5.39.0). Before this fix,
    // ArtworkUrlNotifier.build() returned resp.data?['url'] verbatim — fine
    // for the S3-presigned-shaped absolute URL the earlier tests use, but a
    // relative path handed straight to an Image widget fails exactly like
    // streamUrl did for playback in v5.37.1 (no scheme, no host). This test
    // fails against the pre-fix code with url == the raw relative string.
    test('a relative signed artwork path is made absolute', () async {
      const relative =
          '/api/v1/catalog/albums/album-relative/artwork/file'
          '?exp=1786592689&sig=DeALC--dO9le1LdAYCMSH1GAdRp4lwGCkQk019L6WOM';
      final container = ProviderContainer(
        overrides: [
          dioProvider.overrideWithValue(
            _stubDio(
              hasArtwork: true,
              artworkUrl: relative,
              baseUrl: 'http://10.100.121.134',
            ),
          ),
        ],
      );
      addTearDown(container.dispose);

      final url = await container.read(
        artworkUrlProvider('album-relative').future,
      );
      expect(url, 'http://10.100.121.134$relative');
      final uri = Uri.parse(url!);
      expect(uri.hasScheme, isTrue);
      expect(uri.host, isNotEmpty);
      // The signature must survive the rewrite — it is the only credential
      // this request carries.
      expect(url, contains('exp=1786592689'));
      expect(url, contains('sig=DeALC--dO9le1LdAYCMSH1GAdRp4lwGCkQk019L6WOM'));
    });
  });
}
