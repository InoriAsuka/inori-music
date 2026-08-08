/// Characters Windows rejects in a file name: the reserved punctuation plus
/// every control character. POSIX only reserves `/` and NUL, which is why a
/// name that is fine on macOS and Linux can still be rejected on Windows.
final _illegalFileNameChars = RegExp(r'[<>:"/\\|?*\x00-\x1F]');

/// Makes an identifier safe to use as a file name on every target platform.
///
/// Track ids are used to name the files this app writes into its own storage.
/// Local-import ids carry a `local:` prefix ([localTrackIdPrefix]), and a
/// colon is illegal in a Windows file name — so every guest-mode import on
/// Windows failed with `ERROR_INVALID_NAME`, surfaced as "文件名、目录名或卷标
/// 语法不正确". It went unnoticed for as long as it did because macOS and
/// Linux both accept `:` happily.
///
/// Only the name on disk is sanitised; the id itself keeps its prefix,
/// because `PlayerNotifier` branches on it to tell local tracks from server
/// ones. Server ids are bare UUIDs, so this is a no-op for them and paths
/// written by older builds still resolve.
String safeFileName(String id) => id.replaceAll(_illegalFileNameChars, '_');
