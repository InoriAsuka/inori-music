//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_import_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogImportRequest {
  /// Returns a new [CatalogImportRequest] instance.
  CatalogImportRequest({

     this.albumId,

     this.artistId,

     this.discNumber,

     this.durationMs,

    required  this.mediaObjectId,

     this.sortTitle,

     this.title,

     this.trackNumber,
  });

  @JsonKey(
    
    name: r'albumId',
    required: false,
    includeIfNull: false,
  )


  final String? albumId;



  @JsonKey(
    
    name: r'artistId',
    required: false,
    includeIfNull: false,
  )


  final String? artistId;



          // minimum: 0
  @JsonKey(
    
    name: r'discNumber',
    required: false,
    includeIfNull: false,
  )


  final int? discNumber;



          // minimum: 0
  @JsonKey(
    
    name: r'durationMs',
    required: false,
    includeIfNull: false,
  )


  final int? durationMs;



      /// ID of the media object to import
  @JsonKey(
    
    name: r'mediaObjectId',
    required: true,
    includeIfNull: false,
  )


  final String mediaObjectId;



  @JsonKey(
    
    name: r'sortTitle',
    required: false,
    includeIfNull: false,
  )


  final String? sortTitle;



      /// Track title; falls back to mediaObjectId if empty
  @JsonKey(
    
    name: r'title',
    required: false,
    includeIfNull: false,
  )


  final String? title;



          // minimum: 0
  @JsonKey(
    
    name: r'trackNumber',
    required: false,
    includeIfNull: false,
  )


  final int? trackNumber;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogImportRequest &&
      other.albumId == albumId &&
      other.artistId == artistId &&
      other.discNumber == discNumber &&
      other.durationMs == durationMs &&
      other.mediaObjectId == mediaObjectId &&
      other.sortTitle == sortTitle &&
      other.title == title &&
      other.trackNumber == trackNumber;

    @override
    int get hashCode =>
        albumId.hashCode +
        artistId.hashCode +
        discNumber.hashCode +
        durationMs.hashCode +
        mediaObjectId.hashCode +
        sortTitle.hashCode +
        title.hashCode +
        trackNumber.hashCode;

  factory CatalogImportRequest.fromJson(Map<String, dynamic> json) => _$CatalogImportRequestFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogImportRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

