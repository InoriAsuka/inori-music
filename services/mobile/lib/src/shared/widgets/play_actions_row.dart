// ignore_for_file: implementation_imports
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// "Play all" / "Shuffle" action row for a collection detail header (album,
/// playlist, artist) — these screens previously had no way to start
/// playback other than tapping individual track rows one at a time.
///
/// Matches EchoMusic's detail-page header pattern (a primary play action
/// alongside secondary actions); album/playlist-level favorite/follow are
/// deliberately not included here — this project has no backing API for
/// collection-level favorites, only per-track ones (see [track_list_tile]).
class PlayActionsRow extends ConsumerWidget {
  const PlayActionsRow({super.key, required this.tracksState});

  final AsyncValue<List<CatalogTrack>> tracksState;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tracks = tracksState.valueOrNull;
    final hasTracks = tracks != null && tracks.isNotEmpty;

    void play({required bool shuffle}) {
      final notifier = ref.read(playerProvider.notifier);
      final ids = tracks!.map((t) => t.id).toList();
      if (shuffle) ids.shuffle();
      notifier.playQueue(ids);
    }

    return Row(
      children: [
        FilledButton.icon(
          onPressed: hasTracks ? () => play(shuffle: false) : null,
          icon: const Icon(Icons.play_arrow_rounded, size: 20),
          label: const Text('Play'),
          style: FilledButton.styleFrom(
            backgroundColor: context.skinColors.sakuraPink,
            foregroundColor: Colors.white,
          ),
        ),
        const SizedBox(width: 8),
        IconButton(
          onPressed: hasTracks ? () => play(shuffle: true) : null,
          icon: const Icon(Icons.shuffle),
          tooltip: 'Shuffle',
          color: context.skinColors.onSurfaceVariant,
        ),
      ],
    );
  }
}
