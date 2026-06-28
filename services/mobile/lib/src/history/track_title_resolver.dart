// ignore_for_file: implementation_imports
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';

/// Resolves a trackId → track title. Stores the trackId as its own state
/// and provides a `resolve()` method that fetches and caches the title.
///
/// Usage:
///   // Trigger resolution once in initState:
///   ref.read(trackTitleResolverProvider(trackId).notifier).resolve();
///   // Watch the result reactively:
///   final title = ref.watch(trackTitleResolverProvider(trackId));
class TrackTitleResolver extends AutoDisposeFamilyNotifier<String?, String> {
  @override
  String? build(String trackId) => trackId;

  Future<void> resolve() async {
    final trackId = state;
    if (trackId == null) return;
    try {
      final track = await ref.read(catalogRepositoryProvider).getTrack(trackId);
      state = track.title;
    } catch (_) {
      // Leave state as null — caller falls back to trackId
    }
  }
}

final trackTitleResolverProvider =
    NotifierProvider.autoDispose.family<TrackTitleResolver, String?, String>(
  TrackTitleResolver.new,
);
