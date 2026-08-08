// safe_file_name_test.dart
//
// Regression guard for the v5.26.1 Windows import failure: local track ids
// carry a `local:` prefix and were used verbatim as file names, which Windows
// rejects with ERROR_INVALID_NAME while macOS and Linux accept it.
//
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/local_library/local_library_notifier.dart'
    show localTrackIdPrefix;
import 'package:inori_music/src/shared/safe_file_name.dart';

void main() {
  test('a local track id loses the colon that Windows rejects', () {
    const id = '${localTrackIdPrefix}550e8400-e29b-41d4-a716-446655440000';
    final name = safeFileName(id);

    expect(name, 'local_550e8400-e29b-41d4-a716-446655440000');
    expect(name, isNot(contains(':')));
  });

  test('a bare UUID is left untouched', () {
    // Server ids go through the same helper on the offline-download path;
    // rewriting them would orphan files written by earlier builds.
    const uuid = '550e8400-e29b-41d4-a716-446655440000';
    expect(safeFileName(uuid), uuid);
  });

  test('every character Windows reserves is replaced', () {
    expect(safeFileName(r'a<b>c:d"e/f\g|h?i*j'), 'a_b_c_d_e_f_g_h_i_j');
  });

  test('control characters are replaced too', () {
    expect(safeFileName('a\u0000b\u001Fc'), 'a_b_c');
  });

  test('ordinary punctuation survives', () {
    // Over-sanitising would make names harder to correlate with ids by hand.
    expect(safeFileName('track-01_final (v2).mix'), 'track-01_final (v2).mix');
  });
}
