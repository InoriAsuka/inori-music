// ignore_for_file: implementation_imports
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/api/catalog_api.dart';
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/catalog_album.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:inori_api/src/model/catalog_search_result.dart';
import 'package:inori_api/src/model/playlist.dart';
import 'package:inori_api/src/model/track_playback_descriptor.dart';

import 'package:inori_music/src/api/api_client.dart';

final catalogApiProvider = Provider<CatalogApi>((ref) {
  return CatalogApi(ref.watch(dioProvider));
});

final catalogRepositoryProvider = Provider<CatalogRepository>((ref) {
  return CatalogRepository(ref.watch(catalogApiProvider));
});

class CatalogRepository {
  CatalogRepository(this._api);
  final CatalogApi _api;

  Future<List<CatalogArtist>> listArtists({int limit = 50, int offset = 0}) async {
    final resp = await _api.listCatalogArtists(limit: limit, offset: offset);
    return resp.data?.artists ?? [];
  }

  Future<List<CatalogAlbum>> listAlbums({
    String? artistId,
    int limit = 50,
    int offset = 0,
  }) async {
    final resp = await _api.listCatalogAlbums(
      artistId: artistId,
      limit: limit,
      offset: offset,
    );
    return resp.data?.albums ?? [];
  }

  Future<List<CatalogTrack>> listTracks({int limit = 50, int offset = 0}) async {
    final resp = await _api.listCatalogTracks(limit: limit, offset: offset);
    return resp.data?.tracks ?? [];
  }

  Future<CatalogArtist> getArtist(String id) async {
    final resp = await _api.getCatalogArtist(id: id);
    return resp.data!;
  }

  Future<CatalogAlbum> getAlbum(String id) async {
    final resp = await _api.getCatalogAlbum(id: id);
    return resp.data!;
  }

  Future<CatalogTrack> getTrack(String id) async {
    final resp = await _api.getCatalogTrack(id: id);
    return resp.data!;
  }

  Future<CatalogSearchResult> search(String q, {int limit = 20}) async {
    final resp = await _api.listCatalogSearch(q: q);
    return resp.data!;
  }

  Future<List<CatalogAlbum>> albumsByArtist(
    String artistId, {
    int limit = 50,
    int offset = 0,
  }) async {
    final resp = await _api.listAlbumsByArtist(
      id: artistId,
      limit: limit,
      offset: offset,
    );
    return resp.data?.albums ?? [];
  }

  Future<List<CatalogTrack>> tracksByAlbum(
    String albumId, {
    int limit = 200,
    int offset = 0,
  }) async {
    // Use general tracks endpoint and filter client-side
    final resp = await _api.listCatalogTracks(limit: limit, offset: offset);
    final all = resp.data?.tracks ?? <CatalogTrack>[];
    return all.where((t) => t.albumId == albumId).toList();
  }

  Future<List<CatalogTrack>> tracksByArtist(
    String artistId, {
    int limit = 200,
    int offset = 0,
  }) async {
    final resp = await _api.listCatalogTracks(limit: limit, offset: offset);
    final all = resp.data?.tracks ?? <CatalogTrack>[];
    return all.where((t) => t.artistId == artistId).toList();
  }

  Future<List<Playlist>> listPlaylists({int limit = 50, int offset = 0}) async {
    final resp = await _api.apiV1CatalogPlaylistsGet(limit: limit, offset: offset);
    return resp.data?.playlists ?? [];
  }

  Future<Playlist> getPlaylist(String id) async {
    final resp = await _api.apiV1CatalogPlaylistsIdGet(id: id);
    return resp.data!;
  }

  Future<List<String>> playlistTrackIds(String id) async {
    final resp = await _api.apiV1CatalogPlaylistsIdGet(id: id);
    return resp.data?.trackIds ?? [];
  }

  Future<TrackPlaybackDescriptor> getPlaybackDescriptor(String trackId) async {
    final resp = await _api.getTrackPlaybackDescriptor(id: trackId);
    return resp.data!;
  }
}
