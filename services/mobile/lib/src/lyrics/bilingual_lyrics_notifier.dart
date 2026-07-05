import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _kBilingualLyricsKey = 'lyrics.bilingual';

final bilingualLyricsProvider = NotifierProvider<BilingualLyricsNotifier, bool>(
  BilingualLyricsNotifier.new,
);

/// Persists the bilingual (translation) lyrics display toggle.
class BilingualLyricsNotifier extends Notifier<bool> {
  @override
  bool build() {
    _restore();
    return false;
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    state = prefs.getBool(_kBilingualLyricsKey) ?? false;
  }

  Future<void> setEnabled(bool enabled) async {
    state = enabled;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_kBilingualLyricsKey, enabled);
  }
}
