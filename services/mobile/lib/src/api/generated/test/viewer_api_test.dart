import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for ViewerApi
void main() {
  final instance = InoriApi().getViewerApi();

  group(ViewerApi, () {
    // Stream track audio
    //
    // Proxies audio bytes with HTTP 206 Range support for filesystem-based storage backends. Authenticates via Bearer token in Authorization header or ?token= query parameter.
    //
    //Future apiV1CatalogTracksIdStreamGet(String id, { String token }) async
    test('test apiV1CatalogTracksIdStreamGet', () async {
      // TODO
    });

    // Get viewer track summary
    //
    //Future<MyTrackSummary> getMyTrackSummary(String trackId, { DateTime since, DateTime until, int limit }) async
    test('test getMyTrackSummary', () async {
      // TODO
    });

  });
}
