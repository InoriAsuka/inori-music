//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_lifecycle_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectLifecycleRequest {
  /// Returns a new [MediaObjectLifecycleRequest] instance.
  MediaObjectLifecycleRequest({

    required  this.lifecycleState,
  });

      /// New metadata lifecycle state for the media object. Deleted objects cannot be moved back to a non-deleted state.
  @JsonKey(
    
    name: r'lifecycleState',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectLifecycleRequestLifecycleStateEnum lifecycleState;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectLifecycleRequest &&
      other.lifecycleState == lifecycleState;

    @override
    int get hashCode =>
        lifecycleState.hashCode;

  factory MediaObjectLifecycleRequest.fromJson(Map<String, dynamic> json) => _$MediaObjectLifecycleRequestFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectLifecycleRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

/// New metadata lifecycle state for the media object. Deleted objects cannot be moved back to a non-deleted state.
enum MediaObjectLifecycleRequestLifecycleStateEnum {
    /// New metadata lifecycle state for the media object. Deleted objects cannot be moved back to a non-deleted state.
@JsonValue(r'staged')
staged(r'staged'),
    /// New metadata lifecycle state for the media object. Deleted objects cannot be moved back to a non-deleted state.
@JsonValue(r'active')
active(r'active'),
    /// New metadata lifecycle state for the media object. Deleted objects cannot be moved back to a non-deleted state.
@JsonValue(r'archived')
archived(r'archived'),
    /// New metadata lifecycle state for the media object. Deleted objects cannot be moved back to a non-deleted state.
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectLifecycleRequestLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}


