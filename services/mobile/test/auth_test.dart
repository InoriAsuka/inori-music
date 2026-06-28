import 'package:flutter_test/flutter_test.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';

// ---------------------------------------------------------------------------
// Tests — AuthState value object
// ---------------------------------------------------------------------------

void main() {
  group('AuthState', () {
    test('isAuthenticated is true when status is authenticated', () {
      const state = AuthState(status: AuthStatus.authenticated);
      expect(state.isAuthenticated, isTrue);
    });

    test('isAuthenticated is false when status is unauthenticated', () {
      const state = AuthState(status: AuthStatus.unauthenticated);
      expect(state.isAuthenticated, isFalse);
    });

    test('isAuthenticated is false when status is loading', () {
      const state = AuthState(status: AuthStatus.loading);
      expect(state.isAuthenticated, isFalse);
    });

    test('copyWith preserves unchanged fields', () {
      const state = AuthState(
        status: AuthStatus.authenticated,
        userId: '1',
        username: 'alice',
        token: 'tok',
      );
      final updated = state.copyWith(status: AuthStatus.unauthenticated);
      expect(updated.status, AuthStatus.unauthenticated);
      expect(updated.userId, '1');
      expect(updated.username, 'alice');
      expect(updated.token, 'tok');
    });

    test('copyWith can set error', () {
      const state = AuthState(status: AuthStatus.unauthenticated);
      final withError = state.copyWith(error: 'Bad creds');
      expect(withError.error, 'Bad creds');
    });
  });

  group('AuthStatus', () {
    test('has three values', () {
      expect(AuthStatus.values, hasLength(3));
      expect(AuthStatus.values, containsAll([
        AuthStatus.loading,
        AuthStatus.authenticated,
        AuthStatus.unauthenticated,
      ]));
    });
  });
}
