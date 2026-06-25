//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_selection_filter.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectSelectionFilter {
  /// Returns a new [MediaObjectSelectionFilter] instance.
  MediaObjectSelectionFilter({

     this.assetKind,

     this.backendId,

     this.contentHash,

     this.lifecycleState,

     this.verificationStatus,
  });

  @JsonKey(
    
    name: r'assetKind',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectSelectionFilterAssetKindEnum? assetKind;



  @JsonKey(
    
    name: r'backendId',
    required: false,
    includeIfNull: false,
  )


  final String? backendId;



  @JsonKey(
    
    name: r'contentHash',
    required: false,
    includeIfNull: false,
  )


  final String? contentHash;



  @JsonKey(
    
    name: r'lifecycleState',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectSelectionFilterLifecycleStateEnum? lifecycleState;



  @JsonKey(
    
    name: r'verificationStatus',
    required: false,
    includeIfNull: false,
  )


  final MediaObjectSelectionFilterVerificationStatusEnum? verificationStatus;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectSelectionFilter &&
      other.assetKind == assetKind &&
      other.backendId == backendId &&
      other.contentHash == contentHash &&
      other.lifecycleState == lifecycleState &&
      other.verificationStatus == verificationStatus;

    @override
    int get hashCode =>
        assetKind.hashCode +
        backendId.hashCode +
        contentHash.hashCode +
        lifecycleState.hashCode +
        verificationStatus.hashCode;

  factory MediaObjectSelectionFilter.fromJson(Map<String, dynamic> json) => _$MediaObjectSelectionFilterFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectSelectionFilterToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum MediaObjectSelectionFilterAssetKindEnum {
@JsonValue(r'original_audio')
originalAudio(r'original_audio'),
@JsonValue(r'transcoded_audio')
transcodedAudio(r'transcoded_audio'),
@JsonValue(r'artwork')
artwork(r'artwork'),
@JsonValue(r'lyrics')
lyrics(r'lyrics'),
@JsonValue(r'waveform')
waveform(r'waveform'),
@JsonValue(r'analysis')
analysis(r'analysis'),
@JsonValue(r'import_package')
importPackage(r'import_package'),
@JsonValue(r'backup')
backup(r'backup');

const MediaObjectSelectionFilterAssetKindEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectSelectionFilterLifecycleStateEnum {
@JsonValue(r'staged')
staged(r'staged'),
@JsonValue(r'active')
active(r'active'),
@JsonValue(r'archived')
archived(r'archived'),
@JsonValue(r'deleted')
deleted(r'deleted');

const MediaObjectSelectionFilterLifecycleStateEnum(this.value);

final String value;

@override
String toString() => value;
}



enum MediaObjectSelectionFilterVerificationStatusEnum {
@JsonValue(r'verified')
verified(r'verified'),
@JsonValue(r'failed')
failed(r'failed'),
@JsonValue(r'unknown')
unknown(r'unknown');

const MediaObjectSelectionFilterVerificationStatusEnum(this.value);

final String value;

@override
String toString() => value;
}


