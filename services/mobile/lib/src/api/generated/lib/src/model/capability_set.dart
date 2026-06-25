//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'capability_set.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CapabilitySet {
  /// Returns a new [CapabilitySet] instance.
  CapabilitySet({

     this.crossNodeAccess,

     this.multipartUpload,

     this.nativeLifecyclePolicy,

     this.presignedUrls,

     this.requiresCredentialValidation,

     this.requiresMountValidation,

     this.serverRangeReads,
  });

  @JsonKey(
    
    name: r'crossNodeAccess',
    required: false,
    includeIfNull: false,
  )


  final bool? crossNodeAccess;



  @JsonKey(
    
    name: r'multipartUpload',
    required: false,
    includeIfNull: false,
  )


  final bool? multipartUpload;



  @JsonKey(
    
    name: r'nativeLifecyclePolicy',
    required: false,
    includeIfNull: false,
  )


  final bool? nativeLifecyclePolicy;



  @JsonKey(
    
    name: r'presignedUrls',
    required: false,
    includeIfNull: false,
  )


  final bool? presignedUrls;



  @JsonKey(
    
    name: r'requiresCredentialValidation',
    required: false,
    includeIfNull: false,
  )


  final bool? requiresCredentialValidation;



  @JsonKey(
    
    name: r'requiresMountValidation',
    required: false,
    includeIfNull: false,
  )


  final bool? requiresMountValidation;



  @JsonKey(
    
    name: r'serverRangeReads',
    required: false,
    includeIfNull: false,
  )


  final bool? serverRangeReads;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CapabilitySet &&
      other.crossNodeAccess == crossNodeAccess &&
      other.multipartUpload == multipartUpload &&
      other.nativeLifecyclePolicy == nativeLifecyclePolicy &&
      other.presignedUrls == presignedUrls &&
      other.requiresCredentialValidation == requiresCredentialValidation &&
      other.requiresMountValidation == requiresMountValidation &&
      other.serverRangeReads == serverRangeReads;

    @override
    int get hashCode =>
        crossNodeAccess.hashCode +
        multipartUpload.hashCode +
        nativeLifecyclePolicy.hashCode +
        presignedUrls.hashCode +
        requiresCredentialValidation.hashCode +
        requiresMountValidation.hashCode +
        serverRangeReads.hashCode;

  factory CapabilitySet.fromJson(Map<String, dynamic> json) => _$CapabilitySetFromJson(json);

  Map<String, dynamic> toJson() => _$CapabilitySetToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

