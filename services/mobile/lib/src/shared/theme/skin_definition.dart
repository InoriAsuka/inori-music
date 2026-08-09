import 'dart:convert';
import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'package:inori_music/src/shared/theme/sakura_dusk.dart';

/// The full set of color tokens a skin must define. Field names mirror
/// [SakuraDuskColors] 1:1 on purpose — every call site that used to read the
/// static Sakura Dusk constant now reads the same-named field off the active
/// skin instead, so the token vocabulary app-wide didn't need to change.
class SkinColors {
  const SkinColors({
    required this.sakuraPink,
    required this.sakuraPinkLight,
    required this.sakuraPinkDark,
    required this.background,
    required this.surface,
    required this.surfaceVariant,
    required this.surfaceContainer,
    required this.onBackground,
    required this.onSurface,
    required this.onSurfaceVariant,
    required this.outline,
    required this.outlineVariant,
    required this.error,
    required this.onError,
    required this.playerBar,
    required this.miniPlayerShadow,
    required this.accentCyan,
    required this.accentPink,
  });

  final Color sakuraPink;
  final Color sakuraPinkLight;
  final Color sakuraPinkDark;
  final Color background;
  final Color surface;
  final Color surfaceVariant;
  final Color surfaceContainer;
  final Color onBackground;
  final Color onSurface;
  final Color onSurfaceVariant;
  final Color outline;
  final Color outlineVariant;
  final Color error;
  final Color onError;
  final Color playerBar;
  final Color miniPlayerShadow;
  final Color accentCyan;
  final Color accentPink;

  /// The original Sakura Dusk (light) values, unchanged from the hardcoded
  /// [SakuraDuskColors] constants that shipped through v5.13.0.
  static const sakuraDusk = SkinColors(
    sakuraPink: SakuraDuskColors.sakuraPink,
    sakuraPinkLight: SakuraDuskColors.sakuraPinkLight,
    sakuraPinkDark: SakuraDuskColors.sakuraPinkDark,
    background: SakuraDuskColors.background,
    surface: SakuraDuskColors.surface,
    surfaceVariant: SakuraDuskColors.surfaceVariant,
    surfaceContainer: SakuraDuskColors.surfaceContainer,
    onBackground: SakuraDuskColors.onBackground,
    onSurface: SakuraDuskColors.onSurface,
    onSurfaceVariant: SakuraDuskColors.onSurfaceVariant,
    outline: SakuraDuskColors.outline,
    outlineVariant: SakuraDuskColors.outlineVariant,
    error: SakuraDuskColors.error,
    onError: SakuraDuskColors.onError,
    playerBar: SakuraDuskColors.playerBar,
    miniPlayerShadow: SakuraDuskColors.miniPlayerShadow,
    accentCyan: SakuraDuskColors.accentCyan,
    accentPink: SakuraDuskColors.accentPink,
  );

  /// 月靛 (Moonlit Indigo) — built-in dark ACG skin. Deep indigo/plum ground;
  /// the sakura-pink brand triplet is intentionally identical to
  /// [sakuraDusk] (brand continuity across skins, per design). Every
  /// neutral text/surface pair below was checked with [contrastRatio]
  /// against WCAG's 4.5:1 body-text floor before landing here: onSurface/
  /// surface 13.9:1, onSurfaceVariant/surface 7.2:1, onBackground/
  /// background 15.0:1, accentCyan/background 10.1:1. The one accepted
  /// compromise is sakuraPink used as *text* directly on this dark ground
  /// (≈3.3–3.5:1, clears the large-text 3:1 floor but not the normal-text
  /// one) — brightening it for that case would break the opposite
  /// constraint (white text painted *on* sakuraPink fills, e.g. filled
  /// buttons, needs sakuraPink to stay dark enough for white to read).
  static const moonlitIndigo = SkinColors(
    sakuraPink: Color(0xFFD42062),
    sakuraPinkLight: Color(0xFFFF8FB8),
    sakuraPinkDark: Color(0xFFBD1550),
    background: Color(0xFF1A1626),
    surface: Color(0xFF241B33),
    surfaceVariant: Color(0xFF2E2140),
    surfaceContainer: Color(0xFF33253F),
    onBackground: Color(0xFFF3E9F5),
    onSurface: Color(0xFFF3E9F5),
    onSurfaceVariant: Color(0xFFB9A4C4),
    outline: Color(0xFF5C4770),
    outlineVariant: Color(0xFF3A2C4A),
    error: Color(0xFFFF6B6B),
    onError: Color(0xFF2A0A0A),
    playerBar: Color(0xFF20182E),
    miniPlayerShadow: Color(0x66000000),
    accentCyan: Color(0xFF5CD3E8),
    accentPink: Color(0xFFFF8FB8),
  );
}

/// A named, switchable color scheme. Two ship in the app ([builtInSkins]);
/// more can be added via Settings → Appearance → 皮肤 → 导入 ([parseSkinJson]).
class SkinDefinition {
  const SkinDefinition({
    required this.id,
    required this.displayName,
    required this.brightness,
    required this.colors,
    this.author,
    this.isBuiltIn = false,
  });

  final String id;
  final String displayName;
  final Brightness brightness;
  final SkinColors colors;
  final String? author;
  final bool isBuiltIn;

  static const sakuraDusk = SkinDefinition(
    id: 'sakura-dusk',
    displayName: '樱花薄暮',
    brightness: Brightness.light,
    colors: SkinColors.sakuraDusk,
    author: 'Inori Music',
    isBuiltIn: true,
  );

  static const moonlitIndigo = SkinDefinition(
    id: 'moonlit-indigo',
    displayName: '月靛',
    brightness: Brightness.dark,
    colors: SkinColors.moonlitIndigo,
    author: 'Inori Music',
    isBuiltIn: true,
  );
}

const builtInSkins = [SkinDefinition.sakuraDusk, SkinDefinition.moonlitIndigo];

/// Generalized form of the theme builder that shipped through v5.13.0 as
/// `buildSakuraDuskTheme()` — same structure, reading colors off whichever
/// [SkinDefinition] is active instead of the static Sakura Dusk constants.
ThemeData buildThemeFromSkin(SkinDefinition skin) {
  final c = skin.colors;
  final colorScheme = skin.brightness == Brightness.dark
      ? ColorScheme.dark(
          primary: c.sakuraPink,
          onPrimary: Colors.white,
          primaryContainer: c.sakuraPinkLight,
          onPrimaryContainer: c.sakuraPinkDark,
          secondary: c.accentCyan,
          onSecondary: Colors.white,
          tertiary: c.accentPink,
          onTertiary: c.onBackground,
          error: c.error,
          onError: c.onError,
          surface: c.surface,
          onSurface: c.onSurface,
          surfaceContainerHighest: c.surfaceVariant,
          outline: c.outline,
          outlineVariant: c.outlineVariant,
          scrim: c.miniPlayerShadow,
        )
      : ColorScheme.light(
          primary: c.sakuraPink,
          onPrimary: Colors.white,
          primaryContainer: c.sakuraPinkLight,
          onPrimaryContainer: c.sakuraPinkDark,
          secondary: c.accentCyan,
          onSecondary: Colors.white,
          tertiary: c.accentPink,
          onTertiary: c.onBackground,
          error: c.error,
          onError: c.onError,
          surface: c.surface,
          onSurface: c.onSurface,
          surfaceContainerHighest: c.surfaceVariant,
          outline: c.outline,
          outlineVariant: c.outlineVariant,
          scrim: c.miniPlayerShadow,
        );

  return ThemeData(
    useMaterial3: true,
    colorScheme: colorScheme,
    scaffoldBackgroundColor: c.background,
    fontFamily: 'Inter',
    appBarTheme: AppBarTheme(
      backgroundColor: c.background,
      foregroundColor: c.onBackground,
      elevation: 0,
      scrolledUnderElevation: 0,
      centerTitle: false,
      titleTextStyle: TextStyle(
        fontFamily: 'Inter',
        fontSize: 18,
        fontWeight: FontWeight.w600,
        color: c.onBackground,
      ),
    ),
    navigationBarTheme: NavigationBarThemeData(
      backgroundColor: c.playerBar,
      indicatorColor: c.sakuraPinkDark,
      elevation: 0,
    ),
    navigationRailTheme: NavigationRailThemeData(
      backgroundColor: c.surface,
      indicatorColor: c.sakuraPinkDark,
      selectedIconTheme: IconThemeData(color: c.sakuraPinkDark),
      unselectedIconTheme: IconThemeData(color: c.onSurfaceVariant),
    ),
    cardTheme: CardThemeData(
      color: c.surfaceVariant,
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: c.outlineVariant, width: 0.5),
      ),
    ),
    listTileTheme: const ListTileThemeData(
      tileColor: Colors.transparent,
      contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 4),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: c.surfaceVariant,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: c.outline),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: c.outline),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: c.sakuraPink, width: 1.5),
      ),
      labelStyle: TextStyle(color: c.onSurfaceVariant),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: c.sakuraPink,
        foregroundColor: Colors.white,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        // Height only — NOT `Size(double.infinity, 48)`. That forced every
        // FilledButton in the app to demand infinite minWidth, which Flutter
        // clamps down to whatever the *tightest* available width happens to
        // be at that particular call site. Inside a loose vertical slot
        // (Column, a dialog's OverflowBar) that clamp lands on the slot's own
        // width, which looks like an intentional full-width button — so most
        // call sites never showed a problem. Two shapes do not survive it:
        // a `ListTile.trailing` slot (the v5.30.7 field report — the
        // trailing widget consumed the tile's *entire* width, leaving
        // title/subtitle a sliver too narrow to hold even one CJK character
        // per line) and a plain `Row` sibling with no `Expanded`/`Flexible`
        // around it (confirmed live in `play_actions_row.dart`'s "Play"
        // button and `local_library_screen.dart`'s "播放全部" button — a Row
        // hands non-flex children an *unbounded* main-axis constraint, so
        // "infinite minWidth" isn't merely clamped there, it trips Flutter's
        // own "BoxConstraints forces an infinite width" layout assertion).
        // A button that genuinely wants full width now says so explicitly at
        // its own call site (`SizedBox(width: double.infinity, child: ...)`
        // or a local `crossAxisAlignment: CrossAxisAlignment.stretch`
        // ancestor, e.g. login_screen.dart's submit button) instead of
        // leaning on a theme default that silently breaks the one layout
        // shape it can't survive.
        minimumSize: const Size(0, 48),
      ),
    ),
    sliderTheme: SliderThemeData(
      activeTrackColor: c.sakuraPink,
      inactiveTrackColor: c.outlineVariant,
      thumbColor: c.sakuraPinkDark,
      overlayColor: c.sakuraPinkDark.withValues(alpha: 0.16),
      trackHeight: 3,
    ),
    dividerTheme: DividerThemeData(color: c.outlineVariant, thickness: 0.5),
    iconTheme: IconThemeData(color: c.onSurfaceVariant),
    textTheme: TextTheme(
      displayLarge: TextStyle(
        color: c.onBackground,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w700,
      ),
      displayMedium: TextStyle(
        color: c.onBackground,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w700,
      ),
      headlineLarge: TextStyle(
        color: c.onBackground,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w700,
      ),
      headlineMedium: TextStyle(
        color: c.onBackground,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w600,
      ),
      headlineSmall: TextStyle(
        color: c.onBackground,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w600,
      ),
      titleLarge: TextStyle(
        color: c.onSurface,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w600,
      ),
      titleMedium: TextStyle(
        color: c.onSurface,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w500,
      ),
      titleSmall: TextStyle(
        color: c.onSurfaceVariant,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w500,
      ),
      bodyLarge: TextStyle(color: c.onSurface, fontFamily: 'Inter'),
      bodyMedium: TextStyle(color: c.onSurface, fontFamily: 'Inter'),
      bodySmall: TextStyle(color: c.onSurfaceVariant, fontFamily: 'Inter'),
      labelLarge: TextStyle(
        color: c.onSurface,
        fontFamily: 'Inter',
        fontWeight: FontWeight.w500,
      ),
      labelMedium: TextStyle(color: c.onSurfaceVariant, fontFamily: 'Inter'),
      labelSmall: TextStyle(color: c.onSurfaceVariant, fontFamily: 'Inter'),
    ),
  );
}

/// WCAG 2.1 contrast ratio between two colors (1.0–21.0). Same relative
/// luminance formula used to hand-check [SkinColors.moonlitIndigo] above.
double contrastRatio(Color a, Color b) {
  double relativeLuminance(Color color) {
    double linearize(double channel) {
      return channel <= 0.03928
          ? channel / 12.92
          : math.pow((channel + 0.055) / 1.055, 2.4).toDouble();
    }

    return 0.2126 * linearize(color.r) +
        0.7152 * linearize(color.g) +
        0.0722 * linearize(color.b);
  }

  final lumA = relativeLuminance(a);
  final lumB = relativeLuminance(b);
  final lighter = math.max(lumA, lumB);
  final darker = math.min(lumA, lumB);
  return (lighter + 0.05) / (darker + 0.05);
}

class SkinParseException implements Exception {
  SkinParseException(this.message);
  final String message;

  @override
  String toString() => message;
}

class SkinImportResult {
  const SkinImportResult({required this.skin, required this.warnings});
  final SkinDefinition skin;
  final List<String> warnings;
}

const _requiredColorKeys = [
  'primary',
  'primaryLight',
  'primaryDark',
  'background',
  'surface',
  'surfaceVariant',
  'surfaceContainer',
  'onBackground',
  'onSurface',
  'onSurfaceVariant',
  'outline',
  'outlineVariant',
  'error',
  'onError',
  'playerBar',
  'shadow',
  'accentCyan',
  'accentPink',
];

/// Parses a user-imported skin manifest (single JSON file — see
/// `docs/skin-format.md`-style inline schema below). Throws
/// [SkinParseException] with a Chinese, user-facing message for anything
/// structurally wrong (missing/unparseable fields); returns contrast
/// warnings (does not throw) for any core text/background pair under the
/// WCAG 4.5:1 floor, per the "warn, don't block" rule for imports.
///
/// Expected shape:
/// ```json
/// {
///   "id": "my-skin",
///   "displayName": "My Skin",
///   "author": "optional",
///   "brightness": "light" | "dark",
///   "colors": { "primary": "#RRGGBB", ... (all keys in _requiredColorKeys) }
/// }
/// ```
SkinImportResult parseSkinJson(
  String source, {
  Set<String> existingIds = const {},
}) {
  final Object? raw;
  try {
    raw = jsonDecode(source);
  } on FormatException catch (e) {
    throw SkinParseException('不是合法的 JSON 文件：${e.message}');
  }
  if (raw is! Map<String, dynamic>) {
    throw SkinParseException('皮肤文件格式错误：根节点必须是一个 JSON 对象');
  }

  final id = raw['id'];
  if (id is! String || id.trim().isEmpty) {
    throw SkinParseException('缺少或无效的 "id" 字段');
  }
  if (existingIds.contains(id)) {
    throw SkinParseException('皮肤 id "$id" 已存在，请修改文件中的 "id" 后重新导入');
  }
  final displayName = raw['displayName'];
  if (displayName is! String || displayName.trim().isEmpty) {
    throw SkinParseException('缺少或无效的 "displayName" 字段');
  }
  final brightnessStr = raw['brightness'];
  if (brightnessStr != 'light' && brightnessStr != 'dark') {
    throw SkinParseException('"brightness" 字段必须是 "light" 或 "dark"');
  }
  final colorsRaw = raw['colors'];
  if (colorsRaw is! Map<String, dynamic>) {
    throw SkinParseException('缺少 "colors" 字段');
  }

  Color parseHex(String key) {
    final value = colorsRaw[key];
    if (value is! String) {
      throw SkinParseException('颜色字段 "$key" 缺失或不是字符串');
    }
    var hex = value.trim();
    if (hex.startsWith('#')) hex = hex.substring(1);
    if (hex.length == 6) hex = 'FF$hex';
    if (hex.length != 8) {
      throw SkinParseException(
        '颜色字段 "$key" 格式错误："$value"（应为 #RRGGBB 或 #AARRGGBB）',
      );
    }
    final parsed = int.tryParse(hex, radix: 16);
    if (parsed == null) {
      throw SkinParseException('颜色字段 "$key" 不是合法的十六进制颜色："$value"');
    }
    return Color(parsed);
  }

  for (final key in _requiredColorKeys) {
    if (!colorsRaw.containsKey(key)) {
      throw SkinParseException('"colors" 缺少必填字段 "$key"');
    }
  }

  final colors = SkinColors(
    sakuraPink: parseHex('primary'),
    sakuraPinkLight: parseHex('primaryLight'),
    sakuraPinkDark: parseHex('primaryDark'),
    background: parseHex('background'),
    surface: parseHex('surface'),
    surfaceVariant: parseHex('surfaceVariant'),
    surfaceContainer: parseHex('surfaceContainer'),
    onBackground: parseHex('onBackground'),
    onSurface: parseHex('onSurface'),
    onSurfaceVariant: parseHex('onSurfaceVariant'),
    outline: parseHex('outline'),
    outlineVariant: parseHex('outlineVariant'),
    error: parseHex('error'),
    onError: parseHex('onError'),
    playerBar: parseHex('playerBar'),
    miniPlayerShadow: parseHex('shadow'),
    accentCyan: parseHex('accentCyan'),
    accentPink: parseHex('accentPink'),
  );

  final author = raw['author'];
  final skin = SkinDefinition(
    id: id,
    displayName: displayName,
    brightness: brightnessStr == 'dark' ? Brightness.dark : Brightness.light,
    colors: colors,
    author: author is String && author.trim().isNotEmpty ? author : null,
  );

  final warnings = <String>[];
  void checkPair(String label, Color fg, Color bg) {
    final ratio = contrastRatio(fg, bg);
    if (ratio < 4.5) {
      warnings.add(
        '$label 对比度仅 ${ratio.toStringAsFixed(2)}:1（建议 ≥ 4.5:1），文字可能难以辨认',
      );
    }
  }

  checkPair('正文文字 / 背景', colors.onBackground, colors.background);
  checkPair('正文文字 / 卡片', colors.onSurface, colors.surface);
  checkPair('次要文字 / 卡片', colors.onSurfaceVariant, colors.surface);
  checkPair('错误提示文字 / 错误底色', colors.onError, colors.error);

  return SkinImportResult(skin: skin, warnings: warnings);
}
