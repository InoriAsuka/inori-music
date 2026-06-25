//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_artist_stat_item.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_artist_stats_breakdown.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogArtistStatsBreakdown {
  /// Returns a new [CatalogArtistStatsBreakdown] instance.
  CatalogArtistStatsBreakdown({

    required  this.artists,
  });

      /// Ordered list of per-artist stat rows.
  @JsonKey(
    
    name: r'artists',
    required: true,
    includeIfNull: false,
  )


  final List<CatalogArtistStatItem> artists;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogArtistStatsBreakdown &&
      other.artists == artists;

    @override
    int get hashCode =>
        artists.hashCode;

  factory CatalogArtistStatsBreakdown.fromJson(Map<String, dynamic> json) => _$CatalogArtistStatsBreakdownFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogArtistStatsBreakdownToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

