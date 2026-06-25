//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_album.dart';
import 'package:inori_api/src/model/catalog_pagination_meta.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_catalog_albums_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminCatalogAlbumsGet200Response {
  /// Returns a new [ApiV1AdminCatalogAlbumsGet200Response] instance.
  ApiV1AdminCatalogAlbumsGet200Response({

     this.albums,

     this.pagination,
  });

  @JsonKey(
    
    name: r'albums',
    required: false,
    includeIfNull: false,
  )


  final List<CatalogAlbum>? albums;



  @JsonKey(
    
    name: r'pagination',
    required: false,
    includeIfNull: false,
  )


  final CatalogPaginationMeta? pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminCatalogAlbumsGet200Response &&
      other.albums == albums &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        albums.hashCode +
        pagination.hashCode;

  factory ApiV1AdminCatalogAlbumsGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1AdminCatalogAlbumsGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminCatalogAlbumsGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

