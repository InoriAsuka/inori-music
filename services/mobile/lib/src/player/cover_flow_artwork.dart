import 'package:flutter/material.dart';

/// Old-school "Cover Flow" artwork display mode: the current track's cover
/// faces forward at full size in the centre, its neighbours in the queue
/// recede to the sides at a shrunken, tilted scale, and how many of them fit
/// is a direct function of how much width the player has — not a fixed
/// count. An optional display mode (Settings → Appearance), not a plugin:
/// see requirement.md v5.30.0 for why a full plugin UI DSL was out of scope.
///
/// Deliberately generic over what each slot renders ([itemBuilder]) rather
/// than reaching into [PlayerState.queue] and artwork providers itself — the
/// caller already owns exactly that data (see
/// `_FullPlayerScreenState._playerBlock`), and keeping the geometry here free
/// of provider watches is what makes [visibleSideCount] and this widget
/// testable with a trivial builder instead of real network/local artwork.
class CoverFlowArtwork extends StatelessWidget {
  const CoverFlowArtwork({
    super.key,
    required this.itemCount,
    required this.currentIndex,
    required this.centerSize,
    required this.width,
    required this.itemBuilder,
    this.onSelect,
  });

  /// Total number of tracks in the queue this is flowing through.
  final int itemCount;

  /// Index of the track shown centred/forward-facing.
  final int currentIndex;

  /// Edge length of the centred cover. Side covers scale down from this.
  final double centerSize;

  /// Width available to lay the flow out in — the same region width the
  /// plain single-cover tile centres itself in, so switching the setting on
  /// and off doesn't shift the centred cover's position.
  final double width;

  /// Builds the widget for the cover at [index], sized to [size] (already
  /// scaled down for a side slot, or [centerSize] itself for the centre).
  final Widget Function(BuildContext context, int index, double size)
  itemBuilder;

  /// Invoked with a side cover's queue index when it's tapped, so the caller
  /// can jump playback to it. Null disables the tap (e.g. read-only previews).
  final ValueChanged<int>? onSelect;

  /// Side covers scale to this fraction of the centred cover's size —
  /// EchoMusic-style Cover Flow shrinks neighbours rather than keeping them
  /// full size, so the centre reads as "focused" rather than one of a row of
  /// equals.
  static const _sideScale = 0.72;

  /// Opacity applied to every side cover, fading them slightly so the centre
  /// unambiguously reads as the current track.
  static const _sideOpacity = 0.55;

  /// Horizontal distance from the centre to the first side slot, and between
  /// each subsequent one, as a multiple of the centred cover's own size.
  /// Below 1.0 so consecutive covers overlap slightly (the classic Cover Flow
  /// look) rather than sitting edge to edge.
  static const _stride = 0.68;

  /// Hard ceiling on how many covers show on each side regardless of how
  /// much width is on offer — an ultrawide monitor showing eleven covers
  /// reads as a filmstrip, not a "flow" with one clear focal point.
  static const _maxSidePerSide = 3;

  /// How many covers fit on *each* side of the centred one, given [width]
  /// and [centerSize].
  ///
  /// A pure function — deliberately free of [BuildContext] or any provider —
  /// so it can be unit-tested directly for "the count changes with width"
  /// without pumping a widget at all. Each side gets half the width minus the
  /// half of the centred cover sitting in the middle; how many [_stride]-wide
  /// slots fit in that remainder is the answer, capped at [_maxSidePerSide].
  static int visibleSideCount({
    required double width,
    required double centerSize,
  }) {
    if (centerSize <= 0) return 0;
    final available = width / 2 - centerSize / 2;
    if (available <= 0) return 0;
    final strideWidth = centerSize * _stride;
    return (available / strideWidth).floor().clamp(0, _maxSidePerSide);
  }

  @override
  Widget build(BuildContext context) {
    final sideCount = visibleSideCount(width: width, centerSize: centerSize);

    // Painting order: farthest-from-centre first on each side, centre last —
    // so nearer covers land on top of the farther ones they overlap, and the
    // centre lands on top of everything either side reaches under it.
    final offsets = <int>[
      for (var d = sideCount; d >= 1; d--) ...[-d, d],
      0,
    ];

    final children = <Widget>[];
    for (final offset in offsets) {
      final index = currentIndex + offset;
      if (index < 0 || index >= itemCount) continue;
      final isCenter = offset == 0;
      final size = isCenter ? centerSize : centerSize * _sideScale;
      final dx = offset * centerSize * _stride;

      children.add(
        Positioned(
          left: width / 2 + dx - size / 2,
          top: (centerSize - size) / 2,
          child: GestureDetector(
            onTap: (isCenter || onSelect == null)
                ? null
                : () => onSelect!(index),
            child: Opacity(
              opacity: isCenter ? 1.0 : _sideOpacity,
              child: Transform(
                alignment: Alignment.center,
                transform: Matrix4.identity()
                  ..setEntry(3, 2, 0.001)
                  ..rotateY(isCenter ? 0 : (offset > 0 ? -0.4 : 0.4)),
                child: SizedBox(
                  width: size,
                  height: size,
                  child: itemBuilder(context, index, size),
                ),
              ),
            ),
          ),
        ),
      );
    }

    return SizedBox(
      width: width,
      height: centerSize,
      child: Stack(clipBehavior: Clip.none, children: children),
    );
  }
}
