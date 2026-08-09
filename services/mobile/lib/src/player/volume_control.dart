import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';

/// Remembers the volume level just before the user last hit mute, so
/// unmuting restores that value instead of jumping to a hardcoded default.
///
/// Deliberately a plain Riverpod [StateProvider] rather than local [State] on
/// [VolumeControl] itself: the desktop mini player bar and the full player
/// screen each mount their own, independent [VolumeControl] instance (see
/// `mini_player_bar.dart` / `full_player_screen.dart`), and muting from one
/// is expected to unmute correctly from the other — two separate local
/// memories would drift the moment the user touched both. Not persisted:
/// losing this across app restarts is fine, since [PlayerState.volume]
/// itself always initialises at 1.0 rather than restoring a prior session's
/// level either.
final volumeBeforeMuteProvider = StateProvider<double>((ref) => 1.0);

/// Speaker icon + horizontal slider wired straight to
/// [PlayerNotifier.setVolume] / [PlayerState.volume] — the setter has existed
/// since before v5.30.0 but nothing in the UI ever called it until this.
///
/// Shared by the desktop mini player bar's control row and the full player
/// screen's transport row (see each file's doc comments) rather than
/// duplicated, so the two can't drift into different mute/step behaviour.
///
/// [compact] drops the inline slider in favour of a bare icon button that
/// opens a small popover housing the same slider — the mini bar reaches for
/// this once the row segment it hands the control isn't wide enough for the
/// full inline layout (see `mini_player_bar.dart`'s `_desktopActionsRow`),
/// and the full player screen uses it unconditionally: that screen's control
/// block width is a carefully calibrated, cover-derived dimension (see
/// `playerControlWidth`'s doc comment for the overflow history behind that
/// calibration), and an inline slider is exactly the kind of fixed-width
/// addition that has blown past it before.
class VolumeControl extends ConsumerWidget {
  const VolumeControl({super.key, this.compact = false, this.iconSize});

  final bool compact;
  final double? iconSize;

  /// Width of the inline slider track. Short enough to read as a secondary
  /// control next to the transport trio rather than competing with the main
  /// seek bar for attention.
  static const _sliderWidth = 90.0;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final volume = ref.watch(playerProvider.select((s) => s.volume));

    if (compact) {
      return IconButton(
        icon: Icon(_iconFor(volume), size: iconSize),
        color: context.skinColors.onSurfaceVariant,
        tooltip: 'Volume',
        onPressed: () => _showPopover(context),
      );
    }
    return _InlineVolumeControl(volume: volume, iconSize: iconSize);
  }

  static IconData _iconFor(double volume) {
    if (volume <= 0.0) return Icons.volume_off;
    if (volume < 0.5) return Icons.volume_down;
    return Icons.volume_up;
  }

  /// Anchors a [showMenu] popover to this button's own on-screen position —
  /// the standard Flutter recipe for a menu that opens next to whatever was
  /// tapped rather than at a fixed screen location. The menu item is disabled
  /// (no ink/selection behaviour of its own) since it exists purely to host
  /// the live slider, not to act as a choosable entry.
  void _showPopover(BuildContext context) {
    final button = context.findRenderObject()! as RenderBox;
    final overlay =
        Overlay.of(context).context.findRenderObject()! as RenderBox;
    final position = RelativeRect.fromRect(
      Rect.fromPoints(
        button.localToGlobal(Offset.zero, ancestor: overlay),
        button.localToGlobal(
          button.size.bottomRight(Offset.zero),
          ancestor: overlay,
        ),
      ),
      Offset.zero & overlay.size,
    );

    showMenu<void>(
      context: context,
      position: position,
      items: [
        PopupMenuItem<void>(
          enabled: false,
          child: Consumer(
            builder: (context, ref, _) {
              final volume = ref.watch(playerProvider.select((s) => s.volume));
              return _InlineVolumeControl(volume: volume, iconSize: iconSize);
            },
          ),
        ),
      ],
    );
  }
}

/// The icon-button-plus-slider row shared by [VolumeControl]'s expanded mode
/// and the popover its compact mode opens — one implementation of the actual
/// control, just rendered in two different places.
class _InlineVolumeControl extends ConsumerWidget {
  const _InlineVolumeControl({required this.volume, this.iconSize});

  final double volume;
  final double? iconSize;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        IconButton(
          icon: Icon(VolumeControl._iconFor(volume), size: iconSize),
          color: context.skinColors.onSurfaceVariant,
          tooltip: volume > 0 ? 'Mute' : 'Unmute',
          onPressed: () => _toggleMute(ref),
        ),
        SizedBox(
          width: VolumeControl._sliderWidth,
          child: SliderTheme(
            data: SliderTheme.of(context).copyWith(
              trackHeight: 2,
              activeTrackColor: context.skinColors.sakuraPink,
              inactiveTrackColor: context.skinColors.outline,
              thumbColor: context.skinColors.sakuraPinkLight,
              thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 4),
              overlayShape: const RoundSliderOverlayShape(overlayRadius: 10),
            ),
            child: Slider(
              value: volume.clamp(0.0, 1.0),
              onChanged: (v) => ref.read(playerProvider.notifier).setVolume(v),
            ),
          ),
        ),
      ],
    );
  }

  /// Mutes by remembering the current volume and dropping to zero; unmutes
  /// by restoring whatever was last remembered (1.0 if the user has never
  /// muted through this control this session — see
  /// [volumeBeforeMuteProvider]'s doc comment).
  void _toggleMute(WidgetRef ref) {
    final notifier = ref.read(playerProvider.notifier);
    if (volume > 0) {
      ref.read(volumeBeforeMuteProvider.notifier).state = volume;
      notifier.setVolume(0);
    } else {
      final restore = ref.read(volumeBeforeMuteProvider);
      notifier.setVolume(restore > 0 ? restore : 1.0);
    }
  }
}
