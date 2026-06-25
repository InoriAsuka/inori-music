//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_playlist_stat_item.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_playlist_stats_breakdown.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogPlaylistStatsBreakdown {
  /// Returns a new [CatalogPlaylistStatsBreakdown] instance.
  CatalogPlaylistStatsBreakdown({

    required  this.playlists,
  });

      /// Ordered list of per-playlist stat rows.
  @JsonKey(
    
    name: r'playlists',
    required: true,
    includeIfNull: false,
  )


  final List<CatalogPlaylistStatItem> playlists;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogPlaylistStatsBreakdown &&
      other.playlists == playlists;

    @override
    int get hashCode =>
        playlists.hashCode;

  factory CatalogPlaylistStatsBreakdown.fromJson(Map<String, dynamic> json) => _$CatalogPlaylistStatsBreakdownFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogPlaylistStatsBreakdownToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

