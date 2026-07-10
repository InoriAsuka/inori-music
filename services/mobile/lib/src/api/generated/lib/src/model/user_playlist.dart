//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'user_playlist.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UserPlaylist {
  /// Returns a new [UserPlaylist] instance.
  UserPlaylist({

    required  this.id,

    required  this.userId,

    required  this.name,

     this.description,

    required  this.trackIds,

    required  this.createdAt,

    required  this.updatedAt,
  });

  @JsonKey(
    
    name: r'id',
    required: true,
    includeIfNull: false,
  )


  final String id;



  @JsonKey(
    
    name: r'userId',
    required: true,
    includeIfNull: false,
  )


  final String userId;



  @JsonKey(
    
    name: r'name',
    required: true,
    includeIfNull: false,
  )


  final String name;



  @JsonKey(
    
    name: r'description',
    required: false,
    includeIfNull: false,
  )


  final String? description;



  @JsonKey(
    
    name: r'trackIds',
    required: true,
    includeIfNull: false,
  )


  final List<String> trackIds;



  @JsonKey(
    
    name: r'createdAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime createdAt;



  @JsonKey(
    
    name: r'updatedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime updatedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UserPlaylist &&
      other.id == id &&
      other.userId == userId &&
      other.name == name &&
      other.description == description &&
      other.trackIds == trackIds &&
      other.createdAt == createdAt &&
      other.updatedAt == updatedAt;

    @override
    int get hashCode =>
        id.hashCode +
        userId.hashCode +
        name.hashCode +
        description.hashCode +
        trackIds.hashCode +
        createdAt.hashCode +
        updatedAt.hashCode;

  factory UserPlaylist.fromJson(Map<String, dynamic> json) => _$UserPlaylistFromJson(json);

  Map<String, dynamic> toJson() => _$UserPlaylistToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

