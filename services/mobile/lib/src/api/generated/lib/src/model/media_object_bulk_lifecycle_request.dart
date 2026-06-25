//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object_selection_filter.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_bulk_lifecycle_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectBulkLifecycleRequest {
  /// Returns a new [MediaObjectBulkLifecycleRequest] instance.
  MediaObjectBulkLifecycleRequest({

     this.dryRun = false,

    required  this.filter,

    required  this.lifecycleState,
  });

      /// When true, calculate matching objects and per-object outcomes without persisting lifecycle changes.
  @JsonKey(
    defaultValue: false,
    name: r'dryRun',
    required: false,
    includeIfNull: false,
  )


  final bool? dryRun;



  @JsonKey(
    
    name: r'filter',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectSelectionFilter filter;



  @JsonKey(
    
    name: r'lifecycleState',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectBulkLifecycleRequestLifecycleStateEnum lifecycleState;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectBulkLifecycleRequest &&
      other.dryRun == dryRun &&
      other.filter == filter &&
      other.lifecycleState == lifecycleState;

    @override
    int get hashCode =>
        dryRun.hashCode +
        filter.hashCode +
        lifecycleState.hashCode;

  factory MediaObjectBulkLifecycleRequest.fromJson(Map<String, dynamic> json) => _$MediaObjectBulkLifecycleRequestFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectBulkLifecycleRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectBulkLifecycleRequestLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectBulkLifecycleRequestLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}


