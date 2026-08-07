import 'dart:io';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:palette_generator/palette_generator.dart';

import 'package:inori_music/src/catalog/artwork_provider.dart';

/// Which cover to sample. A track is either server-backed (album id) or a
/// local import (embedded cover written to a file:// path) — both are carried
/// so callers can pass what they have without branching first.
///
/// A record rather than an encoded string so the family key keeps structural
/// equality for free and there's nothing to parse back out.
typedef CoverSource = ({String? albumId, Uri? localArtUri});

/// Colours pulled out of a cover image, reduced to just what the UI asks for.
///
/// Deliberately does not expose [PaletteGenerator] itself: callers should not
/// have to know which of its seven nullable swatches is appropriate, and
/// keeping the package type out of widget code means swapping the extractor
/// later doesn't ripple outwards.
@immutable
class CoverPalette {
  const CoverPalette({
    required this.dominant,
    this.darkMuted,
    this.lightMuted,
    this.darkVibrant,
    this.lightVibrant,
  });

  final Color dominant;
  final Color? darkMuted;
  final Color? lightMuted;
  final Color? darkVibrant;
  final Color? lightVibrant;

  /// Picks the swatch that stays out of the way of foreground text under the
  /// given theme brightness: muted first, vibrant as a second choice, and the
  /// dominant colour only if the image was flat enough to yield nothing else.
  ///
  /// Mirrors Spotube's `usePaletteColor` fallback order — a vibrant swatch
  /// behind lyrics reads as a coloured wash rather than a backdrop, so it is
  /// never the first pick.
  Color backdropFor(Brightness brightness) => brightness == Brightness.dark
      ? (darkMuted ?? darkVibrant ?? dominant)
      : (lightMuted ?? lightVibrant ?? dominant);

  @override
  bool operator ==(Object other) =>
      other is CoverPalette &&
      other.dominant == dominant &&
      other.darkMuted == darkMuted &&
      other.lightMuted == lightMuted &&
      other.darkVibrant == darkVibrant &&
      other.lightVibrant == lightVibrant;

  @override
  int get hashCode =>
      Object.hash(dominant, darkMuted, lightMuted, darkVibrant, lightVibrant);
}

/// Extracts a [CoverPalette] from the current track's cover art, or null when
/// there is no cover, it can't be read, or the image yields no usable colour.
///
/// Every failure path returns null rather than throwing: this drives an
/// optional visual flourish, and a decode error on one cover must never take
/// down the screen showing it. Callers fall back to their skin colour.
final coverPaletteProvider = FutureProvider.autoDispose
    .family<CoverPalette?, CoverSource>((ref, source) async {
      final imageProvider = await _resolveImage(ref, source);
      if (imageProvider == null) return null;

      final PaletteGenerator generator;
      try {
        generator = await PaletteGenerator.fromImageProvider(
          imageProvider,
          // Quantising the full-resolution artwork is pure waste — colour
          // distribution survives the downsample, and this keeps a large
          // cover from stalling the raster thread on first display.
          size: const Size(96, 96),
          maximumColorCount: 8,
        );
      } catch (_) {
        return null;
      }

      final dominant = generator.dominantColor?.color;
      if (dominant == null) return null;
      return CoverPalette(
        dominant: dominant,
        darkMuted: generator.darkMutedColor?.color,
        lightMuted: generator.lightMutedColor?.color,
        darkVibrant: generator.darkVibrantColor?.color,
        lightVibrant: generator.lightVibrantColor?.color,
      );
    });

Future<ImageProvider?> _resolveImage(Ref ref, CoverSource source) async {
  // Same local-vs-server split LyricsBackground already makes for the blurred
  // artwork itself, so the palette is always sampled from the image actually
  // on screen.
  final localArtUri = source.localArtUri;
  if (localArtUri != null && localArtUri.scheme == 'file') {
    final file = File(localArtUri.toFilePath());
    if (!file.existsSync()) return null;
    return FileImage(file);
  }

  final albumId = source.albumId;
  if (albumId == null || albumId.isEmpty) return null;
  final String? url;
  try {
    url = await ref.watch(artworkUrlProvider(albumId).future);
  } catch (_) {
    return null;
  }
  if (url == null || url.isEmpty) return null;
  // The same provider CachedNetworkImage uses, so this reads the already
  // downloaded bytes instead of re-fetching the cover.
  return CachedNetworkImageProvider(url);
}
