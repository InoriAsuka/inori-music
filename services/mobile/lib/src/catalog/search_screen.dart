// ignore_for_file: implementation_imports
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_search_result.dart';
import 'package:inori_api/src/model/search_result_item.dart';
import 'package:inori_api/src/model/search_result_kind.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/shared/widgets/track_list_tile.dart';

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

class _SearchState {
  const _SearchState({this.query = '', this.result, this.isLoading = false, this.error});
  final String query;
  final CatalogSearchResult? result;
  final bool isLoading;
  final String? error;
}

class _SearchNotifier extends AutoDisposeNotifier<_SearchState> {
  Timer? _debounce;

  @override
  _SearchState build() {
    ref.onDispose(() => _debounce?.cancel());
    return const _SearchState();
  }

  void updateQuery(String q) {
    _debounce?.cancel();
    if (q.trim().isEmpty) {
      state = const _SearchState();
      return;
    }
    state = _SearchState(query: q, isLoading: true);
    _debounce = Timer(const Duration(milliseconds: 300), () => _doSearch(q));
  }

  Future<void> _doSearch(String q) async {
    try {
      final result = await ref.read(catalogRepositoryProvider).search(q);
      state = _SearchState(query: q, result: result);
    } catch (e) {
      state = _SearchState(query: q, error: '$e');
    }
  }
}

final _searchNotifierProvider =
    AutoDisposeNotifierProvider<_SearchNotifier, _SearchState>(_SearchNotifier.new);

// ---------------------------------------------------------------------------
// Screen
// ---------------------------------------------------------------------------

class SearchScreen extends ConsumerStatefulWidget {
  const SearchScreen({super.key});

  @override
  ConsumerState<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends ConsumerState<SearchScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabCtrl;
  final _ctrl = TextEditingController();

  @override
  void initState() {
    super.initState();
    _tabCtrl = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _tabCtrl.dispose();
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(_searchNotifierProvider);
    final artists = state.result?.items
            .where((i) => i.kind == SearchResultKind.artist)
            .toList() ??
        [];
    final albums = state.result?.items
            .where((i) => i.kind == SearchResultKind.album)
            .toList() ??
        [];
    final tracks = state.result?.items
            .where((i) => i.kind == SearchResultKind.track)
            .toList() ??
        [];

    return Scaffold(
      appBar: AppBar(
        title: TextField(
          controller: _ctrl,
          autofocus: true,
          decoration: InputDecoration(
            hintText: 'Search artists, albums, tracks…',
            border: InputBorder.none,
            enabledBorder: InputBorder.none,
            focusedBorder: InputBorder.none,
            suffixIcon: _ctrl.text.isNotEmpty
                ? IconButton(
                    icon: const Icon(Icons.clear),
                    onPressed: () {
                      _ctrl.clear();
                      ref.read(_searchNotifierProvider.notifier).updateQuery('');
                    },
                  )
                : null,
          ),
          onChanged: (q) =>
              ref.read(_searchNotifierProvider.notifier).updateQuery(q),
        ),
        bottom: TabBar(
          controller: _tabCtrl,
          tabs: [
            Tab(text: 'Artists (${artists.length})'),
            Tab(text: 'Albums (${albums.length})'),
            Tab(text: 'Tracks (${tracks.length})'),
          ],
        ),
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : state.error != null
              ? Center(child: Text(state.error!))
              : state.query.isEmpty
                  ? const Center(
                      child: Text('Start typing to search…',
                          style: TextStyle(color: NeonShrineColors.onSurfaceVariant)),
                    )
                  : TabBarView(
                      controller: _tabCtrl,
                      children: [
                        _ArtistResults(items: artists),
                        _AlbumResults(items: albums),
                        _TrackResults(items: tracks),
                      ],
                    ),
    );
  }
}

// ---------------------------------------------------------------------------
// Result sub-views
// ---------------------------------------------------------------------------

class _ArtistResults extends StatelessWidget {
  const _ArtistResults({required this.items});
  final List<SearchResultItem> items;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const Center(child: Text('No artists found'));
    }
    return ListView.builder(
      itemCount: items.length,
      itemBuilder: (context, i) {
        final artist = items[i].artist;
        if (artist == null) return const SizedBox();
        return ListTile(
          leading: const CircleAvatar(
            backgroundColor: NeonShrineColors.surfaceContainer,
            child: Icon(Icons.person, color: NeonShrineColors.outlineVariant),
          ),
          title: Text(artist.name),
          onTap: () => context.go(AppRoutes.artistDetailPath(artist.id)),
        );
      },
    );
  }
}

class _AlbumResults extends StatelessWidget {
  const _AlbumResults({required this.items});
  final List<SearchResultItem> items;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const Center(child: Text('No albums found'));
    }
    return ListView.builder(
      itemCount: items.length,
      itemBuilder: (context, i) {
        final album = items[i].album;
        if (album == null) return const SizedBox();
        return ListTile(
          leading: Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: NeonShrineColors.surfaceContainer,
              borderRadius: BorderRadius.circular(6),
            ),
            child: const Icon(Icons.album, color: NeonShrineColors.outlineVariant, size: 28),
          ),
          title: Text(album.title),
          subtitle: album.releaseYear != null ? Text('${album.releaseYear}') : null,
          onTap: () => context.go(AppRoutes.albumDetailPath(album.id)),
        );
      },
    );
  }
}

class _TrackResults extends StatelessWidget {
  const _TrackResults({required this.items});
  final List<SearchResultItem> items;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const Center(child: Text('No tracks found'));
    }
    return ListView.builder(
      itemCount: items.length,
      itemBuilder: (context, i) {
        final track = items[i].track;
        if (track == null) return const SizedBox();
        return TrackListTile(
          track: track,
          isFavorite: track.isFavorite ?? false,
        );
      },
    );
  }
}
