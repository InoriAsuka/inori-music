//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_artist_stat_item.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogArtistStatItem {
  /// Returns a new [CatalogArtistStatItem] instance.
  CatalogArtistStatItem({

    required  this.artistId,

    required  this.name,

    required  this.albumCount,

    required  this.trackCount,
  });

      /// Artist identifier.
  @JsonKey(
    
    name: r'artistId',
    required: true,
    includeIfNull: false,
  )


  final String artistId;



      /// Artist display name.
  @JsonKey(
    
    name: r'name',
    required: true,
    includeIfNull: false,
  )


  final String name;



      /// Number of albums attributed to this artist.
  @JsonKey(
    
    name: r'albumCount',
    required: true,
    includeIfNull: false,
  )


  final int albumCount;



      /// Number of tracks attributed to this artist.
  @JsonKey(
    
    name: r'trackCount',
    required: true,
    includeIfNull: false,
  )


  final int trackCount;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogArtistStatItem &&
      other.artistId == artistId &&
      other.name == name &&
      other.albumCount == albumCount &&
      other.trackCount == trackCount;

    @override
    int get hashCode =>
        artistId.hashCode +
        name.hashCode +
        albumCount.hashCode +
        trackCount.hashCode;

  factory CatalogArtistStatItem.fromJson(Map<String, dynamic> json) => _$CatalogArtistStatItemFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogArtistStatItemToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

