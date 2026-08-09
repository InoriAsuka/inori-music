import 'dart:io';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/catalog/artwork_provider.dart';

/// Resolves and paints a track's cover image content — the "which source
/// wins" logic shared by the mini player bar's [MiniPlayerArtwork] (see
/// `mini_player_bar.dart`) and the full player's own artwork tile (see
/// `full_player_screen.dart`'s `_FullPlayerArtwork`).
///
/// A track's cover comes from one of two disjoint sources, never both:
/// catalog tracks carry [albumId] and have no local file to read a cover
/// from, so [PlayerNotifier] always sets their `MediaItem.artUri` to null
/// (see `player_notifier.dart`'s `_makeMediaItem`); local-library tracks are
/// guest-mode files with no server album at all, so their cover is whatever
/// image `audio_metadata_reader` pulled out of the file itself, exposed as a
/// `file://` URI on `MediaItem.artUri` and passed in here as [localArtUri].
/// Checking [localArtUri] first and falling back to the [albumId] lookup
/// only when it is absent (or not a `file://` URI) covers both without
/// either source needing to know the other exists.
///
/// Before v5.30.6 this logic was duplicated between the two call sites
/// above, and the mini player bar's copy had never actually grown the
/// [localArtUri] branch at all — local tracks played from the local library
/// showed a cover in every list row (which reads straight from the local
/// DB) but a bare music-note placeholder in the bar, since [MiniPlayerArtwork]
/// only ever knew how to ask the server for a cover. This widget is the
/// single place that decision gets made now, so a future change to it (a
/// third source, a different fallback order) cannot land in one call site
/// and miss the other again.
///
/// Deliberately renders only the image (or [fallback]) and nothing else —
/// no background fill, no rounding, no fixed-size box. The two call sites'
/// *containers* differ enough (the mini bar wraps this in a rounded,
/// background-filled `Container`; the full player wraps it in a plain
/// `ClipRRect` over the colour-field backdrop with a much larger fallback
/// glyph) that unifying those too would mean forcing one to look like the
/// other for no real gain — only the source-selection-and-loading-states
/// logic was actually identical, so only that part moved here.
class TrackArtwork extends ConsumerWidget {
  const TrackArtwork({
    super.key,
    required this.size,
    required this.fallback,
    this.albumId,
    this.localArtUri,
  });

  /// Edge length in logical pixels for both the `Image.file` and
  /// `CachedNetworkImage` branches. [fallback] is responsible for its own
  /// sizing — callers that need it to fill the same box pass a
  /// [fallback] that does so itself (see [MiniPlayerArtwork]'s usage).
  final double size;

  final String? albumId;

  /// Embedded cover art extracted from a guest-mode local file, exposed as a
  /// `file://` URI on `MediaItem.artUri`. Anything not a `file://` URI is
  /// treated the same as absent, falling through to the [albumId] lookup.
  final Uri? localArtUri;

  /// Built fresh for every "no image available" outcome (missing source,
  /// still loading, or a load error) rather than a plain [Widget] — both
  /// call sites' fallback glyphs read `context.skinColors`, which needs a
  /// [BuildContext] to resolve.
  final WidgetBuilder fallback;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final artUri = localArtUri;
    if (artUri != null && artUri.scheme == 'file') {
      return Image.file(
        File(artUri.toFilePath()),
        width: size,
        height: size,
        fit: BoxFit.cover,
        errorBuilder: (context, _, _) => fallback(context),
      );
    }

    if (albumId == null || albumId!.isEmpty) {
      return fallback(context);
    }

    final artworkAsync = ref.watch(artworkUrlProvider(albumId!));
    return artworkAsync.when(
      data: (url) {
        if (url == null || url.isEmpty) return fallback(context);
        return CachedNetworkImage(
          imageUrl: url,
          width: size,
          height: size,
          fit: BoxFit.cover,
          placeholder: (context, _) => fallback(context),
          errorWidget: (context, _, error) => fallback(context),
        );
      },
      loading: () => fallback(context),
      error: (error, _) => fallback(context),
    );
  }
}
