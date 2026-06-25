//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_update_album_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogUpdateAlbumRequest {
  /// Returns a new [CatalogUpdateAlbumRequest] instance.
  CatalogUpdateAlbumRequest({

     this.title,

     this.sortTitle,

     this.artistId,

     this.releaseYear,
  });

      /// Album title. Must not be empty if provided.
  @JsonKey(
    
    name: r'title',
    required: false,
    includeIfNull: false,
  )


  final String? title;



      /// Sort key for the album title. May be empty to clear.
  @JsonKey(
    
    name: r'sortTitle',
    required: false,
    includeIfNull: false,
  )


  final String? sortTitle;



      /// ID of the owning artist. Must reference an existing artist.
  @JsonKey(
    
    name: r'artistId',
    required: false,
    includeIfNull: false,
  )


  final String? artistId;



      /// Release year. 0 clears the value.
          // minimum: 0
  @JsonKey(
    
    name: r'releaseYear',
    required: false,
    includeIfNull: false,
  )


  final int? releaseYear;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogUpdateAlbumRequest &&
      other.title == title &&
      other.sortTitle == sortTitle &&
      other.artistId == artistId &&
      other.releaseYear == releaseYear;

    @override
    int get hashCode =>
        title.hashCode +
        sortTitle.hashCode +
        artistId.hashCode +
        releaseYear.hashCode;

  factory CatalogUpdateAlbumRequest.fromJson(Map<String, dynamic> json) => _$CatalogUpdateAlbumRequestFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogUpdateAlbumRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

