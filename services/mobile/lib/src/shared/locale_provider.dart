import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Currently active locale. Defaults to the device locale.
/// Updated by the language picker in SettingsScreen.
final localeProvider = StateProvider<Locale?>((ref) => null);
