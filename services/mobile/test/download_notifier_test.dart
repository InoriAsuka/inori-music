// ignore_for_file: implementation_imports
//
// download_notifier_test.dart
//
// Tests state-machine logic of DownloadNotifier that does not require
// the real OfflineDb / SQLite. The notifier's build() calls
// _restoreFromDb() asynchronously, but the synchronous initial state
// is already returned as {} before that resolves, so we can read it
// immediately without touching a real database.
//
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/offline/download_notifier.dart';

void main() {
  // We manipulate the DownloadNotifier's state directly without touching the DB.
  // We can't easily override the OfflineDb singleton, but the pure-state
  // methods (isDownloaded, isDownloading) and direct state inspection don't
  // require it — build() returns {} synchronously before _restoreFromDb fires.

  ProviderContainer makeContainer() {
    final container = ProviderContainer();
    addTearDown(container.dispose);
    return container;
  }

  test('initial state is an empty map', () {
    final container = makeContainer();
    // Reading synchronously gives us the synchronously-returned {} from build().
    final state = container.read(downloadProvider);
    expect(state, isEmpty,
        reason: 'DownloadNotifier.build() returns {} synchronously');
  });

  test('isDownloaded returns false when track is not in state', () {
    final container = makeContainer();
    final notifier = container.read(downloadProvider.notifier);
    expect(notifier.isDownloaded('track-x'), isFalse);
  });

  test('isDownloading returns false when track is not in state', () {
    final container = makeContainer();
    final notifier = container.read(downloadProvider.notifier);
    expect(notifier.isDownloading('track-x'), isFalse);
  });

  test('isDownloaded returns true after state is set to DownloadDone', () {
    final container = makeContainer();
    final notifier = container.read(downloadProvider.notifier);

    // Inject DownloadDone directly — simulates what _restoreFromDb does.
    notifier.state = {'track-done': const DownloadDone()};

    expect(notifier.isDownloaded('track-done'), isTrue);
    expect(notifier.isDownloading('track-done'), isFalse);
  });

  test('isDownloaded returns false after key is removed from state', () {
    final container = makeContainer();
    final notifier = container.read(downloadProvider.notifier);

    notifier.state = {'track-done': const DownloadDone()};
    expect(notifier.isDownloaded('track-done'), isTrue);

    // Simulate key removal (without DB — just update state).
    final updated = Map<String, DownloadStatus>.from(notifier.state)
      ..remove('track-done');
    notifier.state = updated;

    expect(notifier.isDownloaded('track-done'), isFalse);
  });
}
