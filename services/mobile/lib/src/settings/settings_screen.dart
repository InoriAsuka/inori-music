import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
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

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
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
