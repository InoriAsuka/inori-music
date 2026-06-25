//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'readiness_check.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ReadinessCheck {
  /// Returns a new [ReadinessCheck] instance.
  ReadinessCheck({

     this.message,

    required  this.name,

    required  this.status,
  });

  @JsonKey(
    
    name: r'message',
    required: false,
    includeIfNull: false,
  )


  final String? message;



  @JsonKey(
    
    name: r'name',
    required: true,
    includeIfNull: false,
  )


  final String name;



  @JsonKey(
    
    name: r'status',
    required: true,
    includeIfNull: false,
  )


  final ReadinessCheckStatusEnum status;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ReadinessCheck &&
      other.message == message &&
      other.name == name &&
      other.status == status;

    @override
    int get hashCode =>
        message.hashCode +
        name.hashCode +
        status.hashCode;

  factory ReadinessCheck.fromJson(Map<String, dynamic> json) => _$ReadinessCheckFromJson(json);

  Map<String, dynamic> toJson() => _$ReadinessCheckToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum ReadinessCheckStatusEnum {
@JsonValue(r'ok')
ok(r'ok'),
@JsonValue(r'failed')
failed(r'failed');

const ReadinessCheckStatusEnum(this.value);

final String value;

@override
String toString() => value;
}


