//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/backend_type.dart';
import 'package:inori_api/src/model/backend_config.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'storage_backend_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class StorageBackendRequest {
  /// Returns a new [StorageBackendRequest] instance.
  StorageBackendRequest({

    required  this.config,

    required  this.displayName,

    required  this.enabled,

    required  this.id,

     this.isDefault,

     this.priority,

    required  this.type,
  });

  @JsonKey(
    
    name: r'config',
    required: true,
    includeIfNull: false,
  )


  final BackendConfig config;



  @JsonKey(
    
    name: r'displayName',
    required: true,
    includeIfNull: false,
  )


  final String displayName;



  @JsonKey(
    
    name: r'enabled',
    required: true,
    includeIfNull: false,
  )


  final bool enabled;



  @JsonKey(
    
    name: r'id',
    required: true,
    includeIfNull: false,
  )


  final String id;



  @JsonKey(
    
    name: r'isDefault',
    required: false,
    includeIfNull: false,
  )


  final bool? isDefault;



  @JsonKey(
    
    name: r'priority',
    required: false,
    includeIfNull: false,
  )


  final int? priority;



  @JsonKey(
    
    name: r'type',
    required: true,
    includeIfNull: false,
  )


  final BackendType type;





    @override
    bool operator ==(Object other) => identical(this, other) || other is StorageBackendRequest &&
      other.config == config &&
      other.displayName == displayName &&
      other.enabled == enabled &&
      other.id == id &&
      other.isDefault == isDefault &&
      other.priority == priority &&
      other.type == type;

    @override
    int get hashCode =>
        config.hashCode +
        displayName.hashCode +
        enabled.hashCode +
        id.hashCode +
        isDefault.hashCode +
        priority.hashCode +
        type.hashCode;

  factory StorageBackendRequest.fromJson(Map<String, dynamic> json) => _$StorageBackendRequestFromJson(json);

  Map<String, dynamic> toJson() => _$StorageBackendRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

