import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/shared/absolute_url.dart';

// ---------------------------------------------------------------------------
// v5.39.0: absoluteApiUrl is the generalised form of the v5.37.1 playback fix
// (test/playback_url_test.dart) — same helper, same failure mode, now also
// exercised against an *album artwork* signed path rather than a track
// stream one, since getAlbumArtwork's fallback URL has the identical shape
// (a relative path with an HMAC exp/sig query string) for the same reason:
// storage backends that cannot issue presigned URLs (local/NFS/SMB — most
// deployments). Shipping the server half of v5.39.0 without fixing
// artwork_provider.dart the same way v5.37.1 fixed player_notifier.dart would
// trade "no cover" for "broken image" — strictly worse.
// ---------------------------------------------------------------------------

void main() {
  group('absoluteApiUrl', () {
    const base = 'http://10.100.121.134';
    const signedArtwork =
        '/api/v1/catalog/albums/9bd95177cd52ceab/artwork/file'
        '?exp=1786592689&sig=DeALC--dO9le1LdAYCMSH1GAdRp4lwGCkQk019L6WOM';

    test('a relative artwork path gains the configured host', () {
      expect(
        absoluteApiUrl(signedArtwork, base),
        'http://10.100.121.134/api/v1/catalog/albums/9bd95177cd52ceab/artwork/file'
        '?exp=1786592689&sig=DeALC--dO9le1LdAYCMSH1GAdRp4lwGCkQk019L6WOM',
      );
    });

    test('the signature survives — without it the request 401s', () {
      // Neither a native player nor an image loader attaches an
      // Authorization header, so exp+sig is the only credential the request
      // carries. Dropping the query while fixing the host would trade a
      // broken-image icon for a *different* broken-image icon (401 instead
      // of an unreachable relative path) and look like progress.
      final resolved = absoluteApiUrl(signedArtwork, base);
      expect(resolved, contains('exp=1786592689'));
      expect(
        resolved,
        contains('sig=DeALC--dO9le1LdAYCMSH1GAdRp4lwGCkQk019L6WOM'),
      );
    });

    test('an already-absolute presigned URL is left alone', () {
      // The S3-backend path in getAlbumArtwork still returns a fully
      // qualified presigned URL, commonly on a different host than the API.
      // Rewriting it would break the signature it was issued against.
      const presigned =
          'https://minio.example.internal:9000/media/cover.jpg?X-Amz-Signature=abc';
      expect(absoluteApiUrl(presigned, base), presigned);
    });

    test('a base with an explicit port is preserved', () {
      expect(
        absoluteApiUrl(signedArtwork, 'http://10.100.121.134:8080'),
        startsWith('http://10.100.121.134:8080/api/v1/'),
      );
    });

    test('a trailing slash on the base does not double up', () {
      expect(
        absoluteApiUrl(signedArtwork, 'http://10.100.121.134/'),
        startsWith('http://10.100.121.134/api/v1/catalog/'),
      );
      expect(
        absoluteApiUrl(signedArtwork, 'http://10.100.121.134/'),
        isNot(contains('//api')),
      );
    });

    test('the result always has a scheme and a host', () {
      // The single property a bare relative path lacks, and the one an
      // <img>/native player/Image.network widget needs to do anything at all.
      for (final b in const [
        'http://10.100.121.134',
        'http://10.100.121.134/',
        'http://10.100.121.134:8080',
        'https://music.example.com',
      ]) {
        final uri = Uri.parse(absoluteApiUrl(signedArtwork, b));
        expect(uri.hasScheme, isTrue, reason: 'base $b produced $uri');
        expect(uri.host, isNotEmpty, reason: 'base $b produced $uri');
      }
    });
  });
}
