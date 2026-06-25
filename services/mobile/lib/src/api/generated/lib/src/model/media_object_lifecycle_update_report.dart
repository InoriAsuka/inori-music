//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object_lifecycle_update_result.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_lifecycle_update_report.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectLifecycleUpdateReport {
  /// Returns a new [MediaObjectLifecycleUpdateReport] instance.
  MediaObjectLifecycleUpdateReport({

    required  this.dryRun,

    required  this.failedObjects,

    required  this.lifecycleState,

    required  this.matchedObjects,

    required  this.results,

    required  this.updatedAt,

    required  this.updatedObjects,

    required  this.wouldUpdateObjects,
  });

  @JsonKey(
    
    name: r'dryRun',
    required: true,
    includeIfNull: false,
  )


  final bool dryRun;



          // minimum: 0
  @JsonKey(
    
    name: r'failedObjects',
    required: true,
    includeIfNull: false,
  )


  final int failedObjects;



  @JsonKey(
    
    name: r'lifecycleState',
    required: true,
    includeIfNull: false,
  )


  final MediaObjectLifecycleUpdateReportLifecycleStateEnum lifecycleState;



          // minimum: 0
  @JsonKey(
    
    name: r'matchedObjects',
    required: true,
    includeIfNull: false,
  )


  final int matchedObjects;



  @JsonKey(
    
    name: r'results',
    required: true,
    includeIfNull: false,
  )


  final List<MediaObjectLifecycleUpdateResult> results;



  @JsonKey(
    
    name: r'updatedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime updatedAt;



          // minimum: 0
  @JsonKey(
    
    name: r'updatedObjects',
    required: true,
    includeIfNull: false,
  )


  final int updatedObjects;



          // minimum: 0
  @JsonKey(
    
    name: r'wouldUpdateObjects',
    required: true,
    includeIfNull: false,
  )


  final int wouldUpdateObjects;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectLifecycleUpdateReport &&
      other.dryRun == dryRun &&
      other.failedObjects == failedObjects &&
      other.lifecycleState == lifecycleState &&
      other.matchedObjects == matchedObjects &&
      other.results == results &&
      other.updatedAt == updatedAt &&
      other.updatedObjects == updatedObjects &&
      other.wouldUpdateObjects == wouldUpdateObjects;

    @override
    int get hashCode =>
        dryRun.hashCode +
        failedObjects.hashCode +
        lifecycleState.hashCode +
        matchedObjects.hashCode +
        results.hashCode +
        updatedAt.hashCode +
        updatedObjects.hashCode +
        wouldUpdateObjects.hashCode;

  factory MediaObjectLifecycleUpdateReport.fromJson(Map<String, dynamic> json) => _$MediaObjectLifecycleUpdateReportFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectLifecycleUpdateReportToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectLifecycleUpdateReportLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectLifecycleUpdateReportLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}


