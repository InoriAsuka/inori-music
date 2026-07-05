// ignore_for_file: implementation_imports
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_search_result.dart';
import 'package:inori_api/src/model/search_result_item.dart';
import 'package:inori_api/src/model/search_result_kind.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/catalog/search_history_provider.dart';
import 'package:inori_music/src/player/player_notifier.dart';
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
      ref.read(searchHistoryProvider.notifier).add(q);
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

  List<SearchResultItem> _suggestions = [];
  bool _showSuggestions = false;
  Timer? _suggestionDebounce;

  @override
  void initState() {
    super.initState();
    _tabCtrl = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabCtrl.dispose();
    _ctrl.dispose();
    _suggestionDebounce?.cancel();
    super.dispose();
  }

  void _triggerSuggestions(String q) {
    _suggestionDebounce?.cancel();
    if (q.length < 2) {
      setState(() {
        _suggestions = [];
        _showSuggestions = false;
      });
      return;
    }
    _suggestionDebounce = Timer(const Duration(milliseconds: 150), () async {
      try {
        final result = await ref.read(catalogRepositoryProvider).search(q, limit: 5);
        if (!mounted) return;
        final tracks = result.items.where((i) => i.track != null).take(5).toList();
        setState(() {
          _suggestions = tracks;
          _showSuggestions = tracks.isNotEmpty;
        });
      } catch (_) {}
    });
  }

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    final state = ref.watch(_searchNotifierProvider);
    final history = ref.watch(searchHistoryProvider);
    final allItems = state.result?.items ?? [];
    final artists = allItems.where((i) => i.kind == SearchResultKind.artist).toList();
    final albums = allItems.where((i) => i.kind == SearchResultKind.album).toList();
    final tracks = allItems.where((i) => i.kind == SearchResultKind.track).toList();

    final showEmptyPrompt = state.query.isEmpty && state.result == null && !state.isLoading;

    Widget bodyContent;
    if (state.isLoading) {
      bodyContent = const Center(child: CircularProgressIndicator());
    } else if (state.error != null) {
      bodyContent = Center(child: Text(state.error!));
    } else if (showEmptyPrompt) {
      bodyContent = Center(
        child: Text(t.searchPrompt,
            style: const TextStyle(color: NeonShrineColors.onSurfaceVariant)),
      );
    } else {
      bodyContent = TabBarView(
        controller: _tabCtrl,
        children: [
          _AllResults(items: allItems, t: t),
          _ArtistResults(items: artists, t: t),
          _AlbumResults(items: albums, t: t),
          _TrackResults(items: tracks, t: t),
        ],
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: TextField(
          controller: _ctrl,
          autofocus: true,
          textInputAction: TextInputAction.search,
          onSubmitted: (q) {
            setState(() => _showSuggestions = false);
            if (q.trim().isNotEmpty) {
              ref.read(searchHistoryProvider.notifier).add(q);
            }
          },
          decoration: InputDecoration(
            hintText: t.searchHint,
            border: InputBorder.none,
            enabledBorder: InputBorder.none,
            focusedBorder: InputBorder.none,
            suffixIcon: _ctrl.text.isNotEmpty
                ? IconButton(
                    icon: const Icon(Icons.clear),
                    onPressed: () {
                      _ctrl.clear();
                      ref.read(_searchNotifierProvider.notifier).updateQuery('');
                      setState(() {
                        _suggestions = [];
                        _showSuggestions = false;
                      });
                    },
                  )
                : null,
          ),
          onChanged: (q) {
            ref.read(_searchNotifierProvider.notifier).updateQuery(q);
            _triggerSuggestions(q);
            // rebuild to show/hide history overlay when text becomes empty
            setState(() {});
          },
        ),
        bottom: TabBar(
          controller: _tabCtrl,
          tabs: [
            const Tab(text: 'All'),
            Tab(text: '${t.artists} (${artists.length})'),
            Tab(text: '${t.albums} (${albums.length})'),
            Tab(text: '${t.tracks} (${tracks.length})'),
          ],
        ),
      ),
      body: Stack(
        children: [
          bodyContent,
          if (_showSuggestions && _suggestions.isNotEmpty)
            Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: Card(
                margin: EdgeInsets.zero,
                child: Column(
                  children: _suggestions.map((item) {
                    final track = item.track!;
                    return ListTile(
                      dense: true,
                      leading: const Icon(Icons.music_note, size: 16),
                      title: Text(track.title,
                          style: const TextStyle(fontSize: 13)),
                      onTap: () {
                        setState(() {
                          _showSuggestions = false;
                        });
                        _ctrl.clear();
                        ref.read(playerProvider.notifier).playTrack(track.id);
                      },
                    );
                  }).toList(),
                ),
              ),
            ),
          // search history overlay: visible when field is empty
          if (_ctrl.text.trim().isEmpty && history.isNotEmpty)
            Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: Card(
                margin: EdgeInsets.zero,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Padding(
                      padding: const EdgeInsets.fromLTRB(16, 12, 8, 8),
                      child: Row(
                        children: [
                          Text(
                            t.recentSearches,
                            style: const TextStyle(
                              color: NeonShrineColors.onSurfaceVariant,
                              fontSize: 12,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          const Spacer(),
                          IconButton(
                            icon: const Icon(Icons.delete_outline, size: 18),
                            tooltip: t.clearSearchHistory,
                            onPressed: () => ref
                                .read(searchHistoryProvider.notifier)
                                .clear(),
                          ),
                        ],
                      ),
                    ),
                    ...history.take(10).map((query) => ListTile(
                          dense: true,
                          leading: const Icon(Icons.history, size: 18),
                          title: Text(query, style: const TextStyle(fontSize: 14)),
                          trailing: IconButton(
                            icon: const Icon(Icons.close, size: 16),
                            onPressed: () => ref
                                .read(searchHistoryProvider.notifier)
                                .remove(query),
                          ),
                          onTap: () {
                            _ctrl.text = query;
                            _ctrl.selection = TextSelection.collapsed(
                                offset: query.length);
                            ref
                                .read(_searchNotifierProvider.notifier)
                                .updateQuery(query);
                            setState(() {
                              _showSuggestions = false;
                            });
                          },
                        )),
                    // Leave bottom padding so content doesn't stick to card edge.
                    const SizedBox(height: 4),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Highlight rendering
// ---------------------------------------------------------------------------

/// Parses a backend highlight snippet like "hello <mark>w</mark>orld" into
/// plain [TextSpan]s, with bold + primary-colored marked segments and
/// default-styled unmarked segments. Falls back to a single plain span when
/// the input is null, empty, or has no <mark> tags.
class _Highlighter {
  static const _markOpen = '<mark>';
  static const _markClose = '</mark>';

  /// Returns null when [raw] is null/blank/no-mark-present, signalling the
  /// caller to render plain text instead.
  static List<TextSpan>? spanify(String? raw, TextStyle base) {
    if (raw == null || raw.isEmpty) return null;
    if (!raw.contains(_markOpen)) return null;
    final out = <TextSpan>[];
    var rest = raw;
    while (rest.isNotEmpty) {
      final o = rest.indexOf(_markOpen);
      if (o < 0) {
        out.add(TextSpan(text: rest, style: base));
        break;
      }
      if (o > 0) {
        out.add(TextSpan(text: rest.substring(0, o), style: base));
      }
      final c = rest.indexOf(_markClose, o);
      if (c < 0) {
        // mismatched tag: emit the rest as plain text
        out.add(TextSpan(text: rest.substring(o), style: base));
        break;
      }
      final inner = rest.substring(o + _markOpen.length, c);
      out.add(TextSpan(
        text: inner,
        style: base.copyWith(
          fontWeight: FontWeight.bold,
          color: NeonShrineColors.primaryViolet,
        ),
      ));
      rest = rest.substring(c + _markClose.length);
    }
    return out;
  }
}

// ---------------------------------------------------------------------------
// Result sub-views
// ---------------------------------------------------------------------------

class _AllResults extends StatelessWidget {
  const _AllResults({required this.items, required this.t});
  final List<SearchResultItem> items;
  final AppLocalizations t;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.search_off, size: 64, color: NeonShrineColors.onSurfaceVariant),
            SizedBox(height: 16),
            Text('没有找到相关内容',
                style: TextStyle(color: NeonShrineColors.onSurfaceVariant)),
          ],
        ),
      );
    }
    return ListView.builder(
      itemCount: items.length,
      itemBuilder: (context, i) {
        final item = items[i];
        if (item.artist != null) {
          return ListTile(
            leading: const CircleAvatar(
              backgroundColor: NeonShrineColors.surfaceContainer,
              child: Icon(Icons.person, color: NeonShrineColors.outlineVariant),
            ),
            title: _HighlightedTitle(
                raw: item.highlight, plain: item.artist!.name),
            onTap: () => context.go(AppRoutes.artistDetailPath(item.artist!.id)),
          );
        }
        if (item.album != null) {
          return ListTile(
            leading: Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: NeonShrineColors.surfaceContainer,
                borderRadius: BorderRadius.circular(6),
              ),
              child: const Icon(Icons.album,
                  color: NeonShrineColors.outlineVariant, size: 28),
            ),
            title:
                _HighlightedTitle(raw: item.highlight, plain: item.album!.title),
            onTap: () => context.go(AppRoutes.albumDetailPath(item.album!.id)),
          );
        }
        if (item.track != null) {
          return TrackListTile(
              track: item.track!, isFavorite: item.track!.isFavorite ?? false);
        }
        return const SizedBox.shrink();
      },
    );
  }
}

class _ArtistResults extends StatelessWidget {
  const _ArtistResults({required this.items, required this.t});
  final List<SearchResultItem> items;
  final AppLocalizations t;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.search_off,
                size: 64, color: NeonShrineColors.onSurfaceVariant),
            const SizedBox(height: 16),
            Text(t.noResults,
                style:
                    const TextStyle(color: NeonShrineColors.onSurfaceVariant)),
          ],
        ),
      );
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
          title:
              _HighlightedTitle(raw: items[i].highlight, plain: artist.name),
          onTap: () => context.go(AppRoutes.artistDetailPath(artist.id)),
        );
      },
    );
  }
}

class _AlbumResults extends StatelessWidget {
  const _AlbumResults({required this.items, required this.t});
  final List<SearchResultItem> items;
  final AppLocalizations t;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.search_off,
                size: 64, color: NeonShrineColors.onSurfaceVariant),
            const SizedBox(height: 16),
            Text(t.noResults,
                style:
                    const TextStyle(color: NeonShrineColors.onSurfaceVariant)),
          ],
        ),
      );
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
            child: const Icon(Icons.album,
                color: NeonShrineColors.outlineVariant, size: 28),
          ),
          title:
              _HighlightedTitle(raw: items[i].highlight, plain: album.title),
          subtitle: album.releaseYear != null ? Text('${album.releaseYear}') : null,
          onTap: () => context.go(AppRoutes.albumDetailPath(album.id)),
        );
      },
    );
  }
}

class _TrackResults extends StatelessWidget {
  const _TrackResults({required this.items, required this.t});
  final List<SearchResultItem> items;
  final AppLocalizations t;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.search_off,
                size: 64, color: NeonShrineColors.onSurfaceVariant),
            const SizedBox(height: 16),
            Text(t.noResults,
                style:
                    const TextStyle(color: NeonShrineColors.onSurfaceVariant)),
          ],
        ),
      );
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

// ---------------------------------------------------------------------------
// Highlighted title widget
// ---------------------------------------------------------------------------

/// Title that renders the backend's highlight snippet when present and falls
/// back to plain text otherwise.
class _HighlightedTitle extends StatelessWidget {
  const _HighlightedTitle({required this.raw, required this.plain});
  final String? raw;
  final String plain;

  @override
  Widget build(BuildContext context) {
    const base = TextStyle(
      color: NeonShrineColors.onSurface,
      fontSize: 14,
      fontWeight: FontWeight.w500,
    );
    final spans = _Highlighter.spanify(raw, base);
    if (spans == null) {
      return Text(plain,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: base);
    }
    return RichText(
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      text: TextSpan(children: spans),
    );
  }
}
