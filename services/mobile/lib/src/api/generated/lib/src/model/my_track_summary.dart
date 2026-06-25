//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/track_play_count.dart';
import 'package:inori_api/src/model/user_track_stats.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'my_track_summary.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class MyTrackSummary {
  /// Returns a new [MyTrackSummary] instance.
  MyTrackSummary({

    required  this.stats,

    required  this.recentTracks,
  });

  @JsonKey(
    
    name: r'stats',
    required: true,
    includeIfNull: false,
  )


  final UserTrackStats stats;



  @JsonKey(
    
    name: r'recentTracks',
    required: true,
    includeIfNull: false,
  )


  final List<TrackPlayCount> recentTracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is MyTrackSummary &&
      other.stats == stats &&
      other.recentTracks == recentTracks;

    @override
    int get hashCode =>
        stats.hashCode +
        recentTracks.hashCode;

  factory MyTrackSummary.fromJson(Map<String, dynamic> json) => _$MyTrackSummaryFromJson(json);

  Map<String, dynamic> toJson() => _$MyTrackSummaryToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

