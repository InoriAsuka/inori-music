//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_verification_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectVerificationResult {
  /// Returns a new [MediaObjectVerificationResult] instance.
  MediaObjectVerificationResult({

    required  this.backendId,

    required  this.contentHash,

    required  this.mediaObjectId,

     this.message,

    required  this.objectKey,

    required  this.sizeBytes,

    required  this.status,

    required  this.verifiedAt,
  });

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
    
    name: r'mediaObjectId',
    required: true,
    includeIfNull: false,
  )


  final String mediaObjectId;



  @JsonKey(
    
    name: r'message',
    required: false,
    includeIfNull: false,
  )


  final String? message;



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
    
    name: r'status',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectVerificationResultStatusEnum status;



  @JsonKey(
    
    name: r'verifiedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime verifiedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectVerificationResult &&
      other.backendId == backendId &&
      other.contentHash == contentHash &&
      other.mediaObjectId == mediaObjectId &&
      other.message == message &&
      other.objectKey == objectKey &&
      other.sizeBytes == sizeBytes &&
      other.status == status &&
      other.verifiedAt == verifiedAt;

    @override
    int get hashCode =>
        backendId.hashCode +
        contentHash.hashCode +
        mediaObjectId.hashCode +
        message.hashCode +
        objectKey.hashCode +
        sizeBytes.hashCode +
        status.hashCode +
        verifiedAt.hashCode;

  factory MediaObjectVerificationResult.fromJson(Map<String, dynamic> json) => _$MediaObjectVerificationResultFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectVerificationResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectVerificationResultStatusEnum {
@JsonValue(r'verified')
verified(r'verified'),
@JsonValue(r'failed')
failed(r'failed');

const MediaObjectVerificationResultStatusEnum(this.value);

final String value;

@override
String toString() => value;
}


