//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_album.dart';
import 'package:inori_api/src/model/search_result_kind.dart';
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'search_result_item.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class SearchResultItem {
  /// Returns a new [SearchResultItem] instance.
  SearchResultItem({

    required  this.kind,

     this.artist,

     this.album,

     this.track,

     this.highlight,
  });

  @JsonKey(
    
    name: r'kind',
    required: true,
    includeIfNull: false,
  )


  final SearchResultKind kind;



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



      /// HTML snippet (with <mark> tags around matched terms) for this item's matched field, when the search backend supports highlighting. Empty when the backend is degraded to PostgreSQL full-text search.
  @JsonKey(
    
    name: r'highlight',
    required: false,
    includeIfNull: false,
  )


  final String? highlight;





    @override
    bool operator ==(Object other) => identical(this, other) || other is SearchResultItem &&
      other.kind == kind &&
      other.artist == artist &&
      other.album == album &&
      other.track == track &&
      other.highlight == highlight;

    @override
    int get hashCode =>
        kind.hashCode +
        artist.hashCode +
        album.hashCode +
        track.hashCode +
        highlight.hashCode;

  factory SearchResultItem.fromJson(Map<String, dynamic> json) => _$SearchResultItemFromJson(json);

  Map<String, dynamic> toJson() => _$SearchResultItemToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

