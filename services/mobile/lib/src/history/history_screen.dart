// ignore_for_file: implementation_imports, unnecessary_non_null_assertion
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:inori_api/src/model/batch_delete_request.dart';
import 'package:inori_api/src/model/play_event.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

final historyEventsProvider = FutureProvider<List<PlayEvent>>((ref) async {
  final api = ref.read(historyApiProvider);
  final resp = await api.listPlayEvents(limit: 200);
  return resp.data?.events ?? [];
});

// ---------------------------------------------------------------------------
// History Screen
// ---------------------------------------------------------------------------

class HistoryScreen extends ConsumerStatefulWidget {
  const HistoryScreen({super.key});

  @override
  ConsumerState<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends ConsumerState<HistoryScreen> {
  final Set<String> _selected = {};
  bool _selectMode = false;

  void _toggleSelect(String id) {
    setState(() {
      if (_selected.contains(id)) {
        _selected.remove(id);
        if (_selected.isEmpty) _selectMode = false;
      } else {
        _selected.add(id);
      }
    });
  }

  void _enterSelectMode(String id) {
    setState(() {
      _selectMode = true;
      _selected.add(id);
    });
  }

  Future<void> _deleteSelected() async {
    final ids = _selected.toList();
    if (ids.isEmpty) return;
    final t = AppLocalizations.of(context)!;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: NeonShrineColors.surfaceVariant,
        title: Text(t.deleteHistory, style: const TextStyle(color: NeonShrineColors.onSurface)),
        content: Text(
          'Delete ${ids.length} event${ids.length > 1 ? 's' : ''}?',
          style: const TextStyle(color: NeonShrineColors.onSurfaceVariant),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(t.cancel)),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: Text(t.delete)),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      final api = ref.read(historyApiProvider);
      // Batch delete in chunks of 100
      for (var i = 0; i < ids.length; i += 100) {
        final chunk = ids.skip(i).take(100).toList();
        await api.apiV1MeHistoryBatchDeletePost(
          batchDeleteRequest: BatchDeleteRequest(ids: chunk),
        );
      }
      setState(() {
        _selected.clear();
        _selectMode = false;
      });
      // ignore: unused_result
      ref.refresh(historyEventsProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to delete: $e'), backgroundColor: NeonShrineColors.error),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(historyEventsProvider);
    final t = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: _selectMode ? Text('${_selected.length} selected') : Text(t.history),
        actions: [
          if (_selectMode) ...[
            IconButton(
              icon: const Icon(Icons.delete, color: NeonShrineColors.error),
              tooltip: 'Delete selected',
              onPressed: _selected.isEmpty ? null : _deleteSelected,
            ),
            IconButton(
              icon: const Icon(Icons.close),
              tooltip: 'Cancel',
              onPressed: () => setState(() {
                _selected.clear();
                _selectMode = false;
              }),
            ),
          ] else ...[
            IconButton(
              icon: const Icon(Icons.bar_chart),
              tooltip: 'Statistics',
              onPressed: () => context.push(AppRoutes.historyStats),
            ),
            IconButton(
              icon: const Icon(Icons.refresh),
              tooltip: 'Refresh',
              onPressed: () => ref.refresh(historyEventsProvider),
            ),
          ],
        ],
      ),
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
                onPressed: () => ref.refresh(historyEventsProvider),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (events) => events.isEmpty
            ? const Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.history, size: 64, color: NeonShrineColors.onSurfaceVariant),
                    SizedBox(height: 16),
                    Text('No play history yet', style: TextStyle(fontSize: 18, color: NeonShrineColors.onSurfaceVariant)),
                  ],
                ),
              )
            : ListView.builder(
                itemCount: events.length,
                itemBuilder: (context, i) {
                  final event = events[i];
                  final isSelected = _selected.contains(event.id);

                  return ListTile(
                    leading: _selectMode
                        ? Checkbox(
                            value: isSelected,
                            onChanged: (_) => _toggleSelect(event.id),
                            checkColor: Colors.white,
                            activeColor: NeonShrineColors.primaryViolet,
                          )
                        : const CircleAvatar(
                            backgroundColor: NeonShrineColors.surfaceContainer,
                            child: Icon(Icons.music_note, color: NeonShrineColors.onSurfaceVariant, size: 18),
                          ),
                    title: Text(
                      event.trackId,
                      style: const TextStyle(color: NeonShrineColors.onSurface, fontSize: 14, fontWeight: FontWeight.w500),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    subtitle: Text(
                      _formatEventDate(event.playedAt, t),
                      style: const TextStyle(color: NeonShrineColors.onSurfaceVariant, fontSize: 12),
                    ),
                    selected: isSelected,
                    selectedTileColor: NeonShrineColors.primaryVioletDark.withValues(alpha: 0.2),
                    onTap: _selectMode ? () => _toggleSelect(event.id) : null,
                    onLongPress: _selectMode ? null : () => _enterSelectMode(event.id),
                  );
                },
              ),
      ),
    );
  }

  static String _formatEventDate(DateTime dt, AppLocalizations t) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    final hhmm = '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
    if (diff.inDays == 0) {
      return '${t.today} $hhmm';
    } else if (diff.inDays == 1) {
      return '${t.yesterday} $hhmm';
    }
    return '${dt.year}/${dt.month.toString().padLeft(2, '0')}/${dt.day.toString().padLeft(2, '0')}';
  }
}
