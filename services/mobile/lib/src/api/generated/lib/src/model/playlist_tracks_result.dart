//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'playlist_tracks_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class PlaylistTracksResult {
  /// Returns a new [PlaylistTracksResult] instance.
  PlaylistTracksResult({

    required  this.tracks,
  });

      /// Ordered list of full track objects for the playlist
  @JsonKey(
    
    name: r'tracks',
    required: true,
    includeIfNull: false,
  )


  final List<CatalogTrack> tracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is PlaylistTracksResult &&
      other.tracks == tracks;

    @override
    int get hashCode =>
        tracks.hashCode;

  factory PlaylistTracksResult.fromJson(Map<String, dynamic> json) => _$PlaylistTracksResultFromJson(json);

  Map<String, dynamic> toJson() => _$PlaylistTracksResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

