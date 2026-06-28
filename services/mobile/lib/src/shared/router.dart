import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/auth/login_screen.dart';
import 'package:inori_music/src/catalog/artists_screen.dart';
import 'package:inori_music/src/catalog/artist_detail_screen.dart';
import 'package:inori_music/src/catalog/albums_screen.dart';
import 'package:inori_music/src/catalog/album_detail_screen.dart';
import 'package:inori_music/src/catalog/tracks_screen.dart';
import 'package:inori_music/src/catalog/playlists_screen.dart';
import 'package:inori_music/src/catalog/playlist_detail_screen.dart';
import 'package:inori_music/src/catalog/search_screen.dart';
import 'package:inori_music/src/player/full_player_screen.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/favorites/favorites_screen.dart';
import 'package:inori_music/src/history/history_screen.dart';
import 'package:inori_music/src/history/history_stats_screen.dart';
import 'package:inori_music/src/settings/settings_screen.dart';
import 'package:inori_music/src/shared/widgets/shell_scaffold.dart';

// ---------------------------------------------------------------------------
// Route paths
// ---------------------------------------------------------------------------

abstract class AppRoutes {
  static const login = '/login';
  static const home = '/';
  static const artists = '/artists';
  static const artistDetail = '/artists/:id';
  static const albums = '/albums';
  static const albumDetail = '/albums/:id';
  static const tracks = '/tracks';
  static const playlists = '/playlists';
  static const playlistDetail = '/playlists/:id';
  static const search = '/search';
  static const player = '/player';
  static const favorites = '/library/favorites';
  static const history = '/library/history';
  static const historyStats = '/library/history/stats';
  static const settings = '/settings';

  // Deep-link entry points (inori://tracks/:id, etc.)
  static const deepTrack = '/tracks/:id';
  static const deepAlbum = '/albums/:id';
  static const deepArtist = '/artists/:id';

  static String artistDetailPath(String id) => '/artists/$id';
  static String albumDetailPath(String id) => '/albums/$id';
  static String playlistDetailPath(String id) => '/playlists/$id';
  static String trackDeepLinkPath(String id) => '/tracks/$id';
}

// ---------------------------------------------------------------------------
// Router listenable that bridges Riverpod → GoRouter refresh
// ---------------------------------------------------------------------------

/// A [ChangeNotifier] that listens to [authProvider] and notifies GoRouter
/// when auth state changes. This avoids recreating the GoRouter on every
/// auth state update — the router is created once and only its refresh signal
/// triggers re-evaluation of the redirect callback.
class _AuthChangeNotifier extends ChangeNotifier {
  _AuthChangeNotifier(this._ref) {
    _ref.listen(authProvider, (prev, next) => notifyListeners());
  }
  final Ref _ref;
}

// ---------------------------------------------------------------------------
// Deep-link play screen
// ---------------------------------------------------------------------------

/// Handles `inori://tracks/<id>` deep links.
/// Immediately starts playback for the given track ID, then navigates to the
/// full player screen so the user lands on a meaningful UI.
class _DeepLinkTrackScreen extends ConsumerStatefulWidget {
  const _DeepLinkTrackScreen({required this.trackId});
  final String trackId;

  @override
  ConsumerState<_DeepLinkTrackScreen> createState() => _DeepLinkTrackScreenState();
}

class _DeepLinkTrackScreenState extends ConsumerState<_DeepLinkTrackScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      await ref.read(playerProvider.notifier).playTrack(widget.trackId);
      if (mounted) context.go(AppRoutes.player);
    });
  }

  @override
  Widget build(BuildContext context) {
    // Briefly visible while the async play resolves.
    return const Scaffold(
      body: Center(child: CircularProgressIndicator()),
    );
  }
}

// ---------------------------------------------------------------------------
// Router provider
// ---------------------------------------------------------------------------

final routerProvider = Provider<GoRouter>((ref) {
  final notifier = _AuthChangeNotifier(ref);
  ref.onDispose(notifier.dispose);

  return GoRouter(
    initialLocation: AppRoutes.artists,
    refreshListenable: notifier,
    redirect: (context, state) {
      final authState = ref.read(authProvider);

      // While auth is loading, show a splash instead of flashing content.
      if (authState is AsyncLoading) return AppRoutes.login;

      final isLoggedIn = authState.valueOrNull?.isAuthenticated ?? false;
      final isLoginRoute = state.matchedLocation == AppRoutes.login;

      if (!isLoggedIn && !isLoginRoute) return AppRoutes.login;
      if (isLoggedIn && isLoginRoute) return AppRoutes.artists;
      return null;
    },
    routes: [
      // Login — no shell
      GoRoute(
        path: AppRoutes.login,
        builder: (context, state) => const LoginScreen(),
      ),

      // Full player overlay — no shell
      GoRoute(
        path: AppRoutes.player,
        builder: (context, state) => const FullPlayerScreen(),
      ),

      // Deep link: inori://tracks/<id>  →  play track then go to /player
      GoRoute(
        path: '/tracks/:id',
        builder: (context, state) => _DeepLinkTrackScreen(
          trackId: state.pathParameters['id']!,
        ),
      ),

      // Shell (persistent nav + mini player)
      ShellRoute(
        builder: (context, state, child) => ShellScaffold(child: child),
        routes: [
          GoRoute(
            path: AppRoutes.artists,
            builder: (context, state) => const ArtistsScreen(),
            routes: [
              GoRoute(
                path: ':id',
                builder: (_, state) => ArtistDetailScreen(id: state.pathParameters['id']!),
              ),
            ],
          ),
          GoRoute(
            path: AppRoutes.albums,
            builder: (context, state) => const AlbumsScreen(),
            routes: [
              GoRoute(
                path: ':id',
                builder: (_, state) => AlbumDetailScreen(id: state.pathParameters['id']!),
              ),
            ],
          ),
          GoRoute(
            path: AppRoutes.tracks,
            builder: (context, state) => const TracksScreen(),
          ),
          GoRoute(
            path: AppRoutes.playlists,
            builder: (context, state) => const PlaylistsScreen(),
            routes: [
              GoRoute(
                path: ':id',
                builder: (_, state) => PlaylistDetailScreen(id: state.pathParameters['id']!),
              ),
            ],
          ),
          GoRoute(
            path: AppRoutes.search,
            builder: (context, state) => const SearchScreen(),
          ),
          GoRoute(
            path: AppRoutes.favorites,
            builder: (context, state) => const FavoritesScreen(),
          ),
          GoRoute(
            path: AppRoutes.history,
            builder: (context, state) => const HistoryScreen(),
            routes: [
              GoRoute(
                path: 'stats',
                builder: (context, state) => const HistoryStatsScreen(),
              ),
            ],
          ),
          GoRoute(
            path: AppRoutes.settings,
            builder: (context, state) => const SettingsScreen(),
          ),
        ],
      ),
    ],
  );
});
