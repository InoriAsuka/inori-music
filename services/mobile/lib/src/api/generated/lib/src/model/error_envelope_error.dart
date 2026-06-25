//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'error_envelope_error.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class ErrorEnvelopeError {
  /// Returns a new [ErrorEnvelopeError] instance.
  ErrorEnvelopeError({

    required  this.code,

    required  this.message,
  });

  @JsonKey(
    
    name: r'code',
    required: true,
    includeIfNull: false,
  )


  final ErrorEnvelopeErrorCodeEnum code;



  @JsonKey(
    
    name: r'message',
    required: true,
    includeIfNull: false,
  )


  final String message;





    @override
    bool operator ==(Object other) => identical(this, other) || other is ErrorEnvelopeError &&
      other.code == code &&
      other.message == message;

    @override
    int get hashCode =>
        code.hashCode +
        message.hashCode;

  factory ErrorEnvelopeError.fromJson(Map<String, dynamic> json) => _$ErrorEnvelopeErrorFromJson(json);

  Map<String, dynamic> toJson() => _$ErrorEnvelopeErrorToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}


enum ErrorEnvelopeErrorCodeEnum {
@JsonValue(r'invalid_backend')
invalidBackend(r'invalid_backend'),
@JsonValue(r'unauthorized')
unauthorized(r'unauthorized'),
@JsonValue(r'not_found')
notFound(r'not_found'),
@JsonValue(r'method_not_allowed')
methodNotAllowed(r'method_not_allowed'),
@JsonValue(r'conflict')
conflict(r'conflict'),
@JsonValue(r'probe_unsupported')
probeUnsupported(r'probe_unsupported'),
@JsonValue(r'probe_failed')
probeFailed(r'probe_failed'),
@JsonValue(r'capacity_unsupported')
capacityUnsupported(r'capacity_unsupported'),
@JsonValue(r'internal_error')
internalError(r'internal_error'),
@JsonValue(r'admin_auth_not_configured')
adminAuthNotConfigured(r'admin_auth_not_configured'),
@JsonValue(r'invalid_media_object')
invalidMediaObject(r'invalid_media_object'),
@JsonValue(r'media_registry_not_configured')
mediaRegistryNotConfigured(r'media_registry_not_configured'),
@JsonValue(r'media_object_verification_unsupported')
mediaObjectVerificationUnsupported(r'media_object_verification_unsupported'),
@JsonValue(r'media_object_verification_failed')
mediaObjectVerificationFailed(r'media_object_verification_failed'),
@JsonValue(r'auth_not_configured')
authNotConfigured(r'auth_not_configured'),
@JsonValue(r'invalid_user')
invalidUser(r'invalid_user'),
@JsonValue(r'user_disabled')
userDisabled(r'user_disabled'),
@JsonValue(r'missing_query')
missingQuery(r'missing_query'),
@JsonValue(r'catalog_not_configured')
catalogNotConfigured(r'catalog_not_configured'),
@JsonValue(r'invalid_catalog_entity')
invalidCatalogEntity(r'invalid_catalog_entity'),
@JsonValue(r'import_rejected')
importRejected(r'import_rejected'),
@JsonValue(r'relink_rejected')
relinkRejected(r'relink_rejected'),
@JsonValue(r'validation_error')
validationError(r'validation_error'),
@JsonValue(r'invalid_limit')
invalidLimit(r'invalid_limit'),
@JsonValue(r'playback_unavailable')
playbackUnavailable(r'playback_unavailable'),
@JsonValue(r'invalid_offset')
invalidOffset(r'invalid_offset'),
@JsonValue(r'invalid_sort_order')
invalidSortOrder(r'invalid_sort_order'),
@JsonValue(r'history_not_configured')
historyNotConfigured(r'history_not_configured'),
@JsonValue(r'invalid_since')
invalidSince(r'invalid_since'),
@JsonValue(r'invalid_until')
invalidUntil(r'invalid_until'),
@JsonValue(r'invalid_time_range')
invalidTimeRange(r'invalid_time_range'),
@JsonValue(r'missing_time_filter')
missingTimeFilter(r'missing_time_filter'),
@JsonValue(r'invalid_order')
invalidOrder(r'invalid_order'),
@JsonValue(r'event_forbidden')
eventForbidden(r'event_forbidden'),
@JsonValue(r'invalid_played_at')
invalidPlayedAt(r'invalid_played_at'),
@JsonValue(r'invalid_ids')
invalidIds(r'invalid_ids'),
@JsonValue(r'missing_time_bounds')
missingTimeBounds(r'missing_time_bounds'),
@JsonValue(r'invalid_granularity')
invalidGranularity(r'invalid_granularity'),
@JsonValue(r'storage_backend_is_default')
storageBackendIsDefault(r'storage_backend_is_default'),
@JsonValue(r'storage_backend_in_use')
storageBackendInUse(r'storage_backend_in_use'),
@JsonValue(r'favorites_not_configured')
favoritesNotConfigured(r'favorites_not_configured');

const ErrorEnvelopeErrorCodeEnum(this.value);

final String value;

@override
String toString() => value;
}


