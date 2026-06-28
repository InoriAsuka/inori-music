// ignore_for_file: implementation_imports
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:inori_music/src/catalog/catalog_cache_providers.dart';
import 'package:inori_music/src/offline/download_notifier.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';
import 'package:inori_music/src/user_playlist/user_playlist_notifier.dart';

class TrackListTile extends ConsumerWidget {
  const TrackListTile({
    super.key,
    required this.track,
    this.isFavorite = false,
    this.onFavoriteTap,
    this.onTap,
    this.artworkUrl,
  });

  final CatalogTrack track;
  final bool isFavorite;
  final VoidCallback? onFavoriteTap;
  final VoidCallback? onTap;

  /// Optional artwork URL. When provided, a thumbnail is shown instead of
  /// the track number / music-note icon.
  final String? artworkUrl;

  static String _formatDurationMs(int? ms) {
    if (ms == null) return '';
    final totalSec = ms ~/ 1000;
    final m = totalSec ~/ 60;
    final s = totalSec % 60;
    return '$m:${s.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final durationStr = _formatDurationMs(track.durationMs);

    Widget leading;
    if (artworkUrl != null && artworkUrl!.isNotEmpty) {
      leading = ClipRRect(
        borderRadius: BorderRadius.circular(4),
        child: CachedNetworkImage(
          imageUrl: artworkUrl!,
          width: 40,
          height: 40,
          fit: BoxFit.cover,
          placeholder: (ctx, url) => const _ArtworkPlaceholder(),
          errorWidget: (ctx, url, err) => const _ArtworkPlaceholder(),
        ),
      );
    } else {
      leading = SizedBox(
        width: 40,
        child: Center(
          child: track.trackNumber != null
              ? Text(
                  '${track.trackNumber}',
                  style: const TextStyle(
                    color: NeonShrineColors.onSurfaceVariant,
                    fontSize: 13,
                  ),
                )
              : const Icon(
                  Icons.music_note,
                  size: 18,
                  color: NeonShrineColors.onSurfaceVariant,
                ),
        ),
      );
    }

    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 2),
      leading: leading,
      title: Text(
        track.title,
        style: const TextStyle(
          color: NeonShrineColors.onSurface,
          fontSize: 14,
          fontWeight: FontWeight.w500,
        ),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: track.artistId.isNotEmpty
          ? ref.watch(artistNameProvider(track.artistId)).when(
                data: (name) => Text(
                  name,
                  style: const TextStyle(
                    color: NeonShrineColors.onSurfaceVariant,
                    fontSize: 12,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                loading: () => const SizedBox.shrink(),
                error: (e, st) => Text(
                  track.artistId,
                  style: const TextStyle(
                    color: NeonShrineColors.onSurfaceVariant,
                    fontSize: 12,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              )
          : null,
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (durationStr.isNotEmpty)
            Text(
              durationStr,
              style: const TextStyle(
                color: NeonShrineColors.onSurfaceVariant,
                fontSize: 12,
              ),
            ),
          const SizedBox(width: 8),
          GestureDetector(
            onTap: onFavoriteTap,
            child: Icon(
              isFavorite ? Icons.favorite : Icons.favorite_border,
              size: 20,
              color: isFavorite
                  ? NeonShrineColors.accentPink
                  : NeonShrineColors.onSurfaceVariant,
            ),
          ),
        ],
      ),
      onTap: onTap ??
          () {
            ref.read(playerProvider.notifier).playTrack(track.id);
          },
      onLongPress: () => _showTrackMenu(context, ref),
    );
  }

  void _showTrackMenu(BuildContext context, WidgetRef ref) {
    final isDownloaded =
        ref.read(downloadProvider)[track.id] is DownloadDone;
    showModalBottomSheet<void>(
      context: context,
      builder: (ctx) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // --- Playlist ---
          ListTile(
            leading: const Icon(Icons.playlist_add,
                color: NeonShrineColors.primaryViolet),
            title: const Text('Add to playlist'),
            onTap: () {
              Navigator.pop(ctx);
              showModalBottomSheet<void>(
                context: context,
                builder: (_) => _AddToPlaylistSheet(trackId: track.id),
              );
            },
          ),
          // --- Download ---
          ListTile(
            leading: Icon(
                isDownloaded ? Icons.delete_outline : Icons.download),
            title:
                Text(isDownloaded ? 'Delete download' : 'Download for offline'),
            onTap: () {
              Navigator.pop(ctx);
              if (isDownloaded) {
                ref.read(downloadProvider.notifier).deleteDownload(track.id);
              } else {
                ref.read(downloadProvider.notifier).startDownload(track.id);
              }
            },
          ),
        ],
      ),
    );
  }
}

class _ArtworkPlaceholder extends StatelessWidget {
  const _ArtworkPlaceholder();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 40,
      height: 40,
      color: NeonShrineColors.surfaceContainer,
      child: const Icon(Icons.music_note, size: 18, color: NeonShrineColors.onSurfaceVariant),
    );
  }
}

class _AddToPlaylistSheet extends ConsumerStatefulWidget {
  const _AddToPlaylistSheet({required this.trackId});
  final String trackId;

  @override
  ConsumerState<_AddToPlaylistSheet> createState() => _AddToPlaylistSheetState();
}

class _AddToPlaylistSheetState extends ConsumerState<_AddToPlaylistSheet> {
  Future<void> _createAndAdd() async {
    Navigator.of(context).pop();
    final controller = TextEditingController();
    final name = await showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('New Playlist'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Playlist name'),
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
      final pl = await ref.read(userPlaylistProvider.notifier).create(name);
      if (pl != null) {
        await ref
            .read(userPlaylistProvider.notifier)
            .addTrack(pl.id, widget.trackId);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final playlists = ref.watch(userPlaylistProvider).valueOrNull ?? [];

    return SafeArea(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Padding(
            padding: EdgeInsets.all(16),
            child: Text(
              'Add to Playlist',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: NeonShrineColors.onSurface,
              ),
            ),
          ),
          if (playlists.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Text(
                'No playlists yet.',
                style: TextStyle(color: NeonShrineColors.onSurfaceVariant),
              ),
            )
          else
            ...playlists.map(
              (pl) => ListTile(
                leading: const Icon(Icons.queue_music,
                    color: NeonShrineColors.primaryViolet),
                title: Text(pl.name),
                subtitle: Text('${pl.trackIds.length} tracks'),
                onTap: () async {
                  Navigator.of(context).pop();
                  await ref
                      .read(userPlaylistProvider.notifier)
                      .addTrack(pl.id, widget.trackId);
                },
              ),
            ),
          ListTile(
            leading: const Icon(Icons.add, color: NeonShrineColors.primaryViolet),
            title: const Text('+ New Playlist'),
            onTap: _createAndAdd,
          ),
          const SizedBox(height: 8),
        ],
      ),
    );
  }
}
