// ignore_for_file: implementation_imports
//
// mini_player_bar_test.dart
//
// Widget tests for MiniPlayerBar.
//
// Strategy: override playerProvider with a stub that subclasses PlayerNotifier
// and overrides build() to return a fixed PlayerState without touching
// audio_service or just_audio.
//
import 'package:audio_service/audio_service.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/player/mini_player_bar.dart';

/// Stubs the artwork URL lookup so a test can drive the "server track has a
/// resolvable cover" branch without a network round trip — same shape as
/// full_player_layout_test.dart's own _StubArtworkNotifier.
class _StubArtworkNotifier extends ArtworkUrlNotifier {
  _StubArtworkNotifier(this._url);
  final String? _url;

  @override
  Future<String?> build(String albumId) async => _url;
}

// ---------------------------------------------------------------------------
// Stub PlayerNotifier — subclasses the real one but overrides build() to
// return a pre-built state and never initialise the audio subsystem.
// ---------------------------------------------------------------------------
class _StubPlayerNotifier extends PlayerNotifier {
  _StubPlayerNotifier(this._fixedState);
  final pstate.PlayerState _fixedState;

  int toggleCount = 0;

  @override
  pstate.PlayerState build() => _fixedState;

  @override
  Future<void> togglePlayPause() async {
    toggleCount++;
    state = state.copyWith(
      playbackState: PlaybackState(playing: !state.isPlaying),
    );
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

Widget _buildApp(
  _StubPlayerNotifier stub, {
  List<Override> extraOverrides = const [],
}) {
  return ProviderScope(
    overrides: [playerProvider.overrideWith(() => stub), ...extraOverrides],
    child: MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: const Scaffold(body: MiniPlayerBar()),
    ),
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  testWidgets('MiniPlayerBar shows nothingPlaying text when no media item', (
    tester,
  ) async {
    final stub = _StubPlayerNotifier(pstate.PlayerState());
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    // No track data should appear when there is no media item.
    expect(find.text('Idol'), findsNothing);
    expect(find.text('Yoasobi'), findsNothing);
    // Controls are always rendered.
    expect(
      find.byIcon(Icons.play_arrow_rounded),
      findsOneWidget,
      reason: 'Play icon shown when not playing and no media item',
    );
  });

  testWidgets('MiniPlayerBar shows track title when media item is set', (
    tester,
  ) async {
    const trackTitle = 'Idol';
    const artistName = 'Yoasobi';
    final mediaItem = MediaItem(
      id: 'track-001',
      title: trackTitle,
      artist: artistName,
    );
    final stub = _StubPlayerNotifier(
      pstate.PlayerState(
        queue: [mediaItem],
        currentIndex: 0,
        mediaItem: mediaItem,
        playbackState: PlaybackState(playing: false),
      ),
    );

    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    expect(
      find.text(trackTitle),
      findsOneWidget,
      reason: 'Track title should be displayed in the mini player bar',
    );
    expect(
      find.text(artistName),
      findsOneWidget,
      reason: 'Artist name should be displayed below the title',
    );
  });

  testWidgets('Play button calls togglePlayPause on the notifier', (
    tester,
  ) async {
    final mediaItem = MediaItem(id: 'track-001', title: 'Idol');
    final stub = _StubPlayerNotifier(
      pstate.PlayerState(
        queue: [mediaItem],
        currentIndex: 0,
        mediaItem: mediaItem,
        playbackState: PlaybackState(playing: false),
      ),
    );

    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    final playButton = find.byTooltip('Play');
    expect(
      playButton,
      findsOneWidget,
      reason: 'Play button should be present when not playing',
    );

    await tester.tap(playButton);
    await tester.pump();

    expect(
      stub.toggleCount,
      equals(1),
      reason: 'togglePlayPause should have been called once',
    );
  });

  // ---------------------------------------------------------------------
  // v5.30.0 — EchoMusic scale (requirement.md v5.30.0 / plan Phase v5.30.0)
  // ---------------------------------------------------------------------

  testWidgets('the content row is 84px tall, EchoMusic scale', (tester) async {
    final stub = _StubPlayerNotifier(pstate.PlayerState());
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    final size = tester.getSize(find.byKey(MiniPlayerBar.contentKey));
    expect(
      size.height,
      84,
      reason:
          'EchoMusic\'s PlayerBar.vue footer (h-21) is a fixed 84px; Apple '
          "Music's thinner bar was explicitly rejected as too small once "
          'the player page itself was reworked to match Apple Music.',
    );
  });

  testWidgets('the artwork is 56px, EchoMusic scale (was 44px)', (
    tester,
  ) async {
    final mediaItem = MediaItem(id: 'track-001', title: 'Idol');
    final stub = _StubPlayerNotifier(
      pstate.PlayerState(
        queue: [mediaItem],
        currentIndex: 0,
        mediaItem: mediaItem,
        playbackState: PlaybackState(playing: false),
      ),
    );
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    // No albumId on this mediaItem, so the artwork renders its fallback
    // music-note icon inside the cover box — the box itself is what v5.30.0
    // actually specifies, regardless of what's drawn inside it.
    final box = tester.widget<Container>(
      find.ancestor(
        of: find.byIcon(Icons.music_note),
        matching: find.byType(Container),
      ),
    );
    expect(box.constraints?.maxWidth, 56);
    expect(box.constraints?.maxHeight, 56);
  });

  testWidgets('transport icons render at EchoMusic\'s 22px, not the old 24px', (
    tester,
  ) async {
    final mediaItem = MediaItem(id: 'track-001', title: 'Idol');
    final stub = _StubPlayerNotifier(
      pstate.PlayerState(
        queue: [mediaItem],
        currentIndex: 0,
        mediaItem: mediaItem,
        playbackState: PlaybackState(playing: false),
      ),
    );
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    final prevIcon = tester.widget<Icon>(find.byIcon(Icons.skip_previous));
    final nextIcon = tester.widget<Icon>(find.byIcon(Icons.skip_next));
    expect(prevIcon.size, 22);
    expect(nextIcon.size, 22);
  });

  testWidgets(
    'the transport trio stays centred regardless of which side has the '
    'sleep timer button',
    (tester) async {
      // Before v5.30.0 the sleep timer button sat directly after "next" with
      // nothing to balance it on the title side, so the trio was not
      // actually centred — this is the same three-section centring
      // full_player_screen.dart's own transport row already relies on.
      final mediaItem = MediaItem(
        id: 'track-001',
        title: 'A very very very long track title that eats up space',
        artist: 'Some Artist',
      );
      final stub = _StubPlayerNotifier(
        pstate.PlayerState(
          queue: [mediaItem],
          currentIndex: 0,
          mediaItem: mediaItem,
          playbackState: PlaybackState(playing: false),
        ),
      );
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      final barCentre = tester
          .getCenter(find.byKey(MiniPlayerBar.contentKey))
          .dx;
      final trioCentre =
          (tester.getCenter(find.byIcon(Icons.skip_previous)).dx +
              tester.getCenter(find.byIcon(Icons.skip_next)).dx) /
          2;
      expect((trioCentre - barCentre).abs(), lessThan(4));
    },
  );

  // ---------------------------------------------------------------------
  // v5.30.6 — local-library artwork fix (requirement.md v5.30.6 / plan
  // Phase v5.30.6, item B). A local track has no albumId at all — its cover
  // is an embedded image extracted straight from the file and exposed as a
  // file:// URI on MediaItem.artUri. Before this, MiniPlayerArtwork only
  // ever knew how to ask the server for a cover via albumId, so every
  // guest-mode local track showed the music-note placeholder in the bar
  // even though the exact same track already rendered its cover correctly
  // in every list row (which reads straight from the local DB rather than
  // going through this widget). track_artwork_test.dart covers the shared
  // TrackArtwork logic itself in more depth; these two prove MiniPlayerBar
  // actually threads mediaItem's fields through to it.
  // ---------------------------------------------------------------------

  testWidgets(
    'a local track (artUri, no albumId) renders its embedded cover via '
    'Image.file rather than the placeholder',
    (tester) async {
      final mediaItem = MediaItem(
        id: 'local:track-1',
        title: 'Local Track',
        artUri: Uri.file('/tmp/does-not-need-to-exist.jpg'),
      );
      final stub = _StubPlayerNotifier(
        pstate.PlayerState(
          queue: [mediaItem],
          currentIndex: 0,
          mediaItem: mediaItem,
          playbackState: PlaybackState(playing: false),
        ),
      );
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      final image = tester.widget<Image>(find.byType(Image));
      expect(
        image.image,
        isA<FileImage>(),
        reason:
            'A local track carries its cover as a file:// artUri, not an '
            'albumId — before v5.30.6 the bar could only ever ask the '
            'server, so this always rendered the placeholder instead',
      );
      expect(find.byIcon(Icons.music_note), findsNothing);
    },
  );

  testWidgets(
    'a server track (albumId, no artUri) still resolves its cover through '
    'artworkUrlProvider',
    (tester) async {
      final mediaItem = MediaItem(
        id: 'track-002',
        title: 'Server Track',
        extras: {'albumId': 'album-x'},
      );
      final stub = _StubPlayerNotifier(
        pstate.PlayerState(
          queue: [mediaItem],
          currentIndex: 0,
          mediaItem: mediaItem,
          playbackState: PlaybackState(playing: false),
        ),
      );
      await tester.pumpWidget(
        _buildApp(
          stub,
          extraOverrides: [
            artworkUrlProvider.overrideWith(
              () => _StubArtworkNotifier('https://example/a.jpg'),
            ),
            // v5.32.0: MiniPlayerBar now also watches coverPaletteProvider
            // itself (to prime the cache for the full player screen — see
            // its own doc comment), which without this override would run
            // the real PaletteGenerator.fromImageProvider against a URL
            // that the test HTTP client 400s — same real-network guard
            // full_player_layout_test.dart's own _appWithRouter already
            // needs for the same reason.
            coverPaletteProvider.overrideWith((ref, source) async => null),
          ],
        ),
      );
      // Two pumps: artworkUrlProvider's build() is async, so the first frame
      // only gets as far as AsyncLoading (which TrackArtwork renders as the
      // placeholder) — the second lets that microtask actually complete.
      await tester.pump();
      await tester.pump();

      // Not also asserting Image's absence: CachedNetworkImage wraps a
      // plain Image internally (with a CachedNetworkImageProvider, not a
      // FileImage) — its own presence is what proves this branch was taken.
      expect(find.byType(CachedNetworkImage), findsOneWidget);
    },
  );

  // ---------------------------------------------------------------------
  // v5.30.6 — Apple-style floating shadow (requirement.md v5.30.6 / plan
  // Phase v5.30.6, item D). Material's own `elevation` shadow is fully
  // retired in favour of an explicit two-layer BoxShadow (floating_shadow
  // _test.dart covers that function itself in depth); this just proves the
  // bar actually wires the replacement in rather than merely adding a
  // shadow *alongside* the old one.
  // ---------------------------------------------------------------------

  testWidgets(
    'the bar\'s drop shadow is an explicit BoxShadow; Material elevation is '
    'off, not just superseded',
    (tester) async {
      final stub = _StubPlayerNotifier(pstate.PlayerState());
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      final shadowFinder = find.byWidgetPredicate((widget) {
        if (widget is! DecoratedBox) return false;
        final decoration = widget.decoration;
        return decoration is BoxDecoration &&
            (decoration.boxShadow?.isNotEmpty ?? false);
      });
      expect(shadowFinder, findsOneWidget);
      final decoration =
          tester.widget<DecoratedBox>(shadowFinder).decoration as BoxDecoration;
      expect(decoration.boxShadow, hasLength(2));

      final material = tester.widget<Material>(
        find
            .ancestor(
              of: find.byKey(MiniPlayerBar.contentKey),
              matching: find.byType(Material),
            )
            .first,
      );
      expect(
        material.elevation,
        0,
        reason:
            'A non-zero elevation here would paint a second, duplicate '
            'shadow underneath the explicit BoxShadow above',
      );
    },
  );

  // ---------------------------------------------------------------------
  // v5.32.0 — cover hover-scale affordance (EchoMusic's PlayerBar.vue
  // group-hover:scale-110, telling users the cover is what opens the full
  // player — see MiniPlayerArtwork's own surrounding _HoverScaleCover doc
  // comment).
  // ---------------------------------------------------------------------

  testWidgets(
    'the cover scales up on hover and back down once the pointer leaves',
    (tester) async {
      final mediaItem = MediaItem(id: 'track-001', title: 'Idol');
      final stub = _StubPlayerNotifier(
        pstate.PlayerState(
          queue: [mediaItem],
          currentIndex: 0,
          mediaItem: mediaItem,
          playbackState: PlaybackState(playing: false),
        ),
      );
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      AnimatedScale scale() =>
          tester.widget<AnimatedScale>(find.byType(AnimatedScale));

      expect(scale().scale, 1.0, reason: 'At rest, no hover has happened yet');

      final mouse = await tester.createGesture(kind: PointerDeviceKind.mouse);
      await mouse.addPointer(location: Offset.zero);
      addTearDown(mouse.removePointer);
      await mouse.moveTo(tester.getCenter(find.byType(MiniPlayerArtwork)));
      await tester.pump();

      final hoveredScale = scale().scale;
      expect(
        hoveredScale,
        greaterThan(1.0),
        reason:
            'Hovering the cover must grow it, matching EchoMusic\'s own '
            'group-hover:scale-110 affordance',
      );
      expect(
        hoveredScale,
        inInclusiveRange(1.06, 1.10),
        reason: 'Field report\'s own bound: noticeable but not exaggerated',
      );

      await mouse.moveTo(const Offset(-10, -10));
      await tester.pump();
      expect(
        scale().scale,
        1.0,
        reason: 'Moving the pointer away must shrink it back to rest',
      );
    },
  );
}
