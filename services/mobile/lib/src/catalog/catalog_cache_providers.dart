// ignore_for_file: implementation_imports
import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/catalog_album.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';

// ---------------------------------------------------------------------------
// Artist name cache
//   • Auto-dispose family: entry is dropped when no listeners remain AND the
//     300 s keep-alive window has closed.
//   • keepAlive link is closed after the TTL so the entry can be garbage-
//     collected; a fresh listener will trigger a new fetch.
// ---------------------------------------------------------------------------

final artistNameProvider = AsyncNotifierProvider.autoDispose
    .family<ArtistNameNotifier, String, String>(ArtistNameNotifier.new);

class ArtistNameNotifier
    extends AutoDisposeFamilyAsyncNotifier<String, String> {
  @override
  Future<String> build(String artistId) async {
    // Keep alive for 300 s even when the widget tree detaches temporarily.
    final link = ref.keepAlive();
    final timer = Timer(const Duration(seconds: 300), link.close);
    ref.onDispose(timer.cancel);

    final repo = ref.read(catalogRepositoryProvider);
    final CatalogArtist artist = await repo.getArtist(artistId);
    return artist.name;
  }
}

// ---------------------------------------------------------------------------
// Album title cache — same pattern.
// ---------------------------------------------------------------------------

final albumTitleProvider = AsyncNotifierProvider.autoDispose
    .family<AlbumTitleNotifier, String, String>(AlbumTitleNotifier.new);

class AlbumTitleNotifier
    extends AutoDisposeFamilyAsyncNotifier<String, String> {
  @override
  Future<String> build(String albumId) async {
    final link = ref.keepAlive();
    final timer = Timer(const Duration(seconds: 300), link.close);
    ref.onDispose(timer.cancel);

    final repo = ref.read(catalogRepositoryProvider);
    final CatalogAlbum album = await repo.getAlbum(albumId);
    return album.title;
  }
}
