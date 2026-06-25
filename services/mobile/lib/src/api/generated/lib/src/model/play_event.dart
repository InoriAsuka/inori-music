//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'play_event.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class PlayEvent {
  /// Returns a new [PlayEvent] instance.
  PlayEvent({

    required  this.id,

    required  this.userId,

    required  this.trackId,

    required  this.playedAt,

    required  this.createdAt,
  });

      /// Unique event identifier.
  @JsonKey(
    
    name: r'id',
    required: true,
    includeIfNull: false,
  )


  final String id;



      /// User who played the track.
  @JsonKey(
    
    name: r'userId',
    required: true,
    includeIfNull: false,
  )


  final String userId;



      /// Track that was played.
  @JsonKey(
    
    name: r'trackId',
    required: true,
    includeIfNull: false,
  )


  final String trackId;



      /// Client-reported play timestamp.
  @JsonKey(
    
    name: r'playedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime playedAt;



      /// Server-side record creation time.
  @JsonKey(
    
    name: r'createdAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime createdAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is PlayEvent &&
      other.id == id &&
      other.userId == userId &&
      other.trackId == trackId &&
      other.playedAt == playedAt &&
      other.createdAt == createdAt;

    @override
    int get hashCode =>
        id.hashCode +
        userId.hashCode +
        trackId.hashCode +
        playedAt.hashCode +
        createdAt.hashCode;

  factory PlayEvent.fromJson(Map<String, dynamic> json) => _$PlayEventFromJson(json);

  Map<String, dynamic> toJson() => _$PlayEventToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

