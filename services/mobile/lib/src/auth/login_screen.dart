import 'dart:ui' show ImageFilter;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/shared/theme/sakura_dusk.dart';
import 'package:inori_music/src/shared/widgets/app_background.dart';
import 'package:inori_music/src/shared/widgets/inori_mark.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _usernameCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  final _serverCtrl = TextEditingController(text: 'http://localhost:8080');

  bool _obscurePassword = true;
  bool _showServerField = false;

  @override
  void dispose() {
    _usernameCtrl.dispose();
    _passwordCtrl.dispose();
    _serverCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    await ref.read(authProvider.notifier).login(
          _usernameCtrl.text.trim(),
          _passwordCtrl.text,
          baseUrl: _showServerField ? _serverCtrl.text.trim() : null,
        );
  }

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    final auth = ref.watch(authProvider);
    final isLoading = auth is AsyncLoading;
    final error = auth.valueOrNull?.error;

    return Scaffold(
      body: AppBackground(
        child: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 32),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  // Brand mark floats directly over the background, same
                  // treatment as the splash screen it follows.
                  const InoriMark(size: 76),
                  const SizedBox(height: 16),
                  Text(
                    'Inori Music',
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                          color: SakuraDuskColors.onBackground,
                          fontWeight: FontWeight.w700,
                        ),
                  ),
                  const SizedBox(height: 40),

                  _GlassCard(
                    child: Form(
                      key: _formKey,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Text(
                            'Sign in to your library',
                            textAlign: TextAlign.center,
                            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                                  color: SakuraDuskColors.onSurfaceVariant,
                                ),
                          ),
                          const SizedBox(height: 20),

                          _ServerUrlToggle(
                            show: _showServerField,
                            controller: _serverCtrl,
                            onToggle: () => setState(() => _showServerField = !_showServerField),
                          ),
                          if (_showServerField) const SizedBox(height: 16),

                          TextFormField(
                            controller: _usernameCtrl,
                            decoration: InputDecoration(
                              labelText: t.username,
                              prefixIcon: const Icon(Icons.person_outline),
                            ),
                            textInputAction: TextInputAction.next,
                            autofillHints: const [AutofillHints.username],
                            validator: (v) => (v == null || v.trim().isEmpty) ? t.fieldRequired : null,
                          ),
                          const SizedBox(height: 16),

                          TextFormField(
                            controller: _passwordCtrl,
                            decoration: InputDecoration(
                              labelText: t.password,
                              prefixIcon: const Icon(Icons.lock_outline),
                              suffixIcon: IconButton(
                                icon: Icon(
                                  _obscurePassword ? Icons.visibility_off_outlined : Icons.visibility_outlined,
                                ),
                                onPressed: () => setState(() => _obscurePassword = !_obscurePassword),
                              ),
                            ),
                            obscureText: _obscurePassword,
                            textInputAction: TextInputAction.done,
                            autofillHints: const [AutofillHints.password],
                            onFieldSubmitted: (_) => _submit(),
                            validator: (v) => (v == null || v.isEmpty) ? t.fieldRequired : null,
                          ),
                          const SizedBox(height: 8),

                          if (error != null)
                            Padding(
                              padding: const EdgeInsets.only(top: 4, bottom: 4),
                              child: Text(
                                error,
                                style: TextStyle(
                                  color: Theme.of(context).colorScheme.error,
                                  fontSize: 13,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                          const SizedBox(height: 24),

                          FilledButton(
                            onPressed: isLoading ? null : _submit,
                            child: isLoading
                                ? const SizedBox(
                                    height: 20,
                                    width: 20,
                                    child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                                  )
                                : Text(t.login, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                          ),
                          const SizedBox(height: 12),

                          // Guest entry point — a local-files player, no
                          // account needed.
                          TextButton(
                            onPressed: isLoading
                                ? null
                                : () => ref.read(authProvider.notifier).continueAsGuest(),
                            child: const Text('以游客身份继续'),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Sub-widgets
// ---------------------------------------------------------------------------

/// Frosted card the login form floats on. Must stay readable over an
/// arbitrary user-picked background image: a strong blur plus a
/// near-opaque (72%) surface tint neutralizes most of what's behind it
/// before any text renders on top, so the existing Sakura Dusk contrast
/// ratios (audited against a flat surface) still hold in practice.
class _GlassCard extends StatelessWidget {
  const _GlassCard({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 400),
      child: Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(24),
          boxShadow: [
            BoxShadow(
              color: SakuraDuskColors.miniPlayerShadow,
              blurRadius: 32,
              offset: const Offset(0, 12),
            ),
          ],
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(24),
          child: BackdropFilter(
            filter: ImageFilter.blur(sigmaX: 24, sigmaY: 24),
            child: Container(
              padding: const EdgeInsets.fromLTRB(24, 28, 24, 24),
              decoration: BoxDecoration(
                color: SakuraDuskColors.surface.withValues(alpha: 0.72),
                borderRadius: BorderRadius.circular(24),
                border: Border.all(
                  color: SakuraDuskColors.outlineVariant.withValues(alpha: 0.6),
                ),
              ),
              child: child,
            ),
          ),
        ),
      ),
    );
  }
}

class _ServerUrlToggle extends ConsumerStatefulWidget {
  const _ServerUrlToggle({
    required this.show,
    required this.controller,
    required this.onToggle,
  });

  final bool show;
  final TextEditingController controller;
  final VoidCallback onToggle;

  @override
  ConsumerState<_ServerUrlToggle> createState() => _ServerUrlToggleState();
}

class _ServerUrlToggleState extends ConsumerState<_ServerUrlToggle> {
  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        GestureDetector(
          onTap: widget.onToggle,
          child: Row(
            children: [
              Icon(
                widget.show ? Icons.expand_less : Icons.expand_more,
                size: 18,
                color: SakuraDuskColors.onSurfaceVariant,
              ),
              const SizedBox(width: 4),
              Text(
                'Server URL',
                style: Theme.of(context).textTheme.labelMedium,
              ),
            ],
          ),
        ),
        if (widget.show) ...[
          const SizedBox(height: 8),
          TextFormField(
            controller: widget.controller,
            decoration: InputDecoration(
              labelText: t.serverUrl,
              hintText: 'http://localhost:8080',
              prefixIcon: const Icon(Icons.dns_outlined),
            ),
            keyboardType: TextInputType.url,
            textInputAction: TextInputAction.next,
            validator: (v) {
              if (v == null || v.trim().isEmpty) return t.fieldRequired;
              final uri = Uri.tryParse(v.trim());
              if (uri == null || !uri.hasScheme) return 'Invalid URL';
              return null;
            },
          ),
        ],
      ],
    );
  }
}
