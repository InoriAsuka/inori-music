import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/user_playlist/user_playlist_notifier.dart';

/// Library tab — lists all playlists owned by the current user.
class UserPlaylistListScreen extends ConsumerWidget {
  const UserPlaylistListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final playlistsAsync = ref.watch(userPlaylistProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('My Playlists'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            tooltip: 'New playlist',
            onPressed: () => _showCreateDialog(context, ref),
          ),
        ],
      ),
      body: playlistsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.error_outline,
                  color: NeonShrineColors.error, size: 48),
              const SizedBox(height: 12),
              Text(e.toString(), textAlign: TextAlign.center),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: () => ref.read(userPlaylistProvider.notifier).load(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (playlists) {
          if (playlists.isEmpty) {
            return const Center(child: Text('No playlists yet. Tap + to create one.'));
          }
          return ListView.builder(
            itemCount: playlists.length,
            itemBuilder: (ctx, i) {
              final pl = playlists[i];
              return ListTile(
                leading: const Icon(
                  Icons.queue_music,
                  color: NeonShrineColors.primaryViolet,
                ),
                title: Text(pl.name),
                subtitle: Text('${pl.trackIds.length} tracks'),
                onTap: () =>
                    ctx.push(AppRoutes.myPlaylistDetailPath(pl.id)),
              );
            },
          );
        },
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref) async {
    final controller = TextEditingController();
    final name = await showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('New Playlist'),
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
            child: const Text('Create'),
          ),
        ],
      ),
    );
    if (name != null && name.isNotEmpty) {
      await ref.read(userPlaylistProvider.notifier).create(name);
    }
  }
}
