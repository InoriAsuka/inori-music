//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_duplicate_group.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectDuplicateGroup {
  /// Returns a new [MediaObjectDuplicateGroup] instance.
  MediaObjectDuplicateGroup({

    required  this.contentHash,

    required  this.count,

    required  this.objects,

    required  this.totalSizeBytes,
  });

      /// Content hash shared by every object in the group.
  @JsonKey(
    
    name: r'contentHash',
    required: true,
    includeIfNull: false,
  )


  final String contentHash;



          // minimum: 2
  @JsonKey(
    
    name: r'count',
    required: true,
    includeIfNull: false,
  )


  final int count;



  @JsonKey(
    
    name: r'objects',
    required: true,
    includeIfNull: false,
  )


  final List<MediaObject> objects;



          // minimum: 0
  @JsonKey(
    
    name: r'totalSizeBytes',
    required: true,
    includeIfNull: false,
  )


  final int totalSizeBytes;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectDuplicateGroup &&
      other.contentHash == contentHash &&
      other.count == count &&
      other.objects == objects &&
      other.totalSizeBytes == totalSizeBytes;

    @override
    int get hashCode =>
        contentHash.hashCode +
        count.hashCode +
        objects.hashCode +
        totalSizeBytes.hashCode;

  factory MediaObjectDuplicateGroup.fromJson(Map<String, dynamic> json) => _$MediaObjectDuplicateGroupFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectDuplicateGroupToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

