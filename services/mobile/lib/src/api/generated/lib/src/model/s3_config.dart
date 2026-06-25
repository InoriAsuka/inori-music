//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 's3_config.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class S3Config {
  /// Returns a new [S3Config] instance.
  S3Config({

    required  this.accessKeySecretRef,

    required  this.bucket,

    required  this.endpoint,

     this.pathStyle,

     this.region,

    required  this.secretKeySecretRef,
  });

  @JsonKey(
    
    name: r'accessKeySecretRef',
    required: true,
    includeIfNull: false,
  )


  final String accessKeySecretRef;



  @JsonKey(
    
    name: r'bucket',
    required: true,
    includeIfNull: false,
  )


  final String bucket;



  @JsonKey(
    
    name: r'endpoint',
    required: true,
    includeIfNull: false,
  )


  final String endpoint;



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
    required: true,
    includeIfNull: false,
  )


  final String secretKeySecretRef;





    @override
    bool operator ==(Object other) => identical(this, other) || other is S3Config &&
      other.accessKeySecretRef == accessKeySecretRef &&
      other.bucket == bucket &&
      other.endpoint == endpoint &&
      other.pathStyle == pathStyle &&
      other.region == region &&
      other.secretKeySecretRef == secretKeySecretRef;

    @override
    int get hashCode =>
        accessKeySecretRef.hashCode +
        bucket.hashCode +
        endpoint.hashCode +
        pathStyle.hashCode +
        region.hashCode +
        secretKeySecretRef.hashCode;

  factory S3Config.fromJson(Map<String, dynamic> json) => _$S3ConfigFromJson(json);

  Map<String, dynamic> toJson() => _$S3ConfigToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

