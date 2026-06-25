//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_playlist_stat_item.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogPlaylistStatItem {
  /// Returns a new [CatalogPlaylistStatItem] instance.
  CatalogPlaylistStatItem({

    required  this.playlistId,

    required  this.name,

    required  this.trackCount,
  });

      /// Playlist identifier.
  @JsonKey(
    
    name: r'playlistId',
    required: true,
    includeIfNull: false,
  )


  final String playlistId;



      /// Playlist display name.
  @JsonKey(
    
    name: r'name',
    required: true,
    includeIfNull: false,
  )


  final String name;



      /// Number of track entries in this playlist (duplicates counted separately).
  @JsonKey(
    
    name: r'trackCount',
    required: true,
    includeIfNull: false,
  )


  final int trackCount;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogPlaylistStatItem &&
      other.playlistId == playlistId &&
      other.name == name &&
      other.trackCount == trackCount;

    @override
    int get hashCode =>
        playlistId.hashCode +
        name.hashCode +
        trackCount.hashCode;

  factory CatalogPlaylistStatItem.fromJson(Map<String, dynamic> json) => _$CatalogPlaylistStatItemFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogPlaylistStatItemToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

