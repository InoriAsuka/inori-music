//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'track_play_count.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class TrackPlayCount {
  /// Returns a new [TrackPlayCount] instance.
  TrackPlayCount({

    required  this.trackId,

    required  this.playCount,
  });

  @JsonKey(
    
    name: r'trackId',
    required: true,
    includeIfNull: false,
  )


  final String trackId;



  @JsonKey(
    
    name: r'playCount',
    required: true,
    includeIfNull: false,
  )


  final int playCount;





    @override
    bool operator ==(Object other) => identical(this, other) || other is TrackPlayCount &&
      other.trackId == trackId &&
      other.playCount == playCount;

    @override
    int get hashCode =>
        trackId.hashCode +
        playCount.hashCode;

  factory TrackPlayCount.fromJson(Map<String, dynamic> json) => _$TrackPlayCountFromJson(json);

  Map<String, dynamic> toJson() => _$TrackPlayCountToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

