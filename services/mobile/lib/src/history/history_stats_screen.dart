// ignore_for_file: implementation_imports
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:inori_api/src/model/timeline_bucket.dart';
import 'package:inori_api/src/model/user_history_stats.dart';

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

// ---------------------------------------------------------------------------
// History Stats Screen
// ---------------------------------------------------------------------------

class HistoryStatsScreen extends ConsumerWidget {
  const HistoryStatsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statsState = ref.watch(historyStatsProvider);
    final timelineState = ref.watch(historyTimelineProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('History Stats')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Summary cards
          statsState.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Text('$e', style: const TextStyle(color: NeonShrineColors.error)),
            data: (stats) => Row(
              children: [
                Expanded(child: _StatCard(label: 'Total Plays', value: '${stats.totalEvents}')),
                const SizedBox(width: 12),
                Expanded(child: _StatCard(label: 'Unique Tracks', value: '${stats.uniqueTracks}')),
              ],
            ),
          ),

          const SizedBox(height: 24),

          // 30-day chart
          const Text(
            '30-day Activity',
            style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: NeonShrineColors.onBackground),
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
        ],
      ),
    );
  }
}

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
