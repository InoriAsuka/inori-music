//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_batch_import_result_item.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_batch_import_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogBatchImportResult {
  /// Returns a new [CatalogBatchImportResult] instance.
  CatalogBatchImportResult({

    required  this.total,

    required  this.imported,

    required  this.failed,

    required  this.items,
  });

  @JsonKey(
    
    name: r'total',
    required: true,
    includeIfNull: false,
  )


  final int total;



  @JsonKey(
    
    name: r'imported',
    required: true,
    includeIfNull: false,
  )


  final int imported;



  @JsonKey(
    
    name: r'failed',
    required: true,
    includeIfNull: false,
  )


  final int failed;



  @JsonKey(
    
    name: r'items',
    required: true,
    includeIfNull: false,
  )


  final List<CatalogBatchImportResultItem> items;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogBatchImportResult &&
      other.total == total &&
      other.imported == imported &&
      other.failed == failed &&
      other.items == items;

    @override
    int get hashCode =>
        total.hashCode +
        imported.hashCode +
        failed.hashCode +
        items.hashCode;

  factory CatalogBatchImportResult.fromJson(Map<String, dynamic> json) => _$CatalogBatchImportResultFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogBatchImportResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

