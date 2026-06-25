//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_timeline_event.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectTimelineEvent {
  /// Returns a new [MediaObjectTimelineEvent] instance.
  MediaObjectTimelineEvent({

    required  this.at,

     this.lifecycleState,

     this.message,

     this.previousLifecycleState,

     this.source_,

     this.status,

    required  this.type,
  });

  @JsonKey(
    
    name: r'at',
    required: true,
    includeIfNull: false,
  )


  final DateTime at;



  @JsonKey(
    
    name: r'lifecycleState',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectTimelineEventLifecycleStateEnum? lifecycleState;



  @JsonKey(
    
    name: r'message',
    required: false,
    includeIfNull: false,
  )


  final String? message;



  @JsonKey(
    
    name: r'previousLifecycleState',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectTimelineEventPreviousLifecycleStateEnum? previousLifecycleState;



  @JsonKey(
    
    name: r'source',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectTimelineEventSource_Enum? source_;



  @JsonKey(
    
    name: r'status',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectTimelineEventStatusEnum? status;



  @JsonKey(
    
    name: r'type',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectTimelineEventTypeEnum type;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectTimelineEvent &&
      other.at == at &&
      other.lifecycleState == lifecycleState &&
      other.message == message &&
      other.previousLifecycleState == previousLifecycleState &&
      other.source_ == source_ &&
      other.status == status &&
      other.type == type;

    @override
    int get hashCode =>
        at.hashCode +
        lifecycleState.hashCode +
        message.hashCode +
        previousLifecycleState.hashCode +
        source_.hashCode +
        status.hashCode +
        type.hashCode;

  factory MediaObjectTimelineEvent.fromJson(Map<String, dynamic> json) => _$MediaObjectTimelineEventFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectTimelineEventToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectTimelineEventLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectTimelineEventLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectTimelineEventPreviousLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectTimelineEventPreviousLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectTimelineEventSource_Enum {
@JsonValue(r'single')
single(r'single'),
@JsonValue(r'bulk')
bulk(r'bulk');

const MediaObjectTimelineEventSource_Enum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectTimelineEventStatusEnum {
@JsonValue(r'verified')
verified(r'verified'),
@JsonValue(r'failed')
failed(r'failed');

const MediaObjectTimelineEventStatusEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectTimelineEventTypeEnum {
@JsonValue(r'created')
created(r'created'),
@JsonValue(r'lifecycle_changed')
lifecycleChanged(r'lifecycle_changed'),
@JsonValue(r'verification')
verification(r'verification');

const MediaObjectTimelineEventTypeEnum(this.value);

final String value;

@override
String toString() => value;
}


