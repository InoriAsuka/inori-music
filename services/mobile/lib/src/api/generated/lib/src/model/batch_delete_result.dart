//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'batch_delete_result.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class BatchDeleteResult {
  /// Returns a new [BatchDeleteResult] instance.
  BatchDeleteResult({

    required  this.deleted,
  });

      /// Number of play events actually deleted.
  @JsonKey(
    
    name: r'deleted',
    required: true,
    includeIfNull: false,
  )


  final int deleted;





    @override
    bool operator ==(Object other) => identical(this, other) || other is BatchDeleteResult &&
      other.deleted == deleted;

    @override
    int get hashCode =>
        deleted.hashCode;

  factory BatchDeleteResult.fromJson(Map<String, dynamic> json) => _$BatchDeleteResultFromJson(json);

  Map<String, dynamic> toJson() => _$BatchDeleteResultToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

