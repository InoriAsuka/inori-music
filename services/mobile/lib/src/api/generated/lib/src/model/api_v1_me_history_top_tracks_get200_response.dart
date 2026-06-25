//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/track_play_count.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_me_history_top_tracks_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1MeHistoryTopTracksGet200Response {
  /// Returns a new [ApiV1MeHistoryTopTracksGet200Response] instance.
  ApiV1MeHistoryTopTracksGet200Response({

    required  this.tracks,
  });

  @JsonKey(
    
    name: r'tracks',
    required: true,
    includeIfNull: false,
  )


  final List<TrackPlayCount> tracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1MeHistoryTopTracksGet200Response &&
      other.tracks == tracks;

    @override
    int get hashCode =>
        tracks.hashCode;

  factory ApiV1MeHistoryTopTracksGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1MeHistoryTopTracksGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1MeHistoryTopTracksGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

