//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_catalog_playlists_id_tracks_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminCatalogPlaylistsIdTracksGet200Response {
  /// Returns a new [ApiV1AdminCatalogPlaylistsIdTracksGet200Response] instance.
  ApiV1AdminCatalogPlaylistsIdTracksGet200Response({
  });



    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminCatalogPlaylistsIdTracksGet200Response &&

    @override
    int get hashCode =>

  factory ApiV1AdminCatalogPlaylistsIdTracksGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1AdminCatalogPlaylistsIdTracksGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminCatalogPlaylistsIdTracksGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

