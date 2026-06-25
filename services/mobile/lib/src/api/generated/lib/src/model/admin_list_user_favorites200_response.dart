//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/admin_list_user_favorites200_response_pagination.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'admin_list_user_favorites200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class AdminListUserFavorites200Response {
  /// Returns a new [AdminListUserFavorites200Response] instance.
  AdminListUserFavorites200Response({

     this.trackIds,

     this.pagination,
  });

  @JsonKey(
    
    name: r'trackIds',
    required: false,
    includeIfNull: false,
  )


  final List<String>? trackIds;



  @JsonKey(
    
    name: r'pagination',
    required: false,
    includeIfNull: false,
  )


  final AdminListUserFavorites200ResponsePagination? pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is AdminListUserFavorites200Response &&
      other.trackIds == trackIds &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        trackIds.hashCode +
        pagination.hashCode;

  factory AdminListUserFavorites200Response.fromJson(Map<String, dynamic> json) => _$AdminListUserFavorites200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$AdminListUserFavorites200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

