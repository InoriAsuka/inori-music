//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/user_playlist.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'list_user_playlists200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ListUserPlaylists200Response {
  /// Returns a new [ListUserPlaylists200Response] instance.
  ListUserPlaylists200Response({

     this.playlists,
  });

  @JsonKey(
    
    name: r'playlists',
    required: false,
    includeIfNull: false,
  )


  final List<UserPlaylist>? playlists;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ListUserPlaylists200Response &&
      other.playlists == playlists;

    @override
    int get hashCode =>
        playlists.hashCode;

  factory ListUserPlaylists200Response.fromJson(Map<String, dynamic> json) => _$ListUserPlaylists200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ListUserPlaylists200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

