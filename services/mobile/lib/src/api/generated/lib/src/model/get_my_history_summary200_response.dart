//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/track_play_count.dart';
import 'package:inori_api/src/model/user_history_stats.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'get_my_history_summary200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class GetMyHistorySummary200Response {
  /// Returns a new [GetMyHistorySummary200Response] instance.
  GetMyHistorySummary200Response({

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
    bool operator ==(Object other) => identical(this, other) || other is GetMyHistorySummary200Response &&
      other.stats == stats &&
      other.topTracks == topTracks;

    @override
    int get hashCode =>
        stats.hashCode +
        topTracks.hashCode;

  factory GetMyHistorySummary200Response.fromJson(Map<String, dynamic> json) => _$GetMyHistorySummary200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$GetMyHistorySummary200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

