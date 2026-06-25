//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'track_history_stats.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class TrackHistoryStats {
  /// Returns a new [TrackHistoryStats] instance.
  TrackHistoryStats({

    required  this.totalEvents,

    required  this.uniqueListeners,
  });

      /// Total play events recorded for this track.
  @JsonKey(
    
    name: r'totalEvents',
    required: true,
    includeIfNull: false,
  )


  final int totalEvents;



      /// Number of distinct users who have played this track.
  @JsonKey(
    
    name: r'uniqueListeners',
    required: true,
    includeIfNull: false,
  )


  final int uniqueListeners;





    @override
    bool operator ==(Object other) => identical(this, other) || other is TrackHistoryStats &&
      other.totalEvents == totalEvents &&
      other.uniqueListeners == uniqueListeners;

    @override
    int get hashCode =>
        totalEvents.hashCode +
        uniqueListeners.hashCode;

  factory TrackHistoryStats.fromJson(Map<String, dynamic> json) => _$TrackHistoryStatsFromJson(json);

  Map<String, dynamic> toJson() => _$TrackHistoryStatsToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

