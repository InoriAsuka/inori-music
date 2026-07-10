//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'get_user_playlist_tracks200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class GetUserPlaylistTracks200Response {
  /// Returns a new [GetUserPlaylistTracks200Response] instance.
  GetUserPlaylistTracks200Response({

     this.tracks,
  });

  @JsonKey(
    
    name: r'tracks',
    required: false,
    includeIfNull: false,
  )


  final List<CatalogTrack>? tracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is GetUserPlaylistTracks200Response &&
      other.tracks == tracks;

    @override
    int get hashCode =>
        tracks.hashCode;

  factory GetUserPlaylistTracks200Response.fromJson(Map<String, dynamic> json) => _$GetUserPlaylistTracks200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$GetUserPlaylistTracks200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

