import 'dart:io';

import 'package:file_picker/file_picker.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _kBackgroundPathKey = 'appearance.backgroundImagePath';
const _kBackgroundOpacityKey = 'appearance.backgroundOpacity';

class BackgroundSettings {
  const BackgroundSettings({this.imagePath, this.opacity = 0.45});

  /// Path to the user's custom login-screen background image, copied into
  /// app storage at pick time. Null means the plain themed background.
  final String? imagePath;

  /// How visible the image is behind the login card (0–1). Capped in the UI
  /// (see settings_screen.dart) so text over it never drops below the
  /// contrast the glass card guarantees.
  final double opacity;

  BackgroundSettings copyWith({String? imagePath, bool clearImage = false, double? opacity}) {
    return BackgroundSettings(
      imagePath: clearImage ? null : (imagePath ?? this.imagePath),
      opacity: opacity ?? this.opacity,
    );
  }
}

final backgroundProvider = NotifierProvider<BackgroundNotifier, BackgroundSettings>(
  BackgroundNotifier.new,
);

/// Custom background image for the login/splash screens (not the whole
/// app — the rest of the UI is built on opaque Sakura Dusk surfaces and
/// re-auditing every existing screen for contrast over an arbitrary user
/// photo is a much bigger scope than "let me pick a nicer login backdrop").
class BackgroundNotifier extends Notifier<BackgroundSettings> {
  @override
  BackgroundSettings build() {
    _restore();
    return const BackgroundSettings();
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    final path = prefs.getString(_kBackgroundPathKey);
    final opacity = prefs.getDouble(_kBackgroundOpacityKey) ?? 0.45;
    if (path != null && File(path).existsSync()) {
      state = BackgroundSettings(imagePath: path, opacity: opacity);
    } else {
      state = BackgroundSettings(opacity: opacity);
    }
  }

  Future<void> pickImage() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: const ['png', 'jpg', 'jpeg', 'webp'],
    );
    final path = result?.files.single.path;
    if (path == null) return;

    final dir = await getApplicationSupportDirectory();
    final bgDir = Directory(p.join(dir.path, 'appearance'));
    if (!bgDir.existsSync()) bgDir.createSync(recursive: true);
    // Only one background is ever active — clear any previous copy first so
    // changing the extension (e.g. png -> jpg) doesn't leave an orphan file.
    if (bgDir.existsSync()) {
      for (final entity in bgDir.listSync()) {
        if (entity is File && p.basenameWithoutExtension(entity.path) == 'background') {
          await entity.delete();
        }
      }
    }
    final dest = File(p.join(bgDir.path, 'background${p.extension(path)}'));
    await File(path).copy(dest.path);

    state = state.copyWith(imagePath: dest.path);
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kBackgroundPathKey, dest.path);
  }

  Future<void> clearImage() async {
    final path = state.imagePath;
    if (path != null) {
      final f = File(path);
      if (f.existsSync()) await f.delete();
    }
    state = state.copyWith(clearImage: true);
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_kBackgroundPathKey);
  }

  Future<void> setOpacity(double value) async {
    final clamped = value.clamp(0.1, 0.8);
    state = state.copyWith(opacity: clamped);
    final prefs = await SharedPreferences.getInstance();
    await prefs.setDouble(_kBackgroundOpacityKey, clamped);
  }
}
