// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/playlist.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/playlists_screen.dart'
    show catalogPlaylistsProvider;
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/user_playlist/user_playlist_notifier.dart';

/// The desktop sidebar's "歌单区": EchoMusic's own `自建歌单 | 收藏歌单` split —
/// this app's two playlist subsystems (`user_playlist_notifier.dart`'s
/// `userPlaylistProvider`, this account's own playlists; `catalog/
/// playlists_screen.dart`'s `catalogPlaylistsProvider`, the server's curated
/// catalog playlists) mapped one-to-one onto the two tabs.
///
/// That mapping is a judgement call, not a given: the server has no
/// per-user "favourited playlist" relation at all (checked — there is no
/// `favoritePlaylist`/playlist-favourite endpoint anywhere in the generated
/// API), so "收藏歌单" cannot mean "playlists this account favourited" the
/// way EchoMusic's own does. Catalog playlists — admin-curated collections
/// every account can browse — are the closest existing concept to
/// "collected into the library from the catalog", so that tab points there
/// instead. Flagged in requirement.md's v5.33.0 entry for review.
///
/// Both tabs preview a handful of rows and always offer a "查看全部" link to
/// the existing full-screen list (`UserPlaylistListScreen`/
/// `PlaylistsScreen`) rather than reimplementing pagination/creation here —
/// this section is a preview + shortcut, not a second copy of either
/// screen.
class SidebarPlaylistsSection extends ConsumerStatefulWidget {
  const SidebarPlaylistsSection({super.key});

  @override
  ConsumerState<SidebarPlaylistsSection> createState() =>
      _SidebarPlaylistsSectionState();
}

class _SidebarPlaylistsSectionState
    extends ConsumerState<SidebarPlaylistsSection> {
  bool _showCreated = true;
  bool _sortAlpha = false;

  /// Rows previewed inline before the section defers to "查看全部" — a
  /// sidebar accessory should read as a shortcut, not grow into a second
  /// scrollable list competing with the nav groups above it.
  static const _previewCount = 5;

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 4, 4, 0),
          child: Row(
            children: [
              // Flexible, not Expanded straight on the tab Row — the 220px
              // sidebar (minus this Padding's own insets and the two icon
              // buttons on the right) leaves the tab labels well under 150px
              // to share, and at that width even a compact IconTheme-sized
              // pair of CJK tab labels ("自建歌单"/"收藏歌单") can outgrow it.
              // FittedBox lets the pair scale down as one unit rather than
              // each label wrapping/ellipsing independently and reading as
              // two mismatched sizes — a v5.33.0 field regression
              // (`RenderFlex overflowed ... SidebarPlaylistsSection`) caught
              // in test/shell_scaffold_nav_test.dart before this shipped.
              Expanded(
                child: FittedBox(
                  fit: BoxFit.scaleDown,
                  alignment: Alignment.centerLeft,
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      _TabLabel(
                        label: t.createdPlaylists,
                        selected: _showCreated,
                        onTap: () => setState(() => _showCreated = true),
                      ),
                      const SizedBox(width: 8),
                      _TabLabel(
                        label: t.collectedPlaylists,
                        selected: !_showCreated,
                        onTap: () => setState(() => _showCreated = false),
                      ),
                    ],
                  ),
                ),
              ),
              _HeaderIconButton(
                icon: Icons.sort_by_alpha,
                tooltip: t.sortAlphabetically,
                color: _sortAlpha
                    ? context.skinColors.sakuraPinkLight
                    : context.skinColors.onSurfaceVariant,
                onPressed: () => setState(() => _sortAlpha = !_sortAlpha),
              ),
              _HeaderIconButton(
                icon: Icons.refresh,
                tooltip: t.refresh,
                color: context.skinColors.onSurfaceVariant,
                onPressed: () => _showCreated
                    ? ref.read(userPlaylistProvider.notifier).load()
                    : ref.invalidate(catalogPlaylistsProvider),
              ),
            ],
          ),
        ),
        _showCreated ? _createdTab(context) : _collectedTab(context),
      ],
    );
  }

  Widget _createdTab(BuildContext context) {
    final t = AppLocalizations.of(context);
    final async = ref.watch(userPlaylistProvider);
    return async.when(
      loading: () => const _InlineLoading(),
      error: (_, _) => _EmptyRow(text: t.noCreatedPlaylistsYet),
      data: (playlists) {
        final sorted = _sortAlpha
            ? (List<UserPlaylist>.of(playlists)
                ..sort((a, b) => a.name.compareTo(b.name)))
            : playlists;
        return Column(
          children: [
            if (sorted.isEmpty) _EmptyRow(text: t.noCreatedPlaylistsYet),
            for (final pl in sorted.take(_previewCount))
              _PlaylistRow(
                title: pl.name,
                trackCount: pl.trackIds.length,
                onTap: () =>
                    context.push(AppRoutes.myPlaylistDetailPath(pl.id)),
              ),
            _ViewAllRow(onTap: () => context.go(AppRoutes.myPlaylists)),
          ],
        );
      },
    );
  }

  Widget _collectedTab(BuildContext context) {
    final t = AppLocalizations.of(context);
    final async = ref.watch(catalogPlaylistsProvider);
    return async.when(
      loading: () => const _InlineLoading(),
      error: (_, _) => _EmptyRow(text: t.noCollectedPlaylistsYet),
      data: (playlists) {
        final sorted = _sortAlpha
            ? (List<Playlist>.of(playlists)
                ..sort((a, b) => a.name.compareTo(b.name)))
            : playlists;
        return Column(
          children: [
            if (sorted.isEmpty) _EmptyRow(text: t.noCollectedPlaylistsYet),
            for (final pl in sorted.take(_previewCount))
              _PlaylistRow(
                title: pl.name,
                trackCount: pl.trackIds.length,
                onTap: () => context.go(AppRoutes.playlistDetailPath(pl.id)),
              ),
            _ViewAllRow(onTap: () => context.go(AppRoutes.playlists)),
          ],
        );
      },
    );
  }
}

/// A 22x22 icon button for the playlist section's own header row — plain
/// [IconButton] defaults to a 48x48 minimum tap target regardless of
/// [VisualDensity] (that only nudges the default, it does not remove the
/// floor), which is what actually overflowed this row inside the 220px
/// sidebar (two of them ate ~96px before the tab labels got a single
/// pixel). Explicit tight [BoxConstraints] is the only way to get genuinely
/// smaller than that floor.
class _HeaderIconButton extends StatelessWidget {
  const _HeaderIconButton({
    required this.icon,
    required this.tooltip,
    required this.color,
    required this.onPressed,
  });

  final IconData icon;
  final String tooltip;
  final Color color;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return IconButton(
      icon: Icon(icon, size: 14, color: color),
      tooltip: tooltip,
      padding: EdgeInsets.zero,
      constraints: const BoxConstraints.tightFor(width: 22, height: 22),
      onPressed: onPressed,
    );
  }
}

class _TabLabel extends StatelessWidget {
  const _TabLabel({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(6),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 2),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 12.5,
            fontWeight: selected ? FontWeight.w700 : FontWeight.w400,
            color: selected
                ? context.skinColors.onSurface
                : context.skinColors.onSurfaceVariant,
          ),
        ),
      ),
    );
  }
}

class _PlaylistRow extends StatelessWidget {
  const _PlaylistRow({
    required this.title,
    required this.trackCount,
    required this.onTap,
  });

  final String title;
  final int trackCount;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      dense: true,
      visualDensity: VisualDensity.compact,
      contentPadding: const EdgeInsets.only(left: 16, right: 8),
      leading: Icon(
        Icons.queue_music,
        size: 18,
        color: context.skinColors.onSurfaceVariant,
      ),
      title: Text(
        title,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(fontSize: 13, color: context.skinColors.onSurface),
      ),
      trailing: Text(
        '$trackCount',
        style: TextStyle(
          fontSize: 11,
          color: context.skinColors.onSurfaceVariant,
        ),
      ),
      onTap: onTap,
    );
  }
}

class _ViewAllRow extends StatelessWidget {
  const _ViewAllRow({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      dense: true,
      visualDensity: VisualDensity.compact,
      contentPadding: const EdgeInsets.only(left: 16, right: 8),
      leading: Icon(
        Icons.chevron_right,
        size: 18,
        color: context.skinColors.onSurfaceVariant,
      ),
      title: Text(
        AppLocalizations.of(context).viewAll,
        style: TextStyle(
          fontSize: 12.5,
          color: context.skinColors.onSurfaceVariant,
        ),
      ),
      onTap: onTap,
    );
  }
}

class _EmptyRow extends StatelessWidget {
  const _EmptyRow({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 12,
          color: context.skinColors.onSurfaceVariant,
        ),
      ),
    );
  }
}

class _InlineLoading extends StatelessWidget {
  const _InlineLoading();

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.symmetric(vertical: 12),
      child: Center(
        child: SizedBox(
          width: 16,
          height: 16,
          child: CircularProgressIndicator(strokeWidth: 2),
        ),
      ),
    );
  }
}
