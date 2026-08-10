import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'package:inori_music/src/player/full_player_screen.dart';

/// How long the player page's entrance transition takes to slide fully into
/// view — and, via [playerTransitionReverseDuration], fully back out again.
///
/// 320ms: the midpoint of the 300-360ms band this was scoped to. Weighed
/// against two other durations already in this codebase rather than picked
/// blind: comfortably above the 260ms seek-bar position glide
/// (seek_bar_effects.dart's `positionTweenDuration`, a much smaller-scale,
/// continuous motion) and close to the 300ms lyrics-list auto-scroll
/// (full_player_screen.dart's `_LyricsListState`) — both far more modest
/// motions than "the whole window's content changes", which argues for a
/// slightly longer, more deliberate duration than either. This is also a
/// *frequent* interaction — opened many times a session, unlike e.g. a
/// settings modal — which argues against going any higher in the band.
const playerTransitionDuration = Duration(milliseconds: 320);

/// Same duration as [playerTransitionDuration] for the reverse leg.
/// "Symmetric" (the field report's own word for the entrance/exit curves)
/// describes the curve *shape* — decelerating open, accelerating close, see
/// [_PlayerTransitionShellState._openness]'s doc comment for the derivation
/// — not the timing. Keeping the duration identical means direction is the
/// only thing that changes between opening and closing.
const playerTransitionReverseDuration = playerTransitionDuration;

/// How far (as a fraction of screen height) a drag must travel before
/// release commits to dismissing rather than bouncing back — independent of
/// velocity, see [_flingVelocityThreshold] for the other half of that
/// decision.
///
/// 0.3: below halfway, since forcing a slow, deliberate drag past the
/// midpoint before it commits reads as unresponsive; high enough that a
/// short, accidental drag (e.g. a slightly off-axis attempt to interact with
/// something else near the top bar) reliably bounces back rather than
/// closing the screen out from under the user.
const _dismissExtentThreshold = 0.3;

/// Downward release velocity, in logical pixels per second, that commits to
/// dismissing regardless of how far the drag actually travelled — letting a
/// quick flick close the screen even from just past the gesture's own touch
/// slop.
///
/// 700.0 is not picked in isolation: it is the exact constant Flutter's own
/// [Dismissible] widget uses for the same decision (`_kMinFlingVelocity` in
/// the SDK's `dismissible.dart`), reused rather than re-derived so this
/// gesture's "does that count as a flick" threshold matches the one Flutter
/// already ships elsewhere in this app.
const _flingVelocityThreshold = 700.0;

/// How dark the scrim behind the sliding player gets once fully open —
/// [Colors.black54]'s own alpha (0x8A / 0xFF ≈ 0.54) is the closest existing
/// precedent in the Flutter SDK for "a modal barrier's resting darkness"
/// ([ModalRoute]'s own default barrier colour uses it), rounded to 0.55.
const _scrimMaxAlpha = 0.55;

/// Builds the player route's page transition: the whole screen rises from
/// the bottom to cover the window — Apple Music / EchoMusic's shape,
/// requested directly against this app's inherited Material default
/// ([FadeUpwardsPageTransitionsBuilder], which only offsets 25% of the
/// screen height plus a fade — cheap-looking for a page this central to the
/// app) — with a darkening scrim behind it and a drag-to-dismiss gesture on
/// top (see [FullPlayerScreen.transition]/[_dragHandle]).
///
/// Ignores the `child` `CustomTransitionPage` hands this and constructs its
/// own [FullPlayerScreen] wired to `animation` instead:
/// `CustomTransitionPage.child` is built by go_router's `pageBuilder` before
/// `animation` even exists, so there is no way for the page itself to carry
/// [PlayerTransition] without this indirection — see router.dart's own
/// comment at the call site for the full picture.
Widget playerPageTransitionsBuilder(
  BuildContext context,
  Animation<double> animation,
  Animation<double> secondaryAnimation,
  Widget child,
) {
  return _PlayerTransitionShell(animation: animation);
}

/// Owns the single [AnimationController] that actually drives this route's
/// visible open/close motion once a drag is involved, and reconciles it with
/// the route's own `animation` — see [_PlayerTransitionShellState._openness].
///
/// A second, local controller exists at all — rather than the drag gesture
/// directly manipulating `ModalRoute.of(context)!.controller` the way
/// Flutter's own Cupertino edge-swipe-back gesture does internally — because
/// that controller is `@protected` on `TransitionRoute`. Legitimate for a
/// mixin *on* the route class itself (`CupertinoRouteTransitionMixin` is
/// literally mixed into the route), illegitimate for a plain descendant
/// widget like this one: touching it here would trip the analyzer's
/// `invalid_use_of_protected_member`, and this project's `flutter analyze`
/// gate treats any new warning as a failure. A second controller sidesteps
/// that restriction by construction rather than working around it.
class _PlayerTransitionShell extends StatefulWidget {
  const _PlayerTransitionShell({required this.animation});

  final Animation<double> animation;

  @override
  State<_PlayerTransitionShell> createState() => _PlayerTransitionShellState();
}

class _PlayerTransitionShellState extends State<_PlayerTransitionShell>
    with SingleTickerProviderStateMixin {
  /// "How open" a drag-driven override says this screen is — `1` at rest,
  /// matching [widget.animation]'s own fully-open value of `1`; ramps toward
  /// `0` as the user drags down (`0` == fully dismissed, off-screen); and on
  /// release either animates back to `1` (cancelled) or on to `0` (committed
  /// — see [_dismiss]). Never leaves `1` except in response to a drag, which
  /// is what lets [_openness] combine the two sources with a plain `min`
  /// rather than an explicit "is a drag in progress" flag.
  late final AnimationController _drag;

  /// Screen height captured at drag-start, so a mid-drag resize (never
  /// happens in practice, but the arithmetic below assumes a fixed
  /// denominator) can't retroactively change what a given finger offset
  /// means partway through a single gesture.
  double _dragReferenceHeight = 0;

  @override
  void initState() {
    super.initState();
    _drag = AnimationController(vsync: this, value: 1.0);
  }

  @override
  void dispose() {
    _drag.dispose();
    super.dispose();
  }

  void _handleDragStart() {
    _dragReferenceHeight = MediaQuery.sizeOf(context).height;
    // A still-running cancel/dismiss settle animation from a previous drag
    // must stop driving _drag.value before this new gesture starts setting
    // it directly in _handleDragUpdate — otherwise the two fight each frame.
    _drag.stop();
  }

  /// [deltaY] tracks the pointer 1:1 (no easing) — "拖动跟手" in the field
  /// report's own words. Curves only ever apply to the two *settle*
  /// animations below, once the finger has actually lifted.
  void _handleDragUpdate(double deltaY) {
    if (_dragReferenceHeight <= 0) return;
    _drag.value = (_drag.value - deltaY / _dragReferenceHeight).clamp(0.0, 1.0);
  }

  void _handleDragEnd(double velocityY) {
    if (_dragReferenceHeight <= 0) return;
    final draggedPastThreshold = _drag.value <= 1.0 - _dismissExtentThreshold;
    final flungDown = velocityY > _flingVelocityThreshold;
    if (draggedPastThreshold || flungDown) {
      _dismiss();
    } else {
      _cancel();
    }
  }

  /// Commits to closing: finishes the local drag motion the rest of the way
  /// to `0` — the same accelerating curve the route's own programmatic close
  /// uses (see [_openness]) — and only *then* pops the real route.
  ///
  /// Popping after the local animation finishes, rather than immediately, is
  /// what keeps this glitch-free: `Navigator.maybePop()` starts the route's
  /// own independent reverse transition (`playerTransitionReverseDuration`,
  /// driven by [widget.animation]), which — if popped mid-drag while
  /// [widget.animation] was still sitting at its resting `1.0` — would
  /// otherwise start its own slide from a *different* point than wherever
  /// the finger left off, a visible jump. Waiting sidesteps that instead of
  /// choreographing it: by the time the pop's own reverse transition would
  /// render anything, [_openness]'s `min()` is already pinned at `0` from
  /// this animation having finished, so that second transition is never
  /// actually visible at all, however long it takes underneath.
  void _dismiss() {
    _drag
        .animateTo(
          0.0,
          duration: playerTransitionReverseDuration,
          curve: Curves.easeOutCubic,
        )
        .whenComplete(() {
          if (mounted) Navigator.of(context).maybePop();
        });
  }

  /// Bounces back to fully open — same duration/curve as the entrance
  /// transition itself, since a cancelled drag is, motion-wise, just a
  /// second entrance.
  void _cancel() {
    _drag.animateTo(
      1.0,
      duration: playerTransitionDuration,
      curve: Curves.easeOutCubic,
    );
  }

  /// Combines the route's own entrance/exit animation with any local drag
  /// override into the single value everything in [build] renders from.
  ///
  /// [Curves.easeOutCubic] is applied directly to [widget.animation]'s raw
  /// value rather than through a [CurvedAnimation] with a distinct
  /// `reverseCurve` — deliberately, not an oversight. Flutter's own
  /// `CurvedAnimation.value` implementation applies whichever curve is
  /// active to the parent's raw value regardless of direction (see
  /// `animations.dart`), so leaving `reverseCurve` unset already means "use
  /// `curve` for both directions", and doing the same transform by hand here
  /// avoids the extra object. What actually matters is what that produces:
  /// with `openness` read straight off this curved value (as it is, once no
  /// drag is overriding it), the *entrance* (animation rising 0->1 over
  /// time) reads `easeOutCubic`'s own fast-start/slow-end shape directly —
  /// decelerating, exactly the field report's ask. The *exit* (animation
  /// falling 1->0 over time, e.g. from the close button rather than a drag)
  /// reads the *mirror image* of that same shape: substituting u = 1 - t
  /// into `easeOutCubic(t)` and differentiating shows the resulting motion
  /// is slow at the start of the close and fast at the end — accelerating,
  /// the "对称" (symmetric — same curve, opposite end emphasised) shape the
  /// field report asked for, for free, with no second named curve to keep in
  /// sync with the first.
  double get _openness {
    final routeOpenness = Curves.easeOutCubic.transform(widget.animation.value);
    // Whichever is *more closed* wins. _drag never leaves 1.0 except in
    // response to a drag (see its own doc comment), so before one starts
    // this is always exactly routeOpenness, and once one starts it can only
    // ever pull the rendered value down — never fights the route's own
    // motion, just overrides it when it disagrees.
    return math.min(routeOpenness, _drag.value);
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: Listenable.merge([widget.animation, _drag]),
      builder: (context, _) {
        final openness = _openness;
        final screenHeight = MediaQuery.sizeOf(context).height;
        return Stack(
          children: [
            // Darkens whatever sits behind this route (visible in the gap
            // this screen's own translation opens up — the router wires this
            // page as `opaque: false` specifically so that stays true) in
            // step with how open this screen currently is, so the previous
            // screen never looks starkly bare mid-slide. Blocks interaction
            // with that revealed screen for as long as this one is at all
            // open, matching how a modal barrier normally behaves.
            IgnorePointer(
              ignoring: openness <= 0,
              child: Opacity(
                opacity: (_scrimMaxAlpha * (1 - openness)).clamp(0.0, 1.0),
                child: const ColoredBox(color: Colors.black),
              ),
            ),
            Transform.translate(
              offset: Offset(0, (1 - openness) * screenHeight),
              child: FullPlayerScreen(
                transition: PlayerTransition(
                  progress: widget.animation,
                  onDragStart: _handleDragStart,
                  onDragUpdate: _handleDragUpdate,
                  onDragEnd: _handleDragEnd,
                ),
              ),
            ),
          ],
        );
      },
    );
  }
}
