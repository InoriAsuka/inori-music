//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'set_playlist_tracks_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class SetPlaylistTracksRequest {
  /// Returns a new [SetPlaylistTracksRequest] instance.
  SetPlaylistTracksRequest({

    required  this.trackIds,
  });

      /// Ordered list of track IDs. Replaces the current track list atomically. An empty array clears the playlist.
  @JsonKey(
    
    name: r'trackIds',
    required: true,
    includeIfNull: false,
  )


  final List<String> trackIds;





    @override
    bool operator ==(Object other) => identical(this, other) || other is SetPlaylistTracksRequest &&
      other.trackIds == trackIds;

    @override
    int get hashCode =>
        trackIds.hashCode;

  factory SetPlaylistTracksRequest.fromJson(Map<String, dynamic> json) => _$SetPlaylistTracksRequestFromJson(json);

  Map<String, dynamic> toJson() => _$SetPlaylistTracksRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

