import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/audio/crossfade_notifier.dart';
import 'package:inori_music/src/audio/eq_notifier.dart';
import 'package:inori_music/src/audio/replay_gain_notifier.dart';
import 'package:inori_music/src/audio/sleep_timer_notifier.dart';
import 'package:inori_music/src/audio/speed_notifier.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/lyrics/bilingual_lyrics_notifier.dart';
import 'package:inori_music/src/offline/download_notifier.dart';
import 'package:inori_music/src/offline/offline_db.dart';
import 'package:inori_music/src/shared/locale_provider.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/neon_shrine.dart';

// ---------------------------------------------------------------------------
// Language picker data
// ---------------------------------------------------------------------------

class _LangOption {
  const _LangOption(this.locale, this.label);
  final Locale? locale; // null = system default
  final String label;
}

const _langOptions = [
  _LangOption(null, 'System Default'),
  _LangOption(Locale('en'), 'English'),
  _LangOption(Locale('zh'), '简体中文'),
  _LangOption(Locale('ja'), '日本語'),
];

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  @override
  Widget build(BuildContext context) {
    final auth = ref.watch(authProvider);
    final username = auth.valueOrNull?.username ?? '';
    final locale = ref.read(localeProvider);
    final t = AppLocalizations.of(context);

    return Scaffold(
      appBar: AppBar(title: Text(t.settings)),
      body: ListView(
        children: [
          // Account section
          const _SectionHeader(title: 'Account'),
          ListTile(
            leading: const CircleAvatar(
              backgroundColor: NeonShrineColors.primaryVioletDark,
              child: Icon(Icons.person, color: Colors.white),
            ),
            title: Text(username.isNotEmpty ? username : 'User'),
            subtitle: const Text('Logged in'),
          ),
          ListTile(
            leading: const Icon(Icons.lock_outline),
            title: Text(t.changePassword),
            onTap: () => _showChangePasswordDialog(context, ref, t),
          ),
          const Divider(),

          // Sessions section
          const _SectionHeader(title: 'Sessions'),
          ListTile(
            leading: const Icon(Icons.devices),
            title: Text(t.revokeAll),
            subtitle: Text(t.sessions),
            onTap: () => _confirmRevokeAllSessions(context, ref, t),
          ),
          const Divider(),

          // Language section
          const _SectionHeader(title: 'Appearance'),
          ListTile(
            leading: const Icon(Icons.language),
            title: Text(t.language),
            subtitle: Text(_currentLabel(locale)),
            onTap: () => _showLanguagePicker(context, ref, t),
          ),
          const Divider(),

          // Offline Library section
          const _SectionHeader(title: 'Offline Library'),
          _OfflineLibrarySection(),
          const Divider(),

          // Lyrics section
          const _SectionHeader(title: '歌词'),
          Consumer(
            builder: (context, ref, _) {
              final enabled = ref.watch(bilingualLyricsProvider);
              return SwitchListTile(
                title: const Text('双语歌词'),
                subtitle: const Text('在原文下方显示翻译歌词'),
                value: enabled,
                onChanged: (v) => ref.read(bilingualLyricsProvider.notifier).setEnabled(v),
              );
            },
          ),
          const Divider(),

          // Audio section
          const _SectionHeader(title: '音频'),
          Consumer(
            builder: (context, ref, _) {
              final speed = ref.watch(speedNotifierProvider);
              return ListTile(
                title: const Text('播放速度'),
                trailing: Text('${speed}×'),
                onTap: () => _showSpeedSheet(context, ref),
              );
            },
          ),
          Consumer(
            builder: (context, ref, _) {
              final enabled = ref.watch(replayGainEnabledProvider);
              return SwitchListTile(
                title: const Text('响度归一化 (ReplayGain)'),
                subtitle: const Text('自动调整音量，使不同曲目听感一致'),
                value: enabled,
                onChanged: (v) => ref.read(replayGainEnabledProvider.notifier).setEnabled(v),
              );
            },
          ),
          // EQ section
          const _EqSection(),
          // Sleep timer
          Consumer(
            builder: (context, ref, _) {
              final timerState = ref.watch(sleepTimerProvider);
              final String subtitle;
              if (!timerState.active) {
                subtitle = '未激活';
              } else if (timerState.stopAfterTrack) {
                subtitle = '当前曲目结束后停止';
              } else if (timerState.remaining != null) {
                final m = timerState.remaining!.inMinutes
                    .remainder(60)
                    .toString()
                    .padLeft(2, '0');
                final s = timerState.remaining!.inSeconds
                    .remainder(60)
                    .toString()
                    .padLeft(2, '0');
                subtitle = '剩余 $m:$s';
              } else {
                subtitle = '激活';
              }
              return ListTile(
                leading: const Icon(Icons.bedtime),
                title: const Text('睡眠定时器'),
                subtitle: Text(subtitle),
                onTap: () => _showSleepTimerSheet(context, ref),
              );
            },
          ),
          // Crossfade slider
          Consumer(
            builder: (context, ref, _) {
              final seconds = ref.watch(crossfadeProvider);
              return ListTile(
                leading: const Icon(Icons.swap_horiz),
                title: const Text('交叉淡入淡出'),
                subtitle: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(seconds == 0 ? '关闭' : '$seconds 秒'),
                    Slider(
                      value: seconds.toDouble(),
                      min: 0,
                      max: 8,
                      divisions: 8,
                      label: seconds == 0 ? '关闭' : '$seconds 秒',
                      onChanged: (v) => ref
                          .read(crossfadeProvider.notifier)
                          .setSeconds(v.round()),
                    ),
                  ],
                ),
              );
            },
          ),
          const Divider(),

          // Sign out
          const _SectionHeader(title: 'Account Actions'),
          ListTile(
            leading: const Icon(Icons.logout, color: NeonShrineColors.error),
            title: Text(t.logout, style: const TextStyle(color: NeonShrineColors.error)),
            onTap: () => _confirmLogout(context, ref, t),
          ),
        ],
      ),
    );
  }

  void _showSleepTimerSheet(BuildContext context, WidgetRef ref) {
    final timerActive = ref.read(sleepTimerProvider).active;
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Padding(
              padding: EdgeInsets.all(16),
              child: Text('睡眠定时器',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            ),
            for (final mins in [15, 30, 45, 60])
              ListTile(
                title: Text('$mins 分钟'),
                onTap: () {
                  ref
                      .read(sleepTimerProvider.notifier)
                      .startFixed(Duration(minutes: mins));
                  Navigator.pop(context);
                },
              ),
            ListTile(
              title: const Text('当前曲目结束后停止'),
              onTap: () {
                ref.read(sleepTimerProvider.notifier).startAfterTrack();
                Navigator.pop(context);
              },
            ),
            if (timerActive)
              ListTile(
                title: const Text('取消定时器',
                    style: TextStyle(color: Colors.red)),
                onTap: () {
                  ref.read(sleepTimerProvider.notifier).cancel();
                  Navigator.pop(context);
                },
              ),
          ],
        ),
      ),
    );
  }

  void _showSpeedSheet(BuildContext context, WidgetRef ref) {
    const speeds = [0.5, 0.75, 1.0, 1.25, 1.5, 2.0];
    final current = ref.read(speedNotifierProvider);
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Padding(
              padding: EdgeInsets.all(16),
              child: Text('播放速度', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            ),
            for (final s in speeds)
              ListTile(
                title: Text('${s}×'),
                trailing: s == current ? const Icon(Icons.check) : null,
                onTap: () {
                  ref.read(speedNotifierProvider.notifier).setSpeed(s);
                  Navigator.pop(context);
                },
              ),
          ],
        ),
      ),
    );
  }

  String _currentLabel(Locale? locale) {
    for (final opt in _langOptions) {
      if (_isSameLocale(opt.locale, locale)) return opt.label;
    }
    return locale?.toLanguageTag() ?? 'System Default';
  }

  Future<void> _showChangePasswordDialog(
    BuildContext context,
    WidgetRef ref,
    AppLocalizations t,
  ) async {
    final currentCtrl = TextEditingController();
    final newCtrl = TextEditingController();
    final formKey = GlobalKey<FormState>();

    await showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(t.changePassword),
        content: Form(
          key: formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(
                controller: currentCtrl,
                decoration: InputDecoration(labelText: t.currentPassword),
                obscureText: true,
                validator: (v) => (v == null || v.isEmpty) ? t.fieldRequired : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: newCtrl,
                decoration: InputDecoration(labelText: t.newPassword),
                obscureText: true,
                validator: (v) {
                  if (v == null || v.isEmpty) return t.fieldRequired;
                  if (v.length < 8) return t.passwordMinLength;
                  return null;
                },
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text(t.cancel),
          ),
          FilledButton(
            onPressed: () async {
              if (!formKey.currentState!.validate()) return;
              try {
                await ref
                    .read(authProvider.notifier)
                    .changePassword(currentCtrl.text, newCtrl.text);
                if (ctx.mounted) {
                  Navigator.pop(ctx);
                  ScaffoldMessenger.of(ctx).showSnackBar(
                    SnackBar(content: Text(t.changePasswordSuccess)),
                  );
                }
              } catch (e) {
                if (ctx.mounted) {
                  ScaffoldMessenger.of(ctx).showSnackBar(
                    SnackBar(content: Text('Error: $e')),
                  );
                }
              }
            },
            child: Text(t.save),
          ),
        ],
      ),
    );
    currentCtrl.dispose();
    newCtrl.dispose();
  }

  Future<void> _confirmRevokeAllSessions(
    BuildContext context,
    WidgetRef ref,
    AppLocalizations t,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(t.revokeAll),
        content: Text(t.revokeAllConfirmBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(t.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(t.revokeAll),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(authProvider.notifier).revokeAllOtherSessions();
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(t.revokeAllConfirmBody)),
        );
      }
    }
  }

  Future<void> _showLanguagePicker(
    BuildContext context,
    WidgetRef ref,
    AppLocalizations t,
  ) async {
    final currentLocale = ref.read(localeProvider);

    final selected = await showModalBottomSheet<int>(
      context: context,
      backgroundColor: NeonShrineColors.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Text(
                t.language,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
            ),
            const Divider(height: 1),
            ..._langOptions.asMap().entries.map(
              (entry) {
                final i = entry.key;
                final opt = entry.value;
                final isSelected = _isSameLocale(opt.locale, currentLocale);
                return ListTile(
                  title: Text(opt.label),
                  leading: isSelected
                      ? const Icon(Icons.check, color: NeonShrineColors.primaryViolet)
                      : null,
                  onTap: () => Navigator.pop(ctx, i),
                );
              },
            ),
          ],
        ),
      ),
    );

    if (selected != null && context.mounted) {
      ref.read(localeProvider.notifier).state = _langOptions[selected].locale;
    }
  }

  static bool _isSameLocale(Locale? a, Locale? b) {
    if (a == null && b == null) return true;
    if (a == null || b == null) return false;
    return a.languageCode == b.languageCode;
  }

  Future<void> _confirmLogout(
    BuildContext context,
    WidgetRef ref,
    AppLocalizations t,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(t.logout),
        content: Text(t.signOutConfirmBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(t.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(t.logout),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(authProvider.notifier).logout();
      if (context.mounted) context.go(AppRoutes.login);
    }
  }
}

class _OfflineLibrarySection extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return FutureBuilder<List<OfflineTrack>>(
      future: OfflineDb.instance.queryAll(),
      builder: (context, snapshot) {
        final tracks = snapshot.data ?? [];
        if (tracks.isEmpty) {
          return const Padding(
            padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Text(
              'No downloaded tracks.',
              style: TextStyle(color: NeonShrineColors.onSurfaceVariant),
            ),
          );
        }
        final totalBytes = tracks.fold<int>(0, (sum, t) => sum + t.sizeBytes);
        final totalMb = (totalBytes / (1024 * 1024)).toStringAsFixed(1);
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
              child: Text(
                '${tracks.length} track${tracks.length == 1 ? '' : 's'} · $totalMb MB',
                style: const TextStyle(
                  color: NeonShrineColors.onSurfaceVariant,
                  fontSize: 12,
                ),
              ),
            ),
            ...tracks.map(
              (t) => ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 16),
                leading: const Icon(Icons.download_done,
                    color: NeonShrineColors.primaryViolet),
                title: Text(
                  t.title,
                  style: const TextStyle(
                      color: NeonShrineColors.onSurface, fontSize: 14),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                subtitle: Text(
                  t.artistName,
                  style: const TextStyle(
                      color: NeonShrineColors.onSurfaceVariant, fontSize: 12),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                trailing: IconButton(
                  icon: const Icon(Icons.delete_outline,
                      color: NeonShrineColors.error),
                  onPressed: () {
                    ref
                        .read(downloadProvider.notifier)
                        .deleteDownload(t.trackId);
                  },
                ),
              ),
            ),
            ListTile(
              leading:
                  const Icon(Icons.delete_sweep, color: NeonShrineColors.error),
              title: const Text(
                'Delete all downloads',
                style: TextStyle(color: NeonShrineColors.error),
              ),
              onTap: () async {
                final confirmed = await showDialog<bool>(
                  context: context,
                  builder: (ctx) => AlertDialog(
                    title: const Text('Delete all downloads'),
                    content: const Text(
                        'This will remove all offline tracks from your device.'),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.pop(ctx, false),
                        child: const Text('Cancel'),
                      ),
                      FilledButton(
                        onPressed: () => Navigator.pop(ctx, true),
                        child: const Text('Delete all'),
                      ),
                    ],
                  ),
                );
                if (confirmed == true) {
                  await ref
                      .read(downloadProvider.notifier)
                      .deleteAllDownloads();
                }
              },
            ),
          ],
        );
      },
    );
  }
}

/// EQ 10-band equalizer section for the Settings screen.
class _EqSection extends ConsumerWidget {
  const _EqSection();

  static const _presetKeys = ['flat', 'bassBoost', 'vocal', 'electronic'];
  static const _presetLabels = ['Flat', 'Bass', 'Vocal', 'Electronic'];
  static const _bandLabels = ['31', '62', '125', '250', '500', '1K', '2K', '4K', '8K', '16K'];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final eq = ref.watch(eqNotifierProvider);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Consumer(
          builder: (context, ref, _) {
            final enabled = ref.watch(eqNotifierProvider).enabled;
            return SwitchListTile(
              title: const Text('均衡器 (EQ)'),
              subtitle: const Text('10 频段均衡器'),
              value: enabled,
              onChanged: (v) => ref.read(eqNotifierProvider.notifier).setEnabled(v),
            );
          },
        ),
        if (eq.enabled) ...[
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
            child: SegmentedButton<String>(
              segments: List.generate(
                _presetKeys.length,
                (i) => ButtonSegment<String>(
                  value: _presetKeys[i],
                  label: Text(_presetLabels[i], style: const TextStyle(fontSize: 12)),
                ),
              ),
              selected: {eq.preset.isEmpty || !_presetKeys.contains(eq.preset) ? 'flat' : eq.preset},
              onSelectionChanged: (s) =>
                  ref.read(eqNotifierProvider.notifier).setPreset(s.first),
              multiSelectionEnabled: false,
            ),
          ),
          const SizedBox(height: 8),
          SizedBox(
            height: 160,
            child: Row(
              children: List.generate(10, (i) {
                return Expanded(
                  child: Column(
                    children: [
                      Expanded(
                        child: RotatedBox(
                          quarterTurns: 3,
                          child: Slider(
                            value: eq.bands[i].clamp(-12.0, 12.0),
                            min: -12.0,
                            max: 12.0,
                            divisions: 48,
                            onChanged: (v) =>
                                ref.read(eqNotifierProvider.notifier).setBand(i, v),
                          ),
                        ),
                      ),
                      Text(
                        _bandLabels[i],
                        style: const TextStyle(fontSize: 10, color: NeonShrineColors.onSurfaceVariant),
                      ),
                    ],
                  ),
                );
              }),
            ),
          ),
          const SizedBox(height: 4),
        ],
      ],
    );
  }
}

class _SectionHeader extends StatelessWidget {
  const _SectionHeader({required this.title});
  final String title;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
      child: Text(
        title,
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              color: NeonShrineColors.primaryViolet,
              fontWeight: FontWeight.w600,
            ),
      ),
    );
  }
}
