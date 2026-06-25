//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'patch_user_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class PatchUserRequest {
  /// Returns a new [PatchUserRequest] instance.
  PatchUserRequest({

     this.role,

     this.username,
  });

  @JsonKey(
    
    name: r'role',
    required: false,
    includeIfNull: false,
  )


  final PatchUserRequestRoleEnum? role;



  @JsonKey(
    
    name: r'username',
    required: false,
    includeIfNull: false,
  )


  final String? username;





    @override
    bool operator ==(Object other) => identical(this, other) || other is PatchUserRequest &&
      other.role == role &&
      other.username == username;

    @override
    int get hashCode =>
        role.hashCode +
        username.hashCode;

  factory PatchUserRequest.fromJson(Map<String, dynamic> json) => _$PatchUserRequestFromJson(json);

  Map<String, dynamic> toJson() => _$PatchUserRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum PatchUserRequestRoleEnum {
@JsonValue(r'admin')
admin(r'admin'),
@JsonValue(r'viewer')
viewer(r'viewer');

const PatchUserRequestRoleEnum(this.value);

final String value;

@override
String toString() => value;
}


