// settings_screen_test.dart
//
// v5.30.7: the Settings screen had no test coverage before this phase. It
// earned a narrowly-scoped file specifically to guard the Account section
// rework — the guest-mode login ListTile (with its full-width
// FilledButton.tonal trailing widget, the field report's original bug) was
// deleted outright on the theory that the desktop sidebar's own account
// block (`_GuestSignInPrompt` in shell_scaffold.dart) already covered
// sign-in for a guest.
//
// v5.31.0 fixes the regression that theory introduced: `_GuestSignInPrompt`
// only exists inside `_DesktopSidebar`, which only ever renders at
// >=1200dp — mobile and tablet guests were left with *no* UI path back to a
// real account at all. This screen now grows its own sign-in entry point
// back, gated on `ShellChrome.of(context) == null` (no desktop sidebar
// ancestor) rather than deleted outright, so the two entry points never
// both show at once.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/playback/playback_engine.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';
import 'package:inori_music/src/settings/settings_screen.dart';
import 'package:inori_music/src/shared/widgets/shell_chrome.dart';

import 'support/fake_playback_engine.dart';

class _StubAuthNotifier extends AuthNotifier {
  _StubAuthNotifier(this._state);
  final AuthState _state;

  @override
  Future<AuthState> build() async => _state;
}

/// Same as [_StubAuthNotifier], but records whether the guest sign-in entry
/// point actually reached [AuthNotifier.exitGuestMode] instead of just
/// looking tappable. Overriding the method itself (rather than letting the
/// real implementation run) keeps this test from depending on
/// `SharedPreferences`' test-harness behaviour, which is
/// `exitGuestMode`'s own concern, not this screen's — the same reasoning
/// `shell_scaffold_nav_test.dart` applies by never actually tapping its own
/// `_GuestSignInPrompt` either.
class _SpyAuthNotifier extends AuthNotifier {
  _SpyAuthNotifier(this._state);
  final AuthState _state;
  bool exitGuestModeCalled = false;

  @override
  Future<AuthState> build() async => _state;

  @override
  Future<void> exitGuestMode() async {
    exitGuestModeCalled = true;
  }
}

const _guest = AuthState(status: AuthStatus.guest);
const _signedIn = AuthState(
  status: AuthStatus.authenticated,
  username: 'inori',
  userId: 'u-1',
);

Widget _buildApp(
  AuthNotifier Function() authNotifier, {
  Widget? shellChrome,
  PlaybackCapabilities capabilities = PlaybackCapabilities.none,
}) => ProviderScope(
  overrides: [
    authProvider.overrideWith(authNotifier),
    // _EqSection and the crossfade slider both ask
    // playbackCapabilitiesProvider (derived from this) for whether to
    // show live controls or the "unsupported" placeholder — either
    // branch is safe to render, this just avoids depending on
    // whichever a real engine happens to resolve to on the test host.
    // Defaults to PlaybackCapabilities.none (everything gated off) so
    // existing callers that don't care about a specific capability
    // keep exercising the "engine supports nothing extra" branch.
    playbackEngineProvider.overrideWithValue(
      FakePlaybackEngine(capabilities: capabilities),
    ),
  ],
  child: MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: shellChrome ?? const SettingsScreen(),
  ),
);

void main() {
  group('Account section (v5.30.7 / v5.31.0)', () {
    testWidgets(
      'a signed-in user still sees the Account section with their name '
      'and a change-password row',
      (tester) async {
        await tester.pumpWidget(_buildApp(() => _StubAuthNotifier(_signedIn)));
        await tester.pumpAndSettle();

        expect(tester.takeException(), isNull);
        expect(find.text('Account'), findsOneWidget);
        expect(find.text('inori'), findsOneWidget);
        expect(find.text('Logged in'), findsOneWidget);
      },
    );

    testWidgets(
      'guest mode with no ShellChrome ancestor (mobile/tablet — no desktop '
      'sidebar exists to carry the sign-in prompt instead) shows its own '
      'entry point, and tapping it reaches exitGuestMode',
      (tester) async {
        final spy = _SpyAuthNotifier(_guest);
        await tester.pumpWidget(_buildApp(() => spy));
        await tester.pumpAndSettle();

        expect(tester.takeException(), isNull);
        expect(find.text('Account'), findsOneWidget);
        expect(find.text('Tap to sign in'), findsOneWidget);

        await tester.tap(find.text('Tap to sign in'));
        await tester.pumpAndSettle();

        expect(tester.takeException(), isNull);
        expect(spy.exitGuestModeCalled, isTrue);
      },
    );

    testWidgets('guest mode under a ShellChrome ancestor (a desktop sidebar is '
        'already present above this screen) does not duplicate the sign-in '
        'entry point', (tester) async {
      await tester.pumpWidget(
        _buildApp(
          () => _StubAuthNotifier(_guest),
          shellChrome: const ShellChrome(
            reservesTrafficLightGutter: true,
            child: SettingsScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(
        find.text('Account'),
        findsNothing,
        reason:
            'The section header must not survive with nothing left under '
            'it once the sign-in row defers to the sidebar\'s own block',
      );
      expect(find.text('Tap to sign in'), findsNothing);
    });
  });

  group('Crossfade slider capability gate (v5.38.1)', () {
    // v5.38.0 added MediaKitEngine (Windows), which honestly reports
    // capabilities.crossfade: false and treats crossfadeSeconds as a
    // documented no-op setter. The slider itself had no capability check
    // at all, so on that engine it dragged, showed a number, and
    // persisted a value the engine silently discarded. These tests guard
    // the fix mirroring _EqSection's existing gate.
    testWidgets(
      'crossfade: false (e.g. media_kit on Windows) shows the disabled '
      'explanatory tile and renders no crossfade Slider',
      (tester) async {
        await tester.pumpWidget(
          _buildApp(
            () => _StubAuthNotifier(_signedIn),
            capabilities: const PlaybackCapabilities(crossfade: false),
          ),
        );
        await tester.pumpAndSettle();
        // The crossfade tile sits inside the "音频" section, well past the
        // Account/Appearance/Offline Library/歌词 sections above it — the
        // screen's root ListView only builds slivers near the viewport, so
        // it isn't in the tree until scrolled into range.
        await tester.scrollUntilVisible(find.text('切歌淡入淡出'), 300);
        await tester.pumpAndSettle();

        expect(tester.takeException(), isNull);
        expect(find.text('切歌淡入淡出'), findsOneWidget);
        expect(find.text('当前播放引擎不提供切歌淡入淡出'), findsOneWidget);
        expect(find.byType(Slider), findsNothing);

        final tile = tester.widget<ListTile>(
          find.ancestor(
            of: find.text('当前播放引擎不提供切歌淡入淡出'),
            matching: find.byType(ListTile),
          ),
        );
        expect(tile.enabled, isFalse);
      },
    );

    testWidgets(
      'crossfade: true (just_audio — every platform except Windows) still '
      'renders the working, draggable slider',
      (tester) async {
        await tester.pumpWidget(
          _buildApp(
            () => _StubAuthNotifier(_signedIn),
            capabilities: const PlaybackCapabilities(crossfade: true),
          ),
        );
        await tester.pumpAndSettle();
        await tester.scrollUntilVisible(find.text('切歌淡入淡出'), 300);
        await tester.pumpAndSettle();

        expect(tester.takeException(), isNull);
        expect(find.text('切歌淡入淡出'), findsOneWidget);
        expect(find.text('当前播放引擎不提供切歌淡入淡出'), findsNothing);
        expect(find.byType(Slider), findsOneWidget);
      },
    );
  });
}
