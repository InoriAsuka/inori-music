// ignore_for_file: unnecessary_non_null_assertion
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/audio/sleep_timer_notifier.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/spring_interaction.dart';

/// Persistent mini-player bar displayed at the bottom of the shell scaffold.
///
/// A floating rounded card rather than a full-bleed strip — EchoMusic and
/// Spotube both keep the player bar visually separate from the nav chrome
/// beneath it (a `SurfaceCard`/elevated rounded panel with margins), not a
/// flush edge-to-edge bar. The margin here is what makes the drop shadow
/// from `elevation` actually visible on both sides instead of running off
/// the screen edges.
///
/// Scaled to EchoMusic's `PlayerBar.vue` proportions since v5.30.0 rather
/// than Apple Music's — the user explicitly flagged Apple Music's bar as too
/// small once the player page itself had been reworked to match Apple Music
/// elsewhere. The content row (everything below the slim seek strip) is a
/// fixed 84px (EchoMusic's `h-21`), the artwork is 56px, and the layout is
/// three sections — song info / transport / actions — with the two info/
/// actions sections sharing equal flex so the transport trio in between
/// always lands dead centre, the same centring trick
/// `full_player_screen.dart`'s own transport row uses.
///
/// The slim seek strip predates this rework and stays outside the 84px
/// budget rather than folding into it (EchoMusic has no equivalent — its
/// progress bar lives inside the middle column instead): overlaying it on
/// the content row would need it to win hit-testing priority over whatever
/// sits underneath, which is a UX trade-off this phase didn't set out to make.
class MiniPlayerBar extends ConsumerWidget {
  const MiniPlayerBar({super.key});

  /// EchoMusic's `PlayerBar.vue` footer height (`h-21`).
  static const _contentHeight = 84.0;

  /// EchoMusic's cover size within that footer.
  static const _artworkSize = 56.0;

  /// EchoMusic's transport icon size (`width="22" height="22"` on
  /// prev/next/play-mode in `PlayerBar.vue`). The play/pause glyph itself
  /// stays a little larger, matching the existing (pre-v5.30.0) proportion
  /// between it and its siblings.
  static const _transportIconSize = 22.0;

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

    return Padding(
      padding: const EdgeInsets.fromLTRB(8, 0, 8, 8),
      child: Material(
        color: context.skinColors.playerBar,
        elevation: 8,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        clipBehavior: Clip.antiAlias,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const _MiniPlayerProgressBar(),
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
                    child: Row(
                      children: [
                        // Section 1 — artwork + title/artist. Flexible, so it
                        // absorbs whatever width the fixed-size transport
                        // trio and section 3 don't need.
                        Expanded(
                          child: Row(
                            children: [
                              _MiniPlayerArtwork(
                                albumId:
                                    mediaItem?.extras?['albumId'] as String?,
                                size: _artworkSize,
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
                                          color: context
                                              .skinColors
                                              .onSurfaceVariant,
                                          fontSize: 11,
                                        ),
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),

                        // Section 2 — the transport trio. Fixed-size, kept
                        // dead centre by section 1 and section 3 sharing
                        // equal Expanded flex around it rather than either
                        // side growing to whatever it happens to contain.
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            // The transport trio is the highest-traffic
                            // control surface in the app, so it's where the
                            // spring hover/press motion earns its keep —
                            // SpringInteraction only observes pointer
                            // events, leaving each button's own tap handling
                            // untouched.
                            SpringInteraction(
                              child: IconButton(
                                icon: const Icon(
                                  Icons.skip_previous,
                                  size: _transportIconSize,
                                ),
                                color: context.skinColors.onSurfaceVariant,
                                onPressed: () => ref
                                    .read(playerProvider.notifier)
                                    .previous(),
                                tooltip: 'Previous',
                              ),
                            ),
                            SpringInteraction(
                              child: IconButton(
                                icon: Icon(
                                  isPlaying
                                      ? Icons.pause_rounded
                                      : Icons.play_arrow_rounded,
                                  size: 28,
                                  color: context.skinColors.onBackground,
                                ),
                                tooltip: isPlaying ? 'Pause' : 'Play',
                                onPressed: isBuffering
                                    ? null
                                    : () => ref
                                          .read(playerProvider.notifier)
                                          .togglePlayPause(),
                              ),
                            ),
                            SpringInteraction(
                              child: IconButton(
                                icon: const Icon(
                                  Icons.skip_next,
                                  size: _transportIconSize,
                                ),
                                color: context.skinColors.onSurfaceVariant,
                                onPressed: () =>
                                    ref.read(playerProvider.notifier).next(),
                                tooltip: 'Next',
                              ),
                            ),
                          ],
                        ),

                        // Section 3 — actions. Mirrors section 1's flex so
                        // section 2 actually sits in the middle instead of
                        // drifting toward whichever side has less in it.
                        Expanded(
                          child: Align(
                            alignment: Alignment.centerRight,
                            child: Consumer(
                              builder: (context, ref, _) {
                                final timerState = ref.watch(
                                  sleepTimerProvider,
                                );
                                final active = timerState.active;
                                final remaining = timerState.remaining;
                                final label = timerState.stopAfterTrack
                                    ? '♪'
                                    : (remaining != null
                                          ? _formatRemaining(remaining)
                                          : null);
                                return active
                                    ? TextButton.icon(
                                        icon: Icon(
                                          Icons.alarm,
                                          size: 18,
                                          color: Theme.of(
                                            context,
                                          ).colorScheme.primary,
                                        ),
                                        label: Text(
                                          label ?? '',
                                          style: TextStyle(
                                            fontSize: 10,
                                            color: Theme.of(
                                              context,
                                            ).colorScheme.primary,
                                          ),
                                        ),
                                        onPressed: () =>
                                            _showSleepTimerSheet(context, ref),
                                      )
                                    : IconButton(
                                        icon: const Icon(Icons.alarm, size: 22),
                                        color:
                                            context.skinColors.onSurfaceVariant,
                                        tooltip: 'Sleep timer',
                                        onPressed: () =>
                                            _showSleepTimerSheet(context, ref),
                                      );
                              },
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ],
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

/// Compact draggable progress bar shown above the mini player controls.
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

/// Mini player artwork thumbnail — shows the album cover or a fallback icon.
///
/// [size] defaults to the pre-v5.30.0 44px box size (icon size and corner
/// radius scale proportionately from it rather than staying fixed, so a
/// future caller passing something other than 56 still gets sane
/// proportions); [MiniPlayerBar] itself always passes 56, EchoMusic's
/// `PlayerBar.vue` cover size.
class _MiniPlayerArtwork extends ConsumerWidget {
  const _MiniPlayerArtwork({this.albumId, this.size = 44});

  final String? albumId;
  final double size;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final artworkAsync = albumId != null && albumId!.isNotEmpty
        ? ref.watch(artworkUrlProvider(albumId!))
        : null;

    // EchoMusic's own fallback icon is 24px inside a 56px cover — scaling
    // from that ratio rather than hardcoding 24 keeps the icon proportionate
    // if a future call site passes a different size.
    final iconSize = size * (24 / 56);

    Widget child;
    if (artworkAsync == null) {
      child = Icon(
        Icons.music_note,
        color: context.skinColors.onSurfaceVariant,
        size: iconSize,
      );
    } else {
      child = artworkAsync.when(
        data: (url) {
          if (url == null || url.isEmpty) {
            return Icon(
              Icons.music_note,
              color: context.skinColors.onSurfaceVariant,
              size: iconSize,
            );
          }
          return CachedNetworkImage(
            imageUrl: url,
            width: size,
            height: size,
            fit: BoxFit.cover,
            placeholder: (context, _) => Icon(
              Icons.music_note,
              color: context.skinColors.onSurfaceVariant,
              size: iconSize,
            ),
            errorWidget: (context, _, error) => Icon(
              Icons.music_note,
              color: context.skinColors.onSurfaceVariant,
              size: iconSize,
            ),
          );
        },
        loading: () => Icon(
          Icons.music_note,
          color: context.skinColors.onSurfaceVariant,
          size: iconSize,
        ),
        error: (error, _) => Icon(
          Icons.music_note,
          color: context.skinColors.onSurfaceVariant,
          size: iconSize,
        ),
      );
    }

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
      child: child,
    );
  }
}
