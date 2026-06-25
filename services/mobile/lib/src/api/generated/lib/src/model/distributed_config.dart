//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'distributed_config.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class DistributedConfig {
  /// Returns a new [DistributedConfig] instance.
  DistributedConfig({

     this.accessKeySecretRef,

    required  this.adapter,

     this.bucket,

     this.endpoint,

     this.mountPath,

     this.pathStyle,

     this.region,

     this.secretKeySecretRef,
  });

  @JsonKey(
    
    name: r'accessKeySecretRef',
    required: false,
    includeIfNull: false,
  )


  final String? accessKeySecretRef;



  @JsonKey(
    
    name: r'adapter',
    required: true,
    includeIfNull: false,
  )


  final DistributedConfigAdapterEnum adapter;



  @JsonKey(
    
    name: r'bucket',
    required: false,
    includeIfNull: false,
  )


  final String? bucket;



  @JsonKey(
    
    name: r'endpoint',
    required: false,
    includeIfNull: false,
  )


  final String? endpoint;



  @JsonKey(
    
    name: r'mountPath',
    required: false,
    includeIfNull: false,
  )


  final String? mountPath;



  @JsonKey(
    
    name: r'pathStyle',
    required: false,
    includeIfNull: false,
  )


  final bool? pathStyle;



  @JsonKey(
    
    name: r'region',
    required: false,
    includeIfNull: false,
  )


  final String? region;



  @JsonKey(
    
    name: r'secretKeySecretRef',
    required: false,
    includeIfNull: false,
  )


  final String? secretKeySecretRef;





    @override
    bool operator ==(Object other) => identical(this, other) || other is DistributedConfig &&
      other.accessKeySecretRef == accessKeySecretRef &&
      other.adapter == adapter &&
      other.bucket == bucket &&
      other.endpoint == endpoint &&
      other.mountPath == mountPath &&
      other.pathStyle == pathStyle &&
      other.region == region &&
      other.secretKeySecretRef == secretKeySecretRef;

    @override
    int get hashCode =>
        accessKeySecretRef.hashCode +
        adapter.hashCode +
        bucket.hashCode +
        endpoint.hashCode +
        mountPath.hashCode +
        pathStyle.hashCode +
        region.hashCode +
        secretKeySecretRef.hashCode;

  factory DistributedConfig.fromJson(Map<String, dynamic> json) => _$DistributedConfigFromJson(json);

  Map<String, dynamic> toJson() => _$DistributedConfigToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum DistributedConfigAdapterEnum {
@JsonValue(r's3-compatible')
s3Compatible(r's3-compatible'),
@JsonValue(r'mounted-filesystem')
mountedFilesystem(r'mounted-filesystem'),
@JsonValue(r'dedicated')
dedicated(r'dedicated');

const DistributedConfigAdapterEnum(this.value);

final String value;

@override
String toString() => value;
}


