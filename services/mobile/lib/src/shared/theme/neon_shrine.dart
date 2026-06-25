import 'package:flutter/material.dart';

/// Neon Shrine — Inori Music dark color palette
/// Primary violet: #9b5cff, inherited from v2 web player
abstract class NeonShrineColors {
  static const Color primaryViolet = Color(0xFF9B5CFF);
  static const Color primaryVioletLight = Color(0xFFB98AFF);
  static const Color primaryVioletDark = Color(0xFF6B2EDF);

  static const Color background = Color(0xFF0D0D14);
  static const Color surface = Color(0xFF16161F);
  static const Color surfaceVariant = Color(0xFF1E1E2A);
  static const Color surfaceContainer = Color(0xFF22222F);

  static const Color onBackground = Color(0xFFF0EEFF);
  static const Color onSurface = Color(0xFFE8E4FF);
  static const Color onSurfaceVariant = Color(0xFFB8B0D8);
  static const Color outline = Color(0xFF3A3650);
  static const Color outlineVariant = Color(0xFF2A2640);

  static const Color error = Color(0xFFFF5C7A);
  static const Color onError = Color(0xFF1A0010);

  static const Color playerBar = Color(0xFF1A1A27);
  static const Color miniPlayerShadow = Color(0x80000000);

  /// Neon accent colors
  static const Color accentCyan = Color(0xFF4FC3F7);
  static const Color accentPink = Color(0xFFFF6EAF);
  static const Color accentGreen = Color(0xFF69F0AE);
}

ThemeData buildNeonShrineTheme() {
  const colorScheme = ColorScheme.dark(
    primary: NeonShrineColors.primaryViolet,
    onPrimary: Colors.white,
    primaryContainer: NeonShrineColors.primaryVioletDark,
    onPrimaryContainer: NeonShrineColors.primaryVioletLight,
    secondary: NeonShrineColors.accentCyan,
    onSecondary: Color(0xFF001A26),
    tertiary: NeonShrineColors.accentPink,
    onTertiary: Color(0xFF1A0010),
    error: NeonShrineColors.error,
    onError: NeonShrineColors.onError,
    surface: NeonShrineColors.surface,
    onSurface: NeonShrineColors.onSurface,
    surfaceContainerHighest: NeonShrineColors.surfaceVariant,
    outline: NeonShrineColors.outline,
    outlineVariant: NeonShrineColors.outlineVariant,
    scrim: Color(0xCC000000),
  );

  return ThemeData(
    useMaterial3: true,
    colorScheme: colorScheme,
    scaffoldBackgroundColor: NeonShrineColors.background,
    fontFamily: 'Inter',
    appBarTheme: const AppBarTheme(
      backgroundColor: NeonShrineColors.background,
      foregroundColor: NeonShrineColors.onBackground,
      elevation: 0,
      scrolledUnderElevation: 0,
      centerTitle: false,
      titleTextStyle: TextStyle(
        fontFamily: 'Inter',
        fontSize: 18,
        fontWeight: FontWeight.w600,
        color: NeonShrineColors.onBackground,
      ),
    ),
    navigationBarTheme: const NavigationBarThemeData(
      backgroundColor: NeonShrineColors.playerBar,
      indicatorColor: NeonShrineColors.primaryVioletDark,
      elevation: 0,
    ),
    navigationRailTheme: const NavigationRailThemeData(
      backgroundColor: NeonShrineColors.surface,
      indicatorColor: NeonShrineColors.primaryVioletDark,
      selectedIconTheme: IconThemeData(color: NeonShrineColors.primaryVioletLight),
      unselectedIconTheme: IconThemeData(color: NeonShrineColors.onSurfaceVariant),
    ),
    cardTheme: CardThemeData(
      color: NeonShrineColors.surfaceVariant,
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: NeonShrineColors.outlineVariant, width: 0.5),
      ),
    ),
    listTileTheme: const ListTileThemeData(
      tileColor: Colors.transparent,
      contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 4),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: NeonShrineColors.surfaceVariant,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: const BorderSide(color: NeonShrineColors.outline),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: const BorderSide(color: NeonShrineColors.outline),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: const BorderSide(color: NeonShrineColors.primaryViolet, width: 1.5),
      ),
      labelStyle: const TextStyle(color: NeonShrineColors.onSurfaceVariant),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: NeonShrineColors.primaryViolet,
        foregroundColor: Colors.white,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        minimumSize: const Size(double.infinity, 48),
      ),
    ),
    sliderTheme: const SliderThemeData(
      activeTrackColor: NeonShrineColors.primaryViolet,
      inactiveTrackColor: NeonShrineColors.outlineVariant,
      thumbColor: NeonShrineColors.primaryVioletLight,
      overlayColor: Color(0x299B5CFF),
      trackHeight: 3,
    ),
    dividerTheme: const DividerThemeData(
      color: NeonShrineColors.outlineVariant,
      thickness: 0.5,
    ),
    iconTheme: const IconThemeData(color: NeonShrineColors.onSurfaceVariant),
    textTheme: const TextTheme(
      displayLarge: TextStyle(color: NeonShrineColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w700),
      displayMedium: TextStyle(color: NeonShrineColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w700),
      headlineLarge: TextStyle(color: NeonShrineColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w700),
      headlineMedium: TextStyle(color: NeonShrineColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w600),
      headlineSmall: TextStyle(color: NeonShrineColors.onBackground, fontFamily: 'Inter', fontWeight: FontWeight.w600),
      titleLarge: TextStyle(color: NeonShrineColors.onSurface, fontFamily: 'Inter', fontWeight: FontWeight.w600),
      titleMedium: TextStyle(color: NeonShrineColors.onSurface, fontFamily: 'Inter', fontWeight: FontWeight.w500),
      titleSmall: TextStyle(color: NeonShrineColors.onSurfaceVariant, fontFamily: 'Inter', fontWeight: FontWeight.w500),
      bodyLarge: TextStyle(color: NeonShrineColors.onSurface, fontFamily: 'Inter'),
      bodyMedium: TextStyle(color: NeonShrineColors.onSurface, fontFamily: 'Inter'),
      bodySmall: TextStyle(color: NeonShrineColors.onSurfaceVariant, fontFamily: 'Inter'),
      labelLarge: TextStyle(color: NeonShrineColors.onSurface, fontFamily: 'Inter', fontWeight: FontWeight.w500),
      labelMedium: TextStyle(color: NeonShrineColors.onSurfaceVariant, fontFamily: 'Inter'),
      labelSmall: TextStyle(color: NeonShrineColors.onSurfaceVariant, fontFamily: 'Inter'),
    ),
  );
}
