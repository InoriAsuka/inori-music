//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/capacity_report.dart';
import 'package:inori_api/src/model/health_status.dart';
import 'package:inori_api/src/model/capability_set.dart';
import 'package:inori_api/src/model/backend_type.dart';
import 'package:inori_api/src/model/backend_config.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'storage_backend.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class StorageBackend {
  /// Returns a new [StorageBackend] instance.
  StorageBackend({

    required  this.config,

    required  this.displayName,

    required  this.enabled,

    required  this.id,

     this.isDefault,

     this.priority,

    required  this.type,

     this.capabilities,

     this.createdAt,

     this.healthStatus,

     this.lastCapacity,

     this.lastHealthCheckAt,

     this.updatedAt,
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



  @JsonKey(
    
    name: r'capabilities',
    required: false,
    includeIfNull: false,
  )


  final CapabilitySet? capabilities;



  @JsonKey(
    
    name: r'createdAt',
    required: false,
    includeIfNull: false,
  )


  final DateTime? createdAt;



  @JsonKey(
    
    name: r'healthStatus',
    required: false,
    includeIfNull: false,
  )


  final HealthStatus? healthStatus;



  @JsonKey(
    
    name: r'lastCapacity',
    required: false,
    includeIfNull: false,
  )


  final CapacityReport? lastCapacity;



  @JsonKey(
    
    name: r'lastHealthCheckAt',
    required: false,
    includeIfNull: false,
  )


  final DateTime? lastHealthCheckAt;



  @JsonKey(
    
    name: r'updatedAt',
    required: false,
    includeIfNull: false,
  )


  final DateTime? updatedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is StorageBackend &&
      other.config == config &&
      other.displayName == displayName &&
      other.enabled == enabled &&
      other.id == id &&
      other.isDefault == isDefault &&
      other.priority == priority &&
      other.type == type &&
      other.capabilities == capabilities &&
      other.createdAt == createdAt &&
      other.healthStatus == healthStatus &&
      other.lastCapacity == lastCapacity &&
      other.lastHealthCheckAt == lastHealthCheckAt &&
      other.updatedAt == updatedAt;

    @override
    int get hashCode =>
        config.hashCode +
        displayName.hashCode +
        enabled.hashCode +
        id.hashCode +
        isDefault.hashCode +
        priority.hashCode +
        type.hashCode +
        capabilities.hashCode +
        createdAt.hashCode +
        healthStatus.hashCode +
        lastCapacity.hashCode +
        lastHealthCheckAt.hashCode +
        updatedAt.hashCode;

  factory StorageBackend.fromJson(Map<String, dynamic> json) => _$StorageBackendFromJson(json);

  Map<String, dynamic> toJson() => _$StorageBackendToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

