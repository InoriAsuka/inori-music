import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';

// tests for CatalogUpdateTrackRequest
void main() {
  final CatalogUpdateTrackRequest? instance = /* CatalogUpdateTrackRequest(...) */ null;
  // TODO add properties to the entity

  group(CatalogUpdateTrackRequest, () {
    // Track title. Must not be empty if provided.
    // String title
    test('to test the property `title`', () async {
      // TODO
    });

    // Sort key for the track title. May be empty to clear.
    // String sortTitle
    test('to test the property `sortTitle`', () async {
      // TODO
    });

    // ID of the performing artist. Must reference an existing artist.
    // String artistId
    test('to test the property `artistId`', () async {
      // TODO
    });

    // ID of the parent album. Empty string removes album association.
    // String albumId
    test('to test the property `albumId`', () async {
      // TODO
    });

    // int trackNumber
    test('to test the property `trackNumber`', () async {
      // TODO
    });

    // int discNumber
    test('to test the property `discNumber`', () async {
      // TODO
    });

    // Track duration in milliseconds.
    // int durationMs
    test('to test the property `durationMs`', () async {
      // TODO
    });

    // String genre
    test('to test the property `genre`', () async {
      // TODO
    });

  });
}
