//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'set_user_playlist_tracks_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class SetUserPlaylistTracksRequest {
  /// Returns a new [SetUserPlaylistTracksRequest] instance.
  SetUserPlaylistTracksRequest({

    required  this.trackIds,
  });

  @JsonKey(
    
    name: r'trackIds',
    required: true,
    includeIfNull: false,
  )


  final List<String> trackIds;





    @override
    bool operator ==(Object other) => identical(this, other) || other is SetUserPlaylistTracksRequest &&
      other.trackIds == trackIds;

    @override
    int get hashCode =>
        trackIds.hashCode;

  factory SetUserPlaylistTracksRequest.fromJson(Map<String, dynamic> json) => _$SetUserPlaylistTracksRequestFromJson(json);

  Map<String, dynamic> toJson() => _$SetUserPlaylistTracksRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

