//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'nfs_config.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class NFSConfig {
  /// Returns a new [NFSConfig] instance.
  NFSConfig({

    required  this.expectedRemote,

    required  this.mountPath,
  });

  @JsonKey(
    
    name: r'expectedRemote',
    required: true,
    includeIfNull: false,
  )


  final String expectedRemote;



  @JsonKey(
    
    name: r'mountPath',
    required: true,
    includeIfNull: false,
  )


  final String mountPath;





    @override
    bool operator ==(Object other) => identical(this, other) || other is NFSConfig &&
      other.expectedRemote == expectedRemote &&
      other.mountPath == mountPath;

    @override
    int get hashCode =>
        expectedRemote.hashCode +
        mountPath.hashCode;

  factory NFSConfig.fromJson(Map<String, dynamic> json) => _$NFSConfigFromJson(json);

  Map<String, dynamic> toJson() => _$NFSConfigToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

