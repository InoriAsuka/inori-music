//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_artist.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogArtist {
  /// Returns a new [CatalogArtist] instance.
  CatalogArtist({

    required  this.createdAt,

    required  this.id,

    required  this.name,

     this.sortName,

    required  this.updatedAt,
  });

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



  @JsonKey(
    
    name: r'name',
    required: true,
    includeIfNull: false,
  )


  final String name;



  @JsonKey(
    
    name: r'sortName',
    required: false,
    includeIfNull: false,
  )


  final String? sortName;



  @JsonKey(
    
    name: r'updatedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime updatedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogArtist &&
      other.createdAt == createdAt &&
      other.id == id &&
      other.name == name &&
      other.sortName == sortName &&
      other.updatedAt == updatedAt;

    @override
    int get hashCode =>
        createdAt.hashCode +
        id.hashCode +
        name.hashCode +
        sortName.hashCode +
        updatedAt.hashCode;

  factory CatalogArtist.fromJson(Map<String, dynamic> json) => _$CatalogArtistFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogArtistToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

