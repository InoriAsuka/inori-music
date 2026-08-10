// ignore_for_file: unnecessary_non_null_assertion
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/audio/sleep_timer_notifier.dart';
import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/playback_mode_buttons.dart';
import 'package:inori_music/src/player/queue_drawer_provider.dart';
import 'package:inori_music/src/player/track_artwork.dart';
import 'package:inori_music/src/player/volume_control.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/floating_shadow.dart';
import 'package:inori_music/src/shared/widgets/hover_link_text.dart';
import 'package:inori_music/src/shared/widgets/seek_bar_effects.dart';
import 'package:inori_music/src/shared/widgets/shell_chrome.dart';
import 'package:inori_music/src/shared/widgets/spring_interaction.dart';

/// Persistent mini-player bar displayed at the bottom of the shell scaffold.
///
/// A floating rounded card rather than a full-bleed strip — EchoMusic and
/// Spotube both keep the player bar visually separate from the nav chrome
/// beneath it (a `SurfaceCard`/elevated rounded panel with margins), not a
/// flush edge-to-edge bar. The margin here is what makes the drop shadow
/// (see [floatingShadow]) actually visible on both sides instead of running
/// off the screen edges.
///
/// Scaled to EchoMusic's `PlayerBar.vue` proportions since v5.30.0 rather
/// than Apple Music's — the user explicitly flagged Apple Music's bar as too
/// small once the player page itself had been reworked to match Apple Music
/// elsewhere. The content row is a fixed 84px (EchoMusic's `h-21`), and the
/// artwork is 56px.
///
/// v5.30.5 briefly split this into two shapes behind a `showNowPlaying`
/// constructor flag — mobile/tablet kept the cover+title section, and the
/// desktop shell docked that block in the sidebar instead (see the deleted
/// `SidebarNowPlaying`), passing `showNowPlaying: false` so the bar's freed
/// section carried shuffle/repeat instead. The user's field report on that
/// build was explicit that the cover belongs *with* the transport controls,
/// not in the nav chrome, so v5.30.6 reverted the sidebar placement — and
/// once the cover was back in the bar on every layout, the flag had nothing
/// left to switch on: both call sites wanted the same shape. What actually
/// still needs to differ between a phone-width bar and a desktop-width one
/// is not "is this the desktop shell" but "is there room", so the shuffle/
/// repeat-flanked transport group, the seek row underneath it, and the
/// volume/timer/queue action group are now gated on this bar's own measured
/// width (see [_wideBreakpoint]) rather than on which caller constructed it.
/// Below that width the bar keeps its pre-v5.30.6 narrow shape: transport
/// trio only, a plain seek strip across the top, and just the sleep timer.
class MiniPlayerBar extends ConsumerWidget {
  const MiniPlayerBar({super.key});

  /// EchoMusic's `PlayerBar.vue` footer height (`h-21`).
  static const _contentHeight = 84.0;

  /// EchoMusic's cover size within that footer.
  static const _artworkSize = 56.0;

  /// EchoMusic's transport icon size (`width="22" height="22"` on
  /// prev/next/play-mode in `PlayerBar.vue`). The play/pause glyph itself
  /// stays a little larger, matching the existing (pre-v5.30.0) proportion
  /// between it and its siblings. Also applied to the wide shape's
  /// shuffle/repeat/volume/timer/queue controls — EchoMusic's own
  /// "play-mode" icons share this same 22px, so this doubles as their size
  /// too.
  static const _transportIconSize = 22.0;

  /// Below this width, section 3's volume control has no room for its inline
  /// slider and collapses to an icon-only trigger that opens the slider in a
  /// popover instead — judged against the section's own measured width (via
  /// the LayoutBuilder in [_wideActionsRow]), not the window's, per the
  /// basis-mismatch lesson `full_player_screen.dart`'s
  /// `_compactControlsBreakpoint` documents.
  ///
  /// 240, not a rounder-looking 200 — section 3 in expanded mode measures
  /// exactly 234px (three 48px default IconButton tap targets for
  /// volume/timer/queue, plus [VolumeControl]'s 90px slider track: no inline
  /// gaps, since this Row sets none). 200 left a real gap between "not narrow
  /// enough to switch to compact" and "actually wide enough for 234px",
  /// which overflowed rather than triggering the collapse meant to prevent
  /// exactly that — caught by test/mini_player_bar_desktop_test.dart, not by
  /// inspection. 240 clears 234 with a few px to spare; compact mode itself
  /// only needs 144px (three icons, no slider), so anything below 240 has
  /// ample room once collapsed.
  static const _volumeCompactThreshold = 240.0;

  /// Below this width (the bar's own measured width — see the LayoutBuilder
  /// in [build] — never the window's), the wide shape's three simultaneous
  /// additions (shuffle/repeat flanking the transport trio, a seek row with
  /// time labels underneath it, and volume+timer+queue on the right) do not
  /// fit alongside a 56px cover and title. Below it the bar falls back to
  /// its pre-v5.30.6 narrow shape instead of trying to cram a shrunken
  /// version of the wide one into the same space — see
  /// `test/mini_player_bar_desktop_test.dart` for the overflow probes this
  /// was calibrated against, and `test/mini_player_bar_test.dart` for proof
  /// the narrow shape itself still holds at a 375dp phone width.
  static const _wideBreakpoint = 640.0;

  /// Width of just the transport-group row (shuffle/prev/play/next/repeat),
  /// centred inside the wider middle column [miniPlayerMiddleColumnWidth]
  /// computes (see [_wideContent]) — through v5.31.0 this constant also
  /// doubled as the seek row's own width, which is exactly what the
  /// v5.32.0 field report ("进度条太短，与整条比例失调") called out: capping
  /// the seek row at however wide five fixed-size icons happen to need left
  /// it looking like an afterthought on a spacious desktop bar. The two are
  /// independent now — this constant only ever sizes the icon row.
  ///
  /// 248, not a rounder-looking 240 — the five-icon row (shuffle, prev, the
  /// larger play/pause circle, next, repeat) with `spaceBetween` and no
  /// inline gaps measures exactly 240px, which a first pass at 232 fell
  /// 8px short of (`RenderFlex overflowed by 8.0 pixels`, caught by
  /// mini_player_bar_test.dart running its unpinned-width cases at the
  /// default 800px test surface — comfortably past _wideBreakpoint, so they
  /// exercise this shape too). 248 clears the exact 240px requirement with
  /// a few px to spare rather than sitting flush against it, the same
  /// margin style as [_volumeCompactThreshold]'s own buffer above its
  /// measured 234px.
  static const _transportBlockWidth = 248.0;

  /// Identifies the fixed-height content row in tests, so its rendered size
  /// can be asserted against without relying on it being the only widget of
  /// some common type in the tree.
  static const contentKey = ValueKey('miniPlayerBarContent');

  /// Identifies the floating card's own outer edge (the [DecoratedBox]
  /// carrying the rounded corners + shadow) — distinct from [contentKey],
  /// which spans only the fixed-height content row and excludes the narrow
  /// shape's extra top progress strip. Margin/inset tests need the whole
  /// card's bounds, not just the content row's.
  static const cardKey = ValueKey('miniPlayerBarCard');

  /// Uniform outer margin between the floating bar and whatever surrounds it
  /// (the shell's content column, the window edge…), on all four sides.
  ///
  /// Through v5.32.0 this was `EdgeInsets.fromLTRB(8, 0, 8, 8)` — three sides
  /// at 8, the top at a bare 0 — so the top edge butted directly against
  /// whatever sits above the bar instead of floating clear of it like the
  /// other three (v5.33.0 field report: "控制条完全可以参考 EchoMusic 这样的上下
  /// 左右留白等宽，而不是现在我们这种基本贴边"). One named constant applied via
  /// [EdgeInsets.all] is the actual fix, not just picking a number — four
  /// independent literals is exactly what let three of them agree on 8 and
  /// leave the fourth behind; a fifth ad hoc literal added later to "fix"
  /// this would only repeat the mistake.
  static const _outerMargin = 8.0;

  /// Horizontal inset between the floating card's own border and its
  /// content (the cover, title/artist, transport controls…).
  ///
  /// Derived from, rather than independently chosen alongside, the
  /// *vertical* gap the layout already produces for free: an [_artworkSize]
  /// (56px) cover centred inside the [_contentHeight] (84px) content row
  /// leaves (84-56)/2 = 14px above and below it with no padding of its own.
  /// Through v5.32.0 the horizontal inset was a flat, unrelated 12
  /// (`EdgeInsets.symmetric(horizontal: 12)`), so the cover sat measurably
  /// closer to the card's left edge than to its top/bottom edges — the
  /// other half of the v5.33.0 field report ("四边留白等宽"). Computing this
  /// from the same two constants that already produce the vertical figure —
  /// instead of hand-typing 14 to match today's values — keeps the two
  /// figures from drifting apart again if either constant above ever
  /// changes.
  static const _contentHorizontalInset = (_contentHeight - _artworkSize) / 2;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final playerState = ref.watch(playerProvider);
    final mediaItem = playerState.mediaItem;
    final isPlaying = playerState.isPlaying;
    final isBuffering = playerState.isBuffering;
    final t = AppLocalizations.of(context);

    final title = mediaItem?.title ?? t.nothingPlaying;
    final artist = mediaItem?.artist ?? '';
    final albumId = mediaItem?.extras?['albumId'] as String?;
    // Null for a guest-mode local track (no server-side artist id at all) —
    // see player_notifier.dart's _makeMediaItem doc comment. HoverLinkText
    // renders plain, inert text for a null id rather than a dead-looking
    // link, so that case needs no special handling here beyond passing it
    // through.
    final artistId = mediaItem?.extras?['artistId'] as String?;
    // Embedded cover art extracted from a guest-mode local file — see
    // TrackArtwork's doc comment for why this and albumId are mutually
    // exclusive rather than both being checked against the same track.
    final localArtUri = mediaItem?.artUri;

    // v5.32.0: primes coverPaletteProvider for whatever is currently
    // playing, well before the user can ever open the full player screen —
    // this bar is mounted for the whole time something is loaded, on every
    // platform, unlike the desktop shell's own _LayoutAccentGradient (which
    // only exists ≥1200dp and would otherwise be the *only* continuous
    // watcher). coverPaletteProvider is `autoDispose`, so its cached
    // PaletteGenerator result only survives between the mini bar closing and
    // the full player opening because *some* widget kept watching it the
    // whole time in between — before this, on any layout narrower than
    // 1200dp, nothing did, and the full player recomputed the palette from
    // scratch (a real PaletteGenerator pass over the decoded image) on every
    // single open. The value itself is unused here — this watch exists
    // purely to keep the provider warm, mirroring the image itself already
    // being decoded for MiniPlayerArtwork below.
    if (albumId != null || localArtUri != null) {
      ref.watch(
        coverPaletteProvider((albumId: albumId, localArtUri: localArtUri)),
      );
    }

    return Padding(
      padding: const EdgeInsets.all(_outerMargin),
      // The shadow lives on this outer DecoratedBox, not inside the
      // Material below — see floatingShadow's doc comment on why a
      // clipping ancestor (this Material clips its own rounded corners)
      // must never sit between a BoxShadow and open air.
      child: DecoratedBox(
        key: cardKey,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          boxShadow: floatingShadow(context.skinColors.miniPlayerShadow),
        ),
        child: Material(
          color: context.skinColors.playerBar,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
          clipBehavior: Clip.antiAlias,
          child: LayoutBuilder(
            builder: (context, constraints) {
              final isWide = constraints.maxWidth >= _wideBreakpoint;
              return Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Only the narrow shape carries this plain top strip — the
                  // wide shape's seek control moves into the content row
                  // itself (see _wideContent), directly under the transport
                  // group, with its own time labels.
                  if (!isWide) const _MiniPlayerProgressBar(),
                  SizedBox(
                    key: contentKey,
                    height: _contentHeight,
                    child: SafeArea(
                      top: false,
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: _contentHorizontalInset,
                        ),
                        // No outer InkWell here since v5.30.7 — the whole
                        // content row used to open the full player on tap,
                        // which meant tapping the title/artist (reasonable
                        // muscle memory once they became links to their own
                        // detail pages, see _nowPlayingInfo) fought with
                        // opening the player instead of navigating there.
                        // The field report was explicit that only the cover
                        // should open the player now; _nowPlayingInfo wires
                        // that gesture onto just the artwork instead.
                        child: isWide
                            ? _wideContent(
                                context,
                                ref,
                                title: title,
                                artist: artist,
                                albumId: albumId,
                                artistId: artistId,
                                localArtUri: localArtUri,
                                isPlaying: isPlaying,
                                isBuffering: isBuffering,
                              )
                            : _narrowContent(
                                context,
                                ref,
                                title: title,
                                artist: artist,
                                albumId: albumId,
                                artistId: artistId,
                                localArtUri: localArtUri,
                                isPlaying: isPlaying,
                                isBuffering: isBuffering,
                              ),
                      ),
                    ),
                  ),
                ],
              );
            },
          ),
        ),
      ),
    );
  }

  static String _formatRemaining(Duration d) {
    final m = d.inMinutes.remainder(60).toString().padLeft(2, '0');
    final s = d.inSeconds.remainder(60).toString().padLeft(2, '0');
    return '$m:$s';
  }
}

/// The bar's pre-v5.30.6 shape: cover+title, the transport trio, and just
/// the sleep timer — kept verbatim (down to the plain top seek strip) as the
/// fallback for widths too narrow for [_wideContent]'s additions, rather
/// than a shrunken version of the wide shape. It was already proven not to
/// overflow at phone widths before this phase touched anything, and
/// "retain a simplified representation" (the v5.30.6 field report's own
/// framing) reads most naturally as *this*, not a new third shape.
Widget _narrowContent(
  BuildContext context,
  WidgetRef ref, {
  required String title,
  required String artist,
  required String? albumId,
  required String? artistId,
  required Uri? localArtUri,
  required bool isPlaying,
  required bool isBuffering,
}) {
  return Row(
    children: [
      // Section 1 — artwork + title/artist. Expanded, so it absorbs
      // whatever width the fixed-size transport trio and section 3 don't
      // need.
      Expanded(
        child: _nowPlayingInfo(
          context,
          title: title,
          artist: artist,
          albumId: albumId,
          artistId: artistId,
          localArtUri: localArtUri,
        ),
      ),

      // Section 2 — the transport trio. Fixed-size, kept dead centre by
      // section 1 and section 3 sharing equal Expanded flex around it rather
      // than either side growing to whatever it happens to contain.
      Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          _transportButton(
            context: context,
            icon: Icons.skip_previous,
            tooltip: 'Previous',
            onPressed: () => ref.read(playerProvider.notifier).previous(),
          ),
          _playPauseButton(
            context,
            ref,
            isPlaying: isPlaying,
            isBuffering: isBuffering,
          ),
          _transportButton(
            context: context,
            icon: Icons.skip_next,
            tooltip: 'Next',
            onPressed: () => ref.read(playerProvider.notifier).next(),
          ),
        ],
      ),

      // Section 3 — just the sleep timer, mirroring section 1's flex so
      // section 2 actually sits in the middle instead of drifting toward
      // whichever side has less in it.
      Expanded(
        child: Align(
          alignment: Alignment.centerRight,
          child: _sleepTimerButton(),
        ),
      ),
    ],
  );
}

/// Width of the wide shape's middle column (the transport-group row stacked
/// over the seek row) — derived from the row's own measured width, never
/// the window's (the same basis-mismatch lesson every other breakpoint in
/// this file documents) and never the transport row's icon count, which is
/// what produced the pre-v5.32.0 fixed 248 this replaces as the *column's*
/// width. [MiniPlayerBar._transportBlockWidth] still sizes the transport
/// button row itself, centred inside this wider column (see [_wideContent])
/// — only the seek row actually grows to fill it.
///
/// v5.32.0 field report: "进度条太短，与整条比例失调" — with the seek row
/// capped at the same 248px as the five transport icons, a ~1200px desktop
/// bar left it looking like an afterthought squeezed between two much wider
/// sections. 0.46 of the row's own available width, clamped to
/// [_middleColumnFloor, 640]: at [MiniPlayerBar._wideBreakpoint] (640, the
/// narrowest this shape ever renders at) it already clears the 248px
/// transport row with room to spare (640 * 0.46 ≈ 294); by a spacious
/// 1400px desktop bar it reaches roughly 644, capped at 640 so sections 1
/// and 3 (now-playing info, the action group) always keep a usable share
/// rather than this column eating the whole row on an ultrawide monitor.
double miniPlayerMiddleColumnWidth(double rowWidth) =>
    (rowWidth * 0.46).clamp(_middleColumnFloor, 640.0);

/// Floor for [miniPlayerMiddleColumnWidth] — comfortably above
/// [MiniPlayerBar._transportBlockWidth] (248) so the transport row keeps
/// margin on both sides once centred in the column, even at the clamp's
/// lower bound.
const _middleColumnFloor = 290.0;

/// Width of the seek row (progress bar + time labels) within the middle
/// column — wider than the transport-button row stacked above it (which
/// stays pinned to [MiniPlayerBar._transportBlockWidth] so its five icons
/// don't spread out with ugly gaps at the column's full width), clamped to
/// a usable band.
///
/// 0.7 of the middle column: the field report's own suggested 55%-75% range
/// (see requirement.md's v5.32.0 entry), picked toward the wide end of it
/// because the row's own content — two fixed-width time labels plus the
/// gaps beside them, see [_MiniPlayerSeekRowState._timeLabelGap] — already
/// eats a fixed chunk of whatever width it is given, so the *track* itself
/// (the part that actually reads as "the progress bar") ends up narrower
/// than 0.7 alone suggests. The floor/ceiling bound it independently of the
/// middle column's own clamp.
double miniPlayerSeekRowWidth(double middleColumnWidth) =>
    (middleColumnWidth * 0.7).clamp(220.0, 460.0);

/// The wide shape (v5.30.6): cover+title on the left; a shuffle/repeat-
/// flanked transport group with its own, independently-sized seek row
/// underneath, centred; and volume/timer/queue on the right. Sections 1 and
/// 3 share equal Expanded flex around section 2's own computed
/// [miniPlayerMiddleColumnWidth] so it stays centred regardless of how long
/// the title or how the action group's own width shifts between compact and
/// expanded volume — the same relationship as before v5.32.0, just against
/// a wider, measured column instead of the transport row's fixed size.
Widget _wideContent(
  BuildContext context,
  WidgetRef ref, {
  required String title,
  required String artist,
  required String? albumId,
  required String? artistId,
  required Uri? localArtUri,
  required bool isPlaying,
  required bool isBuffering,
}) {
  return LayoutBuilder(
    builder: (context, constraints) {
      final middleWidth = miniPlayerMiddleColumnWidth(constraints.maxWidth);
      return Row(
        children: [
          Expanded(
            child: _nowPlayingInfo(
              context,
              title: title,
              artist: artist,
              albumId: albumId,
              artistId: artistId,
              localArtUri: localArtUri,
            ),
          ),
          SizedBox(
            width: middleWidth,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // The five-icon row stays at its own compact, calibrated
                // width and is centred in the wider column — stretching it
                // across the full column with spaceBetween would spread the
                // icons apart with the same ugly gaps _transportBlockWidth's
                // own doc comment exists to avoid.
                Center(
                  child: SizedBox(
                    width: MiniPlayerBar._transportBlockWidth,
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const ShuffleButton(
                          iconSize: MiniPlayerBar._transportIconSize,
                        ),
                        _transportButton(
                          context: context,
                          icon: Icons.skip_previous,
                          tooltip: 'Previous',
                          onPressed: () =>
                              ref.read(playerProvider.notifier).previous(),
                        ),
                        _playPauseButton(
                          context,
                          ref,
                          isPlaying: isPlaying,
                          isBuffering: isBuffering,
                        ),
                        _transportButton(
                          context: context,
                          icon: Icons.skip_next,
                          tooltip: 'Next',
                          onPressed: () =>
                              ref.read(playerProvider.notifier).next(),
                        ),
                        const RepeatModeButton(
                          iconSize: MiniPlayerBar._transportIconSize,
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 4),
                SizedBox(
                  width: miniPlayerSeekRowWidth(middleWidth),
                  child: const _MiniPlayerSeekRow(),
                ),
              ],
            ),
          ),
          Expanded(
            child: Align(
              alignment: Alignment.centerRight,
              child: LayoutBuilder(
                builder: (context, constraints) =>
                    _wideActionsRow(context, ref, constraints.maxWidth),
              ),
            ),
          ),
        ],
      );
    },
  );
}

/// Cover + title/artist — section 1 in both [_narrowContent] and
/// [_wideContent]. Pulled out once rather than kept as two copies now that
/// both shapes render it identically; only [_wideContent]'s surrounding
/// sections actually differ from [_narrowContent]'s.
///
/// v5.30.7 gives the cover and the title/artist text three independent
/// gestures instead of one shared one: the cover alone opens the full
/// player (the field report's explicit ask — previously the *entire* content
/// row did this via an outer InkWell in [MiniPlayerBar.build], which fought
/// with turning the title/artist into links), while title and artist each
/// link to their own detail page via [HoverLinkText] when the current track
/// carries the id for it. v5.32.0 adds the hover-scale affordance itself
/// (see [_HoverScaleCover]) that tells users the cover — and only the cover
/// — is what that gesture lives on.
Widget _nowPlayingInfo(
  BuildContext context, {
  required String title,
  required String artist,
  required String? albumId,
  required String? artistId,
  required Uri? localArtUri,
}) {
  return Row(
    children: [
      _HoverScaleCover(albumId: albumId, localArtUri: localArtUri),
      const SizedBox(width: 12),
      Expanded(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            HoverLinkText(
              text: title,
              style: TextStyle(
                color: context.skinColors.onSurface,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              onTap: albumId == null
                  ? null
                  : () => context.push(AppRoutes.albumDetailPath(albumId)),
            ),
            if (artist.isNotEmpty)
              HoverLinkText(
                text: artist,
                style: TextStyle(
                  color: context.skinColors.onSurfaceVariant,
                  fontSize: 11,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                onTap: artistId == null
                    ? null
                    : () => context.push(AppRoutes.artistDetailPath(artistId)),
              ),
          ],
        ),
      ),
    ],
  );
}

/// The cover art in [_nowPlayingInfo], grown slightly on hover — this bar's
/// own version of EchoMusic's `PlayerBar.vue` `group-hover:scale-110`
/// (`transition-transform duration-500`), the visual cue that the cover —
/// and, since v5.30.7, only the cover — is what opens the full player.
/// Needs its own [State] rather than a plain function (like the rest of this
/// file's small widget helpers) because "is the mouse over it right now" is
/// exactly the kind of transient, per-instance value a stateless function
/// has nowhere to keep.
class _HoverScaleCover extends StatefulWidget {
  const _HoverScaleCover({required this.albumId, required this.localArtUri});

  final String? albumId;
  final Uri? localArtUri;

  @override
  State<_HoverScaleCover> createState() => _HoverScaleCoverState();
}

class _HoverScaleCoverState extends State<_HoverScaleCover> {
  bool _hovering = false;

  /// 1.08, inside the field report's own "1.06-1.10" range but toward its
  /// gentler end: this cover (56px, see MiniPlayerBar._artworkSize) sits in
  /// a much denser row than EchoMusic's own player bar, immediately beside
  /// the title/artist text — the same absolute growth reads proportionally
  /// larger at this size, and needs headroom on the right so the scaled
  /// cover doesn't visually collide with that text.
  static const _hoverScale = 1.08;

  /// 200ms rather than EchoMusic's own 500ms — that duration suits a
  /// deliberate, tap-driven CSS transition; hover feedback in a desktop app
  /// reads as laggy at anywhere close to it. 200ms is snappy enough to feel
  /// like direct response to the pointer while still being a visible ease
  /// rather than a snap-cut.
  static const _hoverDuration = Duration(milliseconds: 200);

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hovering = true),
      onExit: (_) => setState(() => _hovering = false),
      child: GestureDetector(
        onTap: () => context.push(AppRoutes.player),
        child: AnimatedScale(
          scale: _hovering ? _hoverScale : 1.0,
          duration: _hoverDuration,
          curve: Curves.easeOut,
          child: MiniPlayerArtwork(
            albumId: widget.albumId,
            localArtUri: widget.localArtUri,
            size: MiniPlayerBar._artworkSize,
          ),
        ),
      ),
    );
  }
}

/// A plain prev/next transport icon. The play/pause button is not built from
/// this — it additionally needs the accent-filled circle, see
/// [_playPauseButton] — but shares the same [SpringInteraction] hover/press
/// motion, since the transport group is the highest-traffic control surface
/// in the app and is where that motion earns its keep.
Widget _transportButton({
  required BuildContext context,
  required IconData icon,
  required String tooltip,
  required VoidCallback onPressed,
}) {
  return SpringInteraction(
    child: IconButton(
      icon: Icon(icon, size: MiniPlayerBar._transportIconSize),
      color: context.skinColors.onSurfaceVariant,
      onPressed: onPressed,
      tooltip: tooltip,
    ),
  );
}

/// Play/pause, filled with the cover-derived accent colour rather than a
/// bare icon — matches the full player's own play/pause button
/// (`full_player_screen.dart`'s transport row) and both v5.30.6 reference
/// screenshots, which give play/pause a solid-circle treatment the
/// surrounding transport icons don't get. `Colors.white` for the glyph
/// mirrors that same existing button and this skin system's own
/// `FilledButtonTheme` (`foregroundColor: Colors.white` against a
/// `sakuraPink` fill) rather than computing a fresh contrast colour here —
/// sakuraPink is deliberately kept dark enough for white to read against it
/// in every skin (see `SkinColors.moonlitIndigo`'s doc comment on that exact
/// constraint), so a second contrast computation would just be re-deriving
/// a guarantee the skin system already provides.
Widget _playPauseButton(
  BuildContext context,
  WidgetRef ref, {
  required bool isPlaying,
  required bool isBuffering,
}) {
  return SpringInteraction(
    child: Container(
      decoration: BoxDecoration(
        color: context.skinColors.sakuraPink,
        shape: BoxShape.circle,
      ),
      child: IconButton(
        icon: Icon(
          isPlaying ? Icons.pause_rounded : Icons.play_arrow_rounded,
          size: 26,
          color: Colors.white,
        ),
        tooltip: isPlaying ? 'Pause' : 'Play',
        onPressed: isBuffering
            ? null
            : () => ref.read(playerProvider.notifier).togglePlayPause(),
      ),
    ),
  );
}

/// The sleep-timer control shared by [_narrowContent]'s section 3 and
/// [_wideActionsRow] — a top-level function rather than code duplicated
/// inline so the two shapes' timer buttons can't drift into looking alike
/// today and diverging tomorrow.
Widget _sleepTimerButton() {
  return Consumer(
    builder: (context, ref, _) {
      final timerState = ref.watch(sleepTimerProvider);
      final active = timerState.active;
      final remaining = timerState.remaining;
      final label = timerState.stopAfterTrack
          ? '♪'
          : (remaining != null
                ? MiniPlayerBar._formatRemaining(remaining)
                : null);
      return active
          ? TextButton.icon(
              icon: Icon(
                Icons.alarm,
                size: 18,
                color: Theme.of(context).colorScheme.primary,
              ),
              label: Text(
                label ?? '',
                style: TextStyle(
                  fontSize: 10,
                  color: Theme.of(context).colorScheme.primary,
                ),
              ),
              onPressed: () => _showSleepTimerSheet(context, ref),
            )
          : IconButton(
              icon: const Icon(Icons.alarm, size: 22),
              color: context.skinColors.onSurfaceVariant,
              tooltip: 'Sleep timer',
              onPressed: () => _showSleepTimerSheet(context, ref),
            );
    },
  );
}

/// The wide shape's section 3, once there is room for it: volume, the sleep
/// timer, and a queue entry point. [availableWidth] is this section's own
/// measured width (from the `LayoutBuilder` in [_wideContent]) — what
/// decides whether the volume slider fits inline or collapses to an
/// icon-only popover trigger, see [MiniPlayerBar._volumeCompactThreshold].
///
/// The queue button's own behaviour is [_openQueue] — see that function's
/// doc comment for why it differs by shell width instead of opening the
/// same thing everywhere.
Widget _wideActionsRow(
  BuildContext context,
  WidgetRef ref,
  double availableWidth,
) {
  final compactVolume = availableWidth < MiniPlayerBar._volumeCompactThreshold;
  return Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      VolumeControl(
        compact: compactVolume,
        iconSize: MiniPlayerBar._transportIconSize,
      ),
      _sleepTimerButton(),
      IconButton(
        icon: const Icon(
          Icons.queue_music,
          size: MiniPlayerBar._transportIconSize,
        ),
        color: context.skinColors.onSurfaceVariant,
        tooltip: 'Queue',
        onPressed: () => _openQueue(context, ref),
      ),
    ],
  );
}

/// Opens the play queue from the mini bar's own queue button.
///
/// Desktop — detected via [ShellChrome.of] returning non-null, the same
/// signal `settings_screen.dart` already uses for "is a desktop sidebar
/// actually present above this screen" — slides a docked drawer in over the
/// content column instead of navigating away. [ShellChrome] is only ever
/// provided by `_DesktopLayout` (shell_scaffold.dart, the >=1200dp shell),
/// so this reads directly off "is the real desktop shell above this bar"
/// rather than guessing from a breakpoint number that lives in a different
/// file. v5.30.5 explicitly deferred building any queue UI at all ("没有从
/// 主界面打开队列的通路就先跳播放页，不要为此新造队列 UI") — this is that deferred
/// work, landing on a *drawer* (not a second queue list) because
/// [QueueList] already exists and needs no duplicate.
///
/// Every narrower layout keeps the pre-v5.33.0 behaviour: push the full
/// player route. [_narrowContent] (the phone-width shape) carries no queue
/// button at all, so in practice this only matters for a tablet-width
/// window — its own mini bar can still be wide enough to render
/// [_wideActionsRow] (see [MiniPlayerBar._wideBreakpoint], which only checks
/// this bar's own measured width) even though there is no desktop sidebar
/// above it. A right-docked drawer competing with a tablet's much narrower
/// content column is a worse trade there than the full player's own queue
/// presentation (a docked panel above its split breakpoint, a bottom sheet
/// below it — see full_player_screen.dart's own queue button), which is
/// already a good experience and needs nothing new built for it.
void _openQueue(BuildContext context, WidgetRef ref) {
  if (ShellChrome.of(context) != null) {
    ref.read(queueDrawerOpenProvider.notifier).toggle();
  } else {
    context.push(AppRoutes.player);
  }
}

void _showSleepTimerSheet(BuildContext context, WidgetRef ref) {
  final timerActive = ref.read(sleepTimerProvider).active;
  showModalBottomSheet<void>(
    context: context,
    builder: (_) => SafeArea(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Padding(
            padding: EdgeInsets.all(16),
            child: Text(
              '睡眠定时器',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
          ),
          for (final mins in [15, 30, 45, 60])
            ListTile(
              title: Text('$mins 分钟'),
              onTap: () {
                ref
                    .read(sleepTimerProvider.notifier)
                    .startFixed(Duration(minutes: mins));
                Navigator.pop(context);
              },
            ),
          ListTile(
            title: const Text('当前曲目结束后停止'),
            onTap: () {
              ref.read(sleepTimerProvider.notifier).startAfterTrack();
              Navigator.pop(context);
            },
          ),
          if (timerActive)
            ListTile(
              title: const Text('取消定时器', style: TextStyle(color: Colors.red)),
              onTap: () {
                ref.read(sleepTimerProvider.notifier).cancel();
                Navigator.pop(context);
              },
            ),
        ],
      ),
    ),
  );
}

/// The narrow shape's plain top seek strip — unchanged since before v5.30.6
/// (see [_narrowContent]'s doc comment on why this phase left it alone
/// rather than trying to fit a shrunken [_MiniPlayerSeekRow] into the same
/// space).
class _MiniPlayerProgressBar extends ConsumerStatefulWidget {
  const _MiniPlayerProgressBar();

  @override
  ConsumerState<_MiniPlayerProgressBar> createState() =>
      _MiniPlayerProgressBarState();
}

class _MiniPlayerProgressBarState
    extends ConsumerState<_MiniPlayerProgressBar> {
  // Local override while the user is actively dragging, so incoming
  // positionStream updates don't fight the gesture and cause jitter.
  double? _dragValue;

  @override
  Widget build(BuildContext context) {
    final playerState = ref.watch(playerProvider);
    final disabled = playerState.isBuffering || playerState.isIdle;
    final maxMs = playerState.duration.inMilliseconds.toDouble() > 0
        ? playerState.duration.inMilliseconds.toDouble()
        : 1.0;
    final positionMs = playerState.position.inMilliseconds.toDouble().clamp(
      0.0,
      maxMs,
    );

    return SizedBox(
      height: 14,
      // v5.30.7: glides between positionStream ticks instead of snapping —
      // see seek_bar_effects.dart's positionTweenDuration doc comment for
      // why a TweenAnimationBuilder rather than a manual controller, and why
      // 260ms. Dragging bypasses it entirely (the builder's own value is
      // ignored below whenever _dragValue is set), so a drag still tracks
      // the pointer 1:1 with no animation lag.
      child: TweenAnimationBuilder<double>(
        tween: Tween<double>(begin: positionMs, end: positionMs),
        duration: positionTweenDuration,
        curve: Curves.linear,
        builder: (context, animatedPositionMs, _) {
          final value = (_dragValue ?? animatedPositionMs).clamp(0.0, maxMs);
          return SliderTheme(
            data: SliderTheme.of(context).copyWith(
              trackHeight: 2,
              activeTrackColor: context.skinColors.sakuraPink,
              inactiveTrackColor: context.skinColors.outline,
              thumbColor: context.skinColors.sakuraPinkLight,
              thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 4),
              overlayShape: const RoundSliderOverlayShape(overlayRadius: 10),
            ),
            child: Slider(
              value: value,
              max: maxMs,
              onChangeStart: disabled
                  ? null
                  : (v) => setState(() => _dragValue = v),
              onChanged: disabled
                  ? null
                  : (v) => setState(() => _dragValue = v),
              onChangeEnd: disabled
                  ? null
                  : (v) {
                      ref
                          .read(playerProvider.notifier)
                          .seekTo(Duration(milliseconds: v.toInt()));
                      setState(() => _dragValue = null);
                    },
            ),
          );
        },
      ),
    );
  }
}

/// The wide shape's seek control (v5.30.6): position/duration labels
/// flanking a thicker, softer-coloured slider, laid out as the second row of
/// the transport block (see [_wideContent]) instead of a full-bleed hairline
/// across the whole bar's top edge — the field report's exact complaint
/// ("时间条样式需要优化一下，太丑了") was that strip's 2px `outline`-coloured
/// track with no time reading at all.
///
/// Kept as a ConsumerStatefulWidget for the same reason
/// [_MiniPlayerProgressBar] already was — a local [_dragValue] override
/// while the user is actively dragging, so incoming positionStream updates
/// don't fight the gesture — now joined by [_hovering] for the thumb
/// reveal-on-hover treatment both v5.30.6 reference screenshots use: a
/// near-invisible track at rest, a visible thumb only once you're about to
/// interact with it.
class _MiniPlayerSeekRow extends ConsumerStatefulWidget {
  const _MiniPlayerSeekRow();

  @override
  ConsumerState<_MiniPlayerSeekRow> createState() => _MiniPlayerSeekRowState();
}

class _MiniPlayerSeekRowState extends ConsumerState<_MiniPlayerSeekRow> {
  double? _dragValue;
  bool _hovering = false;

  /// Horizontal breathing room between each time label and the track next to
  /// it. Through v5.30.6 there was none at all — the labels sat flush
  /// against the slider's own bounding box — which the v5.30.7 field report
  /// called out directly ("进度条前后的时间贴的太紧了").
  static const _timeLabelGap = 10.0;

  @override
  Widget build(BuildContext context) {
    final playerState = ref.watch(playerProvider);
    final disabled = playerState.isBuffering || playerState.isIdle;
    final isPlaying = playerState.isPlaying;
    final maxMs = playerState.duration.inMilliseconds.toDouble() > 0
        ? playerState.duration.inMilliseconds.toDouble()
        : 1.0;
    final positionMs = playerState.position.inMilliseconds.toDouble().clamp(
      0.0,
      maxMs,
    );
    // Thumb/overlay only appear once there's a reason to look at them —
    // hovering (mouse) or actively dragging (any input). At rest the track
    // reads as a plain line, matching both reference screenshots.
    final active = _hovering || _dragValue != null;
    final timeStyle = TextStyle(
      fontSize: 11,
      color: context.skinColors.onSurfaceVariant,
      // Fixed-width digits — without this, every second's redraw nudges the
      // label's own width (a '1' is narrower than an '8' in most fonts),
      // which shows up as the whole row twitching left-right once a second.
      fontFeatures: const [FontFeature.tabularFigures()],
    );

    return MouseRegion(
      onEnter: (_) => setState(() => _hovering = true),
      onExit: (_) => setState(() => _hovering = false),
      // v5.30.7: glides between positionStream ticks instead of snapping —
      // see seek_bar_effects.dart's positionTweenDuration doc comment for
      // why a TweenAnimationBuilder rather than a manual controller.
      // Dragging bypasses it (the animated value is ignored below whenever
      // _dragValue is set), so a drag still tracks the pointer 1:1.
      child: TweenAnimationBuilder<double>(
        tween: Tween<double>(begin: positionMs, end: positionMs),
        duration: positionTweenDuration,
        curve: Curves.linear,
        builder: (context, animatedPositionMs, _) {
          final value = (_dragValue ?? animatedPositionMs).clamp(0.0, maxMs);
          // v5.32.0: reserved off the *duration's* formatted width, not the
          // elapsed label's own — elapsed can never format longer than the
          // duration it's counting up to, so this is always the widest this
          // track's labels will need. Without a fixed slot, elapsed crossing
          // a digit-count boundary (9:59 -> 10:00) widens the label and
          // steals that many px from the slider below — v5.30.6's
          // tabularFigures only keeps *same*-digit-count redraws (9:58 ->
          // 9:59) from jittering; it can't do anything about the digit count
          // itself changing. Measured via TextPainter rather than counting
          // characters: the ':' separator's advance width isn't guaranteed
          // to match a tabular digit's, so a character count would over- or
          // under-reserve depending on the active font.
          final labelWidth = _timeLabelWidth(
            playerState.duration,
            timeStyle,
            MediaQuery.textScalerOf(context),
          );
          return Row(
            children: [
              SizedBox(
                width: labelWidth,
                // Right-aligned so the digits stay adjacent to the track —
                // any slack from reserving room for a wider duration sits on
                // the far side, away from the slider, instead of pushing the
                // label's own digits away from it.
                child: Text(
                  _formatSeekTime(Duration(milliseconds: value.toInt())),
                  textAlign: TextAlign.right,
                  // v5.33.0 fix: without these, a reserved width that's off
                  // by even a fraction of a pixel from what the paragraph
                  // layout actually needs (see _timeLabelWidth's doc
                  // comment) makes Text wrap the overflowing glyph onto a
                  // second line instead of the single line this slot is
                  // sized for — the field report's "0:43" rendering as
                  // "0:4"/"3".
                  maxLines: 1,
                  softWrap: false,
                  style: timeStyle,
                ),
              ),
              const SizedBox(width: _timeLabelGap),
              Expanded(
                child: SliderTheme(
                  data: SliderTheme.of(context).copyWith(
                    trackHeight: 4,
                    // A muted foreground rather than `outline` — outline
                    // reads as a border colour, not a "the rest of this
                    // control" colour, and was one of the two things the
                    // field report's "太丑了" was actually pointing at (the
                    // other being the old hairline's position, fixed by
                    // moving this row at all).
                    inactiveTrackColor: context.skinColors.onSurfaceVariant
                        .withValues(alpha: 0.24),
                    // Slightly desaturated relative to the sakuraPink used
                    // for icons/buttons elsewhere on this bar — both
                    // reference screenshots use a softer fill for the
                    // *track* specifically, saving full saturation for the
                    // one-button accent (play/pause) instead of spreading it
                    // across every pink element on the bar. The gradient
                    // (v5.30.7) lightens *toward* the thumb from this same
                    // muted base rather than introducing a second, brighter
                    // hue — "more layered", not "more saturated".
                    activeTrackColor: _desaturate(
                      context.skinColors.sakuraPink,
                      0.18,
                    ),
                    trackShape: const GradientSliderTrackShape(),
                    thumbColor: context.skinColors.sakuraPinkLight,
                    overlayColor: context.skinColors.sakuraPink.withValues(
                      alpha: 0.12,
                    ),
                    thumbShape: GlowingSliderThumbShape(
                      radius: active ? 7 : 0,
                      // Fixed regardless of `active` — see
                      // GlowingSliderThumbShape.getPreferredSize's doc
                      // comment. This is the v5.32.0 fix for "鼠标放上去出现
                      // 控制点会导致整个进度条收缩一部分": before, this field
                      // and `radius` were the same value, so revealing the
                      // thumb on hover also grew the track's own reserved
                      // layout inset by the same amount, visibly shrinking
                      // the track. Now the reservation never moves; only the
                      // painted thumb does.
                      maxRadius: 7,
                      // Only while the thumb itself is actually visible
                      // (active) — this row's whole design is "near
                      // invisible at rest, revealed on interact"; an ambient
                      // glow with no visible thumb to anchor it would fight
                      // that rather than extend it.
                      glowing: isPlaying && active,
                      glowColor: context.skinColors.sakuraPink.withValues(
                        alpha: 0.28,
                      ),
                    ),
                    // Constant overlayRadius for the same reason `maxRadius`
                    // above is constant: RoundSliderOverlayShape is a
                    // Flutter SDK shape whose own getPreferredSize returns
                    // Size.fromRadius(overlayRadius) unconditionally (not
                    // gated on activation), so toggling this 0/16 with
                    // `active` — the pre-v5.32.0 code here — fed the exact
                    // same max(thumbWidth, overlayWidth) term in
                    // BaseSliderTrackShape.getPreferredRect that the thumb
                    // fix above addresses, via a second shape we don't
                    // control the internals of. Flutter's own thumb
                    // hover/press activation animation already drives the
                    // *painted* overlay radius from 0 up to this constant
                    // (RoundSliderOverlayShape.paint tweens against
                    // activationAnimation, never against this field
                    // directly) — so dropping the ternary costs nothing
                    // visually and removes this shape's own contribution to
                    // the shrink bug.
                    overlayShape: const RoundSliderOverlayShape(
                      overlayRadius: 16,
                    ),
                  ),
                  child: Slider(
                    value: value,
                    max: maxMs,
                    onChangeStart: disabled
                        ? null
                        : (v) => setState(() => _dragValue = v),
                    onChanged: disabled
                        ? null
                        : (v) => setState(() => _dragValue = v),
                    onChangeEnd: disabled
                        ? null
                        : (v) {
                            ref
                                .read(playerProvider.notifier)
                                .seekTo(Duration(milliseconds: v.toInt()));
                            setState(() => _dragValue = null);
                          },
                  ),
                ),
              ),
              const SizedBox(width: _timeLabelGap),
              SizedBox(
                width: labelWidth,
                // Left-aligned, the mirror image of the elapsed label above
                // — its digits stay adjacent to the track's other end.
                child: Text(
                  _formatSeekTime(playerState.duration),
                  // See the elapsed-time label above for why these two are
                  // required, not stylistic.
                  maxLines: 1,
                  softWrap: false,
                  style: timeStyle,
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

/// Width to reserve for a time label showing [duration]'s own formatted
/// text — see the v5.32.0 doc comment where this is called for why both
/// labels share this one measurement rather than each sizing off its own
/// current text.
///
/// [textScaler] must be the same one the real [Text] widgets render with —
/// pass `MediaQuery.textScalerOf(context)`, never the default. Through
/// v5.32.0 this measured at the implicit default (`TextScaler.noScaling`,
/// i.e. 1.0) while the real `Text` below rendered at whatever the ambient
/// `MediaQuery` reported, so a user with a larger system font size got a
/// slot measured for 1.0x and text painted wider than that — the v5.33.0
/// field report's "0:43"/"3:05" splitting across two lines, with the last
/// character landing on the second one.
double _timeLabelWidth(
  Duration duration,
  TextStyle style,
  TextScaler textScaler,
) {
  final painter = TextPainter(
    text: TextSpan(text: _formatSeekTime(duration), style: style),
    textDirection: TextDirection.ltr,
    textScaler: textScaler,
  )..layout();
  // ceilToDouble() plus a further +2, not the raw sub-pixel double
  // `painter.width` returns. `SizedBox.width` reserves *exactly* what it is
  // given with zero tolerance, so even a rounding discrepancy of a few
  // hundredths of a pixel between this measurement pass and the paragraph
  // layout `Text` performs later at paint time is enough to push the last
  // glyph just past the box's edge — and with nothing capping line count,
  // that glyph wraps onto a second line rather than clipping invisibly.
  // maxLines/softWrap on the two call sites (see _MiniPlayerSeekRowState)
  // are the second half of this fix: they turn "reservation was a little
  // short" back into "clip a fraction of a pixel" instead of "wrap".
  return painter.width.ceilToDouble() + 2;
}

/// `mm:ss`, or `h:mm:ss` once a track runs past an hour — long-form local
/// files (audiobooks, DJ mixes, live sets) are common enough in a local
/// library that minutes alone would run to three digits and read as a typo
/// ("125:04") rather than a duration.
String _formatSeekTime(Duration d) {
  final totalSeconds = d.inSeconds;
  final hours = totalSeconds ~/ 3600;
  final minutes = (totalSeconds % 3600) ~/ 60;
  final seconds = totalSeconds % 60;
  final secondsStr = seconds.toString().padLeft(2, '0');
  if (hours > 0) {
    return '$hours:${minutes.toString().padLeft(2, '0')}:$secondsStr';
  }
  return '$minutes:$secondsStr';
}

/// Reduces [color]'s HSL saturation by [amount] (0-1), clamped so it never
/// goes negative. Used to soften the seek row's active track relative to the
/// fully-saturated sakuraPink used for icons/buttons — see
/// [_MiniPlayerSeekRowState.build]'s doc comment.
Color _desaturate(Color color, double amount) {
  final hsl = HSLColor.fromColor(color);
  return hsl
      .withSaturation((hsl.saturation - amount).clamp(0.0, 1.0))
      .toColor();
}

/// Mini player artwork thumbnail — shows the album cover, a local file's
/// embedded cover, or a fallback icon. The actual source-selection logic is
/// [TrackArtwork] (shared with the full player screen's own artwork tile
/// since v5.30.6); this just adds the rounded, background-filled box around
/// it.
///
/// [size] defaults to the pre-v5.30.0 44px box size (icon size and corner
/// radius scale proportionately from it rather than staying fixed, so a
/// future caller passing something other than 56 still gets sane
/// proportions); [MiniPlayerBar] itself always passes 56, EchoMusic's
/// `PlayerBar.vue` cover size.
class MiniPlayerArtwork extends StatelessWidget {
  const MiniPlayerArtwork({
    super.key,
    this.albumId,
    this.localArtUri,
    this.size = 44,
  });

  final String? albumId;

  /// Embedded cover art extracted from a guest-mode local file (a `file://`
  /// URI on `MediaItem.artUri`). Local tracks have no `albumId` at all, so
  /// before this was wired up (v5.30.6) this widget could only ever show the
  /// fallback icon for them, even though the exact same track's cover
  /// already rendered fine in every list row (which reads straight from the
  /// local DB rather than going through this widget).
  final Uri? localArtUri;

  final double size;

  @override
  Widget build(BuildContext context) {
    // EchoMusic's own fallback icon is 24px inside a 56px cover — scaling
    // from that ratio rather than hardcoding 24 keeps the icon proportionate
    // if a future call site passes a different size.
    final iconSize = size * (24 / 56);

    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: context.skinColors.surfaceContainer,
        // EchoMusic's Cover component uses a 10px radius at its 56px cover
        // size — scaled by the same ratio here so a future non-default
        // [size] gets a proportionate corner instead of a fixed one.
        borderRadius: BorderRadius.circular(size * (10 / 56)),
      ),
      clipBehavior: Clip.antiAlias,
      child: TrackArtwork(
        size: size,
        albumId: albumId,
        localArtUri: localArtUri,
        fallback: (context) => Icon(
          Icons.music_note,
          color: context.skinColors.onSurfaceVariant,
          size: iconSize,
        ),
      ),
    );
  }
}
