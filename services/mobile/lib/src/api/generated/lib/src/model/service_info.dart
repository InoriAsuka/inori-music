//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'service_info.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ServiceInfo {
  /// Returns a new [ServiceInfo] instance.
  ServiceInfo({

    required  this.buildTime,

    required  this.commit,

    required  this.name,

    required  this.version,
  });

  @JsonKey(
    
    name: r'buildTime',
    required: true,
    includeIfNull: false,
  )


  final String buildTime;



  @JsonKey(
    
    name: r'commit',
    required: true,
    includeIfNull: false,
  )


  final String commit;



  @JsonKey(
    
    name: r'name',
    required: true,
    includeIfNull: false,
  )


  final String name;



  @JsonKey(
    
    name: r'version',
    required: true,
    includeIfNull: false,
  )


  final String version;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ServiceInfo &&
      other.buildTime == buildTime &&
      other.commit == commit &&
      other.name == name &&
      other.version == version;

    @override
    int get hashCode =>
        buildTime.hashCode +
        commit.hashCode +
        name.hashCode +
        version.hashCode;

  factory ServiceInfo.fromJson(Map<String, dynamic> json) => _$ServiceInfoFromJson(json);

  Map<String, dynamic> toJson() => _$ServiceInfoToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

