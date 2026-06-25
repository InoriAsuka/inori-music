//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_album.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogAlbum {
  /// Returns a new [CatalogAlbum] instance.
  CatalogAlbum({

    required  this.artistId,

    required  this.createdAt,

    required  this.id,

     this.releaseYear,

     this.sortTitle,

    required  this.title,

    required  this.updatedAt,
  });

  @JsonKey(
    
    name: r'artistId',
    required: true,
    includeIfNull: false,
  )


  final String artistId;



  @JsonKey(
    
    name: r'createdAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime createdAt;



  @JsonKey(
    
    name: r'id',
    required: true,
    includeIfNull: false,
  )


  final String id;



          // minimum: 0
  @JsonKey(
    
    name: r'releaseYear',
    required: false,
    includeIfNull: false,
  )


  final int? releaseYear;



  @JsonKey(
    
    name: r'sortTitle',
    required: false,
    includeIfNull: false,
  )


  final String? sortTitle;



  @JsonKey(
    
    name: r'title',
    required: true,
    includeIfNull: false,
  )


  final String title;



  @JsonKey(
    
    name: r'updatedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime updatedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogAlbum &&
      other.artistId == artistId &&
      other.createdAt == createdAt &&
      other.id == id &&
      other.releaseYear == releaseYear &&
      other.sortTitle == sortTitle &&
      other.title == title &&
      other.updatedAt == updatedAt;

    @override
    int get hashCode =>
        artistId.hashCode +
        createdAt.hashCode +
        id.hashCode +
        releaseYear.hashCode +
        sortTitle.hashCode +
        title.hashCode +
        updatedAt.hashCode;

  factory CatalogAlbum.fromJson(Map<String, dynamic> json) => _$CatalogAlbumFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogAlbumToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

