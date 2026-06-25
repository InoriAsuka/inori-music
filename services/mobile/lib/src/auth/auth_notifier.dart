import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'package:inori_music/src/api/api_client.dart';

const _kTokenKey = 'auth_token';
const _kUserIdKey = 'user_id';
const _kUsernameKey = 'username';
const _kBaseUrlKey = 'base_url';

// ---------------------------------------------------------------------------
// Auth state
// ---------------------------------------------------------------------------

enum AuthStatus { loading, authenticated, unauthenticated }

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
    return const AuthState(status: AuthStatus.unauthenticated);
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
