//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_lifecycle_update_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectLifecycleUpdateResult {
  /// Returns a new [MediaObjectLifecycleUpdateResult] instance.
  MediaObjectLifecycleUpdateResult({

    required  this.lifecycleState,

    required  this.mediaObjectId,

     this.message,

     this.object,

    required  this.previousLifecycleState,

    required  this.status,
  });

  @JsonKey(
    
    name: r'lifecycleState',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectLifecycleUpdateResultLifecycleStateEnum lifecycleState;



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
    
    name: r'object',
    required: false,
    includeIfNull: false,
  )


  final MediaObject? object;



  @JsonKey(
    
    name: r'previousLifecycleState',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectLifecycleUpdateResultPreviousLifecycleStateEnum previousLifecycleState;



  @JsonKey(
    
    name: r'status',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectLifecycleUpdateResultStatusEnum status;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectLifecycleUpdateResult &&
      other.lifecycleState == lifecycleState &&
      other.mediaObjectId == mediaObjectId &&
      other.message == message &&
      other.object == object &&
      other.previousLifecycleState == previousLifecycleState &&
      other.status == status;

    @override
    int get hashCode =>
        lifecycleState.hashCode +
        mediaObjectId.hashCode +
        message.hashCode +
        object.hashCode +
        previousLifecycleState.hashCode +
        status.hashCode;

  factory MediaObjectLifecycleUpdateResult.fromJson(Map<String, dynamic> json) => _$MediaObjectLifecycleUpdateResultFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectLifecycleUpdateResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectLifecycleUpdateResultLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectLifecycleUpdateResultLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectLifecycleUpdateResultPreviousLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectLifecycleUpdateResultPreviousLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectLifecycleUpdateResultStatusEnum {
@JsonValue(r'updated')
updated(r'updated'),
@JsonValue(r'would_update')
wouldUpdate(r'would_update'),
@JsonValue(r'failed')
failed(r'failed');

const MediaObjectLifecycleUpdateResultStatusEnum(this.value);

final String value;

@override
String toString() => value;
}


