import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

const _kBaseUrlKey = 'base_url';
const _kTokenKey = 'auth_token';
const _kDefaultBaseUrl = 'http://localhost:8080';

/// Broadcast stream that fires when a 401 causes an automatic logout.
/// [AuthNotifier] listens to this and invalidates itself to redirect the user.
final forceLogoutStream = StreamController<void>.broadcast();

/// Secure storage singleton
final secureStorageProvider = Provider<FlutterSecureStorage>(
  (_) => const FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
    iOptions: IOSOptions(accessibility: KeychainAccessibility.first_unlock_this_device),
  ),
);

/// Base URL provider — read from secure storage, fallback to localhost
final baseUrlProvider = FutureProvider<String>((ref) async {
  final storage = ref.watch(secureStorageProvider);
  return await storage.read(key: _kBaseUrlKey) ?? _kDefaultBaseUrl;
});

/// Token provider — read from secure storage
final tokenProvider = FutureProvider<String?>((ref) async {
  final storage = ref.watch(secureStorageProvider);
  return storage.read(key: _kTokenKey);
});

/// Dio HTTP client with auth interceptor
final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(
    BaseOptions(
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 30),
      headers: {'Accept': 'application/json', 'Content-Type': 'application/json'},
    ),
  );

  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) async {
        final storage = ref.read(secureStorageProvider);
        final token = await storage.read(key: _kTokenKey);
        final baseUrl = await storage.read(key: _kBaseUrlKey) ?? _kDefaultBaseUrl;
        options.baseUrl = baseUrl;
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) async {
        if (error.response?.statusCode == 401) {
          // Token expired — clear storage and broadcast so AuthNotifier can
          // transition to unauthenticated and trigger the login redirect.
          final storage = ref.read(secureStorageProvider);
          await storage.delete(key: _kTokenKey);
          ref.invalidate(tokenProvider);
          forceLogoutStream.add(null);
        }
        handler.next(error);
      },
    ),
  );

  return dio;
});
