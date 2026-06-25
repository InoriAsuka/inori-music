//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_users_get200_response_pagination.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminUsersGet200ResponsePagination {
  /// Returns a new [ApiV1AdminUsersGet200ResponsePagination] instance.
  ApiV1AdminUsersGet200ResponsePagination({

    required  this.limit,

    required  this.offset,

    required  this.total,

    required  this.hasMore,
  });

  @JsonKey(
    
    name: r'limit',
    required: true,
    includeIfNull: false,
  )


  final int limit;



  @JsonKey(
    
    name: r'offset',
    required: true,
    includeIfNull: false,
  )


  final int offset;



  @JsonKey(
    
    name: r'total',
    required: true,
    includeIfNull: false,
  )


  final int total;



  @JsonKey(
    
    name: r'hasMore',
    required: true,
    includeIfNull: false,
  )


  final bool hasMore;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminUsersGet200ResponsePagination &&
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

  factory ApiV1AdminUsersGet200ResponsePagination.fromJson(Map<String, dynamic> json) => _$ApiV1AdminUsersGet200ResponsePaginationFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminUsersGet200ResponsePaginationToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

