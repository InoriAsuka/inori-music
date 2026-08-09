import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as ps;
import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Cycles [ps.PlayerState.repeat] through off -> all -> one -> off, tinting
/// the icon with the accent colour whenever repeat is active in any form.
///
/// Pulled out of `full_player_screen.dart` in v5.30.5 so the desktop mini
/// player bar could offer the same control (see `mini_player_bar.dart`)
/// without re-deriving the cycle order by hand — the field report explicitly
/// asked for the mini bar's repeat/shuffle to reuse the full player's
/// existing semantics rather than grow a second implementation that could
/// drift from it. `iconSize` defaults to [Icon]'s own default (24) to match
/// the full player's un-sized repeat icon; the mini bar passes its own
/// EchoMusic-scale 22px explicitly.
class RepeatModeButton extends ConsumerWidget {
  const RepeatModeButton({super.key, this.iconSize});

  final double? iconSize;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final repeat = ref.watch(playerProvider.select((s) => s.repeat));
    return IconButton(
      icon: Icon(
        Icons.repeat,
        size: iconSize,
        color: repeat != ps.RepeatMode.none
            ? context.skinColors.sakuraPinkLight
            : context.skinColors.onSurfaceVariant,
      ),
      tooltip: 'Repeat: ${repeat.name}',
      onPressed: () {
        final notifier = ref.read(playerProvider.notifier);
        switch (repeat) {
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
    );
  }
}

/// Toggles [ps.PlayerState.shuffle], tinting the icon the same way
/// [RepeatModeButton] does. See that class's doc comment for why this is
/// its own shared widget rather than inline code duplicated per surface.
class ShuffleButton extends ConsumerWidget {
  const ShuffleButton({super.key, this.iconSize});

  final double? iconSize;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isShuffle = ref.watch(playerProvider.select((s) => s.shuffle));
    return IconButton(
      icon: Icon(
        Icons.shuffle,
        size: iconSize,
        color: isShuffle
            ? context.skinColors.sakuraPinkLight
            : context.skinColors.onSurfaceVariant,
      ),
      tooltip: 'Shuffle',
      onPressed: () => ref.read(playerProvider.notifier).setShuffle(!isShuffle),
    );
  }
}
