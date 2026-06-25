//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

import 'dart:async';

// ignore: unused_import
import 'dart:convert';
import 'package:inori_api/src/deserialize.dart';
import 'package:dio/dio.dart';

import 'package:inori_api/src/model/error_envelope.dart';
import 'package:inori_api/src/model/history_stats.dart';
import 'package:inori_api/src/model/timeline_result.dart';
import 'package:inori_api/src/model/top_tracks_result.dart';
import 'package:inori_api/src/model/top_users_result.dart';
import 'package:inori_api/src/model/track_history_stats.dart';
import 'package:inori_api/src/model/user_history_stats.dart';

class AdminHistoryApi {

  final Dio _dio;

  const AdminHistoryApi(this._dio);

  /// Get history aggregate stats
  /// Returns system-wide playback aggregate counts (admin only).
  ///
  /// Parameters:
  /// * [since] - Restrict results to events on or after this RFC3339 timestamp.
  /// * [until] - Restrict results to events before this RFC3339 timestamp (exclusive).
  /// * [cancelToken] - A [CancelToken] that can be used to cancel the operation
  /// * [headers] - Can be used to add additional headers to the request
  /// * [extras] - Can be used to add flags to the request
  /// * [validateStatus] - A [ValidateStatus] callback that can be used to determine request success based on the HTTP status of the response
  /// * [onSendProgress] - A [ProgressCallback] that can be used to get the send progress
  /// * [onReceiveProgress] - A [ProgressCallback] that can be used to get the receive progress
  ///
  /// Returns a [Future] containing a [Response] with a [HistoryStats] as data
  /// Throws [DioException] if API call or serialization fails
  Future<Response<HistoryStats>> getAdminHistoryStats({ 
    DateTime? since,
    DateTime? until,
    CancelToken? cancelToken,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
    ValidateStatus? validateStatus,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    final _path = r'/api/v1/admin/history/stats';
    final _options = Options(
      method: r'GET',
      headers: <String, dynamic>{
        ...?headers,
      },
      extra: <String, dynamic>{
        'secure': <Map<String, String>>[
          {
            'type': 'http',
            'scheme': 'bearer',
            'name': 'bearerAuth',
          },
        ],
        ...?extra,
      },
      validateStatus: validateStatus,
    );

    final _queryParameters = <String, dynamic>{
      if (since != null) r'since': since,
      if (until != null) r'until': until,
    };

    final _response = await _dio.request<Object>(
      _path,
      options: _options,
      queryParameters: _queryParameters,
      cancelToken: cancelToken,
      onSendProgress: onSendProgress,
      onReceiveProgress: onReceiveProgress,
    );

    HistoryStats? _responseData;

    try {
final rawData = _response.data;
_responseData = rawData == null ? null : deserialize<HistoryStats, HistoryStats>(rawData, 'HistoryStats', growable: true);

    } catch (error, stackTrace) {
      throw DioException(
        requestOptions: _response.requestOptions,
        response: _response,
        type: DioExceptionType.unknown,
        error: error,
        stackTrace: stackTrace,
      );
    }

    return Response<HistoryStats>(
      data: _responseData,
      headers: _response.headers,
      isRedirect: _response.isRedirect,
      requestOptions: _response.requestOptions,
      redirects: _response.redirects,
      statusCode: _response.statusCode,
      statusMessage: _response.statusMessage,
      extra: _response.extra,
    );
  }

  /// Get play history grouped by time bucket (admin)
  /// Returns play event counts grouped by day, week, or month within the specified time window. Both since and until are required. Admin only.
  ///
  /// Parameters:
  /// * [since] - Start of the time window (RFC3339, inclusive).
  /// * [until] - End of the time window (RFC3339, exclusive).
  /// * [granularity] - Bucket size: day (default), week, or month.
  /// * [userId] - Optional: restrict to a specific user's events.
  /// * [trackId] - Optional: restrict to events for a specific track.
  /// * [cancelToken] - A [CancelToken] that can be used to cancel the operation
  /// * [headers] - Can be used to add additional headers to the request
  /// * [extras] - Can be used to add flags to the request
  /// * [validateStatus] - A [ValidateStatus] callback that can be used to determine request success based on the HTTP status of the response
  /// * [onSendProgress] - A [ProgressCallback] that can be used to get the send progress
  /// * [onReceiveProgress] - A [ProgressCallback] that can be used to get the receive progress
  ///
  /// Returns a [Future] containing a [Response] with a [TimelineResult] as data
  /// Throws [DioException] if API call or serialization fails
  Future<Response<TimelineResult>> getAdminHistoryTimeline({ 
    required DateTime since,
    required DateTime until,
    String? granularity = 'day',
    String? userId,
    String? trackId,
    CancelToken? cancelToken,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
    ValidateStatus? validateStatus,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    final _path = r'/api/v1/admin/history/timeline';
    final _options = Options(
      method: r'GET',
      headers: <String, dynamic>{
        ...?headers,
      },
      extra: <String, dynamic>{
        'secure': <Map<String, String>>[
          {
            'type': 'http',
            'scheme': 'bearer',
            'name': 'bearerAuth',
          },
        ],
        ...?extra,
      },
      validateStatus: validateStatus,
    );

    final _queryParameters = <String, dynamic>{
      r'since': since,
      r'until': until,
      if (granularity != null) r'granularity': granularity,
      if (userId != null) r'userId': userId,
      if (trackId != null) r'trackId': trackId,
    };

    final _response = await _dio.request<Object>(
      _path,
      options: _options,
      queryParameters: _queryParameters,
      cancelToken: cancelToken,
      onSendProgress: onSendProgress,
      onReceiveProgress: onReceiveProgress,
    );

    TimelineResult? _responseData;

    try {
final rawData = _response.data;
_responseData = rawData == null ? null : deserialize<TimelineResult, TimelineResult>(rawData, 'TimelineResult', growable: true);

    } catch (error, stackTrace) {
      throw DioException(
        requestOptions: _response.requestOptions,
        response: _response,
        type: DioExceptionType.unknown,
        error: error,
        stackTrace: stackTrace,
      );
    }

    return Response<TimelineResult>(
      data: _responseData,
      headers: _response.headers,
      isRedirect: _response.isRedirect,
      requestOptions: _response.requestOptions,
      redirects: _response.redirects,
      statusCode: _response.statusCode,
      statusMessage: _response.statusMessage,
      extra: _response.extra,
    );
  }

  /// Get most-played tracks
  /// Returns tracks with the highest total play count across all users (admin only).
  ///
  /// Parameters:
  /// * [limit] - Maximum number of results (default 10, max 100).
  /// * [since] - Restrict results to events on or after this RFC3339 timestamp.
  /// * [until] - Restrict results to events before this RFC3339 timestamp (exclusive).
  /// * [cancelToken] - A [CancelToken] that can be used to cancel the operation
  /// * [headers] - Can be used to add additional headers to the request
  /// * [extras] - Can be used to add flags to the request
  /// * [validateStatus] - A [ValidateStatus] callback that can be used to determine request success based on the HTTP status of the response
  /// * [onSendProgress] - A [ProgressCallback] that can be used to get the send progress
  /// * [onReceiveProgress] - A [ProgressCallback] that can be used to get the receive progress
  ///
  /// Returns a [Future] containing a [Response] with a [TopTracksResult] as data
  /// Throws [DioException] if API call or serialization fails
  Future<Response<TopTracksResult>> getAdminTopTracks({ 
    int? limit = 10,
    DateTime? since,
    DateTime? until,
    CancelToken? cancelToken,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
    ValidateStatus? validateStatus,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    final _path = r'/api/v1/admin/history/top-tracks';
    final _options = Options(
      method: r'GET',
      headers: <String, dynamic>{
        ...?headers,
      },
      extra: <String, dynamic>{
        'secure': <Map<String, String>>[
          {
            'type': 'http',
            'scheme': 'bearer',
            'name': 'bearerAuth',
          },
        ],
        ...?extra,
      },
      validateStatus: validateStatus,
    );

    final _queryParameters = <String, dynamic>{
      if (limit != null) r'limit': limit,
      if (since != null) r'since': since,
      if (until != null) r'until': until,
    };

    final _response = await _dio.request<Object>(
      _path,
      options: _options,
      queryParameters: _queryParameters,
      cancelToken: cancelToken,
      onSendProgress: onSendProgress,
      onReceiveProgress: onReceiveProgress,
    );

    TopTracksResult? _responseData;

    try {
final rawData = _response.data;
_responseData = rawData == null ? null : deserialize<TopTracksResult, TopTracksResult>(rawData, 'TopTracksResult', growable: true);

    } catch (error, stackTrace) {
      throw DioException(
        requestOptions: _response.requestOptions,
        response: _response,
        type: DioExceptionType.unknown,
        error: error,
        stackTrace: stackTrace,
      );
    }

    return Response<TopTracksResult>(
      data: _responseData,
      headers: _response.headers,
      isRedirect: _response.isRedirect,
      requestOptions: _response.requestOptions,
      redirects: _response.redirects,
      statusCode: _response.statusCode,
      statusMessage: _response.statusMessage,
      extra: _response.extra,
    );
  }

  /// Get most-active users
  /// Returns users with the highest total play event count (admin only).
  ///
  /// Parameters:
  /// * [limit] - Maximum number of results (default 10, max 100).
  /// * [since] - Restrict results to events on or after this RFC3339 timestamp.
  /// * [until] - Restrict results to events before this RFC3339 timestamp (exclusive).
  /// * [cancelToken] - A [CancelToken] that can be used to cancel the operation
  /// * [headers] - Can be used to add additional headers to the request
  /// * [extras] - Can be used to add flags to the request
  /// * [validateStatus] - A [ValidateStatus] callback that can be used to determine request success based on the HTTP status of the response
  /// * [onSendProgress] - A [ProgressCallback] that can be used to get the send progress
  /// * [onReceiveProgress] - A [ProgressCallback] that can be used to get the receive progress
  ///
  /// Returns a [Future] containing a [Response] with a [TopUsersResult] as data
  /// Throws [DioException] if API call or serialization fails
  Future<Response<TopUsersResult>> getAdminTopUsers({ 
    int? limit = 10,
    DateTime? since,
    DateTime? until,
    CancelToken? cancelToken,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
    ValidateStatus? validateStatus,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    final _path = r'/api/v1/admin/history/top-users';
    final _options = Options(
      method: r'GET',
      headers: <String, dynamic>{
        ...?headers,
      },
      extra: <String, dynamic>{
        'secure': <Map<String, String>>[
          {
            'type': 'http',
            'scheme': 'bearer',
            'name': 'bearerAuth',
          },
        ],
        ...?extra,
      },
      validateStatus: validateStatus,
    );

    final _queryParameters = <String, dynamic>{
      if (limit != null) r'limit': limit,
      if (since != null) r'since': since,
      if (until != null) r'until': until,
    };

    final _response = await _dio.request<Object>(
      _path,
      options: _options,
      queryParameters: _queryParameters,
      cancelToken: cancelToken,
      onSendProgress: onSendProgress,
      onReceiveProgress: onReceiveProgress,
    );

    TopUsersResult? _responseData;

    try {
final rawData = _response.data;
_responseData = rawData == null ? null : deserialize<TopUsersResult, TopUsersResult>(rawData, 'TopUsersResult', growable: true);

    } catch (error, stackTrace) {
      throw DioException(
        requestOptions: _response.requestOptions,
        response: _response,
        type: DioExceptionType.unknown,
        error: error,
        stackTrace: stackTrace,
      );
    }

    return Response<TopUsersResult>(
      data: _responseData,
      headers: _response.headers,
      isRedirect: _response.isRedirect,
      requestOptions: _response.requestOptions,
      redirects: _response.redirects,
      statusCode: _response.statusCode,
      statusMessage: _response.statusMessage,
      extra: _response.extra,
    );
  }

  /// Get play history stats for a specific track (admin)
  /// Returns aggregate play counts (total events, unique listeners) for any track. Admin only.
  ///
  /// Parameters:
  /// * [trackId] - ID of the track whose stats to retrieve.
  /// * [since] - Restrict stats to events on or after this RFC3339 timestamp.
  /// * [until] - Restrict stats to events before this RFC3339 timestamp (exclusive).
  /// * [cancelToken] - A [CancelToken] that can be used to cancel the operation
  /// * [headers] - Can be used to add additional headers to the request
  /// * [extras] - Can be used to add flags to the request
  /// * [validateStatus] - A [ValidateStatus] callback that can be used to determine request success based on the HTTP status of the response
  /// * [onSendProgress] - A [ProgressCallback] that can be used to get the send progress
  /// * [onReceiveProgress] - A [ProgressCallback] that can be used to get the receive progress
  ///
  /// Returns a [Future] containing a [Response] with a [TrackHistoryStats] as data
  /// Throws [DioException] if API call or serialization fails
  Future<Response<TrackHistoryStats>> getAdminTrackStats({ 
    required String trackId,
    DateTime? since,
    DateTime? until,
    CancelToken? cancelToken,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
    ValidateStatus? validateStatus,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    final _path = r'/api/v1/admin/history/tracks/{trackId}/stats'.replaceAll('{' r'trackId' '}', trackId.toString());
    final _options = Options(
      method: r'GET',
      headers: <String, dynamic>{
        ...?headers,
      },
      extra: <String, dynamic>{
        'secure': <Map<String, String>>[
          {
            'type': 'http',
            'scheme': 'bearer',
            'name': 'bearerAuth',
          },
        ],
        ...?extra,
      },
      validateStatus: validateStatus,
    );

    final _queryParameters = <String, dynamic>{
      if (since != null) r'since': since,
      if (until != null) r'until': until,
    };

    final _response = await _dio.request<Object>(
      _path,
      options: _options,
      queryParameters: _queryParameters,
      cancelToken: cancelToken,
      onSendProgress: onSendProgress,
      onReceiveProgress: onReceiveProgress,
    );

    TrackHistoryStats? _responseData;

    try {
final rawData = _response.data;
_responseData = rawData == null ? null : deserialize<TrackHistoryStats, TrackHistoryStats>(rawData, 'TrackHistoryStats', growable: true);

    } catch (error, stackTrace) {
      throw DioException(
        requestOptions: _response.requestOptions,
        response: _response,
        type: DioExceptionType.unknown,
        error: error,
        stackTrace: stackTrace,
      );
    }

    return Response<TrackHistoryStats>(
      data: _responseData,
      headers: _response.headers,
      isRedirect: _response.isRedirect,
      requestOptions: _response.requestOptions,
      redirects: _response.redirects,
      statusCode: _response.statusCode,
      statusMessage: _response.statusMessage,
      extra: _response.extra,
    );
  }

  /// Get top listeners for a specific track (admin)
  /// Returns the users who have played a track the most. Admin only.
  ///
  /// Parameters:
  /// * [trackId] - ID of the track whose top listeners to retrieve.
  /// * [limit] - Maximum number of results (default 10, max 100).
  /// * [since] - Restrict results to events on or after this RFC3339 timestamp.
  /// * [until] - Restrict results to events before this RFC3339 timestamp (exclusive).
  /// * [cancelToken] - A [CancelToken] that can be used to cancel the operation
  /// * [headers] - Can be used to add additional headers to the request
  /// * [extras] - Can be used to add flags to the request
  /// * [validateStatus] - A [ValidateStatus] callback that can be used to determine request success based on the HTTP status of the response
  /// * [onSendProgress] - A [ProgressCallback] that can be used to get the send progress
  /// * [onReceiveProgress] - A [ProgressCallback] that can be used to get the receive progress
  ///
  /// Returns a [Future] containing a [Response] with a [TopUsersResult] as data
  /// Throws [DioException] if API call or serialization fails
  Future<Response<TopUsersResult>> getAdminTrackTopListeners({ 
    required String trackId,
    int? limit = 10,
    DateTime? since,
    DateTime? until,
    CancelToken? cancelToken,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
    ValidateStatus? validateStatus,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    final _path = r'/api/v1/admin/history/tracks/{trackId}/top-listeners'.replaceAll('{' r'trackId' '}', trackId.toString());
    final _options = Options(
      method: r'GET',
      headers: <String, dynamic>{
        ...?headers,
      },
      extra: <String, dynamic>{
        'secure': <Map<String, String>>[
          {
            'type': 'http',
            'scheme': 'bearer',
            'name': 'bearerAuth',
          },
        ],
        ...?extra,
      },
      validateStatus: validateStatus,
    );

    final _queryParameters = <String, dynamic>{
      if (limit != null) r'limit': limit,
      if (since != null) r'since': since,
      if (until != null) r'until': until,
    };

    final _response = await _dio.request<Object>(
      _path,
      options: _options,
      queryParameters: _queryParameters,
      cancelToken: cancelToken,
      onSendProgress: onSendProgress,
      onReceiveProgress: onReceiveProgress,
    );

    TopUsersResult? _responseData;

    try {
final rawData = _response.data;
_responseData = rawData == null ? null : deserialize<TopUsersResult, TopUsersResult>(rawData, 'TopUsersResult', growable: true);

    } catch (error, stackTrace) {
      throw DioException(
        requestOptions: _response.requestOptions,
        response: _response,
        type: DioExceptionType.unknown,
        error: error,
        stackTrace: stackTrace,
      );
    }

    return Response<TopUsersResult>(
      data: _responseData,
      headers: _response.headers,
      isRedirect: _response.isRedirect,
      requestOptions: _response.requestOptions,
      redirects: _response.redirects,
      statusCode: _response.statusCode,
      statusMessage: _response.statusMessage,
      extra: _response.extra,
    );
  }

  /// Get play history stats for a specific user (admin)
  /// Returns aggregate play counts (total events, unique tracks) for any user. Admin only.
  ///
  /// Parameters:
  /// * [userId] - ID of the user whose stats to retrieve.
  /// * [since] - Restrict stats to events on or after this RFC3339 timestamp.
  /// * [until] - Restrict stats to events before this RFC3339 timestamp (exclusive).
  /// * [cancelToken] - A [CancelToken] that can be used to cancel the operation
  /// * [headers] - Can be used to add additional headers to the request
  /// * [extras] - Can be used to add flags to the request
  /// * [validateStatus] - A [ValidateStatus] callback that can be used to determine request success based on the HTTP status of the response
  /// * [onSendProgress] - A [ProgressCallback] that can be used to get the send progress
  /// * [onReceiveProgress] - A [ProgressCallback] that can be used to get the receive progress
  ///
  /// Returns a [Future] containing a [Response] with a [UserHistoryStats] as data
  /// Throws [DioException] if API call or serialization fails
  Future<Response<UserHistoryStats>> getAdminUserStats({ 
    required String userId,
    DateTime? since,
    DateTime? until,
    CancelToken? cancelToken,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
    ValidateStatus? validateStatus,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    final _path = r'/api/v1/admin/history/users/{userId}/stats'.replaceAll('{' r'userId' '}', userId.toString());
    final _options = Options(
      method: r'GET',
      headers: <String, dynamic>{
        ...?headers,
      },
      extra: <String, dynamic>{
        'secure': <Map<String, String>>[
          {
            'type': 'http',
            'scheme': 'bearer',
            'name': 'bearerAuth',
          },
        ],
        ...?extra,
      },
      validateStatus: validateStatus,
    );

    final _queryParameters = <String, dynamic>{
      if (since != null) r'since': since,
      if (until != null) r'until': until,
    };

    final _response = await _dio.request<Object>(
      _path,
      options: _options,
      queryParameters: _queryParameters,
      cancelToken: cancelToken,
      onSendProgress: onSendProgress,
      onReceiveProgress: onReceiveProgress,
    );

    UserHistoryStats? _responseData;

    try {
final rawData = _response.data;
_responseData = rawData == null ? null : deserialize<UserHistoryStats, UserHistoryStats>(rawData, 'UserHistoryStats', growable: true);

    } catch (error, stackTrace) {
      throw DioException(
        requestOptions: _response.requestOptions,
        response: _response,
        type: DioExceptionType.unknown,
        error: error,
        stackTrace: stackTrace,
      );
    }

    return Response<UserHistoryStats>(
      data: _responseData,
      headers: _response.headers,
      isRedirect: _response.isRedirect,
      requestOptions: _response.requestOptions,
      redirects: _response.redirects,
      statusCode: _response.statusCode,
      statusMessage: _response.statusMessage,
      extra: _response.extra,
    );
  }

  /// Get top tracks for a specific user (admin)
  /// Returns the tracks most frequently played by any user. Admin only.
  ///
  /// Parameters:
  /// * [userId] - ID of the user whose top tracks to retrieve.
  /// * [limit] - Maximum number of results (default 10, max 100).
  /// * [since] - Restrict results to events on or after this RFC3339 timestamp.
  /// * [until] - Restrict results to events before this RFC3339 timestamp (exclusive).
  /// * [cancelToken] - A [CancelToken] that can be used to cancel the operation
  /// * [headers] - Can be used to add additional headers to the request
  /// * [extras] - Can be used to add flags to the request
  /// * [validateStatus] - A [ValidateStatus] callback that can be used to determine request success based on the HTTP status of the response
  /// * [onSendProgress] - A [ProgressCallback] that can be used to get the send progress
  /// * [onReceiveProgress] - A [ProgressCallback] that can be used to get the receive progress
  ///
  /// Returns a [Future] containing a [Response] with a [TopTracksResult] as data
  /// Throws [DioException] if API call or serialization fails
  Future<Response<TopTracksResult>> getAdminUserTopTracks({ 
    required String userId,
    int? limit = 10,
    DateTime? since,
    DateTime? until,
    CancelToken? cancelToken,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? extra,
    ValidateStatus? validateStatus,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    final _path = r'/api/v1/admin/history/users/{userId}/top-tracks'.replaceAll('{' r'userId' '}', userId.toString());
    final _options = Options(
      method: r'GET',
      headers: <String, dynamic>{
        ...?headers,
      },
      extra: <String, dynamic>{
        'secure': <Map<String, String>>[
          {
            'type': 'http',
            'scheme': 'bearer',
            'name': 'bearerAuth',
          },
        ],
        ...?extra,
      },
      validateStatus: validateStatus,
    );

    final _queryParameters = <String, dynamic>{
      if (limit != null) r'limit': limit,
      if (since != null) r'since': since,
      if (until != null) r'until': until,
    };

    final _response = await _dio.request<Object>(
      _path,
      options: _options,
      queryParameters: _queryParameters,
      cancelToken: cancelToken,
      onSendProgress: onSendProgress,
      onReceiveProgress: onReceiveProgress,
    );

    TopTracksResult? _responseData;

    try {
final rawData = _response.data;
_responseData = rawData == null ? null : deserialize<TopTracksResult, TopTracksResult>(rawData, 'TopTracksResult', growable: true);

    } catch (error, stackTrace) {
      throw DioException(
        requestOptions: _response.requestOptions,
        response: _response,
        type: DioExceptionType.unknown,
        error: error,
        stackTrace: stackTrace,
      );
    }

    return Response<TopTracksResult>(
      data: _responseData,
      headers: _response.headers,
      isRedirect: _response.isRedirect,
      requestOptions: _response.requestOptions,
      redirects: _response.redirects,
      statusCode: _response.statusCode,
      statusMessage: _response.statusMessage,
      extra: _response.extra,
    );
  }

}
