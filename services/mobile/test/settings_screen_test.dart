// settings_screen_test.dart
//
// v5.30.7: the Settings screen had no test coverage before this phase. It
// earns a narrowly-scoped file now specifically to guard the Account
// section rework — the guest-mode login ListTile (with its full-width
// FilledButton.tonal trailing widget, the field report's original bug) was
// deleted outright once the desktop sidebar's own account block made it a
// duplicate (see shell_scaffold_nav_test.dart's `_GuestSignInPrompt`
// coverage for that entry point). This does not attempt to cover the rest
// of the screen (language/skin/EQ/crossfade/etc. — all unrelated to this
// phase's change) beyond what's needed to render it safely.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';
import 'package:inori_music/src/settings/settings_screen.dart';

import 'support/fake_playback_engine.dart';

class _StubAuthNotifier extends AuthNotifier {
  _StubAuthNotifier(this._state);
  final AuthState _state;

  @override
  Future<AuthState> build() async => _state;
}

const _guest = AuthState(status: AuthStatus.guest);
const _signedIn = AuthState(
  status: AuthStatus.authenticated,
  username: 'inori',
  userId: 'u-1',
);

Widget _buildApp(AuthState auth) => ProviderScope(
  overrides: [
    authProvider.overrideWith(() => _StubAuthNotifier(auth)),
    // _EqSection asks playbackCapabilitiesProvider (derived from this) for
    // whether to show live EQ controls or the "unsupported" placeholder —
    // either branch is safe to render, this just avoids depending on
    // whichever a real engine happens to resolve to on the test host.
    playbackEngineProvider.overrideWithValue(FakePlaybackEngine()),
  ],
  child: const MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: SettingsScreen(),
  ),
);

void main() {
  group('Account section (v5.30.7)', () {
    testWidgets(
      'guest mode shows no Account section at all — no header, no login '
      'row, no orphaned caption',
      (tester) async {
        await tester.pumpWidget(_buildApp(_guest));
        await tester.pumpAndSettle();

        expect(tester.takeException(), isNull);
        expect(
          find.text('Account'),
          findsNothing,
          reason:
              'The section header must not survive with nothing left under '
              'it once the guest login row is gone',
        );
        expect(find.text('以游客身份使用'), findsNothing);
        expect(find.text('登录后可使用云端曲库、收藏与跨设备续播'), findsNothing);
        expect(
          find.widgetWithText(FilledButton, '登录'),
          findsNothing,
          reason:
              'The guest-facing login button was removed outright — the '
              'sidebar\'s own account block is the only entry point now',
        );
      },
    );

    testWidgets(
      'a signed-in user still sees the Account section with their name '
      'and a change-password row',
      (tester) async {
        await tester.pumpWidget(_buildApp(_signedIn));
        await tester.pumpAndSettle();

        expect(tester.takeException(), isNull);
        expect(find.text('Account'), findsOneWidget);
        expect(find.text('inori'), findsOneWidget);
        expect(find.text('Logged in'), findsOneWidget);
      },
    );
  });
}
