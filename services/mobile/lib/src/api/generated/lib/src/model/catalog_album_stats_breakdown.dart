//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_album_stat_item.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_album_stats_breakdown.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogAlbumStatsBreakdown {
  /// Returns a new [CatalogAlbumStatsBreakdown] instance.
  CatalogAlbumStatsBreakdown({

    required  this.albums,
  });

      /// Ordered list of per-album stat rows.
  @JsonKey(
    
    name: r'albums',
    required: true,
    includeIfNull: false,
  )


  final List<CatalogAlbumStatItem> albums;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogAlbumStatsBreakdown &&
      other.albums == albums;

    @override
    int get hashCode =>
        albums.hashCode;

  factory CatalogAlbumStatsBreakdown.fromJson(Map<String, dynamic> json) => _$CatalogAlbumStatsBreakdownFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogAlbumStatsBreakdownToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

