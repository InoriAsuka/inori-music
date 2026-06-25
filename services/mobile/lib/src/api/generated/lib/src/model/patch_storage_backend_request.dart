//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'patch_storage_backend_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class PatchStorageBackendRequest {
  /// Returns a new [PatchStorageBackendRequest] instance.
  PatchStorageBackendRequest({

     this.displayName,

     this.priority,
  });

      /// New human-readable name
  @JsonKey(
    
    name: r'displayName',
    required: false,
    includeIfNull: false,
  )


  final String? displayName;



      /// Scheduling priority (lower = higher priority)
  @JsonKey(
    
    name: r'priority',
    required: false,
    includeIfNull: false,
  )


  final int? priority;





    @override
    bool operator ==(Object other) => identical(this, other) || other is PatchStorageBackendRequest &&
      other.displayName == displayName &&
      other.priority == priority;

    @override
    int get hashCode =>
        displayName.hashCode +
        priority.hashCode;

  factory PatchStorageBackendRequest.fromJson(Map<String, dynamic> json) => _$PatchStorageBackendRequestFromJson(json);

  Map<String, dynamic> toJson() => _$PatchStorageBackendRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

