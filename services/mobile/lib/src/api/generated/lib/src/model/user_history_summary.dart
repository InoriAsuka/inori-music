//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/track_play_count.dart';
import 'package:inori_api/src/model/user_history_stats.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'user_history_summary.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UserHistorySummary {
  /// Returns a new [UserHistorySummary] instance.
  UserHistorySummary({

    required  this.stats,

    required  this.topTracks,
  });

  @JsonKey(
    
    name: r'stats',
    required: true,
    includeIfNull: false,
  )


  final UserHistoryStats stats;



  @JsonKey(
    
    name: r'topTracks',
    required: true,
    includeIfNull: false,
  )


  final List<TrackPlayCount> topTracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UserHistorySummary &&
      other.stats == stats &&
      other.topTracks == topTracks;

    @override
    int get hashCode =>
        stats.hashCode +
        topTracks.hashCode;

  factory UserHistorySummary.fromJson(Map<String, dynamic> json) => _$UserHistorySummaryFromJson(json);

  Map<String, dynamic> toJson() => _$UserHistorySummaryToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

