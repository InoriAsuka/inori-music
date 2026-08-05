import 'dart:async';

import 'package:dio/dio.dart';
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

    final token = await _storage.read(key: _kTokenKey);
    final userId = await _storage.read(key: _kUserIdKey);
    final username = await _storage.read(key: _kUsernameKey);

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
    final prefs = await SharedPreferences.getInstance();
    if (prefs.getBool(_kLastModeGuestKey) ?? false) {
      return const AuthState(status: AuthStatus.guest);
    }
    return const AuthState(status: AuthStatus.unauthenticated);
  }

  /// Enter guest mode: purely local, no network call. Persists the choice so
  /// a relaunch goes straight back into guest mode (see [build]).
  Future<void> continueAsGuest() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_kLastModeGuestKey, true);
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

  Future<void> login(String username, String password, {String? baseUrl}) async {
    state = const AsyncLoading();

    if (baseUrl != null && baseUrl.isNotEmpty) {
      await _storage.write(key: _kBaseUrlKey, value: baseUrl);
    }

    try {
      final dio = _dio;
      final savedBase = await _storage.read(key: _kBaseUrlKey) ?? 'http://localhost:8080';
      final response = await dio.post(
        '$savedBase/api/v1/auth/login',
        data: {'username': username, 'password': password},
        options: Options(headers: {'Authorization': null}), // no auth on login
      );

      final data = response.data as Map<String, dynamic>;
      final token = data['token'] as String;
      final userId = data['userId'] as String;

      await _storage.write(key: _kTokenKey, value: token);
      await _storage.write(key: _kUserIdKey, value: userId);
      await _storage.write(key: _kUsernameKey, value: username);
      // A real login always wins over a previously-remembered guest session.
      (await SharedPreferences.getInstance()).remove(_kLastModeGuestKey);

      state = AsyncData(AuthState(
        status: AuthStatus.authenticated,
        token: token,
        userId: userId,
        username: username,
      ));
    } on DioException catch (e) {
      final msg = _extractError(e);
      state = AsyncData(AuthState(
        status: AuthStatus.unauthenticated,
        error: msg,
      ));
    }
  }

  Future<void> logout() async {
    final token = await _storage.read(key: _kTokenKey);
    if (token != null) {
      try {
        final savedBase = await _storage.read(key: _kBaseUrlKey) ?? 'http://localhost:8080';
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
    final savedBase = await _storage.read(key: _kBaseUrlKey) ?? 'http://localhost:8080';
    final token = await _storage.read(key: _kTokenKey);
    await _dio.post(
      '$savedBase/api/v1/me/change-password',
      data: {'currentPassword': current, 'newPassword': newPwd},
      options: Options(headers: {'Authorization': 'Bearer $token'}),
    );
  }

  /// Revoke all active sessions except the current one.
  Future<void> revokeAllOtherSessions() async {
    final savedBase = await _storage.read(key: _kBaseUrlKey) ?? 'http://localhost:8080';
    final token = await _storage.read(key: _kTokenKey);
    await _dio.delete(
      '$savedBase/api/v1/me/sessions/revoke-all',
      options: Options(headers: {'Authorization': 'Bearer $token'}),
    );
  }

  Future<Map<String, dynamic>> _fetchMe(String token) async {
    final savedBase = await _storage.read(key: _kBaseUrlKey) ?? 'http://localhost:8080';
    final response = await _dio.get(
      '$savedBase/api/v1/me',
      options: Options(headers: {'Authorization': 'Bearer $token'}),
    );
    return response.data as Map<String, dynamic>;
  }

  Future<void> _clearStorage() async {
    await _storage.delete(key: _kTokenKey);
    await _storage.delete(key: _kUserIdKey);
    await _storage.delete(key: _kUsernameKey);
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
}

final authProvider = AsyncNotifierProvider<AuthNotifier, AuthState>(
  AuthNotifier.new,
);
