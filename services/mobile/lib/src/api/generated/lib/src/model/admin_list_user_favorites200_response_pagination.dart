//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'admin_list_user_favorites200_response_pagination.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class AdminListUserFavorites200ResponsePagination {
  /// Returns a new [AdminListUserFavorites200ResponsePagination] instance.
  AdminListUserFavorites200ResponsePagination({

     this.limit,

     this.offset,

     this.total,

     this.hasMore,
  });

  @JsonKey(
    
    name: r'limit',
    required: false,
    includeIfNull: false,
  )


  final int? limit;



  @JsonKey(
    
    name: r'offset',
    required: false,
    includeIfNull: false,
  )


  final int? offset;



  @JsonKey(
    
    name: r'total',
    required: false,
    includeIfNull: false,
  )


  final int? total;



  @JsonKey(
    
    name: r'hasMore',
    required: false,
    includeIfNull: false,
  )


  final bool? hasMore;





    @override
    bool operator ==(Object other) => identical(this, other) || other is AdminListUserFavorites200ResponsePagination &&
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

  factory AdminListUserFavorites200ResponsePagination.fromJson(Map<String, dynamic> json) => _$AdminListUserFavorites200ResponsePaginationFromJson(json);

  Map<String, dynamic> toJson() => _$AdminListUserFavorites200ResponsePaginationToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

