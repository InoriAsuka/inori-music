//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'add_playlist_track_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class AddPlaylistTrackRequest {
  /// Returns a new [AddPlaylistTrackRequest] instance.
  AddPlaylistTrackRequest({

    required  this.trackId,
  });

  @JsonKey(
    
    name: r'trackId',
    required: true,
    includeIfNull: false,
  )


  final String trackId;





    @override
    bool operator ==(Object other) => identical(this, other) || other is AddPlaylistTrackRequest &&
      other.trackId == trackId;

    @override
    int get hashCode =>
        trackId.hashCode;

  factory AddPlaylistTrackRequest.fromJson(Map<String, dynamic> json) => _$AddPlaylistTrackRequestFromJson(json);

  Map<String, dynamic> toJson() => _$AddPlaylistTrackRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

