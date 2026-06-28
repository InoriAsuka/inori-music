// ignore_for_file: implementation_imports, unnecessary_non_null_assertion
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/catalog_artist.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

final _artistsProvider = FutureProvider<List<CatalogArtist>>((ref) {
  return ref.watch(catalogRepositoryProvider).listArtists(limit: 200);
});

class ArtistsScreen extends ConsumerWidget {
  const ArtistsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final state = ref.watch(_artistsProvider);
    return Scaffold(
      appBar: AppBar(title: Text(t.artists)),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.error_outline, color: NeonShrineColors.error, size: 48),
              const SizedBox(height: 12),
              Text('$e', textAlign: TextAlign.center),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: () => ref.refresh(_artistsProvider),
                child: Text(t.retry),
              ),
            ],
          ),
        ),
        data: (artists) => artists.isEmpty
            ? Center(child: Text(t.noData))
            : GridView.builder(
                padding: const EdgeInsets.all(12),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2,
                  crossAxisSpacing: 12,
                  mainAxisSpacing: 12,
                  childAspectRatio: 0.85,
                ),
                itemCount: artists.length,
                itemBuilder: (context, i) {
                  final artist = artists[i];
                  return _ArtistCard(
                    artist: artist,
                    onTap: () => context.go(AppRoutes.artistDetailPath(artist.id)),
                  );
                },
              ),
      ),
    );
  }
}

class _ArtistCard extends StatelessWidget {
  const _ArtistCard({required this.artist, required this.onTap});

  final CatalogArtist artist;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        decoration: BoxDecoration(
          color: NeonShrineColors.surfaceVariant,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: NeonShrineColors.outlineVariant, width: 0.5),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Expanded(
              child: Container(
                width: double.infinity,
                decoration: const BoxDecoration(
                  color: NeonShrineColors.surfaceContainer,
                  borderRadius: BorderRadius.vertical(top: Radius.circular(12)),
                ),
                child: const Icon(
                  Icons.person,
                  size: 56,
                  color: NeonShrineColors.outlineVariant,
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(8),
              child: Text(
                artist.name,
                style: Theme.of(context).textTheme.titleSmall,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                textAlign: TextAlign.center,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
