//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'session_view.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class SessionView {
  /// Returns a new [SessionView] instance.
  SessionView({

    required  this.userId,

    required  this.expiresAt,

    required  this.createdAt,
  });

  @JsonKey(
    
    name: r'userId',
    required: true,
    includeIfNull: false,
  )


  final String userId;



  @JsonKey(
    
    name: r'expiresAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime expiresAt;



  @JsonKey(
    
    name: r'createdAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime createdAt;





    @override
    bool operator ==(Object other) => identical(this, other) || other is SessionView &&
      other.userId == userId &&
      other.expiresAt == expiresAt &&
      other.createdAt == createdAt;

    @override
    int get hashCode =>
        userId.hashCode +
        expiresAt.hashCode +
        createdAt.hashCode;

  factory SessionView.fromJson(Map<String, dynamic> json) => _$SessionViewFromJson(json);

  Map<String, dynamic> toJson() => _$SessionViewToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

