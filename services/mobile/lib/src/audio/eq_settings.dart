/// EQ band center frequencies (Hz) for the 10-band equalizer.
const eqBandFrequencies = [31, 62, 125, 250, 500, 1000, 2000, 4000, 8000, 16000];

/// Named EQ presets. Each list contains 10 gain values in dB, one per band.
const eqPresets = <String, List<double>>{
  'flat': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
  'bassBoost': [6, 5, 4, 2, 0, 0, 0, 0, 0, 0],
  'vocal': [0, 0, 0, 2, 4, 4, 3, 2, 0, 0],
  'electronic': [4, 3, 0, -2, 0, 2, 3, 3, 4, 4],
};

/// Immutable EQ configuration state.
class EqSettings {
  const EqSettings({
    required this.enabled,
    required this.bands,
    required this.preset,
    this.customPresets = const {},
  });

  final bool enabled;

  /// Gain values in dB for each of the 10 bands.
  final List<double> bands;

  /// Key into [eqPresets], a key into [customPresets], or `'custom'` for an
  /// unsaved ad-hoc adjustment.
  final String preset;

  /// User-saved presets, keyed by user-chosen name.
  final Map<String, List<double>> customPresets;

  factory EqSettings.defaults() => EqSettings(
        enabled: false,
        bands: List<double>.from(eqPresets['flat']!),
        preset: 'flat',
        customPresets: const {},
      );

  EqSettings copyWith({
    bool? enabled,
    List<double>? bands,
    String? preset,
    Map<String, List<double>>? customPresets,
  }) =>
      EqSettings(
        enabled: enabled ?? this.enabled,
        bands: bands ?? List<double>.from(this.bands),
        preset: preset ?? this.preset,
        customPresets: customPresets ?? this.customPresets,
      );
}
