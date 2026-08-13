import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _kBaseUrlKey = 'base_url';
const _kTokenKey = 'auth_token';
const _kDefaultBaseUrl = 'http://localhost:8080';

/// Broadcast stream that fires when a 401 causes an automatic logout.
/// [AuthNotifier] listens to this and invalidates itself to redirect the user.
final forceLogoutStream = StreamController<void>.broadcast();

/// Secure storage singleton — reserved for actual credentials as of v5.37.2
/// (just the bearer token). `base_url`/`user_id`/`username` used to live
/// here too even though none of them is a secret; see [AuthCache] for why
/// that mattered enough to move them to SharedPreferences.
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

/// Holds the single process-lifetime [AuthCache] instance.
final authCacheProvider = Provider<AuthCache>((ref) => AuthCache());

/// Caches the bearer token and base URL for the process lifetime so no code
/// path — in particular [dioProvider]'s interceptor — reads secure storage
/// more than once per launch.
///
/// v5.37.2: the interceptor used to call `storage.read` twice on *every*
/// HTTP request (token + base URL). On macOS, [secureStorageProvider]
/// deliberately uses the legacy file-based keychain because this app is
/// ad-hoc signed (see that provider's doc comment) — and the legacy keychain
/// asks the user to authorise access on every read from an app it doesn't
/// already trust. Opening one artist page fires dozens of requests (artist +
/// albums + tracks + per-album artwork lookups), so two reads per request
/// turned into dozens of authorisation dialogs back to back: "keeps popping
/// up over and over" from the user's side, a scaling bug from the code's.
///
/// Both fields cache a [Future] rather than a resolved value so that several
/// callers racing the very first read (an artist page fires off artist +
/// albums + tracks without awaiting between them) share the one in-flight
/// read instead of each seeing "nothing cached yet" and starting their own —
/// the `??=` below runs synchronously, before the read it stores ever gets a
/// chance to complete.
class AuthCache {
  Future<String?>? _tokenFuture;
  Future<String>? _baseUrlFuture;

  /// Read-through: reads secure storage on the first call ever made; every
  /// later call — including ones racing the first — is served from
  /// [_tokenFuture].
  Future<String?> token(FlutterSecureStorage storage) {
    return _tokenFuture ??= storage.read(key: _kTokenKey);
  }

  /// Read-through with a one-time migration. v5.37.0/v5.37.1 stored the base
  /// URL in the keychain despite it not being a credential; a device
  /// upgrading from either version still has it there instead of in
  /// [prefs]. That migration read is the one keychain access this class
  /// does not avoid — it happens at most once ever per install (the
  /// keychain copy is deleted right after) and only for installs that ran
  /// one of those two versions. See requirement.md's v5.37.2 entry.
  Future<String> baseUrl(
    SharedPreferences prefs,
    FlutterSecureStorage storage,
  ) {
    return _baseUrlFuture ??= _loadBaseUrl(prefs, storage);
  }

  Future<String> _loadBaseUrl(
    SharedPreferences prefs,
    FlutterSecureStorage storage,
  ) async {
    final stored = prefs.getString(_kBaseUrlKey);
    if (stored != null) return stored;
    final legacy = await storage.read(key: _kBaseUrlKey);
    if (legacy != null) {
      await prefs.setString(_kBaseUrlKey, legacy);
      await storage.delete(key: _kBaseUrlKey);
      return legacy;
    }
    return _kDefaultBaseUrl;
  }

  /// Overwrites the cached token immediately. Called right after a keychain
  /// write (login) or delete (logout, or the 401 force-logout path in
  /// [dioProvider]) so a still-warm cache can never keep serving a value
  /// storage no longer has — a stale token surviving logout would let a
  /// signed-out session keep making authenticated requests.
  void setToken(String? value) {
    _tokenFuture = Future.value(value);
  }

  /// Overwrites the cached base URL immediately. Called right after login
  /// saves a (possibly new) server address, so a request made moments later
  /// doesn't silently keep talking to the old server.
  void setBaseUrl(String value) {
    _baseUrlFuture = Future.value(value);
  }
}

/// Base URL provider — read-through [AuthCache], falls back to localhost.
final baseUrlProvider = FutureProvider<String>((ref) async {
  final cache = ref.watch(authCacheProvider);
  final storage = ref.watch(secureStorageProvider);
  final prefs = await SharedPreferences.getInstance();
  return cache.baseUrl(prefs, storage);
});

/// Token provider — read-through [AuthCache].
final tokenProvider = FutureProvider<String?>((ref) async {
  final cache = ref.watch(authCacheProvider);
  final storage = ref.watch(secureStorageProvider);
  return cache.token(storage);
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
        final cache = ref.read(authCacheProvider);
        final storage = ref.read(secureStorageProvider);
        final prefs = await SharedPreferences.getInstance();
        // Both of these are served from AuthCache after the first request of
        // the process's lifetime — see its doc comment for why that matters
        // on macOS. Neither `storage.read` nor `prefs.getString` runs here on
        // steady state.
        final token = await cache.token(storage);
        final baseUrl = await cache.baseUrl(prefs, storage);
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
          ref.read(authCacheProvider).setToken(null);
          ref.invalidate(tokenProvider);
          forceLogoutStream.add(null);
        }
        handler.next(error);
      },
    ),
  );

  return dio;
});
