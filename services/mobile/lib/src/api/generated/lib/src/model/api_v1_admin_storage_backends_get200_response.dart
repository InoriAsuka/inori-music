//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/storage_backend.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'api_v1_admin_storage_backends_get200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ApiV1AdminStorageBackendsGet200Response {
  /// Returns a new [ApiV1AdminStorageBackendsGet200Response] instance.
  ApiV1AdminStorageBackendsGet200Response({

    required  this.backends,
  });

  @JsonKey(
    
    name: r'backends',
    required: true,
    includeIfNull: false,
  )


  final List<StorageBackend> backends;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ApiV1AdminStorageBackendsGet200Response &&
      other.backends == backends;

    @override
    int get hashCode =>
        backends.hashCode;

  factory ApiV1AdminStorageBackendsGet200Response.fromJson(Map<String, dynamic> json) => _$ApiV1AdminStorageBackendsGet200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ApiV1AdminStorageBackendsGet200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

