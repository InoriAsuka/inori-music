//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_pagination_meta.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_catalog_tracks_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminCatalogTracksGet200Response {
  /// Returns a new [ApiV1AdminCatalogTracksGet200Response] instance.
  ApiV1AdminCatalogTracksGet200Response({

     this.tracks,

     this.pagination,
  });

  @JsonKey(
    
    name: r'tracks',
    required: false,
    includeIfNull: false,
  )


  final List<CatalogTrack>? tracks;



  @JsonKey(
    
    name: r'pagination',
    required: false,
    includeIfNull: false,
  )


  final CatalogPaginationMeta? pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminCatalogTracksGet200Response &&
      other.tracks == tracks &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        tracks.hashCode +
        pagination.hashCode;

  factory ApiV1AdminCatalogTracksGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1AdminCatalogTracksGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminCatalogTracksGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

