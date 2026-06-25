//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object_verification_result.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_verification_report.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectVerificationReport {
  /// Returns a new [MediaObjectVerificationReport] instance.
  MediaObjectVerificationReport({

    required  this.checkedAt,

    required  this.results,
  });

  @JsonKey(
    
    name: r'checkedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime checkedAt;



  @JsonKey(
    
    name: r'results',
    required: true,
    includeIfNull: false,
  )


  final List<MediaObjectVerificationResult> results;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectVerificationReport &&
      other.checkedAt == checkedAt &&
      other.results == results;

    @override
    int get hashCode =>
        checkedAt.hashCode +
        results.hashCode;

  factory MediaObjectVerificationReport.fromJson(Map<String, dynamic> json) => _$MediaObjectVerificationReportFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectVerificationReportToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

