//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_album_stat_item.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogAlbumStatItem {
  /// Returns a new [CatalogAlbumStatItem] instance.
  CatalogAlbumStatItem({

    required  this.albumId,

    required  this.title,

    required  this.artistId,

    required  this.trackCount,
  });

      /// Album identifier.
  @JsonKey(
    
    name: r'albumId',
    required: true,
    includeIfNull: false,
  )


  final String albumId;



      /// Album title.
  @JsonKey(
    
    name: r'title',
    required: true,
    includeIfNull: false,
  )


  final String title;



      /// Identifier of the artist who owns this album.
  @JsonKey(
    
    name: r'artistId',
    required: true,
    includeIfNull: false,
  )


  final String artistId;



      /// Number of tracks belonging to this album.
  @JsonKey(
    
    name: r'trackCount',
    required: true,
    includeIfNull: false,
  )


  final int trackCount;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogAlbumStatItem &&
      other.albumId == albumId &&
      other.title == title &&
      other.artistId == artistId &&
      other.trackCount == trackCount;

    @override
    int get hashCode =>
        albumId.hashCode +
        title.hashCode +
        artistId.hashCode +
        trackCount.hashCode;

  factory CatalogAlbumStatItem.fromJson(Map<String, dynamic> json) => _$CatalogAlbumStatItemFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogAlbumStatItemToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

