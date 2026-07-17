import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/api/api_client.dart';

// ---------------------------------------------------------------------------
// Hand-written "me" API client (v5.4.0 cross-device sync).
//
// The generated inori_api MeApi only covers /me/sessions. The cross-device
// player-state and search-history endpoints are consumed here through the same
// Dio instance (auth + baseUrl interceptor in api_client.dart) rather than
// regenerating the OpenAPI client. Wire shapes mirror the server:
//   * player-state  -> internal/httpapi/handler_playerstate.go (playerStateRequest / PlayerState)
//   * search-history -> internal/searchhistory (Service works on []string queries)
// ---------------------------------------------------------------------------

/// Cross-device playback snapshot returned by `GET /api/v1/me/player-state`
/// and echoed back (with server `updatedAt`) by the PUT.
class PlayerStateDto {
  const PlayerStateDto({
    required this.queue,
    required this.currentIndex,
    required this.positionSeconds,
    required this.repeat,
    required this.shuffle,
    required this.volume,
    required this.speed,
    required this.status,
    this.updatedAt,
  });

  final List<String> queue;
  final int currentIndex;
  final double positionSeconds;

  /// Repeat mode as understood by the server: "off" | "one" | "all".
  final String repeat;
  final bool shuffle;
  final double volume;
  final double speed;

  /// Transport state: "playing" | "paused" | "stopped".
  final String status;

  /// Server-assigned last-write-wins clock. Null on locally-built payloads.
  final DateTime? updatedAt;

  Map<String, dynamic> toJson() => {
        'queue': queue,
        'currentIndex': currentIndex,
        'positionSeconds': positionSeconds,
        'repeat': repeat,
        'shuffle': shuffle,
        'volume': volume,
        'speed': speed,
        'status': status,
      };

  factory PlayerStateDto.fromJson(Map<String, dynamic> json) {
    final rawUpdated = json['updatedAt'];
    return PlayerStateDto(
      queue: (json['queue'] as List<dynamic>?)?.cast<String>() ?? const [],
      currentIndex: (json['currentIndex'] as num?)?.toInt() ?? -1,
      positionSeconds: (json['positionSeconds'] as num?)?.toDouble() ?? 0,
      repeat: (json['repeat'] as String?) ?? 'off',
      shuffle: (json['shuffle'] as bool?) ?? false,
      volume: (json['volume'] as num?)?.toDouble() ?? 1.0,
      speed: (json['speed'] as num?)?.toDouble() ?? 1.0,
      status: (json['status'] as String?) ?? 'stopped',
      updatedAt: rawUpdated is String ? DateTime.tryParse(rawUpdated) : null,
    );
  }
}

class MeApi {
  const MeApi(this._dio);

  final Dio _dio;

  // ---- player state ---------------------------------------------------------

  /// Fetch the cross-device player state, or null when the server has never
  /// recorded one (HTTP 404).
  Future<PlayerStateDto?> getPlayerState({CancelToken? cancelToken}) async {
    try {
      final resp = await _dio.get<Map<String, dynamic>>(
        '/api/v1/me/player-state',
        cancelToken: cancelToken,
      );
      final data = resp.data;
      return data == null ? null : PlayerStateDto.fromJson(data);
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return null;
      rethrow;
    }
  }

  /// Upsert the player state (last-write-wins). Returns the stored snapshot,
  /// which carries the server-assigned `updatedAt`.
  Future<PlayerStateDto?> putPlayerState(
    PlayerStateDto state, {
    CancelToken? cancelToken,
  }) async {
    final resp = await _dio.put<Map<String, dynamic>>(
      '/api/v1/me/player-state',
      data: state.toJson(),
      cancelToken: cancelToken,
    );
    final data = resp.data;
    return data == null ? null : PlayerStateDto.fromJson(data);
  }

  // ---- search history -------------------------------------------------------

  /// Fetch the user's recent search queries, newest first (server caps at 20).
  ///
  /// Parses defensively: the server may return a bare array or an envelope
  /// keyed by `queries` / `history` / `entries` / `items`, whose elements are
  /// either plain strings or `{ "query": ... }` objects.
  Future<List<String>> getSearchHistory({CancelToken? cancelToken}) async {
    final resp = await _dio.get<dynamic>(
      '/api/v1/me/search-history',
      cancelToken: cancelToken,
    );
    return _parseQueries(resp.data);
  }

  /// Replace the user's entire search history (server trims to 20, newest
  /// first). An empty list clears it.
  Future<void> putSearchHistory(
    List<String> queries, {
    CancelToken? cancelToken,
  }) async {
    await _dio.put<dynamic>(
      '/api/v1/me/search-history',
      data: {'queries': queries},
      cancelToken: cancelToken,
    );
  }

  /// Clear the user's search history on the server.
  Future<void> deleteSearchHistory({CancelToken? cancelToken}) async {
    await _dio.delete<dynamic>(
      '/api/v1/me/search-history',
      cancelToken: cancelToken,
    );
  }

  static List<String> _parseQueries(dynamic data) {
    List<dynamic>? list;
    if (data is List) {
      list = data;
    } else if (data is Map) {
      final map = data.cast<String, dynamic>();
      for (final key in const ['queries', 'history', 'entries', 'items']) {
        final v = map[key];
        if (v is List) {
          list = v;
          break;
        }
      }
    }
    if (list == null) return const [];
    final out = <String>[];
    for (final e in list) {
      if (e is String) {
        if (e.trim().isNotEmpty) out.add(e);
      } else if (e is Map) {
        final q = e['query'];
        if (q is String && q.trim().isNotEmpty) out.add(q);
      }
    }
    return out;
  }
}

/// Cross-device sync API keyed off the shared authenticated Dio client.
final meApiProvider = Provider<MeApi>((ref) => MeApi(ref.read(dioProvider)));
