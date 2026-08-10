import 'package:flutter/material.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/desktop_app_bar.dart';

/// "为您推荐" — the sidebar's own "发现音乐" group carries this destination
/// because the EchoMusic skeleton it's modelled on does (see
/// shell_scaffold.dart's `_discoverItems` doc comment), but this app has no
/// recommendation engine behind it yet: no listening-history-driven ranking,
/// no taste model, nothing server-side to call. v5.33.0's brief was explicit
/// that the *skeleton* should still exist even where the backing capability
/// doesn't ("摆出完整骨架，缺的做空状态") — so this is a real, reachable screen
/// with a real destination-shaped guiding message, not a dead nav item, and
/// not a fabricated feed built to look occupied.
///
/// Contrast with `explore_screen.dart`'s sibling destination, which *does*
/// have a real server endpoint behind it (recently-added albums) — the
/// difference between the two is exactly "we have data for this" vs. "we
/// don't", not a stylistic choice between them.
class ForYouScreen extends StatelessWidget {
  const ForYouScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    return Scaffold(
      appBar: DesktopAppBar(title: Text(t.forYou)),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.auto_awesome_outlined,
                size: 56,
                color: context.skinColors.onSurfaceVariant,
              ),
              const SizedBox(height: 16),
              Text(
                t.forYou,
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w600,
                  color: context.skinColors.onSurface,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                t.forYouComingSoon,
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  color: context.skinColors.onSurfaceVariant,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
