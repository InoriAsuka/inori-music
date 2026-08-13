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
    iOptions: IOSOptions(
      accessibility: KeychainAccessibility.first_unlock_this_device,
    ),
    // macOS (v5.37.0): the plugin defaults to the *data protection* keychain
    // (`kSecUseDataProtectionKeychain = true`). On macOS that keychain only
    // admits apps that carry a keychain access group entitlement, and such an
    // entitlement is only meaningful with a real Team ID prefix. This app is
    // ad-hoc signed (`CODE_SIGN_IDENTITY = "-"`, no `DEVELOPMENT_TEAM`), so
    // there is no prefix to claim and every `SecItemAdd` came back
    // `errSecMissingEntitlement` (-34018).
    //
    // The failure was write-only, which is why it hid for so long: a *read*
    // miss maps to `errSecItemNotFound`, which the plugin reports as "no
    // value" rather than an error, so the app started normally and showed the
    // login form. Only the first write — i.e. the moment someone actually
    // tried to sign in — ever hit it.
    //
    // The legacy file-based keychain needs no entitlement at all: under App
    // Sandbox an app always reaches its own keychain items. Nothing to
    // migrate, because no macOS build ever managed to write a single item.
    // Affects macOS only — iOS ignores this option (and always has a valid
    // application-identifier from provisioning).
    mOptions: MacOsOptions(useDataProtectionKeyChain: false),
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
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
    ),
  );

  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) async {
        final storage = ref.read(secureStorageProvider);
        final token = await storage.read(key: _kTokenKey);
        final baseUrl =
            await storage.read(key: _kBaseUrlKey) ?? _kDefaultBaseUrl;
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
