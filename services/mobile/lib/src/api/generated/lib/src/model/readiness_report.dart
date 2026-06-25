//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/readiness_check.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'readiness_report.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ReadinessReport {
  /// Returns a new [ReadinessReport] instance.
  ReadinessReport({

    required  this.checks,

    required  this.ready,
  });

  @JsonKey(
    
    name: r'checks',
    required: true,
    includeIfNull: false,
  )


  final List<ReadinessCheck> checks;



  @JsonKey(
    
    name: r'ready',
    required: true,
    includeIfNull: false,
  )


  final bool ready;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ReadinessReport &&
      other.checks == checks &&
      other.ready == ready;

    @override
    int get hashCode =>
        checks.hashCode +
        ready.hashCode;

  factory ReadinessReport.fromJson(Map<String, dynamic> json) => _$ReadinessReportFromJson(json);

  Map<String, dynamic> toJson() => _$ReadinessReportToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

