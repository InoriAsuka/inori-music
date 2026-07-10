//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'update_user_playlist_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UpdateUserPlaylistRequest {
  /// Returns a new [UpdateUserPlaylistRequest] instance.
  UpdateUserPlaylistRequest({

     this.name,

     this.description,
  });

  @JsonKey(
    
    name: r'name',
    required: false,
    includeIfNull: false,
  )


  final String? name;



  @JsonKey(
    
    name: r'description',
    required: false,
    includeIfNull: false,
  )


  final String? description;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UpdateUserPlaylistRequest &&
      other.name == name &&
      other.description == description;

    @override
    int get hashCode =>
        name.hashCode +
        description.hashCode;

  factory UpdateUserPlaylistRequest.fromJson(Map<String, dynamic> json) => _$UpdateUserPlaylistRequestFromJson(json);

  Map<String, dynamic> toJson() => _$UpdateUserPlaylistRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

