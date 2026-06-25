//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'track_playback_descriptor.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class TrackPlaybackDescriptor {
  /// Returns a new [TrackPlaybackDescriptor] instance.
  TrackPlaybackDescriptor({

    required  this.trackId,

    required  this.mediaObjectId,

    required  this.mimeType,

    required  this.durationMs,

    required  this.backendId,

     this.backendType,

    required  this.objectKey,

     this.presignedUrl,

     this.streamUrl,
  });

      /// Catalog track identifier.
  @JsonKey(
    
    name: r'trackId',
    required: true,
    includeIfNull: false,
  )


  final String trackId;



      /// Media object identifier for the linked audio file.
  @JsonKey(
    
    name: r'mediaObjectId',
    required: true,
    includeIfNull: false,
  )


  final String mediaObjectId;



      /// MIME type of the audio file, e.g. audio/flac.
  @JsonKey(
    
    name: r'mimeType',
    required: true,
    includeIfNull: false,
  )


  final String mimeType;



      /// Track duration in milliseconds.
  @JsonKey(
    
    name: r'durationMs',
    required: true,
    includeIfNull: false,
  )


  final int durationMs;



      /// Storage backend identifier where the object resides.
  @JsonKey(
    
    name: r'backendId',
    required: true,
    includeIfNull: false,
  )


  final String backendId;



      /// Storage backend type (local, nfs, smb, s3, distributed). Omitted when the backend record cannot be resolved.
  @JsonKey(
    
    name: r'backendType',
    required: false,
    includeIfNull: false,
  )


  final String? backendType;



      /// Object key within the storage backend.
  @JsonKey(
    
    name: r'objectKey',
    required: true,
    includeIfNull: false,
  )


  final String objectKey;



      /// AWS Signature Version 4 presigned GET URL, valid for 15 minutes. Present only when the backend supports presigned URLs and credentials are configured; omitted otherwise.
  @JsonKey(
    
    name: r'presignedUrl',
    required: false,
    includeIfNull: false,
  )


  final String? presignedUrl;



      /// Server-proxied streaming URL for filesystem backends (local, NFS, SMB). Append ?token=<bearer> to authenticate. Present only when presignedUrl is absent and the backend supports server-side streaming.
  @JsonKey(
    
    name: r'streamUrl',
    required: false,
    includeIfNull: false,
  )


  final String? streamUrl;





    @override
    bool operator ==(Object other) => identical(this, other) || other is TrackPlaybackDescriptor &&
      other.trackId == trackId &&
      other.mediaObjectId == mediaObjectId &&
      other.mimeType == mimeType &&
      other.durationMs == durationMs &&
      other.backendId == backendId &&
      other.backendType == backendType &&
      other.objectKey == objectKey &&
      other.presignedUrl == presignedUrl &&
      other.streamUrl == streamUrl;

    @override
    int get hashCode =>
        trackId.hashCode +
        mediaObjectId.hashCode +
        mimeType.hashCode +
        durationMs.hashCode +
        backendId.hashCode +
        backendType.hashCode +
        objectKey.hashCode +
        presignedUrl.hashCode +
        streamUrl.hashCode;

  factory TrackPlaybackDescriptor.fromJson(Map<String, dynamic> json) => _$TrackPlaybackDescriptorFromJson(json);

  Map<String, dynamic> toJson() => _$TrackPlaybackDescriptorToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

