//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/health_status.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'probe_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ProbeResult {
  /// Returns a new [ProbeResult] instance.
  ProbeResult({

    required  this.backendId,

     this.checkedAt,

     this.message,

    required  this.status,
  });

  @JsonKey(
    
    name: r'backendId',
    required: true,
    includeIfNull: false,
  )


  final String backendId;



  @JsonKey(
    
    name: r'checkedAt',
    required: false,
    includeIfNull: false,
  )


  final DateTime? checkedAt;



  @JsonKey(
    
    name: r'message',
    required: false,
    includeIfNull: false,
  )


  final String? message;



  @JsonKey(
    
    name: r'status',
    required: true,
    includeIfNull: false,
  )


  final HealthStatus status;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ProbeResult &&
      other.backendId == backendId &&
      other.checkedAt == checkedAt &&
      other.message == message &&
      other.status == status;

    @override
    int get hashCode =>
        backendId.hashCode +
        checkedAt.hashCode +
        message.hashCode +
        status.hashCode;

  factory ProbeResult.fromJson(Map<String, dynamic> json) => _$ProbeResultFromJson(json);

  Map<String, dynamic> toJson() => _$ProbeResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

