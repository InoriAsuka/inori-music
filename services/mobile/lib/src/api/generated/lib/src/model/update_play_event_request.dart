//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'update_play_event_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class UpdatePlayEventRequest {
  /// Returns a new [UpdatePlayEventRequest] instance.
  UpdatePlayEventRequest({

    required  this.playedAt,
  });

      /// New client-reported play timestamp (RFC3339).
  @JsonKey(
    
    name: r'playedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime playedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is UpdatePlayEventRequest &&
      other.playedAt == playedAt;

    @override
    int get hashCode =>
        playedAt.hashCode;

  factory UpdatePlayEventRequest.fromJson(Map<String, dynamic> json) => _$UpdatePlayEventRequestFromJson(json);

  Map<String, dynamic> toJson() => _$UpdatePlayEventRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

