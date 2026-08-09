// track_artwork_test.dart
//
// Direct widget tests for the shared TrackArtwork "which source wins" logic
// (see lib/src/player/track_artwork.dart) introduced in v5.30.6 to fix a
// real bug: MiniPlayerArtwork (mini_player_bar.dart) only ever knew how to
// ask the server for a cover via albumId, so a guest-mode local track —
// which has no albumId at all, only an embedded-cover file:// URI on
// MediaItem.artUri — always fell back to the placeholder icon in the player
// bar, even though the exact same track already rendered its cover
// correctly in every list row. _FullPlayerArtwork (full_player_screen.dart)
// already had the localArtUri branch; this is that same logic pulled out so
// both call sites share one implementation instead of two that can drift.
//
// mini_player_bar_test.dart separately covers the integration — that
// MiniPlayerBar actually threads mediaItem?.artUri through to this widget.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:cached_network_image/cached_network_image.dart';

import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/player/track_artwork.dart';

/// Stubs the artwork URL lookup so a test can drive the "has a cover"/"no
/// cover" branches without a network round trip — same shape as
/// full_player_layout_test.dart's own _StubArtworkNotifier.
class _StubArtworkNotifier extends ArtworkUrlNotifier {
  _StubArtworkNotifier(this._url);
  final String? _url;

  @override
  Future<String?> build(String albumId) async => _url;
}

const _fallbackKey = ValueKey('fallback');

Widget _fallback(BuildContext context) =>
    const Icon(Icons.music_note, key: _fallbackKey);

/// [stubArtwork] is a separate flag from [artworkUrl] rather than inferring
/// "don't stub" from a null URL — a test resolving to "no artwork found"
/// needs the stub installed and returning null, which is a different thing
/// from not touching artworkUrlProvider at all. Conflating the two the first
/// time this was written meant the "resolves to no artwork" case fell
/// through to the *real* provider (a real Dio call, whose retry/timeout
/// timer then outlived the test — `!timersPending`) instead of the stub.
Widget _app(Widget child, {bool stubArtwork = false, String? artworkUrl}) =>
    ProviderScope(
      overrides: [
        if (stubArtwork)
          artworkUrlProvider.overrideWith(
            () => _StubArtworkNotifier(artworkUrl),
          ),
      ],
      child: MaterialApp(home: Scaffold(body: child)),
    );

/// artworkUrlProvider's `build()` is async even in this stub (it always
/// resolves synchronously-in-spirit, but `async =>` still yields a
/// microtask) — one pump only gets as far as AsyncLoading, which
/// TrackArtwork renders as [fallback], indistinguishable from "no artwork".
/// A second pump lets that microtask actually complete.
Future<void> _settle(WidgetTester tester) async {
  await tester.pump();
  await tester.pump();
}

void main() {
  testWidgets(
    'a file:// localArtUri renders via Image.file, even with no albumId',
    (tester) async {
      await tester.pumpWidget(
        _app(
          TrackArtwork(
            size: 56,
            localArtUri: Uri.file('/tmp/does-not-need-to-exist.jpg'),
            fallback: _fallback,
          ),
        ),
      );
      await tester.pump();

      final image = tester.widget<Image>(find.byType(Image));
      expect(
        image.image,
        isA<FileImage>(),
        reason: 'A file:// artUri must always win over needing an albumId',
      );
      expect(find.byKey(_fallbackKey), findsNothing);
    },
  );

  testWidgets(
    'a non-file localArtUri is treated as absent and falls through to the '
    'albumId lookup',
    (tester) async {
      await tester.pumpWidget(
        _app(
          TrackArtwork(
            size: 56,
            // Not a scheme this widget knows how to read a file from —
            // must not be treated the same as a real file:// URI.
            localArtUri: Uri.parse('content://media/external/audio/1'),
            albumId: 'album-x',
            fallback: _fallback,
          ),
          stubArtwork: true,
          artworkUrl: 'https://example/a.jpg',
        ),
      );
      await _settle(tester);

      // CachedNetworkImage is itself the proof of which branch was taken —
      // it wraps a plain Image internally (visible with a
      // CachedNetworkImageProvider, not a FileImage), so asserting Image's
      // *absence* here would be wrong, not just redundant.
      expect(find.byType(CachedNetworkImage), findsOneWidget);
    },
  );

  testWidgets('no localArtUri, an albumId that resolves: renders via '
      'CachedNetworkImage', (tester) async {
    await tester.pumpWidget(
      _app(
        TrackArtwork(size: 56, albumId: 'album-x', fallback: _fallback),
        stubArtwork: true,
        artworkUrl: 'https://example/a.jpg',
      ),
    );
    await _settle(tester);

    // Not also asserting the fallback key's absence: CachedNetworkImage's
    // own `placeholder` parameter *is* this same fallback (see
    // TrackArtwork's `data:` branch), and widget tests never complete a
    // real network fetch (Flutter's test binding 400s every HTTP
    // request), so the placeholder legitimately renders as
    // CachedNetworkImage's own child throughout the test. Choosing
    // CachedNetworkImage over the file:// / no-source paths is what
    // matters here, and its presence alone proves that.
    expect(find.byType(CachedNetworkImage), findsOneWidget);
  });

  testWidgets('an albumId that resolves to no artwork falls back', (
    tester,
  ) async {
    await tester.pumpWidget(
      _app(
        TrackArtwork(size: 56, albumId: 'album-x', fallback: _fallback),
        stubArtwork: true,
        artworkUrl: null,
      ),
    );
    await _settle(tester);

    expect(find.byType(CachedNetworkImage), findsNothing);
    expect(find.byKey(_fallbackKey), findsOneWidget);
  });

  testWidgets(
    'neither a localArtUri nor an albumId falls back without watching any '
    'provider',
    (tester) async {
      await tester.pumpWidget(
        _app(TrackArtwork(size: 56, fallback: _fallback)),
      );
      await tester.pump();

      expect(find.byKey(_fallbackKey), findsOneWidget);
      expect(find.byType(Image), findsNothing);
      expect(find.byType(CachedNetworkImage), findsNothing);
    },
  );
}
