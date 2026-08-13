import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/services.dart' show PlatformException;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/api/api_client.dart';

const _kTokenKey = 'auth_token';
const _kUserIdKey = 'user_id';
const _kUsernameKey = 'username';
const _kBaseUrlKey = 'base_url';

/// SharedPreferences key for "last session was a guest session" — non-secret
/// UI preference, deliberately kept out of [FlutterSecureStorage] (that's
/// reserved for real credentials). Lets a guest relaunch straight into guest
/// mode instead of flashing the login screen again.
///
/// v5.37.2: `_kUserIdKey`/`_kUsernameKey`/`_kBaseUrlKey` above used to live in
/// [FlutterSecureStorage] too, violating exactly this convention — none of
/// them is a credential, and every one was another keychain ACL prompt on
/// macOS. They now follow this key's example; see [AuthCache] in
/// api_client.dart.
const _kLastModeGuestKey = 'auth.lastModeGuest';

// ---------------------------------------------------------------------------
// Auth state
// ---------------------------------------------------------------------------

enum AuthStatus { loading, authenticated, unauthenticated, guest }

class AuthState {
  const AuthState({
    required this.status,
    this.userId,
    this.username,
    this.token,
    this.error,
  });

  final AuthStatus status;
  final String? userId;
  final String? username;
  final String? token;
  final String? error;

  bool get isAuthenticated => status == AuthStatus.authenticated;
  bool get isGuest => status == AuthStatus.guest;
  // "Past the login gate" — either a real account or an explicit guest choice.
  // The router redirect and the shell nav both key off this rather than
  // `isAuthenticated` alone, so guest mode doesn't get bounced back to /login.
  bool get isPastGate => isAuthenticated || isGuest;

  AuthState copyWith({
    AuthStatus? status,
    String? userId,
    String? username,
    String? token,
    String? error,
  }) {
    return AuthState(
      status: status ?? this.status,
      userId: userId ?? this.userId,
      username: username ?? this.username,
      token: token ?? this.token,
      error: error,
    );
  }
}

// ---------------------------------------------------------------------------
// Auth notifier
// ---------------------------------------------------------------------------

class AuthNotifier extends AsyncNotifier<AuthState> {
  FlutterSecureStorage get _storage => ref.read(secureStorageProvider);
  Dio get _dio => ref.read(dioProvider);
  AuthCache get _cache => ref.read(authCacheProvider);

  @override
  Future<AuthState> build() async {
    // When the Dio interceptor detects a 401 (token expired / revoked), it
    // fires forceLogoutStream.  We listen here and transition to unauthenticated,
    // which triggers the go_router redirect to /login.
    final logoutSub = forceLogoutStream.stream.listen((_) async {
      await _clearStorage();
      state = const AsyncData(AuthState(status: AuthStatus.unauthenticated));
    });
    ref.onDispose(logoutSub.cancel);

    final prefs = await SharedPreferences.getInstance();
    // The only remaining keychain read on this path — AuthCache serves every
    // later call (this launch's, and dioProvider's interceptor) from memory.
    final token = await _cache.token(_storage);
    // user_id/username are plain SharedPreferences as of v5.37.2; on a
    // device still holding either in the keychain from v5.37.0/v5.37.1, this
    // migrates it once and deletes the keychain copy.
    final userId = await _migrateToPrefs(prefs, _kUserIdKey);
    final username = await _migrateToPrefs(prefs, _kUsernameKey);

    if (token != null && userId != null) {
      // Validate token by fetching /me
      try {
        await _fetchMe(token);
        return AuthState(
          status: AuthStatus.authenticated,
          token: token,
          userId: userId,
          username: username,
        );
      } catch (_) {
        await _clearStorage();
        return const AuthState(status: AuthStatus.unauthenticated);
      }
    }

    // No valid session — if the last thing this install did was explicitly
    // continue as a guest, skip straight back into guest mode instead of
    // flashing the login screen on every cold start.
    if (prefs.getBool(_kLastModeGuestKey) ?? false) {
      return const AuthState(status: AuthStatus.guest);
    }
    return const AuthState(status: AuthStatus.unauthenticated);
  }

  /// One-time migration for a value v5.37.0/v5.37.1 mistakenly stored in the
  /// keychain: non-secret UI state that never needed keychain protection,
  /// and every read of it was another macOS authorisation prompt (see
  /// [AuthCache] in api_client.dart). If [prefs] already has the value —
  /// true for every launch except the first one after upgrading from either
  /// of those two versions — this is a synchronous lookup and never touches
  /// the keychain at all. On a device with nothing to migrate (a fresh
  /// install, or one that never had this key set), the keychain lookup below
  /// still runs once per launch, but resolves to "item not found" — Keychain
  /// Services reports that without any authorisation UI, since there is no
  /// existing ACL to evaluate; only access to an item that exists prompts.
  Future<String?> _migrateToPrefs(SharedPreferences prefs, String key) async {
    final existing = prefs.getString(key);
    if (existing != null) return existing;
    final legacy = await _storage.read(key: key);
    if (legacy != null) {
      await prefs.setString(key, legacy);
      await _storage.delete(key: key);
    }
    return legacy;
  }

  /// Read-through helper mirroring [AuthCache.baseUrl] with the
  /// [SharedPreferences] instance fetched for the caller — used by every
  /// network call this notifier makes outside of [login] (which already has
  /// its own `prefs` in hand).
  Future<String> _resolvedBaseUrl() async {
    final prefs = await SharedPreferences.getInstance();
    return _cache.baseUrl(prefs, _storage);
  }

  /// Enter guest mode: purely local, no network call. Persists the choice so
  /// a relaunch goes straight back into guest mode (see [build]).
  Future<void> continueAsGuest() async {
    // Remembering the choice is a convenience (skip the login screen on the
    // next cold start), not a precondition for guest mode — which is purely
    // local and needs no storage at all. So a storage failure must not be
    // able to stop someone from getting in. This is bound straight to a
    // button's onPressed, where anything thrown is an uncaught async error:
    // the button would simply do nothing, forever, with no explanation.
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(_kLastModeGuestKey, true);
    } catch (_) {
      // Guest mode still works; it just won't be remembered next launch.
    }
    state = const AsyncData(AuthState(status: AuthStatus.guest));
  }

  /// Drop from guest mode back to genuinely unauthenticated, so the router's
  /// normal "not past the gate → /login" rule takes over. Used by the "Log
  /// in" entry point surfaced to guests in Settings — going through
  /// unauthenticated (rather than a special guest-only login path) means the
  /// router doesn't need a second, parallel set of login-reachability rules.
  Future<void> exitGuestMode() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_kLastModeGuestKey);
    state = const AsyncData(AuthState(status: AuthStatus.unauthenticated));
  }

  /// Releases the splash gate when auth never resolved on its own.
  ///
  /// The router holds the UI on the splash screen for as long as auth is
  /// [AsyncLoading], and that screen has no controls — it was written on the
  /// assumption that resolution always happens "well under a second", so when
  /// the assumption fails it becomes a dead end. The splash offers this after
  /// a timeout. Dropping to unauthenticated hands control back to the
  /// router's ordinary rules, which land on /login with [reason] displayed.
  ///
  /// Safe to call while [build] is still in flight: if that future later
  /// completes it simply publishes its own real result over this one.
  void abandonPendingAuth(String reason) {
    state = AsyncData(
      AuthState(status: AuthStatus.unauthenticated, error: reason),
    );
  }

  /// Signs in and **always** leaves [state] in a resolved (non-loading) state,
  /// on every path. That total-ness is load-bearing, not defensive style: the
  /// router pins the UI to the splash screen for exactly as long as auth is
  /// [AsyncLoading] (see `router.dart`), so a single escaping exception used
  /// to strand the user on a static logo screen permanently, with nothing
  /// logged and nothing shown. See [_describeUnexpected].
  Future<void> login(
    String username,
    String password, {
    String? baseUrl,
  }) async {
    state = const AsyncLoading();

    try {
      final prefs = await SharedPreferences.getInstance();
      // A caller-supplied base URL is a plain SharedPreferences write as of
      // v5.37.2 — no keychain, no entitlement/authorisation failure mode.
      // `_cache.setBaseUrl` makes it visible to this very call (below) and to
      // dioProvider's interceptor immediately, without a re-read.
      if (baseUrl != null && baseUrl.isNotEmpty) {
        await prefs.setString(_kBaseUrlKey, baseUrl);
        _cache.setBaseUrl(baseUrl);
      }

      final dio = _dio;
      final savedBase = await _cache.baseUrl(prefs, _storage);
      final response = await dio.post(
        '$savedBase/api/v1/auth/login',
        data: {'username': username, 'password': password},
        options: Options(headers: {'Authorization': null}), // no auth on login
      );

      final data = response.data as Map<String, dynamic>;
      final token = data['token'] as String;
      final userId = data['userId'] as String;

      // The one keychain write left in this method — and it stays inside
      // this try/catch on purpose. Keychain writes fail for reasons that
      // have nothing to do with the credentials (see secureStorageProvider);
      // this used to sit above the try, so on macOS the very first statement
      // after AsyncLoading threw straight past every handler.
      await _storage.write(key: _kTokenKey, value: token);
      _cache.setToken(token);
      await prefs.setString(_kUserIdKey, userId);
      await prefs.setString(_kUsernameKey, username);
      // A real login always wins over a previously-remembered guest session.
      await prefs.remove(_kLastModeGuestKey);

      state = AsyncData(
        AuthState(
          status: AuthStatus.authenticated,
          token: token,
          userId: userId,
          username: username,
        ),
      );
    } on DioException catch (e) {
      final msg = _extractError(e);
      state = AsyncData(
        AuthState(status: AuthStatus.unauthenticated, error: msg),
      );
    } catch (e) {
      // Catch-all on purpose. Transport failures are only one of the ways this
      // method can fail; it also writes to the keychain once (the token),
      // touches SharedPreferences several times (base URL, user id,
      // username, guest flag), and casts the response body. None of those
      // throw a DioException, and any of them escaping leaves the router
      // parked on the splash screen forever.
      state = AsyncData(
        AuthState(
          status: AuthStatus.unauthenticated,
          error: _describeUnexpected(e),
        ),
      );
    }
  }

  Future<void> logout() async {
    final token = await _cache.token(_storage);
    if (token != null) {
      try {
        final savedBase = await _resolvedBaseUrl();
        await _dio.post(
          '$savedBase/api/v1/auth/logout',
          options: Options(headers: {'Authorization': 'Bearer $token'}),
        );
      } catch (_) {
        // ignore — clear local state regardless
      }
    }
    await _clearStorage();
    (await SharedPreferences.getInstance()).remove(_kLastModeGuestKey);
    state = const AsyncData(AuthState(status: AuthStatus.unauthenticated));
  }

  Future<void> changePassword(String current, String newPwd) async {
    final savedBase = await _resolvedBaseUrl();
    final token = await _cache.token(_storage);
    await _dio.post(
      '$savedBase/api/v1/me/change-password',
      data: {'currentPassword': current, 'newPassword': newPwd},
      options: Options(headers: {'Authorization': 'Bearer $token'}),
    );
  }

  /// Revoke all active sessions except the current one.
  Future<void> revokeAllOtherSessions() async {
    final savedBase = await _resolvedBaseUrl();
    final token = await _cache.token(_storage);
    await _dio.delete(
      '$savedBase/api/v1/me/sessions/revoke-all',
      options: Options(headers: {'Authorization': 'Bearer $token'}),
    );
  }

  Future<Map<String, dynamic>> _fetchMe(String token) async {
    final savedBase = await _resolvedBaseUrl();
    final response = await _dio.get(
      '$savedBase/api/v1/me',
      options: Options(headers: {'Authorization': 'Bearer $token'}),
    );
    return response.data as Map<String, dynamic>;
  }

  /// Clears the token (cache + keychain) and the two migrated-out-of-keychain
  /// identity fields. Does not touch `base_url` — signing out of one account
  /// shouldn't forget which server the app was pointed at.
  Future<void> _clearStorage() async {
    await _storage.delete(key: _kTokenKey);
    _cache.setToken(null);
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_kUserIdKey);
    await prefs.remove(_kUsernameKey);
  }

  String _extractError(DioException e) {
    if (e.response?.statusCode == 401 || e.response?.statusCode == 403) {
      return 'Invalid username or password';
    }
    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.receiveTimeout) {
      return 'Connection timed out. Check server URL.';
    }
    if (e.type == DioExceptionType.connectionError) {
      return 'Cannot connect to server. Check URL and network.';
    }
    final body = e.response?.data;
    if (body is Map) {
      return body['message'] as String? ??
          body['error']?.toString() ??
          'Login failed';
    }
    return 'Login failed';
  }

  /// Turns a non-transport failure into text a user can actually report back.
  ///
  /// Deliberately keeps the raw platform code rather than collapsing
  /// everything to "Login failed": the difference between "wrong password"
  /// and "this build cannot write to the keychain at all" is the difference
  /// between a user error and a packaging bug, and only one of them is worth
  /// the user retyping their password over. A macOS keychain rejection
  /// (`-34018`) reaching the screen as an opaque "Login failed" would have
  /// been almost as useless as the frozen splash it replaces.
  String _describeUnexpected(Object e) {
    if (e is PlatformException) {
      // Secure storage lands here — the plugin surfaces every non-`noErr`
      // Keychain status as a PlatformException carrying the OSStatus in its
      // message, formatted as "Code: <status>, Message: <…>".
      final detail = e.message ?? '';
      // macOS gates keychain access behind a system authorisation prompt, and
      // this app is ad-hoc signed — its code signature changes with every
      // build, so the prompt returns for each new build the user installs,
      // not only on first run. The write that triggers it is the one in
      // login(), which means the first sign-in after an update races the
      // dialog and loses. Tell the user what to do instead of printing an
      // OSStatus at them; the form keeps its values, so retrying is one click.
      const authPrompt = [
        '-25293', // errSecAuthFailed
        '-25308', // errSecInteractionNotAllowed
        '-25315', // errSecInteractionRequired
        '-128', // errSecUserCanceled
      ];
      for (final code in authPrompt) {
        if (detail.contains(code)) {
          return 'Keychain access was not granted ($code). Click "Allow" in '
              'the macOS dialog, then press Sign In again.';
        }
      }
      return 'Secure storage error (${e.code}): '
          '${detail.isEmpty ? 'no detail' : detail}';
    }
    if (e is TypeError) {
      // A response that parsed as JSON but not into the shape we cast to —
      // usually a URL pointing at something that is not this API, or a server
      // too old for this client.
      return 'Unexpected response from server. Check the server URL and that '
          'it is running a compatible version.';
    }
    return 'Login failed: $e';
  }
}

final authProvider = AsyncNotifierProvider<AuthNotifier, AuthState>(
  AuthNotifier.new,
);
