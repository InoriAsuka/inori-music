//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:json_annotation/json_annotation.dart';


enum BackendType {
      @JsonValue(r'local')
      local(r'local'),
      @JsonValue(r'nfs')
      nfs(r'nfs'),
      @JsonValue(r'smb')
      smb(r'smb'),
      @JsonValue(r's3')
      s3(r's3'),
      @JsonValue(r'distributed')
      distributed(r'distributed');

  const BackendType(this.value);

  final String value;

  @override
  String toString() => value;
}
