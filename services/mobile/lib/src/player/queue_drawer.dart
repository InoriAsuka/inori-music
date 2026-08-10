import 'package:flutter/material.dart';
import 'package:flutter/services.dart'
    show KeyDownEvent, KeyEvent, LogicalKeyboardKey;
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/player_transition.dart';
import 'package:inori_music/src/player/queue_drawer_provider.dart';
import 'package:inori_music/src/player/queue_list.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/glass_panel.dart';

/// The desktop shell's docked queue panel — Spotify/EchoMusic's "slide a
/// drawer in from the right, stay on the current page" pattern, which is
/// what the mini bar's queue button always should have opened (see
/// `mini_player_bar.dart`'s `_openQueue` doc comment: v5.30.5 explicitly
/// deferred building this exact thing).
///
/// Always mounted (by `_DesktopLayout`, shell_scaffold.dart) rather than
/// conditionally inserted/removed from the tree — an [AnimationController]
/// needs to survive across open/close to actually animate the transition,
/// and [IgnorePointer] below is what keeps it from intercepting clicks
/// meant for the content underneath while closed.
///
/// Content is [QueueList] — the same class the full player screen's own
/// docked side panel and bottom sheet already share (v5.33.0 pulled it out
/// of `full_player_screen.dart` into its own file for exactly this reason:
/// a third caller with its own reorder/delete/jump-to-track implementation
/// would only ever drift from the other two).
class QueueDrawer extends ConsumerStatefulWidget {
  const QueueDrawer({super.key});

  /// Fixed panel width — comfortably wide enough for a track title +
  /// artist without the ellipsis firing on every other row, while still
  /// reading as a docked accessory rather than a second content column
  /// competing with whatever page is open underneath it.
  static const width = 360.0;

  @override
  ConsumerState<QueueDrawer> createState() => _QueueDrawerState();
}

class _QueueDrawerState extends ConsumerState<QueueDrawer>
    with SingleTickerProviderStateMixin {
  // Built in initState, not `late final` with lazy assignment — this
  // codebase has been bitten by deferred AnimationController construction
  // before (see player_transition.dart's own controller for the established
  // convention this follows).
  late final AnimationController _controller;
  final _focusNode = FocusNode(debugLabel: 'queue-drawer');

  @override
  void initState() {
    super.initState();
    // Reuses playerTransitionDuration/Curves.easeOutCubic rather than a
    // second, independently-tuned constant — the field report was explicit
    // that this drawer's motion should read as the same visual vocabulary
    // as the player page's own slide (v5.32.0), not a bespoke animation
    // invented just for this panel.
    _controller = AnimationController(
      vsync: this,
      duration: playerTransitionDuration,
      reverseDuration: playerTransitionReverseDuration,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _close() => ref.read(queueDrawerOpenProvider.notifier).close();

  KeyEventResult _handleKey(FocusNode node, KeyEvent event) {
    if (event is KeyDownEvent &&
        event.logicalKey == LogicalKeyboardKey.escape) {
      _close();
      return KeyEventResult.handled;
    }
    return KeyEventResult.ignored;
  }

  @override
  Widget build(BuildContext context) {
    final open = ref.watch(queueDrawerOpenProvider);
    // ref.listen (not watch) drives the controller and focus as *side
    // effects* of the open/close flag flipping — the animated value itself
    // (read below via _controller, an Animation this widget owns) is what
    // actually drives the visuals frame to frame, so build() doesn't need
    // to run again for every tick.
    ref.listen<bool>(queueDrawerOpenProvider, (previous, next) {
      if (next) {
        _controller.animateTo(1.0, curve: Curves.easeOutCubic);
        _focusNode.requestFocus();
      } else {
        _controller.animateTo(0.0, curve: Curves.easeOutCubic);
      }
    });

    return AnimatedBuilder(
      animation: _controller,
      builder: (context, _) {
        final t = _controller.value;
        // Closed AND settled (not merely "the provider says closed" —
        // mid-reverse-animation the drawer is still visibly sliding off and
        // must keep intercepting the barrier's own taps, and mid-*opening*
        // animation the controller can still read 0.0 on its very first
        // frame) is the only state that should let clicks fall through to
        // the content underneath, and — see below — the only state that
        // should tear the panel's own content back down.
        final interactive = open || t > 0;
        return IgnorePointer(
          ignoring: !interactive,
          child: Stack(
            children: [
              // Barrier — invisible, not a darkening scrim: this drawer
              // overlays the content column, it does not modally block it
              // the way the full player's own drag-to-dismiss scrim does
              // (Spotify/EchoMusic's own queue panel does not dim the page
              // behind it either). Its only job is "a click outside the
              // panel itself closes the drawer".
              Positioned.fill(
                child: IgnorePointer(
                  ignoring: t <= 0,
                  child: GestureDetector(
                    behavior: HitTestBehavior.opaque,
                    onTap: _close,
                  ),
                ),
              ),
              Positioned(
                top: 0,
                right: 0,
                bottom: 0,
                width: QueueDrawer.width,
                child: FractionalTranslation(
                  // t=0 (closed): shifted one full panel-width to the right,
                  // off-screen. t=1 (open): no shift. FractionalTranslation
                  // scales by the child's *own* size, so this stays correct
                  // if QueueDrawer.width is ever tuned without touching this
                  // math.
                  translation: Offset(1 - t, 0),
                  // `interactive`, not unconditionally built: this widget is
                  // permanently mounted (see the class doc comment on why
                  // the AnimationController needs to outlive a single
                  // open/close cycle), but QueueList underneath does real
                  // work — it watches playerProvider and renders one
                  // ListTile per queued track — and leaving it *built* at
                  // rest, merely translated off-screen, kept every one of
                  // those track titles in the tree and findable by
                  // find.text/find.byType even while invisible. That
                  // silently duplicated whatever the mini player bar itself
                  // was already showing for the current track in every test
                  // that renders the desktop shell, not just ones about this
                  // drawer. Swapping in a bare SizedBox.shrink() once fully
                  // closed removes QueueList (and its provider subscription)
                  // from the tree entirely — cheap, and the FractionalTranslation/
                  // Positioned above still reserve the same 360px slot either
                  // way since Positioned(width: ...) gives a tight
                  // constraint regardless of what the child inside asks for.
                  child: interactive
                      ? Focus(
                          focusNode: _focusNode,
                          onKeyEvent: _handleKey,
                          child: Padding(
                            padding: const EdgeInsets.fromLTRB(0, 8, 8, 8),
                            child: GlassPanel(
                              padding: const EdgeInsets.fromLTRB(8, 8, 8, 8),
                              child: Column(
                                children: [
                                  Row(
                                    children: [
                                      const SizedBox(width: 8),
                                      Text(
                                        AppLocalizations.of(context).queue,
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
                                          color: context
                                              .skinColors
                                              .onSurfaceVariant,
                                        ),
                                        tooltip: 'Close',
                                        onPressed: _close,
                                      ),
                                    ],
                                  ),
                                  const Expanded(child: QueueList()),
                                ],
                              ),
                            ),
                          ),
                        )
                      : const SizedBox.shrink(),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
