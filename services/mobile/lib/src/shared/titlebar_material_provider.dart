import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Windows-only (see `DesktopAppBar`/Settings — macOS has its own vibrancy
/// conventions and isn't in scope here). Backed by `flutter_acrylic`'s
/// `WindowEffect.disabled/.acrylic/.mica`.
enum TitlebarMaterial { none, acrylic, mica }

const _kTitlebarMaterialKey = 'appearance.titlebarMaterial';

final titlebarMaterialProvider =
    NotifierProvider<TitlebarMaterialNotifier, TitlebarMaterial>(
      TitlebarMaterialNotifier.new,
    );

/// Persists which translucency material (if any) is applied to the title
/// bar area by [DesktopIntegration.applyTitlebarMaterial].
class TitlebarMaterialNotifier extends Notifier<TitlebarMaterial> {
  @override
  TitlebarMaterial build() {
    _restore();
    return TitlebarMaterial.none;
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_kTitlebarMaterialKey);
    state = TitlebarMaterial.values.firstWhere(
      (m) => m.name == raw,
      orElse: () => TitlebarMaterial.none,
    );
  }

  Future<void> setMaterial(TitlebarMaterial material) async {
    state = material;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kTitlebarMaterialKey, material.name);
  }
}
