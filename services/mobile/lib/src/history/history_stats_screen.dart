// ignore_for_file: implementation_imports, unnecessary_non_null_assertion
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/timeline_bucket.dart';
import 'package:inori_api/src/model/track_play_count.dart';
import 'package:inori_api/src/model/user_history_stats.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/history/track_title_resolver.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

final historyStatsProvider = FutureProvider<UserHistoryStats>((ref) async {
  final api = ref.read(historyApiProvider);
  final resp = await api.apiV1MeHistoryStatsGet();
  return resp.data!;
});

final historyTimelineProvider = FutureProvider<List<TimelineBucket>>((ref) async {
  final api = ref.read(historyApiProvider);
  // Last 30 days
  final until = DateTime.now();
  final since = until.subtract(const Duration(days: 30));
  final resp = await api.getMyHistoryTimeline(since: since, until: until, granularity: 'day');
  return resp.data?.buckets ?? [];
});

final historyTopTracksProvider = FutureProvider<List<TrackPlayCount>>((ref) async {
  final api = ref.read(historyApiProvider);
  final resp = await api.apiV1MeHistoryTopTracksGet(limit: 5);
  return resp.data?.tracks ?? [];
});

// ---------------------------------------------------------------------------
// History Stats Screen
// ---------------------------------------------------------------------------

class HistoryStatsScreen extends ConsumerWidget {
  const HistoryStatsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statsState = ref.watch(historyStatsProvider);
    final timelineState = ref.watch(historyTimelineProvider);
    final topTracksState = ref.watch(historyTopTracksProvider);
    final t = AppLocalizations.of(context);

    return Scaffold(
      appBar: AppBar(title: Text(t.historyStats)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Summary cards
          statsState.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Text('$e', style: const TextStyle(color: NeonShrineColors.error)),
            data: (stats) => Row(
              children: [
                Expanded(child: _StatCard(label: t.totalPlays, value: '${stats.totalEvents}')),
                const SizedBox(width: 12),
                Expanded(child: _StatCard(label: t.uniqueTracks, value: '${stats.uniqueTracks}')),
              ],
            ),
          ),

          const SizedBox(height: 24),

          // 30-day chart
          Text(
            t.activityChart,
            style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: NeonShrineColors.onBackground),
          ),
          const SizedBox(height: 12),
          timelineState.when(
            loading: () => const SizedBox(height: 150, child: Center(child: CircularProgressIndicator())),
            error: (e, _) => SizedBox(height: 80, child: Center(child: Text('$e', style: const TextStyle(color: NeonShrineColors.error)))),
            data: (buckets) => buckets.isEmpty
                ? const SizedBox(height: 80, child: Center(child: Text('No data', style: TextStyle(color: NeonShrineColors.onSurfaceVariant))))
                : SizedBox(
                    height: 160,
                    child: BarChart(
                      BarChartData(
                        backgroundColor: Colors.transparent,
                        borderData: FlBorderData(show: false),
                        gridData: FlGridData(show: false),
                        titlesData: FlTitlesData(
                          show: true,
                          leftTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
                          rightTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
                          topTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
                          bottomTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
                        ),
                        barGroups: buckets.asMap().entries.map((entry) {
                          return BarChartGroupData(
                            x: entry.key,
                            barRods: [
                              BarChartRodData(
                                toY: entry.value.eventCount.toDouble(),
                                color: NeonShrineColors.primaryViolet,
                                width: 6,
                                borderRadius: BorderRadius.circular(3),
                              ),
                            ],
                          );
                        }).toList(),
                        barTouchData: BarTouchData(enabled: false),
                      ),
                    ),
                  ),
          ),

          const SizedBox(height: 24),

          // Top tracks
          Text(
            t.topTracks,
            style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: NeonShrineColors.onBackground),
          ),
          const SizedBox(height: 8),
          topTracksState.when(
            loading: () => const SizedBox(height: 80, child: Center(child: CircularProgressIndicator())),
            error: (e, _) => Text('$e', style: const TextStyle(color: NeonShrineColors.error)),
            data: (tracks) => tracks.isEmpty
                ? const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Text('No history yet', style: TextStyle(color: NeonShrineColors.onSurfaceVariant)),
                  )
                : Column(
                    children: tracks.asMap().entries.map((entry) {
                      final rank = entry.key + 1;
                      final tc = entry.value;
                      return _TopTrackRow(key: ValueKey(tc.trackId), rank: rank, trackId: tc.trackId, playCount: tc.playCount);
                    }).toList(),
                  ),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Widgets
// ---------------------------------------------------------------------------

class _StatCard extends StatelessWidget {
  const _StatCard({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: NeonShrineColors.surfaceVariant,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: NeonShrineColors.outlineVariant, width: 0.5),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(value, style: const TextStyle(fontSize: 28, fontWeight: FontWeight.w700, color: NeonShrineColors.primaryVioletLight)),
          const SizedBox(height: 4),
          Text(label, style: const TextStyle(fontSize: 13, color: NeonShrineColors.onSurfaceVariant)),
        ],
      ),
    );
  }
}

class _TopTrackRow extends ConsumerStatefulWidget {
  const _TopTrackRow({super.key, required this.rank, required this.trackId, required this.playCount});

  final int rank;
  final String trackId;
  final int playCount;

  @override
  ConsumerState<_TopTrackRow> createState() => _TopTrackRowState();
}

class _TopTrackRowState extends ConsumerState<_TopTrackRow> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        ref.read(trackTitleResolverProvider(widget.trackId).notifier).resolve();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final title = ref.watch(trackTitleResolverProvider(widget.trackId));
    final displayName = title ?? widget.trackId;

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: NeonShrineColors.surfaceVariant,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          SizedBox(
            width: 28,
            child: Text(
              '#${widget.rank}',
              style: const TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w700,
                color: NeonShrineColors.primaryVioletLight,
              ),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              displayName,
              style: const TextStyle(
                fontSize: 14,
                color: NeonShrineColors.onSurface,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 8),
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.play_arrow, size: 14, color: NeonShrineColors.onSurfaceVariant),
              const SizedBox(width: 2),
              Text(
                '${widget.playCount}',
                style: const TextStyle(fontSize: 13, color: NeonShrineColors.onSurfaceVariant),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
