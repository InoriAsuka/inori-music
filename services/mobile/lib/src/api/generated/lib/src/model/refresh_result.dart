//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/probe_result.dart';
import 'package:inori_api/src/model/capacity_report.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'refresh_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class RefreshResult {
  /// Returns a new [RefreshResult] instance.
  RefreshResult({

    required  this.backendId,

     this.capacity,

     this.errors,

     this.probe,

    required  this.skipped,
  });

  @JsonKey(
    
    name: r'backendId',
    required: true,
    includeIfNull: false,
  )


  final String backendId;



  @JsonKey(
    
    name: r'capacity',
    required: false,
    includeIfNull: false,
  )


  final CapacityReport? capacity;



  @JsonKey(
    
    name: r'errors',
    required: false,
    includeIfNull: false,
  )


  final List<String>? errors;



  @JsonKey(
    
    name: r'probe',
    required: false,
    includeIfNull: false,
  )


  final ProbeResult? probe;



  @JsonKey(
    
    name: r'skipped',
    required: true,
    includeIfNull: false,
  )


  final bool skipped;





    @override
    bool operator ==(Object other) => identical(this, other) || other is RefreshResult &&
      other.backendId == backendId &&
      other.capacity == capacity &&
      other.errors == errors &&
      other.probe == probe &&
      other.skipped == skipped;

    @override
    int get hashCode =>
        backendId.hashCode +
        capacity.hashCode +
        errors.hashCode +
        probe.hashCode +
        skipped.hashCode;

  factory RefreshResult.fromJson(Map<String, dynamic> json) => _$RefreshResultFromJson(json);

  Map<String, dynamic> toJson() => _$RefreshResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

