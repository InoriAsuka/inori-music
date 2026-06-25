//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/catalog_pagination_meta.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_catalog_artists_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminCatalogArtistsGet200Response {
  /// Returns a new [ApiV1AdminCatalogArtistsGet200Response] instance.
  ApiV1AdminCatalogArtistsGet200Response({

     this.artists,

     this.pagination,
  });

  @JsonKey(
    
    name: r'artists',
    required: false,
    includeIfNull: false,
  )


  final List<CatalogArtist>? artists;



  @JsonKey(
    
    name: r'pagination',
    required: false,
    includeIfNull: false,
  )


  final CatalogPaginationMeta? pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminCatalogArtistsGet200Response &&
      other.artists == artists &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        artists.hashCode +
        pagination.hashCode;

  factory ApiV1AdminCatalogArtistsGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1AdminCatalogArtistsGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminCatalogArtistsGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

