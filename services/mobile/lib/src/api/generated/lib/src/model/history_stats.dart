//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'history_stats.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class HistoryStats {
  /// Returns a new [HistoryStats] instance.
  HistoryStats({

    required  this.totalEvents,

    required  this.uniqueUsers,

    required  this.uniqueTracks,
  });

  @JsonKey(
    
    name: r'totalEvents',
    required: true,
    includeIfNull: false,
  )


  final int totalEvents;



  @JsonKey(
    
    name: r'uniqueUsers',
    required: true,
    includeIfNull: false,
  )


  final int uniqueUsers;



  @JsonKey(
    
    name: r'uniqueTracks',
    required: true,
    includeIfNull: false,
  )


  final int uniqueTracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is HistoryStats &&
      other.totalEvents == totalEvents &&
      other.uniqueUsers == uniqueUsers &&
      other.uniqueTracks == uniqueTracks;

    @override
    int get hashCode =>
        totalEvents.hashCode +
        uniqueUsers.hashCode +
        uniqueTracks.hashCode;

  factory HistoryStats.fromJson(Map<String, dynamic> json) => _$HistoryStatsFromJson(json);

  Map<String, dynamic> toJson() => _$HistoryStatsToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

