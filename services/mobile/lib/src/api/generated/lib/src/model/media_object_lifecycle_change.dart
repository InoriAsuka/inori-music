//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_lifecycle_change.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectLifecycleChange {
  /// Returns a new [MediaObjectLifecycleChange] instance.
  MediaObjectLifecycleChange({

    required  this.changedAt,

    required  this.lifecycleState,

    required  this.previousLifecycleState,

    required  this.source_,
  });

  @JsonKey(
    
    name: r'changedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime changedAt;



  @JsonKey(
    
    name: r'lifecycleState',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectLifecycleChangeLifecycleStateEnum lifecycleState;



  @JsonKey(
    
    name: r'previousLifecycleState',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectLifecycleChangePreviousLifecycleStateEnum previousLifecycleState;



      /// Committed lifecycle update source.
  @JsonKey(
    
    name: r'source',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectLifecycleChangeSource_Enum source_;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectLifecycleChange &&
      other.changedAt == changedAt &&
      other.lifecycleState == lifecycleState &&
      other.previousLifecycleState == previousLifecycleState &&
      other.source_ == source_;

    @override
    int get hashCode =>
        changedAt.hashCode +
        lifecycleState.hashCode +
        previousLifecycleState.hashCode +
        source_.hashCode;

  factory MediaObjectLifecycleChange.fromJson(Map<String, dynamic> json) => _$MediaObjectLifecycleChangeFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectLifecycleChangeToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectLifecycleChangeLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectLifecycleChangeLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectLifecycleChangePreviousLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectLifecycleChangePreviousLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}


/// Committed lifecycle update source.
enum MediaObjectLifecycleChangeSource_Enum {
    /// Committed lifecycle update source.
@JsonValue(r'single')
single(r'single'),
    /// Committed lifecycle update source.
@JsonValue(r'bulk')
bulk(r'bulk');

const MediaObjectLifecycleChangeSource_Enum(this.value);

final String value;

@override
String toString() => value;
}


