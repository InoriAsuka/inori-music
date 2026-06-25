//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/updated_catalog_item.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'updated_catalog_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UpdatedCatalogResult {
  /// Returns a new [UpdatedCatalogResult] instance.
  UpdatedCatalogResult({

    required  this.items,
  });

      /// Updated artist, album, track, and playlist entries ordered newest-first.
  @JsonKey(
    
    name: r'items',
    required: true,
    includeIfNull: false,
  )


  final List<UpdatedCatalogItem> items;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UpdatedCatalogResult &&
      other.items == items;

    @override
    int get hashCode =>
        items.hashCode;

  factory UpdatedCatalogResult.fromJson(Map<String, dynamic> json) => _$UpdatedCatalogResultFromJson(json);

  Map<String, dynamic> toJson() => _$UpdatedCatalogResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

