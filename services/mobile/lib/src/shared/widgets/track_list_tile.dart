// ignore_for_file: implementation_imports
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/catalog/catalog_cache_providers.dart';
import 'package:inori_music/src/offline/download_notifier.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/user_playlist/user_playlist_notifier.dart';

class TrackListTile extends ConsumerStatefulWidget {
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

  @override
  ConsumerState<TrackListTile> createState() => _TrackListTileState();
}

class _TrackListTileState extends ConsumerState<TrackListTile> {
  /// Pointer-driven only — MouseRegion never fires for touch input, so this
  /// stays false on phones and the tile renders exactly as it did before.
  bool _hovering = false;

  static String _formatDurationMs(int? ms) {
    if (ms == null) return '';
    final totalSec = ms ~/ 1000;
    final m = totalSec ~/ 60;
    final s = totalSec % 60;
    return '$m:${s.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    final track = widget.track;
    final artworkUrl = widget.artworkUrl;
    final isFavorite = widget.isFavorite;
    final onFavoriteTap = widget.onFavoriteTap;
    final onTap = widget.onTap;

    final durationStr = _formatDurationMs(track.durationMs);

    // Prefer the explicit artworkUrl; fall back to artworkUrlProvider when albumId
    // is available.
    final albumId = track.albumId;
    final resolvedUrl = artworkUrl ??
        (albumId != null && albumId.isNotEmpty
            ? ref.watch(artworkUrlProvider(albumId)).value
            : null);

    Widget leading;
    if (resolvedUrl != null && resolvedUrl.isNotEmpty) {
      leading = ClipRRect(
        borderRadius: BorderRadius.circular(4),
        child: CachedNetworkImage(
          imageUrl: resolvedUrl,
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
                  style: TextStyle(
                    color: context.skinColors.onSurfaceVariant,
                    fontSize: 13,
                  ),
                )
              : Icon(
                  Icons.music_note,
                  size: 18,
                  color: context.skinColors.onSurfaceVariant,
                ),
        ),
      );
    }

    // Hovering the row scrims the thumbnail and reveals a play glyph over it,
    // matching Spotube's TrackTile. Purely an affordance — the tap target is
    // still the whole row, which already starts playback.
    leading = SizedBox(
      width: 40,
      height: 40,
      child: Stack(
        fit: StackFit.expand,
        children: [
          leading,
          if (_hovering)
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: Container(
                color: Colors.black.withValues(alpha: 0.45),
                child: const Icon(
                  Icons.play_arrow_rounded,
                  size: 22,
                  color: Colors.white,
                ),
              ),
            ),
        ],
      ),
    );

    final tile = ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 2),
      leading: leading,
      title: Text(
        track.title,
        style: TextStyle(
          color: context.skinColors.onSurface,
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
                  style: TextStyle(
                    color: context.skinColors.onSurfaceVariant,
                    fontSize: 12,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                loading: () => const SizedBox.shrink(),
                error: (e, st) => Text(
                  track.artistId,
                  style: TextStyle(
                    color: context.skinColors.onSurfaceVariant,
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
              style: TextStyle(
                color: context.skinColors.onSurfaceVariant,
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
                  ? context.skinColors.accentPink
                  : context.skinColors.onSurfaceVariant,
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

    return MouseRegion(
      onEnter: (_) => setState(() => _hovering = true),
      onExit: (_) => setState(() => _hovering = false),
      child: Listener(
        // Right-click opens the same menu long-press does — a desktop user has
        // no long-press gesture, so without this the "add to playlist" /
        // "download" actions were mouse-unreachable entirely.
        onPointerDown: (event) {
          if (event.kind == PointerDeviceKind.mouse &&
              event.buttons == kSecondaryMouseButton) {
            _showTrackMenu(context, ref);
          }
        },
        child: tile,
      ),
    );
  }

  void _showTrackMenu(BuildContext context, WidgetRef ref) {
    final track = widget.track;
    final isDownloaded =
        ref.read(downloadProvider)[track.id] is DownloadDone;
    showModalBottomSheet<void>(
      context: context,
      builder: (ctx) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // --- Playlist ---
          ListTile(
            leading: Icon(Icons.playlist_add,
                color: context.skinColors.sakuraPink),
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
      color: context.skinColors.surfaceContainer,
      child: Icon(Icons.music_note, size: 18, color: context.skinColors.onSurfaceVariant),
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
          Padding(
            padding: const EdgeInsets.all(16),
            child: Text(
              'Add to Playlist',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: context.skinColors.onSurface,
              ),
            ),
          ),
          if (playlists.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Text(
                'No playlists yet.',
                style: TextStyle(color: context.skinColors.onSurfaceVariant),
              ),
            )
          else
            ...playlists.map(
              (pl) => ListTile(
                leading: Icon(Icons.queue_music,
                    color: context.skinColors.sakuraPink),
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
            leading: Icon(Icons.add, color: context.skinColors.sakuraPink),
            title: const Text('+ New Playlist'),
            onTap: _createAndAdd,
          ),
          const SizedBox(height: 8),
        ],
      ),
    );
  }
}
