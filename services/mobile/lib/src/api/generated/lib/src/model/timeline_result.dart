//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/timeline_bucket.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'timeline_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class TimelineResult {
  /// Returns a new [TimelineResult] instance.
  TimelineResult({

    required  this.buckets,
  });

      /// Ordered list of time buckets (oldest first).
  @JsonKey(
    
    name: r'buckets',
    required: true,
    includeIfNull: false,
  )


  final List<TimelineBucket> buckets;





    @override
    bool operator ==(Object other) => identical(this, other) || other is TimelineResult &&
      other.buckets == buckets;

    @override
    int get hashCode =>
        buckets.hashCode;

  factory TimelineResult.fromJson(Map<String, dynamic> json) => _$TimelineResultFromJson(json);

  Map<String, dynamic> toJson() => _$TimelineResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

