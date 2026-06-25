//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'record_play_event_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class RecordPlayEventRequest {
  /// Returns a new [RecordPlayEventRequest] instance.
  RecordPlayEventRequest({

    required  this.trackId,

     this.playedAt,
  });

  @JsonKey(
    
    name: r'trackId',
    required: true,
    includeIfNull: false,
  )


  final String trackId;



      /// Client-reported play time; defaults to server time when omitted.
  @JsonKey(
    
    name: r'playedAt',
    required: false,
    includeIfNull: false,
  )


  final DateTime? playedAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is RecordPlayEventRequest &&
      other.trackId == trackId &&
      other.playedAt == playedAt;

    @override
    int get hashCode =>
        trackId.hashCode +
        playedAt.hashCode;

  factory RecordPlayEventRequest.fromJson(Map<String, dynamic> json) => _$RecordPlayEventRequestFromJson(json);

  Map<String, dynamic> toJson() => _$RecordPlayEventRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

