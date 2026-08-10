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
import 'package:inori_music/src/discover/explore_screen.dart';
import 'package:inori_music/src/discover/for_you_screen.dart';
import 'package:inori_music/src/player/full_player_screen.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_transition.dart';
import 'package:inori_music/src/favorites/favorites_screen.dart';
import 'package:inori_music/src/history/history_screen.dart';
import 'package:inori_music/src/history/history_stats_screen.dart';
import 'package:inori_music/src/local_library/local_library_screen.dart';
import 'package:inori_music/src/settings/settings_screen.dart';
import 'package:inori_music/src/shared/splash_screen.dart';
import 'package:inori_music/src/shared/widgets/shell_scaffold.dart';
import 'package:inori_music/src/user_playlist/user_playlist_detail_screen.dart';
import 'package:inori_music/src/user_playlist/user_playlist_list_screen.dart';

// ---------------------------------------------------------------------------
// Route paths
// ---------------------------------------------------------------------------

abstract class AppRoutes {
  static const login = '/login';
  static const splash = '/splash';
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
  // v5.33.0 — the desktop sidebar's "发现音乐" group (EchoMusic骨架, see
  // shell_scaffold.dart's _discoverItems). Nested under /discover rather
  // than sitting at the root alongside artists/albums/etc. so the two read
  // as a related pair rather than two unrelated top-level destinations.
  static const forYou = '/discover/for-you';
  static const explore = '/discover/explore';
  static const favorites = '/library/favorites';
  static const history = '/library/history';
  static const historyStats = '/library/history/stats';
  static const settings = '/settings';
  static const localLibrary = '/local-library';
  static const myPlaylists = '/library/my-playlists';
  static const myPlaylistDetail = '/library/my-playlists/:id';

  // Deep-link entry points (inori://tracks/:id, etc.)
  // tracks/:id  → handled by top-level GoRoute (_DeepLinkTrackScreen)
  // albums/:id  → resolved by ShellRoute sub-route (AlbumDetailScreen)
  // artists/:id → resolved by ShellRoute sub-route (ArtistDetailScreen)
  static const deepTrack = '/tracks/:id';

  static String artistDetailPath(String id) => '/artists/$id';
  static String albumDetailPath(String id) => '/albums/$id';
  static String playlistDetailPath(String id) => '/playlists/$id';
  static String trackDeepLinkPath(String id) => '/tracks/$id';
  static String myPlaylistDetailPath(String id) => '/library/my-playlists/$id';
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
  ConsumerState<_DeepLinkTrackScreen> createState() =>
      _DeepLinkTrackScreenState();
}

class _DeepLinkTrackScreenState extends ConsumerState<_DeepLinkTrackScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      try {
        await ref.read(playerProvider.notifier).playTrack(widget.trackId);
        if (mounted) context.go(AppRoutes.player);
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(
            context,
          ).showSnackBar(SnackBar(content: Text('Could not play track: $e')));
          context.go(AppRoutes.artists);
        }
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    // Briefly visible while the async play resolves.
    return const Scaffold(body: Center(child: CircularProgressIndicator()));
  }
}

/// Handles `inori://albums/<id>` and `inori://artists/<id>` deep links.
///
/// Both paths are already covered by the ShellRoute sub-routes:
///   /albums/:id  →  AlbumDetailScreen (inside ShellScaffold)
///   /artists/:id →  ArtistDetailScreen (inside ShellScaffold)
///
/// GoRouter resolves these paths directly to the shell widgets with no extra
/// top-level route needed.  `deepAlbum` and `deepArtist` constants are kept
/// for documentation purposes but require no dedicated handler.

// ---------------------------------------------------------------------------
// Guest-mode route allow-list
// ---------------------------------------------------------------------------

/// Routes reachable in guest mode (no account). Guest mode is a local-files
/// player — every server-catalog route (Artists/Albums/Tracks/Playlists/
/// Search/Favorites/History/MyPlaylists) is meaningless without an account,
/// so this is an allow-list rather than a block-list: any new server-backed
/// route added later is safe-by-default instead of silently leaking through.
/// `/login` is handled separately in the redirect (a guest must always be
/// able to reach it to upgrade to a real account).
const _guestAllowedRoutes = [
  AppRoutes.localLibrary,
  AppRoutes.settings,
  AppRoutes.player,
];

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

      // While auth is loading, show a real splash instead of flashing content.
      if (authState is AsyncLoading) return AppRoutes.splash;

      final authValue = authState.valueOrNull;
      final isLoggedIn = authValue?.isAuthenticated ?? false;
      final isGuest = authValue?.isGuest ?? false;
      final isPastGate = isLoggedIn || isGuest;
      final isLoginRoute = state.matchedLocation == AppRoutes.login;

      // Splash is only ever valid while loading (handled above) — once auth
      // has resolved, always move off it immediately no matter the result,
      // rather than falling through to the isLoginRoute-keyed rules below
      // (which don't match "/splash" and would otherwise strand the user
      // there indefinitely).
      if (state.matchedLocation == AppRoutes.splash) {
        if (isLoggedIn) return AppRoutes.artists;
        if (isGuest) return AppRoutes.localLibrary;
        return AppRoutes.login;
      }

      if (!isPastGate && !isLoginRoute) return AppRoutes.login;
      if (isLoggedIn && isLoginRoute) return AppRoutes.artists;
      // Mirrors the isLoggedIn rule above: once past the gate (as a guest),
      // /login is never something you "sit on" — landing there (e.g. right
      // after tapping "Continue as Guest") bounces into the app. A guest who
      // wants to log in for real instead calls AuthNotifier.exitGuestMode(),
      // which drops back to genuinely unauthenticated so this same rule set
      // routes them to /login normally (see settings_screen.dart).
      if (isGuest && isLoginRoute) return AppRoutes.localLibrary;
      if (isGuest &&
          !_guestAllowedRoutes.any(
            (r) => state.matchedLocation.startsWith(r),
          )) {
        return AppRoutes.localLibrary;
      }
      return null;
    },
    routes: [
      // Splash — no shell, transient (see redirect above)
      GoRoute(
        path: AppRoutes.splash,
        builder: (context, state) => const SplashScreen(),
      ),

      // Login — no shell
      GoRoute(
        path: AppRoutes.login,
        builder: (context, state) => const LoginScreen(),
      ),

      // Full player overlay — no shell. pageBuilder (not builder) so this
      // gets a CustomTransitionPage instead of the platform's default
      // PageTransitionsTheme (FadeUpwardsPageTransitionsBuilder, which only
      // offsets 25% of the screen height plus a fade) — see
      // player_transition.dart for the actual slide/scrim/drag-to-dismiss
      // motion and why it's not defined inline here.
      GoRoute(
        path: AppRoutes.player,
        pageBuilder: (context, state) => CustomTransitionPage<void>(
          key: state.pageKey,
          // Only exists to satisfy CustomTransitionPage's required `child` —
          // playerPageTransitionsBuilder ignores it and builds its own
          // FullPlayerScreen instead, because `child` here is built *before*
          // `animation` exists (this pageBuilder callback never sees it),
          // while FullPlayerScreen specifically needs that live Animation to
          // gate its own first-frame cost (see FullPlayerScreen.transition's
          // doc comment) — there is no way to construct the real, wired-up
          // instance at this call site.
          child: const FullPlayerScreen(),
          // False, not the CustomTransitionPage default of true: this route
          // rises to *cover* the shell rather than replacing it outright, so
          // the shell must stay in the tree (and visible, in the gap this
          // page's own translation opens up) for the scrim/slide in
          // player_transition.dart to have anything to sit over.
          opaque: false,
          transitionDuration: playerTransitionDuration,
          reverseTransitionDuration: playerTransitionReverseDuration,
          transitionsBuilder: playerPageTransitionsBuilder,
        ),
      ),

      // Deep link: inori://tracks/<id>  →  play track then go to /player
      GoRoute(
        path: '/tracks/:id',
        builder: (context, state) =>
            _DeepLinkTrackScreen(trackId: state.pathParameters['id']!),
      ),

      // inori://albums/<id> and inori://artists/<id> are handled by GoRouter's
      // shell sub-routes (/albums/:id, /artists/:id) without a top-level override.
      // Adding duplicate top-level routes here would cause redirect loops.

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
                builder: (_, state) =>
                    ArtistDetailScreen(id: state.pathParameters['id']!),
              ),
            ],
          ),
          GoRoute(
            path: AppRoutes.albums,
            builder: (context, state) => const AlbumsScreen(),
            routes: [
              GoRoute(
                path: ':id',
                builder: (_, state) =>
                    AlbumDetailScreen(id: state.pathParameters['id']!),
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
                builder: (_, state) =>
                    PlaylistDetailScreen(id: state.pathParameters['id']!),
              ),
            ],
          ),
          GoRoute(
            path: AppRoutes.search,
            builder: (context, state) => const SearchScreen(),
          ),
          GoRoute(
            path: AppRoutes.forYou,
            builder: (context, state) => const ForYouScreen(),
          ),
          GoRoute(
            path: AppRoutes.explore,
            builder: (context, state) => const ExploreScreen(),
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
          GoRoute(
            path: AppRoutes.localLibrary,
            builder: (context, state) => const LocalLibraryScreen(),
          ),
          GoRoute(
            path: AppRoutes.myPlaylists,
            builder: (context, state) => const UserPlaylistListScreen(),
            routes: [
              GoRoute(
                path: ':id',
                builder: (_, state) => UserPlaylistDetailScreen(
                  playlistId: state.pathParameters['id']!,
                ),
              ),
            ],
          ),
        ],
      ),
    ],
  );
});
