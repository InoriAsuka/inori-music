//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'local_config.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class LocalConfig {
  /// Returns a new [LocalConfig] instance.
  LocalConfig({

    required  this.rootPath,
  });

  @JsonKey(
    
    name: r'rootPath',
    required: true,
    includeIfNull: false,
  )


  final String rootPath;





    @override
    bool operator ==(Object other) => identical(this, other) || other is LocalConfig &&
      other.rootPath == rootPath;

    @override
    int get hashCode =>
        rootPath.hashCode;

  factory LocalConfig.fromJson(Map<String, dynamic> json) => _$LocalConfigFromJson(json);

  Map<String, dynamic> toJson() => _$LocalConfigToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

