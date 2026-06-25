import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';

// tests for TrackPlaybackDescriptor
void main() {
  final TrackPlaybackDescriptor? instance = /* TrackPlaybackDescriptor(...) */ null;
  // TODO add properties to the entity

  group(TrackPlaybackDescriptor, () {
    // Catalog track identifier.
    // String trackId
    test('to test the property `trackId`', () async {
      // TODO
    });

    // Media object identifier for the linked audio file.
    // String mediaObjectId
    test('to test the property `mediaObjectId`', () async {
      // TODO
    });

    // MIME type of the audio file, e.g. audio/flac.
    // String mimeType
    test('to test the property `mimeType`', () async {
      // TODO
    });

    // Track duration in milliseconds.
    // int durationMs
    test('to test the property `durationMs`', () async {
      // TODO
    });

    // Storage backend identifier where the object resides.
    // String backendId
    test('to test the property `backendId`', () async {
      // TODO
    });

    // Storage backend type (local, nfs, smb, s3, distributed). Omitted when the backend record cannot be resolved.
    // String backendType
    test('to test the property `backendType`', () async {
      // TODO
    });

    // Object key within the storage backend.
    // String objectKey
    test('to test the property `objectKey`', () async {
      // TODO
    });

    // AWS Signature Version 4 presigned GET URL, valid for 15 minutes. Present only when the backend supports presigned URLs and credentials are configured; omitted otherwise.
    // String presignedUrl
    test('to test the property `presignedUrl`', () async {
      // TODO
    });

    // Server-proxied streaming URL for filesystem backends (local, NFS, SMB). Append ?token=<bearer> to authenticate. Present only when presignedUrl is absent and the backend supports server-side streaming.
    // String streamUrl
    test('to test the property `streamUrl`', () async {
      // TODO
    });

  });
}
