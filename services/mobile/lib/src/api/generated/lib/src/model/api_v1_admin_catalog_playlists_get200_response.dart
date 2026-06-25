//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/playlist.dart';
import 'package:inori_api/src/model/catalog_pagination_meta.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_catalog_playlists_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminCatalogPlaylistsGet200Response {
  /// Returns a new [ApiV1AdminCatalogPlaylistsGet200Response] instance.
  ApiV1AdminCatalogPlaylistsGet200Response({

     this.playlists,

     this.pagination,
  });

  @JsonKey(
    
    name: r'playlists',
    required: false,
    includeIfNull: false,
  )


  final List<Playlist>? playlists;



  @JsonKey(
    
    name: r'pagination',
    required: false,
    includeIfNull: false,
  )


  final CatalogPaginationMeta? pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminCatalogPlaylistsGet200Response &&
      other.playlists == playlists &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        playlists.hashCode +
        pagination.hashCode;

  factory ApiV1AdminCatalogPlaylistsGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1AdminCatalogPlaylistsGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminCatalogPlaylistsGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

