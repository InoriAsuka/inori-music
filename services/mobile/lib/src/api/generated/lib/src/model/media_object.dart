//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object_verification_result.dart';
import 'package:inori_api/src/model/media_object_lifecycle_change.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObject {
  /// Returns a new [MediaObject] instance.
  MediaObject({

    required  this.assetKind,

    required  this.backendId,

    required  this.contentHash,

    required  this.id,

    required  this.lifecycleState,

    required  this.mimeType,

    required  this.objectKey,

    required  this.sizeBytes,

    required  this.createdAt,

     this.lastLifecycleChange,

     this.lastVerification,

    required  this.updatedAt,
  });

  @JsonKey(
    
    name: r'assetKind',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectAssetKindEnum assetKind;



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


  final MediaObjectLifecycleStateEnum lifecycleState;



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



  @JsonKey(
    
    name: r'createdAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime createdAt;



  @JsonKey(
    
    name: r'lastLifecycleChange',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectLifecycleChange? lastLifecycleChange;



  @JsonKey(
    
    name: r'lastVerification',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectVerificationResult? lastVerification;



  @JsonKey(
    
    name: r'updatedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime updatedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObject &&
      other.assetKind == assetKind &&
      other.backendId == backendId &&
      other.contentHash == contentHash &&
      other.id == id &&
      other.lifecycleState == lifecycleState &&
      other.mimeType == mimeType &&
      other.objectKey == objectKey &&
      other.sizeBytes == sizeBytes &&
      other.createdAt == createdAt &&
      other.lastLifecycleChange == lastLifecycleChange &&
      other.lastVerification == lastVerification &&
      other.updatedAt == updatedAt;

    @override
    int get hashCode =>
        assetKind.hashCode +
        backendId.hashCode +
        contentHash.hashCode +
        id.hashCode +
        lifecycleState.hashCode +
        mimeType.hashCode +
        objectKey.hashCode +
        sizeBytes.hashCode +
        createdAt.hashCode +
        lastLifecycleChange.hashCode +
        lastVerification.hashCode +
        updatedAt.hashCode;

  factory MediaObject.fromJson(Map<String, dynamic> json) => _$MediaObjectFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectAssetKindEnum {
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

const MediaObjectAssetKindEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}


