//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:json_annotation/json_annotation.dart';


enum HealthStatus {
      @JsonValue(r'unknown')
      unknown(r'unknown'),
      @JsonValue(r'healthy')
      healthy(r'healthy'),
      @JsonValue(r'unhealthy')
      unhealthy(r'unhealthy'),
      @JsonValue(r'disabled')
      disabled(r'disabled');

  const HealthStatus(this.value);

  final String value;

  @override
  String toString() => value;
}
