//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_stats.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogStats {
  /// Returns a new [CatalogStats] instance.
  CatalogStats({

    required  this.artists,

    required  this.albums,

    required  this.tracks,

    required  this.playlists,
  });

      /// Total number of artists.
  @JsonKey(
    
    name: r'artists',
    required: true,
    includeIfNull: false,
  )


  final int artists;



      /// Total number of albums.
  @JsonKey(
    
    name: r'albums',
    required: true,
    includeIfNull: false,
  )


  final int albums;



      /// Total number of tracks.
  @JsonKey(
    
    name: r'tracks',
    required: true,
    includeIfNull: false,
  )


  final int tracks;



      /// Total number of playlists.
  @JsonKey(
    
    name: r'playlists',
    required: true,
    includeIfNull: false,
  )


  final int playlists;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogStats &&
      other.artists == artists &&
      other.albums == albums &&
      other.tracks == tracks &&
      other.playlists == playlists;

    @override
    int get hashCode =>
        artists.hashCode +
        albums.hashCode +
        tracks.hashCode +
        playlists.hashCode;

  factory CatalogStats.fromJson(Map<String, dynamic> json) => _$CatalogStatsFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogStatsToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

