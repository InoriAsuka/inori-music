import 'dart:io';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/shared/theme/skin_definition.dart';

const _kSelectedSkinIdKey = 'appearance.skinId';

class SkinState {
  const SkinState({required this.installed, required this.selectedId});

  /// Built-ins first, then imported skins in the order they were added.
  final List<SkinDefinition> installed;
  final String selectedId;

  SkinDefinition get active {
    for (final skin in installed) {
      if (skin.id == selectedId) return skin;
    }
    return SkinDefinition.sakuraDusk;
  }
}

final skinProvider = NotifierProvider<SkinNotifier, SkinState>(SkinNotifier.new);

/// Owns the installed-skins registry (built-ins + anything imported via
/// Settings → Appearance → 皮肤 → 导入) and which one is active. Imported
/// skins persist as individual JSON files under the app support directory's
/// `skins/` folder — a skin is "a file", matching how a single-file manifest
/// was designed to be shareable in the first place.
class SkinNotifier extends Notifier<SkinState> {
  @override
  SkinState build() {
    _restore();
    return SkinState(installed: builtInSkins, selectedId: SkinDefinition.sakuraDusk.id);
  }

  Future<Directory> _skinsDir() async {
    final dir = await getApplicationSupportDirectory();
    final skinsDir = Directory(p.join(dir.path, 'skins'));
    if (!skinsDir.existsSync()) skinsDir.createSync(recursive: true);
    return skinsDir;
  }

  Future<void> _restore() async {
    final imported = <SkinDefinition>[];
    try {
      final dir = await _skinsDir();
      for (final entity in dir.listSync()) {
        if (entity is! File || !entity.path.endsWith('.json')) continue;
        try {
          imported.add(parseSkinJson(await entity.readAsString()).skin);
        } catch (_) {
          // A hand-edited-into-corruption file shouldn't block startup —
          // just skip it; it stays on disk in case the user wants to fix it.
        }
      }
    } catch (_) {
      // No skins directory yet (fresh install) — built-ins only.
    }

    final installed = [...builtInSkins, ...imported];
    final prefs = await SharedPreferences.getInstance();
    final storedId = prefs.getString(_kSelectedSkinIdKey);
    final selectedId = installed.any((s) => s.id == storedId) ? storedId! : SkinDefinition.sakuraDusk.id;
    state = SkinState(installed: installed, selectedId: selectedId);
  }

  Future<void> selectSkin(String id) async {
    if (!state.installed.any((s) => s.id == id)) return;
    state = SkinState(installed: state.installed, selectedId: id);
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kSelectedSkinIdKey, id);
  }

  /// Validates and installs a skin manifest's raw JSON text. Returns any
  /// non-blocking contrast warnings; throws [SkinParseException] (caught by
  /// the caller and shown as a dialog) if the file is structurally invalid.
  Future<List<String>> importSkin(String jsonSource) async {
    final existingIds = state.installed.map((s) => s.id).toSet();
    final result = parseSkinJson(jsonSource, existingIds: existingIds);

    final dir = await _skinsDir();
    final file = File(p.join(dir.path, '${result.skin.id}.json'));
    await file.writeAsString(jsonSource);

    state = SkinState(installed: [...state.installed, result.skin], selectedId: state.selectedId);
    return result.warnings;
  }

  /// Removes a previously-imported skin. Built-ins can't be deleted. If the
  /// deleted skin was active, falls back to the default Sakura Dusk skin.
  Future<void> deleteSkin(String id) async {
    SkinDefinition? target;
    for (final skin in state.installed) {
      if (skin.id == id) target = skin;
    }
    if (target == null || target.isBuiltIn) return;

    final dir = await _skinsDir();
    final file = File(p.join(dir.path, '$id.json'));
    if (file.existsSync()) await file.delete();

    final installed = state.installed.where((s) => s.id != id).toList();
    final wasActive = state.selectedId == id;
    final selectedId = wasActive ? SkinDefinition.sakuraDusk.id : state.selectedId;
    state = SkinState(installed: installed, selectedId: selectedId);
    if (wasActive) {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_kSelectedSkinIdKey, selectedId);
    }
  }

  /// Opens a file picker for a `.json` skin manifest and imports it.
  /// Returns null if the user cancelled the picker.
  Future<List<String>?> pickAndImportSkin() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: const ['json'],
    );
    final path = result?.files.single.path;
    if (path == null) return null;
    final jsonSource = await File(path).readAsString();
    return importSkin(jsonSource);
  }
}

/// Exposes the active [SkinDefinition] to the whole widget tree below the
/// single insertion point in `main.dart`, so any widget — Consumer or plain
/// Stateless — can read live skin colors via `context.skinColors` without
/// needing to be rebuilt into a Riverpod Consumer just for this.
class SkinScope extends InheritedWidget {
  const SkinScope({super.key, required this.skin, required super.child});

  final SkinDefinition skin;

  static SkinColors of(BuildContext context) {
    final scope = context.dependOnInheritedWidgetOfExactType<SkinScope>();
    // Falls back to the default skin rather than asserting — widget tests
    // routinely pump an isolated widget under a bare MaterialApp with no
    // SkinScope ancestor, and should still render something sensible.
    return (scope?.skin ?? SkinDefinition.sakuraDusk).colors;
  }

  @override
  bool updateShouldNotify(SkinScope oldWidget) => oldWidget.skin.id != skin.id;
}

extension SkinColorsContext on BuildContext {
  SkinColors get skinColors => SkinScope.of(this);
}
