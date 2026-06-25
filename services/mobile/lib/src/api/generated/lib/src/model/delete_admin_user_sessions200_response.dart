//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'delete_admin_user_sessions200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class DeleteAdminUserSessions200Response {
  /// Returns a new [DeleteAdminUserSessions200Response] instance.
  DeleteAdminUserSessions200Response({

    required  this.revoked,
  });

  @JsonKey(
    
    name: r'revoked',
    required: true,
    includeIfNull: false,
  )


  final int revoked;





    @override
    bool operator ==(Object other) => identical(this, other) || other is DeleteAdminUserSessions200Response &&
      other.revoked == revoked;

    @override
    int get hashCode =>
        revoked.hashCode;

  factory DeleteAdminUserSessions200Response.fromJson(Map<String, dynamic> json) => _$DeleteAdminUserSessions200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$DeleteAdminUserSessions200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

