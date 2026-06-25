//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/media_object_timeline_event.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'media_object_timeline.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MediaObjectTimeline {
  /// Returns a new [MediaObjectTimeline] instance.
  MediaObjectTimeline({

    required  this.events,

    required  this.mediaObjectId,
  });

  @JsonKey(
    
    name: r'events',
    required: true,
    includeIfNull: false,
  )


  final List<MediaObjectTimelineEvent> events;



  @JsonKey(
    
    name: r'mediaObjectId',
    required: true,
    includeIfNull: false,
  )


  final String mediaObjectId;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MediaObjectTimeline &&
      other.events == events &&
      other.mediaObjectId == mediaObjectId;

    @override
    int get hashCode =>
        events.hashCode +
        mediaObjectId.hashCode;

  factory MediaObjectTimeline.fromJson(Map<String, dynamic> json) => _$MediaObjectTimelineFromJson(json);

  Map<String, dynamic> toJson() => _$MediaObjectTimelineToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

