//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:inori_api/src/model/session_view.dart';
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'get_admin_user_sessions200_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class GetAdminUserSessions200Response {
  /// Returns a new [GetAdminUserSessions200Response] instance.
  GetAdminUserSessions200Response({

    required  this.sessions,

    required  this.count,
  });

  @JsonKey(
    
    name: r'sessions',
    required: true,
    includeIfNull: false,
  )


  final List<SessionView> sessions;



  @JsonKey(
    
    name: r'count',
    required: true,
    includeIfNull: false,
  )


  final int count;





    @override
    bool operator ==(Object other) => identical(this, other) || other is GetAdminUserSessions200Response &&
      other.sessions == sessions &&
      other.count == count;

    @override
    int get hashCode =>
        sessions.hashCode +
        count.hashCode;

  factory GetAdminUserSessions200Response.fromJson(Map<String, dynamic> json) => _$GetAdminUserSessions200ResponseFromJson(json);

  Map<String, dynamic> toJson() => _$GetAdminUserSessions200ResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

