// keychain_read_scaling_test.dart
//
// Guard for the v5.37.2 "keychain prompt storm" fix.
//
// Report: after signing in on macOS, the keychain authorisation dialog kept
// popping up over and over — type the password, it immediately asks again,
// endlessly. Root cause: the dio interceptor in api_client.dart called
// `storage.read` twice on EVERY HTTP request (token + base URL). macOS uses
// the legacy file-based keychain here (see secureStorageProvider's doc
// comment — this app is ad-hoc signed, so it can't use the data-protection
// keychain), and that keychain asks the user to authorise access on every
// read from an app it doesn't already trust. Opening one artist page fires
// dozens of requests (artist + albums + tracks + per-album artwork lookups),
// so two keychain reads per request turned into dozens of dialogs — visually
// indistinguishable from an infinite loop.
//
// The fix is AuthCache (api_client.dart): an in-memory read-through cache so
// the interceptor touches secure storage at most once per process lifetime,
// no matter how many requests are issued. This test proves that holds by
// counting real `read` calls on a spy storage while driving the *actual*
// dioProvider interceptor through several requests with a stubbed transport
// (no network).
//
// Falsification (recorded per the v5.37.2 requirement to prove this guard is
// not vacuous): with the interceptor temporarily reverted to calling
// `storage.read` directly on every request (the pre-fix code), this test was
// observed to fail — read count tracked the request count exactly instead of
// staying at or below 1. See requirement.md's v5.37.2 entry for the numbers.
//
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/api/api_client.dart';

/// Mutable counter shared between the test and the spy storage below —
/// simpler than mocking a call-count matcher, and the number is what this
/// guard is actually about.
class _ReadCounter {
  int reads = 0;
}

/// Spy on [FlutterSecureStorage] that counts every `read` call and otherwise
/// behaves like a device that already has a token stored (the steady state
/// every ordinary browsing session runs under, post-login). Override
/// signatures copied from `_KeychainRejectingStorage` in
/// login_failure_visibility_test.dart.
class _ReadCountingStorage extends FlutterSecureStorage {
  const _ReadCountingStorage(this._counter);

  final _ReadCounter _counter;

  // 'auth_token' mirrors the private _kTokenKey constant in api_client.dart
  // (not importable — it's file-private by design, since nothing outside
  // that file should construct keys itself).
  static const _tokenKey = 'auth_token';

  @override
  Future<String?> read({
    required String key,
    IOSOptions? iOptions,
    AndroidOptions? aOptions,
    LinuxOptions? lOptions,
    WebOptions? webOptions,
    MacOsOptions? mOptions,
    WindowsOptions? wOptions,
  }) async {
    _counter.reads++;
    return key == _tokenKey ? 'steady-state-token' : null;
  }

  @override
  Future<void> write({
    required String key,
    required String? value,
    IOSOptions? iOptions,
    AndroidOptions? aOptions,
    LinuxOptions? lOptions,
    WebOptions? webOptions,
    MacOsOptions? mOptions,
    WindowsOptions? wOptions,
  }) async {}

  @override
  Future<void> delete({
    required String key,
    IOSOptions? iOptions,
    AndroidOptions? aOptions,
    LinuxOptions? lOptions,
    WebOptions? webOptions,
    MacOsOptions? mOptions,
    WindowsOptions? wOptions,
  }) async {}
}

/// Answers every request instantly with an empty JSON object. This test is
/// about how many times the keychain gets touched, not about response
/// handling, so no real network access is involved.
class _StubAdapter implements HttpClientAdapter {
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return ResponseBody.fromString(
      '{}',
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    // Simulates a device that already migrated base_url out of the keychain
    // (see AuthCache.baseUrl / auth_notifier.dart's _migrateToPrefs) — i.e.
    // every ordinary launch after the one-time migration following the
    // upgrade. That one-time migration read is a separate, deliberately
    // un-guarded cost (see requirement.md's v5.37.2 entry: "may prompt once;
    // that is acceptable and self-limiting"). What this test protects is the
    // steady state everyone spends the rest of their time in, where a
    // request must not touch the keychain for the base URL at all.
    SharedPreferences.setMockInitialValues({
      'base_url': 'http://10.0.0.5:8080',
    });
  });

  ProviderContainer buildContainer(_ReadCounter counter) {
    final container = ProviderContainer(
      overrides: [
        secureStorageProvider.overrideWithValue(_ReadCountingStorage(counter)),
      ],
    );
    addTearDown(container.dispose);
    return container;
  }

  test('issuing many sequential requests reads the keychain at most once, '
      'not once per request', () async {
    final counter = _ReadCounter();
    final container = buildContainer(counter);
    final dio = container.read(dioProvider)..httpClientAdapter = _StubAdapter();

    const requestCount = 12;
    for (var i = 0; i < requestCount; i++) {
      await dio.get('/api/v1/probe/$i');
    }

    expect(
      counter.reads,
      lessThanOrEqualTo(1),
      reason:
          'the interceptor must serve the token from the in-memory '
          'AuthCache after the first request, not read the keychain on '
          'every request — otherwise every keychain read is another '
          'macOS authorisation prompt, and $requestCount requests means '
          'up to $requestCount prompts. This is the exact reported bug: '
          'opening one artist page fires dozens of requests.',
    );
  });

  test('a burst of concurrent requests also coalesces onto a single keychain '
      'read', () async {
    // The sequential test above already proves the cache persists across
    // requests once warm. This proves the *first wave* doesn't fan out:
    // several requests fired without awaiting between them — as a real
    // screen does, e.g. an artist page kicking off artist + albums +
    // tracks together — must still share one in-flight read rather than
    // each independently observing "nothing cached yet" and starting its
    // own.
    final counter = _ReadCounter();
    final container = buildContainer(counter);
    final dio = container.read(dioProvider)..httpClientAdapter = _StubAdapter();

    await Future.wait([
      for (var i = 0; i < 8; i++) dio.get('/api/v1/probe/$i'),
    ]);

    expect(counter.reads, lessThanOrEqualTo(1));
  });
}
