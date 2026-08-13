import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/shared/theme/sakura_dusk.dart';
import 'package:inori_music/src/shared/widgets/app_background.dart';
import 'package:inori_music/src/shared/widgets/gate_window_chrome.dart';
import 'package:inori_music/src/shared/widgets/inori_mark.dart';

/// How long auth may stay unresolved before this screen stops pretending
/// everything is fine. Long enough to clear a slow `/me` round-trip (dio
/// allows 15s to connect), short enough that nobody sits here wondering.
const _slowThreshold = Duration(seconds: 8);

/// Shown while auth status resolves — normally a single `/me` round-trip,
/// well under a second.
///
/// This used to be a pure splash: a 260ms fade-in and then a completely
/// static image, on the stated assumption that it is "a transient gate, not a
/// screen a user should ever need to act on". v5.37.0 disproved that. A login
/// that threw anything other than a `DioException` left auth pinned at
/// `AsyncLoading`; the router holds the UI here for as long as that lasts;
/// and the result was a motionless logo with no spinner, no text and no way
/// out — reported, reasonably, as "卡死".
///
/// The strand itself is fixed at the source ([AuthNotifier.login] is now
/// total, and the router no longer pulls anyone off the login form). But a
/// gate with no exit is the wrong shape regardless of whether today's code
/// can still reach it, so this screen now always shows that it is alive, and
/// always offers a way back once it has waited too long.
class SplashScreen extends ConsumerStatefulWidget {
  const SplashScreen({super.key});

  @override
  ConsumerState<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends ConsumerState<SplashScreen>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _scale;
  late final Animation<double> _opacity;

  Timer? _slowTimer;
  bool _slow = false;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 260),
    );
    _scale = Tween(
      begin: 0.9,
      end: 1.0,
    ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeOutCubic));
    _opacity = Tween(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeOutCubic));
    _controller.forward();

    _slowTimer = Timer(_slowThreshold, () {
      if (mounted) setState(() => _slow = true);
    });
  }

  @override
  void dispose() {
    _slowTimer?.cancel();
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    return Scaffold(
      body: Stack(
        children: [
          AppBackground(
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
                        style: Theme.of(context).textTheme.headlineMedium
                            ?.copyWith(
                              color: SakuraDuskColors.onBackground,
                              fontWeight: FontWeight.w700,
                            ),
                      ),
                      const SizedBox(height: 28),

                      // Always present, always animating. Its only job is to
                      // be the difference between "working" and "hung" — the
                      // question this screen previously gave the user no way
                      // to answer.
                      const SizedBox(
                        height: 22,
                        width: 22,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: SakuraDuskColors.sakuraPink,
                        ),
                      ),
                      const SizedBox(height: 16),

                      ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 320),
                        child: Text(
                          _slow ? t.splashTakingLonger : t.splashConnecting,
                          textAlign: TextAlign.center,
                          style: Theme.of(context).textTheme.bodySmall
                              ?.copyWith(
                                color: SakuraDuskColors.onSurfaceVariant,
                              ),
                        ),
                      ),

                      if (_slow) ...[
                        const SizedBox(height: 8),
                        TextButton(
                          onPressed: () => ref
                              .read(authProvider.notifier)
                              .abandonPendingAuth(t.splashTakingLonger),
                          child: Text(t.splashBackToSignIn),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ),
          ),
          const Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: GateWindowChrome(),
          ),
        ],
      ),
    );
  }
}
