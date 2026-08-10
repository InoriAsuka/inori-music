import 'dart:io';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/shared/theme/artwork_overlay_skin.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/cover_fluid_background.dart';

/// Full-bleed cover-derived backdrop shared by [KaraokeScreen] and
/// [FullPlayerScreen].
///
/// Since v5.26.0 this is EchoMusic's drifting colour field
/// ([CoverFluidBackground]) rather than a single blurred copy of the cover:
/// the artwork's own colours slowly reorganise behind the content instead of
/// sitting still. Falls back to the current skin's flat background colour
/// when there's no artwork to derive anything from (no album, no embedded
/// cover, or the image fails to load).
class LyricsBackground extends ConsumerWidget {
  const LyricsBackground({
    super.key,
    required this.albumId,
    required this.localArtUri,
    required this.child,
    this.useFluidBackground = true,
  });

  /// Server catalog album id — mutually exclusive with [localArtUri] in
  /// practice (a track is either server-backed or a local import).
  final String? albumId;

  /// Embedded cover art extracted from a guest-mode local file (file:// URI).
  final Uri? localArtUri;

  final Widget child;

  /// Whether to render the real [CoverFluidBackground] — four rotating
  /// image quadrants under a heavy 64-sigma backdrop blur, see that class's
  /// own doc comment — or [_CheapBackdrop], a flat/gradient stand-in built
  /// from the same derived accent colour.
  ///
  /// Both branches compute and apply the identical [artworkOverlaySkin]; only
  /// the decorative layer behind [child] differs, so nothing that reads
  /// `context.skinColors` (the play/pause button's accent, foreground ink,
  /// etc.) changes when this flips. [FullPlayerScreen] is the only caller
  /// that ever passes `false`, and only while its own entrance/drag
  /// transition is actually moving (see its `transitionProgress` doc
  /// comment) — v5.32.0 field report: "点击封面展开播放页，会卡顿后才弹出播放页",
  /// traced to this backdrop's own first-frame cost (compositing four image
  /// draws plus the blur filter's GPU work) landing on exactly the frames a
  /// slide-up transition needs to stay smooth. [KaraokeScreen], the only
  /// other caller, never has a transition to protect and always gets the
  /// real thing (the default).
  final bool useFluidBackground;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final image = resolveCoverImage(
      ref,
      albumId: albumId,
      localArtUri: localArtUri,
    );

    // No artwork means no backdrop, and content keeps the user's real skin —
    // light ink on a light background, exactly as before.
    if (image == null) {
      return ColoredBox(color: context.skinColors.background, child: child);
    }

    final palette = ref
        .watch(
          coverPaletteProvider((albumId: albumId, localArtUri: localArtUri)),
        )
        .valueOrNull;
    final overlaySkin = artworkOverlaySkin(
      ref.watch(skinProvider).active,
      palette: palette,
    );
    // Everything above the backdrop reads its colours from the derived
    // overlay skin rather than the user's — see [artworkOverlaySkin]. Shared
    // by both branches below so a caller flipping useFluidBackground never
    // changes anything but the decorative layer underneath this.
    final skinnedChild = SkinScope(skin: overlaySkin, child: child);

    if (!useFluidBackground) {
      return _CheapBackdrop(
        palette: palette,
        fallbackColor: context.skinColors.background,
        child: skinnedChild,
      );
    }

    return CoverFluidBackground(
      image: image,
      fallbackColor: context.skinColors.background,
      child: skinnedChild,
    );
  }
}

/// Stand-in for [CoverFluidBackground] while the player's own page
/// transition is still moving — a plain two-stop gradient built from the
/// already-resolved accent colour (see [LyricsBackground.useFluidBackground]
/// for why this exists). No image compositing, no colour-matrix filter, no
/// backdrop blur, no Ticker: cheap enough to cost nothing on frames a slide
/// animation needs to stay smooth.
class _CheapBackdrop extends StatelessWidget {
  const _CheapBackdrop({
    required this.palette,
    required this.fallbackColor,
    required this.child,
  });

  final CoverPalette? palette;
  final Color fallbackColor;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    // No palette yet is rare in practice — MiniPlayerBar primes
    // coverPaletteProvider for the current track continuously (see its own
    // doc comment) well before the player page can be opened — but falls
    // back to a flat skin colour rather than flashing an arbitrary one,
    // matching CoverFluidBackground's own no-image fallback.
    final accent = palette?.accentOverArtwork;
    if (accent == null) {
      return ColoredBox(color: fallbackColor, child: child);
    }
    return DecoratedBox(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          // Darkened toward the bottom rather than a flat fill — a hint of
          // the same depth CoverFluidBackground's own blurred, scrim-darkened
          // field has, so swapping between the two (once the transition
          // settles) reads as a refinement of the same backdrop rather than
          // a visible substitution.
          colors: [accent, Color.lerp(accent, Colors.black, 0.35)!],
        ),
      ),
      child: child,
    );
  }
}

/// Resolves the cover for a track to an [ImageProvider], or null when there
/// isn't one. Shared so the fluid backdrop, the palette extractor and the
/// artwork tile all sample the same image — a backdrop derived from a
/// different picture than the one on screen is worse than no backdrop.
///
/// Network covers go through [CachedNetworkImageProvider] so this reads bytes
/// that are already on disk rather than re-fetching them.
ImageProvider? resolveCoverImage(
  WidgetRef ref, {
  required String? albumId,
  required Uri? localArtUri,
}) {
  if (localArtUri != null && localArtUri.scheme == 'file') {
    final file = File(localArtUri.toFilePath());
    if (!file.existsSync()) return null;
    return FileImage(file);
  }
  if (albumId == null || albumId.isEmpty) return null;
  final url = ref.watch(artworkUrlProvider(albumId)).valueOrNull;
  if (url == null || url.isEmpty) return null;
  return CachedNetworkImageProvider(url);
}
