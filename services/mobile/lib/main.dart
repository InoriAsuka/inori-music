import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/audio_handler.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/desktop_integration.dart';
import 'package:inori_music/src/shared/locale_provider.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

/// Global [InoriAudioHandler] instance shared between main.dart and PlayerNotifier.
late final InoriAudioHandler audioHandler;

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  // Initialize audio_service so background audio, lock-screen controls, and
  // OS media sessions are available before any widget is created.
  audioHandler = await InoriAudioHandler.create();
  audioHandler.initCrossfade();
  runApp(
    const ProviderScope(
      child: InoriMusicApp(),
    ),
  );
}

class InoriMusicApp extends ConsumerStatefulWidget {
  const InoriMusicApp({super.key});

  @override
  ConsumerState<InoriMusicApp> createState() => _InoriMusicAppState();
}

class _InoriMusicAppState extends ConsumerState<InoriMusicApp>
    with WidgetsBindingObserver {
  DesktopIntegration? _desktop;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    if (DesktopIntegration.isDesktop) {
      _desktop = DesktopIntegration(ref);
      // init is async; fire-and-forget — failures are logged inside the class.
      _desktop!.init();
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _desktop?.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    // Cross-device sync (v5.4.0): flush the player state immediately when the
    // app leaves the foreground, so progress survives a background kill.
    if (state == AppLifecycleState.paused ||
        state == AppLifecycleState.inactive ||
        state == AppLifecycleState.hidden) {
      ref.read(playerProvider.notifier).reportStateOnBackground();
    }
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(routerProvider);
    final locale = ref.watch(localeProvider);

    return MaterialApp.router(
      title: 'Inori Music',
      theme: buildNeonShrineTheme(),
      darkTheme: buildNeonShrineTheme(),
      themeMode: ThemeMode.dark,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
      locale: locale,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
    );
  }
}
