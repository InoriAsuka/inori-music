//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectRequest {
  /// Returns a new [MediaObjectRequest] instance.
  MediaObjectRequest({

    required  this.assetKind,

    required  this.backendId,

    required  this.contentHash,

    required  this.id,

    required  this.lifecycleState,

    required  this.mimeType,

    required  this.objectKey,

    required  this.sizeBytes,
  });

  @JsonKey(
    
    name: r'assetKind',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectRequestAssetKindEnum assetKind;



  @JsonKey(
    
    name: r'backendId',
    required: true,
    includeIfNull: false,
  )


  final String backendId;



  @JsonKey(
    
    name: r'contentHash',
    required: true,
    includeIfNull: false,
  )


  final String contentHash;



  @JsonKey(
    
    name: r'id',
    required: true,
    includeIfNull: false,
  )


  final String id;



  @JsonKey(
    
    name: r'lifecycleState',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectRequestLifecycleStateEnum lifecycleState;



  @JsonKey(
    
    name: r'mimeType',
    required: true,
    includeIfNull: false,
  )


  final String mimeType;



      /// Relative slash-delimited storage key.
  @JsonKey(
    
    name: r'objectKey',
    required: true,
    includeIfNull: false,
  )


  final String objectKey;



          // minimum: 0
  @JsonKey(
    
    name: r'sizeBytes',
    required: true,
    includeIfNull: false,
  )


  final int sizeBytes;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectRequest &&
      other.assetKind == assetKind &&
      other.backendId == backendId &&
      other.contentHash == contentHash &&
      other.id == id &&
      other.lifecycleState == lifecycleState &&
      other.mimeType == mimeType &&
      other.objectKey == objectKey &&
      other.sizeBytes == sizeBytes;

    @override
    int get hashCode =>
        assetKind.hashCode +
        backendId.hashCode +
        contentHash.hashCode +
        id.hashCode +
        lifecycleState.hashCode +
        mimeType.hashCode +
        objectKey.hashCode +
        sizeBytes.hashCode;

  factory MediaObjectRequest.fromJson(Map<String, dynamic> json) => _$MediaObjectRequestFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectRequestAssetKindEnum {
@JsonValue(r'original_audio')
originalAudio(r'original_audio'),
@JsonValue(r'transcoded_audio')
transcodedAudio(r'transcoded_audio'),
@JsonValue(r'artwork')
artwork(r'artwork'),
@JsonValue(r'lyrics')
lyrics(r'lyrics'),
@JsonValue(r'waveform')
waveform(r'waveform'),
@JsonValue(r'analysis')
analysis(r'analysis'),
@JsonValue(r'import_package')
importPackage(r'import_package'),
@JsonValue(r'backup')
backup(r'backup');

const MediaObjectRequestAssetKindEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectRequestLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectRequestLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}


