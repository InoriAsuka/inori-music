//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'patch_media_object_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class PatchMediaObjectRequest {
  /// Returns a new [PatchMediaObjectRequest] instance.
  PatchMediaObjectRequest({

     this.assetKind,

     this.mimeType,
  });

  @JsonKey(
    
    name: r'assetKind',
    required: false,
    includeIfNull: false,
  )


  final String? assetKind;



  @JsonKey(
    
    name: r'mimeType',
    required: false,
    includeIfNull: false,
  )


  final String? mimeType;





    @override
    bool operator ==(Object other) => identical(this, other) || other is PatchMediaObjectRequest &&
      other.assetKind == assetKind &&
      other.mimeType == mimeType;

    @override
    int get hashCode =>
        assetKind.hashCode +
        mimeType.hashCode;

  factory PatchMediaObjectRequest.fromJson(Map<String, dynamic> json) => _$PatchMediaObjectRequestFromJson(json);

  Map<String, dynamic> toJson() => _$PatchMediaObjectRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

