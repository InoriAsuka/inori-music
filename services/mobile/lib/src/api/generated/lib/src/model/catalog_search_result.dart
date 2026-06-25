//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/search_result_item.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_search_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogSearchResult {
  /// Returns a new [CatalogSearchResult] instance.
  CatalogSearchResult({

    required  this.query,

    required  this.items,
  });

      /// The search query string as submitted.
  @JsonKey(
    
    name: r'query',
    required: true,
    includeIfNull: false,
  )


  final String query;



      /// Ordered result items: artists first, then albums, then tracks.
  @JsonKey(
    
    name: r'items',
    required: true,
    includeIfNull: false,
  )


  final List<SearchResultItem> items;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogSearchResult &&
      other.query == query &&
      other.items == items;

    @override
    int get hashCode =>
        query.hashCode +
        items.hashCode;

  factory CatalogSearchResult.fromJson(Map<String, dynamic> json) => _$CatalogSearchResultFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogSearchResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

