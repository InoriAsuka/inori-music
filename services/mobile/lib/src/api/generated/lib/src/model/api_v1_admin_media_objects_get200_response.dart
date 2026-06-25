//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object.dart';
import 'package:inori_api/src/model/pagination_metadata.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_media_objects_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminMediaObjectsGet200Response {
  /// Returns a new [ApiV1AdminMediaObjectsGet200Response] instance.
  ApiV1AdminMediaObjectsGet200Response({

    required  this.objects,

    required  this.pagination,
  });

  @JsonKey(
    
    name: r'objects',
    required: true,
    includeIfNull: false,
  )


  final List<MediaObject> objects;



  @JsonKey(
    
    name: r'pagination',
    required: true,
    includeIfNull: false,
  )


  final PaginationMetadata pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminMediaObjectsGet200Response &&
      other.objects == objects &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        objects.hashCode +
        pagination.hashCode;

  factory ApiV1AdminMediaObjectsGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1AdminMediaObjectsGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminMediaObjectsGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

