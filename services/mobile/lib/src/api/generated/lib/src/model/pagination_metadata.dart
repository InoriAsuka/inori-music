//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'pagination_metadata.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class PaginationMetadata {
  /// Returns a new [PaginationMetadata] instance.
  PaginationMetadata({

    required  this.hasMore,

    required  this.limit,

    required  this.offset,

    required  this.total,
  });

  @JsonKey(
    
    name: r'hasMore',
    required: true,
    includeIfNull: false,
  )


  final bool hasMore;



          // minimum: 1
          // maximum: 500
  @JsonKey(
    
    name: r'limit',
    required: true,
    includeIfNull: false,
  )


  final int limit;



          // minimum: 0
  @JsonKey(
    
    name: r'offset',
    required: true,
    includeIfNull: false,
  )


  final int offset;



          // minimum: 0
  @JsonKey(
    
    name: r'total',
    required: true,
    includeIfNull: false,
  )


  final int total;





    @override
    bool operator ==(Object other) => identical(this, other) || other is PaginationMetadata &&
      other.hasMore == hasMore &&
      other.limit == limit &&
      other.offset == offset &&
      other.total == total;

    @override
    int get hashCode =>
        hasMore.hashCode +
        limit.hashCode +
        offset.hashCode +
        total.hashCode;

  factory PaginationMetadata.fromJson(Map<String, dynamic> json) => _$PaginationMetadataFromJson(json);

  Map<String, dynamic> toJson() => _$PaginationMetadataToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

