import 'dart:ui' show ImageFilter;

import 'package:flutter/material.dart';

import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/floating_shadow.dart';

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
    this.borderRadiusOverride,
  });

  final Widget child;
  final EdgeInsetsGeometry padding;
  final double borderRadius;

  /// Kept well below the backdrop's own 64 — this blurs what is *already*
  /// blurred, so it only needs to add the sense of a second surface.
  final double blurSigma;

  /// Escape hatch for a panel that needs *different* rounding per corner —
  /// added v5.30.7 for the desktop sidebar, whose top-left corner has to
  /// match the macOS window's own corner radius once the panel sits flush
  /// against it (see `_DesktopSidebar` in shell_scaffold.dart), while its
  /// other three corners keep the panel's usual rounding. `null` (the
  /// default) keeps every existing call site — which only ever wanted one
  /// uniform [borderRadius] — completely unchanged; this is additive, not a
  /// replacement for the simple case.
  final BorderRadius? borderRadiusOverride;

  @override
  Widget build(BuildContext context) {
    final radius = borderRadiusOverride ?? BorderRadius.circular(borderRadius);
    // The shadow lives on this outer DecoratedBox rather than inside the
    // ClipRRect below — ClipRRect clips everything painted within it,
    // including a shadow drawn by whatever it wraps, so a shadow added to
    // the Material or the ClipRRect itself would simply be cut off at the
    // panel's own edge instead of spreading past it. See floatingShadow's
    // doc comment for why this is two BoxShadow layers instead of Material's
    // single `elevation` shadow.
    return DecoratedBox(
      decoration: BoxDecoration(
        borderRadius: radius,
        boxShadow: floatingShadow(context.skinColors.miniPlayerShadow),
      ),
      child: ClipRRect(
        borderRadius: radius,
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: blurSigma, sigmaY: blurSigma),
          // Material, not a DecoratedBox: ListTile and friends paint their
          // background and ink onto the nearest Material ancestor, so a
          // coloured box in between swallows both. Exactly the defect the
          // sidebar had in v5.22.0, and panels are full of list rows.
          child: Material(
            color: context.skinColors.surface,
            shape: RoundedRectangleBorder(
              borderRadius: radius,
              side: BorderSide(
                color: context.skinColors.outlineVariant,
                width: 0.8,
              ),
            ),
            child: Padding(padding: padding, child: child),
          ),
        ),
      ),
    );
  }
}
