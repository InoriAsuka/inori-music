//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'smb_config.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class SMBConfig {
  /// Returns a new [SMBConfig] instance.
  SMBConfig({

    required  this.expectedShare,

    required  this.mountPath,
  });

  @JsonKey(
    
    name: r'expectedShare',
    required: true,
    includeIfNull: false,
  )


  final String expectedShare;



  @JsonKey(
    
    name: r'mountPath',
    required: true,
    includeIfNull: false,
  )


  final String mountPath;





    @override
    bool operator ==(Object other) => identical(this, other) || other is SMBConfig &&
      other.expectedShare == expectedShare &&
      other.mountPath == mountPath;

    @override
    int get hashCode =>
        expectedShare.hashCode +
        mountPath.hashCode;

  factory SMBConfig.fromJson(Map<String, dynamic> json) => _$SMBConfigFromJson(json);

  Map<String, dynamic> toJson() => _$SMBConfigToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

