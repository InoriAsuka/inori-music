import 'dart:io';
import 'dart:math' as math;

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, TargetPlatform;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/audio/eq_notifier.dart';
import 'package:inori_music/src/audio/sleep_timer_notifier.dart';
import 'package:inori_music/src/audio/speed_notifier.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/favorites/track_favorite_notifier.dart';
import 'package:inori_music/src/local_library/audio_quality.dart';
import 'package:inori_music/src/local_library/local_library_db.dart';
import 'package:inori_music/src/local_library/local_library_notifier.dart'
    show localTrackIdPrefix;
import 'package:inori_music/src/lyrics/bilingual_lyrics_notifier.dart';
import 'package:inori_music/src/lyrics/local_lyrics_provider.dart';
import 'package:inori_music/src/lyrics/lyric_line.dart';
import 'package:inori_music/src/lyrics/lyrics_background.dart';
import 'package:inori_music/src/lyrics/lyrics_provider.dart';
import 'package:inori_music/src/player/cover_flow_artwork.dart';
import 'package:inori_music/src/player/cover_flow_mode_provider.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/karaoke_screen.dart';
import 'package:inori_music/src/player/player_state.dart' as ps;
import 'package:inori_music/src/shared/desktop_integration.dart';
import 'package:inori_music/src/shared/system_titlebar_provider.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/glass_panel.dart';
import 'package:inori_music/src/shared/widgets/spring_interaction.dart';

/// Full-screen player overlay with progress bar, controls, and queue sheet.
class FullPlayerScreen extends ConsumerStatefulWidget {
  const FullPlayerScreen({super.key});

  @override
  ConsumerState<FullPlayerScreen> createState() => _FullPlayerScreenState();
}

/// Which panel, if any, is docked beside the player.
enum _SidePanel { none, lyrics, queue }

/// Cover edge length for the wide (Apple Music-aligned) player layout, scaled
/// to the available region instead of fixed. The ratios below are measured
/// off four differently-sized Apple Music screenshots (requirement.md
/// v5.29.0): the cover consistently spans about 42% of the region's width or
/// 40% of its height, whichever is tighter, clamped to a sane on-screen size
/// at either end.
///
/// Only used above [_FullPlayerScreenState._splitBreakpoint] — the narrow
/// layout keeps its own fixed [_FullPlayerScreenState._narrowArtworkSize]
/// instead, since every reference screenshot behind this formula was a wide
/// desktop window and the same ratio would shrink a phone-width cover below
/// its pre-v5.29.0 size.
double playerArtworkSize({
  required double regionWidth,
  required double regionHeight,
}) => math.min(regionWidth * 0.42, regionHeight * 0.40).clamp(160.0, 420.0);

/// Width of the title/artist block, progress bar and transport controls,
/// derived from the cover rather than stretched across the whole region. The
/// same reference screenshots put this ratio at 1.38-1.52x the cover's
/// width; 1.45 is their midpoint.
///
/// Clamped to a [_controlWidthFloor] usability floor and, via `math.max`, a
/// ceiling of `regionWidth - 48` that itself can never drop below that floor
/// — a plain `.clamp(floor, regionWidth - 48)` would invert (throw
/// `RangeError`) once `regionWidth` falls under `floor + 48`, which the left
/// half of a split layout reaches on anything narrower than an ultra-wide
/// monitor.
double playerControlWidth({
  required double artworkSize,
  required double regionWidth,
}) => (artworkSize * 1.45).clamp(
  _controlWidthFloor,
  math.max(_controlWidthFloor, regionWidth - 48),
);

/// Usability floor for [playerControlWidth].
///
/// Not the 280 the shape of this formula might suggest — that was this
/// value's first estimate, and measuring the actual compact-mode transport
/// bar (`_ControlDensity`) showed it does not fit in 280. The bar is three
/// unequal groups — secondaries left, the transport trio centre, secondaries
/// right — laid out as two `Expanded` siblings around a fixed-size middle so
/// the trio stays centred regardless of how many secondaries sit on each
/// side (see the transport `Row` in [_FullPlayerScreenState._playerBlock]).
/// That symmetry means the *narrower* side's spare space is not available to
/// the wider one: at 280 the middle trio alone (3 x 36px) leaves 2 x 86px for
/// the sides, but the right group (speed text button + sleep + favourite)
/// measured 116.4px even in compact mode — the left group's unused slack
/// cannot cover that, so the right side overflowed by 40px. 400 leaves each
/// side roughly 132px, comfortably above that measurement. Raising this
/// floor is preferred over reworking the trio-centring split: that symmetry
/// is what keeps the transport trio centred in every other window size (see
/// test/full_player_layout_test.dart's grouping and centring cases), and is
/// out of scope for a floor that only ever binds at the narrowest split
/// windows.
const _controlWidthFloor = 400.0;

class _FullPlayerScreenState extends ConsumerState<FullPlayerScreen> {
  late final PageController _pageController;
  int _pageIndex = 0;

  _SidePanel _sidePanel = _SidePanel.none;

  /// Below this the window cannot give both the player and a panel a usable
  /// width, so lyrics and the queue keep their pre-v5.28.0 presentation —
  /// a pushed route and a bottom sheet.
  static const _splitBreakpoint = 900.0;

  /// Below this control-block width, eight transport controls cannot sit on
  /// one row at full size.
  ///
  /// Measured against the control block's own width, not the window's. Before
  /// v5.29.0 the side panel was a fixed 380px, so the window width was a
  /// reasonable proxy for how much room the controls had; now the panel takes
  /// half the window, and a window well above this number can still hand the
  /// control block a region narrower than it — the same kind of basis
  /// mismatch that let v5.28.0's spaceEvenly overflow go unnoticed until it
  /// was actually measured. Calibrated against the mid-width split-panel case
  /// in test/full_player_layout_test.dart; adjust there first if this value
  /// ever needs to move.
  static const _compactControlsBreakpoint = 480.0;

  /// Cover edge length for the narrow (single-column, phone-width) layout.
  /// Predates [playerArtworkSize] and stays a fixed value rather than folding
  /// into that formula — see its doc comment for why.
  static const _narrowArtworkSize = 280.0;

  @override
  void initState() {
    super.initState();
    _pageController = PageController();
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(playerProvider);
    final isPlaying = state.isPlaying;
    final isBuffering = state.isBuffering;
    final trackId = state.mediaItem?.id ?? '';
    final position = ref.watch(playerProvider.select((s) => s.position));

    final width = MediaQuery.sizeOf(context).width;
    final canSplit = width >= _splitBreakpoint;
    final splitPanel = canSplit ? _sidePanel : _SidePanel.none;

    return Scaffold(
      backgroundColor: context.skinColors.background,
      // The whole player now sits on the cover-derived colour field, not just
      // the lyrics tab inside it — the artwork drives the page's colour and
      // the controls read as frosted panes over it. With no artwork this is a
      // plain skin-coloured background and nothing below changes.
      body: LyricsBackground(
        albumId: state.mediaItem?.extras?['albumId'] as String?,
        localArtUri: state.mediaItem?.artUri,
        child: SafeArea(
          child: Column(
            children: [
              // Top bar. Wrapped in its own Builder so `context` here is a
              // *descendant* of the SkinScope LyricsBackground inserts a few
              // widgets up — without it, every context.skinColors call below
              // resolved via the FullPlayerScreen element itself, which sits
              // *above* that SkinScope in the tree (inherited-widget lookups
              // only walk up, never down into what build() is about to
              // return). That is a real, separate bug from the one this
              // block's scrim addresses: the top bar was never actually
              // picking up the artwork-derived colours at all, in either
              // brightness branch, regardless of the cover — confirmed by
              // probing it directly (the resolved colour was byte-identical
              // to the ambient skin's fixed onSurfaceVariant no matter what
              // palette was fed in). A muted grey sized for Sakura Dusk's
              // light background, painted over a colour field that can be
              // any brightness, is unreadable exactly the way the v5.29.0
              // field report described — this was the actual root cause, not
              // a global-vs-local mismatch in an already-dynamic colour.
              Builder(
                builder: (context) => _topBarScrim(
                  context: context,
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 8,
                    ),
                    child: Row(
                      children: [
                        // macOS draws its traffic lights over whatever Flutter
                        // renders, at a fixed offset from the window's top-left.
                        // This screen is a full-bleed route with no DesktopAppBar,
                        // so without a gutter the close button sits underneath
                        // them — the same reservation DesktopAppBar already makes.
                        if (_needsMacTrafficLightGutter(ref))
                          const SizedBox(width: 70),
                        IconButton(
                          icon: Icon(
                            Icons.keyboard_arrow_down,
                            size: 32,
                            color: context.skinColors.onBackground,
                          ),
                          tooltip: 'Close player',
                          onPressed: () => Navigator.of(context).maybePop(),
                        ),
                        Expanded(
                          child: Text(
                            AppLocalizations.of(context).nowPlaying,
                            textAlign: TextAlign.center,
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: context.skinColors.onSurfaceVariant,
                            ),
                          ),
                        ),
                        IconButton(
                          icon: Icon(
                            Icons.queue_music,
                            color: context.skinColors.onSurfaceVariant,
                          ),
                          tooltip: 'Queue',
                          onPressed: () => canSplit
                              ? _togglePanel(_SidePanel.queue)
                              : _showQueueSheet(context, ref),
                        ),
                        IconButton(
                          icon: Icon(
                            Icons.mic_external_on,
                            color: context.skinColors.onSurfaceVariant,
                          ),
                          tooltip: 'Karaoke',
                          // Only gated on "is anything playing at all". It used to
                          // also require a local track to carry an embedded lyrics
                          // tag, which silently disabled the button for every
                          // instrumental / untagged local file — a dead control with
                          // no explanation, and the only way into the lyrics screen.
                          // KaraokeScreen already renders "No lyrics available" over
                          // the artwork backdrop, which is exactly what a server
                          // track with no lyrics has always done.
                          onPressed: trackId.isEmpty
                              ? null
                              : () => canSplit
                                    ? _togglePanel(_SidePanel.lyrics)
                                    : Navigator.of(context).push(
                                        MaterialPageRoute<void>(
                                          builder: (_) => const KaraokeScreen(),
                                          fullscreenDialog: true,
                                        ),
                                      ),
                        ),
                        // Technical detail (sample rate/bitrate/format) — only
                        // meaningful for local files; the server catalog has no
                        // equivalent metadata to show (out of scope, see
                        // requirement.md v5.19.0).
                        if (trackId.startsWith(localTrackIdPrefix))
                          IconButton(
                            icon: Icon(
                              Icons.info_outline,
                              color: context.skinColors.onSurfaceVariant,
                            ),
                            tooltip: '详情',
                            onPressed: () =>
                                _showLocalTrackDetails(context, trackId),
                          ),
                        // EQ icon button
                        Consumer(
                          builder: (ctx, ref2, _) {
                            final eqEnabled = ref2
                                .watch(eqNotifierProvider)
                                .enabled;
                            return IconButton(
                              icon: Icon(
                                Icons.equalizer,
                                color: eqEnabled
                                    ? context.skinColors.sakuraPinkLight
                                    : context.skinColors.onSurfaceVariant,
                              ),
                              tooltip: 'Equalizer',
                              onPressed: () => ref2
                                  .read(eqNotifierProvider.notifier)
                                  .setEnabled(!eqEnabled),
                            );
                          },
                        ),
                      ],
                    ),
                  ),
                ),
              ),

              // Centred when nothing is open; split once lyrics or the queue
              // are shown, with the player itself staying centred in whatever
              // width is left. Apple Music's shape, and it means the same
              // block serves both states instead of two layouts drifting
              // apart. Only wide windows split — below the breakpoint the
              // panel would leave neither half usable, so there it stays a
              // sheet / pushed route as before. The split itself is
              // proportional (flex 1:1) rather than a fixed-width sidebar —
              // see [_playerBlock] and [playerControlWidth]: a fixed sidebar
              // left the control block stretching to fill whatever width was
              // left over, with no relationship to the cover's own size.
              Expanded(
                child: Row(
                  children: [
                    Expanded(
                      flex: 1,
                      child: LayoutBuilder(
                        builder: (context, constraints) => _playerBlock(
                          context,
                          state: state,
                          isPlaying: isPlaying,
                          isBuffering: isBuffering,
                          trackId: trackId,
                          position: position,
                          canSplit: canSplit,
                          regionWidth: constraints.maxWidth,
                          regionHeight: constraints.maxHeight,
                        ),
                      ),
                    ),
                    if (splitPanel != _SidePanel.none)
                      Expanded(
                        flex: 1,
                        // The panel's Expanded still claims half the row, so
                        // the player side keeps exactly half no matter how
                        // wide the window gets; ConstrainedBox only stops the
                        // panel's *content* from stretching past a readable
                        // width once that half becomes very wide. Align is
                        // what lets the space freed by the constraint actually
                        // show, instead of leaving a blank gap where the
                        // panel used to reach.
                        child: Align(
                          alignment: Alignment.centerRight,
                          child: ConstrainedBox(
                            constraints: const BoxConstraints(maxWidth: 560),
                            child: _PlayerSidePanel(
                              panel: splitPanel,
                              trackId: trackId,
                              position: position,
                              onClose: () =>
                                  setState(() => _sidePanel = _SidePanel.none),
                            ),
                          ),
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// Cover, title/artist, progress bar and transport controls as one block —
  /// its width derived from the cover, and the whole thing centred in
  /// whatever [regionWidth]/[regionHeight] it is handed. Unsplit, that region
  /// is the whole window; split, it is the left half. The same method serves
  /// both states, so there is one layout to keep in sync rather than two that
  /// can drift apart.
  Widget _playerBlock(
    BuildContext context, {
    required ps.PlayerState state,
    required bool isPlaying,
    required bool isBuffering,
    required String trackId,
    required Duration position,
    required bool canSplit,
    required double regionWidth,
    required double regionHeight,
  }) {
    final coverSize = canSplit
        ? playerArtworkSize(
            regionWidth: regionWidth,
            regionHeight: regionHeight,
          )
        : _narrowArtworkSize;
    final controlWidth = playerControlWidth(
      artworkSize: coverSize,
      regionWidth: regionWidth,
    );
    final compactControls = controlWidth < _compactControlsBreakpoint;
    // Cover Flow only replaces the wide-layout artwork tile (see
    // cover_flow_artwork.dart's doc comment): the narrow layout's PageView
    // already owns horizontal swipe gestures for paging to lyrics, and a
    // second horizontal gesture for flowing through the queue underneath it
    // would fight the first rather than compose with it. A single-track
    // queue has no neighbours to flow through, so it falls back to the plain
    // tile too rather than rendering a Cover Flow of one.
    final coverFlowEnabled =
        canSplit && ref.watch(coverFlowModeProvider) && state.queue.length > 1;

    return Column(
      children: [
        const Spacer(),

        // Wide windows dock lyrics in the side panel, so the artwork/lyrics
        // PageView — and the page-dot indicator that came with it — only
        // earns its keep below the breakpoint. Above it the dots were a
        // stray mark under the cover with nothing to page to.
        if (coverFlowEnabled)
          CoverFlowArtwork(
            itemCount: state.queue.length,
            currentIndex: state.currentIndex,
            centerSize: coverSize,
            width: regionWidth,
            itemBuilder: (context, index, size) => ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: _FullPlayerArtwork(
                size: size,
                albumId: state.queue[index].extras?['albumId'] as String?,
                localArtUri: state.queue[index].artUri,
              ),
            ),
            onSelect: (index) => ref
                .read(playerProvider.notifier)
                .playQueue(
                  state.queue.map((m) => m.id).toList(),
                  initialIndex: index,
                ),
          )
        else if (canSplit)
          _artworkTile(context, state, size: coverSize)
        else ...[
          SizedBox(
            width: _narrowArtworkSize,
            height: _narrowArtworkSize,
            child: PageView(
              controller: _pageController,
              onPageChanged: (i) => setState(() => _pageIndex = i),
              children: [
                _artworkTile(context, state, size: _narrowArtworkSize),
                _LyricsPage(trackId: trackId, position: position),
              ],
            ),
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: List.generate(2, (i) {
              return Container(
                margin: const EdgeInsets.symmetric(horizontal: 3),
                width: _pageIndex == i ? 10 : 6,
                height: 6,
                decoration: BoxDecoration(
                  color: _pageIndex == i
                      ? context.skinColors.sakuraPink
                      : context.skinColors.onSurfaceVariant.withValues(
                          alpha: 0.4,
                        ),
                  borderRadius: BorderRadius.circular(3),
                ),
              );
            }),
          ),
        ],

        const Spacer(),

        // Title / artist. Width matches the control block below rather than
        // a fixed inset from the region's edges — the whole point of this
        // block is that title, progress bar and transport controls share one
        // width axis instead of each being sized independently.
        SizedBox(
          width: controlWidth,
          child: Column(
            children: [
              Text(
                state.mediaItem?.title ?? 'Unknown Track',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w700,
                  color: context.skinColors.onBackground,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 4),
              Text(
                state.mediaItem?.artist ?? '',
                style: TextStyle(
                  fontSize: 15,
                  color: context.skinColors.onSurfaceVariant,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),

        const SizedBox(height: 24),

        // Seek bar and transport share one frosted pane, so they read as
        // a single control surface floating over the cover's colour field
        // instead of loose widgets scattered across it.
        SizedBox(
          width: controlWidth,
          child: GlassPanel(
            padding: const EdgeInsets.fromLTRB(14, 6, 14, 2),
            child: _ControlDensity(
              compact: compactControls,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  SliderTheme(
                    data: SliderTheme.of(context).copyWith(
                      trackHeight: 3,
                      thumbShape: const RoundSliderThumbShape(
                        enabledThumbRadius: 6,
                      ),
                      overlayShape: const RoundSliderOverlayShape(
                        overlayRadius: 14,
                      ),
                    ),
                    child: Slider(
                      value: isBuffering
                          ? 0
                          : state.position.inMilliseconds.toDouble().clamp(
                              0,
                              state.duration.inMilliseconds.toDouble() > 0
                                  ? state.duration.inMilliseconds.toDouble()
                                  : 1,
                            ),
                      max: state.duration.inMilliseconds.toDouble() > 0
                          ? state.duration.inMilliseconds.toDouble()
                          : 1,
                      onChanged: isBuffering
                          ? null
                          : (v) => ref
                                .read(playerProvider.notifier)
                                .seekTo(Duration(milliseconds: v.toInt())),
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          _formatDuration(state.position),
                          style: TextStyle(
                            fontSize: 12,
                            color: context.skinColors.onSurfaceVariant,
                          ),
                        ),
                        Text(
                          _formatDuration(state.duration),
                          style: TextStyle(
                            fontSize: 12,
                            color: context.skinColors.onSurfaceVariant,
                          ),
                        ),
                      ],
                    ),
                  ),
                  // Three groups rather than eight equal slots.
                  // spaceEvenly gave play/prev/next exactly the same
                  // visual weight as the sleep timer, so the controls
                  // reached for constantly were impossible to find by
                  // muscle memory. Secondary controls hug the edges; the
                  // transport trio sits tight in the middle and stays
                  // centred however many secondaries each side has.
                  //
                  // Eight controls do not fit one row at phone
                  // widths — they overflowed under the old
                  // spaceEvenly too. Rather than dropping or
                  // clipping any, narrow windows shrink every
                  // button's footprint through the surrounding
                  // theme, which keeps the grouping identical
                  // at every size instead of maintaining two
                  // arrangements that can drift apart.
                  Row(
                    children: [
                      Expanded(
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.start,
                          children: [
                            IconButton(
                              icon: Icon(
                                Icons.repeat,
                                color: state.repeat != ps.RepeatMode.none
                                    ? context.skinColors.sakuraPinkLight
                                    : context.skinColors.onSurfaceVariant,
                              ),
                              onPressed: () {
                                final notifier = ref.read(
                                  playerProvider.notifier,
                                );
                                switch (state.repeat) {
                                  case ps.RepeatMode.none:
                                    notifier.setRepeat(ps.RepeatMode.all);
                                    break;
                                  case ps.RepeatMode.all:
                                    notifier.setRepeat(ps.RepeatMode.one);
                                    break;
                                  case ps.RepeatMode.one:
                                    notifier.setRepeat(ps.RepeatMode.none);
                                    break;
                                }
                              },
                              tooltip: 'Repeat: ${state.repeat.name}',
                            ),
                            Consumer(
                              builder: (context2, ref2, child2) {
                                final isShuffle = ref2
                                    .watch(playerProvider)
                                    .shuffle;
                                return IconButton(
                                  icon: Icon(
                                    Icons.shuffle,
                                    color: isShuffle
                                        ? context.skinColors.sakuraPinkLight
                                        : context.skinColors.onSurfaceVariant,
                                  ),
                                  onPressed: () => ref2
                                      .read(playerProvider.notifier)
                                      .setShuffle(!isShuffle),
                                  tooltip: 'Shuffle',
                                );
                              },
                            ),
                          ],
                        ),
                      ),
                      // The transport trio carries the spring hover/press
                      // motion (see SpringInteraction); the surrounding
                      // secondary controls are deliberately left plain so
                      // the primary actions stay the ones that respond.
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          SpringInteraction(
                            child: IconButton(
                              icon: Icon(
                                Icons.skip_previous,
                                size: 36,
                                color: context.skinColors.onSurface,
                              ),
                              onPressed: () =>
                                  ref.read(playerProvider.notifier).previous(),
                            ),
                          ),
                          // Play / Pause button
                          SpringInteraction(
                            child: Container(
                              decoration: BoxDecoration(
                                color: context.skinColors.sakuraPink,
                                shape: BoxShape.circle,
                              ),
                              child: IconButton(
                                icon: Icon(
                                  isBuffering
                                      ? Icons.play_arrow_rounded
                                      : (isPlaying
                                            ? Icons.pause_rounded
                                            : Icons.play_arrow_rounded),
                                  size: 36,
                                  color: Colors.white,
                                ),
                                onPressed: isBuffering
                                    ? null
                                    : () => ref
                                          .read(playerProvider.notifier)
                                          .togglePlayPause(),
                              ),
                            ),
                          ),
                          SpringInteraction(
                            child: IconButton(
                              icon: Icon(
                                Icons.skip_next,
                                size: 36,
                                color: context.skinColors.onSurface,
                              ),
                              onPressed: () =>
                                  ref.read(playerProvider.notifier).next(),
                            ),
                          ),
                        ],
                      ),
                      Expanded(
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.end,
                          children: [
                            // Speed control button
                            Consumer(
                              builder: (context, ref, _) {
                                final speed = ref.watch(speedNotifierProvider);
                                return TextButton(
                                  onPressed: () =>
                                      _showSpeedSheet(context, ref),
                                  child: Text(
                                    '$speed×',
                                    style: const TextStyle(fontSize: 14),
                                  ),
                                );
                              },
                            ),
                            // Sleep timer button
                            Consumer(
                              builder: (context, ref, _) {
                                final timerState = ref.watch(
                                  sleepTimerProvider,
                                );
                                final active = timerState.active;
                                return IconButton(
                                  icon: Icon(
                                    Icons.bedtime,
                                    color: active
                                        ? context.skinColors.sakuraPinkLight
                                        : context.skinColors.onSurfaceVariant,
                                  ),
                                  tooltip: 'Sleep timer',
                                  onPressed: () =>
                                      _showSleepTimerSheet(context, ref),
                                );
                              },
                            ),
                            // Favorite button — wrapped in Consumer so icon and onPressed
                            // always use the same live trackId from the reactive ref.
                            Consumer(
                              builder: (context2, ref2, child2) {
                                final trackId = ref2
                                    .watch(playerProvider)
                                    .mediaItem
                                    ?.id;
                                // Local (guest-mode) tracks have no server-side favorite state.
                                final isLocal =
                                    trackId?.startsWith(localTrackIdPrefix) ??
                                    false;
                                final isFav = (trackId != null && !isLocal)
                                    ? ref2.watch(trackFavoriteProvider(trackId))
                                    : false;
                                return IconButton(
                                  icon: Icon(
                                    isFav
                                        ? Icons.favorite
                                        : Icons.favorite_border,
                                    color: isFav
                                        ? context.skinColors.accentPink
                                        : (trackId != null && !isLocal
                                              ? context.skinColors.onSurface
                                              : context
                                                    .skinColors
                                                    .onSurfaceVariant),
                                  ),
                                  onPressed: (trackId == null || isLocal)
                                      ? null
                                      : () => ref2
                                            .read(
                                              trackFavoriteProvider(
                                                trackId,
                                              ).notifier,
                                            )
                                            .toggle(),
                                  tooltip: 'Favorite',
                                );
                              },
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),

        const SizedBox(height: 32),
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

  void _showSpeedSheet(BuildContext context, WidgetRef ref) {
    final current = ref.read(speedNotifierProvider);
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Padding(
              padding: EdgeInsets.all(16),
              child: Text(
                '播放速度',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
            ),
            for (final s in speedPresets)
              ListTile(
                title: Text('$s×'),
                trailing: s == current ? const Icon(Icons.check) : null,
                onTap: () {
                  ref.read(speedNotifierProvider.notifier).setSpeed(s);
                  Navigator.pop(context);
                },
              ),
          ],
        ),
      ),
    );
  }

  Future<void> _showLocalTrackDetails(
    BuildContext context,
    String trackId,
  ) async {
    final track = await LocalLibraryDb.instance.query(trackId);
    if (!context.mounted) return;
    final rows = <(String, String)>[
      ('格式', track?.fileFormat ?? '未知'),
      (
        '采样率',
        track?.sampleRate != null
            ? '${(track!.sampleRate! / 1000).toStringAsFixed(1)} kHz'
            : '未知',
      ),
      // bitrate is stored in bps (audio_metadata_reader's own unit) — kbps
      // is what listeners actually recognize (e.g. "320" for MP3).
      (
        '码率',
        track?.bitrate != null
            ? '${(track!.bitrate! / 1000).round()} kbps'
            : '未知',
      ),
      (
        '文件大小',
        track != null
            ? '${(track.sizeBytes / (1024 * 1024)).toStringAsFixed(1)} MB'
            : '未知',
      ),
      if (track?.genre != null) ('流派', track!.genre!),
    ];
    await showModalBottomSheet<void>(
      context: context,
      backgroundColor: context.skinColors.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text('曲目详情', style: Theme.of(ctx).textTheme.headlineSmall),
                  if (isHiResAudio(
                    sampleRate: track?.sampleRate,
                    bitrate: track?.bitrate,
                  )) ...[
                    const SizedBox(width: 10),
                    const _HiResBadge(),
                  ],
                ],
              ),
            ),
            const Divider(height: 1),
            for (final (label, value) in rows)
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 10,
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        label,
                        style: TextStyle(
                          color: ctx.skinColors.onSurfaceVariant,
                        ),
                      ),
                    ),
                    Text(
                      value,
                      style: TextStyle(color: ctx.skinColors.onSurface),
                    ),
                  ],
                ),
              ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  /// The cover tile. Shared by the wide layout (which shows it alone) and the
  /// narrow one (which pages between it and the lyrics).
  Widget _artworkTile(
    BuildContext context,
    ps.PlayerState state, {
    required double size,
  }) => Container(
    width: size,
    height: size,
    decoration: BoxDecoration(
      color: context.skinColors.surfaceVariant,
      borderRadius: BorderRadius.circular(16),
      boxShadow: [
        BoxShadow(
          color: context.skinColors.sakuraPink.withValues(alpha: 0.15),
          blurRadius: 32,
          offset: const Offset(0, 8),
        ),
      ],
    ),
    child: ClipRRect(
      borderRadius: BorderRadius.circular(16),
      child: _FullPlayerArtwork(
        size: size,
        albumId: state.mediaItem?.extras?['albumId'] as String?,
        localArtUri: state.mediaItem?.artUri,
      ),
    ),
  );

  /// Opens [panel], or closes it if it is already the one showing — the
  /// button doubles as the way out, so the panel never needs its own
  /// dismiss affordance to be reachable.
  void _togglePanel(_SidePanel panel) => setState(
    () => _sidePanel = _sidePanel == panel ? _SidePanel.none : panel,
  );

  /// Whether macOS will paint its traffic lights over this screen's top-left
  /// corner. False on every other platform, and false when the user has opted
  /// into the OS title bar — then the lights live in a real title bar above
  /// the content instead of on top of it.
  static bool _needsMacTrafficLightGutter(WidgetRef ref) {
    if (!DesktopIntegration.isDesktop) return false;
    if (defaultTargetPlatform != TargetPlatform.macOS) return false;
    return !ref.watch(systemTitleBarProvider);
  }

  void _showQueueSheet(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => DraggableScrollableSheet(
        initialChildSize: 0.6,
        maxChildSize: 0.9,
        minChildSize: 0.3,
        builder: (_, controller) => Container(
          decoration: BoxDecoration(
            color: context.skinColors.surface,
            borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
          ),
          child: Column(
            children: [
              const SizedBox(height: 12),
              Container(
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: context.skinColors.outlineVariant,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(12),
                child: Text(
                  'Queue',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                    color: context.skinColors.onBackground,
                  ),
                ),
              ),
              Expanded(child: _QueueList(scrollController: controller)),
            ],
          ),
        ),
      ),
    );
  }

  static String _formatDuration(Duration d) {
    final mins = d.inMinutes;
    final secs = d.inSeconds % 60;
    return '$mins:${secs.toString().padLeft(2, '0')}';
  }
}

/// Guarantees the top bar's own title/icons stay legible regardless of what
/// locally sits behind them.
///
/// [artworkOverlaySkin] already decides ink colour (white or near-black) from
/// the backdrop's *global* dominant-colour luminance, but
/// [CoverFluidBackground] is a moving, multi-swatch colour field — the strip
/// directly behind the top bar can easily land on a patch that disagrees with
/// that global call. A gradient scrim in the same polarity as the resolved
/// ink (dark behind white ink, light behind dark ink) swamps that local
/// variance instead of trying to read it — the same technique most players
/// use behind a nav bar sitting over artwork. [context] must already be a
/// descendant of the derived overlay SkinScope (see the Builder wrapping this
/// call in [_FullPlayerScreenState.build]) or this reads the wrong ink.
Widget _topBarScrim({required BuildContext context, required Widget child}) {
  final inkIsLight = context.skinColors.onBackground.computeLuminance() > 0.5;
  final scrim = inkIsLight ? Colors.black : Colors.white;
  return DecoratedBox(
    decoration: BoxDecoration(
      gradient: LinearGradient(
        begin: Alignment.topCenter,
        end: Alignment.bottomCenter,
        colors: [scrim.withValues(alpha: 0.5), scrim.withValues(alpha: 0.0)],
      ),
    ),
    child: child,
  );
}

/// Large artwork widget for the full player screen.
/// Watches [artworkUrlProvider] for the album and shows CachedNetworkImage when
/// a URL is available; falls back to a music-note icon otherwise.
class _FullPlayerArtwork extends ConsumerWidget {
  const _FullPlayerArtwork({
    required this.size,
    this.albumId,
    this.localArtUri,
  });

  /// Edge length in logical pixels — driven by [playerArtworkSize] on wide
  /// layouts, or the narrow layout's fixed size below the split breakpoint.
  final double size;
  final String? albumId;
  // Embedded cover art extracted from a guest-mode local file (file:// URI).
  final Uri? localArtUri;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final artUri = localArtUri;
    if (artUri != null && artUri.scheme == 'file') {
      return Image.file(
        File(artUri.toFilePath()),
        width: size,
        height: size,
        fit: BoxFit.cover,
        errorBuilder: (_, _, _) => const _ArtworkFallback(),
      );
    }
    if (albumId == null || albumId!.isEmpty) {
      return const _ArtworkFallback();
    }
    final artworkAsync = ref.watch(artworkUrlProvider(albumId!));
    return artworkAsync.when(
      data: (url) {
        if (url == null || url.isEmpty) return const _ArtworkFallback();
        return CachedNetworkImage(
          imageUrl: url,
          width: size,
          height: size,
          fit: BoxFit.cover,
          placeholder: (context, _) => const _ArtworkFallback(),
          errorWidget: (context, _, error) => const _ArtworkFallback(),
        );
      },
      loading: () => const _ArtworkFallback(),
      error: (error, _) => const _ArtworkFallback(),
    );
  }
}

class _ArtworkFallback extends StatelessWidget {
  const _ArtworkFallback();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Icon(
        Icons.music_note_rounded,
        size: 80,
        color: context.skinColors.sakuraPink,
      ),
    );
  }
}

/// Lyrics page widget shown in the second page of the FullPlayerScreen
/// PageView.
///
/// No longer carries its own [LyricsBackground]: since v5.26.0 the whole
/// screen sits on one, and nesting a second would run a second set of
/// rotating tiles and a second 64px backdrop blur for no visual gain.
class _LyricsPage extends StatelessWidget {
  const _LyricsPage({required this.trackId, required this.position});

  final String trackId;
  final Duration position;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(16),
      child: _LyricsBody(trackId: trackId, position: position),
    );
  }
}

Widget _emptyLyricsMessage(BuildContext context) => Center(
  child: Text(
    '暂无歌词',
    style: TextStyle(
      fontSize: 15,
      color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.5),
    ),
  ),
);

class _LyricsBody extends ConsumerWidget {
  const _LyricsBody({required this.trackId, required this.position});

  final String trackId;
  final Duration position;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (trackId.isEmpty) {
      return _emptyLyricsMessage(context);
    }
    final isLocal = trackId.startsWith(localTrackIdPrefix);
    final lyricsAsync = isLocal
        ? ref.watch(localLyricsProvider(trackId))
        : ref.watch(lyricsProvider(trackId));
    if (lyricsAsync.isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    final lines = lyricsAsync.valueOrNull;
    if (lines == null || lines.isEmpty) {
      return _emptyLyricsMessage(context);
    }
    final currentIndex = lines.lastIndexWhere((l) => l.timestamp <= position);
    final bilingual = ref.watch(bilingualLyricsProvider);
    return _LyricsList(
      lines: lines,
      currentIndex: currentIndex,
      position: position,
      bilingual: bilingual,
    );
  }
}

class _LyricsList extends StatefulWidget {
  const _LyricsList({
    required this.lines,
    required this.currentIndex,
    required this.position,
    required this.bilingual,
  });
  final List<LyricLine> lines;
  final int currentIndex;
  final Duration position;
  final bool bilingual;

  @override
  State<_LyricsList> createState() => _LyricsListState();
}

class _LyricsListState extends State<_LyricsList> {
  final ScrollController _scrollController = ScrollController();

  double get _itemHeight => widget.bilingual ? 64.0 : 48.0;

  @override
  void didUpdateWidget(_LyricsList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.currentIndex != widget.currentIndex &&
        widget.currentIndex >= 0) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!_scrollController.hasClients) return;
        final itemHeight = _itemHeight;
        final offset =
            (widget.currentIndex * itemHeight -
                    _scrollController.position.viewportDimension / 2 +
                    itemHeight / 2)
                .clamp(
                  _scrollController.position.minScrollExtent,
                  _scrollController.position.maxScrollExtent,
                );
        _scrollController.animateTo(
          offset,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
        );
      });
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
      itemCount: widget.lines.length,
      itemBuilder: (context, i) {
        final isCurrent = i == widget.currentIndex;
        final line = widget.lines[i];
        final showTranslation =
            widget.bilingual &&
            line.translation != null &&
            line.translation!.isNotEmpty;
        return SizedBox(
          height: _itemHeight,
          child: Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                _buildLineText(context, line, isCurrent),
                if (showTranslation)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text(
                      line.translation!,
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: isCurrent ? 13 : 11,
                        color: Theme.of(context).colorScheme.onSurfaceVariant
                            .withValues(alpha: isCurrent ? 0.85 : 0.4),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }

  /// Renders a lyric line, using per-word gradient highlighting for the
  /// current line when word-level timing is available, and falling back to
  /// whole-line highlighting otherwise.
  Widget _buildLineText(BuildContext context, LyricLine line, bool isCurrent) {
    final dimColor = Theme.of(
      context,
    ).colorScheme.onSurface.withValues(alpha: 0.5);
    final activeColor = Theme.of(context).colorScheme.primary;
    final words = line.words;
    if (isCurrent && words != null && words.isNotEmpty) {
      final spans = <TextSpan>[];
      for (var i = 0; i < words.length; i++) {
        final word = words[i];
        final wordEnd = i + 1 < words.length ? words[i + 1].offset : null;
        double progress;
        if (widget.position <= word.offset) {
          progress = 0.0;
        } else if (wordEnd == null) {
          progress = 1.0;
        } else if (widget.position >= wordEnd) {
          progress = 1.0;
        } else {
          final totalMs = (wordEnd - word.offset).inMilliseconds;
          final doneMs = (widget.position - word.offset).inMilliseconds;
          progress = totalMs > 0 ? (doneMs / totalMs).clamp(0.0, 1.0) : 1.0;
        }
        spans.add(
          TextSpan(
            text: word.text,
            style: TextStyle(
              color: Color.lerp(dimColor, activeColor, progress),
            ),
          ),
        );
      }
      return RichText(
        textAlign: TextAlign.center,
        text: TextSpan(
          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
          children: spans,
        ),
      );
    }
    return Text(
      line.text,
      textAlign: TextAlign.center,
      style: TextStyle(
        fontSize: isCurrent ? 18 : 15,
        fontWeight: isCurrent ? FontWeight.w600 : FontWeight.normal,
        color: isCurrent ? activeColor : dimColor,
      ),
    );
  }
}

/// "HQ" chip beside the track-detail heading, shown only for files whose
/// sample rate and bitrate together indicate Hi-Res audio — see
/// [isHiResAudio] for why bitrate stands in for bit depth.
class _HiResBadge extends StatelessWidget {
  const _HiResBadge();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: context.skinColors.sakuraPink,
        borderRadius: BorderRadius.circular(4),
      ),
      child: const Text(
        'HQ',
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w700,
          letterSpacing: 0.5,
          color: Colors.white,
        ),
      ),
    );
  }
}

/// Lyrics or the queue, docked beside the player on wide windows.
///
/// Frosted rather than opaque so the cover's colour field still reads
/// continuously across the whole window — an opaque panel would cut the
/// backdrop in half and undo v5.26.0.
class _PlayerSidePanel extends StatelessWidget {
  const _PlayerSidePanel({
    required this.panel,
    required this.trackId,
    required this.position,
    required this.onClose,
  });

  final _SidePanel panel;
  final String trackId;
  final Duration position;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(0, 8, 16, 16),
      child: GlassPanel(
        padding: const EdgeInsets.fromLTRB(8, 8, 8, 8),
        child: Column(
          children: [
            Row(
              children: [
                const SizedBox(width: 8),
                Text(
                  panel == _SidePanel.lyrics ? '歌词' : '播放队列',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: context.skinColors.onSurface,
                  ),
                ),
                const Spacer(),
                IconButton(
                  icon: Icon(
                    Icons.close,
                    size: 18,
                    color: context.skinColors.onSurfaceVariant,
                  ),
                  tooltip: '关闭',
                  onPressed: onClose,
                ),
              ],
            ),
            Expanded(
              child: switch (panel) {
                _SidePanel.lyrics => _LyricsBody(
                  trackId: trackId,
                  position: position,
                ),
                _SidePanel.queue => const _QueueList(),
                _SidePanel.none => const SizedBox.shrink(),
              },
            ),
          ],
        ),
      ),
    );
  }
}

/// Reorderable play queue, shared by the bottom sheet (narrow windows) and
/// the docked side panel (wide ones) so the two can't drift apart.
class _QueueList extends ConsumerWidget {
  const _QueueList({this.scrollController});

  /// Supplied by the DraggableScrollableSheet so dragging the sheet and
  /// scrolling the list stay one gesture; null when docked.
  final ScrollController? scrollController;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final playerState = ref.watch(playerProvider);
    final queue = playerState.queue;
    final currentIndex = playerState.currentIndex;

    return ReorderableListView.builder(
      scrollController: scrollController,
      itemCount: queue.length,
      onReorderItem: (oldIdx, newIdx) {
        ref.read(playerProvider.notifier).reorderQueue(oldIdx, newIdx);
      },
      itemBuilder: (_, i) {
        final item = queue[i];
        final isCurrent = i == currentIndex;
        return ListTile(
          key: ValueKey(item.id),
          leading: Icon(
            Icons.music_note,
            color: isCurrent
                ? context.skinColors.sakuraPinkLight
                : context.skinColors.onSurfaceVariant,
          ),
          title: Text(
            item.title,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              color: isCurrent
                  ? context.skinColors.sakuraPinkLight
                  : context.skinColors.onSurface,
              fontWeight: isCurrent ? FontWeight.w600 : FontWeight.normal,
            ),
          ),
          subtitle: Text(
            item.artist ?? '',
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(color: context.skinColors.onSurfaceVariant),
          ),
          trailing: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (isCurrent && playerState.isPlaying)
                Icon(
                  Icons.equalizer,
                  color: context.skinColors.sakuraPinkLight,
                  size: 20,
                ),
              Icon(
                Icons.drag_handle,
                color: context.skinColors.onSurfaceVariant,
                size: 20,
              ),
            ],
          ),
          onTap: () => ref
              .read(playerProvider.notifier)
              .playQueue(queue.map((m) => m.id).toList(), initialIndex: i),
        );
      },
    );
  }
}

/// Shrinks every button inside the transport panel when the window is too
/// narrow for them at full size.
///
/// A theme rather than a second arrangement of the same controls: the
/// grouping then stays byte-identical at every width, and there is no
/// alternate layout to keep in sync when a control is added or removed.
class _ControlDensity extends StatelessWidget {
  const _ControlDensity({required this.compact, required this.child});

  final bool compact;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    if (!compact) return child;
    const dense = BoxConstraints(minWidth: 34, minHeight: 34);
    return IconButtonTheme(
      data: IconButtonThemeData(
        style:
            IconButton.styleFrom(
              padding: EdgeInsets.zero,
              minimumSize: const Size(34, 34),
              visualDensity: VisualDensity.compact,
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ).copyWith(
              // styleFrom has no constraints slot, and without it IconButton keeps
              // its 48dp floor no matter what minimumSize says.
              fixedSize: WidgetStatePropertyAll(dense.biggest),
            ),
      ),
      child: TextButtonTheme(
        data: TextButtonThemeData(
          style: TextButton.styleFrom(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            minimumSize: const Size(34, 34),
            visualDensity: VisualDensity.compact,
            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          ),
        ),
        child: child,
      ),
    );
  }
}
