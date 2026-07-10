//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'add_user_playlist_track_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class AddUserPlaylistTrackRequest {
  /// Returns a new [AddUserPlaylistTrackRequest] instance.
  AddUserPlaylistTrackRequest({

    required  this.trackId,
  });

  @JsonKey(
    
    name: r'trackId',
    required: true,
    includeIfNull: false,
  )


  final String trackId;





    @override
    bool operator ==(Object other) => identical(this, other) || other is AddUserPlaylistTrackRequest &&
      other.trackId == trackId;

    @override
    int get hashCode =>
        trackId.hashCode;

  factory AddUserPlaylistTrackRequest.fromJson(Map<String, dynamic> json) => _$AddUserPlaylistTrackRequestFromJson(json);

  Map<String, dynamic> toJson() => _$AddUserPlaylistTrackRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

