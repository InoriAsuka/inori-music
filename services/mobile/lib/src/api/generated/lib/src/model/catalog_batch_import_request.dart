//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_import_request.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_batch_import_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogBatchImportRequest {
  /// Returns a new [CatalogBatchImportRequest] instance.
  CatalogBatchImportRequest({

    required  this.items,
  });

      /// Import requests to process independently
  @JsonKey(
    
    name: r'items',
    required: true,
    includeIfNull: false,
  )


  final List<CatalogImportRequest> items;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogBatchImportRequest &&
      other.items == items;

    @override
    int get hashCode =>
        items.hashCode;

  factory CatalogBatchImportRequest.fromJson(Map<String, dynamic> json) => _$CatalogBatchImportRequestFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogBatchImportRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

