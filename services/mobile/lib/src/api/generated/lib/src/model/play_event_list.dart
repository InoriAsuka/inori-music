//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/catalog_pagination_meta.dart';
import 'package:inori_api/src/model/play_event.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'play_event_list.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class PlayEventList {
  /// Returns a new [PlayEventList] instance.
  PlayEventList({

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


  final CatalogPaginationMeta pagination;





    @override
    bool operator ==(Object other) => identical(this, other) || other is PlayEventList &&
      other.events == events &&
      other.pagination == pagination;

    @override
    int get hashCode =>
        events.hashCode +
        pagination.hashCode;

  factory PlayEventList.fromJson(Map<String, dynamic> json) => _$PlayEventListFromJson(json);

  Map<String, dynamic> toJson() => _$PlayEventListToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

