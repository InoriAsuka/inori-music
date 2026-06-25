//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_batch_import_result_item.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogBatchImportResultItem {
  /// Returns a new [CatalogBatchImportResultItem] instance.
  CatalogBatchImportResultItem({

    required  this.index,

    required  this.mediaObjectId,

     this.track,

     this.error,

     this.errorCode,
  });

      /// Zero-based position of this item in the request array
  @JsonKey(
    
    name: r'index',
    required: true,
    includeIfNull: false,
  )


  final int index;



  @JsonKey(
    
    name: r'mediaObjectId',
    required: true,
    includeIfNull: false,
  )


  final String mediaObjectId;



  @JsonKey(
    
    name: r'track',
    required: false,
    includeIfNull: false,
  )


  final CatalogTrack? track;



  @JsonKey(
    
    name: r'error',
    required: false,
    includeIfNull: false,
  )


  final String? error;



  @JsonKey(
    
    name: r'errorCode',
    required: false,
    includeIfNull: false,
  )


  final String? errorCode;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogBatchImportResultItem &&
      other.index == index &&
      other.mediaObjectId == mediaObjectId &&
      other.track == track &&
      other.error == error &&
      other.errorCode == errorCode;

    @override
    int get hashCode =>
        index.hashCode +
        mediaObjectId.hashCode +
        track.hashCode +
        error.hashCode +
        errorCode.hashCode;

  factory CatalogBatchImportResultItem.fromJson(Map<String, dynamic> json) => _$CatalogBatchImportResultItemFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogBatchImportResultItemToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

