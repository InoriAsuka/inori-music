//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'capacity_report.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CapacityReport {
  /// Returns a new [CapacityReport] instance.
  CapacityReport({

    required  this.availableBytes,

    required  this.backendId,

    required  this.checkedAt,

    required  this.totalBytes,

    required  this.usedBytes,
  });

          // minimum: 0
  @JsonKey(
    
    name: r'availableBytes',
    required: true,
    includeIfNull: false,
  )


  final int availableBytes;



  @JsonKey(
    
    name: r'backendId',
    required: true,
    includeIfNull: false,
  )


  final String backendId;



  @JsonKey(
    
    name: r'checkedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime checkedAt;



          // minimum: 0
  @JsonKey(
    
    name: r'totalBytes',
    required: true,
    includeIfNull: false,
  )


  final int totalBytes;



          // minimum: 0
  @JsonKey(
    
    name: r'usedBytes',
    required: true,
    includeIfNull: false,
  )


  final int usedBytes;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CapacityReport &&
      other.availableBytes == availableBytes &&
      other.backendId == backendId &&
      other.checkedAt == checkedAt &&
      other.totalBytes == totalBytes &&
      other.usedBytes == usedBytes;

    @override
    int get hashCode =>
        availableBytes.hashCode +
        backendId.hashCode +
        checkedAt.hashCode +
        totalBytes.hashCode +
        usedBytes.hashCode;

  factory CapacityReport.fromJson(Map<String, dynamic> json) => _$CapacityReportFromJson(json);

  Map<String, dynamic> toJson() => _$CapacityReportToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

