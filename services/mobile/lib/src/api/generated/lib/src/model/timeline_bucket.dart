//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'timeline_bucket.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class TimelineBucket {
  /// Returns a new [TimelineBucket] instance.
  TimelineBucket({

    required  this.bucketStart,

    required  this.eventCount,
  });

      /// UTC start of the bucket (day/week/month boundary).
  @JsonKey(
    
    name: r'bucketStart',
    required: true,
    includeIfNull: false,
  )


  final DateTime bucketStart;



      /// Number of play events in this bucket.
  @JsonKey(
    
    name: r'eventCount',
    required: true,
    includeIfNull: false,
  )


  final int eventCount;





    @override
    bool operator ==(Object other) => identical(this, other) || other is TimelineBucket &&
      other.bucketStart == bucketStart &&
      other.eventCount == eventCount;

    @override
    int get hashCode =>
        bucketStart.hashCode +
        eventCount.hashCode;

  factory TimelineBucket.fromJson(Map<String, dynamic> json) => _$TimelineBucketFromJson(json);

  Map<String, dynamic> toJson() => _$TimelineBucketToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

