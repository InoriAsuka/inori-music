import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/api/api_client.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/shared/router.dart';

// ---------------------------------------------------------------------------
// Guards for the v5.37.0 "login goes to the logo screen and never comes back"
// failure.
//
// The report was: fill in server / username / password, press Sign In, land on
// a static logo screen, and nothing ever happens — no spinner, no message, no
// error, no timeout. Three separate defects lined up to produce it:
//
//   1. secureStorageProvider used the macOS *data protection* keychain, which
//      rejects an ad-hoc-signed app with errSecMissingEntitlement (-34018) on
//      every write. Reads were fine (a miss is reported as "no value"), so the
//      app started normally and only died at the moment of signing in.
//   2. AuthNotifier.login caught only DioException. The keychain throws a
//      PlatformException, which escaped, leaving state pinned at AsyncLoading.
//   3. The router sends anything AsyncLoading to /splash — a screen that was
//      deliberately built with no controls, because it was assumed to be
//      transient. Pinned auth turned it into a dead end.
//
// Each of these is guarded below. Every guard was run against the pre-fix code
// first and observed to fail.
// ---------------------------------------------------------------------------

/// Stands in for the macOS keychain as it behaved before the fix: reads report
/// "nothing stored" (which is why the app got as far as the login form), and
/// every write is refused outright.
class _KeychainRejectingStorage extends FlutterSecureStorage {
  const _KeychainRejectingStorage();

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
  }) async {
    throw PlatformException(
      code: 'Unexpected security result code',
      message: "Code: -34018, Message: A required entitlement isn't present.",
    );
  }

  @override
  Future<String?> read({
    required String key,
    IOSOptions? iOptions,
    AndroidOptions? aOptions,
    LinuxOptions? lOptions,
    WebOptions? webOptions,
    MacOsOptions? mOptions,
    WindowsOptions? wOptions,
  }) async => null;
}

ProviderContainer _containerWith(FlutterSecureStorage storage) {
  final container = ProviderContainer(
    overrides: [secureStorageProvider.overrideWithValue(storage)],
  );
  addTearDown(container.dispose);
  return container;
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() => SharedPreferences.setMockInitialValues({}));

  group('AuthNotifier.login is total', () {
    test(
      'a keychain rejection does not escape and does not pin AsyncLoading',
      () async {
        final container = _containerWith(const _KeychainRejectingStorage());
        await container.read(authProvider.future);

        // The bug: this threw, so the await never returned normally and state
        // stayed AsyncLoading forever.
        await container
            .read(authProvider.notifier)
            .login('admin', 'hunter2', baseUrl: 'http://10.0.0.1:8080');

        final state = container.read(authProvider);
        expect(
          state,
          isA<AsyncData<AuthState>>(),
          reason: 'state left in AsyncLoading strands the router on /splash',
        );
        expect(state.value!.status, AuthStatus.unauthenticated);
      },
    );

    test(
      'the keychain OSStatus survives into the message shown to the user',
      () async {
        final container = _containerWith(const _KeychainRejectingStorage());
        await container.read(authProvider.future);

        await container
            .read(authProvider.notifier)
            .login('admin', 'hunter2', baseUrl: 'http://10.0.0.1:8080');

        final error = container.read(authProvider).value!.error;
        expect(error, isNotNull);
        // Collapsing this to a generic "Login failed" would be nearly as
        // useless as the frozen splash: -34018 is a packaging bug, not a typo
        // in the password, and only the code says which.
        expect(error, contains('-34018'));
      },
    );

    test('abandonPendingAuth releases the gate', () async {
      final container = _containerWith(const _KeychainRejectingStorage());
      await container.read(authProvider.future);

      container.read(authProvider.notifier).abandonPendingAuth('timed out');

      final state = container.read(authProvider);
      expect(state, isA<AsyncData<AuthState>>());
      expect(state.value!.status, AuthStatus.unauthenticated);
      expect(state.value!.error, 'timed out');
    });
  });

  group('secureStorageProvider platform options', () {
    test('macOS avoids the data protection keychain', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final storage = container.read(secureStorageProvider);
      expect(
        storage.mOptions.toMap()['useDataProtectionKeyChain'],
        'false',
        reason:
            'the data protection keychain needs a keychain access group '
            'entitlement, which needs a Team ID prefix; this app is ad-hoc '
            'signed and has none, so every write fails with -34018',
      );
    });
  });

  group('resolveAuthRedirect', () {
    const loading = AsyncLoading<AuthState>();
    const authed = AsyncData(AuthState(status: AuthStatus.authenticated));
    const anon = AsyncData(AuthState(status: AuthStatus.unauthenticated));
    const guest = AsyncData(AuthState(status: AuthStatus.guest));

    test('a sign-in in flight does not tear the user off the login form', () {
      // The bug: this returned '/splash', which is how pressing Sign In moved
      // the user to a screen with no controls. The form renders its own
      // spinner and error text for exactly this state.
      expect(
        resolveAuthRedirect(authState: loading, location: AppRoutes.login),
        isNull,
      );
    });

    test('a failed sign-in stays put so the error is readable', () {
      expect(
        resolveAuthRedirect(authState: anon, location: AppRoutes.login),
        isNull,
      );
    });

    test('loading anywhere else still shows the splash', () {
      expect(
        resolveAuthRedirect(authState: loading, location: AppRoutes.artists),
        AppRoutes.splash,
      );
    });

    test('a successful sign-in leaves the login form', () {
      expect(
        resolveAuthRedirect(authState: authed, location: AppRoutes.login),
        AppRoutes.artists,
      );
    });

    test('splash is never a resting place once auth resolves', () {
      expect(
        resolveAuthRedirect(authState: authed, location: AppRoutes.splash),
        AppRoutes.artists,
      );
      expect(
        resolveAuthRedirect(authState: guest, location: AppRoutes.splash),
        AppRoutes.localLibrary,
      );
      expect(
        resolveAuthRedirect(authState: anon, location: AppRoutes.splash),
        AppRoutes.login,
      );
    });

    test('the guest allow-list is unchanged by the loading exception', () {
      expect(
        resolveAuthRedirect(authState: guest, location: AppRoutes.localLibrary),
        isNull,
      );
      expect(
        resolveAuthRedirect(authState: guest, location: AppRoutes.artists),
        AppRoutes.localLibrary,
      );
      expect(
        resolveAuthRedirect(authState: anon, location: AppRoutes.artists),
        AppRoutes.login,
      );
    });
  });
}
