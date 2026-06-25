//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'user_track_stats.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UserTrackStats {
  /// Returns a new [UserTrackStats] instance.
  UserTrackStats({

    required  this.trackId,

    required  this.totalPlays,

     this.firstPlayedAt,

     this.lastPlayedAt,
  });

  @JsonKey(
    
    name: r'trackId',
    required: true,
    includeIfNull: false,
  )


  final String trackId;



          // minimum: 0
  @JsonKey(
    
    name: r'totalPlays',
    required: true,
    includeIfNull: false,
  )


  final int totalPlays;



  @JsonKey(
    
    name: r'firstPlayedAt',
    required: false,
    includeIfNull: false,
  )


  final DateTime? firstPlayedAt;



  @JsonKey(
    
    name: r'lastPlayedAt',
    required: false,
    includeIfNull: false,
  )


  final DateTime? lastPlayedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UserTrackStats &&
      other.trackId == trackId &&
      other.totalPlays == totalPlays &&
      other.firstPlayedAt == firstPlayedAt &&
      other.lastPlayedAt == lastPlayedAt;

    @override
    int get hashCode =>
        trackId.hashCode +
        totalPlays.hashCode +
        firstPlayedAt.hashCode +
        lastPlayedAt.hashCode;

  factory UserTrackStats.fromJson(Map<String, dynamic> json) => _$UserTrackStatsFromJson(json);

  Map<String, dynamic> toJson() => _$UserTrackStatsToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

