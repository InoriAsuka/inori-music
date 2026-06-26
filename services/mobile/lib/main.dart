import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/audio_handler.dart';
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
  runApp(
    const ProviderScope(
      child: InoriMusicApp(),
    ),
  );
}

class InoriMusicApp extends ConsumerWidget {
  const InoriMusicApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
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
