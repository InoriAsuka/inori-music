// platform_network_capability_test.dart
//
// Architecture test for the v5.36.0 platform-connectivity fix. The three
// facts asserted here — macOS App Sandbox's outbound-network entitlement,
// Android's release-build INTERNET permission, and Android's cleartext HTTP
// allowance — live entirely in native platform config (Info.plist-style
// entitlements, AndroidManifest.xml). None of it is Dart source: `flutter
// analyze` doesn't parse it and no other test in this suite reads it. That
// blindness is exactly how all three gaps went unnoticed since day one — the
// client has never successfully reached a server on any platform. If a
// future `flutter create` regen, a Xcode template upgrade, or a careless
// entitlements edit drops any of these again, this file is the only thing
// standing between that regression and another silent, undiagnosable outage.
// See .plan/20260812-072-v5.36.0-platform-connectivity-gaps.md for the full
// root-cause writeup.
//
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// Strips XML/plist comments before matching, so a key only *mentioned* in a
/// nearby comment (not actually set) can't accidentally satisfy a check.
String _withoutXmlComments(String xml) =>
    xml.replaceAll(RegExp(r'<!--.*?-->', dotAll: true), '');

/// True only if the plist dict sets [key] to boolean true — the `<key>`
/// element must be immediately followed by a `<true/>` value, not merely
/// present somewhere in the file.
bool _plistBoolTrue(String plist, String key) {
  final pattern = RegExp('<key>${RegExp.escape(key)}</key>\\s*<true\\s*/>');
  return pattern.hasMatch(_withoutXmlComments(plist));
}

void main() {
  test(
    'macOS Debug and Release entitlements grant outbound network access',
    () {
      // App Sandbox refuses every outbound connect() at the kernel level
      // (Seatbelt/MACF) unless com.apple.security.network.client is present
      // — enforced below the HTTP stack, so it applies equally to dio's
      // dart:io sockets and just_audio's native players. Both files sandbox
      // the app (com.apple.security.app-sandbox = true) and both are wired
      // as a real CODE_SIGN_ENTITLEMENTS in the Xcode project, so both need
      // the key independently — there is no debug-only carve-out.
      const files = [
        'macos/Runner/DebugProfile.entitlements',
        'macos/Runner/Release.entitlements',
      ];
      final offenders = <String>[];
      for (final path in files) {
        final content = File(path).readAsStringSync();
        if (!_plistBoolTrue(content, 'com.apple.security.network.client')) {
          offenders.add(path);
        }
      }

      expect(
        offenders,
        isEmpty,
        reason:
            'These entitlements files sandbox the app (app-sandbox = true) '
            'but do not grant com.apple.security.network.client. Without '
            'it, every outbound connection — login, catalog browsing, '
            'playback — is refused by the kernel before it ever reaches '
            "dio or just_audio; the app can't tell, because the connection "
            'never leaves the process.',
      );
    },
  );

  test('Android release builds declare the INTERNET permission', () {
    final content = _withoutXmlComments(
      File('android/app/src/main/AndroidManifest.xml').readAsStringSync(),
    );
    final hasInternetPermission = RegExp(
      r'<uses-permission\s+android:name="android\.permission\.INTERNET"\s*/>',
    ).hasMatch(content);

    expect(
      hasInternetPermission,
      isTrue,
      reason:
          'android/app/src/main/AndroidManifest.xml is missing the '
          'INTERNET permission. The debug and profile manifests declare it '
          "only for the Flutter tool's own use (hot reload, DevTools) — "
          'that grant does not carry into release builds, so a release APK '
          'would have zero network access no matter what the Dart code '
          'does.',
    );
  });

  test('Android allows cleartext HTTP for the still-HTTPS-less self-hosted '
      'server', () {
    final content = _withoutXmlComments(
      File('android/app/src/main/AndroidManifest.xml').readAsStringSync(),
    );
    final applicationTag = RegExp(
      r'<application[^>]*>',
      dotAll: true,
    ).firstMatch(content)?.group(0);

    expect(
      applicationTag,
      isNotNull,
      reason: 'Could not find an <application> tag to inspect at all.',
    );
    expect(
      applicationTag,
      contains('android:usesCleartextTraffic="true"'),
      reason:
          'Android 9+ (API 28+) blocks cleartext traffic by default. The '
          'self-hosted server has no HTTPS in front of it, and its '
          'address is user-entered at runtime, so there is no fixed '
          'domain known at build time to scope a network security '
          'config allowlist to instead — see the v5.36.0 plan for why '
          'this is a deliberate, single-switch trade-off. Without it, '
          "just_audio's Android backend (a real ExoPlayer instance, "
          'unlike dio) refuses to open the http:// audio stream.',
    );
  });
}
