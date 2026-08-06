import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/shared/background_provider.dart';
import 'package:inori_music/src/shared/theme/sakura_dusk.dart';

/// Backdrop for the login/splash screens. Renders the user's custom image
/// (if set) at their chosen opacity, or the plain themed background
/// otherwise — either way, [child] (the glass card / brand mark) is meant
/// to float on top, not this widget's problem to keep readable on its own.
class AppBackground extends ConsumerWidget {
  const AppBackground({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(backgroundProvider);
    final path = settings.imagePath;
    final hasImage = path != null && File(path).existsSync();

    return Stack(
      fit: StackFit.expand,
      children: [
        ColoredBox(color: SakuraDuskColors.background),
        if (hasImage)
          Opacity(
            opacity: settings.opacity,
            child: Image.file(
              File(path),
              fit: BoxFit.cover,
              errorBuilder: (_, _, _) => const SizedBox.shrink(),
            ),
          ),
        child,
      ],
    );
  }
}
