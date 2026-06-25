//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object_duplicate_group.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_duplicate_report.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectDuplicateReport {
  /// Returns a new [MediaObjectDuplicateReport] instance.
  MediaObjectDuplicateReport({

     this.backendId,

    required  this.groups,

    required  this.minCopies,

    required  this.totalGroups,

    required  this.totalObjects,

    required  this.totalSizeBytes,
  });

      /// Backend filter used to calculate this duplicate report, omitted for all backends.
  @JsonKey(
    
    name: r'backendId',
    required: false,
    includeIfNull: false,
  )


  final String? backendId;



  @JsonKey(
    
    name: r'groups',
    required: true,
    includeIfNull: false,
  )


  final List<MediaObjectDuplicateGroup> groups;



          // minimum: 2
  @JsonKey(
    
    name: r'minCopies',
    required: true,
    includeIfNull: false,
  )


  final int minCopies;



          // minimum: 0
  @JsonKey(
    
    name: r'totalGroups',
    required: true,
    includeIfNull: false,
  )


  final int totalGroups;



          // minimum: 0
  @JsonKey(
    
    name: r'totalObjects',
    required: true,
    includeIfNull: false,
  )


  final int totalObjects;



          // minimum: 0
  @JsonKey(
    
    name: r'totalSizeBytes',
    required: true,
    includeIfNull: false,
  )


  final int totalSizeBytes;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectDuplicateReport &&
      other.backendId == backendId &&
      other.groups == groups &&
      other.minCopies == minCopies &&
      other.totalGroups == totalGroups &&
      other.totalObjects == totalObjects &&
      other.totalSizeBytes == totalSizeBytes;

    @override
    int get hashCode =>
        backendId.hashCode +
        groups.hashCode +
        minCopies.hashCode +
        totalGroups.hashCode +
        totalObjects.hashCode +
        totalSizeBytes.hashCode;

  factory MediaObjectDuplicateReport.fromJson(Map<String, dynamic> json) => _$MediaObjectDuplicateReportFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectDuplicateReportToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

