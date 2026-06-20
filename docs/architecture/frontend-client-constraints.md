# Frontend & Client Architecture Constraints

## Scope

This document defines the technology stack, design language, and boundary constraints for the two
web-facing products and the cross-platform Flutter client:

| Product | Target | Entry point |
|---------|--------|-------------|
| **inori-web** | User-facing music player (viewer role) | `packages/web/` |
| **inori-admin** | Back-office management console (admin role) | `packages/admin/` |
| **inori-app** | Cross-platform Flutter client (Android / iOS / Desktop / Web) | `packages/app/` |

All three products consume the same REST API defined in
`packages/api-contract/openapi/storage-admin.v1.json`.

---

## 1. Web Products — Shared Technical Constraints

### 1.1 Core Stack (non-negotiable)

| Layer | Choice | Version floor |
|-------|--------|---------------|
| Runtime framework | **React** | 19+ |
| Build tool | **Vite** | 8+ |
| Language | **TypeScript** | 5.5+ (strict mode) |
| Component primitives | **shadcn/ui** | latest (Radix UI foundation) |
| Styling engine | **Tailwind CSS** | 4+ (CSS-first config, `@theme` directive) |
| State management | **Zustand** | 5+ (atomic stores per domain) |
| Server state / data fetching | **TanStack Query** | 5+ |
| Routing | **TanStack Router** | 1+ (file-based, type-safe) |
| Form handling | **React Hook Form** + **Zod** | latest |
| Animation | **Motion (formerly Framer Motion)** | 11+ |
| Icons | **Lucide React** + **Phosphor Icons** | latest |

### 1.2 API Client

- Generated from `packages/api-contract/openapi/storage-admin.v1.json` via **openapi-typescript** + **openapi-fetch**.
- All API types are auto-generated; never hand-write request/response types.
- Token is stored in `localStorage` (web) and injected via TanStack Query's default headers.

### 1.3 Audio Engine (inori-web only)

- **Howler.js** for cross-browser HTML5 audio with Web Audio API fallback.
- Playback state (current track, queue, position, volume) lives in a dedicated Zustand store.
- Presigned URLs fetched from `GET /api/v1/catalog/tracks/{id}/playback`; TTL is 15 min —
  client must re-fetch before the URL expires during long sessions.

### 1.4 Build & Tooling

- **pnpm** workspaces mono-repo (root `pnpm-workspace.yaml`).
- **Biome** for formatting and linting (replaces ESLint + Prettier).
- **Vitest** for unit and component tests; **Playwright** for E2E.
- **Storybook** 8+ for component development and design-system documentation.

---

## 2. Design Language

### 2.1 Visual Identity — "Neon Shrine"

The visual identity fuses three aesthetics into one coherent system:

```
ACG / Anime-Manga-Game (Japanese pop-culture aesthetic)
  ↕ layered with
Cyberpunk / Synthwave (neon-lit, high-contrast technology aesthetic)
  ↕ grounded by
Futuristic HiFi Player (precision instrument, near-future interface)
```

The result: a dark, luminous, high-contrast UI that feels like a high-end Japanese music player
rendered in a near-future city — clean geometry, subtle holographic glows, and deliberate otaku
cultural references woven into the motion design.

### 2.2 Color System (Tailwind 4 `@theme`)

```css
@theme {
  /* Base canvas */
  --color-void:          #070711;   /* near-black canvas */
  --color-surface:       #0d0d1a;   /* card / panel background */
  --color-surface-raised:#12122b;   /* elevated surface */
  --color-border:        #1e1e3f;   /* default border */
  --color-border-glow:   #2d2d6b;   /* interactive border */

  /* Primary — electric violet */
  --color-primary-50:    #f0e8ff;
  --color-primary-400:   #b97dff;
  --color-primary-500:   #9b5cff;   /* main accent */
  --color-primary-600:   #7c3aed;
  --color-primary-900:   #1a0533;

  /* Secondary — cyan-teal */
  --color-secondary-400: #22e5d4;
  --color-secondary-500: #0fd4c0;   /* secondary accent */
  --color-secondary-900: #022926;

  /* Tertiary — sakura pink */
  --color-sakura-300:    #ffb3d1;
  --color-sakura-500:    #ff5fa0;   /* highlight / active state */
  --color-sakura-700:    #c2185b;

  /* Semantic */
  --color-success:       #00e5a0;
  --color-warning:       #ffd23f;
  --color-danger:        #ff4d6d;
  --color-info:          #38bdf8;

  /* Text */
  --color-text-primary:  #e8e8f4;
  --color-text-secondary:#9090b8;
  --color-text-muted:    #4a4a7a;
}
```

### 2.3 Typography

| Role | Font | Weight |
|------|------|--------|
| Display / Hero | **Orbitron** (Google Fonts) | 700–900 |
| UI Labels | **Inter** | 400 / 500 |
| Monospace / ID | **JetBrains Mono** | 400 |
| Japanese glyphs | **Noto Sans JP** | 400 / 700 |

Rules:
- Track titles: Inter 500 with optional Noto Sans JP fallback for CJK characters.
- Player time codes, bitrates, IDs: JetBrains Mono.
- Section headers and hero callouts: Orbitron in `tracking-widest`.
- Never mix Orbitron with Noto Sans JP in the same line element.

### 2.4 Motion & Animation

All animation uses **Motion** (Framer Motion v11+).

| Pattern | Spec |
|---------|------|
| Page transitions | Shared-layout slide + fade, 200 ms, `easeOutExpo` |
| Track card hover | Scale 1.02 + border-glow pulse, 150 ms |
| Playback progress | Spring physics scrubber, `stiffness: 200, damping: 26` |
| Equalizer bars | SVG path animation driven by Web Audio API `AnalyserNode` |
| Admin table row enter | Stagger fade-up, 30 ms delay per row, max 8 rows |
| Modal / sheet open | Slide-up from bottom on mobile, fade-scale on desktop |
| Skeleton loader | Gradient shimmer via CSS `@keyframes`, no JS |
| Glitch text | CSS `clip-path` glitch effect on hover for album / artist headers |

Reduced-motion: all animations collapse to a 0 ms fade when `prefers-reduced-motion: reduce`.

### 2.5 Spatial & Depth System

- **3 elevation layers**: void (canvas) → surface → raised → overlay.
- Background of panels: semi-transparent `backdrop-blur-xl` glass with `border-border/40`.
- Glow effects: `box-shadow` with primary/secondary color at 15–30% opacity; never pure white.
- Scanline texture: repeating 1 px horizontal CSS gradient at 3% opacity overlaid on hero sections.
- Grid overlay: subtle dot-grid (`radial-gradient`) on admin dashboard backgrounds.

### 2.6 ACG / Japanese-style Details

- Album art displayed in rounded-sm frames with a colored halo border matching dominant palette.
- Furigana-style sub-labels (artist sort-name, track sort-title) rendered in `text-xs text-muted` above the main title where space permits.
- Loading states use a spinning SVG gate/sakura-blossom motif (single line-art, no heavy illustration).
- Empty states use haiku-format copy (3 lines, evocative, never "No data found").
- Notification toasts appear from top-right with a cherry-blossom petal CSS particle (2–3 petals, pure CSS, no canvas).

### 2.7 Cyberpunk / Synthwave Details

- Active nav items: left-border `border-l-2 border-primary-500` with a `text-shadow` glow.
- Data tables: alternating row tints (`surface` / `surface-raised`), monospace cell values for IDs and sizes.
- Waveform visualizer: canvas-based frequency bar chart in primary-to-secondary gradient.
- "Online" / health indicators: pulsing green dot (`animate-ping` at 0.5 opacity behind solid dot).
- Error states: red scanline flicker CSS animation (3 frames, 80 ms total).

---

## 3. inori-web — User Player Constraints

### 3.1 Functional Boundaries

| In scope | Out of scope |
|----------|-------------|
| Auth (login / logout / change password) | User registration (admin creates accounts) |
| Browse catalog (artists / albums / tracks / playlists) | Catalog editing |
| Full-text search | Admin search filters |
| Playback with queue management | Upload / import |
| Playback history (view / delete own events) | Other users' history |
| Personal stats & timelines | Global admin stats |
| Session management (list / revoke) | Admin user management |

### 3.2 Layout

```
┌─────────────────────────────────────────────────────┐
│  Sidebar (collapsible, 240px)  │  Main content area  │
│  ─────────────────────────     │  ─────────────────  │
│  [Logo: inori]                 │  [Page header]      │
│  ─ Browse                      │                     │
│    Artists / Albums / Tracks   │  [Scrollable body]  │
│    Playlists                   │                     │
│  ─ History                     │                     │
│    Recent / Stats / Timeline   │                     │
│  ─ Settings                    │                     │
│    Account / Sessions          │                     │
├────────────────────────────────┴─────────────────────┤
│  Player bar (fixed bottom, 72px)                     │
│  [Art] [Title / Artist]  [Controls]  [Progress]      │
│  [Volume] [Queue] [Lyrics stub]                      │
└─────────────────────────────────────────────────────┘
```

- **Mobile**: bottom nav bar (5 icons) + full-screen player sheet; sidebar hidden.
- **Tablet**: sidebar collapses to icon-only (56px).
- **Desktop**: sidebar at 240px, player bar always visible.

### 3.3 Player Bar Specifics

- Scrubber: custom range input styled with Tailwind 4; position updated at 250 ms interval.
- Waveform visualizer: 64-bar canvas FFT display, lazy-initialized only when audio context is active.
- Queue drawer: slide-in sheet from the right, drag-to-reorder via `@dnd-kit/core`.
- Keyboard shortcuts: `Space` play/pause, `←/→` seek ±5 s, `↑/↓` volume, `N/P` next/prev.

### 3.4 Responsive Breakpoints (Tailwind 4)

```
sm:  640px   (mobile landscape)
md:  768px   (tablet portrait)
lg:  1024px  (tablet landscape / small desktop)
xl:  1280px  (desktop)
2xl: 1536px  (large desktop / wide)
```

---

## 4. inori-admin — Management Console Constraints

### 4.1 Functional Boundaries

| Domain | Operations |
|--------|-----------|
| Users | List (paginate / sort / filter) · Create · Get · Patch · Disable / Enable · Delete · Force-change-password · View sessions · Revoke sessions |
| Catalog | Artists / Albums / Tracks / Playlists full CRUD · Import from media object · Batch import · Relink track media · Search |
| Storage backends | Register · List · Set default · Disable · Probe · Health · Capacity |
| Media objects | Register · List · Get · Timeline · Verify (single / batch) · Lifecycle (single / bulk / dry-run) · Duplicates · Stats |
| History | Global list · Stats · Top-tracks · Top-users · Timeline · Per-user / per-track views · Bulk delete · Window delete |

### 4.2 Layout

```
┌──────────────────────────────────────────────────────┐
│  Top bar (56px)                                       │
│  [Logo]  [Global search]  [Notifications]  [Avatar]  │
├──────────────┬───────────────────────────────────────┤
│  Left nav    │  Main content area                    │
│  (200px)     │  ┌──────────────────────────────────┐ │
│              │  │  Breadcrumb / page title          │ │
│  Dashboard   │  ├──────────────────────────────────┤ │
│  Catalog ▾  │  │                                  │ │
│    Artists   │  │  Data table / form / detail view │ │
│    Albums    │  │                                  │ │
│    Tracks    │  └──────────────────────────────────┘ │
│    Playlists │                                       │
│  Media ▾    │                                       │
│    Objects   │                                       │
│    Storage   │                                       │
│  Users       │                                       │
│  History ▾  │                                       │
│  Settings    │                                       │
└──────────────┴───────────────────────────────────────┘
```

- **No mobile layout required**; minimum viewport 1024px.
- Left nav collapses to 56px icon-only mode (toggle button in top bar).

### 4.3 Data Table Standard

- Built on **TanStack Table** v8+ with server-side pagination/sort/filter.
- All list endpoints use the API's `limit` / `offset` / `sortBy` / `sortOrder` query params directly.
- Columns: resizable, hideable via column visibility menu.
- Row actions: inline icon buttons (edit, delete, view) + kebab overflow menu for secondary actions.
- Bulk select: checkbox column → bulk-action bar slides up from bottom.
- Export: CSV export of current filtered/sorted page (client-side, no new API endpoint needed).

### 4.4 Admin-specific Components

| Component | Description |
|-----------|-------------|
| `StorageHealthBadge` | Pulsing dot + status text: healthy / unhealthy / unknown / disabled |
| `LifecyclePill` | Color-coded pill: active (green) · archived (amber) · deleted (red) |
| `MediaObjectTimeline` | Vertical timeline: registration → verification → lifecycle change |
| `HistoryTimelineChart` | Recharts `AreaChart` with day/week/month granularity toggle |
| `ImportWizard` | 3-step sheet: select media object → fill metadata → confirm |
| `BulkLifecycleModal` | Filter input → dry-run preview table → confirm commit |
| `CatalogSearchCombobox` | Debounced full-text search with grouped artist/album/track results |

---

## 5. Flutter Client (inori-app) Constraints

### 5.1 Core Stack

| Layer | Choice |
|-------|--------|
| Framework | **Flutter** 3.22+ (Dart 3.4+) |
| State management | **Riverpod** 2+ (code-generated with `riverpod_generator`) |
| Navigation | **go_router** 14+ |
| API client | **Dio** + generated client from OpenAPI spec |
| Local persistence | **Drift** (SQLite ORM, code-generated) |
| Audio | **just_audio** + **audio_service** (background playback, lock-screen controls) |
| Media cache | **cached_network_image** for artwork; custom audio segment cache via `just_audio` |

### 5.2 Platform targets

- Android 8.0+ (API 26+)
- iOS 14+
- macOS 12+ (Catalyst)
- Windows 10+ (secondary)
- Web (Flutter web — progressive enhancement only, not primary target)

### 5.3 Local SQLite Schema (Drift)

| Table | Purpose |
|-------|---------|
| `cached_artists` | Offline artist list cache with `etag` / `cached_at` |
| `cached_albums` | Offline album cache |
| `cached_tracks` | Offline track metadata cache |
| `cached_playlists` | Offline playlist + ordered track IDs |
| `play_queue` | Current playback queue (ordered) |
| `pending_history_events` | Buffered `POST /api/v1/me/history` events to flush when online |
| `auth_session` | Single-row: token, user_id, expires_at |

### 5.4 Design Language (Flutter)

- Same "Neon Shrine" color palette defined in Section 2.2, implemented as a Flutter `ThemeData`.
- Custom `NeonCard` widget: `BoxDecoration` with `color: surface`, `border: Border.all(color: borderGlow)`, `BoxShadow` glow.
- Hero animation on album art when navigating from list → detail.
- `just_audio` visualizer: custom `CustomPainter` frequency bar widget, 32 bars.
- Haptic feedback on play/pause toggle (`HapticFeedback.lightImpact`).

---

## 6. Shared Cross-Cutting Constraints

### 6.1 Authentication

- Bearer token stored in `localStorage` (web) / `FlutterSecureStorage` (mobile).
- Auto-logout on `401` response; TanStack Query / Riverpod error handler redirects to login.
- Session expiry warning toast shown 5 min before `expiresAt` (polled by a background timer).

### 6.2 Offline & Error States

- All list views show a skeleton loader matching the expected row/card layout (no generic spinner).
- Network error: in-line retry button + last-fetch timestamp.
- Empty state: haiku-format copy + a contextual illustration (SVG, < 4 KB).
- `503` from API (service not configured): specific message per domain, not generic "server error".

### 6.3 Accessibility

- WCAG 2.1 AA minimum contrast on all text (verified by Storybook a11y addon).
- All interactive elements keyboard-navigable; focus ring uses `ring-2 ring-primary-500`.
- `aria-live` regions for player status and notification toasts.
- Screen-reader-friendly track list: `role="list"` + `aria-label` per item.

### 6.4 Internationalisation (i18n)

- **Web**: **i18next** with `react-i18next`; locale files under `src/locales/{en,zh,ja}.json`.
- **Flutter**: `flutter_localizations` + `intl` package; `.arb` files for en / zh / ja.
- Default locale: English; Japanese and Simplified Chinese are tier-1 targets given the ACG audience.
- Date/time: always format with locale-aware `Intl.DateTimeFormat` / Dart `intl`; never hardcode.

### 6.5 Performance Budgets

| Metric | Target |
|--------|--------|
| Web LCP (desktop) | < 1.5 s |
| Web LCP (mobile) | < 2.5 s |
| Web JS bundle (initial) | < 200 KB gzipped |
| Web JS bundle (per-route chunk) | < 80 KB gzipped |
| Flutter app start (cold) | < 2 s on mid-range Android |
| API list request round-trip | < 300 ms p95 on LAN |

### 6.6 Package & Monorepo Layout

```
packages/
├── api-contract/         # OpenAPI spec (already exists)
│   └── openapi/
│       └── storage-admin.v1.json
├── api-client/           # Auto-generated TS API client (new)
│   ├── src/generated/
│   └── package.json
├── web/                  # inori-web (new)
│   ├── src/
│   └── package.json
├── admin/                # inori-admin (new)
│   ├── src/
│   └── package.json
├── ui/                   # Shared shadcn/ui component library (new)
│   ├── src/components/
│   └── package.json
└── app/                  # Flutter client (new)
    ├── lib/
    └── pubspec.yaml
```

Shared UI components (`packages/ui/`) are consumed by both `web` and `admin`; the Flutter client
has its own widget library and does not share components with the web packages.

---

## 7. Non-Goals (Explicitly Out of Scope for Now)

- Server-side rendering (SSR) or Next.js — Vite SPA only for 1.x.
- PWA offline-first for web — Flutter handles offline; web is network-first.
- Custom audio codec support — browser native and `just_audio` supported codecs only.
- Social features (comments, following, sharing) — no API surface exists yet.
- Desktop Electron wrapper — Flutter desktop covers that use case.
- Admin mobile layout — console is desktop-only.
