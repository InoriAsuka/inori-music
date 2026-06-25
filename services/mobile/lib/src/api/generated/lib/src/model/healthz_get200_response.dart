//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'healthz_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class HealthzGet200Response {
  /// Returns a new [HealthzGet200Response] instance.
  HealthzGet200Response({

    required  this.status,
  });

  @JsonKey(
    
    name: r'status',
    required: true,
    includeIfNull: false,
  )


  final HealthzGet200ResponseStatusEnum status;





    @override
    bool operator ==(Object other) => identical(this, other) || other is HealthzGet200Response &&
      other.status == status;

    @override
    int get hashCode =>
        status.hashCode;

  factory HealthzGet200Response.fromJson(Map<String, dynamic> json) => _$HealthzGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$HealthzGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum HealthzGet200ResponseStatusEnum {
@JsonValue(r'ok')
ok(r'ok');

const HealthzGet200ResponseStatusEnum(this.value);

final String value;

@override
String toString() => value;
}


