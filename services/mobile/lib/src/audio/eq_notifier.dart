import 'dart:convert';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/audio/eq_settings.dart';
import 'package:inori_music/src/playback/playback_engine.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';

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

  /// The engine's equalizer, or null when it has none.
  ///
  /// Asks the engine rather than testing `Platform.isAndroid`: the platform
  /// was only ever a proxy for "does the current engine wire up an EQ
  /// effect", and it becomes the wrong answer the moment the engine changes.
  EngineEqualizer? get _eq => ref.read(playbackEngineProvider).equalizer;

  Future<void> setEnabled(bool enabled) async {
    final eq = _eq;
    // No equalizer at all — nothing to enable *or* disable. This guard used
    // to cover only the enable path, so a disable call still reached the
    // platform channel (fixed in v5.20.2).
    if (eq == null) return;
    state = state.copyWith(enabled: enabled);
    await eq.setEnabled(enabled);
    if (enabled) {
      await _applyEqualizerBands(state.bands);
    } else {
      await _resetEqualizerBands();
    }
    await _persist();
  }

  Future<void> setPreset(String presetKey) async {
    final bands = eqPresets[presetKey];
    if (bands == null) return;
    state = state.copyWith(preset: presetKey, bands: List<double>.from(bands));
    if (state.enabled) {
      await _applyEqualizerBands(state.bands);
    }
    await _persist();
  }

  Future<void> setBand(int index, double gainDb) async {
    if (index < 0 || index >= 10) return;
    final updated = List<double>.from(state.bands);
    updated[index] = gainDb;
    state = state.copyWith(bands: updated, preset: 'custom');
    if (state.enabled) {
      await _applyEqualizerBands(state.bands);
    }
    await _persist();
  }

  /// Save the current band configuration as a named custom preset.
  Future<void> saveCurrentAsPreset(String name) async {
    final trimmed = name.trim();
    if (trimmed.isEmpty) return;
    final updated = Map<String, List<double>>.from(state.customPresets);
    updated[trimmed] = List<double>.from(state.bands);
    state = state.copyWith(customPresets: updated, preset: trimmed);
    await _persist();
  }

  /// Switch to a previously saved custom preset.
  Future<void> selectCustomPreset(String name) async {
    final bands = state.customPresets[name];
    if (bands == null) return;
    state = state.copyWith(preset: name, bands: List<double>.from(bands));
    if (state.enabled) {
      await _applyEqualizerBands(state.bands);
    }
    await _persist();
  }

  /// Delete a saved custom preset. Falls back to 'flat' if it was selected.
  Future<void> deleteCustomPreset(String name) async {
    if (!state.customPresets.containsKey(name)) return;
    final updated = Map<String, List<double>>.from(state.customPresets)
      ..remove(name);
    final wasSelected = state.preset == name;
    state = state.copyWith(
      customPresets: updated,
      preset: wasSelected ? 'flat' : state.preset,
      bands: wasSelected ? List<double>.from(eqPresets['flat']!) : state.bands,
    );
    if (wasSelected && state.enabled) {
      await _applyEqualizerBands(state.bands);
    }
    await _persist();
  }

  // ---- Equalizer device I/O ----

  /// Map the UI's 10 fixed bands onto the device's actual band count
  /// (commonly 5 on Android) via nearest-neighbor, and push gains.
  Future<void> _applyEqualizerBands(List<double> bands) async {
    final eq = _eq;
    if (eq == null) return;
    final count = await eq.bandCount();
    if (count == 0) return;
    final range = await eq.gainRange();
    for (var i = 0; i < count; i++) {
      final uiIdx = (i * bands.length / count).floor();
      await eq.setBandGain(i, bands[uiIdx].clamp(range.min, range.max));
    }
  }

  Future<void> _resetEqualizerBands() async {
    final eq = _eq;
    if (eq == null) return;
    final count = await eq.bandCount();
    for (var i = 0; i < count; i++) {
      await eq.setBandGain(i, 0);
    }
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
      final rawCustom = map['customPresets'] as Map<String, dynamic>?;
      final customPresets = <String, List<double>>{
        if (rawCustom != null)
          for (final entry in rawCustom.entries)
            entry.key: (entry.value as List<dynamic>)
                .map((e) => (e as num).toDouble())
                .toList(),
      };
      final effectiveEnabled = enabled && _eq != null;
      state = EqSettings(
        enabled: effectiveEnabled,
        bands: bands,
        preset: preset,
        customPresets: customPresets,
      );
      if (effectiveEnabled) {
        await _eq?.setEnabled(true);
        await _applyEqualizerBands(bands);
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
          'customPresets': state.customPresets,
        }),
      );
    } catch (_) {}
  }
}
