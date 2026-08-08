import 'dart:ui' show ImageFilter;

import 'package:flutter/material.dart';

import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Frosted pane for content sitting over the cover-art backdrop.
///
/// Three things together make this read as glass rather than as a
/// semi-transparent box: the backdrop blur, a fill light enough to let the
/// colour field through, and a hairline top-lit border. Drop any one of them
/// and it flattens out.
///
/// The colours come from skin tokens, so over the artwork backdrop this picks
/// up the translucent whites of the derived overlay skin, and on a plain
/// screen it degrades to an ordinary surface card with no special casing at
/// the call site.
class GlassPanel extends StatelessWidget {
  const GlassPanel({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(16),
    this.borderRadius = 20,
    this.blurSigma = 18,
  });

  final Widget child;
  final EdgeInsetsGeometry padding;
  final double borderRadius;

  /// Kept well below the backdrop's own 64 — this blurs what is *already*
  /// blurred, so it only needs to add the sense of a second surface.
  final double blurSigma;

  @override
  Widget build(BuildContext context) {
    final radius = BorderRadius.circular(borderRadius);
    return ClipRRect(
      borderRadius: radius,
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: blurSigma, sigmaY: blurSigma),
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: context.skinColors.surface,
            borderRadius: radius,
            border: Border.all(
              color: context.skinColors.outlineVariant,
              width: 0.8,
            ),
          ),
          child: Padding(padding: padding, child: child),
        ),
      ),
    );
  }
}
