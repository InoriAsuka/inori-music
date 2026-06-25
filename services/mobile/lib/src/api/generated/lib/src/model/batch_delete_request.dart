//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'batch_delete_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class BatchDeleteRequest {
  /// Returns a new [BatchDeleteRequest] instance.
  BatchDeleteRequest({

    required  this.ids,
  });

      /// List of play event IDs to delete (1–100 entries).
  @JsonKey(
    
    name: r'ids',
    required: true,
    includeIfNull: false,
  )


  final List<String> ids;





    @override
    bool operator ==(Object other) => identical(this, other) || other is BatchDeleteRequest &&
      other.ids == ids;

    @override
    int get hashCode =>
        ids.hashCode;

  factory BatchDeleteRequest.fromJson(Map<String, dynamic> json) => _$BatchDeleteRequestFromJson(json);

  Map<String, dynamic> toJson() => _$BatchDeleteRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

