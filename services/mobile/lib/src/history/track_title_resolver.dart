// ignore_for_file: implementation_imports
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';

/// Resolves a trackId → track title. Auto-disposes when no widget is watching.
///
/// Usage:
///   // Trigger resolution once in initState:
///   ref.read(trackTitleResolverProvider(trackId).notifier).resolve();
///   // Watch the result reactively:
///   final title = ref.watch(trackTitleResolverProvider(trackId));
class TrackTitleResolver extends AutoDisposeFamilyNotifier<String?, String> {
  @override
  String? build(String trackId) => null;

  Future<void> resolve() async {
    if (state != null) return; // already resolved
    try {
      final track = await ref.read(catalogRepositoryProvider).getTrack(arg);
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
