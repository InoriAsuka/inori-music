import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/src/local_library/local_library_db.dart';
import 'package:inori_music/src/local_library/local_library_notifier.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/desktop_app_bar.dart';

/// Guest mode's home screen: a flat list of locally-imported audio files.
/// No account, no server — this is what makes the app usable without login.
/// v1 is intentionally a single flat list (artist/album/title sorted) rather
/// than a full Artists→Albums→Tracks hierarchy: a personal local file
/// collection is typically far smaller than a server catalog, so grouped
/// browsing is a later candidate, not core scope.
class LocalLibraryScreen extends ConsumerWidget {
  const LocalLibraryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(localLibraryProvider);
    final playerState = ref.watch(playerProvider);

    return Scaffold(
      appBar: DesktopAppBar(
        title: const Text('本地曲库'),
        actions: [
          PopupMenuButton<_ImportAction>(
            icon: const Icon(Icons.add),
            tooltip: '导入',
            onSelected: (action) {
              final notifier = ref.read(localLibraryProvider.notifier);
              switch (action) {
                case _ImportAction.files:
                  notifier.importFiles();
                case _ImportAction.folder:
                  notifier.importFolder();
              }
            },
            itemBuilder: (context) => const [
              PopupMenuItem(value: _ImportAction.files, child: Text('导入文件')),
              PopupMenuItem(value: _ImportAction.folder, child: Text('导入文件夹')),
            ],
          ),
          IconButton(
            icon: const Icon(Icons.sort),
            tooltip: '排序',
            onPressed: () => _showSortSheet(context, ref),
          ),
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            tooltip: '设置',
            onPressed: () => context.push(AppRoutes.settings),
          ),
        ],
      ),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Text('$e', style: TextStyle(color: context.skinColors.error)),
        ),
        data: (tracks) {
          if (tracks.isEmpty) return const _EmptyLocalLibrary();
          final ids = tracks.map((t) => t.id).toList();
          return ListView.builder(
            itemCount: tracks.length,
            itemBuilder: (context, i) {
              final track = tracks[i];
              final isCurrent = playerState.mediaItem?.id == track.id;
              return _LocalTrackTile(
                key: ValueKey(track.id),
                track: track,
                isCurrent: isCurrent,
                isPlaying: isCurrent && playerState.isPlaying,
                onTap: () => ref
                    .read(playerProvider.notifier)
                    .playQueue(ids, initialIndex: i),
                onDelete: () => _confirmDelete(context, ref, track),
              );
            },
          );
        },
      ),
    );
  }

  Future<void> _confirmDelete(
    BuildContext context,
    WidgetRef ref,
    LocalLibraryTrack track,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('移除曲目'),
        content: Text('从本地曲库中移除「${track.title}」？不会删除原始文件。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('移除'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(localLibraryProvider.notifier).remove(track.id);
    }
  }

  static String _sortLabel(LocalLibrarySortOrder sort) => switch (sort) {
    LocalLibrarySortOrder.artistAlbumTitle => '艺术家 / 专辑 / 标题',
    LocalLibrarySortOrder.recentlyAdded => '最近导入',
    LocalLibrarySortOrder.titleAZ => '标题 A-Z',
    LocalLibrarySortOrder.duration => '时长',
  };

  Future<void> _showSortSheet(BuildContext context, WidgetRef ref) async {
    await showModalBottomSheet<void>(
      context: context,
      backgroundColor: context.skinColors.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (ctx) => Consumer(
        builder: (ctx, ref, _) {
          final current = ref.watch(localLibrarySortProvider);
          return SafeArea(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Text(
                    '排序方式',
                    style: Theme.of(ctx).textTheme.headlineSmall,
                  ),
                ),
                const Divider(height: 1),
                for (final sort in LocalLibrarySortOrder.values)
                  ListTile(
                    title: Text(_sortLabel(sort)),
                    leading: current == sort
                        ? Icon(Icons.check, color: ctx.skinColors.sakuraPink)
                        : const SizedBox(width: 24),
                    onTap: () {
                      ref.read(localLibrarySortProvider.notifier).setSort(sort);
                      Navigator.pop(ctx);
                    },
                  ),
              ],
            ),
          );
        },
      ),
    );
  }
}

enum _ImportAction { files, folder }

class _EmptyLocalLibrary extends ConsumerWidget {
  const _EmptyLocalLibrary();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifier = ref.read(localLibraryProvider.notifier);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.library_music_outlined,
              size: 64,
              color: context.skinColors.onSurfaceVariant,
            ),
            const SizedBox(height: 16),
            Text(
              '本地曲库还是空的',
              style: TextStyle(
                fontSize: 18,
                color: context.skinColors.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              '以游客身份使用时，音乐来自你设备上的文件，不需要账号或服务器。',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 13,
                color: context.skinColors.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 24),
            FilledButton.icon(
              icon: const Icon(Icons.audio_file_outlined),
              label: const Text('导入文件'),
              onPressed: notifier.importFiles,
            ),
            const SizedBox(height: 8),
            OutlinedButton.icon(
              icon: const Icon(Icons.folder_outlined),
              label: const Text('导入文件夹'),
              onPressed: notifier.importFolder,
            ),
          ],
        ),
      ),
    );
  }
}

class _LocalTrackTile extends StatelessWidget {
  const _LocalTrackTile({
    super.key,
    required this.track,
    required this.isCurrent,
    required this.isPlaying,
    required this.onTap,
    required this.onDelete,
  });

  final LocalLibraryTrack track;
  final bool isCurrent;
  final bool isPlaying;
  final VoidCallback onTap;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    final subtitleParts = [
      if (track.artistName.isNotEmpty) track.artistName,
      if (track.durationMs != null)
        _formatDuration(Duration(milliseconds: track.durationMs!)),
    ];
    return ListTile(
      leading: _Cover(coverPath: track.coverArtPath, highlighted: isCurrent),
      title: Text(
        track.title,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          color: isCurrent
              ? context.skinColors.sakuraPink
              : context.skinColors.onSurface,
          fontWeight: isCurrent ? FontWeight.w600 : FontWeight.normal,
        ),
      ),
      subtitle: subtitleParts.isEmpty
          ? null
          : Text(
              subtitleParts.join(' · '),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (track.fileFormat != null) _FormatBadge(format: track.fileFormat!),
          if (isCurrent && isPlaying)
            Padding(
              padding: const EdgeInsets.only(left: 6),
              child: Icon(
                Icons.equalizer,
                color: context.skinColors.sakuraPinkLight,
                size: 20,
              ),
            ),
          IconButton(
            icon: Icon(
              Icons.delete_outline,
              color: context.skinColors.onSurfaceVariant,
            ),
            onPressed: onDelete,
          ),
        ],
      ),
      onTap: onTap,
    );
  }

  static String _formatDuration(Duration d) {
    final mins = d.inMinutes;
    final secs = d.inSeconds % 60;
    return '$mins:${secs.toString().padLeft(2, '0')}';
  }
}

/// Compact file-format hint (e.g. "FLAC") — the list row equivalent of the
/// full technical-detail panel on the player screen. Only shown for tracks
/// imported after v5.19.0 added the capture (older rows have a null format).
class _FormatBadge extends StatelessWidget {
  const _FormatBadge({required this.format});
  final String format;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(right: 4),
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: context.skinColors.surfaceContainer,
        borderRadius: BorderRadius.circular(4),
        border: Border.all(
          color: context.skinColors.outlineVariant,
          width: 0.5,
        ),
      ),
      child: Text(
        format,
        style: TextStyle(
          fontSize: 10,
          fontWeight: FontWeight.w600,
          color: context.skinColors.onSurfaceVariant,
        ),
      ),
    );
  }
}

class _Cover extends StatelessWidget {
  const _Cover({required this.coverPath, required this.highlighted});
  final String? coverPath;
  final bool highlighted;

  @override
  Widget build(BuildContext context) {
    final path = coverPath;
    return Container(
      width: 44,
      height: 44,
      decoration: BoxDecoration(
        color: context.skinColors.surfaceContainer,
        borderRadius: BorderRadius.circular(8),
      ),
      clipBehavior: Clip.antiAlias,
      child: path != null
          ? Image.file(
              File(path),
              fit: BoxFit.cover,
              errorBuilder: (_, _, _) => _fallbackIcon(context),
            )
          : _fallbackIcon(context),
    );
  }

  Widget _fallbackIcon(BuildContext context) => Icon(
    Icons.music_note_rounded,
    color: highlighted
        ? context.skinColors.sakuraPink
        : context.skinColors.onSurfaceVariant,
  );
}
