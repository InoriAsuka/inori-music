//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/track_history_stats.dart';
import 'package:inori_api/src/model/user_play_count.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'track_history_summary.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class TrackHistorySummary {
  /// Returns a new [TrackHistorySummary] instance.
  TrackHistorySummary({

    required  this.stats,

    required  this.topListeners,
  });

  @JsonKey(
    
    name: r'stats',
    required: true,
    includeIfNull: false,
  )


  final TrackHistoryStats stats;



  @JsonKey(
    
    name: r'topListeners',
    required: true,
    includeIfNull: false,
  )


  final List<UserPlayCount> topListeners;





    @override
    bool operator ==(Object other) => identical(this, other) || other is TrackHistorySummary &&
      other.stats == stats &&
      other.topListeners == topListeners;

    @override
    int get hashCode =>
        stats.hashCode +
        topListeners.hashCode;

  factory TrackHistorySummary.fromJson(Map<String, dynamic> json) => _$TrackHistorySummaryFromJson(json);

  Map<String, dynamic> toJson() => _$TrackHistorySummaryToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

