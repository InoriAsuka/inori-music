//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'album_artwork_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class AlbumArtworkResponse {
  /// Returns a new [AlbumArtworkResponse] instance.
  AlbumArtworkResponse({

    required  this.url,

    required  this.expiresIn,
  });

      /// Presigned URL for the artwork
  @JsonKey(
    
    name: r'url',
    required: true,
    includeIfNull: false,
  )


  final String url;



      /// Seconds until the URL expires
  @JsonKey(
    
    name: r'expiresIn',
    required: true,
    includeIfNull: false,
  )


  final int expiresIn;





    @override
    bool operator ==(Object other) => identical(this, other) || other is AlbumArtworkResponse &&
      other.url == url &&
      other.expiresIn == expiresIn;

    @override
    int get hashCode =>
        url.hashCode +
        expiresIn.hashCode;

  factory AlbumArtworkResponse.fromJson(Map<String, dynamic> json) => _$AlbumArtworkResponseFromJson(json);

  Map<String, dynamic> toJson() => _$AlbumArtworkResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

