//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/track_play_count.dart';
import 'package:inori_api/src/model/history_stats.dart';
import 'package:inori_api/src/model/user_play_count.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'global_history_summary.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class GlobalHistorySummary {
  /// Returns a new [GlobalHistorySummary] instance.
  GlobalHistorySummary({

    required  this.stats,

    required  this.topTracks,

    required  this.topUsers,
  });

  @JsonKey(
    
    name: r'stats',
    required: true,
    includeIfNull: false,
  )


  final HistoryStats stats;



  @JsonKey(
    
    name: r'topTracks',
    required: true,
    includeIfNull: false,
  )


  final List<TrackPlayCount> topTracks;



  @JsonKey(
    
    name: r'topUsers',
    required: true,
    includeIfNull: false,
  )


  final List<UserPlayCount> topUsers;





    @override
    bool operator ==(Object other) => identical(this, other) || other is GlobalHistorySummary &&
      other.stats == stats &&
      other.topTracks == topTracks &&
      other.topUsers == topUsers;

    @override
    int get hashCode =>
        stats.hashCode +
        topTracks.hashCode +
        topUsers.hashCode;

  factory GlobalHistorySummary.fromJson(Map<String, dynamic> json) => _$GlobalHistorySummaryFromJson(json);

  Map<String, dynamic> toJson() => _$GlobalHistorySummaryToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

