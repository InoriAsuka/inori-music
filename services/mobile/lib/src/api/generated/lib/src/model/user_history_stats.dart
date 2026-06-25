//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'user_history_stats.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UserHistoryStats {
  /// Returns a new [UserHistoryStats] instance.
  UserHistoryStats({

    required  this.totalEvents,

    required  this.uniqueTracks,
  });

      /// Total play events recorded by this user.
  @JsonKey(
    
    name: r'totalEvents',
    required: true,
    includeIfNull: false,
  )


  final int totalEvents;



      /// Number of distinct tracks played by this user.
  @JsonKey(
    
    name: r'uniqueTracks',
    required: true,
    includeIfNull: false,
  )


  final int uniqueTracks;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UserHistoryStats &&
      other.totalEvents == totalEvents &&
      other.uniqueTracks == uniqueTracks;

    @override
    int get hashCode =>
        totalEvents.hashCode +
        uniqueTracks.hashCode;

  factory UserHistoryStats.fromJson(Map<String, dynamic> json) => _$UserHistoryStatsFromJson(json);

  Map<String, dynamic> toJson() => _$UserHistoryStatsToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

