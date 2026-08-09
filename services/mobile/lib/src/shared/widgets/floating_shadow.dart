import 'package:flutter/material.dart';

/// Two-layer "floating panel" shadow in the Apple style — large blur radii
/// and low, skin-derived alpha — used in place of Material's `elevation`,
/// which paints one tight, high-contrast shadow that is always plain black
/// no matter which skin is active. The v5.30.5 field report called this out
/// by name against Apple Music/macOS panels ("和苹果相差一节"): the mini
/// player bar's `Material(elevation: 8)` and every `GlassPanel` (sidebar,
/// the full player's control/side panels) all shared this same tight-and-
/// black look.
///
/// A single [BoxShadow] cannot read as both "clearly grounded" and "soft and
/// diffuse" at once — a blur wide enough to feel airy (~28+) barely
/// registers right at the panel's own edge, and a blur tight enough to read
/// as contact (~10) does not spread. Real elevated surfaces stack two: a
/// close, slightly stronger layer for the edge, and a large, faint one for
/// the ambient spread beneath it — this app's own pre-existing
/// `_GlassCard` shadow in `login_screen.dart` already used this exact
/// outer-Container-plus-inner-ClipRRect shape for a single layer; this just
/// generalises it to two and shares it with every other floating panel.
///
/// [base] should be a skin's own [SkinColors.miniPlayerShadow] rather than a
/// literal black — Sakura Dusk's token is a low-alpha dark plum tuned for
/// its cream background (a literal black at the same alpha reads as soot on
/// a warm surface), while Moonlit Indigo's is a much higher-alpha black,
/// since a dark ground needs a stronger shadow to register as separation at
/// all. Deriving both layers' alpha *relative to* [base]'s own alpha —
/// instead of hardcoding one absolute alpha for every skin — keeps each skin
/// inside the range it was already tuned for, rather than picking a single
/// constant that would only suit one of them.
List<BoxShadow> floatingShadow(Color base) {
  // The skin author's own judgment of "how strong does a shadow need to be
  // to register against this skin's background" — Sakura Dusk: 0x26/255 ≈
  // 0.15; Moonlit Indigo: 0x66/255 = 0.4.
  final strength = base.a;
  return [
    // Tight contact shadow: close blur, small downward offset — reads as
    // the panel meeting whatever sits directly behind it. Deliberately
    // outside the Apple-reference 24-32 blur band below; that band
    // describes the *ambient* layer, and this one exists specifically to be
    // tighter than that.
    BoxShadow(
      color: base.withValues(alpha: (strength * 1.3).clamp(0.0, 1.0)),
      blurRadius: 10,
      offset: const Offset(0, 3),
    ),
    // Wide ambient spread: this is the layer that actually reads as
    // "floating" the way Apple Music/macOS panels do. Material's own
    // elevation shadow has no equivalent to this layer, which is exactly
    // why it looked tight and hard by comparison. 0.75x lands Sakura Dusk's
    // resulting alpha (~0.11) inside the 0.10-0.18 reference band; Moonlit
    // Indigo's (~0.30) sits above it on purpose, per this function's own doc
    // comment on why a dark ground needs more.
    BoxShadow(
      color: base.withValues(alpha: (strength * 0.75).clamp(0.0, 1.0)),
      blurRadius: 28,
      offset: const Offset(0, 8),
    ),
  ];
}
