//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/api_v1_admin_users_get200_response_pagination.dart';
import 'package:inori_api/src/model/user_view.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_users_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminUsersGet200Response {
  /// Returns a new [ApiV1AdminUsersGet200Response] instance.
  ApiV1AdminUsersGet200Response({

    required  this.users,

    required  this.pagination,
  });

  @JsonKey(
    
    name: r'users',
    required: true,
    includeIfNull: false,
  )


  final List<UserView> users;



  @JsonKey(
    
    name: r'pagination',
    required: true,
    includeIfNull: false,
  )


  final ApiV1AdminUsersGet200ResponsePagination pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminUsersGet200Response &&
      other.users == users &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        users.hashCode +
        pagination.hashCode;

  factory ApiV1AdminUsersGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1AdminUsersGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminUsersGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

