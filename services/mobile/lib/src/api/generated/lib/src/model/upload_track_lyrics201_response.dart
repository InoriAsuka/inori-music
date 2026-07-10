//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'upload_track_lyrics201_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UploadTrackLyrics201Response {
  /// Returns a new [UploadTrackLyrics201Response] instance.
  UploadTrackLyrics201Response({

    required  this.mediaObjectId,

     this.translationMediaObjectId,
  });

  @JsonKey(
    
    name: r'mediaObjectId',
    required: true,
    includeIfNull: false,
  )


  final String mediaObjectId;



  @JsonKey(
    
    name: r'translationMediaObjectId',
    required: false,
    includeIfNull: false,
  )


  final String? translationMediaObjectId;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UploadTrackLyrics201Response &&
      other.mediaObjectId == mediaObjectId &&
      other.translationMediaObjectId == translationMediaObjectId;

    @override
    int get hashCode =>
        mediaObjectId.hashCode +
        translationMediaObjectId.hashCode;

  factory UploadTrackLyrics201Response.fromJson(Map<String, dynamic> json) => _$UploadTrackLyrics201ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$UploadTrackLyrics201ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

