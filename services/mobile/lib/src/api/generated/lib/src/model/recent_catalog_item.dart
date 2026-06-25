//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_album.dart';
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/recent_item_kind.dart';
import 'package:inori_api/src/model/playlist.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'recent_catalog_item.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class RecentCatalogItem {
  /// Returns a new [RecentCatalogItem] instance.
  RecentCatalogItem({

    required  this.kind,

     this.artist,

     this.album,

     this.track,

     this.playlist,

    required  this.addedAt,
  });

  @JsonKey(
    
    name: r'kind',
    required: true,
    includeIfNull: false,
  )


  final RecentItemKind kind;



  @JsonKey(
    
    name: r'artist',
    required: false,
    includeIfNull: false,
  )


  final CatalogArtist? artist;



  @JsonKey(
    
    name: r'album',
    required: false,
    includeIfNull: false,
  )


  final CatalogAlbum? album;



  @JsonKey(
    
    name: r'track',
    required: false,
    includeIfNull: false,
  )


  final CatalogTrack? track;



  @JsonKey(
    
    name: r'playlist',
    required: false,
    includeIfNull: false,
  )


  final Playlist? playlist;



      /// Creation timestamp used for ordering the unified timeline.
  @JsonKey(
    
    name: r'addedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime addedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is RecentCatalogItem &&
      other.kind == kind &&
      other.artist == artist &&
      other.album == album &&
      other.track == track &&
      other.playlist == playlist &&
      other.addedAt == addedAt;

    @override
    int get hashCode =>
        kind.hashCode +
        artist.hashCode +
        album.hashCode +
        track.hashCode +
        playlist.hashCode +
        addedAt.hashCode;

  factory RecentCatalogItem.fromJson(Map<String, dynamic> json) => _$RecentCatalogItemFromJson(json);

  Map<String, dynamic> toJson() => _$RecentCatalogItemToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

