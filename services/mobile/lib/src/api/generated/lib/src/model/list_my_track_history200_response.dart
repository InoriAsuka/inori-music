//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/pagination_metadata.dart';
import 'package:inori_api/src/model/play_event.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'list_my_track_history200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ListMyTrackHistory200Response {
  /// Returns a new [ListMyTrackHistory200Response] instance.
  ListMyTrackHistory200Response({

    required  this.events,

    required  this.pagination,
  });

  @JsonKey(
    
    name: r'events',
    required: true,
    includeIfNull: false,
  )


  final List<PlayEvent> events;



  @JsonKey(
    
    name: r'pagination',
    required: true,
    includeIfNull: false,
  )


  final PaginationMetadata pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ListMyTrackHistory200Response &&
      other.events == events &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        events.hashCode +
        pagination.hashCode;

  factory ListMyTrackHistory200Response.fromJson(Map<String, dynamic> json) => _$ListMyTrackHistory200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ListMyTrackHistory200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

