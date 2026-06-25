//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/refresh_result.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'refresh_report.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class RefreshReport {
  /// Returns a new [RefreshReport] instance.
  RefreshReport({

    required  this.completedAt,

    required  this.results,

    required  this.startedAt,
  });

  @JsonKey(
    
    name: r'completedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime completedAt;



  @JsonKey(
    
    name: r'results',
    required: true,
    includeIfNull: false,
  )


  final List<RefreshResult> results;



  @JsonKey(
    
    name: r'startedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime startedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is RefreshReport &&
      other.completedAt == completedAt &&
      other.results == results &&
      other.startedAt == startedAt;

    @override
    int get hashCode =>
        completedAt.hashCode +
        results.hashCode +
        startedAt.hashCode;

  factory RefreshReport.fromJson(Map<String, dynamic> json) => _$RefreshReportFromJson(json);

  Map<String, dynamic> toJson() => _$RefreshReportToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

