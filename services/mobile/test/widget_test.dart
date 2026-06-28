import 'package:flutter_test/flutter_test.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';

void main() {
  test('placeholder test', () {
    expect(1 + 1, 2);
  });

  test('AuthStatus enum has three values', () {
    expect(AuthStatus.values, hasLength(3));
    expect(AuthStatus.values, containsAll([
      AuthStatus.loading,
      AuthStatus.authenticated,
      AuthStatus.unauthenticated,
    ]));
  });
}
