//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_pagination_meta.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogPaginationMeta {
  /// Returns a new [CatalogPaginationMeta] instance.
  CatalogPaginationMeta({

    required  this.limit,

    required  this.offset,

    required  this.total,

    required  this.hasMore,
  });

      /// Maximum number of items in the page.
  @JsonKey(
    
    name: r'limit',
    required: true,
    includeIfNull: false,
  )


  final int limit;



      /// Zero-based index of the first item in the page.
  @JsonKey(
    
    name: r'offset',
    required: true,
    includeIfNull: false,
  )


  final int offset;



      /// Total number of items across all pages.
  @JsonKey(
    
    name: r'total',
    required: true,
    includeIfNull: false,
  )


  final int total;



      /// True when more items exist after the current page.
  @JsonKey(
    
    name: r'hasMore',
    required: true,
    includeIfNull: false,
  )


  final bool hasMore;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogPaginationMeta &&
      other.limit == limit &&
      other.offset == offset &&
      other.total == total &&
      other.hasMore == hasMore;

    @override
    int get hashCode =>
        limit.hashCode +
        offset.hashCode +
        total.hashCode +
        hasMore.hashCode;

  factory CatalogPaginationMeta.fromJson(Map<String, dynamic> json) => _$CatalogPaginationMetaFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogPaginationMetaToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

