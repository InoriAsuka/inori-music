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
  });

  /// Server catalog album id — mutually exclusive with [localArtUri] in
  /// practice (a track is either server-backed or a local import).
  final String? albumId;

  /// Embedded cover art extracted from a guest-mode local file (file:// URI).
  final Uri? localArtUri;

  final Widget child;

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

    return CoverFluidBackground(
      image: image,
      fallbackColor: context.skinColors.background,
      // Everything above the backdrop reads its colours from the derived
      // overlay skin rather than the user's — see [artworkOverlaySkin].
      child: SkinScope(skin: overlaySkin, child: child),
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
