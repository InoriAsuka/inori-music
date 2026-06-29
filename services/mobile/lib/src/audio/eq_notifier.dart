import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'dart:convert';

import 'package:inori_music/main.dart' show audioHandler;
import 'package:inori_music/src/audio/eq_settings.dart';

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

final eqNotifierProvider = NotifierProvider<EqNotifier, EqSettings>(
  EqNotifier.new,
);

// ---------------------------------------------------------------------------
// Notifier
// ---------------------------------------------------------------------------

class EqNotifier extends Notifier<EqSettings> {
  static const _prefKey = 'audio.eq';

  @override
  EqSettings build() {
    _restore();
    return EqSettings.defaults();
  }

  // ---- Public API ----

  Future<void> setEnabled(bool enabled) async {
    state = state.copyWith(enabled: enabled);
    if (enabled) {
      audioHandler.applyEqualizerBands(state.bands);
    } else {
      audioHandler.resetEqualizer();
    }
    await _persist();
  }

  Future<void> setPreset(String presetKey) async {
    final bands = eqPresets[presetKey];
    if (bands == null) return;
    state = state.copyWith(preset: presetKey, bands: List<double>.from(bands));
    if (state.enabled) {
      audioHandler.applyEqualizerBands(state.bands);
    }
    await _persist();
  }

  Future<void> setBand(int index, double gainDb) async {
    if (index < 0 || index >= 10) return;
    final updated = List<double>.from(state.bands);
    updated[index] = gainDb;
    state = state.copyWith(bands: updated, preset: 'custom');
    if (state.enabled) {
      audioHandler.applyEqualizerBands(state.bands);
    }
    await _persist();
  }

  // ---- Persistence ----

  Future<void> _restore() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final raw = prefs.getString(_prefKey);
      if (raw == null) return;
      final map = jsonDecode(raw) as Map<String, dynamic>;
      final enabled = map['enabled'] as bool? ?? false;
      final preset = map['preset'] as String? ?? 'flat';
      final rawBands = map['bands'] as List<dynamic>?;
      final bands = rawBands != null
          ? rawBands.map((e) => (e as num).toDouble()).toList()
          : List<double>.from(eqPresets['flat']!);
      state = EqSettings(enabled: enabled, bands: bands, preset: preset);
      if (enabled) {
        audioHandler.applyEqualizerBands(bands);
      }
    } catch (_) {
      // Ignore corrupt prefs.
    }
  }

  Future<void> _persist() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(
        _prefKey,
        jsonEncode({
          'enabled': state.enabled,
          'preset': state.preset,
          'bands': state.bands,
        }),
      );
    } catch (_) {}
  }
}
