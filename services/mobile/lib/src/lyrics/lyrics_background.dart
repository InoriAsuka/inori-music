import 'dart:io';
import 'dart:ui';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Full-bleed blurred-artwork backdrop shared by [KaraokeScreen] and
/// [FullPlayerScreen]'s lyrics tab. Falls back to the current skin's plain
/// background color when there's no artwork to show (no album, no embedded
/// cover, or the image fails to load) — the gradient scrim is still applied
/// in that case so [child] doesn't need two different contrast assumptions.
///
/// Since v5.24.0 the scrim is tinted with a colour sampled from the artwork
/// ([coverPaletteProvider]) instead of always being the flat skin background,
/// so the backdrop shifts with each track the way EchoMusic's and
/// OriginalSound's lyrics pages do.
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
    final skin = ref.watch(skinProvider).active;
    final palette = ref
        .watch(
          coverPaletteProvider((albumId: albumId, localArtUri: localArtUri)),
        )
        .valueOrNull;
    // valueOrNull covers loading, extraction failure and "no artwork at all"
    // in one branch: this is decoration, and it must never leave the lyrics
    // sitting on an unscrimmed image while a palette resolves.
    final scrim =
        palette?.backdropFor(skin.brightness) ?? context.skinColors.background;

    return Stack(
      fit: StackFit.expand,
      children: [
        ColoredBox(color: context.skinColors.background),
        _BlurredArtwork(albumId: albumId, localArtUri: localArtUri),
        // Two fixed opacity stops rather than an animated transition — the
        // blurred artwork underneath already changes with the track, and a
        // crossfading scrim on top of it reads as a lag, not a flourish.
        AnimatedContainer(
          duration: const Duration(milliseconds: 400),
          curve: Curves.easeOut,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                scrim.withValues(alpha: 0.55),
                scrim.withValues(alpha: 0.82),
              ],
            ),
          ),
        ),
        child,
      ],
    );
  }
}

/// Resolves the same artwork [FullPlayerScreen]'s artwork tile shows (local
/// file:// cover vs server [artworkUrlProvider]), blurred heavily enough to
/// stay a backdrop rather than compete with the lyrics text. Renders nothing
/// when there's no artwork or it fails to load — [LyricsBackground]'s flat
/// [ColoredBox] shows through underneath either way.
class _BlurredArtwork extends ConsumerWidget {
  const _BlurredArtwork({required this.albumId, required this.localArtUri});

  final String? albumId;
  final Uri? localArtUri;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final artUri = localArtUri;
    Widget image;
    if (artUri != null && artUri.scheme == 'file') {
      image = Image.file(
        File(artUri.toFilePath()),
        fit: BoxFit.cover,
        errorBuilder: (_, _, _) => const SizedBox.shrink(),
      );
    } else if (albumId != null && albumId!.isNotEmpty) {
      final url = ref.watch(artworkUrlProvider(albumId!)).valueOrNull;
      if (url == null || url.isEmpty) return const SizedBox.shrink();
      image = CachedNetworkImage(
        imageUrl: url,
        fit: BoxFit.cover,
        placeholder: (_, _) => const SizedBox.shrink(),
        errorWidget: (_, _, _) => const SizedBox.shrink(),
      );
    } else {
      return const SizedBox.shrink();
    }
    return ImageFiltered(
      imageFilter: ImageFilter.blur(sigmaX: 40, sigmaY: 40),
      child: image,
    );
  }
}
