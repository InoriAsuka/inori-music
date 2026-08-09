// ignore_for_file: unnecessary_non_null_assertion
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/audio/sleep_timer_notifier.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/playback_mode_buttons.dart';
import 'package:inori_music/src/player/track_artwork.dart';
import 'package:inori_music/src/player/volume_control.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/floating_shadow.dart';
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

  /// Width shared by the wide shape's transport-group row and the seek row
  /// beneath it, so the two visually align as one block (same width, same
  /// centre axis) instead of the seek row either overflowing past the
  /// transport group or leaving mismatched slack beside it. A literal shared
  /// constant rather than an `IntrinsicWidth`/measure-and-mirror trick: the
  /// five transport-group icons are a fixed, known size, so their row's
  /// natural width is itself effectively a constant.
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
    // Embedded cover art extracted from a guest-mode local file — see
    // TrackArtwork's doc comment for why this and albumId are mutually
    // exclusive rather than both being checked against the same track.
    final localArtUri = mediaItem?.artUri;

    return Padding(
      padding: const EdgeInsets.fromLTRB(8, 0, 8, 8),
      // The shadow lives on this outer DecoratedBox, not inside the
      // Material below — see floatingShadow's doc comment on why a
      // clipping ancestor (this Material clips its own rounded corners)
      // must never sit between a BoxShadow and open air.
      child: DecoratedBox(
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
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                        child: InkWell(
                          onTap: () => context.push(AppRoutes.player),
                          borderRadius: BorderRadius.circular(8),
                          child: isWide
                              ? _wideContent(
                                  context,
                                  ref,
                                  title: title,
                                  artist: artist,
                                  albumId: albumId,
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
                                  localArtUri: localArtUri,
                                  isPlaying: isPlaying,
                                  isBuffering: isBuffering,
                                ),
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

/// The wide shape (v5.30.6): cover+title on the left; a shuffle/repeat-
/// flanked transport group with its own seek row underneath, centred; and
/// volume/timer/queue on the right. Sections 1 and 3 share equal Expanded
/// flex around section 2's fixed [MiniPlayerBar._transportBlockWidth] so it
/// stays centred regardless of how long the title or how the action group's
/// own width shifts between compact and expanded volume.
Widget _wideContent(
  BuildContext context,
  WidgetRef ref, {
  required String title,
  required String artist,
  required String? albumId,
  required Uri? localArtUri,
  required bool isPlaying,
  required bool isBuffering,
}) {
  return Row(
    children: [
      Expanded(
        child: _nowPlayingInfo(
          context,
          title: title,
          artist: artist,
          albumId: albumId,
          localArtUri: localArtUri,
        ),
      ),
      SizedBox(
        width: MiniPlayerBar._transportBlockWidth,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const ShuffleButton(iconSize: MiniPlayerBar._transportIconSize),
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
                const RepeatModeButton(
                  iconSize: MiniPlayerBar._transportIconSize,
                ),
              ],
            ),
            const SizedBox(height: 4),
            const _MiniPlayerSeekRow(),
          ],
        ),
      ),
      Expanded(
        child: Align(
          alignment: Alignment.centerRight,
          child: LayoutBuilder(
            builder: (context, constraints) =>
                _wideActionsRow(context, constraints.maxWidth),
          ),
        ),
      ),
    ],
  );
}

/// Cover + title/artist — section 1 in both [_narrowContent] and
/// [_wideContent]. Pulled out once rather than kept as two copies now that
/// both shapes render it identically; only [_wideContent]'s surrounding
/// sections actually differ from [_narrowContent]'s.
Widget _nowPlayingInfo(
  BuildContext context, {
  required String title,
  required String artist,
  required String? albumId,
  required Uri? localArtUri,
}) {
  return Row(
    children: [
      MiniPlayerArtwork(
        albumId: albumId,
        localArtUri: localArtUri,
        size: MiniPlayerBar._artworkSize,
      ),
      const SizedBox(width: 12),
      Expanded(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: TextStyle(
                color: context.skinColors.onSurface,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            if (artist.isNotEmpty)
              Text(
                artist,
                style: TextStyle(
                  color: context.skinColors.onSurfaceVariant,
                  fontSize: 11,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
          ],
        ),
      ),
    ],
  );
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
/// The queue button doesn't open a queue view of its own — there isn't one
/// outside [FullPlayerScreen] (docked beside the player on wide windows, a
/// bottom sheet otherwise), and building a second queue UI just for this
/// entry point would be a second thing to keep in sync with the first. It
/// opens the same route the bar's own title/artist tap already does.
Widget _wideActionsRow(BuildContext context, double availableWidth) {
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
        onPressed: () => context.push(AppRoutes.player),
      ),
    ],
  );
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
    final value = (_dragValue ?? positionMs).clamp(0.0, maxMs);

    return SizedBox(
      height: 14,
      child: SliderTheme(
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
          onChanged: disabled ? null : (v) => setState(() => _dragValue = v),
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
    final value = (_dragValue ?? positionMs).clamp(0.0, maxMs);
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
      child: Row(
        children: [
          Text(
            _formatSeekTime(Duration(milliseconds: value.toInt())),
            style: timeStyle,
          ),
          Expanded(
            child: SliderTheme(
              data: SliderTheme.of(context).copyWith(
                trackHeight: 4,
                // A muted foreground rather than `outline` — outline reads
                // as a border colour, not a "the rest of this control"
                // colour, and was one of the two things the field report's
                // "太丑了" was actually pointing at (the other being the old
                // hairline's position, fixed by moving this row at all).
                inactiveTrackColor: context.skinColors.onSurfaceVariant
                    .withValues(alpha: 0.24),
                // Slightly desaturated relative to the sakuraPink used for
                // icons/buttons elsewhere on this bar — both reference
                // screenshots use a softer fill for the *track* specifically,
                // saving full saturation for the one-button accent
                // (play/pause) instead of spreading it across every pink
                // element on the bar.
                activeTrackColor: _desaturate(
                  context.skinColors.sakuraPink,
                  0.18,
                ),
                thumbColor: context.skinColors.sakuraPinkLight,
                overlayColor: context.skinColors.sakuraPink.withValues(
                  alpha: 0.12,
                ),
                thumbShape: RoundSliderThumbShape(
                  enabledThumbRadius: active ? 7 : 0,
                ),
                overlayShape: RoundSliderOverlayShape(
                  overlayRadius: active ? 16 : 0,
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
          Text(_formatSeekTime(playerState.duration), style: timeStyle),
        ],
      ),
    );
  }
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
