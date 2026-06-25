//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/recent_catalog_item.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'recent_catalog_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class RecentCatalogResult {
  /// Returns a new [RecentCatalogResult] instance.
  RecentCatalogResult({

    required  this.items,
  });

      /// Recent artist, album, track, and playlist entries ordered newest-first.
  @JsonKey(
    
    name: r'items',
    required: true,
    includeIfNull: false,
  )


  final List<RecentCatalogItem> items;





    @override
    bool operator ==(Object other) => identical(this, other) || other is RecentCatalogResult &&
      other.items == items;

    @override
    int get hashCode =>
        items.hashCode;

  factory RecentCatalogResult.fromJson(Map<String, dynamic> json) => _$RecentCatalogResultFromJson(json);

  Map<String, dynamic> toJson() => _$RecentCatalogResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

