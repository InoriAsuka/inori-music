import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Reorderable play queue.
///
/// Pulled out of `full_player_screen.dart` (where it lived as a private
/// `_QueueList`) and made public in v5.33.0 so a third call site — the
/// desktop shell's docked `QueueDrawer` (`queue_drawer.dart`) — can share it
/// too, rather than growing a second queue UI that would inevitably drift
/// from this one on reorder/delete/jump-to-track behaviour. The two
/// existing call sites (`FullPlayerScreen`'s bottom sheet on narrow windows,
/// its docked side panel on wide ones) are unchanged in shape — only the
/// class's visibility and file moved.
class QueueList extends ConsumerWidget {
  const QueueList({super.key, this.scrollController});

  /// Supplied by a `DraggableScrollableSheet` so dragging the sheet and
  /// scrolling the list stay one gesture; null when docked (the side panel,
  /// the drawer).
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
