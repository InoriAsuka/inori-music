import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/shared/theme/skin_definition.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';

String _validSkinJson({
  String id = 'test-skin',
  String brightness = 'light',
  String? author = 'tester',
  Map<String, String>? colorOverrides,
  List<String> omitColorKeys = const [],
}) {
  final colors = <String, String>{
    'primary': '#D42062',
    'primaryLight': '#FF8FB8',
    'primaryDark': '#BD1550',
    'background': '#FFF7F2',
    'surface': '#FFFFFF',
    'surfaceVariant': '#FFF0F5',
    'surfaceContainer': '#FCE7EF',
    'onBackground': '#3B2A3F',
    'onSurface': '#3B2A3F',
    'onSurfaceVariant': '#6B5570',
    'outline': '#B77A96',
    'outlineVariant': '#F2D9E4',
    'error': '#C81E2C',
    'onError': '#FFFFFF',
    'playerBar': '#FFF0F5',
    'shadow': '#263B2A3F',
    'accentCyan': '#0A7D94',
    'accentPink': '#FF8FB8',
    ...?colorOverrides,
  };
  for (final key in omitColorKeys) {
    colors.remove(key);
  }
  final colorsJson = colors.entries.map((e) => '"${e.key}": "${e.value}"').join(',');
  final authorJson = author == null ? '' : '"author": "$author", ';
  return '{"id": "$id", "displayName": "Test Skin", $authorJson"brightness": "$brightness", "colors": {$colorsJson}}';
}

void main() {
  group('contrastRatio', () {
    test('black on white is the maximum 21:1', () {
      expect(contrastRatio(Colors.black, Colors.white), closeTo(21.0, 0.01));
    });

    test('a color against itself is 1:1', () {
      expect(contrastRatio(Colors.red, Colors.red), closeTo(1.0, 0.01));
    });

    test('is symmetric regardless of argument order', () {
      final a = contrastRatio(Colors.black, Colors.yellow);
      final b = contrastRatio(Colors.yellow, Colors.black);
      expect(a, closeTo(b, 0.001));
    });
  });

  group('buildThemeFromSkin', () {
    test('light skin produces a light ColorScheme with mapped tokens', () {
      final theme = buildThemeFromSkin(SkinDefinition.sakuraDusk);
      expect(theme.colorScheme.brightness, Brightness.light);
      expect(theme.colorScheme.primary, SkinColors.sakuraDusk.sakuraPink);
      expect(theme.scaffoldBackgroundColor, SkinColors.sakuraDusk.background);
      expect(theme.colorScheme.surface, SkinColors.sakuraDusk.surface);
    });

    test('dark skin produces a dark ColorScheme with mapped tokens', () {
      final theme = buildThemeFromSkin(SkinDefinition.moonlitIndigo);
      expect(theme.colorScheme.brightness, Brightness.dark);
      expect(theme.scaffoldBackgroundColor, SkinColors.moonlitIndigo.background);
      expect(theme.colorScheme.surface, SkinColors.moonlitIndigo.surface);
    });

    test('built-in dark skin clears WCAG 4.5:1 on core text/surface pairs', () {
      const c = SkinColors.moonlitIndigo;
      expect(contrastRatio(c.onBackground, c.background), greaterThanOrEqualTo(4.5));
      expect(contrastRatio(c.onSurface, c.surface), greaterThanOrEqualTo(4.5));
      expect(contrastRatio(c.onSurfaceVariant, c.surface), greaterThanOrEqualTo(4.5));
      expect(contrastRatio(c.onError, c.error), greaterThanOrEqualTo(4.5));
    });
  });

  group('parseSkinJson — valid input', () {
    test('parses a well-formed manifest into a SkinDefinition', () {
      final result = parseSkinJson(_validSkinJson());
      expect(result.skin.id, 'test-skin');
      expect(result.skin.displayName, 'Test Skin');
      expect(result.skin.author, 'tester');
      expect(result.skin.brightness, Brightness.light);
      expect(result.skin.isBuiltIn, isFalse);
      expect(result.skin.colors.sakuraPink, const Color(0xFFD42062));
      expect(result.warnings, isEmpty);
    });

    test('accepts 6-digit hex (no alpha) and defaults to opaque', () {
      final result = parseSkinJson(_validSkinJson(colorOverrides: {'background': '#123456'}));
      expect(result.skin.colors.background, const Color(0xFF123456));
    });

    test('accepts 8-digit ARGB hex', () {
      final result = parseSkinJson(_validSkinJson(colorOverrides: {'shadow': '#8000FF00'}));
      expect(result.skin.colors.miniPlayerShadow, const Color(0x8000FF00));
    });

    test('dark brightness maps correctly', () {
      final result = parseSkinJson(_validSkinJson(brightness: 'dark'));
      expect(result.skin.brightness, Brightness.dark);
    });

    test('missing author is left null rather than an empty string', () {
      final result = parseSkinJson(_validSkinJson(author: null));
      expect(result.skin.author, isNull);
    });
  });

  group('parseSkinJson — structural errors', () {
    test('throws on invalid JSON', () {
      expect(() => parseSkinJson('not json'), throwsA(isA<SkinParseException>()));
    });

    test('throws when root is not an object', () {
      expect(() => parseSkinJson('[1, 2, 3]'), throwsA(isA<SkinParseException>()));
    });

    test('throws when id is missing', () {
      const json = '{"displayName": "Test", "brightness": "light", "colors": {}}';
      expect(() => parseSkinJson(json), throwsA(isA<SkinParseException>()));
    });

    test('throws when id collides with an existing skin', () {
      expect(
        () => parseSkinJson(_validSkinJson(id: 'sakura-dusk'), existingIds: {'sakura-dusk'}),
        throwsA(isA<SkinParseException>()),
      );
    });

    test('throws when brightness is neither light nor dark', () {
      final json = _validSkinJson(brightness: 'sepia');
      expect(() => parseSkinJson(json), throwsA(isA<SkinParseException>()));
    });

    test('throws when a required color key is missing', () {
      final json = _validSkinJson(omitColorKeys: ['accentPink']);
      expect(() => parseSkinJson(json), throwsA(isA<SkinParseException>()));
    });

    test('throws on an unparseable hex value', () {
      final json = _validSkinJson(colorOverrides: {'background': 'not-a-color'});
      expect(() => parseSkinJson(json), throwsA(isA<SkinParseException>()));
    });

    test('throws on a hex string of the wrong length', () {
      final json = _validSkinJson(colorOverrides: {'background': '#FFF'});
      expect(() => parseSkinJson(json), throwsA(isA<SkinParseException>()));
    });
  });

  group('parseSkinJson — WCAG contrast warnings', () {
    test('low-contrast onSurface/surface produces a warning but still imports', () {
      final result = parseSkinJson(_validSkinJson(colorOverrides: {
        'surface': '#202020',
        'onSurface': '#303030',
      }));
      expect(result.skin.colors.onSurface, const Color(0xFF303030));
      expect(result.warnings, isNotEmpty);
      expect(result.warnings.any((w) => w.contains('卡片')), isTrue);
    });

    test('a fully legible manifest produces no warnings', () {
      final result = parseSkinJson(_validSkinJson());
      expect(result.warnings, isEmpty);
    });
  });

  group('SkinState.active', () {
    test('resolves the selected id among installed skins', () {
      const state = SkinState(installed: builtInSkins, selectedId: 'moonlit-indigo');
      expect(state.active.id, 'moonlit-indigo');
    });

    test('falls back to Sakura Dusk when selectedId matches nothing installed', () {
      const state = SkinState(installed: builtInSkins, selectedId: 'does-not-exist');
      expect(state.active.id, SkinDefinition.sakuraDusk.id);
    });
  });

  group('SkinScope.of', () {
    testWidgets('falls back to the default skin colors with no ancestor SkinScope', (tester) async {
      late SkinColors resolved;
      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) {
              resolved = context.skinColors;
              return const SizedBox();
            },
          ),
        ),
      );
      expect(resolved.sakuraPink, SkinColors.sakuraDusk.sakuraPink);
    });

    testWidgets('resolves the provided skin when wrapped in SkinScope', (tester) async {
      late SkinColors resolved;
      await tester.pumpWidget(
        MaterialApp(
          home: SkinScope(
            skin: SkinDefinition.moonlitIndigo,
            child: Builder(
              builder: (context) {
                resolved = context.skinColors;
                return const SizedBox();
              },
            ),
          ),
        ),
      );
      expect(resolved.background, SkinColors.moonlitIndigo.background);
    });
  });
}
