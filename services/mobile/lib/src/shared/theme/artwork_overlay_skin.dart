import 'package:flutter/material.dart';

import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/shared/theme/skin_definition.dart';

/// Derives the skin to use for content sitting **on top of** the cover-art
/// backdrop ([CoverFluidBackground]).
///
/// The player and lyrics screens were written against flat skin surfaces —
/// dark ink on a near-white background under Sakura Dusk. Dropping that
/// unchanged onto a saturated, moving colour field makes it unreadable. Rather
/// than rewriting ~30 call sites to special-case the backdrop, this returns a
/// *derived skin*, and the screen wraps its content in a `SkinScope` carrying
/// it: every existing `context.skinColors.x` inside then resolves to the
/// overlay value with no call-site changes. This is exactly what the v5.14.0
/// token migration made possible.
///
/// Two things change:
///
/// * **Foregrounds go light and surfaces go glassy.** Text, icons and outlines
///   become white at varying opacity; the surface tokens become translucent
///   blacks so cards read as frosted panes over the artwork instead of opaque
///   blocks hiding it.
/// * **The accent follows the artwork.** [palette] replaces the skin's brand
///   pink, so the progress bar, the play button and every selected state take
///   their colour from the cover. When there's no palette yet the skin's own
///   accent stays, so nothing flashes through an uncoloured frame.
SkinDefinition artworkOverlaySkin(
  SkinDefinition base, {
  required CoverPalette? palette,
}) {
  // Brightness stays dark regardless of the underlying skin: the scrim over
  // the artwork is dark by construction, so a light skin's ColorScheme would
  // hand Material the wrong defaults for anything not covered by a token.
  final accent = palette?.accentOverArtwork ?? base.colors.sakuraPink;

  return SkinDefinition(
    // Distinct id because SkinScope.updateShouldNotify compares ids — without
    // the accent in the key, changing tracks would not propagate the new
    // accent to descendants.
    id: '${base.id}#artwork-${accent.toARGB32()}',
    displayName: base.displayName,
    brightness: Brightness.dark,
    author: base.author,
    colors: SkinColors(
      sakuraPink: accent,
      sakuraPinkLight: Color.lerp(accent, Colors.white, 0.3)!,
      sakuraPinkDark: Color.lerp(accent, Colors.black, 0.25)!,
      // Transparent, not a colour: the fluid backdrop is painted underneath
      // and anything opaque here would simply hide it.
      background: Colors.transparent,
      surface: Colors.white.withValues(alpha: 0.10),
      surfaceVariant: Colors.white.withValues(alpha: 0.08),
      surfaceContainer: Colors.white.withValues(alpha: 0.14),
      onBackground: Colors.white,
      onSurface: Colors.white,
      onSurfaceVariant: Colors.white.withValues(alpha: 0.72),
      outline: Colors.white.withValues(alpha: 0.34),
      outlineVariant: Colors.white.withValues(alpha: 0.18),
      error: const Color(0xFFFF6B6B),
      onError: Colors.white,
      playerBar: Colors.white.withValues(alpha: 0.10),
      miniPlayerShadow: Colors.black.withValues(alpha: 0.30),
      accentCyan: base.colors.accentCyan,
      accentPink: Color.lerp(accent, Colors.white, 0.3)!,
    ),
  );
}

extension CoverPaletteAccent on CoverPalette {
  /// Accent to use over the artwork backdrop.
  ///
  /// The opposite choice from [backdropFor], which deliberately avoids vibrant
  /// swatches: here the colour sits on *controls*, where the whole point is
  /// that it pops off the backdrop it was sampled from. Muted is the fallback,
  /// not the first pick.
  Color get accentOverArtwork =>
      lightVibrant ?? darkVibrant ?? lightMuted ?? dominant;
}
