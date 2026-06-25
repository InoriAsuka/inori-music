import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';

// tests for CatalogUpdateAlbumRequest
void main() {
  final CatalogUpdateAlbumRequest? instance = /* CatalogUpdateAlbumRequest(...) */ null;
  // TODO add properties to the entity

  group(CatalogUpdateAlbumRequest, () {
    // Album title. Must not be empty if provided.
    // String title
    test('to test the property `title`', () async {
      // TODO
    });

    // Sort key for the album title. May be empty to clear.
    // String sortTitle
    test('to test the property `sortTitle`', () async {
      // TODO
    });

    // ID of the owning artist. Must reference an existing artist.
    // String artistId
    test('to test the property `artistId`', () async {
      // TODO
    });

    // Release year. 0 clears the value.
    // int releaseYear
    test('to test the property `releaseYear`', () async {
      // TODO
    });

  });
}
