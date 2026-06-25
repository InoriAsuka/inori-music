//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/track_play_count.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'top_tracks_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class TopTracksResult {
  /// Returns a new [TopTracksResult] instance.
  TopTracksResult({

    required  this.tracks,
  });

  @JsonKey(
    
    name: r'tracks',
    required: true,
    includeIfNull: false,
  )


  final List<TrackPlayCount> tracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is TopTracksResult &&
      other.tracks == tracks;

    @override
    int get hashCode =>
        tracks.hashCode;

  factory TopTracksResult.fromJson(Map<String, dynamic> json) => _$TopTracksResultFromJson(json);

  Map<String, dynamic> toJson() => _$TopTracksResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

