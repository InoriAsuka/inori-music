import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_acrylic/flutter_acrylic.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sqflite_common_ffi/sqflite_ffi.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/player/audio_handler.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/desktop_integration.dart';
import 'package:inori_music/src/shared/locale_provider.dart';
import 'package:inori_music/src/playback/engine_selection.dart';
import 'package:inori_music/src/playback/just_audio_engine.dart';
import 'package:inori_music/src/playback/media_kit_engine.dart';
import 'package:inori_music/src/playback/playback_engine.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_definition.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/titlebar_material_provider.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  // sqflite has no native implementation on Windows/Linux desktop — without
  // this, the first OfflineDb/LocalLibraryDb call on those platforms throws
  // "databaseFactory not initialized" (macOS/Android/iOS use sqflite's own
  // native plugin and don't need this). Must run before anything touches a
  // database, so this comes first in main().
  if (Platform.isWindows || Platform.isLinux) {
    sqfliteFfiInit();
    databaseFactory = databaseFactoryFfi;
  }
  // Windows-only Fluent Design backdrop (Mica/acrylic) — must be initialized
  // before any Window.setEffect() call. macOS/Linux/mobile skip this
  // entirely; see DesktopIntegration.applyTitlebarMaterial for the scope note.
  if (Platform.isWindows) {
    await Window.initialize();
  }
  // Build the playback engine and the OS media session before runApp: both
  // have to exist before any widget reads them, and audio_service's init is
  // async. They are injected as provider overrides rather than parked in a
  // global — a global is what previously let four unrelated notifiers reach
  // into main.dart for the audio player, which made every one of them
  // untestable without booting the real audio stack.
  //
  // choosePlaybackEngineKind is the only place that reads the platform for
  // this decision; everything below it is a pure function of that result,
  // so the choice itself is unit-tested without booting either audio stack
  // (see engine_selection_test.dart).
  final PlaybackEngine engine = switch (choosePlaybackEngineKind(
    Platform.operatingSystem,
  )) {
    EngineKind.mediaKit => MediaKitEngine.create(),
    EngineKind.justAudio => JustAudioEngine.create(),
  };
  final mediaSession = await InoriAudioHandler.create(engine);
  runApp(
    ProviderScope(
      overrides: [
        playbackEngineProvider.overrideWithValue(engine),
        mediaSessionProvider.overrideWithValue(mediaSession),
      ],
      child: const InoriMusicApp(),
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

  void _applyTitlebarMaterial(WidgetRef ref) {
    if (!Platform.isWindows) return;
    final material = ref.read(titlebarMaterialProvider);
    final skin = ref.read(skinProvider).active;
    DesktopIntegration.applyTitlebarMaterial(
      material,
      tint: skin.colors.playerBar.withValues(alpha: 0.8),
      isDark: skin.brightness == Brightness.dark,
    );
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(routerProvider);
    final locale = ref.watch(localeProvider);
    final skin = ref.watch(skinProvider).active;

    // Resize the desktop window when crossing the login gate in either
    // direction (narrow fixed window pre-login, wide resizable shell after).
    // Registered here rather than inside DesktopIntegration because
    // WidgetRef.listen is only valid within a widget's build method.
    ref.listen(authProvider, (prev, next) {
      final wasPastGate = prev?.valueOrNull?.isPastGate ?? false;
      final isPastGate = next.valueOrNull?.isPastGate ?? false;
      if (wasPastGate != isPastGate) {
        _desktop?.applyWindowForAuthState(isPastGate);
      }
    });

    // Watching (not just listening) means this also runs on the very first
    // build, with whatever the not-yet-restored default is — and again once
    // the persisted preference/skin actually resolve, since either watch
    // changing re-triggers this build. Window.setEffect() is idempotent, so
    // re-applying on unrelated rebuilds (e.g. a locale change) is harmless.
    ref.watch(titlebarMaterialProvider);
    _applyTitlebarMaterial(ref);

    return MaterialApp.router(
      title: 'Inori Music',
      theme: buildThemeFromSkin(skin),
      themeMode: ThemeMode.light,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
      locale: locale,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      builder: (context, child) => SkinScope(skin: skin, child: child!),
    );
  }
}
