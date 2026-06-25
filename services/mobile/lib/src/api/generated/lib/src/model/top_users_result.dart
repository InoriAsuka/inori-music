//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/user_play_count.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'top_users_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class TopUsersResult {
  /// Returns a new [TopUsersResult] instance.
  TopUsersResult({

    required  this.users,
  });

  @JsonKey(
    
    name: r'users',
    required: true,
    includeIfNull: false,
  )


  final List<UserPlayCount> users;





    @override
    bool operator ==(Object other) => identical(this, other) || other is TopUsersResult &&
      other.users == users;

    @override
    int get hashCode =>
        users.hashCode;

  factory TopUsersResult.fromJson(Map<String, dynamic> json) => _$TopUsersResultFromJson(json);

  Map<String, dynamic> toJson() => _$TopUsersResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

