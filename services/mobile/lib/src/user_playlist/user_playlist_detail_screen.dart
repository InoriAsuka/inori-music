import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/user_playlist/user_playlist_notifier.dart';

/// Playlist detail screen — loads tracks for a user playlist and allows
/// play, rename, and delete.
class UserPlaylistDetailScreen extends ConsumerStatefulWidget {
  const UserPlaylistDetailScreen({super.key, required this.playlistId});

  final String playlistId;

  @override
  ConsumerState<UserPlaylistDetailScreen> createState() =>
      _UserPlaylistDetailScreenState();
}

class _UserPlaylistDetailScreenState
    extends ConsumerState<UserPlaylistDetailScreen> {
  List<String>? _trackIds;
  bool _loading = true;
  String? _error;
  String _name = '';

  @override
  void initState() {
    super.initState();
    _loadTracks();
  }

  Future<void> _loadTracks() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final notifier = ref.read(userPlaylistProvider.notifier);
      final ids = await notifier.getTrackIds(widget.playlistId);
      // Get name from the cached list
      final playlists = ref.read(userPlaylistProvider).valueOrNull ?? [];
      final pl = playlists.where((p) => p.id == widget.playlistId).firstOrNull;
      setState(() {
        _trackIds = ids;
        _name = pl?.name ?? '';
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  Future<void> _showRenameDialog() async {
    final controller = TextEditingController(text: _name);
    final newName = await showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Rename Playlist'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Name'),
        ),
        actions: [
          TextButton(
            onPressed: () => ctx.pop(),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => ctx.pop(controller.text.trim()),
            child: const Text('Save'),
          ),
        ],
      ),
    );
    if (newName != null && newName.isNotEmpty && mounted) {
      await ref
          .read(userPlaylistProvider.notifier)
          .rename(widget.playlistId, newName);
      setState(() => _name = newName);
    }
  }

  Future<void> _confirmDelete() async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Delete Playlist'),
        content: Text('Delete "$_name"? This cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => ctx.pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => ctx.pop(true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (ok == true && mounted) {
      await ref
          .read(userPlaylistProvider.notifier)
          .delete(widget.playlistId);
      if (mounted) context.pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_name.isEmpty ? 'My Playlist' : _name),
        actions: [
          IconButton(
            icon: const Icon(Icons.edit_outlined),
            tooltip: 'Rename',
            onPressed: _showRenameDialog,
          ),
          IconButton(
            icon: const Icon(Icons.delete_outline),
            tooltip: 'Delete',
            onPressed: _confirmDelete,
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.error_outline,
                          color: NeonShrineColors.error, size: 48),
                      const SizedBox(height: 12),
                      Text(_error!, textAlign: TextAlign.center),
                      const SizedBox(height: 12),
                      FilledButton(
                        onPressed: _loadTracks,
                        child: const Text('Retry'),
                      ),
                    ],
                  ),
                )
              : _trackIds == null || _trackIds!.isEmpty
                  ? const Center(child: Text('No tracks in this playlist'))
                  : Column(
                      children: [
                        Padding(
                          padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
                          child: Row(
                            children: [
                              Text(
                                '${_trackIds!.length} tracks',
                                style: const TextStyle(
                                    color: NeonShrineColors.onSurfaceVariant),
                              ),
                              const Spacer(),
                              FilledButton.icon(
                                icon: const Icon(Icons.play_arrow, size: 18),
                                label: const Text('Play All'),
                                onPressed: () {
                                  // PlayerNotifier.playQueue is available via playerProvider
                                  // We just navigate to player after triggering queue
                                  context.go(AppRoutes.player);
                                },
                              ),
                            ],
                          ),
                        ),
                        Expanded(
                          child: ListView.builder(
                            itemCount: _trackIds!.length,
                            itemBuilder: (ctx, i) {
                              final tid = _trackIds![i];
                              return ListTile(
                                leading: SizedBox(
                                  width: 40,
                                  child: Center(
                                    child: Text(
                                      '${i + 1}',
                                      style: const TextStyle(
                                          color: NeonShrineColors
                                              .onSurfaceVariant),
                                    ),
                                  ),
                                ),
                                title: Text(
                                  tid,
                                  style: const TextStyle(
                                    color: NeonShrineColors.onSurface,
                                    fontSize: 13,
                                  ),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                trailing: IconButton(
                                  icon: const Icon(Icons.remove_circle_outline,
                                      size: 20,
                                      color: NeonShrineColors.onSurfaceVariant),
                                  onPressed: () async {
                                    await ref
                                        .read(userPlaylistProvider.notifier)
                                        .removeTrack(widget.playlistId, tid);
                                    await _loadTracks();
                                  },
                                ),
                              );
                            },
                          ),
                        ),
                      ],
                    ),
    );
  }
}
