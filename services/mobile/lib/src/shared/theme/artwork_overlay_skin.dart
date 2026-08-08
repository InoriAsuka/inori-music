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
/// * **Foregrounds follow the backdrop's actual brightness.** v5.26.0 through
///   v5.28.0 assumed the scrim over the artwork was always dark by
///   construction and hardcoded white ink. That assumption is wrong:
///   [CoverFluidBackground] only paints a 24% black scrim over a *brightened*
///   copy of the cover, so a light cover still yields a light backdrop, and
///   white text/icons on it are unreadable — reported against a light cover
///   in v5.29.0, where the track title and several secondary controls were
///   nearly invisible. [backdropLuminance] recomputes what the backdrop
///   actually renders (the same boost matrix, then the same scrim) and the
///   foreground flips to dark ink whenever that comes out light.
/// * **The accent follows the artwork.** [palette] replaces the skin's brand
///   pink, so the progress bar, the play button and every selected state take
///   their colour from the cover. When there's no palette yet the skin's own
///   accent stays, so nothing flashes through an uncoloured frame. This is
///   unconditional — unlike the foreground tokens above, the accent does not
///   flip with backdrop brightness, because its whole job is to pop off
///   whatever the backdrop is doing rather than to blend in for legibility.
SkinDefinition artworkOverlaySkin(
  SkinDefinition base, {
  required CoverPalette? palette,
}) {
  final accent = palette?.accentOverArtwork ?? base.colors.sakuraPink;

  // No palette yet (still extracting, or extraction failed) — the backdrop
  // is already on screen regardless, since CoverFluidBackground only needs
  // the image, not the palette. Dark is the only sensible guess here: it
  // matches every frame this skin produced before v5.29.0, and it is
  // normally gone within one frame once the palette resolves.
  final luminance = palette == null ? null : backdropLuminance(palette);
  final isLightBackdrop = (luminance ?? 0.0) >= 0.5;
  final foreground = isLightBackdrop
      ? _darkInkForeground()
      : _whiteInkForeground(warm: palette != null && _isCoolHue(palette));

  return SkinDefinition(
    // The brightness tier rides along in the id, not just the accent: two
    // tracks can share a muted accent while one has a light cover and the
    // other a dark one, and SkinScope.updateShouldNotify only compares ids —
    // without this, switching between such tracks would leave the previous
    // foreground on screen instead of flipping.
    id:
        '${base.id}#artwork-${accent.toARGB32()}-'
        '${isLightBackdrop ? 'light' : 'dark'}',
    displayName: base.displayName,
    brightness: isLightBackdrop ? Brightness.light : Brightness.dark,
    author: base.author,
    colors: SkinColors(
      sakuraPink: accent,
      sakuraPinkLight: Color.lerp(accent, Colors.white, 0.3)!,
      sakuraPinkDark: Color.lerp(accent, Colors.black, 0.25)!,
      // Transparent, not a colour: the fluid backdrop is painted underneath
      // and anything opaque here would simply hide it.
      background: Colors.transparent,
      surface: foreground.surface,
      surfaceVariant: foreground.surfaceVariant,
      surfaceContainer: foreground.surfaceContainer,
      onBackground: foreground.ink,
      onSurface: foreground.ink,
      onSurfaceVariant: foreground.inkVariant,
      outline: foreground.outline,
      outlineVariant: foreground.outlineVariant,
      error: const Color(0xFFFF6B6B),
      onError: Colors.white,
      playerBar: foreground.surface,
      miniPlayerShadow: Colors.black.withValues(alpha: 0.30),
      accentCyan: base.colors.accentCyan,
      accentPink: Color.lerp(accent, Colors.white, 0.3)!,
    ),
  );
}

/// The luminance ([Color.computeLuminance]) of what [CoverFluidBackground]
/// actually paints behind the content, given the cover's dominant colour.
///
/// Recomputes the same boost the backdrop applies (saturate(1.3) ×
/// brightness(1.5) — see `CoverFluidBackground._boost`) plus the same 24%
/// black scrim it paints on top, via [_boostedBackdropColor]. Exposed as a
/// pure function so [artworkOverlaySkin] can branch on it and tests can
/// assert against it directly, without pumping the animated backdrop widget.
double backdropLuminance(CoverPalette palette) =>
    _boostedBackdropColor(palette.dominant).computeLuminance();

/// Reproduces `CoverFluidBackground`'s boost matrix and scrim on a single
/// swatch, so the backdrop's resulting colour can be computed without
/// rendering it.
Color _boostedBackdropColor(Color swatch) {
  double clampChannel(double v) => v.clamp(0.0, 1.0);
  final r = swatch.r, g = swatch.g, b = swatch.b;
  // Same 5x4 matrix as CoverFluidBackground._boost, applied directly to the
  // 0-1 channel values Color already exposes rather than round-tripping
  // through 0-255 — the matrix has no translation column, so the two spaces
  // agree exactly.
  final boosted = Color.from(
    alpha: 1.0,
    red: clampChannel(1.85433 * r - 0.32184 * g - 0.03249 * b),
    green: clampChannel(-0.09567 * r + 1.62816 * g - 0.03249 * b),
    blue: clampChannel(-0.09567 * r - 0.32184 * g + 1.91751 * b),
  );
  // CoverFluidBackground paints this scrim over an opaque backdrop, so
  // alphaBlend's opaque-background branch applies: the result is just the
  // boosted colour scaled by (1 - 0.24).
  return Color.alphaBlend(Colors.black.withValues(alpha: 0.24), boosted);
}

/// Whether the boosted backdrop colour sits in the cyan-through-violet range
/// (hue 180–280) — used only to nudge white foreground text slightly warm
/// when it would otherwise sit on a cool dark backdrop. Deliberately not
/// mirrored on the light-backdrop branch: the v5.29.0 scope decision was
/// brightness-driven contrast plus this one warmth nudge, not a full
/// complementary-colour system.
bool _isCoolHue(CoverPalette palette) {
  final hue = HSLColor.fromColor(_boostedBackdropColor(palette.dominant)).hue;
  return hue >= 180 && hue <= 280;
}

/// The seven foreground/surface tokens [artworkOverlaySkin] derives, grouped
/// so each branch below can build one in a single expression.
typedef _OverlayForeground = ({
  Color ink,
  Color inkVariant,
  Color outline,
  Color outlineVariant,
  Color surface,
  Color surfaceVariant,
  Color surfaceContainer,
});

/// Light backdrop → dark ink. Near-black rather than pure black reads softer
/// against a colour field that is still, by construction, saturated.
_OverlayForeground _darkInkForeground() => (
  ink: Colors.black.withValues(alpha: 0.88),
  inkVariant: Colors.black.withValues(alpha: 0.60),
  outline: Colors.black.withValues(alpha: 0.28),
  outlineVariant: Colors.black.withValues(alpha: 0.14),
  surface: Colors.black.withValues(alpha: 0.10),
  surfaceVariant: Colors.black.withValues(alpha: 0.08),
  surfaceContainer: Colors.black.withValues(alpha: 0.14),
);

/// Dark backdrop → white ink — the only branch that existed before v5.29.0.
/// [warm] nudges the white toward a warm white for cool-hued backdrops, per
/// the brightness-first, hue-second scope decision above.
_OverlayForeground _whiteInkForeground({required bool warm}) {
  const warmWhite = Color(0xFFFFF1E0);
  final ink = warm ? Color.lerp(Colors.white, warmWhite, 0.16)! : Colors.white;
  return (
    ink: ink,
    inkVariant: ink.withValues(alpha: 0.72),
    outline: ink.withValues(alpha: 0.34),
    outlineVariant: ink.withValues(alpha: 0.18),
    surface: ink.withValues(alpha: 0.10),
    surfaceVariant: ink.withValues(alpha: 0.08),
    surfaceContainer: ink.withValues(alpha: 0.14),
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
