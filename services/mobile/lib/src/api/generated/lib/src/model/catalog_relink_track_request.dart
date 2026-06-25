//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_relink_track_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogRelinkTrackRequest {
  /// Returns a new [CatalogRelinkTrackRequest] instance.
  CatalogRelinkTrackRequest({

    required  this.mediaObjectId,
  });

      /// ID of the new media object. Must be active and of kind original_audio or transcoded_audio.
  @JsonKey(
    
    name: r'mediaObjectId',
    required: true,
    includeIfNull: false,
  )


  final String mediaObjectId;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogRelinkTrackRequest &&
      other.mediaObjectId == mediaObjectId;

    @override
    int get hashCode =>
        mediaObjectId.hashCode;

  factory CatalogRelinkTrackRequest.fromJson(Map<String, dynamic> json) => _$CatalogRelinkTrackRequestFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogRelinkTrackRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

