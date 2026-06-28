import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/api/api_client.dart';

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

class UserPlaylist {
  const UserPlaylist({
    required this.id,
    required this.userId,
    required this.name,
    required this.description,
    required this.trackIds,
  });

  final String id;
  final String userId;
  final String name;
  final String description;
  final List<String> trackIds;

  factory UserPlaylist.fromJson(Map<String, dynamic> json) {
    return UserPlaylist(
      id: json['id'] as String,
      userId: json['userId'] as String,
      name: json['name'] as String,
      description: (json['description'] as String?) ?? '',
      trackIds: (json['trackIds'] as List<dynamic>?)
              ?.cast<String>() ??
          [],
    );
  }
}

// ---------------------------------------------------------------------------
// Notifier
// ---------------------------------------------------------------------------

class UserPlaylistNotifier extends AsyncNotifier<List<UserPlaylist>> {
  late Dio _dio;

  @override
  Future<List<UserPlaylist>> build() async {
    _dio = ref.read(dioProvider);
    return _fetchPlaylists();
  }

  Future<List<UserPlaylist>> _fetchPlaylists() async {
    final resp = await _dio.get<Map<String, dynamic>>('/api/v1/me/playlists');
    final data = resp.data;
    if (data == null) return [];
    final list = (data['playlists'] as List<dynamic>?) ?? [];
    return list
        .cast<Map<String, dynamic>>()
        .map(UserPlaylist.fromJson)
        .toList();
  }

  Future<void> load() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(_fetchPlaylists);
  }

  Future<UserPlaylist?> create(String name, {String description = ''}) async {
    try {
      final resp = await _dio.post<Map<String, dynamic>>(
        '/api/v1/me/playlists',
        data: {'name': name, 'description': description},
      );
      final pl = UserPlaylist.fromJson(resp.data!);
      state = AsyncData([pl, ...state.valueOrNull ?? []]);
      return pl;
    } catch (_) {
      return null;
    }
  }

  Future<void> delete(String id) async {
    await _dio.delete('/api/v1/me/playlists/$id');
    state = AsyncData(
      (state.valueOrNull ?? []).where((p) => p.id != id).toList(),
    );
  }

  Future<void> addTrack(String playlistId, String trackId) async {
    await _dio.post<void>(
      '/api/v1/me/playlists/$playlistId/tracks',
      data: {'trackId': trackId},
    );
    // Refresh to get updated trackIds
    await load();
  }

  Future<void> removeTrack(String playlistId, String trackId) async {
    await _dio.delete('/api/v1/me/playlists/$playlistId/tracks/$trackId');
    await load();
  }

  Future<List<String>> getTrackIds(String playlistId) async {
    final resp =
        await _dio.get<Map<String, dynamic>>('/api/v1/me/playlists/$playlistId/tracks');
    final data = resp.data;
    if (data == null) return [];
    return ((data['trackIds'] as List<dynamic>?) ?? []).cast<String>();
  }

  Future<UserPlaylist?> rename(String id, String newName) async {
    try {
      final resp = await _dio.patch<Map<String, dynamic>>(
        '/api/v1/me/playlists/$id',
        data: {'name': newName},
      );
      final updated = UserPlaylist.fromJson(resp.data!);
      state = AsyncData(
        (state.valueOrNull ?? []).map((p) => p.id == id ? updated : p).toList(),
      );
      return updated;
    } catch (_) {
      return null;
    }
  }
}

final userPlaylistProvider =
    AsyncNotifierProvider<UserPlaylistNotifier, List<UserPlaylist>>(
  UserPlaylistNotifier.new,
);
