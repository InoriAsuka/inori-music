import 'dart:io';

import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/local_library/local_library_db.dart';
import 'package:inori_music/src/local_library/local_library_notifier.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/desktop_app_bar.dart';

/// Guest mode's home screen: a flat list of locally-imported audio files.
/// No account, no server — this is what makes the app usable without login.
/// v1 is intentionally a single flat list (artist/album/title sorted) rather
/// than a full Artists→Albums→Tracks hierarchy: a personal local file
/// collection is typically far smaller than a server catalog, so grouped
/// browsing is a later candidate, not core scope.
///
/// The toolbar/row interactions follow Spotube's local-library page: a
/// play-all/shuffle pair, an in-library filter, hover-to-play on the artwork,
/// and long-press multi-select for bulk removal.
class LocalLibraryScreen extends ConsumerStatefulWidget {
  const LocalLibraryScreen({super.key});

  @override
  ConsumerState<LocalLibraryScreen> createState() => _LocalLibraryScreenState();
}

class _LocalLibraryScreenState extends ConsumerState<LocalLibraryScreen> {
  final _searchController = TextEditingController();

  /// Purely client-side filter over the already-loaded rows — the local
  /// library is a single small table, so there is nothing to query remotely
  /// and no debounce worth adding.
  String _query = '';

  /// Non-empty means selection mode is active. Ids rather than indices, so a
  /// concurrent import/removal reordering the list can't shift the selection
  /// onto different tracks.
  final _selected = <String>{};

  bool get _selectionMode => _selected.isNotEmpty;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  List<LocalLibraryTrack> _filter(List<LocalLibraryTrack> tracks) {
    if (_query.isEmpty) return tracks;
    final q = _query.toLowerCase();
    return tracks
        .where(
          (t) =>
              t.title.toLowerCase().contains(q) ||
              t.artistName.toLowerCase().contains(q) ||
              t.albumTitle.toLowerCase().contains(q),
        )
        .toList();
  }

  void _exitSelection() => setState(_selected.clear);

  void _toggleSelected(String id) {
    setState(() {
      if (!_selected.remove(id)) _selected.add(id);
    });
  }

  /// Runs an import and always says what happened. Every outcome used to look
  /// the same on screen — nothing — so a failing picker, an unreadable file
  /// and a plain cancel were indistinguishable.
  Future<void> _runImport(Future<ImportOutcome> Function() action) async {
    final outcome = await action();
    if (!mounted || outcome.cancelled) return;
    final String message;
    if (outcome.imported == 0 && outcome.hasError) {
      message = '导入失败：${outcome.firstError}';
    } else if (outcome.failed > 0) {
      message =
          '已导入 ${outcome.imported} 首，${outcome.failed} 首失败'
          '（${outcome.firstError}）';
    } else if (outcome.imported == 0) {
      message = '没有找到可导入的音频文件';
    } else {
      message = '已导入 ${outcome.imported} 首';
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }

  void _play(List<LocalLibraryTrack> tracks, {required bool shuffle}) {
    if (tracks.isEmpty) return;
    final ids = tracks.map((t) => t.id).toList();
    if (shuffle) ids.shuffle();
    ref.read(playerProvider.notifier).playQueue(ids);
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(localLibraryProvider);
    final playerState = ref.watch(playerProvider);

    return Scaffold(
      appBar: _selectionMode ? _selectionAppBar() : _browseAppBar(),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Text('$e', style: TextStyle(color: context.skinColors.error)),
        ),
        data: (tracks) {
          if (tracks.isEmpty) {
            final notifier = ref.read(localLibraryProvider.notifier);
            return _EmptyLocalLibrary(
              onImportFiles: () => _runImport(notifier.importFiles),
              onImportFolder: () => _runImport(notifier.importFolder),
            );
          }
          final visible = _filter(tracks);
          return Column(
            children: [
              _Toolbar(
                controller: _searchController,
                onQueryChanged: (q) => setState(() => _query = q),
                trackCount: visible.length,
                onPlayAll: visible.isEmpty
                    ? null
                    : () => _play(visible, shuffle: false),
                onShuffle: visible.isEmpty
                    ? null
                    : () => _play(visible, shuffle: true),
              ),
              const Divider(height: 1),
              Expanded(
                child: visible.isEmpty
                    ? Center(
                        child: Text(
                          '没有匹配「$_query」的曲目',
                          style: TextStyle(
                            color: context.skinColors.onSurfaceVariant,
                          ),
                        ),
                      )
                    : ListView.builder(
                        itemCount: visible.length,
                        itemBuilder: (context, i) {
                          final track = visible[i];
                          final isCurrent =
                              playerState.mediaItem?.id == track.id;
                          return _LocalTrackTile(
                            key: ValueKey(track.id),
                            track: track,
                            isCurrent: isCurrent,
                            isPlaying: isCurrent && playerState.isPlaying,
                            selectionMode: _selectionMode,
                            isSelected: _selected.contains(track.id),
                            // In selection mode a plain tap extends the
                            // selection instead of starting playback — the
                            // usual mobile file-manager convention.
                            onTap: () => _selectionMode
                                ? _toggleSelected(track.id)
                                : ref
                                      .read(playerProvider.notifier)
                                      .playQueue(
                                        visible.map((t) => t.id).toList(),
                                        initialIndex: i,
                                      ),
                            onToggleSelected: () => _toggleSelected(track.id),
                            onDelete: () => _confirmDelete(context, track),
                          );
                        },
                      ),
              ),
            ],
          );
        },
      ),
    );
  }

  DesktopAppBar _browseAppBar() => DesktopAppBar(
    title: const Text('本地曲库'),
    actions: [
      PopupMenuButton<_ImportAction>(
        icon: const Icon(Icons.add),
        tooltip: '导入',
        onSelected: (action) {
          final notifier = ref.read(localLibraryProvider.notifier);
          switch (action) {
            case _ImportAction.files:
              _runImport(notifier.importFiles);
            case _ImportAction.folder:
              _runImport(notifier.importFolder);
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
    ],
  );

  DesktopAppBar _selectionAppBar() => DesktopAppBar(
    automaticallyImplyLeading: false,
    leading: IconButton(
      icon: const Icon(Icons.close),
      tooltip: '退出选择',
      onPressed: _exitSelection,
    ),
    title: Text('已选择 ${_selected.length} 项'),
    actions: [
      IconButton(
        icon: const Icon(Icons.delete_outline),
        tooltip: '移除所选',
        onPressed: _confirmDeleteSelected,
      ),
    ],
  );

  Future<void> _confirmDelete(
    BuildContext context,
    LocalLibraryTrack track,
  ) async {
    final confirmed = await _confirm(
      context,
      title: '移除曲目',
      body: '从本地曲库中移除「${track.title}」？不会删除原始文件。',
    );
    if (confirmed) {
      await ref.read(localLibraryProvider.notifier).remove(track.id);
    }
  }

  Future<void> _confirmDeleteSelected() async {
    final count = _selected.length;
    final confirmed = await _confirm(
      context,
      title: '移除所选曲目',
      body: '从本地曲库中移除 $count 首曲目？不会删除原始文件。',
    );
    if (!confirmed) return;
    // Snapshot before clearing: _exitSelection empties the same set.
    final ids = _selected.toList();
    _exitSelection();
    await ref.read(localLibraryProvider.notifier).removeAll(ids);
  }

  Future<bool> _confirm(
    BuildContext context, {
    required String title,
    required String body,
  }) async {
    final result = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(title),
        content: Text(body),
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
    return result ?? false;
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

/// Play-all / shuffle / filter strip above the list. Before this the screen
/// had no way to start playback other than tapping a row, and no way to find
/// a track in a large import other than scrolling.
class _Toolbar extends StatelessWidget {
  const _Toolbar({
    required this.controller,
    required this.onQueryChanged,
    required this.trackCount,
    required this.onPlayAll,
    required this.onShuffle,
  });

  final TextEditingController controller;
  final ValueChanged<String> onQueryChanged;
  final int trackCount;
  final VoidCallback? onPlayAll;
  final VoidCallback? onShuffle;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 8),
      child: Row(
        children: [
          FilledButton.icon(
            onPressed: onPlayAll,
            icon: const Icon(Icons.play_arrow_rounded, size: 20),
            label: const Text('播放全部'),
            style: FilledButton.styleFrom(
              backgroundColor: context.skinColors.sakuraPink,
              foregroundColor: Colors.white,
            ),
          ),
          const SizedBox(width: 8),
          IconButton(
            onPressed: onShuffle,
            icon: const Icon(Icons.shuffle),
            tooltip: '随机播放',
            color: context.skinColors.onSurfaceVariant,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: TextField(
              controller: controller,
              onChanged: onQueryChanged,
              decoration: InputDecoration(
                isDense: true,
                hintText: '在曲库中搜索',
                prefixIcon: const Icon(Icons.search, size: 20),
                suffixIcon: controller.text.isEmpty
                    ? null
                    : IconButton(
                        icon: const Icon(Icons.clear, size: 18),
                        onPressed: () {
                          controller.clear();
                          onQueryChanged('');
                        },
                      ),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(20),
                ),
                contentPadding: const EdgeInsets.symmetric(vertical: 8),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Text(
            '$trackCount 首',
            style: TextStyle(
              fontSize: 12,
              color: context.skinColors.onSurfaceVariant,
            ),
          ),
        ],
      ),
    );
  }
}

class _EmptyLocalLibrary extends StatelessWidget {
  const _EmptyLocalLibrary({
    required this.onImportFiles,
    required this.onImportFolder,
  });

  // Handed down rather than calling the notifier directly, so these go
  // through the screen's _runImport and report their outcome. This screen is
  // exactly where a broken import is invisible: an empty library that stays
  // empty looks the same whether the picker failed or the user cancelled.
  final VoidCallback onImportFiles;
  final VoidCallback onImportFolder;

  @override
  Widget build(BuildContext context) {
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
              onPressed: onImportFiles,
            ),
            const SizedBox(height: 8),
            OutlinedButton.icon(
              icon: const Icon(Icons.folder_outlined),
              label: const Text('导入文件夹'),
              onPressed: onImportFolder,
            ),
          ],
        ),
      ),
    );
  }
}

class _LocalTrackTile extends StatefulWidget {
  const _LocalTrackTile({
    super.key,
    required this.track,
    required this.isCurrent,
    required this.isPlaying,
    required this.selectionMode,
    required this.isSelected,
    required this.onTap,
    required this.onToggleSelected,
    required this.onDelete,
  });

  final LocalLibraryTrack track;
  final bool isCurrent;
  final bool isPlaying;
  final bool selectionMode;
  final bool isSelected;
  final VoidCallback onTap;
  final VoidCallback onToggleSelected;
  final VoidCallback onDelete;

  @override
  State<_LocalTrackTile> createState() => _LocalTrackTileState();
}

class _LocalTrackTileState extends State<_LocalTrackTile> {
  /// Pointer-driven only — MouseRegion never fires for touch, so phones see
  /// exactly the row this screen has always rendered.
  bool _hovering = false;

  @override
  Widget build(BuildContext context) {
    final track = widget.track;
    final subtitleParts = [
      if (track.artistName.isNotEmpty) track.artistName,
      if (track.durationMs != null)
        _formatDuration(Duration(milliseconds: track.durationMs!)),
    ];

    final tile = ListTile(
      selected: widget.isSelected,
      selectedTileColor: context.skinColors.sakuraPinkDark.withValues(
        alpha: 0.25,
      ),
      leading: widget.selectionMode
          ? Checkbox(
              value: widget.isSelected,
              onChanged: (_) => widget.onToggleSelected(),
            )
          : _Cover(
              coverPath: track.coverArtPath,
              highlighted: widget.isCurrent,
              showPlayOverlay: _hovering,
              isPlaying: widget.isCurrent && widget.isPlaying,
            ),
      title: Text(
        track.title,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          color: widget.isCurrent
              ? context.skinColors.sakuraPink
              : context.skinColors.onSurface,
          fontWeight: widget.isCurrent ? FontWeight.w600 : FontWeight.normal,
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
          if (widget.isCurrent && widget.isPlaying)
            Padding(
              padding: const EdgeInsets.only(left: 6),
              child: Icon(
                Icons.equalizer,
                color: context.skinColors.sakuraPinkLight,
                size: 20,
              ),
            ),
          // Per-row delete would be redundant (and easy to mis-tap) once the
          // selection app bar owns removal.
          if (!widget.selectionMode)
            IconButton(
              icon: Icon(
                Icons.delete_outline,
                color: context.skinColors.onSurfaceVariant,
              ),
              onPressed: widget.onDelete,
            ),
        ],
      ),
      onTap: widget.onTap,
      onLongPress: widget.onToggleSelected,
    );

    return MouseRegion(
      onEnter: (_) => setState(() => _hovering = true),
      onExit: (_) => setState(() => _hovering = false),
      child: Listener(
        // Right-click enters/extends the selection: desktop has no long-press
        // gesture, so multi-select would otherwise be mouse-unreachable.
        onPointerDown: (event) {
          if (event.kind == PointerDeviceKind.mouse &&
              event.buttons == kSecondaryMouseButton) {
            widget.onToggleSelected();
          }
        },
        child: tile,
      ),
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
  const _Cover({
    required this.coverPath,
    required this.highlighted,
    this.showPlayOverlay = false,
    this.isPlaying = false,
  });

  final String? coverPath;
  final bool highlighted;
  final bool showPlayOverlay;

  /// Drives which glyph the hover overlay shows — pausing the row you are
  /// already listening to is the useful action there, not restarting it.
  final bool isPlaying;

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
      child: Stack(
        fit: StackFit.expand,
        children: [
          if (path != null)
            Image.file(
              File(path),
              fit: BoxFit.cover,
              errorBuilder: (_, _, _) => _fallbackIcon(context),
            )
          else
            _fallbackIcon(context),
          if (showPlayOverlay)
            Container(
              color: Colors.black.withValues(alpha: 0.45),
              child: Icon(
                isPlaying ? Icons.pause_rounded : Icons.play_arrow_rounded,
                size: 24,
                color: Colors.white,
              ),
            ),
        ],
      ),
    );
  }

  Widget _fallbackIcon(BuildContext context) => Icon(
    Icons.music_note_rounded,
    color: highlighted
        ? context.skinColors.sakuraPink
        : context.skinColors.onSurfaceVariant,
  );
}
