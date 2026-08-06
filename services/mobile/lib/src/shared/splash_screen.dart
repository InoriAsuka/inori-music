import 'package:flutter/material.dart';

import 'package:inori_music/src/shared/theme/sakura_dusk.dart';
import 'package:inori_music/src/shared/widgets/app_background.dart';
import 'package:inori_music/src/shared/widgets/inori_mark.dart';

/// Shown while auth status resolves (a single `/me` round-trip, typically
/// well under a second). No interactive controls — this is a transient
/// gate, not a screen a user should ever need to act on.
class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _scale;
  late final Animation<double> _opacity;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 260),
    );
    _scale = Tween(begin: 0.9, end: 1.0)
        .animate(CurvedAnimation(parent: _controller, curve: Curves.easeOutCubic));
    _opacity = Tween(begin: 0.0, end: 1.0)
        .animate(CurvedAnimation(parent: _controller, curve: Curves.easeOutCubic));
    _controller.forward();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: AppBackground(
        child: Center(
          child: FadeTransition(
            opacity: _opacity,
            child: ScaleTransition(
              scale: _scale,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const InoriMark(size: 88),
                  const SizedBox(height: 20),
                  Text(
                    'Inori Music',
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                          color: SakuraDuskColors.onBackground,
                          fontWeight: FontWeight.w700,
                        ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
