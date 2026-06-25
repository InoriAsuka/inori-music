//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'force_change_password_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ForceChangePasswordRequest {
  /// Returns a new [ForceChangePasswordRequest] instance.
  ForceChangePasswordRequest({

    required  this.newPassword,
  });

  @JsonKey(
    
    name: r'newPassword',
    required: true,
    includeIfNull: false,
  )


  final String newPassword;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ForceChangePasswordRequest &&
      other.newPassword == newPassword;

    @override
    int get hashCode =>
        newPassword.hashCode;

  factory ForceChangePasswordRequest.fromJson(Map<String, dynamic> json) => _$ForceChangePasswordRequestFromJson(json);

  Map<String, dynamic> toJson() => _$ForceChangePasswordRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

