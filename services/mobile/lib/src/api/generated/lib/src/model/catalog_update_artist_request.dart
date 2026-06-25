//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_update_artist_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogUpdateArtistRequest {
  /// Returns a new [CatalogUpdateArtistRequest] instance.
  CatalogUpdateArtistRequest({

     this.name,

     this.sortName,
  });

      /// Display name. Must not be empty if provided.
  @JsonKey(
    
    name: r'name',
    required: false,
    includeIfNull: false,
  )


  final String? name;



      /// Sort key for the artist name. May be empty to clear.
  @JsonKey(
    
    name: r'sortName',
    required: false,
    includeIfNull: false,
  )


  final String? sortName;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogUpdateArtistRequest &&
      other.name == name &&
      other.sortName == sortName;

    @override
    int get hashCode =>
        name.hashCode +
        sortName.hashCode;

  factory CatalogUpdateArtistRequest.fromJson(Map<String, dynamic> json) => _$CatalogUpdateArtistRequestFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogUpdateArtistRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

