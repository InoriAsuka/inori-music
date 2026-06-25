//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/nfs_config.dart';
import 'package:inori_api/src/model/s3_config.dart';
import 'package:inori_api/src/model/local_config.dart';
import 'package:inori_api/src/model/distributed_config.dart';
import 'package:inori_api/src/model/smb_config.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'backend_config.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class BackendConfig {
  /// Returns a new [BackendConfig] instance.
  BackendConfig({

     this.distributed,

     this.local,

     this.nfs,

     this.s3,

     this.smb,
  });

  @JsonKey(
    
    name: r'distributed',
    required: false,
    includeIfNull: false,
  )


  final DistributedConfig? distributed;



  @JsonKey(
    
    name: r'local',
    required: false,
    includeIfNull: false,
  )


  final LocalConfig? local;



  @JsonKey(
    
    name: r'nfs',
    required: false,
    includeIfNull: false,
  )


  final NFSConfig? nfs;



  @JsonKey(
    
    name: r's3',
    required: false,
    includeIfNull: false,
  )


  final S3Config? s3;



  @JsonKey(
    
    name: r'smb',
    required: false,
    includeIfNull: false,
  )


  final SMBConfig? smb;





    @override
    bool operator ==(Object other) => identical(this, other) || other is BackendConfig &&
      other.distributed == distributed &&
      other.local == local &&
      other.nfs == nfs &&
      other.s3 == s3 &&
      other.smb == smb;

    @override
    int get hashCode =>
        distributed.hashCode +
        local.hashCode +
        nfs.hashCode +
        s3.hashCode +
        smb.hashCode;

  factory BackendConfig.fromJson(Map<String, dynamic> json) => _$BackendConfigFromJson(json);

  Map<String, dynamic> toJson() => _$BackendConfigToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

