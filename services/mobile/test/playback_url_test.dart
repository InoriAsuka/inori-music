import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/player/player_notifier.dart';

// ---------------------------------------------------------------------------
// Guard for the v5.37.1 "Playback failed: (-1002) unsupported URL" failure.
//
// The playback descriptor endpoint returns streamUrl as a *relative* path:
//
//   /api/v1/catalog/tracks/<id>/stream?exp=1786592689&sig=DeALC--dO9le1Ld…
//
// which is correct for the same-origin web client and unusable for a native
// player. resolvePlaybackUrl returned it verbatim — the branch sat *above*
// the fallback that does prepend the base URL, so the correct concatenation
// was unreachable — and AVPlayer answered with NSURLErrorUnsupportedURL.
//
// Nothing about the string looks wrong at a glance, which is the point: it
// carries a valid HMAC signature and a valid path. It just has no host.
// ---------------------------------------------------------------------------

void main() {
  group('absolutePlaybackUrl', () {
    const base = 'http://10.100.121.134';
    const signed =
        '/api/v1/catalog/tracks/4977b05d33167cff/stream'
        '?exp=1786592689&sig=DeALC--dO9le1LdAYCMSH1GAdRp4lwGCkQk019L6WOM';

    test('a relative stream path gains the configured host', () {
      expect(
        absolutePlaybackUrl(signed, base),
        'http://10.100.121.134/api/v1/catalog/tracks/4977b05d33167cff/stream'
        '?exp=1786592689&sig=DeALC--dO9le1LdAYCMSH1GAdRp4lwGCkQk019L6WOM',
      );
    });

    test('the signature survives — without it the stream 401s', () {
      // The engine attaches no Authorization header, so exp+sig is the only
      // credential the request carries. Dropping the query while fixing the
      // host would trade -1002 for a 401 and look like a different bug.
      final resolved = absolutePlaybackUrl(signed, base);
      expect(resolved, contains('exp=1786592689'));
      expect(
        resolved,
        contains('sig=DeALC--dO9le1LdAYCMSH1GAdRp4lwGCkQk019L6WOM'),
      );
    });

    test('an already-absolute presigned URL is left alone', () {
      // Object storage hands back a fully-qualified URL, commonly on a
      // different host than the API. Rewriting it would break the signature
      // it was issued against.
      const presigned =
          'https://minio.example.internal:9000/media/obj.flac?X-Amz-Signature=abc';
      expect(absolutePlaybackUrl(presigned, base), presigned);
    });

    test('a base with an explicit port is preserved', () {
      expect(
        absolutePlaybackUrl(signed, 'http://10.100.121.134:8080'),
        startsWith('http://10.100.121.134:8080/api/v1/'),
      );
    });

    test('a trailing slash on the base does not double up', () {
      expect(
        absolutePlaybackUrl(signed, 'http://10.100.121.134/'),
        startsWith('http://10.100.121.134/api/v1/catalog/'),
      );
      expect(
        absolutePlaybackUrl(signed, 'http://10.100.121.134/'),
        isNot(contains('//api')),
      );
    });

    test('the result always has a scheme and a host', () {
      // The single property that -1002 was reporting the absence of.
      for (final b in const [
        'http://10.100.121.134',
        'http://10.100.121.134/',
        'http://10.100.121.134:8080',
        'https://music.example.com',
      ]) {
        final uri = Uri.parse(absolutePlaybackUrl(signed, b));
        expect(uri.hasScheme, isTrue, reason: 'base $b produced $uri');
        expect(uri.host, isNotEmpty, reason: 'base $b produced $uri');
      }
    });
  });
}
