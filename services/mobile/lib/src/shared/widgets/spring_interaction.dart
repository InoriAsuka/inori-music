import 'package:flutter/material.dart';
import 'package:flutter/physics.dart';

/// Wraps a control so it lifts slightly under the pointer and springs back
/// when pressed, the way OriginalSound HQ Player's transport buttons do.
///
/// Deliberately *observes* pointer events with a [Listener] rather than
/// handling them with a [GestureDetector]: the child keeps whatever tap
/// handling it already had (these wrap existing [IconButton]s), and this adds
/// only the motion. Nothing here consumes a gesture or joins the arena.
///
/// The damping ratios are the ones read out of OriginalSound's own animation
/// definitions (0.30 for hover, 0.40 for press). Stiffness is not something
/// that source specifies in a directly portable form, so it is tuned here for
/// a ~250ms settle — fast enough to feel like a response rather than an
/// animation.
class SpringInteraction extends StatefulWidget {
  const SpringInteraction({
    super.key,
    required this.child,
    this.hoverLift = 2.0,
    this.pressedScale = 0.90,
  });

  final Widget child;

  /// Logical pixels the child rises by while hovered.
  final double hoverLift;

  /// Scale the child springs down to while held.
  final double pressedScale;

  @override
  State<SpringInteraction> createState() => _SpringInteractionState();
}

class _SpringInteractionState extends State<SpringInteraction>
    with TickerProviderStateMixin {
  // Unbounded controllers: an under-damped spring overshoots its target, and
  // a bounded controller would clamp exactly the overshoot that makes this
  // read as springy rather than as a linear tween.
  late final AnimationController _hover = AnimationController.unbounded(
    vsync: this,
    value: 0,
  );
  late final AnimationController _press = AnimationController.unbounded(
    vsync: this,
    value: 1,
  );

  static final _hoverSpring = SpringDescription.withDampingRatio(
    mass: 1,
    stiffness: 500,
    ratio: 0.30,
  );
  static final _pressSpring = SpringDescription.withDampingRatio(
    mass: 1,
    stiffness: 600,
    ratio: 0.40,
  );

  @override
  void dispose() {
    _hover.dispose();
    _press.dispose();
    super.dispose();
  }

  void _springTo(
    AnimationController controller,
    SpringDescription spring,
    double target,
  ) {
    // Carrying the current velocity across means reversing mid-flight (a
    // quick press-release, or the pointer skimming the edge) continues from
    // where the motion actually was instead of restarting from rest.
    controller.animateWith(
      SpringSimulation(spring, controller.value, target, controller.velocity),
    );
  }

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => _springTo(_hover, _hoverSpring, 1),
      onExit: (_) => _springTo(_hover, _hoverSpring, 0),
      child: Listener(
        // Translucent, not the default deferToChild: otherwise the press
        // animation only fires when the child happens to be hit-testable in
        // its own right, and a wrapper that silently does nothing for some
        // children is a trap. Translucent (rather than opaque) keeps anything
        // painted behind this reachable, since this observes rather than owns
        // the interaction.
        behavior: HitTestBehavior.translucent,
        onPointerDown: (_) =>
            _springTo(_press, _pressSpring, widget.pressedScale),
        onPointerUp: (_) => _springTo(_press, _pressSpring, 1),
        onPointerCancel: (_) => _springTo(_press, _pressSpring, 1),
        child: AnimatedBuilder(
          animation: Listenable.merge([_hover, _press]),
          builder: (context, child) => Transform.translate(
            offset: Offset(0, -widget.hoverLift * _hover.value),
            child: Transform.scale(scale: _press.value, child: child),
          ),
          child: widget.child,
        ),
      ),
    );
  }
}
