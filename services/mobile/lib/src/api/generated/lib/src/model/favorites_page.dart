//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/admin_list_user_favorites200_response_pagination.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'favorites_page.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class FavoritesPage {
  /// Returns a new [FavoritesPage] instance.
  FavoritesPage({

    required  this.trackIds,

    required  this.pagination,

     this.tracks,
  });

      /// Ordered list of favorite track IDs (newest first)
  @JsonKey(
    
    name: r'trackIds',
    required: true,
    includeIfNull: false,
  )


  final List<String> trackIds;



  @JsonKey(
    
    name: r'pagination',
    required: true,
    includeIfNull: false,
  )


  final AdminListUserFavorites200ResponsePagination pagination;



      /// Full track objects, each with isFavorite=true
  @JsonKey(
    
    name: r'tracks',
    required: false,
    includeIfNull: false,
  )


  final List<CatalogTrack>? tracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is FavoritesPage &&
      other.trackIds == trackIds &&
      other.pagination == pagination &&
      other.tracks == tracks;

    @override
    int get hashCode =>
        trackIds.hashCode +
        pagination.hashCode +
        tracks.hashCode;

  factory FavoritesPage.fromJson(Map<String, dynamic> json) => _$FavoritesPageFromJson(json);

  Map<String, dynamic> toJson() => _$FavoritesPageToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

