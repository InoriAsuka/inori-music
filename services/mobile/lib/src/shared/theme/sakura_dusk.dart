import 'package:flutter/material.dart';

/// Sakura Dusk — Inori Music shared light color palette.
/// Matches `packages/ui/src/styles/sakura-dusk.css` token values.
abstract class SakuraDuskColors {
  static const Color sakuraPink = Color(0xFFD42062);
  static const Color sakuraPinkLight = Color(0xFFFF8FB8);
  static const Color sakuraPinkDark = Color(0xFFBD1550);

  static const Color background = Color(0xFFFFF7F2);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color surfaceVariant = Color(0xFFFFF0F5);
  static const Color surfaceContainer = Color(0xFFFCE7EF);

  static const Color onBackground = Color(0xFF3B2A3F);
  static const Color onSurface = Color(0xFF3B2A3F);
  static const Color onSurfaceVariant = Color(0xFF6B5570);
  static const Color outline = Color(0xFFB77A96);
  static const Color outlineVariant = Color(0xFFF2D9E4);

  static const Color error = Color(0xFFC81E2C);
  static const Color onError = Color(0xFFFFFFFF);

  static const Color playerBar = Color(0xFFFFF0F5);
  static const Color miniPlayerShadow = Color(0x263B2A3F);

  static const Color accentCyan = Color(0xFF0A7D94);
  static const Color accentPink = Color(0xFFFF8FB8);
}

ThemeData buildSakuraDuskTheme() {
  const colorScheme = ColorScheme.light(
    primary: SakuraDuskColors.sakuraPink,
    onPrimary: Colors.white,
    primaryContainer: SakuraDuskColors.sakuraPinkLight,
    onPrimaryContainer: SakuraDuskColors.sakuraPinkDark,
    secondary: SakuraDuskColors.accentCyan,
    onSecondary: Colors.white,
    tertiary: SakuraDuskColors.accentPink,
    onTertiary: SakuraDuskColors.onBackground,
    error: SakuraDuskColors.error,
    onError: SakuraDuskColors.onError,
    surface: SakuraDuskColors.surface,
    onSurface: SakuraDuskColors.onSurface,
    surfaceContainerHighest: SakuraDuskColors.surfaceVariant,
    outline: SakuraDuskColors.outline,
    outlineVariant: SakuraDuskColors.outlineVariant,
    scrim: Color(0x263B2A3F),
  );

  return ThemeData(
    useMaterial3: true,
    colorScheme: colorScheme,
    scaffoldBackgroundColor: SakuraDuskColors.background,
    fontFamily: 'Inter',
    appBarTheme: const AppBarTheme(
      backgroundColor: SakuraDuskColors.background,
      foregroundColor: SakuraDuskColors.onBackground,
      elevation: 0,
      scrolledUnderElevation: 0,
      centerTitle: false,
      titleTextStyle: TextStyle(
        fontFamily: 'Inter',
        fontSize: 18,
        fontWeight: FontWeight.w600,
        color: SakuraDuskColors.onBackground,
      ),
    ),
    navigationBarTheme: const NavigationBarThemeData(
      backgroundColor: SakuraDuskColors.playerBar,
      indicatorColor: SakuraDuskColors.sakuraPinkDark,
      elevation: 0,
    ),
    navigationRailTheme: const NavigationRailThemeData(
      backgroundColor: SakuraDuskColors.surface,
      indicatorColor: SakuraDuskColors.sakuraPinkDark,
      selectedIconTheme: IconThemeData(color: SakuraDuskColors.sakuraPinkDark),
      unselectedIconTheme: IconThemeData(color: SakuraDuskColors.onSurfaceVariant),
    ),
    cardTheme: CardThemeData(
      color: SakuraDuskColors.surfaceVariant,
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: SakuraDuskColors.outlineVariant, width: 0.5),
      ),
    ),
    listTileTheme: const ListTileThemeData(
      tileColor: Colors.transparent,
      contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 4),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: SakuraDuskColors.surfaceVariant,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: const BorderSide(color: SakuraDuskColors.outline),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: const BorderSide(color: SakuraDuskColors.outline),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: const BorderSide(color: SakuraDuskColors.sakuraPink, width: 1.5),
      ),
      labelStyle: const TextStyle(color: SakuraDuskColors.onSurfaceVariant),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: SakuraDuskColors.sakuraPink,
        foregroundColor: Colors.white,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        minimumSize: const Size(double.infinity, 48),
      ),
    ),
    sliderTheme: const SliderThemeData(
      activeTrackColor: SakuraDuskColors.sakuraPink,
      inactiveTrackColor: SakuraDuskColors.outlineVariant,
      thumbColor: SakuraDuskColors.sakuraPinkDark,
      overlayColor: Color(0x29BD1550),
      trackHeight: 3,
    ),
    dividerTheme: const DividerThemeData(
      color: SakuraDuskColors.outlineVariant,
      thickness: 0.5,
    ),
    iconTheme: const IconThemeData(color: SakuraDuskColors.onSurfaceVariant),
    textTheme: const TextTheme(
      displayLarge: TextStyle(color: SakuraDuskColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w700),
      displayMedium: TextStyle(color: SakuraDuskColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w700),
      headlineLarge: TextStyle(color: SakuraDuskColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w700),
      headlineMedium: TextStyle(color: SakuraDuskColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w600),
      headlineSmall: TextStyle(color: SakuraDuskColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w600),
      titleLarge: TextStyle(color: SakuraDuskColors.onSurface, fontFamily: 'Inter', fontWeight: FontWeight.w600),
      titleMedium: TextStyle(color: SakuraDuskColors.onSurface, fontFamily: 'Inter', fontWeight: FontWeight.w500),
      titleSmall: TextStyle(color: SakuraDuskColors.onSurfaceVariant, fontFamily: 'Inter', fontWeight: FontWeight.w500),
      bodyLarge: TextStyle(color: SakuraDuskColors.onSurface, fontFamily: 'Inter'),
      bodyMedium: TextStyle(color: SakuraDuskColors.onSurface, fontFamily: 'Inter'),
      bodySmall: TextStyle(color: SakuraDuskColors.onSurfaceVariant, fontFamily: 'Inter'),
      labelLarge: TextStyle(color: SakuraDuskColors.onSurface, fontFamily: 'Inter', fontWeight: FontWeight.w500),
      labelMedium: TextStyle(color: SakuraDuskColors.onSurfaceVariant, fontFamily: 'Inter'),
      labelSmall: TextStyle(color: SakuraDuskColors.onSurfaceVariant, fontFamily: 'Inter'),
    ),
  );
}
