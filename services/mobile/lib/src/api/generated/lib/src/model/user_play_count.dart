//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'user_play_count.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UserPlayCount {
  /// Returns a new [UserPlayCount] instance.
  UserPlayCount({

    required  this.userId,

    required  this.playCount,
  });

  @JsonKey(
    
    name: r'userId',
    required: true,
    includeIfNull: false,
  )


  final String userId;



  @JsonKey(
    
    name: r'playCount',
    required: true,
    includeIfNull: false,
  )


  final int playCount;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UserPlayCount &&
      other.userId == userId &&
      other.playCount == playCount;

    @override
    int get hashCode =>
        userId.hashCode +
        playCount.hashCode;

  factory UserPlayCount.fromJson(Map<String, dynamic> json) => _$UserPlayCountFromJson(json);

  Map<String, dynamic> toJson() => _$UserPlayCountToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

