# Inori Music — Flutter Client

Cross-platform music viewer built with Flutter. Supports Android, iOS, macOS, Windows, and Linux.

## Tech Stack

| Layer | Choice |
|-------|--------|
| Framework | Flutter 3.x (Dart) |
| State Management | Riverpod 2.x (`riverpod_annotation` + `hooks_riverpod`) |
| Routing | go_router 15.x |
| HTTP | dio 5.x (via openapi-generator `dart-dio`) |
| Audio Engine | just_audio + audio_service |
| Secure Storage | flutter_secure_storage |
| Image Cache | cached_network_image |
| Charts | fl_chart |
| Code Gen | freezed + json_serializable + riverpod_generator |

## Architecture

```
main.dart
  └─ InoriMusicApp (MaterialApp.router + Riverpod + l10n)
      └─ GoRouter
          ├─ /login → LoginScreen
          ├─ /player → FullPlayerScreen (overlay)
          └─ ShellRoute → ShellScaffold (adaptive nav + MiniPlayerBar)
              ├─ /artists → ArtistsScreen → /:id → ArtistDetailScreen
              ├─ /albums → AlbumsScreen → /:id → AlbumDetailScreen
              ├─ /tracks → TracksScreen
              ├─ /playlists → PlaylistsScreen → /:id → PlaylistDetailScreen
              ├─ /search → SearchScreen
              ├─ /library/favorites → FavoritesScreen
              ├─ /library/history → HistoryScreen → /stats → HistoryStatsScreen
              └─ /settings → SettingsScreen
```

**Providers** (Riverpod):
- `ApiClientProvider` — dio instance with token interceptor
- `AuthNotifier` — login/logout/token persistence
- `CatalogRepository` — catalog browsing (artists, albums, tracks, playlists, search)
- `PlayerNotifier` — queue, playback, position, volume, shuffle, repeat
- `AudioHandler` — audio_service bridge (MediaSession, lock screen, notifications)
- `TrackFavoriteNotifier` — per-track favorite toggle
- `HistoryNotifier` — play events + stats

## Getting Started

```bash
cd services/mobile
flutter pub get
flutter analyze
flutter test
```

### Generate API Client

```bash
make gen:api
```

### Build Runner

```bash
dart run build_runner build --delete-conflicting-outputs
# or watch mode:
dart run build_runner watch --delete-conflicting-outputs
```

## Platform Targets

| Platform | Command |
|----------|---------|
| Android | `flutter build apk --release` |
| iOS | `flutter build ipa --release` |
| macOS | `flutter build macos --release` |
| Windows | `flutter build windows --release` |
| Linux | `flutter build linux --release` |

## Localization

Three languages supported: English, 简体中文, 日本語.
ARB files: `lib/l10n/app_{en,zh,ja}.arb`

## CI / CD

GitHub Actions runs `flutter analyze` + `flutter test` + `flutter build apk` on every push.
