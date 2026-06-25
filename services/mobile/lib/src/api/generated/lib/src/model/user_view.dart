//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/user_role.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'user_view.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UserView {
  /// Returns a new [UserView] instance.
  UserView({

    required  this.createdAt,

    required  this.enabled,

    required  this.id,

    required  this.role,

    required  this.updatedAt,

    required  this.username,
  });

  @JsonKey(
    
    name: r'createdAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime createdAt;



  @JsonKey(
    
    name: r'enabled',
    required: true,
    includeIfNull: false,
  )


  final bool enabled;



  @JsonKey(
    
    name: r'id',
    required: true,
    includeIfNull: false,
  )


  final String id;



  @JsonKey(
    
    name: r'role',
    required: true,
    includeIfNull: false,
  )


  final UserRole role;



  @JsonKey(
    
    name: r'updatedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime updatedAt;



  @JsonKey(
    
    name: r'username',
    required: true,
    includeIfNull: false,
  )


  final String username;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UserView &&
      other.createdAt == createdAt &&
      other.enabled == enabled &&
      other.id == id &&
      other.role == role &&
      other.updatedAt == updatedAt &&
      other.username == username;

    @override
    int get hashCode =>
        createdAt.hashCode +
        enabled.hashCode +
        id.hashCode +
        role.hashCode +
        updatedAt.hashCode +
        username.hashCode;

  factory UserView.fromJson(Map<String, dynamic> json) => _$UserViewFromJson(json);

  Map<String, dynamic> toJson() => _$UserViewToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

