//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_stats.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectStats {
  /// Returns a new [MediaObjectStats] instance.
  MediaObjectStats({

     this.backendId,

    required  this.byAssetKind,

    required  this.byBackendId,

    required  this.byLifecycleState,

    required  this.byVerificationStatus,

    required  this.totalObjects,

    required  this.totalSizeBytes,
  });

      /// Backend filter used to calculate this stats response, omitted for all backends.
  @JsonKey(
    
    name: r'backendId',
    required: false,
    includeIfNull: false,
  )


  final String? backendId;



  @JsonKey(
    
    name: r'byAssetKind',
    required: true,
    includeIfNull: false,
  )


  final Map<String, int> byAssetKind;



  @JsonKey(
    
    name: r'byBackendId',
    required: true,
    includeIfNull: false,
  )


  final Map<String, int> byBackendId;



  @JsonKey(
    
    name: r'byLifecycleState',
    required: true,
    includeIfNull: false,
  )


  final Map<String, int> byLifecycleState;



      /// Counts by latest verification state. Includes verified, failed, and unknown buckets.
  @JsonKey(
    
    name: r'byVerificationStatus',
    required: true,
    includeIfNull: false,
  )


  final Map<String, int> byVerificationStatus;



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
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectStats &&
      other.backendId == backendId &&
      other.byAssetKind == byAssetKind &&
      other.byBackendId == byBackendId &&
      other.byLifecycleState == byLifecycleState &&
      other.byVerificationStatus == byVerificationStatus &&
      other.totalObjects == totalObjects &&
      other.totalSizeBytes == totalSizeBytes;

    @override
    int get hashCode =>
        backendId.hashCode +
        byAssetKind.hashCode +
        byBackendId.hashCode +
        byLifecycleState.hashCode +
        byVerificationStatus.hashCode +
        totalObjects.hashCode +
        totalSizeBytes.hashCode;

  factory MediaObjectStats.fromJson(Map<String, dynamic> json) => _$MediaObjectStatsFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectStatsToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

