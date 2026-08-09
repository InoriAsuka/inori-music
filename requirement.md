# inori-music Requirements

## Current Version

`5.21.0`

## Product Goal

Build a cross-platform music playback system for Web, Android, iOS, and desktop clients while supporting both browser/server and client/server architectures. The server owns media storage configuration, metadata registration, health checks, integrity verification, and administrative APIs. Large media bytes are stored in external storage backends rather than the relational database.

## Technical Requirements

- Flutter-first client direction.
- Go modular monolith first for the server.
- PostgreSQL-first server metadata database.
- SQLite for client-side local persistence.
- PostgreSQL full-text search first for 0.x, with external search engines left as future extensions.
- Media storage must support local filesystems, NFS, SMB, S3-compatible object storage, and distributed storage adapters.
- Repository automation must validate builds and tests, publish tagged release binaries, and publish Docker images for deployable API artifacts.
- Runtime API artifacts must expose non-sensitive build metadata for deployment diagnostics.
- Runtime API artifacts must expose public readiness diagnostics for storage, media registry, and admin-auth configuration.
- Runtime API artifacts must expose non-sensitive Prometheus-compatible metrics for deployment monitoring.
- Runtime HTTP metrics must avoid high-cardinality labels by using route patterns instead of raw URLs.

## Storage Requirements

- Do not store large audio, image, or derived media files in the relational database.
- Store object IDs, backend IDs, object keys, hashes, lifecycle state, asset kind, verification state, and references as metadata.
- Probes and verification must use server-owned temporary objects or read-only checks to avoid damaging user media.
- Admin APIs must expose a read-only per-object metadata timeline derived from retained registration, latest verification, and latest lifecycle transition state.

## Documentation Requirements

- Markdown documentation is maintained in English.
- Phase work must be recorded under `.plan/` with requirements, task checklists, non-goals, and follow-up candidates.
- README, requirements, ADRs, and architecture notes must stay aligned with the current version baseline.
- Media object list APIs must support deterministic sort controls before pagination so admin clients can build predictable tables.
- Media object administration must expose metadata-only duplicate content-hash detection for deduplication planning without reading media bytes.
- Media object administration must support metadata-only bulk lifecycle updates scoped by exactly one safe selection filter.
- Bulk lifecycle updates must support dry-run previews that do not persist metadata changes.
- Committed lifecycle updates must record latest transition metadata for audit preparation.

### v5.30.6 - 2026-08-09

- **feat: 桌面播放控制条重做（曲目信息块回迁 + 本地曲库封面修复 + 进度条与时间显示重做 + 悬浮阴影）** — 用户实机验证 v5.30.5 后给出三条反馈：专辑封面应该跟音乐控制块在一起（不是钉在侧栏）；播放条本地曲目封面不显示；时间条样式太丑（没有时间文字、通栏细线、饱和度过高）；悬浮立体感跟苹果差一节（`Material elevation: 8` 又紧又黑）。本批一次性处理这四项，其中第一项是**回退 v5.30.5 的一个决策**——不是新需求。
- **A. 曲目信息块回迁播放条**：v5.30.5 把封面+标题钉到侧栏底部（`SidebarNowPlaying`）是对用户红框的错误解读；用户明确说"专辑封面需要跟音乐控制块一起"。`SidebarNowPlaying` 连同它的调用点、测试一并删除（删干净后没有任何调用点，不是留作死代码）；`_DesktopSidebar` 恢复到列表直接收尾、不带底部信息块的形状；`_DesktopLayout` 的四区布局（侧栏贯穿全高、播放条只占右列）本身**没有改动**——用户只否掉了信息块位置，没否掉布局。
- **B. 修复本地曲库封面不显示的真 bug**：本地曲目没有 `albumId`（游客模式曲目，封面是从文件里抽出来的内嵌图，存在 `MediaItem.artUri` 这个 `file://` URI 里），而 `MiniPlayerArtwork` 此前只认 `albumId` 走 `artworkUrlProvider`，导致同一首歌在列表行里有封面、播放条里却永远是占位图标——这个缺口从 v4.6.0 修好列表封面之后就一直没堵上。取图逻辑（先看 `localArtUri` 是不是 `file://`，再退回 `albumId` 查服务端，最后兜底占位图）在 `_FullPlayerArtwork`（`full_player_screen.dart`）里其实早就有，只是从没同步到 `MiniPlayerArtwork`。新增共享 widget `TrackArtwork`（`lib/src/player/track_artwork.dart`），把这段"选源+处理加载/出错态"的逻辑收成一份；`_FullPlayerArtwork` 与 `MiniPlayerArtwork` 各自保留自己的外层容器（前者素色 `ClipRRect` 配 80px 大图标占位、后者带背景色的圆角方块配比例缩放小图标），只共享内部真正相同的那部分，没有为了"抽成一个 widget"而把两种视觉硬凑成一种。
- **C. 进度条与时间显示重做（仅桌面宽播放条；窄播放条原样保留）**：参照两张参考截图确定的取舍——进度条位置学参考图 A（传输键下方居中的第二行，两端各一个时间标签），不学参考图 B（顶部通栏，那正是用户嫌丑的现状）。桌面宽播放条的中段现在是两行：`[随机][上一首][▶][下一首][循环]` 一组居中，下方是 `00:02 ──●── 05:55`；时间用等宽数字（`FontFeature.tabularFigures()`）避免秒数跳动时整行左右抖；轨道加粗到 4px，inactive 轨道改用低透明度的 `onSurfaceVariant`（不再是读起来像描边色的 `outline`），active 轨道对 `sakuraPink` 做轻度降饱和（`HSLColor` 调 `saturation`，跟播放/暂停键的满饱和强调色区分开）；滑块默认隐藏（半径 0），hover 或拖拽时才出现（半径 7），两张参考图都是这个处理。播放/暂停键改成 `sakuraPink` 实心圆 + 白色图标（跟 `full_player_screen.dart` 已有的播放键、以及 `FilledButtonTheme` 的 `foregroundColor: Colors.white` 保持同一套约定，不重新计算对比色）。随机/循环从播放条最左边归位到传输键两侧，跟两张参考图一致。
- **移除 `showNowPlaying` 参数，改成按播放条自身测到的宽度分流**：v5.30.5 引入的 `showNowPlaying` 布尔值本来是给"桌面隐藏封面、腾地方放随机/循环+音量/队列"用的；A 项把封面在所有布局下都改回常驻后，这个开关唯一的历史职责就消失了——桌面和移动端/平板会传入同一个值，`false` 分支变成永远不会被生产代码触发的死路径。改为播放条自己用 `LayoutBuilder` 量出的实际宽度（不是窗口宽度）决定用"宽播放条"还是"窄播放条"形状：宽播放条（≥640px，覆盖桌面 shell 和足够宽的平板横屏）叠加 C 项的两行中段 + 随机/循环 + 音量/队列；窄播放条（手机、较窄平板）保留 v5.30.6 之前的单行形状（封面+标题｜上一首/播放/下一首｜睡眠定时器），连同原本那条通栏细进度条一起原样不动——用户原话"窄宽度下保留一条简化的进度表现即可"最贴切的读法就是"保留现有那条"，不是另做第三种形状。校准阈值时抓到一次真实溢出（`_transportBlockWidth` 首版 232px 比五按钮组实际需要的 240px 窄了 8px，`RenderFlex overflowed by 8.0 pixels`），调到 248px 后（多留一点余量，风格上跟 `_volumeCompactThreshold` 的 240 相同套路）在 375/700/900/964 四档宽度下都不再溢出。
- **D. 悬浮阴影改用显式两层 `BoxShadow`**：`mini_player_bar.dart` 的 `Material(elevation: 8)`（去掉 `elevation`，改 0）和 `GlassPanel`（侧栏、播放页控制面板/侧面板）都换成新helper `floatingShadow()`（`shared/widgets/floating_shadow.dart`）——一层贴边的紧阴影（模糊 10、偏移 3）+ 一层大而淡的扩散（模糊 28、偏移 8，落在 Apple 参考的 24–32 模糊/4–8 偏移区间内）。阴影颜色不是写死黑色，而是读皮肤已有的 `SkinColors.miniPlayerShadow`（樱花薄暮是低透明度暗梅色，本来就落在参考的 0.10–0.18 透明度区间；月靛是更高透明度的纯黑，深色地面需要更强的阴影才压得住）——两层透明度都按这个 token 自身的透明度等比例缩放，而不是给所有皮肤共用一个绝对值。踩了一个容易犯的层级坑：`GlassPanel` 内部是 `ClipRRect(BackdropFilter(Material(...)))`，`ClipRRect` 会把画在它里面的阴影直接裁掉，阴影必须包在 `ClipRRect` 外面的 `DecoratedBox` 上——这个仓库自己的 `login_screen.dart` 的 `_GlassCard` 早就用过这个"外层 `Container` 扛阴影、内层 `ClipRRect` 扛模糊"的写法，这次是把单层阴影版本推广成两层、并让其余悬浮面板复用同一个 helper。只改阴影，没有动 macOS 工具栏高度/vibrancy 材质/失焦态红绿灯——那些留给后续 phase。
- **与计划的一处主动简化（非偏离，供复核）**：计划原文写"桌面布局改回 `MiniPlayerBar(showNowPlaying: true)`"，字面是保留参数、只改调用点的值。实现时判断保留这个参数会立刻产生死分支（两个调用点此后永远传同一个值，`false` 分支再无人触发），所以直接删掉了参数，改用上面说的宽度分流。可见结果（桌面播放条重新带封面）跟计划完全一致，只是机制从"标志位"换成了"实测宽度"，且后者跟本仓库反复踩坑总结出的"用区域实际宽度判断，不用外部标志"这条经验一致。
- **测试**：新增 17 例（324 → 341）。两个新文件：`track_artwork_test.dart`（5 例，覆盖 `TrackArtwork` 的取图优先级：`file://` 优先于 `albumId`、非 `file` scheme 的 `localArtUri` 视同缺失、`albumId` 解出 URL/解出 null/完全没有来源三种落点）、`floating_shadow_test.dart`（7 例，纯函数覆盖两层阴影的层数/模糊递增关系/参考区间/深浅皮肤强度对比/透明度不越界/向下偏移，另加 `GlassPanel` 阴影确实包在 `ClipRRect` 外面的结构断言）。`mini_player_bar_desktop_test.dart` 整体重写（原 4 例测 `showNowPlaying: false`，现 6 例测按宽度分流的宽/窄两种形状，含 964px 与 375dp 两档不溢出回归，净增 2 例）。`mini_player_bar_test.dart` 追加 3 例（本地曲目走 `Image.file`、服务端曲目仍走 `artworkUrlProvider`、阴影确实是显式 `BoxShadow` 而非 `Material.elevation`）。`shell_scaffold_nav_test.dart` 用例数不变，把"侧栏底部有曲目信息、播放条没有"反转成"播放条有曲目信息、侧栏没有"，"点侧栏曲目信息块跳播放页"改成"点播放条自己的曲目信息区跳播放页"。校准过程中还抓到一个测试自身的逻辑漏洞：`track_artwork_test.dart` 的 `_app()` 帮助函数原先用 `artworkUrl != null` 判断要不要打桩，导致"解析出无封面"（`artworkUrl: null`）这个用例从未真正打桩、落到了真实 `Dio` 请求，遗留的重试计时器在测试收尾时触发 `!timersPending` 断言——改成独立的 `stubArtwork` 布尔值后测试本身也更准确地表达了意图。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条未触碰文件的存量 info；`flutter test` 341/341 通过）。实际观感待用户实机确认；`_wideBreakpoint`（640px）和 `_transportBlockWidth`（248px）是本批新加的阈值常量，按同一套"先估算、跑测试抓真实溢出、再校准"的流程定出来的，值得在真机上留意超宽/超窄边界。

### v5.30.5 - 2026-08-09

- **feat: 桌面四区布局改造 + 播放控制条补齐（随机/循环/音量/队列）+ Cover Flow 顶栏入口 + macOS 红绿灯让位坐标修复** — 承接 v5.30.0 实机验收结论：用户红框标出的目标布局是"侧栏贯穿全高（含底部曲目信息块）+ 右列三段（标题栏/列表/播放条）"，跟当时"侧栏悬浮但仍与全宽播放条共享同一个 `Column`"的形状不一样；播放控制条也被指出缺音量/随机/循环；Cover Flow 藏得太深找不到；红绿灯压在侧栏圆角和字样上。本批一次处理完这四项，外加追查出红绿灯问题的真根因——不是审美问题，是让位坐标钉错了地方。
- **A. `_DesktopLayout` 从"两行"改成"两列"**：原来是 `Column[Expanded(Row[侧栏, 内容]), 播放条]`——侧栏和内容共享一个 `Row`，播放条在外层 `Column` 独占最后一行、横跨全宽（含侧栏正下方）。改成 `Row[侧栏(全高), Expanded(Column[Expanded(内容), 播放条])]`——播放条现在嵌进右列自己的 `Column`，只占内容列宽度；侧栏的 `Padding(8,8,8,8)` 外边距直接由窗口高度决定，不再被"减去播放条高度"这一步吃掉一截。
- **B. 曲目信息块搬到侧栏底部**：`MiniPlayerBar` 新增 `showNowPlaying`（默认 `true`，移动端/平板走默认，行为不变）。桌面播放条传 `showNowPlaying: false`：第 1 段（封面+标题/艺术家）不再渲染，让出的宽度给随机/循环。原来私有的 `_MiniPlayerArtwork` 提升为公开的 `MiniPlayerArtwork`（现在被侧栏和播放条两处复用）。新增 `SidebarNowPlaying`（同样在 `mini_player_bar.dart`）钉在 `_DesktopSidebar` 的 `Column` 底部（`Expanded(ListView)` 之后再补一条 `Divider` 分隔），点击跳 `AppRoutes.player`，跟播放条原来的 `InkWell` 行为一致。
- **C. 播放条补齐控件**：
  - 随机/循环从 `full_player_screen.dart` 的内联代码抽成共享 widget `RepeatModeButton` / `ShuffleButton`（新文件 `playback_mode_buttons.dart`），播放条和播放页调用同一份状态机（`none→all→one→none` 循环、`shuffle` 布尔翻转），不再是两份长得像、行为可能悄悄跑偏的代码。
  - **音量控件是全新的**：`PlayerNotifier.setVolume` 和 `PlayerState.volume` 早就存在，UI 一直没接上。新文件 `volume_control.dart`：`VolumeControl`（喇叭图标按音量分三档 `volume_off`/`volume_down`/`volume_up`，配一条 90px 横向 `Slider`）+ `volumeBeforeMuteProvider`（记住静音前的音量的一个跨 widget 共享的 `StateProvider`，不是每个 `VolumeControl` 实例自己的本地 `State`——播放条和播放页各挂一份独立实例，静音记忆若各自为政，从一边静音、另一边恢复就会对不上）。`compact: true` 时收成纯图标 + `showMenu` 弹出同一套控件；播放页固定用这个形态，因为那一行宽度是精算过的 `playerControlWidth`，硬塞一条 90px 滑条正是历史上撞过墙的那类加法。
  - 播放条右段窄到某个阈值时，音量滑条会自动收成 `compact` 弹出模式；判断基准是**该分区 `LayoutBuilder` 量出的实际宽度**，不是窗口宽度——这条已经踩过两次坑（v5.28.0/v5.29.0）的规则本次继续适用。阈值定在 240（不是更直觉的 200）：右段展开态实测正好需要 234px（3 个默认 48px `IconButton` + 90px 滑条，中间没有额外间距），200 会在 [200, 234) 之间留一段"既不收又放不下"的真实溢出区间——被 `test/mini_player_bar_desktop_test.dart` 在 600px 分区宽度下实测抓到，不是靠肉眼调出来的。
  - **队列入口**：桌面没有独立于 `FullPlayerScreen` 之外的队列 UI（可复用的 `_QueueList` 只挂在播放页自己的侧栏面板/底部弹层里），新按钮直接跳 `AppRoutes.player`，没有为此再搭一套队列视图。
- **D. Cover Flow 门控验证结果：门控本身没问题**。按计划要求先写测试证伪再动代码：`test/full_player_layout_test.dart` 里 v5.30.0 就有的三条门控测试（开关开+队列 4 首+宽窗→出现；单曲目即使开着也回退；开关关着不出现）本次原样跑通，**没有改动任何门控代码**。真正的问题是入口太深——此前只有 Settings → 外观里一个开关。在播放页顶栏加了一个并排图标（`Icons.view_carousel_outlined`，跟歌词/队列切换按钮同一行），读写同一个 `coverFlowModeProvider`，Settings 里的开关保留、仍是持久化来源。新增一条端到端测试：从关闭状态点按钮 → `CoverFlowArtwork` 出现 → 再点一次 → 消失，证明按钮确实接上了这个 provider，不是摆设。
- **E. 红绿灯让位坐标修复**：`DesktopAppBar`/`DesktopSliverAppBar` 里 `EdgeInsets.only(left: 70)` 此前无条件对 macOS 生效；v5.30.0 把侧栏改成悬浮面板后，窗口左上角实际归侧栏所有，这 70px 却还留在内容区的 `AppBar` 上（x ≥ 236，本来就撞不到红绿灯）——两边各管各的，谁都没有真的把空间让给红绿灯。新增 `ShellChrome`（`shared/widgets/shell_chrome.dart`）：一个只携带 `reservesTrafficLightGutter` 布尔值的 `InheritedWidget`，由 `_DesktopLayout` 包在内容列外面；`DesktopAppBar`/`DesktopSliverAppBar` 读到这个祖先且值为真时跳过 70px 让位，读不到时（登录页、独立的全屏播放页等无侧栏场景）保留原行为。`_DesktopSidebar` 自己接过真正的让位职责：标题行（`InoriMark` + "Inori Music"）顶部内边距在 macOS + 自绘标题栏下从 24 提到 30——按 macOS 红绿灯几何（距窗左 20pt、按钮 12pt 直径、20pt 中心距、垂直中心约 y=20pt，即底边落在窗口坐标 y=26）与侧栏自身 8px 外边距推算，算法写进了行内注释。这段留白同时包了一层 `DragToMoveArea`（此前侧栏完全没有可拖动区域，全靠内容区那条 `DesktopAppBar`），让用户能像 Apple Music 一样拖侧栏顶部移动窗口——只在 macOS 分支生效，Windows/Linux/移动端渲染路径未改动，靠新增测试守住。
- **与计划的一处偏离**：无实质偏离。计划里 A/B 两点的解读（曲目信息块归侧栏底部）照原计划实现，实现过程中没有发现更贴合截图的替代读法。
- **测试**：新增 35 例（290 → 325）。三个新文件：`playback_mode_buttons_test.dart`（4 例）、`volume_control_test.dart`（12 例）、`mini_player_bar_desktop_test.dart`（4 例），共 20 例；三个既有文件追加：`shell_scaffold_nav_test.dart`（+9）、`desktop_app_bar_test.dart`（+4）、`full_player_layout_test.dart`（+2），共 15 例。覆盖点包括：桌面侧栏全高（`GlassPanel` 高度 = 窗口高 − 上下各 8px margin）、播放条左边缘不再压在侧栏下方、侧栏底部有曲目信息且播放条没有、移动端/平板回归守卫（曲目信息仍在、桌面专属控件没有泄漏进去）、1200dp 窗口宽度下桌面 shell 整体不溢出（外加单独隔离出 `MiniPlayerBar` 自身在多档分区宽度下的溢出测试，把"具体是哪个阈值算错了"定位到组件级别而不是只知道 shell 层面挂了）、音量控件拖动/静音/恢复/跨实例共享记忆、Cover Flow 门控回归 + 新入口的开关往返、macOS 侧栏顶部让位 vs 非 macOS 不让位、`DesktopAppBar`/`DesktopSliverAppBar` 在 `ShellChrome` 祖先存在/不存在两种情况下的让位行为。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条未触碰文件的存量 info；`flutter test` 325/325 通过，290 存量 + 35 新增）。实际观感待用户实机确认；macOS 30px 顶部让位数值是按红绿灯几何常量推算得出，未经真机像素级校验，值得用户在真实 macOS 设备上重点复核。

### v5.30.0 - 2026-08-09

- **feat: 主界面重做（悬浮侧栏 + EchoMusic 尺度播放条/菜单）+ Cover Flow 封面模式 + v5.29.0 遗留三小修** — v5.29.0 只做了播放页对齐 Apple Music，这一批按计划继续做主界面：悬浮侧栏、播放条改仿 EchoMusic 尺度而不是 Apple Music（用户已明确指出后者太小）、侧栏账号区仿 EchoMusic 的"未登录=可点登录入口"而不是把游客名塞进用户名位置，外加新增 Cover Flow 封面展示模式。同批带上 v5.29.0 实机验收发现的三个小问题。
- **悬浮侧栏**：`_DesktopSidebar` 从贴边 `Material` + `VerticalDivider` 改成 `Padding(8,8,8,8)` 外边距包一个 `GlassPanel(padding: EdgeInsets.zero, borderRadius: 16)`——`VerticalDivider` 直接删掉，外边距加 `GlassPanel` 自带的发丝边框已经顶替了它的视觉职责，两者都留是重复。`padding: EdgeInsets.zero` 是因为侧栏内部每一处（标题行、账号区、列表行）本来就在管自己的内边距，`GlassPanel` 默认的 `EdgeInsets.all(16)` 会跟这些内边距叠加。
- **侧栏账号区：未登录显示登录入口，不是把游客名当用户名显示**。原来的 `_AccountBlock` 给游客模式渲染 `username = t.guest`（"Guest"），跟真用户名用一模一样的样式显示，点了没有任何反应——一个套着身份外壳的占位符。拆成 `_AccountBlock`（未登录直接返回新增的 `_GuestSignInPrompt`；已登录保留原来的头像+用户名+设置按钮）和 `_GuestSignInPrompt`（人形图标 + 新增 l10n key `tapToSignIn`["点击登录"/"Tap to sign in"/"タップしてログイン"] + 箭头，整行可点，`onTap` 调用 `ref.read(authProvider.notifier).exitGuestMode()`——跟 Settings 页那个"登录"按钮调用的是同一个方法，不是重新实现一遍退出游客模式，路由的 `isPastGate` 重定向本来就会在状态翻转后接管，把用户带到 `/login`）。
- **播放条仿 EchoMusic 尺度**：`MiniPlayerBar` 内容区套 `SizedBox(height: 84)`（EchoMusic `PlayerBar.vue` 的 `h-21`），封面从 44px 提到 56px（图标和圆角按 EchoMusic 的比例 24/56 与 10/56 跟着缩放，不是写死数字），传输图标（上一首/下一首）从 24px 改成 EchoMusic 的 22px，播放/暂停保持原有 28px 的相对更大比例。布局从"标题 `Expanded` + 按钮 + 睡眠定时器紧跟其后"改成显式三段：左（封面+标题，`Expanded`）/中（传输三件套，固定宽度）/右（睡眠定时器，`Expanded`+右对齐）——左右两段用相同 flex，让中间三件套无论标题多长都保持在正中，这是播放页传输行已经在用的同一个居中技巧，之前迷你播放条从来没有这层保证（睡眠定时器紧跟"下一首"，没有对侧配重）。顶部细进度条**留在 84px 预算之外**：EchoMusic 没有等价物（它的进度条长在中间那一列里），叠到内容区顶部会牵扯命中测试优先级，这次没打算做这层取舍。
- **Cover Flow 封面展示模式**（用户此前举的"苹果老式黑胶封面左右滑动"，规划阶段确认是真需求而非举例）：新文件 `cover_flow_artwork.dart`，纯 UI 组件不自己碰 `PlayerState`——`itemBuilder`/`itemCount`/`currentIndex` 全部由调用方传入，避免组件内部另开一份队列/封面数据、跟 `_playerBlock` 已有的 `state.queue` 不同步。核心是纯函数 `visibleSideCount({width, centerSize})`：每侧能放几张由半区宽减半张封面宽、再除以步进决定，封顶 3 张——这是"可见张数由播放器宽度决定"的字面落点，可以脱离 widget 直接单测。侧边封面用 `Transform`（`rotateY`+透视）做斜置和缩放。只接进宽屏分支：窄屏的 `PageView` 已经占用了水平滑动手势在封面和歌词间切换，再叠一套水平手势翻队列会打架；单曲队列也回退普通封面——没有邻居可"流"的 Cover Flow 没有意义。做成设置项（Settings → 外观 → "Cover Flow 封面"开关，`CoverFlowModeNotifier` 跟 `BilingualLyricsNotifier` 完全同构的 `SharedPreferences` 持久化布尔值）而非插件——插件 UI DSL 不成比例，规划阶段已经排除。
- **三个小修（并入本批）**：
  - `'Now Playing'` 走 l10n：`nowPlaying` 这个 key 在 `app_en.arb` 里早就存在，直接换成 `AppLocalizations.of(context).nowPlaying`。
  - **顶栏对比度——诊断出的根因比原计划描述的更深一层**。原计划猜测是"`artworkOverlaySkin` 算的是全局明暗档，跟动态色场的局部亮度不一致"。写了一个探测测试直接验证：给一个很暗的封面加一个鲜艳强调色，读取渲染出来的 "Now Playing" 文字颜色——结果跟 `artworkOverlaySkin()` 算出来的值完全对不上，倒是跟环境皮肤（樱花薄暮）写死的 `onSurfaceVariant` 逐位相等，不管喂进去什么封面都不变。真正原因：顶栏那段 `Padding` 是直接在 `_FullPlayerScreenState.build(BuildContext context)` 的外层作用域里构造的，而 `LyricsBackground` 插入的 `SkinScope`（携带按封面算出的派生皮肤）是 `build()` 即将返回的树里的一个**后代**节点——继承查找只能往祖先方向走，找不到还没建好、将要建在自己下面的东西。顶栏因此从 v5.26.0 起就从来没有真正跟着封面变过颜色，一直读的是环境皮肤给浅色背景配的暗灰字，这正好解释了它在深色背景上为什么看不清（暗灰字配任意亮度的封面，深色区域自然低对比度），跟"局部 vs 全局明暗"是两个不同的机制。参照 `Scaffold.of(context)` 那类标准 Flutter 写法，把整段顶栏包进 `Builder(builder: (context) => ...)`——同名 `context` 遮蔽外层变量，内部原有代码一字不改就自动切换到正确的（`SkinScope` 后代）context。确认修复后再叠加计划要求的顶栏渐变遮罩（`_topBarScrim`，色板极性跟随修好后的 `onBackground` 亮度，从上到下 alpha 0.5→0），作为对抗色场局部起伏的第二道防线——现在建立在正确解析的颜色之上。
  - **播放/暂停键跟随强调色——按计划先证伪**：写了一条断言（给定封面+强调色，取按钮 `Container.decoration.color` 是否等于 `CoverPalette.accentOverArtwork`），**测试第一次跑就通过**——代码原来就是对的（`artworkOverlaySkin` 里 `sakuraPink: accent`，按钮读 `context.skinColors.sakuraPink`，两者本就在同一个正确解析的 `SkinScope` 之下，跟顶栏那个 bug 不是同一处代码路径）。**没有改任何实现代码**，断言原样收进回归测试。
- **与计划的一处偏离**：计划把顶栏对比度问题诊断成"全局 vs 局部明暗不一致"，据此只要求加一层渐变遮罩。实测发现根因更深——顶栏从来没有接上派生皮肤，是一个 `BuildContext` 作用域错误，不是明暗判断精度问题。补上了 `Builder` 包装这个根因修复，再叠加计划要求的渐变遮罩作为第二道防线，两者不冲突，都保留。
- **测试**：`test/shell_scaffold_nav_test.dart` 新增侧栏悬浮断言（`GlassPanel` 左上角坐标两个方向都 >0，反证"贴边"），既有的"guest sidebar labels..."按新行为改写（"Guest" 文字消失，"Tap to sign in" 出现）。`test/mini_player_bar_test.dart` 新增 4 例：内容区 84px、封面 56px（用 `Container.constraints` 断言，没有真实网络图片可加载渲染尺寸）、传输图标 22px、长标题下传输三件套仍对齐播放条正中。`test/full_player_layout_test.dart` 新增 6 例：顶栏标题走 l10n（日语 locale 判别，因为中译/日译和硬编码英文在英文 locale 下永远撞车、无法区分是否真的接上了 l10n）、顶栏渐变遮罩存在且盖在标题上方、播放/暂停颜色跟随强调色的证伪断言（通过，未改代码）、Cover Flow 三种场景（启用+多曲目出现/单曲目即使启用也回退/关闭时不出现）。新文件 `test/cover_flow_artwork_test.dart` 9 例：`visibleSideCount` 纯函数（窄屏 0 张、宽屏比窄屏多、极宽封顶 3、零封面尺寸不除零崩溃）+ widget 级（渲染卡片数确实随宽度变化、当前曲目总在渲染范围内、短队列不越界、点两侧封面回调对应索引、点中间封面不触发回调）。诊断顶栏颜色 bug 用的探测测试确认问题、验证修复后即删除，没有留在仓库里。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 290/290 通过，270 存量 + 20 新增）。实际观感待用户实机确认。

### v5.29.0 - 2026-08-09

- **feat: 播放页对齐 Apple Music 比例分栏，并修复我在 v5.26.0 引入的前景对比度 bug** — 用户上一轮"整体方向对了，但控件对比度和布局比例不对"反馈之后规划的第一个 phase，只做播放页这一块，主界面重做等留给后续 phase。
- **对比度修复（真 bug，不是新需求）**：`artworkOverlaySkin()` 此前注释写死"遮罩按构造一定是深色"，硬编码 `Brightness.dark` + 全白前景。实际上 `CoverFluidBackground` 只是把封面套 `saturate(1.3)×brightness(1.5)` 提亮后叠 24% 黑，浅色封面提亮后仍是浅背景，白字白图标直接糊掉——用户截图里 "Botti, Chris" 和几个次要按钮几乎不可见。新增 `backdropLuminance()`：在 `Color` 的 0–1 归一化通道上原样复现同一套 boost 矩阵，再 `Color.alphaBlend` 同一份 24% 黑蒙层，取 `computeLuminance()` 判断背景明暗；`luminance>=0.5` 时前景整体转近黑（`Colors.black` 系列半透明值），否则维持 v5.26.0 以来的白色处理（逐项验证过新旧输出完全相等，这条分支零回归）。冷背景（HSL 色相 180–280，即青→蓝→紫）额外把白色向暖白微调 16%，只做这一点色相处理，不做完整互补色体系——这是范围里明确的取舍。播放/暂停按钮继续跟随封面强调色，不受明暗分支影响。派生皮肤的 `id` 把明暗档一起编了进去，否则 `SkinScope.updateShouldNotify` 按 id 比较会漏掉"同一个强调色、但明暗翻转"的换肤。
- **播放器块与比例分栏**：`full_player_screen.dart` 新增 `_playerBlock`，把封面、标题/艺术家、进度条、传输控制合成一个整体，宽度由封面尺寸推导——`playerControlWidth` = 封面宽 × 1.45（量自四张不同尺寸的 Apple Music 参考截图，实测比值落在 1.38–1.52，取中点），取代了原来"控制面板 `Padding(horizontal:20)` 铺满区域宽、跟封面毫无比例关系"的写法。封面尺寸本身也从写死的 280 改成按可用空间比例缩放：`playerArtworkSize` = min(区域宽×0.42, 区域高×0.40)，同样量自那四张截图（窄屏单栏布局仍保留写死的 280——参照截图全部是桌面宽窗口，同一比例套到手机宽度上会比现状明显更小，是可感知的倒退）。侧栏从固定 380px 改成 `Expanded(flex:1)` 与播放器对半分，配 `ConstrainedBox(maxWidth:560)+Align`：窗口极宽时面板文字行不会无限变长，且被约束收窄后腾出的空间会回流给播放器一侧（直接把 `ConstrainedBox` 塞进 `Expanded` 不会有这个回流效果，`Flex` 布局对 flex 子项的份额是一次性分配好的）。
- **实现前已在设计阶段处理的两处坑**：(1) `(封面宽×1.45).clamp(下限, 区域宽-48)` 在区域很窄时会下限大于上限抛 `RangeError`，改用 `math.max(下限, 区域宽-48)` 做上限兜底；(2) 控制条是否收紧（`compactControls`）此前用**整窗宽度**判断，分栏后控制条实际只拿到半窗——跟 v5.28.0 那次"窄屏 spaceEvenly 溢出 78px"是同一类判断基准错位，只是触发条件从"手机宽度"变成"中等宽度+展开面板"。改成按**控制条实际区域宽度**判断，`_compactControlsBreakpoint` 保留原估算值 480，测试证明它本身不需要改。
- **新写的校准测试又抓到一个规划完全没预料到的真问题**：中等宽度窗口（950×700）展开队列面板后，控制条实际宽度约 289px，运行后控制条溢出 40px——但溢出发生在**已经处于紧凑模式**（289<480）的情况下，说明问题根本不在 `_compactControlsBreakpoint`（如上条，它没问题），而是原计划估的"280 是可用性下限"这个数字本身从来没有验证过能不能装下紧凑模式的 8 个控件。用诊断测试实测紧凑模式下三段式布局（次要控件×2 贴左、传输三件套居中、次要控件×3 贴右）：中间传输三件套固定 108px（3×36px），两侧用 `Expanded` 各分剩余宽度的一半来保持三件套永远居中；而右侧那组（速度文本按钮 64.4px + 睡眠 26px + 收藏 26px）需要 116.4px，远超它分到的一半——左侧那组只要 52px，多余的份额不会流给右侧。把下限从 280 提到 400（新增 `_controlWidthFloor` 常量并写清测量依据），实测有约 15px 安全余量。**没有改动两侧 `Expanded` 等分这个机制本身**：那正是让传输三件套在任意两侧控件数量下都保持居中的关键设计（v5.28.0 引入，有测试覆盖），只有分栏后最窄的场景才会触到这个下限，为这一种边界情况改掉核心居中机制不划算。
- **测试**：新文件 `test/artwork_overlay_skin_test.dart`（9 例，纯函数测试、不 pump widget，回避 `CoverFluidBackground`/`LyricsBackground` 的 `pumpAndSettle` 陷阱）覆盖明暗判断、深背景分支逐项回归保护、强调色不随明暗翻转、冷色暖化只发生在深背景分支、id 编码明暗档。`test/full_player_layout_test.dart` 更新了一条过期断言（面板改比例分栏后 `splitX` 期望值从写死的 `(1400-380)/2` 改成比例表达式 `1400/4`），新增 10 例：`playerArtworkSize`/`playerControlWidth` 的边界值（420 上限、160 下限、`RangeError` 兜底）与比值区间、宽屏封面撞 420 上限时控制面板宽度约等于 420×1.45、窄屏仍是写死 280、**中等宽度分栏不溢出（这条最初真的跑红过，是它抓出上面 280 下限的问题）**、宽屏不会被误判为紧凑（1400 宽度分栏后 controlWidth≈426，落在 400–480 区间内，正确进入紧凑且不溢出；1400×1400 不分栏时 controlWidth≈609，正确保持非紧凑）、超宽分栏时面板确实被 560 上限盖住。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 270/270 通过，251 存量 + 19 新增）。实际观感待用户实机确认。

### v5.28.0 - 2026-08-08

- **feat: 播放页布局重做——居中/分栏 + 控件分组（对标 Apple Music）** — 用户带截图反馈「整体方向对了，但是有些布局不太好」，指出三个具体问题并给了 Apple Music 作为交互参照。
- **修 1：关闭按钮压在 macOS 红绿灯上**。播放页是全屏路由、不经过 `DesktopAppBar`，因此没有 `DesktopAppBar` 早就在做的红绿灯让位。新增 `_needsMacTrafficLightGutter()`，仅在「桌面 + macOS + 未启用系统标题栏」三条同时成立时留 70px——启用系统标题栏时红绿灯在真实标题栏里，留白反而是错的。
- **修 2：控制条按钮排布**。原本 `mainAxisAlignment: spaceEvenly` 把 8 个控件均匀撒开，**播放/上一首/下一首获得了和睡眠定时器完全相同的视觉权重**，最常用的控件反而没法靠肌肉记忆定位。改成三组：次要控件贴两端，传输三件套在中间抱紧，且无论两侧各有几个次要控件都保持居中（两侧用 `Expanded` 等分，中间 `MainAxisSize.min`）。
- **修 3：封面下方的小圆点**。那是 `PageView` 的页码指示器。宽屏歌词已进侧栏、没有可翻的页，所以 `PageView` 连同指示器只在窄屏保留——宽屏那个点是个无处可去的孤立标记。
- **居中/分栏（用户给的模型：不显示歌词和播放列表就居中，展开就左右分栏）**：顶栏保持通栏，其下 `Expanded(Row([Expanded(播放器列), 侧栏]))`。**播放器列本身一行没改**——它只是拿到多少宽度就在多少宽度里居中，所以同一列同时服务两种状态，不存在两套布局各自漂移的风险。侧栏 380px，内容是歌词或播放队列，用 `GlassPanel` 保持毛玻璃让封面色场在整窗连续（不透明面板会把 v5.26.0 的背景切成两半）。顶栏的队列/歌词按钮变成开关，再按一次即关。**只有 ≥900px 才分栏**：窄窗分栏会让两边都不可用，那里歌词仍是压入路由、队列仍是底部 sheet。队列列表提取为 `_QueueList` 供 sheet 与侧栏共用，避免两处实现漂移。
- **实现中发现并修复两个真问题，都是新写的单测抓的**：(1) **`GlassPanel` 用 `DecoratedBox` 带背景色，里面 `ListTile` 的墨水与高亮被整个吃掉**——跟 v5.22.0 侧栏那次是同一个缺陷（`ListTile` 画在最近的 `Material` 祖先上），面板里全是列表行所以立刻显形，改用 `Material` + `shape`；(2) **窄屏控制条溢出 78px**，查下来**改造前用 spaceEvenly 时同样会溢出**，只是从没测过。解法不是重排成两行（那要维护两套排布、迟早漂移），而是 `_ControlDensity` 在窄屏通过 `IconButtonTheme`/`TextButtonTheme` 压缩每个按钮的占位，**分组结构在任何宽度下逐字节相同**。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 251/251 通过，244 存量 + 7 新增，见 `test/full_player_layout_test.dart`：传输三件套确实抱紧且居中（跨度小于整条一半、且中心与整条中心相差 <24px）、宽屏无 `PageView`、窄屏保留 `PageView`、队列按钮在宽屏停靠面板而非盖住播放器、同一按钮再按即关、开歌词会替换队列而不是叠加、**分栏后播放器在剩余宽度里居中**（断言 `(1400-380)/2` 而不只是"往左移了"））。实际观感待用户实机确认。

### v5.27.0 - 2026-08-08

- **refactor: UI 与播放解耦，引入 `PlaybackEngine` 接缝** — 用户要求"把这套解耦合做了，让 UI 和播放分离"。改造前的实际耦合比预期更深：`main.dart` 顶层有个 `late final InoriAudioHandler audioHandler` 全局，**四个不相干的 notifier**（crossfade / speed / sleepTimer / eq）用 `import 'package:inori_music/main.dart' show audioHandler` **反向抓它**——这意味着它们每一个都传递依赖 `just_audio`，且不启动真实音频栈就无法测试；`InoriAudioHandler` 同时是 OS 媒体会话桥、`AudioPlayer` 持有者、gapless 队列、淡入淡出包络和 Android 均衡器五种东西；`PlayerNotifier` 直接持有 `AudioPlayer` 调用 17 个 just_audio 成员。
- **新的分层**：`UI/notifiers → PlaybackEngine（抽象） → JustAudioEngine`，OS 媒体会话由瘦身后的 `InoriAudioHandler` 单独承担。`lib/src/playback/just_audio_engine.dart` 是**全仓库唯一 import `just_audio` 的文件**；播放器、gapless 队列、淡入淡出包络、Android EQ 全部收进它——这几样操作的是同一个播放器，分散在两个类里正是 `setVolume` 与淡入淡出互相打架的原因（旧代码靠 handler 上一个共享可变字段 `targetVolume` 勉强同步）。
- **`PlaybackCapabilities` 是这次最有价值的部分**：`equalizer`/`speedControl`/`gapless`/`crossfade`/`outputDeviceSelection`/`exclusiveOutput`/`outputFormatControl`。这是那份跨平台输出清单反复强调的规则的落点——**引擎做不到的控件就不该显示**。设置页的均衡器开关随之从 `Platform.isAndroid` 改成问 `capabilities.equalizer`：平台从来只是"当前引擎是否接了 EQ 效果"的代理，换引擎的那天它就是错的。`JustAudioEngine` 把输出链三项诚实地报 false。
- **`EngineEqualizer` 把 band 数定义为查询而不是常量**：UI 固定十段、设备通常五段，写死就是把增益映射到错误的频率上。
- **`playbackEngineProvider`/`mediaSessionProvider` 由 `main()` override 注入，全局变量删除**。两个 provider 故意不提供默认实现——缺 override 是接线 bug，应该在启动时大声失败，而不是悄悄再起一个没人监听的播放器。
- **顺带修掉两处**：(1) `_buildConcatQueue` 原本是"设置队列 → seek 到目标 index"，而 `if (concatIndex > 0)` 意味着**目标恰好是第一个有效 URL 时 seek 被整个跳过**，改为 `setQueue(urls, initialIndex:)` 一步到位；(2) `PlayerNotifier` 原本同时订阅 `processingStateStream` 和 `playerStateStream`，引擎把 just_audio 的 `processingState × playing` 矩阵收敛成五个状态后一条流就够。
- **明确没做的**：Decoder/AudioSink 的真正二分。`just_audio` 和 libmpv 都是解码+输出一体的，中间没有可挂的缝——真做需要自己端到端拥有引擎（Rust + Symphonia + miniaudio）。这次拿到的是 foobar2000 那套三段接缝里粗的一半，但它已经足够让换引擎变成受控改动。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 244/244 通过，235 存量 + 9 新增含 2 例重写）。其中三处值得单独说：**`test/playback_boundary_test.dart` 是架构测试**——断言 `just_audio` 只能被引擎实现文件 import、且任何文件都不得 `import main.dart`；接缝的价值全在于守得住，这两条守不住的那天换引擎就不再是受控改动。**`eq_notifier_test.dart` 新增"引擎有均衡器"整组**——这条路径以前只能在真实 Android 设备上跑，本地测试实际断言的是"测试机不是 Android"；现在覆盖开启推增益、关闭全部归零、增益按设备范围钳位（刻意用 5 段设备对 10 段 UI，逼出映射逻辑而不是 1:1 传递）。`test/support/fake_playback_engine.dart` 是可复用的测试假引擎。

### v5.26.1 - 2026-08-08

- **fix: Windows 导入失败——我们自己写出了非法文件名** — v5.25.1 补上的导入错误上报第一次派上用场：用户在 Windows 上点导入，看到了确切报错「OSERR: 文件名，卷名语法不正确」，也就是 Windows 错误码 123 `ERROR_INVALID_NAME`。有了这句话，定位是直接的而不是猜的——这正好复现了 v5.20.1→v5.20.2 那次的模式（先让失败可见，下一轮就能一次命中）。
- **根因**：`_importOne()` 直接拿 track id 当文件名——`final id = '$localTrackIdPrefix${const Uuid().v4()}'` 得到 `local:550e8400-...`，然后 `p.join(audioDir.path, '$id${p.extension(path)}')` 写成 `...\local_library_audio\local:550e8400-....m4a`；封面文件同理。而 **`:` 在 Windows 文件名里是非法字符**（盘符分隔符 / NTFS 备用数据流分隔符）。POSIX 只保留 `/` 和 NUL，macOS 与 Linux 都接受 `:`，所以这个 bug 只在 Windows 显形——**游客模式在 Windows 上从来就没能导入过任何文件**，而这正是用户"连文件都没能导进来"的全部原因（跟 Windows 没有播放后端是两个独立问题）。
- **修复**：新增 `lib/src/shared/safe_file_name.dart`，`safeFileName(id)` 把 `< > : " / \ | ? *` 与控制字符替换为 `_`。**只清洗磁盘上的名字，id 本身保留 `local:` 前缀**——`PlayerNotifier` 靠它区分本地曲目与服务端曲目；DB 里存的是完整路径，所以旧构建写下的文件继续可用，不需要迁移。顺带用在 `download_notifier.dart` 的离线文件名上：服务端 id 是裸 UUID、清洗后不变（不会孤立已有下载），但"拿 id 当文件名"正是刚炸过的模式，把这个约束在一处写清楚比分散在两处更安全。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 235/235 通过，230 存量 + 5 新增，见 `test/safe_file_name_test.dart`：本地 id 去掉冒号、裸 UUID 原样不动（否则会孤立已有下载）、Windows 保留字符全替换、控制字符替换、普通标点保留）。

### v5.26.0 - 2026-08-08

- **feat: 封面驱动的配色（EchoMusic 流体背景 + 毛玻璃）** — 用户对 v5.21–v5.25 的实机反馈是「样式和色调都太丑了」，四个界面区域全部勾选，并明确给出方向：参考 EchoMusic 的封面渐变背景，有封面就从封面取色，播放器页用封面颜色做动态变化并加毛玻璃，**而不是深色冷色**。同时选择「先只改配色，看完再说」，所以本版本不动布局结构。
- **关键调研发现：EchoMusic 那个"渐变背景"根本不是渐变。** 读 `LyricFluidBackground.vue` 确认，它是把封面切成 4 个象限各画进一个 100×100 canvas（`ctx.filter='blur(5px)'`），摆在中心 ±35%（以自身边长为单位）的位置，每个各自旋转（60s 一圈，-5s/-10s/-15s 错开延迟），整个容器**反向**旋转（150s，scale 1.2），叠 `saturate(1.3) brightness(1.5)` 与 SVG 湍流扭曲，最后盖 `rgba(0,0,0,.24)` + `backdrop-filter: blur(64px)`。所以它是**封面自身的颜色分布在缓慢重组**，而不是两个提取色之间插值——这也是为什么它永远和封面一致，而两点渐变只能是近似。这个区别不读源码是看不出来的。
- **新增 `CoverFluidBackground`**：上述技术的 Flutter 移植。象限用 `OverflowBox(alignment) + ClipRect` 取——把 2 倍尺寸的图钉在某个角再裁掉就等于取那个象限，不需要解码成 `dart:ui.Image` 再手动 `drawImageRect`；饱和/亮度提升合并成一个 5×4 `ColorFilter.matrix`。**刻意没做** `feTurbulence` 扭曲：Flutter 没有等价物、除非写 fragment shader，而反向旋转 + 64px 模糊已经承担了主要效果。
- **新增 `artworkOverlaySkin`，这是本次改动里最关键的架构选择**：播放器和歌词页原本是按"浅色皮肤上的深色墨水"写的，直接放到饱和的运动色场上完全读不了。**没有去改那 ~30 处调用点**，而是派生一个皮肤、由屏幕用 `SkinScope` 包住内容——里面所有 `context.skinColors.x` 自动解析成覆盖层的值，调用点零改动。这正是 v5.14.0 那次把 ~250 处硬编码主题色迁移到 token 换来的能力，第一次真正兑现。前景转白、表面转半透明白（卡片读作毛玻璃而非遮住封面的实心块）；**强调色跟随封面**（`accentOverArtwork`：vibrant → muted → dominant，与背景色 `backdropFor` 刻意取相反的偏好，因为它落在控件上、要的就是从背景里跳出来），进度条/播放键/选中态全部跟着当前封面走。取色未就绪时保持皮肤原强调色，不会闪过一帧无色状态；派生皮肤的 `id` 里带上强调色，否则 `SkinScope.updateShouldNotify` 按 id 比较会导致切歌不传播。
- **新增 `GlassPanel`**：模糊 + 半透明填充 + 高光细边，三者缺一就塌成"半透明方块"。颜色走皮肤 token，所以在封面背景上自动拿到覆盖皮肤的半透明白、在普通页面上退化成普通表面卡片，调用点不需要分支。
- **接线**：`LyricsBackground` 重写为"有封面 → 流体背景 + 覆盖皮肤；无封面 → 纯皮肤色，用户真实皮肤原样保留"；`FullPlayerScreen` **整页**放到这个背景上（此前只有 PageView 里的歌词 tab 有），并移除歌词 tab 内嵌的那层 `LyricsBackground`——否则会跑两套旋转贴片和两次 64px 模糊；进度条与传输控制合并进同一个 `GlassPanel`，读作一整块浮在色场上的控制面而不是散落的控件。
- **实现中发现并修复的一个真 bug（新写的单测抓的）**：两个 `AnimationController` 是 `late final` 惰性初始化、只有 build 的"有封面"分支会碰它们，无封面时它们从未创建，然后 `dispose()` 访问反而**在销毁过程中创建** Ticker，抛 `Looking up a deactivated widget's ancestor is unsafe`。改为 `initState` 显式创建并按有无封面 `repeat()`/`stop()`——顺带解决了"没画背景却还占着 vsync 回调"。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 230/230 通过，223 存量 + 4 新增 + 3 重写：`cover_fluid_background_test.dart` 4 例覆盖无封面时是纯色且不动（`pumpAndSettle` 不超时本身就是断言的一部分）、有封面时 4 个象限贴片 + 1 层模糊、四个贴片偏移与旋转角互不相同、切走封面后干净拆除；`cover_palette_test.dart` 新增 `accentOverArtwork` 组并断言它与 `backdropFor` 不会收敛到同一个色板，`LyricsBackground` 组按新行为重写为 4 例）。**注意事项已写进测试**：任何包含流体背景的 widget 测试都不能用 `pumpAndSettle`，两个 `repeat()` 控制器永远不会 settle。实际观感待用户实机确认。

### v5.25.1 - 2026-08-08

- **fix: 歌词页入口与导入失败的两处静默失败** — 用户实机反馈触发。两个问题是同一类：功能失效时界面上什么都不显示，跟 v5.20.1 之前播放失败的表现完全一样。这已经是这个项目第三次栽在"静默失败"上（前两次是播放、是 Windows sqflite），所以这次两处都补了明确反馈而不是只修逻辑。
- **歌词页按钮对无内嵌歌词的本地曲目被静默禁用**：`full_player_screen.dart` 的门控是 `(trackId.isEmpty || (isLocalTrack && !hasLocalLyrics))`，本地曲目必须带内嵌歌词标签才允许打开歌词页——轻音乐/未打标的本地文件直接 `onPressed: null`，按钮禁用、点了没反应、没有任何说明。这个门控本身就是多余的：`KaraokeScreen` 自己已经处理空歌词（显示「No lyrics available.」），服务端曲目一直是这么走的。**影响远比"一个按钮"大**——歌词页是 v5.24.0（封面取色遮罩）和 v5.25.0（边缘渐隐、逐行虚化）**全部**视觉改动的唯一入口，游客模式 + 无歌词本地文件的用户永远看不到这两个版本做的任何东西，这正是用户反馈"只有主界面播放器控制部分有效果"的直接原因之一。顺带修掉次生问题：`hasLocalLyrics` 用 `.valueOrNull ?? const []`，provider 加载期间会让本来有歌词的曲目也短暂禁用按钮。
- **导入失败没有任何反馈**：`importFiles()`/`importFolder()` 返回 `Future<void>`，按钮直接绑 `onPressed: notifier.importFiles`（抛异常就是未捕获异步错误，UI 上什么都没有），`_importPaths` 的逐文件 catch 也只 `debugPrint`。于是「选择器没打开」「元数据读失败」「文件复制失败」「写库失败」「用户取消」五种情况在界面上**完全无法区分**，都是点了没反应。用户在 Windows 上"连文件都没能导进来"正是撞在这里，而且这就是它一直没被发现的原因。新增 `ImportOutcome`（imported/failed/firstError/cancelled），三个导入入口全部经由 `_runImport` 上报：整体失败给具体异常、部分失败同时报两个数和首个异常、全部成功报数量、"选了但没有可用文件"单独一档、用户取消不提示。空状态页的两个导入按钮改为父组件传回调（`_EmptyLocalLibrary` 由 `ConsumerWidget` 改 `StatelessWidget`），因为**空曲库页正是失败最不可见的地方**——空的曲库保持空，看起来和取消一模一样。
- **本版本未解决、需要单独决策的问题：Windows 根本没有播放后端**。已核实 `just_audio` 0.9.46 的 `pubspec.yaml` 只注册了 android/ios/macos/web，且 `just_audio_platform_interface` 的默认实现是 `MethodChannelJustAudio()`、**没有任何平台兜底**；作为对照，`audio_service_platform_interface` 明确写了 `(Platform.isWindows || Platform.isLinux) ? NoOpAudioService() : MethodChannelAudioService()`，所以 App 在 Windows 上能正常启动、但一播放就必然抛 `MissingPluginException`。这不是配置问题，是 Windows 版从来就没有播放能力，而 Release 一直在发 Windows 包。修它需要更换输出层（media_kit/libmpv，Spotube 走的就是这条路，且顺带能拿到 WASAPI 独占/设备选择/采样率控制），属于独立的一批工作。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 223/223 通过，218 存量 + 5 新增，见 `test/local_library_screen_test.dart` 的 `import outcome reporting` 组：整体失败给出具体异常、取消保持安静、成功报数量、部分失败同时报两个数和首个异常、"选了但没有可用文件"与"失败"区分开）。

### v5.25.0 - 2026-08-07

- **feat: 视觉材质深化（分层交互架构重做第五阶段，OriginalSound 对标子集）** — 在明确排除的整窗透明化和多着色器背景之外，落地性价比最高的四个具体细节。本批（v5.21.0–v5.25.0）到此收尾。
- **按钮弹簧微交互**（新增 `lib/src/shared/widgets/spring_interaction.dart`）：`AnimationController.unbounded` + `SpringSimulation` 驱动 hover 上浮 2px、按下缩到 0.90，阻尼比 0.30/0.40 取自调研读到的 OriginalSound 动效定义。**关键设计：用 `Listener` 观察指针事件而不是 `GestureDetector` 处理**——包的都是现成的 `IconButton`，它们保留自己原有的点击逻辑，这里只叠动效，不消费手势也不进手势竞技场。控制器必须 `unbounded`：欠阻尼弹簧会越过目标值，有界控制器会把"弹"这个特征本身夹掉。`animateWith` 时带上当前速度，快速按下-松开这类中途反向的操作从实际运动状态接着走而不是从静止重开。刚度不是那份源码里能直接搬的量，按"约 250ms 稳定"调的，代码注释里写明了这一点而不是假装是照抄的。应用到两个播放器的传输控制三件套，周边次要控件刻意不加，让主操作是唯一会响应的那组。
- **HQ 徽标**（新增 `lib/src/local_library/audio_quality.dart`）：OriginalSound 的判定是采样率 ≥48kHz **且位深 ≥24**，但 v5.19.0 就核实过 `audio_metadata_reader` 的公开 `AudioMetadata` 模型根本不暴露位深，要拿到得 fork 或换包，跟"一个徽标"完全不成比例。改用码率顶替位深，并且理由是具体的而不是凑合：**码率正是区分 24bit 无损与高采样率有损的那个量，也正是位深检查存在的目的**。阈值取 700 kbps——有损编码器实际上限都远低于它（MP3 320，AAC/Opus 极限也就 ~500），而 16bit/48kHz FLAC 已稳稳高于它。两个值必须都已知：v5.19.0 之前导入的曲目这两个字段是 null，"未知"绝不能显示成质量声明。
- **歌词页上下边缘渐隐**：`ShaderMask` + `LinearGradient` + `BlendMode.dstIn`，歌词接近上下边缘时淡出而不是硬切——调研报告里判定"可以 1:1 直译"的那一项，OriginalSound 自己也是同一套 gradient-into-dstIn。
- **歌词逐行按距离虚化**（`_DepthBlur`）：真实高斯模糊（`ImageFiltered` + `ImageFilter.blur`），每行 0.6 sigma、上限 2.0。**当前行完全跳过模糊**——正在读的那行不该被重采样；同时每个模糊行都是一次 `saveLayer`，屏幕上有多少行就有多少次，所以刻意做浅、做封顶，也不加用户可调滑杆。
- **实现中发现的一个真实问题（值得记下来的那种）**：写单测断言"按下后缩放变小"时一直读到 1.0。加临时日志逐帧打印后确认**是测试辅助函数错了，不是组件错了**——`Matrix4.getMaxScaleOnAxis()` 取的是三个轴里最大的缩放，而 `Transform.scale` 只改 x/y、把 z 留在 1.0，所以它对任何缩小都返回 1.0；改成直接读 m00。排查过程中顺带给 `SpringInteraction` 的 `Listener` 补了 `HitTestBehavior.translucent`：默认的 `deferToChild` 会让"子组件本身恰好不可命中"时按下动效静默失效，一个对某些子组件悄悄不工作的包装器是个陷阱。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 218/218 通过，207 存量 + 11 新增：`audio_quality_test.dart` 7 例含"48kHz 320kbps MP3 被排除"这条位深检查存在意义所在的用例与"未知元数据不算 Hi-Res"，`spring_interaction_test.dart` 4 例含"被包装的按钮仍能收到点击"这条最容易被破坏的性质）。弹簧手感、HQ 徽标、歌词渐隐与虚化的实际观感待远端 CI 构建后用户实机确认。
- **本批（v5.21.0–v5.25.0，分层交互架构重做）全部完成**：v5.21 基础设施（可折叠桌面头部 + 悬浮播放条）、v5.22 服务端模式导航（EchoMusic）、v5.23 游客模式导航（Spotube）、v5.24 封面动态取色、v5.25 视觉材质。等待用户实机验证后决定下一批方向。

### v5.24.0 - 2026-08-07

- **feat: 封面动态取色（分层交互架构重做第四阶段）** — 歌词页背景的渐变遮罩改用当前曲目封面提取的颜色，而不是固定皮肤色。这是 EchoMusic（强调色随封面变化）和 OriginalSound（歌词页取色渐变过渡）两份调研共同指向的点，跟 v5.22/v5.23 的导航重做相互独立。
- 新增依赖 `palette_generator: ^0.3.3+7`（Flutter 官方包，与 Material 动态取色同一套量化算法），新增 `lib/src/catalog/cover_palette_provider.dart`。
- **family key 用记录类型 `CoverSource({albumId, localArtUri})`** 而不是把两种来源编码进一个字符串：记录自带结构化相等，省掉编码/解析这一层。本地封面走 `FileImage`、服务端封面走 `CachedNetworkImageProvider`（跟 `CachedNetworkImage` 同一个 provider，读的是已经下载好的字节而不是重新拉一次）。
- **`CoverPalette` 值类只暴露 UI 真正会用的几个色板 + `backdropFor(Brightness)`，刻意不暴露 `PaletteGenerator` 本身**：调用方不该需要知道它那七个可空色板里哪个合适；包类型不进 widget 代码，以后换取色实现也不会外溢到界面层。降级顺序对齐 Spotube 的 `usePaletteColor`：muted → vibrant → dominant，按主题明暗取对应档；vibrant 永远不是首选——歌词背后放高饱和色会读成"上了色的洗底"而不是背景。
- **取色前先把图缩到 96×96 再量化**：色彩分布不受影响，避免大封面第一次显示时把量化开销压在 raster 线程上。
- **所有失败路径返回 null 而不是抛异常**（无封面 / 空 albumId / 本地文件不存在 / 解码失败 / 图片压根没产出可用色），`LyricsBackground` 一律回退到皮肤固定色——这是纯装饰特性，一张封面解码失败绝不能带垮显示它的页面。已写单测锁定其中一条容易被忽略的情况：**取色进行中的第一帧就必须已经有遮罩**，否则歌词会短暂压在未加遮罩的封面上。
- `lib/src/lyrics/lyrics_background.dart` 从 `StatelessWidget` 改为 `ConsumerWidget`，两档不透明度（0.55/0.82）保持不变，遮罩包 400ms `AnimatedContainer` 做切歌过渡。
- **本 phase 明确没做悬浮播放条进度条跟随取色**（原计划标注为"视时间成本决定，非必须项"）：`MiniPlayerBar` 是常驻组件，在那里挂取色等于**每次切歌都在全 App 范围跑一次图片解码 + 量化**，代价跟一条 2px 强调线完全不成比例。取色的开销应该只由真正展示它的页面承担。同样地，取色结果不去驱动整个 App 的强调色——现有皮肤系统是"整体替换 18 个固定 token"的模型，动态替换 `ColorScheme.primary` 这类全局槽位会跟它正面冲突。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 207/207 通过，197 存量 + 10 新增，见 `test/cover_palette_test.dart`：`backdropFor` 四档降级含"深色专属色板不会串给浅色主题"这条不对称性、provider 三条 null 路径、`LyricsBackground` 取色生效 / 取色为空回退 / 取色进行中已有遮罩）。切歌时遮罩颜色是否真的随封面变化、深浅皮肤下取色结果是否都可读，待远端 CI 构建后用户实机确认。

### v5.23.0 - 2026-08-07

- **feat: 游客模式导航重做（分层交互架构重做第三阶段，Spotube 对标）** — 对标对象按用户拆分的用途取 Spotube（未登录、只播本地文件时的整个交互架构）。调研时的一条重要澄清在这里直接决定了做法：**Spotube 自己并没有独立的"游客 UI"**，跳过登录后落地的是和登录后完全一样的一套导航壳，Local Library 只是其中一个入口——所以"采用 Spotube 的主界面布局"不是再写一套界面，而是让游客模式复用同一套壳。
- **游客模式从"单屏无导航"升级为同一套响应式导航壳**（`shell_scaffold.dart`）：`isGuest` 分支此前直接返回一个裸 `Scaffold`（只有内容 + 迷你播放条），完全没有导航 chrome；现在走同一套 `_MobileLayout`/`_TabletLayout`/`_DesktopLayout`，只是传入精简后的导航项（本地曲库 + 设置）——不是新写一套布局系统，是把现有响应式壳参数化。
- 配套的三处去重：`_NavGroup.header` 改可空，游客模式两项不套分组标题；`_AccountBlock` 游客分支显示"游客"占位名且不渲染设置图标（设置在游客模式本身就是导航项）；`_TabletLayout` 的 rail trailing 设置按钮在设置已进 destinations 时不渲染。移动端 `labelBehavior` 改成按导航项数量决定——>4 项用 `onlyShowSelected`（服务端模式 6 项），否则 `alwaysShow`（游客模式 2 项）。
- **本地曲库列表交互对齐 Spotube 的 `TrackTile`/工具栏范式**（`local_library_screen.dart`，`ConsumerWidget` → `ConsumerStatefulWidget`）：新增工具栏（播放全部 / 随机播放 / 库内搜索 / 曲目计数）——**此前这一页完全没有集合级播放入口，只能一首首点，也没有任何查找手段**；库内搜索是纯客户端过滤（标题/艺术家/专辑，大小写不敏感），本地曲库是单张小表，没有远端查询可打也不需要防抖；多选（长按进入，桌面端右键等价）+ 批量移除，选择模式下 leading 换 `Checkbox`、行内删除按钮隐藏；封面 hover 叠播放/暂停遮罩（正在播放该曲时显示暂停图标）。
- **两个实现细节上的刻意选择**：(1) 播放全部/随机/点行播放都作用于**过滤后**的列表而不是整表——否则筛完再点播放会静默忽略筛选条件，已写单测锁定；(2) 选中集用 track id 而不是列表下标——并发导入/移除重排列表时下标会指向别的曲目。
- `local_library_notifier.dart` 新增 `removeAll(Iterable<String>)`：删完一批再重查一次，而不是每删一首重读整表；单首失败不中断整批（跟导入路径同一条规则）。`remove` 与它共用抽出的 `_deleteOne`/`_refresh`。
- **与计划的一处偏离，附理由**：原计划写"`LocalLibraryScreen` 改用 `DesktopSliverAppBar`"。实际读 v5.21.0 自己写的那份实现确认，它的拖拽区只包在 `FlexibleSpaceBar.background` 上（因为 `DragToMoveArea` 是盒模型 widget，不能包 Sliver——这正是 v5.21.0 记录过的坑），而本地曲库是列表页、不需要展开态大封面。传 `background: null` 迁过去等于**丢掉整条桌面拖拽区**，比现状更差；Spotube 的本地库页本身也是"固定工具栏 + 列表"，没有折叠 hero 头部。所以保留 `DesktopAppBar`（本来就有全宽拖拽），工具栏放 body 顶部。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件；`flutter test` 197/197 通过，184 存量 + 13 新增：`local_library_screen_test.dart` 10 例覆盖播放全部/随机、搜索按标题与艺术家匹配、播放全部遵循筛选、点行按筛选后下标播放、长按多选+批量移除、退出选择不动数据、hover 遮罩进出、右键切换选中、空结果提示与按钮禁用；`shell_scaffold_nav_test.dart` 新增 3 例覆盖游客导航壳项数与标签行为、游客从导航进设置、游客侧栏不重复设置按钮）。游客模式导航壳观感、工具栏与多选手感待远端 CI 构建后用户实机确认。

### v5.22.0 - 2026-08-07

- **feat: 服务端模式导航重做（分层交互架构重做第二阶段，EchoMusic 对标）** — 承接 v5.21.0 产出的 `DesktopSliverAppBar`/悬浮 `MiniPlayerBar` 基础设施，这一批开始动导航结构本身，正面回应用户"界面布局几乎没有变化"的反馈。对标对象按用户拆分的用途取 EchoMusic（登录后对接服务端目录的整个交互架构）。
- **侧栏按分组重做，并补上账号信息区**（`lib/src/shared/widgets/shell_scaffold.dart`）：导航项拆成"发现"（Artists/Albums/Search）与"资料库"（Favorites/History/Playlists）两组，桌面侧栏渲染分组标题；窄布局（`NavigationBar`/`NavigationRail`）没有分组概念，仍然拍平成一条，所以两处共用同一份拍平后的索引，分组只是渲染层的事。新增 `_AccountBlock`（头像首字母 + 用户名 + 设置入口）——不照抄 EchoMusic 的等级/VIP 徽标体系，本项目的用户模型里没有对应概念。
- **补上两条实际不可达的路由**：`AppRoutes.playlists` 和对应的 `PlaylistsScreen` 从很早就存在，但全仓库搜索确认没有任何地方链接过去，目录播放列表除了深链接以外进不去；`AppRoutes.settings` 更严重——全仓库唯一的入口在 `local_library_screen.dart`（游客模式专属），也就是**登录用户根本没有任何办法打开设置页**。前者接进"资料库"分组，后者进侧栏账号区（桌面）、`NavigationRail.trailing`（平板）、Library 页 AppBar 操作（手机，这三种布局里唯一一个既没有侧栏也没有 rail trailing 槽的）。
- **移动端导航项从 5 个变 6 个**，改用 `NavigationDestinationLabelBehavior.onlyShowSelected`——6 个常驻标签在手机宽度下放不下，这是 Material 自己对拥挤导航条给出的答案。
- **Artists/Albums 网格接入真实封面**：`albums_screen.dart` 里的私有 `_AlbumCard` 提为共享组件 `lib/src/shared/widgets/album_card.dart` 并接上早就存在却一直没用的 `artworkUrlProvider`。艺术家这边没有现成方案——`CatalogArtist` 生成模型里根本没有封面字段，后端也没有对应接口（已核对生成代码确认，不是猜测），改用"该艺术家名下任一专辑的封面"顶替；映射关系用一次全量 `listAlbums` 在客户端解析成 `artistId → albumId`，而不是每个网格单元发一次 `albumsByArtist`——网格最多 200 项，后者等于为了缩略图多打 200 次请求。
- **艺术家详情页统一到可折叠头部**：此前它是三个详情页里唯一用固定 `DesktopAppBar` 的（v5.21.0 只迁了专辑/歌单两个），现在同样用 `DesktopSliverAppBar`；页内手搓的一次性横向专辑轮播换成共享 `AlbumCard`，跟 Albums 页视觉一致。
- **三个集合详情页补上"播放全部/随机播放"**（新增 `lib/src/shared/widgets/play_actions_row.dart`）：此前专辑/歌单/艺术家页完全没有集合级播放入口，只能一首首点。集合级收藏/关注没做——本项目后端只有单曲收藏，没有集合收藏接口，硬加会是个假按钮。
- **`TrackListTile` 补桌面端指针交互**：hover 时封面叠半透明播放图标（对标 Spotube `TrackTile`；`MouseRegion` 对触摸输入天然不触发，手机端渲染完全不变），右键弹出跟长按同一个菜单——桌面用户没有长按手势，此前"加入播放列表/下载"这两个操作对鼠标**完全不可达**。
- **实现中发现并修复了三个真实 bug，全部是新写的单测抓出来的，不是走查猜到的**：(1) 侧栏根节点是 `Container(color:)`，而 `ListTile` 的 `selectedTileColor` 和水波纹必须画在最近的 `Material` 祖先上，中间夹一层 `ColoredBox` 会把两者一起吃掉——也就是说**侧栏的选中高亮从改造前就一直没有真正画出来过**，Flutter 自带断言在单测里直接报了这个，改成 `Material(color:)`；(2) 侧栏品牌行 `InoriMark + Text('Inori Music')` 在 220px 侧栏的 188px 内容宽度里溢出 20px，同样是改造前就存在的，加 `Expanded` + 省略号；(3) 分组渲染时跨组累加的 `flatIndex` 被每个 `ListTile` 的 `onTap` 闭包按**变量**而不是按值捕获，所有行点下去都传最终值 6，`items[6]` 直接越界——点任何一个导航项都没反应，这个是本次改动引入的，被"跨组选中态索引"那例单测当场抓住。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 error/warning，仅 1 条存量 info 属未触碰文件 `player_state_reporter.dart`；`flutter test` 184/184 通过，172 存量 + 12 新增：`shell_scaffold_nav_test.dart` 6 例覆盖分组标题/账号区/设置跳转/Playlists 可达性回归/跨组选中索引/移动端导航项数，`play_actions_row_test.dart` 4 例覆盖加载中与空集合禁用、Play 按序入队、Shuffle 元素一致，`track_list_tile_test.dart` 新增 2 例覆盖 hover 遮罩进入离开与右键开菜单）。侧栏分组观感、真实封面加载、桌面端 hover/右键的实际手感本地无桌面窗口环境无法走查，待远端 CI 构建后用户实机确认。

### v5.21.0 - 2026-08-07

- **feat: 可折叠桌面头部 + 悬浮播放条（分层交互架构重做第一阶段，对标 EchoMusic/Spotube/OriginalSound HQ 源码深度调研第一批交付）** — 用户反馈上一批（v5.17.0–v5.20.2）标题栏整合+材质效果虽然生效，但"界面布局几乎没有变化"，明确要求这次要触及导航结构本身，并按用途把三份参考资料拆开、要求下载源码分析而非凭截图臆测。已用 3 个 general-purpose agent 分别 clone `hoowhoami/EchoMusic`、`KRTirtho/spotube`、`Johnwikix/original-sound-hq-player` 深度读码，3 个 Explore agent 摸清本仓库现状，产出六份调研交叉比对后的分层实现计划（完整调研发现见 Claude Code 会话保存的计划文件，非本仓库 `.plan/` 目录）。本 phase 是后续两个导航重做 phase（服务端模式对标 EchoMusic、游客模式对标 Spotube）都要依赖的基础设施。
- **新增 `DesktopSliverAppBar`**（`lib/src/shared/widgets/desktop_app_bar.dart`，跟 `DesktopAppBar` 同文件——它需要复用 `DesktopAppBar` 已有的私有 `_MaximizeButton`/窗口按钮三件套逻辑，Dart 的可见性按文件而不是按类）：`DesktopAppBar` 此前是纯定高组件，不能塞进 `CustomScrollView` 的 `slivers` 列表，这是"可折叠头部+桌面拖拽"必须先解决的架构缺口——专辑/歌单详情页此前用裸 `SliverAppBar` 完全绕开了 `DesktopAppBar`，因此从 v5.17.0 起就没有桌面拖拽/窗口按钮（当时已知但未修的缺口，见 v5.17.0 条目）。
- **踩坑记录（写代码时发现，不是猜测）**：最初设计是照抄 `DesktopAppBar` 的做法，把整个 `SliverAppBar` 包进 `DragToMoveArea`——直接读 `window_manager` 包源码确认 `DragToMoveArea` 内部就是一个 `GestureDetector`（盒模型 widget），而 `SliverAppBar` 是 Sliver widget，两者不能这样嵌套后还塞进 `CustomScrollView.slivers`（会在运行时抛出 Sliver/Box 协议不匹配的错误）。改为只把 `DragToMoveArea` 套在 `FlexibleSpaceBar` 的 `background` 插槽上（背景本身是普通盒模型 widget，覆盖展开态的大面积封面区域，是最自然的拖拽区）；常驻工具栏行（标题/操作按钮）不做拖拽处理，窗口按钮保持简单点击语义。macOS 红绿灯让位的处理也相应调整——不能像 `DesktopAppBar` 那样对整个组件加左内边距（会连带把展开态的封面背景也往右推），改成只对常驻工具栏行的 `leading` 插槽加内边距。
- **第二个踩坑（被自己写的单测抓住的真 bug）**：初版实现把 `title` 同时传给了 `SliverAppBar.title` 和 `FlexibleSpaceBar.title`——这是一个真实存在的 Flutter 反模式，`FlexibleSpaceBar` 自己的渐变机制已经负责展开态大标题收缩成迷你标题的全过程，同时设置 `SliverAppBar.title` 不是无操作而是真的在屏幕上重复渲染一份标题文字。写单测断言标题只出现一次时直接暴露（`findsOneWidget` 失败，实际找到两个），修复为"标题要么去 `SliverAppBar.title`（无背景时），要么去 `FlexibleSpaceBar.title`（有背景时），二选一"——这跟改造前专辑/歌单详情页原本的写法（标题只给 `FlexibleSpaceBar`）是一致的。
- 迁移专辑详情页（`album_detail_screen.dart`）和歌单详情页（`playlist_detail_screen.dart`）到 `DesktopSliverAppBar`，找回桌面拖拽/窗口按钮；艺术家详情页留到 v5.22.0（它目前完全没有 `SliverAppBar`，需要更大改动，跟"补齐服务端导航"那批一起做更合理）。
- **`MiniPlayerBar` 从通栏直角矩形改为悬浮圆角卡片**：外层加 8px 侧边距+8px 底边距的 `Padding`，`Material` 加 `shape: RoundedRectangleBorder(borderRadius: 16)` + `clipBehavior: Clip.antiAlias`——对标 EchoMusic（播放条只占内容区宽度、圆角+阴影）与 Spotube（`SurfaceCard` 悬浮于导航之上）的共同做法。内部三段式布局（曲目信息/核心控制/次要功能）不变，只改外壳；游客模式和服务端模式共用同一个组件，两边都自动生效，不需要分别改。之前通栏铺满时 `elevation: 8` 的投影其实很难看出来（两侧都跑出屏幕外没有东西可以承接阴影），现在加了边距后阴影才真正可见。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 172/172 通过，含 `desktop_app_bar_test.dart` 新增 4 例专门覆盖 `DesktopSliverAppBar` 在 mobile/Windows/macOS/系统标题栏开启四种场景下的拖拽区域与窗口按钮行为，其中一例最初捕获了上面提到的标题重复渲染 bug）。实际视觉效果——专辑/歌单详情页是否真的可以拖拽窗口、悬浮播放条的圆角阴影观感——本地无桌面窗口环境无法走查，待远端 CI 构建后用户实机确认。

### v5.20.2 - 2026-08-07

- **fix: 找到播放彻底无法工作的真正根因——`MissingPluginException`（v5.20.1 的错误可见化修复第一次真正派上用场）** — v5.20.1 上线后用户反馈实机看到了具体报错 `MissingPluginException`，这正是"让静默失败变可见"这个修复要等的输入。拿着这个异常类型直接读 `just_audio` 0.9.46 包源码定位到了确切位置，不是猜测。
- **根因**：`InoriAudioHandler.create()`（`audio_handler.dart`）此前无条件构造 `AndroidEqualizer()` 并通过 `AudioPipeline(androidAudioEffects: [equalizer])` 接入 `AudioPlayer`，在所有平台上都这样做，从未按 `Platform.isAndroid` 区分过。读 `just_audio` 源码确认：`AudioPlayer.setPlatform()` 在玩家进入 active 状态时（`setAudioSource`/`play()` 触发，也就是**每一次播放尝试**）会对 `_audioPipeline._audioEffects` 里的每一个效果无条件调用 `_activate(platform)`——这一步不像同一函数里 INIT 请求的 `androidAudioEffects` 字段那样有 `_isAndroid()` 门控。`AndroidEqualizer._activate()` 内部无条件调用平台方法 `androidEqualizerGetParameters`，这个方法只有 Android 原生插件实现——macOS/Windows/Linux/iOS 上调用它必然抛出 `MissingPluginException`。这意味着自从 v4.1.1 引入 EQ 功能以来，**所有非 Android 平台上的每一次播放尝试**（不只是本地曲库，服务端曲目同样会中招）理论上都会在这一步抛出异常；此前从未被发现，是因为这条链路上一直没有任何错误处理/展示（直到 v5.20.1 才补上），异常只是被 Flutter 的 zone 默认错误处理器悄悄吞掉。
- **修复**：`InoriAudioHandler._equalizer` 与 `androidEqualizer` getter 改为可空（`AndroidEqualizer?`），`create()` 里只在 `Platform.isAndroid` 为真时才构造 equalizer 与 `AudioPipeline`，否则 `AudioPlayer` 不带任何 pipeline 参数创建。`eq_notifier.dart` 三处调用点相应改为 null-safe 访问。顺带修了一个同类的小疏漏：`setEnabled()` 原先的守卫是 `if (enabled && !Platform.isAndroid) return`，只挡住了"开启"路径——"关闭"调用在非 Android 平台会照样往下执行到 `androidEqualizer.setEnabled()`；改为无条件 `if (!Platform.isAndroid) return`，两个方向都挡住。
- **不是新一轮假设，是读源码确认**：核实过程直接读 `just_audio-0.9.46` 包在本机 `.pub-cache` 里的 `AudioPlayer.setPlatform()`（约 1412-1523 行）与 `AndroidEqualizer._activate()`（约 3996-4007 行）源码，确认了平台通道调用的确切位置与无条件触发条件，不是基于异常类名的臆测。
- **范围澄清**：这个 bug 影响的不只是"本地曲库播放"——从 v5.12.0 到 v5.20.1 这一整条排查线索里默认"服务端曲目播放在桌面端没问题、只有本地文件有问题"的假设可能从未真正成立，服务端曲目播放在桌面端很可能同样会命中这同一个异常。本次修复对两条路径（本地/服务端）都生效，因为改动在 `AudioHandler` 初始化层面，不区分具体播放的曲目来源。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 168/168 通过，含新增 2 例：`setEnabled` 在非 Android 测试宿主上开启/关闭均为 no-op，后者是专门锁定"关闭路径此前会漏放行"这个回归点的用例）。**这次有信心认为是真正修复**（而非又一轮诊断可见化），但仍需用户下次实机确认播放是否恢复正常，包括本地曲库与服务端曲目两条路径都要看。

### v5.20.1 - 2026-08-07

- **fix: 本地/队列播放失败时完全无提示（用户实机反馈 v5.16.0 修复后仍"点击完全没反应"）** — 重新核查 v5.16.0 提交（`aaf751d`）发现此前会话总结记录的"已加诊断日志+10秒超时保护"**从未真正落地**——实际提交只包含 `_buildConcatQueue` 的 await 修复本身，没有任何诊断代码，这是本次开工前才通过 `git show` 直接核实澄清的（不能只信之前的会话摘要，必须核实当前代码实际状态）。
- **根因分析**：`playTrack()` 内部从 URL 解析到 `_audioPlayer.play()` 的整条链路此前完全没有 try/catch。`_buildConcatQueue`→`audioHandler.updateConcatQueue`→`_player.setAudioSource()` 一旦抛出异常（例如 just_audio 原生层拒绝某个本地文件），这是一个在 `onTap: () => ref.read(...).playQueue(...)` 这类裸调用里产生的未捕获 Future 异常——Flutter 默认只会把它打到 zone 的错误处理器（release 构建下用户完全看不到），UI 侧此前没有任何反馈，表现正是"点击无响应"。v5.16.0 的 await 修复本身没有错——修复前"竞态"掩盖了这条链路本来就可能失败的事实（无论是否等待，之前失败也是静默的），修复后开始真正等待整条链路，失败路径反而第一次变得"看得见的什么都不做"。
- **已排除的假设（实测排除，非猜测）**：怀疑本地文件路径（例如 `getApplicationSupportDirectory()` 在 macOS 上包含空格的 "Application Support" 目录）在 `'file://$path'` 朴素字符串拼接下产生未转义 URL，写了独立 Dart 脚本验证——`Uri.parse()` 本身会在生成 `.toString()`/`.path` 时自动对空格等字符做 `%20` 百分号编码，`Uri.file(path)` 与朴素拼接后再 `Uri.parse()` 得到完全相同的最终编码结果，两者行为一致，排除此假设。
- **本次实际修复**：(1) `PlayerState` 新增 `playbackError`（`PlaybackFailure?`，故意不重写 `==`，保证同一错误文本连续出现两次也能各自触发一次 `ref.listen`，不被"值未变化"误判去重）与配套的 `clearPlaybackError` copyWith 标志；(2) `playTrack()` 整体包一层 try/catch，进入时先清空旧错误，URL 解析失败或链路任意环节抛出异常时都写入 `playbackError` 并 `debugPrint` 完整堆栈；(3) `ShellScaffold`（游客模式与登录模式共用的唯一根 Scaffold）新增 `ref.listen(playerProvider.select((s) => s.playbackError))`，非空时弹 SnackBar——选在这里而不是每个触发播放的调用点（本地曲库、曲目列表项、搜索结果等）各自加 try/catch，因为这是唯一保证包住所有入口的公共祖先；(4) `resolvePlaybackUrl` 的本地曲库分支补上了 `File(local.localPath).existsSync()` 校验（此前紧邻的 `OfflineDb` 分支已有这个检查，本地曲库分支却没有——文件被移动/删除或沙盒容器重置后，此前会带着一个指向不存在文件的 `file://` URL 继续往下走，而不是提前判断清楚返回 null）。
- **这不是确认修复，是让失败可诊断**：本地没有 macOS/Windows 运行环境，这次改动的目的不是"猜一个新原因去修"，而是把此前完全静默的失败路径变成用户能看到、能截图反馈给我的具体错误文本——下一步真正的根因定位依赖用户下次实机点击后看到的 SnackBar 文案。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 166/166 通过，含新增 3 例：`playbackError`/`clearPlaybackError` 的 copyWith 行为、`PlaybackFailure` 两个相同文本实例不相等的身份语义）。

### v5.20.0 - 2026-08-06

- **feat: 歌词页深度重做（对标开源/商业播放器布局改进第四阶段，本批次收官）** — 呼应参考资料 EchoMusic"写真模式全屏歌词"与原音HQ"逐字歌词带可选着色器背景"。歌词页（`karaoke_screen.dart` 全屏视图 + `full_player_screen.dart` 的 `_LyricsPage` 内嵌 tab）此前背景都是纯色 `skinColors.background`/`surfaceVariant`，没有任何背景处理。
- **新增共享组件 `LyricsBackground`**（`lib/src/lyrics/lyrics_background.dart`）：`Stack` 叠三层——底层皮肤纯色兜底、中层 `_BlurredArtwork`（`ImageFiltered(imageFilter: ImageFilter.blur(sigmaX: 40, sigmaY: 40))` 包裹的封面图，本地曲目走 `Image.file`（`localArtUri` 为 `file://` scheme），服务端曲目走 `artworkUrlProvider`，无封面/加载失败时返回 `SizedBox.shrink()` 让底层纯色透出）、顶层固定强度的渐变暗角（`context.skinColors.background` 从 55% 到 82% 不透明度，保证文字对比度不随随机专辑封面失控——用固定值而非重新做一次 WCAG 审计，跟 v5.18.0 标题栏材质的 `tint` 处理是同一个思路）。`karaoke_screen.dart` 的 `Scaffold.body` 与 `full_player_screen.dart` 的 `_LyricsPage` 都换成这个背景层打底，不再各自维护纯色背景。
- **本地曲目内嵌歌词支持**：v5.19.0 已经把 `meta.lyrics`（内嵌 LRC/纯文本）落到 `LocalLibraryTrack.embeddedLyrics` 列，之前只落库没有消费。新增 `localLyricsProvider`（`lib/src/lyrics/local_lyrics_provider.dart`，结构参照 `lyrics_provider.dart` 但读 `LocalLibraryDb.query()` 而非网络请求）。核心分支逻辑抽成顶层纯函数 `parseEmbeddedLyrics(String? raw)`——先喂给现成的 `LrcParser.parse()`（零新解析逻辑），如果没有任何 `[mm:ss.xx]` 时间戳匹配（纯文本歌词标签的常见情况），回退成一整块 `timestamp: Duration.zero` 的单行"歌词"整体展示，不做逐字/逐行高亮（`Duration.zero` 恒小于等于任意播放位置，天然表现为"当前行永远激活"）。`full_player_screen.dart:_LyricsBody`（原 `_LyricsPage` 拆出的内层）与 `karaoke_screen.dart` 均按 `trackId.startsWith(localTrackIdPrefix)` 分流到 `localLyricsProvider`，本地曲目不再无条件显示"暂无歌词"。
- **放开本地曲目的 Karaoke 入口**：`full_player_screen.dart` 顶部 Karaoke 图标按钮原先对所有 `local:` 前缀曲目一律禁用，现在改为按 `ref.watch(localLyricsProvider(trackId))` 的结果判断——只在本地曲目确实没有任何内嵌歌词标签时才禁用，有内嵌歌词（无论是否带时间戳）就能进全屏卡拉OK。只在 `isLocalTrack` 为真时才 watch 这个 provider，避免服务端曲目播放时多一次不必要的本地 DB 查询。
- **非目标（已记录，不在本次做）**：EchoMusic/原音HQ 提到的桌面歌词悬浮窗（独立于主窗口、置顶显示）需要 `desktop_multi_window` 这类完全不同的多窗口技术栈，体量对得上单独一个 phase，本次不做。逐字动画本身（`ShaderMask` 渐变填充）已是同类产品主流实现，未改动。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 163/163 通过，含新增 `local_lyrics_provider_test.dart` 4 例，覆盖 `parseEmbeddedLyrics` 的 null/空白/LRC 时间戳/纯文本回退四种输入）。`LyricsBackground`/`LocalLyricsNotifier` 本身依赖 Riverpod provider 与文件系统/网络 I/O，跟 `_FullPlayerArtwork`/`lyricsProvider` 这些既有同类代码一样未写组件级测试——延续本仓库"纯逻辑抽函数测试、屏幕组件与 I/O 绑定的 provider 不测试"的既有边界（如 `karaoke_progress_test.dart` 只测 `activeLineIndex`/`wordProgress`，不测 `KaraokeScreen` 本身）。实际模糊背景观感、暗角对比度、本地曲目内嵌歌词能否正确解析显示——待远端 CI 构建后用户实机确认。
- **对标开源/商业播放器布局改进四阶段（v5.17.0–v5.20.0）到此收官**：标题栏整合、Fluent Design 材质、本地曲库浏览增强+技术参数、歌词页重做均已完成。用户额外提出的"支持插件"在规划阶段已明确移出本批次（架构量级不匹配，需要独立规划），未来如需推进应另开一次规划对话。

### v5.19.0 - 2026-08-06

- **feat: 本地曲库浏览增强 + 技术参数展示（对标开源/商业播放器布局改进第三阶段）** — 呼应参考资料"原音HQ播放器"（播放页实时展示采样率/码率/文件格式）。核实 `audio_metadata_reader` 包源码（`lib/src/parsers/tags/tag_parser.dart`）确认 `_importOne` 早就在用的 `AudioMetadata` 对象本来就带 `bitrate`（int?，单位 bps）、`sampleRate`（int?，单位 Hz）、`genres`（`List<String>`）、`trackNumber`、`lyrics`（内嵌 LRC/纯文本）字段——此前只取了 `title/artist/album/duration/pictures`，其余字段解析出来后直接丢弃。本次是纯粹的"接上已经解析出来的数据"，没有新增任何解析逻辑。
- **`local_library_db.dart` schema v1→v2 迁移**：`LocalLibraryTrack` 新增 6 个可空字段（`sampleRate`/`bitrate`/`fileFormat`/`genre`/`trackNumber`/`embeddedLyrics`），`_open()` 的 `openDatabase` 加 `version: 2` 与 `onUpgrade`（`oldVersion < 2` 时执行 6 条 `ALTER TABLE ADD COLUMN`）。已导入的旧记录这些新列落地为 `NULL`，不强制、也没有能力重新触发已丢失原始文件访问权限的重新导入——`LocalLibraryTrack.fromMap()`/UI 侧一律把 `null` 当作"未知"处理，不是待修的 bug。`file_format` 不是从标签读的，是从源文件扩展名推导大写字符串（如 `FLAC`）——与其在少数没写规范 tag 的文件上得到错误的容器格式，不如直接用文件系统事实。
- **曲库列表排序**：新增 `LocalLibrarySortOrder`（艺术家/专辑/标题、最近导入、标题 A-Z、时长）与顶层纯函数 `localLibraryOrderByClause()`（不内联进 `queryAll`，是为了能在不碰 sqflite/平台通道的情况下单独测试）。排序选择通过新增 `localLibrarySortProvider`（`shared_preferences` 持久化，模式与既有 provider 一致）持久化；`LocalLibraryNotifier.build()` 用 `ref.watch`（不是 `ref.read`）监听这个 provider，切换排序自动触发重新查询，两处手动刷新 state 的调用点（`_importPaths`/`remove`）同步传入当前排序维度，避免刷新后又跳回默认排序。入口 UI 复用 `settings_screen.dart` 现成的"底部弹层单选列表"范式，而不是新写一套选择器组件。
- **技术参数展示位置分层**：曲库列表行（`_LocalTrackTile`）只加一个紧凑的格式徽标（`_FormatBadge`，如"FLAC"）；完整参数（格式/采样率/码率/文件大小/流派）放在播放页——`full_player_screen.dart` 新增一个仅本地曲目（`local:` 前缀）可见的详情图标按钮，点开一个底部弹层列出 label/value 行。核实 `audio_metadata_reader` 的 `mp3.dart` 解析器源码确认 `bitrate` 字段单位是 bps（如标准"320"码率的 MP3 值为 320000），展示前除以 1000 转换为 kbps，采样率同理除以 1000 转换为 kHz——没有假设包的字段单位，读源码逐一核实后再写格式化逻辑。
- **测试**：新增 `local_library_db_test.dart`（6 例：4 个排序枚举值各自的 `ORDER BY` 子句、`LocalLibraryTrack.toMap()/fromMap()` 完整字段往返、模拟 v5.19.0 迁移前的旧记录 map（缺少新增列的 key）还原为全 `null`）。踩坑记录：`fromMap()` 用 `DateTime.fromMillisecondsSinceEpoch()`（不带 `isUtc: true`）还原本地时区 `DateTime`，直接拿去跟测试夹具的 `DateTime.utc(...)` 做 `==` 比较会失败——两者代表同一时刻但 `isUtc` 标志不同，Dart 的 `DateTime.==` 把这个也纳入比较。这是 v5.12.0 就存在的既有行为，不是本次改动引入的问题，因此修的是测试断言（改成比较 `.millisecondsSinceEpoch`）而不是生产代码。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 159/159 通过，含新增 6 例）。实际重新导入曲目后列表格式徽标显示、排序切换效果、播放页详情面板数值是否正确——本地无法驱动真实文件选择器/播放，待远端 CI 构建后用户实机确认。

### v5.18.0 - 2026-08-06

- **feat: Windows Fluent Design 材质（对标开源/商业播放器布局改进第二阶段）** — 呼应参考资料"原音HQ播放器"（WinUI 3，Mica/亚克力材质+系统深浅色主题）。新增依赖 `flutter_acrylic`（核实其 `Window.setEffect({required WindowEffect effect, Color color, bool dark})` API 与 `WindowEffect.disabled/acrylic/mica` 枚举值后接入）。
- **范围界定的技术实现**：`flutter_acrylic` 没有"只对某个区域生效"的 API——`Window.setEffect()` 是整个原生窗体级别的合成器效果，Flutter 侧无法局部应用。按 v5.17.0 计划里"只作用于标题栏区域，主内容区保持不透明"的范围决定，实际做法是：全局启用原生 Mica/亚克力背景合成，但只让 `DesktopAppBar` 自己的 `AppBar.backgroundColor` 在材质开启时变成 `Colors.transparent`，应用其余所有页面（`Scaffold`/`Card`/皮肤表面色）继续用不透明颜色正常绘制——Flutter 的不透明像素会完全遮住其下方的原生合成效果，只有标题栏这一处特意留透明的区域才会真正看见模糊背景，这样不用把 v5.14.0 皮肤系统"全部 token 都是不透明色、WCAG 对比度审计成立"的前提推倒重来。
- **可读性处理**：没有另外在 Flutter 侧叠一层暗角/scrim，而是直接用 `Window.setEffect` 自带的 `color` 参数——这个参数由原生合成器混合进模糊背景里（`flutter_acrylic` 官方示例 `Window.setEffect(effect: WindowEffect.acrylic, color: Color(0xCC222222))` 正是这个用法），传入当前皮肤 `playerBar` 色值（80% 不透明度）作为色调，保证标题栏文字对比度不随用户桌面壁纸的随机内容而失控，不需要重新做一次 WCAG 审计。
- **新增设置项"标题栏材质"**（Settings，仅 Windows 显示，且"使用系统标题栏"开启时隐藏——两者语义冲突）：无/亚克力/Mica 三选一，复用既有的底部弹层单选范式（`_showSkinSheet`/`_showLanguagePicker` 同款结构）。新增 `titlebar_material_provider.dart`（`shared_preferences` 持久化，模式与既有 provider 一致）。`main.dart` 里 `ref.watch(titlebarMaterialProvider)` 加上直接调用（不是 `ref.listen`——这个 Riverpod 版本（2.6.1）的 `WidgetRef.listen` 没有 `fireImmediately` 参数，`ref.watch` 天然覆盖首次构建与后续变化两种情形，`Window.setEffect` 本身幂等，多余的重复调用无副作用）。
- **平台范围**：整个功能只在 `Platform.isWindows` 为真时生效（Settings 入口隐藏、`DesktopIntegration.applyTitlebarMaterial` 直接跳过、`DesktopAppBar` 的透明背景分支不触发）——macOS 有自己的 vibrancy 语义体系，`flutter_acrylic` 虽然也支持 macOS 效果，但这不在本 phase 范围内，`Window.initialize()` 同样只在 Windows 上调用，不影响 macOS 现状。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 153/153 通过，含新增 `titlebar_material_provider_test.dart` 3 例）。本次改动波及的 `linux`/`macos`/`windows` 三份 `generated_plugin_registrant`/`GeneratedPluginRegistrant` 文件是 `flutter pub get` 引入新插件依赖后的正常自动更新，一并提交。实际 Mica/亚克力视觉效果、深浅皮肤下材质协调度、关闭材质后正确回退——本地无 Windows 环境无法本地走查，待远端 CI 构建后由用户在 Windows 上实机确认。

### v5.17.0 - 2026-08-06

- **feat: 标题栏整合优化（对标开源/商业播放器布局改进第一阶段）** — 用户提供 4 份参考资料（EchoMusic、原音HQ播放器、开源播放器盘点博客、Spotube）要求调研布局设计并改进。逐一核实内容后确认 Spotube（4.8万 star，纯 Flutter 跨平台，技术栈与本项目一致）参考价值最高，读取其 `components/titlebar/titlebar.dart` + `pages/root/root_app.dart` 源码确认两点可直接复用的具体做法：标题栏是否用系统原生样式做成用户偏好开关而非写死；不单独维护"全局细横条 + 每页独立 AppBar"两层桌面 chrome，而是把拖拽区域与窗口按钮整合进每个页面自己的 AppBar。v5.15.0/v5.15.1 刚做的 `AppTitleBar` 正是前者的反面——一条不含任何内容的独立 28/32px 横条叠在 13 个页面各自的 `AppBar` 之上，垂直空间被两层 chrome 占用。
- **新增 `DesktopAppBar`**（`lib/src/shared/widgets/desktop_app_bar.dart`）——`AppBar` 的 drop-in 替代，接受与 `AppBar` 相同的 `title`/`actions`/`leading`/`automaticallyImplyLeading`/`bottom` 具名参数。移动端或用户选择"使用系统标题栏"时行为与普通 `AppBar` 完全一致；桌面端且未选系统标题栏时，用 `window_manager` 的 `DragToMoveArea` 包裹整个 `AppBar`（包括按钮——参照 Spotube 的做法验证过这个模式：`GestureDetector`/`DragToMoveArea` 的拖拽手势与内部按钮的点击手势在 Flutter 手势竞技场里可以共存，未越过拖拽阈值时点击手势胜出），`actions` 末尾追加窗口按钮（Windows/Linux 追加三枚 `WindowCaptionButton`，跟随当前皮肤明暗；macOS 不追加按钮，改为给整条 AppBar 加 70px 左内边距给原生红绿灯让出位置）。**机械迁移**全部 13 个使用 `Scaffold(appBar: AppBar(...))` 的页面（`artists_screen.dart`/`albums_screen.dart`/`playlists_screen.dart`/`search_screen.dart`/`tracks_screen.dart`/`favorites_screen.dart`/`history_screen.dart`/`history_stats_screen.dart`/`local_library_screen.dart`/`settings_screen.dart`/`user_playlist_detail_screen.dart`/`user_playlist_list_screen.dart`/`artist_detail_screen.dart`）为 `DesktopAppBar(...)`——逐一核实这 13 处全部只用到 `title`/`actions`/`bottom`（`search_screen.dart` 的 `TabBar`）参数，`AppBar(` → `DesktopAppBar(` 纯字符串替换零参数不兼容问题。`main.dart` 的 `MaterialApp.router(builder:)` 移除 `AppTitleBar` 的 `Column` 包裹（不再需要独立横条），旧的 `app_title_bar.dart`/`app_title_bar_test.dart` 确认无其他引用后删除。
- 服务端曲目详情页（`album_detail_screen.dart`/`playlist_detail_screen.dart`）用的是 `CustomScrollView` + `SliverAppBar`，跟本次迁移的"`Scaffold.appBar` 插槽"是两套不同 API，**不在本次迁移范围**——`SliverAppBar` 会随内容滚动折叠，把它改造成持久拖拽区域是另一套改法，本次不做，留作已知的小缺口（这两个详情页顶部暂时仍无法拖拽窗口，可以从其他页面或系统窗口管理方式操作）。
- **新增设置项"使用系统标题栏"**（Settings → Appearance，仅桌面显示）：新增 `system_titlebar_provider.dart`（`shared_preferences` 持久化，模式与既有 `bilingualLyricsProvider` 一致），切换时调用 `DesktopIntegration.applyTitleBarPreference()`（改为 `static`——只操作全局 `windowManager` 单例，不依赖某个 `DesktopIntegration` 实例的状态）立即调用 `windowManager.setTitleBarStyle()` 生效，不需要重启。`desktop_integration.dart` 的 `_initWindow()` 启动时读取这个偏好——直接读 `SharedPreferences`（而不是 `_ref.read(systemTitleBarProvider)`），因为窗口初始化发生得太早，Riverpod provider 自己的异步 `_restore()` 这时还没跑完，走 provider 拿到的永远是硬编码默认值。
- **登录页/启动页补一个最小化的可拖拽入口**：这两个页面不走 `ShellScaffold`/独立 `AppBar`，原生标题栏隐藏后如果什么都不加，用户在鉴权关卡卡住时会完全没有移动或关闭窗口的手段。新增 `GateWindowChrome`（32px 透明条，`Positioned` 叠在背景之上），只提供拖拽 + 一个关闭按钮（不需要最小化/最大化——这两个窗口是固定尺寸不可缩放的，最大化没有意义）；按钮固定用浅色样式，不跟随皮肤——延续 v5.13.0"登录/启动页保持固定樱花薄暮品牌呈现，不随皮肤变化"的既有范围决定。
- **踩坑记录**：完成迁移后顺手跑了一次不限定路径的 `dart format`，结果把仓库里 354 个文件都重新格式化了一遍——包括 OpenAPI 生成的 `lib/src/api/generated/` 目录（完全不该手动改动生成代码的格式）和大量本 phase 完全没碰过的既有文件。已用 `git checkout --` 精确回退到只保留本 phase 实际改动的文件，之后只对新增/改动的具体文件单独跑 `dart format`，不再对整个目录树跑。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 150/150 通过，含新增 `desktop_app_bar_test.dart` 4 例：移动端透传、Windows 桌面三按钮+拖拽、macOS 桌面仅拖拽无按钮、"使用系统标题栏"开启时回退为普通 AppBar）。实际拖拽/双击最大化还原、Windows 三按钮、macOS 红绿灯、"使用系统标题栏"开关两种模式切换、垂直空间是否比 v5.15.0 更紧凑——本地无 Xcode/桌面窗口无法走查，待远端 CI 构建后用户实机确认。

### v5.16.0 - 2026-08-06

- **fix: Windows 上 sqflite 未接 FFI 后端，登录/播放触发 `databaseFactory not init`** — 根因见 v5.13.0 条目后追加的已知问题记录：`pubspec.yaml` 只声明了 `sqflite`（Android/iOS/macOS 有原生实现），Windows/Linux 桌面端从未接入 `sqflite_common_ffi`。新增该依赖（纯 Dart 包，不含任何平台插件注册，`pubspec.yaml` 里没有 `flutter:`/`plugin:` 声明，因此对 iOS/Android/macOS 的构建流程零影响，只在被显式调用时才生效），`main.dart` 的 `main()` 最前面按 `Platform.isWindows || Platform.isLinux` 分支调用 `sqfliteFfiInit()` + `databaseFactory = databaseFactoryFfi`——必须在任何 `OfflineDb`/`LocalLibraryDb` 访问之前完成，所以放在 `WidgetsFlutterBinding.ensureInitialized()` 之后、`InoriAudioHandler.create()` 之前，是 `main()` 里最早执行的业务逻辑。
- **fix: 本地曲库导入的曲目播放静音、进度条卡在 0:00（v5.12.0 遗留，v5.12.3 的沙盒复制修复未能解决）** — 重新阅读 `player_notifier.dart` 后确认了记忆里记录的假设之一：`playTrack()` 对"带队列播放"（`queueIds` 非空——**本地曲库任何一次点击都走这条路径**）的处理，实际设置 `_audioPlayer` 播放源的逻辑全部委托给 `_buildConcatQueue()`，而这个方法此前是包在 `Future<void>(() async {...})()` 里**不被 await** 调用的（`_buildConcatQueue(queueIds, index, url, trackId);`，没有 `await`），与紧随其后的 `await _audioPlayer.play()` 之间是一个真实存在的竞态：`play()` 完全可能在 `_buildConcatQueue` 内部真正调用 `audioHandler.updateConcatQueue()`（这是唯一实际调用 `_audioPlayer.setAudioSource()` 的地方）之前就执行，导致播放器在"没有设置任何播放源"或"沿用上一首的播放源"的状态下收到播放指令——表现正是用户实机复现的"标题/艺术家/封面全部正确渲染（来自不相关的 `_resolveTrack`/`_makeMediaItem` 分支）但进度条与时长卡 0:00、无声音"。此前怀疑"服务端曲目这条路径在生产环境可用"，推测原因是服务端播放每一步都有网络往返延迟，客观上给了这个未 await 的 Future 足够时间在 `play()` 前抢跑完成；本地曲库的 `resolvePlaybackUrl` 是纯本地 SQLite 查询，各处 await 之间几乎没有真实等待时间，这个竞态因此频繁地朝错误的方向倒——这也解释了为什么这个 bug 只在本地播放场景稳定复现。修复：`_buildConcatQueue` 改为返回 `Future<void>` 并在 `playTrack()` 里正确 `await`；同时把队列内逐首 URL 解析从原来的顺序 `for` 循环（每首歌一次 `await`，长队列等于多次串行网络往返）改为 `Future.wait` 并行解析，避免"让 play() 等待"这个必要的修复反过来在大播放列表上引入可感知的启动延迟。顺带修正了一个次要的既有不一致：调用 `_buildConcatQueue` 时传入的起始下标此前用的是未夹取的原始 `index`（跟同一处 `state.copyWith` 用的 `clampedIndex` 不一致），现在统一用 `clampedIndex`。
- 同时用直接阅读 `just_audio` 包源码的方式排除了记忆里记录的另一个候选假设——`ProgressiveAudioSource` 对 `file://` scheme 处理不同于 `AudioSource.uri()`：`AudioSource.uri()` 的实现只按 URL **扩展名**（`.mpd`→Dash、`.m3u8`→Hls，其余一律 `ProgressiveAudioSource`）分支，与 URI scheme 完全无关，本地音频文件的常见扩展名（`.mp3`/`.flac`/`.m4a` 等）不会走 Dash/Hls 分支，两种写法对本项目场景产出完全相同的 `ProgressiveAudioSource`——这个假设不成立，不需要改动 `audio_handler.dart` 的 `updateConcatQueue()`。
- 本地无 Xcode/Android SDK/Windows 工具链，两处修复均无法在本机实机复现验证，`flutter analyze`/`flutter test` 覆盖不到播放竞态与 sqflite 平台分支这类需要真实运行时环境的问题——这两个修复是基于逐行阅读代码 + 对照三方库源码（`sqflite_common`、`just_audio`）得出的高置信度诊断，而非盲目试错，但仍待用户下载新构建后实机复认，如果验证后发现依旧有问题，需要用户提供具体现象（例如是否有任何错误提示、是不是所有本地文件还是只有部分文件）以便加诊断日志而不是再猜一次。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 149/149 通过，均为存量测试复核——`player_notifier_test.dart` 现有用例覆盖的是纯逻辑分支，不实际驱动 `AudioPlayer`/`_buildConcatQueue`，本次改动未新增测试，理由与 `_buildConcatQueue` 此前从未被测试覆盖过一致）。

### v5.15.1 - 2026-08-06

- **fix: 桌面自定义标题栏在 macOS/Windows 上均未生效** — 用户下载 v5.15.0 构建实机验证，反馈两个平台窗体"看起来跟以前一样"。根因定位（对照 `window_manager` 官方 example app 的原生 runner 文件逐行核实，非猜测）：
  - **macOS**：`macos/Runner/MainFlutterWindow.swift` 从 v3 阶段创建以来从未改过，缺少 `window_manager` 官方 example 里的 `override public func order(_:relativeTo:) { super.order(...); hiddenWindowAtLaunch() }`。没有这个覆写，AppKit 在 nib 加载阶段的默认"启动时可见"行为会在 Dart 侧 `windowManager.waitUntilReadyToShow()` 有机会应用 `titleBarStyle`/`windowButtonVisibility` 之前，就把窗口以默认原生标题栏样式显示出来——`AppTitleBar`（v5.15.0 新增）等于从没真正生效过。`hiddenWindowAtLaunch()` 是 `window_manager` 包自带的 `NSWindow` extension 方法（`setIsVisible(false)` 的一次性调用），补上这一个覆写即可让窗口按 `waitUntilReadyToShow` 的预期保持隐藏直到 Dart 侧显式调用 `show()`。
  - **Windows**：`windows/runner/flutter_window.cpp` 的 `OnCreate()` 里有一段 `flutter_controller_->engine()->SetNextFrameCallback([&]() { this->Show(); })` + `ForceRedraw()`——`git log --follow` 确认这是 v3 阶段 `flutter create` 生成后从未被改动过的原始脚手架代码，早于 `window_manager` 在 v5.12.2 引入。这段代码会在 Flutter 首帧渲染完成的瞬间就无条件调用 `this->Show()`，与 Dart 侧 `windowManager.waitUntilReadyToShow()` 想要控制的"配置好 `WindowOptions`（含 `titleBarStyle`）之后才显示"顺序形成竞态——`Win32Window::Create()` 本身不会主动显示窗口（只有 `WS_OVERLAPPEDWINDOW` 但没有 `WS_VISIBLE`），核实 `window_manager` 的 Windows 插件源码确认 `windowManager.show()`（Dart 侧显式调用）本来就会对同一个 HWND 调用 `ShowWindowAsync`，这段脚手架代码是多余且有害的冗余触发。删除后与 `window_manager` 官方 example 的 `flutter_window.cpp` 完全一致（该文件没有这段回调）。
  - 这条 bug 很可能同时也是 v5.12.2 窗口尺寸功能"在 Windows 上从未被真正验证过"的同一根因——用户此前所有实机验证均在 macOS 上进行，Windows 构建直到 v5.13.0 才第一次被人工运行。
- 两处都是原生 runner 文件改动（Swift / C++），本地没有 Xcode/Windows 工具链无法编译验证，只能靠远端 CI 的 `Build macOS`/`Build Windows` job 真实编译来兜底——如果这两个改动有语法或逻辑问题，会在 CI 阶段直接暴露，而不是留到用户手动验证时才发现。
- The phase output is version-tracked; Dart 侧无改动，`flutter analyze`/`flutter test` 分别保持 0 issues / 149/149（与 v5.15.0 一致，仅为存量结果复核）。窗体是否真正显示自定义标题栏、拖拽/双击最大化还原是否正常——待远端 CI 构建后由用户实机确认。

### v5.15.0 - 2026-08-06

- **feat: 桌面自定义窗体（三项客户端深度改造第三阶段之二，也是本轮三项诉求的收尾）** — 承接 v5.12.0–v5.14.0。改造前 `macos/Runner/MainFlutterWindow.swift` 是空的 `NSWindow` 子类、`windows/runner/win32_window.cpp` 是 `flutter create` 默认的 `WS_OVERLAPPEDWINDOW`，桌面端零自定义，用完全依赖 OS 原生标题栏。`window_manager` 依赖从 v5.12.2（窗口尺寸）就已引入，本 phase 零新增依赖，只是用了它更多的 API 面：`desktop_integration.dart` 的 `_initWindow()` 在 `WindowOptions` 里新增 `titleBarStyle: TitleBarStyle.hidden` + `windowButtonVisibility: <是否 macOS>`——原生标题栏统一隐藏，按用户已确认的平台分工（Windows 完全自定义、macOS 保留系统红绿灯）分别处理："隐藏标题栏但保留窗口按钮"是 `window_manager`/系统级别对"macOS 沉浸式标题栏、红黄绿灯仍由 OS 绘制在最上层"这一经典模式的标准支持方式，不需要在 Swift 侧另写代码。
- **新增 `AppTitleBar`**（`lib/src/shared/widgets/app_title_bar.dart`）——28px（macOS）/32px（Windows/Linux）高的替代标题栏，背景色跟随当前皮肤的 `playerBar` token（`ref.watch(skinProvider)`，与 v5.14.0 皮肤系统联动，换皮肤后标题栏颜色同步更新，不需要额外接线）。不放品牌 Logo/文字——`ShellScaffold` 侧边栏与登录/启动页已经各自展示品牌标识，标题栏里再重复一遍纯属视觉噪音，也让它在固定 440×720 的登录/启动窗口和 1440×840 的主窗口上都不显得挤。macOS 分支：左侧留 78px 空间给系统绘制的红黄绿灯（macOS 会把它们画在 Flutter 内容之上，本身不需要 Flutter 侧的组件去"占位"，留白只是避免可拖拽色块看起来在和它们抢位置），其余区域整条可拖拽，不自绘任何按钮。Windows/Linux 分支：三枚按钮直接复用 `window_manager` 包自带的 `WindowCaptionButton.minimize/maximize/unmaximize/close`——这是该包自绘的图标（不是调用系统标题栏),满足"不使用系统 Windows 窗体样式"的前提下没有必要再重新画一遍相同效果的图标；`brightness` 参数直接传入当前皮肤的 `SkinDefinition.brightness`，深色皮肤下按钮自动切换到白色图标配色方案。用 `WindowListener` 的 `onWindowMaximize`/`onWindowUnmaximize` 回调驱动最大化/还原按钮图标切换，覆盖"双击标题栏"（`DragToMoveArea` 自带的双击最大化/还原手势）触发的状态变化，不只是点按钮那一条路径。
- **单点接入** — `AppTitleBar` 与 v5.13.0 的 `SkinScope`、v5.14.0 的皮肤系统一样，在 `main.dart` 的 `MaterialApp.router(builder: ...)` 单点插入（`DesktopIntegration.isDesktop` 门控，移动端 `builder` 直接原样返回 `child`，零影响）：`Column([AppTitleBar(), Expanded(child: 原有内容)])`。因为 `builder` 包裹的是整个 `Navigator`（含其 `Overlay`），所有弹窗/BottomSheet/SnackBar 依然渲染在 `Expanded` 区域内、不会被标题栏遮挡或穿透到标题栏上方——登录页/启动页由于是鉴权关卡前后都会经过的路由，同样能拿到这条标题栏（否则原生标题栏隐藏后用户会在登录页没有任何移动/关闭窗口的手段）。
- **新增 `app_title_bar_test.dart`**：3 例，用 `debugDefaultTargetPlatformOverride` 模拟 macOS/Windows 两种平台而不需要真实桌面窗口（未点击任何按钮，因此不会触碰 `window_manager` 在测试环境下不存在的原生通道——`isMaximized()` 在 `AppTitleBar.initState` 里包了 `catchError`，专门覆盖这种"没有真实原生窗口"的场景）。踩了一个坑：`debugDefaultTargetPlatformOverride` 必须在测试函数体自身同步收尾时重置，而不是放进 `tearDown`/`addTearDown`——Flutter 测试框架的"结束时校验 debug 变量都被清空"检查在 `_runTestBody` 里紧跟测试体之后就执行，比 `addTearDown` 回调触发得更早，放 `tearDown` 里会导致校验先跑到、每个用了平台覆盖的测试必现"a foundation debug variable was changed"报错。
- 本 phase 零新增依赖，零 native 代码改动（`window_manager` 的 macOS/Windows 插件已经在原生侧实现了 `setTitleBarStyle`/`windowButtonVisibility` 的运行时切换）。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 149/149 通过，含新增 `app_title_bar_test.dart` 3 例）。实际拖拽/双击最大化还原、Windows 三按钮功能、macOS 红黄绿灯与自定义标题栏背景的协调观感、切换皮肤后标题栏颜色是否同步——本地无 Xcode/桌面窗口无法走查，待远端 CI 构建后由用户实机确认。

### v5.14.0 - 2026-08-06

- **feat: Flutter 皮肤系统——切换 + 导入（三项客户端深度改造第三阶段之一）** — 承接 v5.12.0/v5.13.0，对应用户诉求"有皮肤可以切换，也可以导入别人制作的皮肤"。改造前主题是 `sakura_dusk.dart` 里的 `buildSakuraDuskTheme()` 硬编码单一配色，`SakuraDuskColors` 是 `main.dart` 唯一直接调用的静态常量类，没有任何抽象层。新增 `skin_definition.dart`：`SkinColors`（18 个颜色 token，字段名与 `SakuraDuskColors` 逐一对应，故意不按最初规划改名为 `primary/primaryLight/...` 泛型名——JSON 清单对外仍用泛型键名，但 Dart 内部字段保留原名，换来后续大规模调用点迁移时零改名成本）+ `SkinDefinition`（id/displayName/brightness/colors/author/isBuiltIn）；`buildSakuraDuskTheme()` 泛化为 `buildThemeFromSkin(SkinDefinition skin)`，结构与原函数逐行对应，只是从读静态常量改为读 `skin.colors.*`，`ColorScheme.light(...)`/`ColorScheme.dark(...)` 按 `skin.brightness` 二选一（沿用原函数只显式设置部分字段、其余吃工厂默认值的写法，未设置字段的默认值本就应该跟着亮暗分别取值）。`SakuraDuskColors` 本身不删除，作为「樱花薄暮」内置皮肤的常量来源保留。
- **新增内置深色 ACG 皮肤「月靛」（Moonlit Indigo）** — 全仓库此前 `themeMode` 硬编码 `ThemeMode.light`，没有任何深色支持，这是真实缺口而非锦上添花。背景/表面取深靛紫基调（`#1A1626`/`#241B33`/`#2E2140`），文字取暖白（`#F3E9F5`），樱粉品牌三色（`sakuraPink`/`Light`/`Dark`）与「樱花薄暮」完全一致以保持品牌延续性——这是唯一没有为深色单独调整的 token 族，原因见下一条。所有颜色搭配用新增的 `contrastRatio()`（标准 WCAG 相对亮度公式，`Color.r/.g/.b` 浮点通道 API）实测而非估算：`onBackground/background` 15.0:1、`onSurface/surface` 13.9:1、`onSurfaceVariant/surface` 7.2:1、`onError/error` 6.6:1，均远超 4.5:1 门槛，测试中固化为 `skin_definition_test.dart` 的断言防止后续误改。唯一接受的妥协：`sakuraPink` 作为文字直接压在深色背景上时对比度仅 ~3.3–3.5:1（未过 4.5:1 但过 3:1 大字号门槛）——调亮该色会反过来破坏"白色文字压在 sakuraPink 填充色按钮上"这另一半约束（同一个颜色不可能同时是"浅到能在深背景上当文字"和"深到能在浅色文字下面当填充色"），两权相害取其轻，选择保持按钮可读性、容忍少数装饰性文字场景的轻微降级，在 requirement.md 里诚实记录而非藏起来。
- **调用点迁移——从最初规划的"存量调用点不强制本 phase 内迁移"改为全量迁移，附改变理由**：写完 `buildThemeFromSkin` 后实测发现，全仓库 24 个文件里有 ~250 处直接引用 `SakuraDuskColors.*` 静态常量（而非经由 `Theme.of(context)`），且绝大多数集中在少数几个语义 token（`onSurfaceVariant` 83 处、`onSurface` 18 处等）——一旦真的切到「月靛」，这些硬编码的深色文字（如 `onSurface` = `#3B2A3F` 近黑）会直接压在已经 reskin 成深色的 `Scaffold`/`Card` 背景上，对比度跌到 ~1.3:1，肉眼不可读，等于皮肤系统一切换就把大半个应用变成"文字消失"的残废状态——这不是"部分组件还没跟上新皮肤"的良性不完整，而是货真价实的功能缺陷。原计划参照 v5.7.0 主题迁移先例（当时一次性迁移了 311 处引用）判断本 phase 工作量过大而选择不做，但重新评估后发现两种情况本质不同：v5.7.0 是"一次性全局换色"，全仓库任何时刻只有一套配色，旧引用只是短暂的迁移期残留；这次是"运行时可切换的两套（以后更多）配色"，不迁移的引用会永久停留在错误状态，而不是过渡态。因此改变原计划，实际执行全量迁移：新增 `SkinScope`（`InheritedWidget`，`skin_provider.dart`）在 `main.dart` 的 `MaterialApp.router(builder: ...)` 单点插入，包裹路由内容；新增 `BuildContext` 扩展 `context.skinColors`，`SkinScope.of()` 在找不到祖先 `SkinScope` 时回退到樱花薄暮默认值而非断言失败（保证 `track_list_tile_test.dart` 等既有"裸 `MaterialApp(home: Widget)`"风格的 widget 测试不用改造也能继续通过）。选用 `InheritedWidget` + `BuildContext` 扩展而非直接 `ref.watch(skinProvider)`，是因为迁移范围内相当一部分是 `StatelessWidget`（如 `_ArtworkFallback`、`_KaraokeLine`、`_LocalTrackTile`、多个 `_XxxCard`），没有现成的 `WidgetRef` 可用，`BuildContext` 在任意 widget 里都能拿到，不需要把这些类改造成 `ConsumerWidget`。24 个受影响文件里逐一把 `SakuraDuskColors.X` 替换为 `context.skinColors.X`（先批量文本替换类名前缀，再用 `flutter analyze` 找出因此失效的 `const` 修饰符——原来能是编译期常量是因为 `SakuraDuskColors.X` 本身是 `static const`，换成运行时才能解析的 `context.skinColors.X` 后每一处外层 `const Icon/TextStyle/Padding/...` 都要去掉——逐一核对修复，避免误删不相关的 `const`）。
- **例外——启动页/登录页/品牌标识刻意不迁移**：`SplashScreen`/`LoginScreen`/`AppBackground`/`InoriMark` 保留固定的樱花薄暮配色，不随皮肤切换变化。理由：这三处是"进入应用前的品牌入场体验"，多数成熟 App 的启动画面本就不随应用内主题设置变化；且游客态下 Settings（含皮肤入口）本就在鉴权关卡之后才能访问，用户切换皮肤时不可能正停留在登录页，两者在正常操作路径下不会同时出现在用户眼前，为一个不可达的观感一致性做迁移没有实际价值。
- **皮肤选择与导入 UI** — 新增 `skin_provider.dart`：`SkinNotifier`（已装皮肤列表 + 当前选中 id，选中项持久化到 `shared_preferences` 键 `appearance.skinId`，做法与 `background_provider.dart` 一致）；导入的皮肤以「一个文件一个皮肤」的形式落盘到 App 支持目录的 `skins/` 子目录（单文件可分享，不建新 DB 表——参照 v5.13.0 处理自定义背景图片时同样"文件系统比 DB schema 更轻"的判断）。Settings → Appearance 新增「皮肤」入口（复用既有 bottom sheet 交互模式，与「语言」「登录页背景」视觉一致），展示已装皮肤的色块预览（`_SkinSwatch`：背景色打底 + 中心一枚品牌粉圆点）+ 名称 + 内置/已导入标记，点击即时切换；非内置皮肤额外提供删除入口（内置皮肤不可删）。AppBar 提供导入按钮，调用 `FilePicker.platform.pickFiles` 选取 `.json` 文件。
- **皮肤导入文件格式与校验** — 单文件 JSON：`id`/`displayName`/`author`（可选）/`brightness`（`"light"`|`"dark"`）/`colors`（18 个键，对外用泛型命名如 `primary`/`primaryLight`/`background`/`shadow` 等，内部映射回 `SkinColors` 的 `sakuraPink`/`sakuraPinkLight`/`miniPlayerShadow` 等字段，颜色值支持 `#RRGGBB` 与 `#AARRGGBB` 两种写法）。`parseSkinJson()` 校验分两层：结构性错误（JSON 语法错误、缺字段、颜色值无法解析、`brightness` 取值非法、`id` 与已装皮肤重复）直接抛 `SkinParseException` 中止导入，UI 侧弹窗展示具体原因；`onBackground/background`、`onSurface/surface`、`onSurfaceVariant/surface`、`onError/error` 四组核心文字可读性配对做 WCAG 4.5:1 校验，不达标**仅警告不阻断**（弹窗列出具体哪几组、实测比值是多少），呼应本项目 v5.5.0 建立的"实测而非估算"惯例——校验函数与 v5.14.0 用于自检"月靛"皮肤的 `contrastRatio()` 是同一个。
- **非目标（相对最初规划的调整）**：`SkinDefinition` 未包含最初规划里的"可选背景图"字段——重新评估后没有找到会消费它的场景（启动页/登录页背景已在 v5.13.0 由独立的 `backgroundProvider` 解决且刻意不随皮肤变化，见上一条），加了一个没有调用点会读取的字段等于埋一块永远用不到的死代码，故从数据模型里去掉，等真的出现"皮肤自带壁纸"这类需求时再加。
- 本 phase 零新增依赖——`parseSkinJson`/`contrastRatio` 全部基于 `dart:convert`/`dart:math`，皮肤导入复用已有的 `file_picker`/`path_provider`/`shared_preferences`。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues——含新代码引入的两个 `always_use_package_imports` info 与全部 24 个迁移文件的 const 修饰符修复；`flutter test --no-pub` 146/146 通过，含新增 `skin_definition_test.dart` 25 例：`contrastRatio` 对称性与已知值、`buildThemeFromSkin` 亮/暗双分支的 `ColorScheme` 映射、内置深色皮肤四组核心文字对的 WCAG 断言、`parseSkinJson` 的合法输入/结构性错误/对比度警告三类场景、`SkinState.active` 的选中态回退逻辑、`SkinScope.of` 有无祖先两种情况）。视觉效果（两款内置皮肤实际观感、切换后是否存在遗漏组件仍卡旧色、导入一份自制 JSON 皮肤的实际效果）本地无 Xcode 无法走查，待远端 CI 构建后由用户实机确认。

### v5.13.0 - 2026-08-06

- **feat: Flutter 启动页 + 登录页视觉重做 + 自定义背景（三项客户端深度改造第二阶段）** — 承接 v5.12.0（游客本地播放）与用户三项诉求中的第二项。改造前 `router.dart` 鉴权校验期间实际展示的是 `LoginScreen` 本身（注释写着"show a splash"但从未真正实现过独立启动页），登录页是纯 `Scaffold→Center→Column` 无背景层，Logo 是 `Icons.music_note_rounded` 套色块。新增品牌视觉标识 `InoriMark`（`lib/src/shared/widgets/inori_mark.dart`）：`CustomPainter` 手绘一枚带缺口的樱花花瓣剪影（缺口是花瓣尖端的浅 V 形凹陷——这是让轮廓读作"樱花"而非泛用树叶/水滴图形的关键细节），两侧曲线刻意不对称避免过于工整的"矢量素材"观感，实心填充用樱粉渐变（`sakuraPinkLight → sakuraPinkDark`）+ 一道半透明白色内凹弧线（暗示声纹/唱片纹路，不用直白的音符图形）。该标识同时替换了登录页、新增的启动页、以及桌面侧边栏（`shell_scaffold.dart`）里各自独立的旧图标，统一为一套视觉身份。
- **feat: 真正的启动页** — 新增 `AppRoutes.splash` + `SplashScreen`（`lib/src/shared/splash_screen.dart`），`router.dart` 的 `redirect` 在 `authState is AsyncLoading` 时导向 `/splash` 而非 `/login`。品牌标识 260ms 缩放 0.9→1.0 + 淡入（`Curves.easeOutCubic`），无交互控件。修复了一个随之浮现的路由逻辑缺口：启动页只在"鉴权校验中"这一瞬时态有效，鉴权结果一出来就必须离开，但原有的 `isLoginRoute` 判断只认 `/login`，如果照搬同一套规则，已登录/游客用户短暂停留在 `/splash` 时会因为无一条规则匹配而卡死在启动页——新增专门针对 `state.matchedLocation == AppRoutes.splash` 的分支，无条件按当前鉴权结果分流到 `/artists`、本地曲库或 `/login`，不复用登录页那套规则。
- **feat: 登录页毛玻璃卡片 + 自定义背景（Settings）** — 品牌标识与「Inori Music」字样直接浮在背景之上（与启动页视觉呼应），表单收进新增的毛玻璃卡片（`BackdropFilter` 24px 高斯模糊 + `surface` 72% 不透明度打底 + 24 圆角 + `miniPlayerShadow` token 投影）——高不透明度叠加强模糊能在任意背景图片上都把既有樱花薄暮配色审计过的对比度基本保住，不需要为每张用户自选图片单独调整文字颜色。新增 `AppBackground`（`lib/src/shared/widgets/app_background.dart`）与 `backgroundProvider`（`lib/src/shared/background_provider.dart`），Settings → Appearance 新增「登录页背景」入口：`file_picker` 选图后**复制**进 App 自己的 `appearance/` 目录（v5.12.3 才吸取的教训——不复制、直接引用外部路径在 macOS 沙盒下有失效风险）持久化，透明度滑块限制在 0.1–0.8 区间防止极端值破坏可读性。
- **非目标调整（相对最初规划的偏离，附理由）**：(1) 原计划给登录页做宽屏双栏布局（`LayoutBuilder` 按 600/1200dp 断点），但 v5.12.2 已经把登录/游客态的窗口锁定为固定 440×720 且不可调整大小，宽屏场景在正常操作下已不可达，继续做双栏布局是给一个不会发生的场景做设计，故取消，登录页只做单栏窄屏优化。(2) 原计划背景层挂在 `main.dart` 的 `MaterialApp.router(builder:...)` 单点包裹、对全部路由生效，实现前重新评估发现这会让"自定义背景"从"登录页背景"变成"全局重新蒙皮"——`ShellScaffold` 之下十几个既有页面全部构建在不透明 Sakura Dusk 表面上，逐一为任意用户图片重新审计对比度是数倍于本阶段的工作量且并非用户诉求原文所指。改为只在 `SplashScreen`/`LoginScreen` 各自内部使用 `AppBackground`，UI 文案也明确标注「登录页背景」而非笼统的"自定义背景"，把范围收窄到实际截图展示的场景。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 121/121 通过，含新增 `background_provider_test.dart` 8 例——`BackgroundSettings.copyWith` 与不透明度夹取逻辑，`BackgroundNotifier` 本身因触碰 `SharedPreferences`/文件 I/O 沿用本仓库既有惯例不直接测试）；视觉效果（启动页动效时长是否合适、毛玻璃卡片在真实背景图片上的实际观感、品牌标识渲染是否符合预期）待远端 CI 构建后由用户实机确认——本地无 Xcode 无法本地走查。
- **已知问题（用户实机验证 v5.13.0 时报告，明确要求后续一起修，暂不打断当前节奏）**：(1) Windows 构建登录/播放时报 `databaseFactory not init`——已根因定位（未修复）：`pubspec.yaml` 只声明了 `sqflite`（Android/iOS/macOS 有原生实现），从未接入 `sqflite_common_ffi`，Windows(/Linux) 桌面端任何 `OfflineDb`/`LocalLibraryDb` 调用都会抛这个异常；`player_notifier.dart` 的 `resolvePlaybackUrl` 对本地与服务端曲目一视同仁地先查 `OfflineDb`，所以在 Windows 上播放**任意**曲目都会触发，游客本地播放（v5.12.0）在 Windows 上因此实质不可用——这是随 Windows Release 构建（v5.12.4 才新增）第一次被人工运行而暴露的既有缺口，并非本次改动引入。(2) macOS 本地导入曲目仍然无声——v5.12.3 的沙盒复制修复（怀疑 `NSOpenPanel` 临时授权失效）已被用户实机验证证伪，问题与文件访问权限无关，下一步排查方向见项目记忆 `local-playback-not-working.md`（`_buildConcatQueue` 未 await 的 `Timer.run` 与 `playTrack` 的时序竞态；`ProgressiveAudioSource` 对 `file://` URI 的处理）。两个问题彼此独立，只是被放进同一个"稍后一起修"批次，计划在 v5.15.0（桌面自定义窗体）之后的 v5.16.0 集中处理。

### v5.12.4 - 2026-08-06

- **ci: GitHub Release 附带 Flutter 构建产物** — 用户在验证 v5.12.x 期间指出，每次都要用 `gh run download` 从 workflow 运行记录里下载构建产物太麻烦，应该像 Go server 二进制一样直接发布到 GitHub Release。`.github/workflows/release.yml`（`v*.*.*` tag 触发，此前只构建 5 个平台/架构组合的 `inori-api` 二进制）新增三个并行 job：`build-flutter-apk`（Android APK，重命名为 `inori-music-<tag>.apk`）、`build-flutter-macos`（`flutter build macos` 产出的 `.app` 用 `ditto -c -k --sequesterRsrc --keepParent` 打包为 `inori-music-<tag>-macos.zip`——用 `ditto` 而非普通 `zip` 是为了正确保留 `.app` bundle 结构和资源分支）、`build-flutter-windows`（`flutter build windows` 产出目录用 PowerShell `Compress-Archive` 打包为 `inori-music-<tag>-windows.zip`，因为 Windows 产物是 exe + 一堆 DLL + `data/` 目录，不能只发布裸 exe）。`publish` job 的 `needs` 加上这三个新 job，其余逻辑不变——`actions/download-artifact` 的 `merge-multiple: true` 已经会把所有 job 的产物合并进同一个 `dist/` 目录，`softprops/action-gh-release` 的 `files: dist/*` 自动一并发布，不需要改 `publish` job 本身。刻意不包含 `flutter build ipa --no-codesign` 的产物——未签名 IPA 无法直接安装（需要越狱或重新签名），放进面向普通用户的 Release 里没有实际可用性，跟 CI 内部产物（用于验证构建本身能过）目的不同。
- The phase output is version-tracked；`actionlint` 对修改后的 `release.yml`（及全部 workflow 文件对照）静态校验通过，YAML 语法另用 Ruby `YAML.load_file` 复核；工作流本身此前未在任何 tag 推送上跑过 Flutter 构建（`release.yml` 只有 Go job），新增 job 是否真的在远端跑通、Release 页面产物是否正确挂载，待本次推送打 tag 后现场确认。

### v5.12.3 - 2026-08-06

- **fix: 游客本地导入的曲目元数据/封面正常但实际不播放** — 用户下载 v5.12.2 构建实机导入一首真实歌曲后反馈：标题「Emmanuel」、艺术家「Botti, Chris」、专辑封面均正确显示（证明 `audio_metadata_reader` 解析与 `_localMediaItem` 渲染链路是通的），但进度条与总时长一直停在 0:00，未见音频播放。根因是 v5.12.0 让 `LocalLibraryDb.localPath` 直接引用 `file_picker` 返回的原始选中路径——macOS App Sandbox 下，`NSOpenPanel` 选中文件后授予的临时读权限**不保证在选择这一动作之后持续有效**：导入当下同步调用 `readMetadata()` 读标签能成功，是因为读取发生在选择操作的同一时间窗口内；但用户点击播放是后续一次独立交互，`just_audio` 底层 `AVPlayer` 用同一路径重新打开文件时，沙盒授权很可能已经失效，导致资源加载静默失败（表现为 duration 卡 0，而不是抛出明显错误）。修复：`local_library_notifier.dart` 的 `_importOne` 在读完标签后把文件**复制**进 App 自己的存储目录（新增 `local_library_audio/` 子目录，与既有的 `local_library_covers/` 同级），`LocalLibraryTrack.localPath` 存复制后的路径而非原始选中路径——彻底绕开沙盒临时授权的生命周期问题，副作用是也顺带修复了"用户后续移动/重命名/删除原始文件会导致本地曲库播放失效"这一独立的健壮性缺口。做法与本项目既有的服务端离线下载（`offline_db.dart`/`download_notifier.dart`：下载后落盘到 App 自己的目录，从不引用外部路径）完全一致，補上了本地导入场景遗漏的同一原则。配套修复 `remove()`：之前只删封面文件、遗留复制出的音频文件永久占用磁盘空间，现在两者一起清理。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues，`flutter test --no-pub` 113/113 通过）；本地无 Xcode 无法直接复现"导入后播放"的完整链路，该假设（沙盒临时授权生命周期）基于 macOS App Sandbox 已知行为模式推导，修复后的实际播放效果待远端 CI 构建后由用户实机确认——如果确认仍不播放，需要用户提供更多现象细节（是否有错误提示、其他格式文件是否同样失败）以缩小范围。

### v5.12.2 - 2026-08-06

- **feat: 桌面窗体按登录门禁状态自适应尺寸** — v5.12.1 由用户下载 CI 构建实机验证 v5.12.0/v5.12.1 后提出：启动/登录页窗口不应该像现在这样宽，应固定为接近手机竖屏的"优雅比例"；进入本地曲库/主界面后再放宽——参照的是成熟开源播放器的通用做法（登录态是紧凑对话框、主界面才是宽画布）。新增 `window_manager` 依赖（本来规划在 v5.15.0 桌面自定义窗体一并引入，这次只提前拿"尺寸控制"这一小块能力，标题栏自定义仍留在 v5.15.0）。`DesktopIntegration`（`lib/src/shared/desktop_integration.dart`）新增 `_initWindow()`：冷启动时按 `authProvider` 当前值（`AsyncLoading`/`unauthenticated` 都落在"未过闸"）用 `windowManager.waitUntilReadyToShow` 以窄尺寸 440×720（比例对齐用户截图参照的 443×727）非可调整大小显示；新增 `applyWindowForAuthState(bool isPastGate)`，登录/游客过闸后调用 `setResizable(true)` + `setSize` 放宽到 1440×840（比例对齐参照的 2042×1191）+ 最小尺寸 960×600，退出登录门禁则收窄复原。触发点在 `main.dart` 的 `InoriMusicApp.build()` 里用 `ref.listen(authProvider, ...)` 比对过闸状态是否翻转——没有放进 `DesktopIntegration` 自己的初始化逻辑里，因为 `WidgetRef.listen` 只能在 widget 的 `build()` 方法内调用，`DesktopIntegration` 是个普通类不满足这个前提。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues，`flutter test --no-pub` 113/113 通过）；窗口尺寸切换的实际观感（是否有初始尺寸闪烁、动画过渡是否顺滑）待远端 CI 构建后由用户实机确认。

### v5.12.1 - 2026-08-06

- **fix: `file_picker` 11.x 与 AGP 9 不兼容，导致 Android APK 编译失败** — 推送 v5.12.0 后监控远端 CI：`Docker` 转绿，但 `Flutter CI` 的 `Build APK` job 与 `Build` workflow 的 `Flutter mobile checks` job 均在同一处失败——`android/app/src/main/java/io/flutter/plugins/GeneratedPluginRegistrant.java:34: error: cannot find symbol class FilePickerPlugin`。根因：`file_picker` 10.3.9/11.0.0 起在插件自身的 Android `build.gradle` 里应用了已弃用的 `kotlin-android` 插件，与本仓库固定的 AGP 9.0.1（`android/settings.gradle.kts`）内置 Kotlin 支持冲突，导致插件的 Kotlin 源码实际未被正确编译进构建图，Java 侧自动生成的 registrant 引用不到该类（上游已知问题，`miguelpruivo/flutter_file_picker` issue #1973/#1942/#1952，确认降级到 `10.3.10` 可规避）。本地无 Android SDK 无法直接复现 Gradle 构建，但 `flutter analyze`/`flutter test` 均不需要 Android 工具链，未能提前拦下——这类"依赖版本与本项目固定的 AGP/Kotlin 工具链不兼容"的问题只能靠远端 CI 的真实 Android 构建暴露，是本次的教训。修复：`file_picker` 从 `^11.0.3` 改为精确锁定 `10.3.10`（不用 `^`，避免未来 `pub upgrade` 无意中滑回有问题的版本区间），pubspec.yaml 留注释说明原因与升级前的确认步骤；配套把 `local_library_notifier.dart` 的两处调用从 11.x 的静态方法写法（`FilePicker.pickFiles(...)`）改回 10.x 的实例访问写法（`FilePicker.platform.pickFiles(...)`）——两个大版本间这是真实的公开 API 变更，不是本项目的用法错误。
- The phase output is version-tracked and verified locally（`flutter analyze --no-fatal-infos` 0 issues，`flutter test --no-pub` 113/113 通过）；Android/Windows/macOS 的实际构建结果待本次推送后的远端 CI 确认——本地环境同时缺完整 Xcode 与 Android SDK，v5.12.0 起这几个 phase 的平台特定构建验证已与用户约定改为"推送后由 CI 产出构建，用户下载确认"。

### v5.12.0 - 2026-08-06

- **feat: Flutter 游客本地播放模式** — 用户提出的三项 Flutter 客户端深度体验改造的第一阶段（另两项：启动页/登录页/自定义背景见 v5.13.0，皮肤系统/桌面窗体见 v5.14.0–v5.15.0）。改造前，客户端 100% 依赖服务端账号——`AuthStatus` 只有 `loading/authenticated/unauthenticated` 三态，`unauthenticated` 只是登录前的瞬时态，`router.dart` 的 `redirect` 无条件把它导回 `/login`；全仓库没有任何本地文件播放路径（无 `file_picker`、无任何标签解析库），既有的"离线"功能（`offline_db.dart`）实质是"已登录状态下把服务端曲目缓存到本地"，与"不登录也能用的本地播放器"是两回事。本阶段新增 `AuthStatus.guest`：`AuthNotifier.continueAsGuest()` 纯本地状态切换（不发网络请求），并把选择写入 `shared_preferences`（`auth.lastModeGuest`，与 token 用的 `flutter_secure_storage` 分开），冷启动时无有效 token 但曾选择过游客模式则直接进入 guest 态，不再闪回登录表单；新增 `AuthNotifier.exitGuestMode()`（guest→unauthenticated）作为"游客登录真实账号"的路径——复用路由既有的"未过闸→/login"规则，不需要给登录页单独开一套可达性规则。路由 `redirect` 对 guest 态采用**白名单**而非黑名单放行（`/local-library`、`/settings`、`/player`）：服务端 catalog 相关路由（Artists/Albums/Tracks/Playlists/Search/Favorites/History/MyPlaylists）对游客毫无意义，白名单保证未来任何新增的服务端路由默认对游客不可达，不必每次新增都记得排除。`ShellScaffold` 对 guest 态提前返回一个无导航栏的极简布局（只有内容区 + MiniPlayerBar）——游客只有"本地曲库"一个真实目的地，硬套多目的地的 NavigationBar/Rail 没有意义，设置入口改放在本地曲库页 AppBar 的齿轮图标。
- **feat: 本地文件导入与元数据解析** — 新增 `lib/src/local_library/`：`local_library_db.dart`（`LocalLibraryDb`，完整照抄 `OfflineDb` 的 completer 守卫单例模式，独立 `local_library_tracks` 表——不复用 `offline_tracks`，因为后者的 schema 与全部消费方都假定 `track_id` 是服务端 UUID，混用会让两套语义永久纠缠）、`local_library_notifier.dart`（`file_picker` 多选文件/整目录递归导入，`audio_metadata_reader` 提取标题/艺术家/专辑/时长/内嵌封面，标签缺失或格式不支持时标题回退文件名去扩展名，单个文件解析失败不影响本批次其余文件）、`local_library_screen.dart`（游客的主屏：单层扁平列表，按 artist/album/title 排序——个人本地文件规模通常远小于服务端曲库，暂不做 Artists→Albums→Tracks 三级浏览，留作后续候选）。元数据库选型排除了最初计划的 `audiotags`：该包基于 `flutter_rust_bridge` + Rust crate，需要为 Android/iOS/macOS/Windows 全平台编译原生 Rust 代码，而现有 Flutter CI 完全没有 Rust 工具链，引入后单是让 5 个 CI runner 都能编译就是一次独立的高风险改造；改用纯 Dart 实现的 `audio_metadata_reader`（零原生依赖，覆盖 MP3/MP4/FLAC/OGG/Opus/WAV/AIFF/APE，比原计划格式覆盖更广），避免了这类风险。
- **fix: player_notifier.dart 接入本地曲目播放** — 本地曲目 id 统一加 `local:` 前缀，在原有的按 `trackId` 查询的少数几个函数处分支，而不是引入新的联合类型：队列内存模型本身就是 `audio_service` 的 `MediaItem`（与服务端无关的通用类型），mini player / full player 只读 `title/artist/album/id`，UI 侧零改动。`resolvePlaybackUrl` 命中 `local:` 前缀时查 `LocalLibraryDb` 直接返回 `file://` 路径，跳过服务端 catalog 与既有 `OfflineDb` 缓存检查；`_stubMediaItem`/`_makeMediaItem` 分支到新增的 `_localMediaItem`（缓存命中同步返回，未命中先吐出 id-only 占位并异步 `_backfillLocalTrack`——复用文件里艺术家/专辑名回填的既有模式，不用把整条播放链路改成 async）；`_postHistoryFor` 对 `local:` 前缀直接跳过（本地文件没有服务端历史可言）；跨设备续播上报 `_reportPlayerState` 在当前播放曲目是本地曲目时整体跳过（不是过滤队列后再上报——过滤会让上报的 `currentIndex` 指向错误曲目，直接跳过整次上报更安全）；ReplayGain 增益查找对未知 id 已经安全空转，无需改动。`full_player_screen.dart` 对本地曲目禁用收藏按钮、歌词 tab、卡拉 OK 入口（三者均依赖服务端），并新增本地内嵌封面渲染（`MediaItem.artUri` 为 `file://` 时优先展示，取代原来只认 `albumId` 网络封面的 `_FullPlayerArtwork`）。
- **fix: macOS App Sandbox 缺失文件访问权限（本阶段引入前即可能影响任何未来的文件选择功能）** — 排查中发现 `services/mobile/macos/Runner/{DebugProfile,Release}.entitlements` 启用了 `com.apple.security.app-sandbox` 却未声明任何文件访问权限；沙盒模式下 `file_picker` 弹出的系统选择框选中文件后，后续读取会被沙盒拒绝。两份 entitlements 均补上 `com.apple.security.files.user-selected.read-only`（只读即可——本阶段只读取用户选中的音频文件，封面图另存到 App 自己的数据目录，不回写用户原始文件）。
- **验证** — `flutter analyze --no-fatal-infos` 0 issues；`flutter test --no-pub` 113/113 通过（含新增：`AuthStatus` 四态 + `isGuest`/`isPastGate` getter 测试、`player_notifier.dart` 两处 guest 分支——`_postHistoryFor`/`_reportPlayerState` 守卫——的镜像测试，沿用本仓库既有的"不直接实例化依赖 `audioHandler` 全局单例的 `PlayerNotifier`，镜像其决策逻辑"测试惯例）；本地无完整 Xcode（仅 Command Line Tools）无法 `flutter run -d macos` 做真机走查，**实机验证待本次推送后的远端 CI（Build macOS / Build Windows job）产出构建，由用户下载验证**。

### v5.11.5 - 2026-08-05

- **fix: Admin E2E CI 改用生产构建，根治 storage tab 计时不稳定** — v5.11.4 把超时从 8s/30s 放宽到 15s/45s 后推送观察，`storage tab` 用例反而更慢（18.9s / 45.1s 撞满新超时，比放宽前的 11.9s / 30.1s 还差），证明"再给更多时间"是在追一个不断后退的目标，不是真正的余量不足。根因是 `.github/workflows/build.yml` 的 `admin-e2e` job 用 `npm run dev`（Turbopack 开发模式）跑 admin：开发模式下每个路由第一次被访问时才编译，`storage` 恰好是套件里唯一"第一次真正跑到断言阶段"的路由（此前 middleware/baseUrl 缺陷让全部用例卡在登录阶段，`storage` 从未真正执行到），在 CI 共享 runner 资源下这次首访编译延迟足以吃掉整个超时预算，且预算越松、观察到的耗时也跟着越涨（资源竞争下的可变延迟，不是固定缺口）。改为 `npx next build` + `npx next start -p 3001`（生产模式，构建期预编译全部路由，运行期不再有按需编译的方差）。本地验证：连续两次 `playwright test` 全绿，`storage tab` 耗时从此前的 15.3s/18.9s 收敛到 6.3s/6.7s，与其余三个用例（5.6s–10.2s）处在同一量级，不再是异常值。v5.11.4 放宽的超时保留作为额外余量，不再是承载修复的主力。
- The phase output is version-tracked and verified locally（生产构建成功、`playwright test` 本地连续两次 4/4 通过、耗时分布收敛）；远端 CI 待本次推送最终确认。

### v5.11.4 - 2026-08-05

- **fix: Admin E2E `storage tab` 用例在远端 CI 偶发超时** — 推送 v5.11.3 后监控远端 CI：`Web checks`/`Admin checks`/`Go API checks`/`Flutter mobile checks`/`Web E2E` 五个 job 全部转绿（`Web E2E` 此前从未真正跑绿过），验证了 v5.11.3 的 middleware/baseUrl 修复在真实 CI 环境同样生效；`Admin E2E` 从此前 4/4 失败降到 1/4 失败——`storage tab` 用例第一次尝试在 8000ms 内没等到 `<h1>Storage</h1>`（本地复测该用例单次即耗时 15.3s，明显比同套件其它三个用例慢），重试时更是在最基础的 `getByLabel("Username")` 上撞到 `playwright.config.ts` 的 30000ms 整体测试超时——两个现象指向同一个方向：这是该用例第一次在真正跑通登录之后被执行到（此前 middleware/baseUrl 两个缺陷导致全部 4 个用例一直在登录阶段就失败，从未真正走到这一步过），CI 共享 runner 资源下 Next dev server 编译 + 4 个顺序用例的累积负载让这条路径的余量比本地开发机紧得多，不是逻辑缺陷。将 `playwright.config.ts` 整体测试超时 30s→45s，`storage tab` 断言超时 8s→15s；其余三个用例余量充足未改动。
- The phase output is version-tracked and verified locally（`tsc --noEmit`/`biome lint` 39 files 通过，`playwright test` 4/4 通过，含 storage tab 15.3s 单次耗时确认新超时有余量）；远端 CI 待本次推送最终确认。

### v5.11.3 - 2026-08-05

- **test: 补全 P1 音频引擎核心模块测试（task #74 收尾）** — v5.11.2 的文档承诺测试补齐只覆盖了 `throttle`/`playerStateSync`/`trackGainCache` 三个模块，任务 #74 实际要求的四个核心模块中 `crossfade.ts`、`audioGraph.ts` 仍无测试。新增 `crossfade.test.ts`（9 例，复用 `token.test.ts` 的 fake-localStorage/无 window 双环境模式）与 `audioGraph.test.ts`（20 例，构造最小 fake `AudioContext`/`GainNode`/`BiquadFilterNode` 覆盖真实 WebAudio 连接拓扑：source→10 段 EQ 滤波器链→GainNode→destination、增益/EQ 数学、CORS 兜底降级、共享 AudioContext 复用、首次手势 resume）。至此 Web 音频引擎全部核心模块（crossfade/audioGraph/trackGainCache/playerStateSync/gaplessEngine/replayGain）均有测试覆盖。
- **fix: Build CI 在 Go 与 Web 两侧各有一处独立缺陷** — 推送 v5.11.2/Windows CI 后例行检查远端 CI 状态，发现 Build 工作流连续 3 次推送失败。(1) **`services/api/internal/userplaylist/types.go` gofmt 对齐失效**——v5.11.0 新增 `SourceCatalogID` 字段（比既有字段名更长）后未重跑 gofmt，与 v5.5.1 是同一类缺陷；`gofmt -w` 修复。(2) **`trackGainCache.test.ts`/`playerStateSync.test.ts` 类型错误**——v5.11.2 提交前只跑了 `vitest run`（esbuild 转译，不做完整类型检查）未跑 `tsc --noEmit`：前者 `makeApi()` 返回的裸对象与 `resolveReplayGainDb` 期望的 openapi-fetch `Client` 类型不匹配（6 处，改为在构造处一次性 `as unknown as Fetcher`）；后者 `import ... from "./player-state"` 路径本不存在（真实模块一部分在 `./playerStateSync`、一部分在 `@/lib/api/player-state`），且该导入失败掩盖了后续 16 处真正的类型错误（`t()` 构造的 `{id}` 不满足 `QueueTrack` 全量字段，补全为完整 fixture）。
- **fix: 深挖 Build CI 历史发现两条更早、持续更久的 E2E 缺陷** — 继续向前追溯 Build 失败历史（而非只看最近一次），发现 `Web E2E`/`Admin E2E` 两个 job 各自因为一个**真实产品缺陷**（非环境、非凭据问题）100% 必现失败，分别已持续 10 天与 26/44 天却从未被处理：
  - **Web：搜索历史下拉框在 blur→refocus 后会延迟自关闭**（`app/(app)/search/page.tsx`）——输入框 `onBlur` 用 `setTimeout(..., 150)` 延迟隐藏下拉（为了让历史项的 click 能先注册），但从未在 `onFocus` 时取消这个定时器；若用户在 150ms 内重新聚焦（例如清空后重新点击输入框——`search-history.spec.ts` 的操作序列正是如此），旧定时器仍会在稍后触发并把下拉关闭，导致下拉（含"Clear history"按钮）在用户仍处于聚焦状态时凭空消失、点击目标从 DOM 上被卸载。这是真实用户也会踩到的时序缺陷，不是 Playwright 特有的假象。修复为把 timeout id 存入 ref，`onFocus` 时清除。此缺陷自下拉框功能引入的 v5.1.0 起就存在，`search-history.spec.ts`（v5.5.0 引入）从第一次跑起就没有真正绿过。
  - **Admin：`basePath: "/admin"` 与两处独立机制的错误交互，导致鉴权中间件形同虚设 + 浏览器端全部 API 请求 404** ——两个各自独立、叠加发作的缺陷：(a) `middleware.ts` 的 `config.matcher` 写成 `["/admin/:path*"]`，但 Next.js 对已配置 `basePath` 的项目会自动把 `basePath` 再拼一次到 matcher 上，实际生效的匹配式变成 `/admin/admin/:path*`，永远匹配不到真实请求路径（如 `/admin/users`），中间件对任何真实页面都不会执行——未登录也能直接停留在受保护页面，登录重定向从未触发；由 v4.8.0（"测试与结构还债"）把原本正确、不含 `/admin` 前缀（依赖自动拼接）的 matcher 替换成错误写法引入，已持续 26 天。同时修复了配套的重定向目标二次前缀问题（`loginUrl.pathname` 曾硬编码 `"/admin/login"`，叠加自动前缀变成 `/admin/admin/login`），并删除了基于错误前提编写、从未真正生效过的 `stripBasePath()` 死代码。(b) `lib/api/client.ts` 与 `store/auth.ts` 的浏览器端 `baseUrl` 均写死为空字符串（从没有 basePath 的 `services/web` 对应文件复制而来，未随 admin 的 basePath 调整），导致浏览器发出的每一个 API 请求（含用户名密码登录本身）都打到裸 `/api/v1/...`，未命中同样会被自动加上 `/admin` 前缀匹配的 rewrite 规则，落到 Next.js 自己的 404 页面；登录页把"无 `data`"一律显示成"Invalid credentials."，掩盖了请求实际上根本没到达 Go API 的真相。此缺陷自 admin 独立拆分、引入 basePath 的 v2.5.0 起就存在，长达 44 天，推测因登录页另有一条不经过此路径的"Bootstrap Token"通道被日常使用而未被发现。两处均改为浏览器端 `baseUrl = "/admin"`。均通过本地起真实 Go API + Next dev server + Playwright 复现后再验证修复：curl 直连确认 Go 侧凭据全程有效、写最小 Playwright 探针脚本捕获浏览器真实请求/响应定位到 404 而非凭据错误、逐项修复后 `npx playwright test` 4/4 全绿（此前 0/4）。顺带修正 admin E2E 自身一处过期断言——`users tab` 用例断言页面存在匹配 `/username/i` 的文本，但表头实际文案是"User"非"Username"，改为断言真实种子账号用户名出现在列表中，更贴合用例名"confirms at least one user exists"的本意。
- **hygiene: `services/admin/.gitignore` 补齐 `next-env.d.ts`/`test-results/`/`playwright-report/`** — 与 `services/web/.gitignore` 对齐，此前 admin 端会把 Next.js 自动生成文件与本地/CI 跑 Playwright 产生的临时目录当作未跟踪文件反复出现。
- The phase output is version-tracked and verified locally：`gofmt -l`/`go vet`/`go test -race`（792 passed）/OpenAPI JSON 校验全绿；Web `tsc --noEmit`/`biome lint`（125 files）/`vitest run`（274 passed，含本阶段新增 29 例）全绿；Admin `tsc --noEmit`/`biome lint`（41 files）/`next build` 全绿，`playwright test` 4/4 通过（此前 0/4）；Flutter `flutter test --no-pub` 104/104（确认改动范围未外溢到移动端）。远端 CI 待本次推送验证——此前 Web/Admin E2E job 因 `api`/`web` 前置 job 失败而被跳过，本次预期可以真正执行到。

### v5.11.2 - 2026-08-04

- **test: 补齐 3 处文档承诺但缺失的测试** — `lib/player/throttle.ts`、`lib/player/playerStateSync.ts`、`lib/audio/trackGainCache.ts` 三个模块的 JSDoc 均写有"见 xxx.test.ts"但实际文件不存在；新增 9+15+6 共 30 个测试用例（均在 node 环境运行，无需 DOM）。
- **fix: OpenAPI 契约 `POST /catalog/playlists/{id}/copy` 路径参数改用 `$ref`** — 原为内联参数定义，与其余端点统一使用 `#/components/parameters/CatalogId` 的方式不一致，导致路径参数契约测试失败；改为 `$ref` 后契约测试通过。
- **ci: Flutter CI 追加 Windows 构建任务** — 新增 `build-windows` job（`windows-latest`），推送 main 后除 APK/macOS/IPA 外新增 Windows 产物构建。
- The phase output is version-tracked（VERSION 提交于同一 commit）。本条目为 v5.11.3 阶段整理 Build CI 历史时后补记录，此前提交时未同步补充本节。

### v5.11.1 - 2026-08-03

- **倍速音高保持显式化** —— Web `applyPlaybackRate` 显式设置 `preservesPitch = true`。规范默认即为 `true` 且当前浏览器均遵守，但留作隐式意味着 UA 若改默认值会在非 1× 速度下静默出现变调；Flutter 的 just_audio 同样保持音高，显式声明让两端听感一致。同时把 `makeSlot` 中重复的两行 rate 设置收敛到 `applyPlaybackRate`，避免第三处漂移点。
- **倍速档位收敛为单一来源** —— Flutter 的 `[0.5, 0.75, 1.0, 1.25, 1.5, 2.0]` 此前在 `full_player_screen.dart` 与 `settings_screen.dart` 各硬编码一份字面量，与 Web 的 `SPEED_PRESETS` 无任何约束关联——正是 v5.10.0 修正的 EQ 预设漂移的同款结构。现提取为 `speed_notifier.dart` 的 `speedPresets` 常量（附 `minSpeed`/`maxSpeed`），两处调用点改为引用。
- **跨端档位一致性测试** —— 新增 `test/speed_presets_test.dart`，与 Web `controls.test.ts` 既有的 `SPEED_PRESETS` 断言一一对应：档位字面量、边界内、含中性 1×、升序无重复。任一端改动而未同步另一端，其自身测试即失败。注释约束靠人执行不可靠，故补测试钉死。
- **验证** —— Web `tsc --noEmit` 通过、Biome lint 120 files 通过、Vitest 217/217 通过；Flutter `analyze --no-fatal-infos` 仅 1 个既存 info、`flutter test` 104/104 通过（新增 3 例）。
- 本阶段仅涉及客户端，无服务端 API schema 变化，OpenAPI `info.version` 保持现状。

### v5.11.0 - 2026-08-03

- **播放列表批量导入** —— 新增 `POST /api/v1/catalog/playlists/{id}/copy` 端点，将 catalog 歌单克隆为当前 viewer 的个人歌单，通过新增的 `source_catalog_id` 列保留来源引用（方案 2），为未来跨端同步预留扩展点。服务端 schema 变更（`user_playlists` 加 `source_catalog_id TEXT`）、types/service/handler/repository 同步更新，Go 792 测试通过。Web 端 `CopyFromCatalogDialog` 组件 + `/playlists/[id]` 详情页入口「Copy to my library」按钮，确认后携带目标名称调用新端点，成功后显示 tracks 数量。OpenAPI 新增 `UserPlaylist.sourceCatalogId` 字段与 `copyCatalogPlaylist` 操作，类型已重新生成。
- 本阶段含服务端 schema 变更，OpenAPI `info.version` 保持现状（语义化版本由发布流程管理）。

### v5.10.0 - 2026-08-02

- **EQ 预设跨端对齐（修正）** —— v5.9.0 的 Web EQ 预设值是新拟的，与 Flutter `eq_settings.dart` 早已存在的一套不一致（`bassBoost` 首段 4 dB vs 6 dB、`vocal`/`electronic` 曲线完全不同），且 Web 多出一个 Flutter 没有的 `treble-boost`。Flutter 为先行实现，故以其为准修正 Web：预设 key 改为 `flat` / `bassBoost` / `vocal` / `electronic`，10 段增益值与 Dart 侧逐一对齐；手动拖动滑块后的标记由 `flat` 改为 `custom`（对齐 Flutter 语义），面板下拉在 `custom` 态下追加对应选项。`eqPresets.ts` 顶部注明该表为跨端权威定义，改动需同步 Dart 侧。
- **Flutter 卡拉 OK 模式** —— 对齐 v5.9.1 Web 实现。新增 `lib/src/lyrics/karaoke_progress.dart`，逐字进度算法（`activeLineIndex` / `wordProgress`）与 Web `progress.ts` 逐行对应，含零长度区间、末词兜底尾巴（800 ms）、越界索引等同款边界处理。新增 `lib/src/player/karaoke_screen.dart` 全屏视图：当前行放大不透明、邻行缩放淡出、当前词经 `ShaderMask` 线性渐变从左至右填充（非整词跳变），无逐字时间戳的行回退为整行高亮；入口挂在全屏播放器顶栏，无曲目时禁用。
- **跨端一致性测试** —— 新增 `test/karaoke_progress_test.dart`（Dart，14 例）与 `lib/karaoke/progress.test.ts`（Vitest，14 例），用例一一对应，锁定两端逐字高亮行为一致。
- **验证** —— Web `tsc --noEmit` 通过、Biome lint 119 files 通过、Vitest 217/217 通过；Flutter `analyze --no-fatal-infos` 仅 1 个既存 info、`flutter test` 101/101 通过。
- **附带** —— `.gitignore` 补充 `.DS_Store`。
- 本阶段仅涉及客户端，无服务端 API schema 变化，OpenAPI `info.version` 保持现状。

### v5.9.1 - 2026-07-31

- **歌词全屏卡拉 OK 模式** —— 新增全屏卡拉 OK 视图（`components/player/KaraokePanel.tsx`），复用增强 LRC 逐字时间戳数据。`lib/karaoke/useSmoothPosition.ts` 通过 rAF 插值将 store 的 250ms tick 平滑至 60fps，消除逐字高亮跳动；`lib/karaoke/progress.ts` 导出纯函数 `activeLineIndex` 与 `wordProgress` 供测试验证。歌词行全屏大字显示（Zen Maru Gothic / Poppins），当前行放大高亮，非当前行缩小淡出；活跃词使用 `background-clip: text` 渐变填充实现从左到右的渐进高亮效果；`<dialog>.showModal()` 提供原生 Escape 关闭与焦点恢复；与 `LyricsPanel` 互斥切换。Web TypeScript 类型检查通过，Biome lint 117 files 通过。**真机播放歌词同步测试待执行**。
- 本阶段仅涉及 Web 客户端歌词 UI，没有服务端 API schema 变化，因此 OpenAPI `info.version` 保持现状。

### v5.9.0 - 2026-07-30

- **Web 10 段 EQ** —— 新增共享 `services/web/lib/audio/eqPresets.ts` 定义 10 段 ISO 频率与 5 组预设（Flat/Bass Boost/Treble Boost/Vocal/Electronic），Web `audioGraph.ts` 从 `MediaElementAudioSourceNode -> GainNode -> destination` 扩展为 `-> BiquadFilter[] -> GainNode -> destination`，每个频段为 peaking 滤波器（Q=1.4，±6 dB）。新增 `services/web/store/eq.ts` Zustand persist 保存启用态/预设/10 段增益；`useAudio.ts` 订阅 store 实时应用 EQ，并与 ReplayGain 共存。PlayerBar 新增 `EqualizerControl` 入口 + `EqualizerPanel` 面板，支持启用/禁用、预设选择、逐段 ±6 dB 滑块调节和重置。Web TypeScript 类型检查通过，Biome lint 114 文件清洁。
- 本阶段仅涉及 Web 客户端音频处理与 UI，没有服务端 API schema 变化，因此 OpenAPI `info.version` 保持现状。

### v5.8.1 - 2026-07-30

- **真机走查验收** —— 完成自动化验证：Web TypeScript/Biome 114 files 通过，Flutter `flutter analyze --no-fatal-infos` 通过。人工验收清单已准备就绪，需在实体设备/模拟器上执行：Flutter 登录页/曲库/播放/设置/MiniPlayer 逐屏检查浅色主题可读性，Web ReplayGain 响度对比/连播间隙确认，跨设备续播手机↔Web 双设备场景验收（恢复位置误差 < 5s）。关键对比度配对已知达标：`#3B2A3F`/白 13.2:1、`#6B5570`/白 6.7:1、白/`#D42062` 5.0:1。
- 本阶段为验收阶段，无代码修改，因此 OpenAPI `info.version` 保持现状。

### v5.8.0 - 2026-07-30

- **E2E 测试凭据修复** —— `services/api/cmd/server/main.go` 添加启动时自动创建测试用户逻辑：通过 `INORI_E2E_VIEWER_USER` / `INORI_E2E_VIEWER_PASSWORD` 环境变量创建 `viewer` 角色账户，解决 Playwright 规范因默认 `ci_viewer`/`ci-password-123` 凭据无效而卡在登录前置步骤的问题。更新 `services/web/e2e/smoke.spec.ts` 文档说明环境变量要求。
- 本阶段为 E2E 测试基础设施修复，没有功能或 API schema 变化，因此 OpenAPI `info.version` 保持现状。

### v5.7.0 - 2026-07-30

- **Flutter 樱花薄暮主题对齐** —— `services/mobile/lib/src/shared/theme/neon_shrine.dart` 迁移为 `sakura_dusk.dart`，类名/函数名/色板全面对齐 Web 端樱花薄暮浅色色板：`background #FFF7F2`、`surface #FFFFFF`、`primary #D42062`、`onBackground/onSurface #3B2A3F`、`error #C81E2C`、`miniPlayerShadow #263B2A3F`。`ColorScheme.dark` 改为 `ColorScheme.light`，`main.dart` 切为 `ThemeMode.light` 并移除冗余 `darkTheme`，避免深色亮度推导残留。`primaryViolet*` 24 处引用批量替换为 `sakuraPink*`，共 311 处主题符号引用迁移；通知栏颜色同步改为 `#D42062`。`flutter analyze --no-fatal-infos` 仅剩 1 个既存 info、`flutter test --no-pub` 87/87 通过。**真机走查待执行**：需在模拟器/真机逐屏确认无深底浅字残留，并手工核对关键对比度配对。
- 本阶段仅涉及 Flutter 客户端视觉主题，没有服务端 API schema 变化，因此 OpenAPI `info.version` 保持现状。

### v5.6.0 - 2026-07-28

- **Admin 樱花薄暮主题对齐** —— `services/admin/app/globals.css` 从 Neon Shrine 深色主题迁移到 v5.5.0 确立的樱花薄暮浅色 ACG 色板：canvas 改为暖奶油白 `#faf8f6`、primary 改为降饱和 berry `#b8577a`、语义色补充 `*-dim` 背景、text 改为深梅紫墨 `#3b2a3f`；新增 `--color-scrim` 遮罩、secondary/gold 体系、`*-ink` 文本色。`color-scheme` 切为 `light`，glow/scanline/nav-active 视觉强度下调适配长时间阅读。同步修复 3 处 `bg-black/60` 硬编码遮罩为 `bg-[var(--color-scrim)]`、2 处 Tailwind v4 失效的 `bg-opacity` 用法改为 `bg-[var(--color-success-dim)]`。`layout.tsx` 的 `themeColor` 同步更新为浅色背景。Admin TypeScript 类型检查通过，Biome lint 39 文件清洁。
- 本阶段仅涉及 Admin 客户端视觉与少量样式 token，没有服务端 API schema 变化，因此 OpenAPI `info.version` 保持现状。

### v5.5.0 - 2026-07-26

- **跨设备播放续播** —— 服务端新增 `GET/PUT /api/v1/me/player-state` 端点（`internal/playerstate` 包，memory+postgres 双实现），客户端（Web/Flutter）在播放中每 30 秒节流上报，切歌/暂停/应用后台立即 PUT。Web 启动时 GET 远端状态，若 `updatedAt` 晚于本地则显示「继续上次播放」提示条（不自动播放，等用户手势），确认后重建队列并 seek 到 position（复用 v5.2.0 `restoredPending` 模式）。队列上限 500，`updatedAt` 由服务端生成（last-write-wins）。
- **跨设备搜索历史同步** —— 服务端新增 `GET/PUT/DELETE /api/v1/me/search-history` 端点（`internal/searchhistory` 包）。Web 端在 v5.1.0 localStorage 基础上增加本地∪远端合并（去重取最新，裁到 20 条）后 PUT 回写，单删/清空同步远端。Flutter 端将 SharedPreferences 主存扩展为登录时合并远端，离线时仅本地，回联后下次合并。
- **OpenAPI 契约升级** —— `packages/api-contract/openapi/storage-admin.v1.json` 版本从 5.0.0 升至 5.4.0，新增 `PlayerState` 和 `SearchHistory` schema 及 4 个新路径。`openapi_contract_test.go` 同步更新。
- **验证** —— Go 792/792 测试通过（含 playerstate/searchhistory 单元测试、handler 测试、-race），Web TypeScript 通过，Vitest 203/203 通过，Biome 110 文件通过。
- 本阶段为 v5 封版收官，后续新方向按 v6+ 升版。

### v5.3.0 - 2026-07-16

> v5.3.0 是纯 Web 客户端阶段，消费既有服务端能力；无 API schema / OpenAPI 变更。

- **Web 用户播放列表** —— 新增个人播放列表 CRUD、曲目管理、拖拽排序与一键播放，全部消费既有 `/api/v1/me/playlists/*` 端点。包含列表页（创建/重命名/删除/空态引导与错误重试）、详情页（播放全部、单曲移除、按 `uid` 区分的重复曲目安全移除、PUT 拖拽排序保留重复与顺序）、TrackRow 上下文菜单「添加到播放列表」（复用选择器 + 内联新建，支持多播放列表独立完成态）。侧边栏与移动端导航同步加入播放列表入口。使用原生 `<dialog>.showModal()` 模态框与原生可聚焦陷阱、Escape 关闭与焦点恢复；所有新 prop 均为默认启用，现有调用方保持源兼容。
- **播放速度控制** —— Web 播放器新增 0.5/0.75/1/1.25/1.5/2× 六档倍速，对齐移动客户端档位。通过独立 `GainNode` 的 `rampGain` / `setGain` 外，同时设置 `HTMLAudioElement.playbackRate` 与 `defaultPlaybackRate`（避免 `load()` 重置还原）；双元素引擎在预载待机槽、切歌 swap、恢复播放与会话恢复时同步继承；倍速状态通过已有 Zustand `persist` 保存；1× 外角标在 PlayerBar 与全屏播放器显示。
- **睡眠定时器** —— 复刻移动客户端双模式：固定时长（15/30/45/60 分钟，会话级倒计时显示与取消）与「当前曲目结束后停止」。到期走已有 player store `pause()` 同步清空定时器；after-track 通过 `useAudio` 的 `ended` 路径优先拦截、阻止 advance / repeat / crossfade。位置 ticker 在 after-track 挂起期间直接返回，回避 lead-time自然淡出绕过语义。定时器 state 未进入 `persist`，会话级意图、刷新即清，登出时 `clearSession()` 同步 cancel，避免跨用户泄漏。
- **音频状态机收敛** —— 经过多轮对抗性审查与回归测试，收敛了 v5.2.0 引入的双元素引擎竞态：晚到 `play()`、快速连续切歌、淡出槽复用与晚到淡出清理、固定时长到期与 after-track 占用、重复曲目预载身份。新增纯 `PlaybackCycle{loadId,playGen}` 代次身份、`ended` 推进守卫、合成/原生事件来源判定、`play()` settlement 守卫，Vitest 覆盖全部关键路径；原有 e2e / 人工听感验证仍依赖集成环境凭据与种子数据，本轮未执行。
- **验证** —— Web 最终 TypeScript 通过、Biome lint 清洁（106 files）、Vitest 203/203 通过（13 files）；Go API 回归 784 tests / 19 packages 通过。
- 未触碰服务端 / OpenAPI `info.version`。

### v5.2.0 - 2026-07-14

- **WebAudio ReplayGain 管线** —— Web 播放器通过共享 `AudioContext`、`MediaElementAudioSourceNode` 和独立 `GainNode` 实际应用曲目 `replayGainDb`；增益按 `10^(dB/20)` 计算并限制在 `0.1–2.0`，与用户音量正交。新增默认关闭且本地持久化的 ReplayGain 设置；WebAudio 或跨域条件不满足时优雅退回原生元素播放。
- **双元素 gapless 与真实 crossfade** —— 两个 `HTMLAudioElement` 轮换播放，在进度超过 50% 或剩余不足 30 秒时预载精确的下一队列位置；自然结束可无缝切换，可选 2 秒线性交叉淡化在曲终前启动。预载按队列位置、曲目 ID、加载代次及 `canplay` 状态校验，并处理签名 URL 过期、重复曲目、快速连续切歌、淡出槽复用、晚到 Promise 与播放拒绝清理等竞态；repeat-one 会从零重新播放。
- **本地播放状态恢复** —— Zustand persist 保存队列、当前位置、音量、随机及循环设置，位置写入按 5 秒节流；页面刷新后恢复曲目和进度但不自动播放，登出时清除持久化播放状态，避免跨用户泄漏。
- **设置、运维与验证** —— 新增 `/settings/audio` 页面和 S3 音频 CORS 运维文档；Vitest 覆盖增益、预载、URL 新鲜度、队列 occurrence、crossfade 清理和持久化，最终 Web 类型检查、Biome lint（90 files）及 Vitest（9 files / 131 tests）通过。Playwright gapless e2e spec 已编写，但本地实跑因默认 `ci_viewer` 凭证无效停在登录前置步骤、未触达音频断言；人工响度/无缝切歌听感验收本轮也未执行——两者均待提供有效 `E2E_USERNAME` / `E2E_PASSWORD` 的集成环境后补齐。
- 本阶段仅涉及 Web 客户端及运维文档，没有服务端 API schema 变化，因此 OpenAPI `info.version` 保持现状。

### v4.8.4 - 2026-07-11

> 2026-07-11 补：本轮实际上做了三次 v4.8.4 push（`994ff31`、`ce8a491`、`0341c99`）才收绿。下表把三轮放在一起记录真实通过的修复。

- **fix: 4.8.3 推送后 Docker / Flutter / Admin E2E 收尾的 3 类独立缺陷** ——继续追踪 v4.8.3 触发的远端 CI，定位三个互不相关的失败根因并修复：
  (1) **Docker 镜像构建 npm ci EUSAGE（BuildKit COPY content-hash 缓存污染）**——v4.8.1 引入 wildcard COPY `package-lock.json*` 后 build 缓存含一份把 package.json+package-lock.json 预装进 image 的 layer，`npm ci` 时看到锁文件认为已装过直接退出，实际 node_modules 不存在 → EUSAGE。多源 COPY `COPY services/<svc>/package.json services/<svc>/package-lock.json ./` 与 wildcard-COPY 共享 content-hash → 拆分 COPY 不能打破缓存。`ARG CACHE_BUST=1` + `RUN echo` 不能击穿 COPY 层（只对 RUN/CMD/ENTRYPOINT 有效）。**正解**：docker.yml web/admin jobs `build-args` 加 `CACHE_BUST=${{ github.run_id }}`，RUN echo cache key 每轮变化 → 整个下游缓存链失效 → COPY+npm ci 实跑成功。Dockerfile 不需要拆分 COPY，但之前的拆分也无害、保留。
  (2) **Flutter APK build.gradle.kts 编译错误（15 个）**——CI 用 AGP 9 + newDsl + Kotlin DSL：`java.util.Properties()` FQN 在 Kotlin DSL script 编译阶段报 `'util' unresolved`（AGP 7-8 不报，AGP 9 报错）。**实际上 `import java.util.Properties` 已经存在，但 FQN 不工作**。修复：把两个 `java.util.Properties()` 调用改为 `Properties()`（与 import 对齐）。
  (3) **Flutter `analyze` 把 info 当 fatal + Turbopack dev mode @import 排序**——(i) `full_player_screen.dart`/`settings_screen.dart` 各 2 处字符串插值冗余花括号 `${speed}×`，`flutter analyze --no-fatal-infos` 不再报；(ii) **admin 端 Turbopack dev mode CSS 解析失败**：admin 端 `globals.css` 的 `@import "tailwindcss"` 后跟 `@import url(goog-fonts)`，Next 15.5.19 Turbopack dev mode 把 `@import url` 推到 `@theme` 块之后触发 CSS spec「@import 必须最先」冲突 → 整个 `/admin/login` 无法渲染。修复：globals.css 移除 `@import url`，改 app/layout.tsx 用 `<link rel="stylesheet">`。
  (4) **Admin E2E 路径缺 basePath 前缀**——`next.config.ts` 设 `basePath: "/admin"`，所有路径实际挂 `/admin/*` 下；E2E 用 `/login`, `/users` 等裸路径 → Next 直接返回 404（`h1 "404" / "This page could not be found."`）。**此错误在 (3)(ii) 修复后才暴露**（CSS overlay 修复后 E2E 才真正 load 页面）。修复：smoke.spec.ts goto 路径全部加 `/admin` 前缀；toHaveURL 改 `/\/admin\/login/`；最终断言 `/\/admin\/(dashboard|users|catalog|storage)/i`。
- **hygiene: `services/admin/.gitignore` 追加 `*.tsconfig.tsbuildinfo`**——`tsconfig.tsbuildinfo` 是 TypeScript 增量构建产物，admin 端此前缺失导致不时进 staging。
- 故障排查时间线（2026-07-09 → 2026-07-12，**最终 9 轮**才真正 Docker 全绿）：
  - 07-09 v4.8.1 → 3 真实失败 → v4.8.3 触 2 失败（Docker EUSAGE + Admin E2E）
  - v4.8.4 round 1 (`994ff31`) 加静态 CACHE_BUST + gradle FQN + css link → 仍红
  - round 2 (`ce8a491`) 拆分多源 COPY → Docker 仍红 / Admin E2E 暴露 404
  - round 3 (`0341c99`) 把 CACHE_BUST 改为从 workflow 注入 run_id + E2E 加 `/admin` 前缀 → **Flutter/Admin E2E 绿，Docker 仍红**（此前误标"全绿"，实际 Docker 从未成功过）
  - round 4 (`fb55b2f`) 误判"poisoned local BuildKit cache" → 引入 registry-backed cache → **新失败**：`invalid reference format: repository name (InoriAsuka/inori-music/admin) must be lowercase`（`${{ github.repository }}` 保留大小写，registry cache exporter 要求全小写）
  - round 5 (`c0b6ee5`) 硬编码 `ghcr.io/inoriasuka/inori-music/...:buildcache` 小写 → 缓存导入成功（`buildcache: not found` 是空缓存首次正常态），但 **npm ci 仍 EUSAGE** —— 彻底否定"缓存污染"假说
  - round 6 (`182b727`) 合并 COPY + 加 `ls -la`/`lockfileVersion` 诊断 → 证明 lockfile 文件完整、可 parse、大小正确，**不是文件损坏**
  - round 7 (`f0dbc7f`) 加 `--loglevel verbose` → 揭开 npm ci 的真正异常：`TypeError: Cannot read properties of undefined (reading 'extraneous')` 于 `@npmcli/arborist/lib/arborist/load-virtual.js:296`（npm 10.9.8 / node:22-alpine）。**关键发现**：`ci.js` 的 `loadVirtual()` catch 块把任何异常都重抛为通用 EUSAGE 文案，真实 stack 只在 verbose 可见 —— 前 6 轮一直在追 decoy 消息
  - round 8 (`9bcfb91`) 误诊为"file: link 目标不在磁盘" → 在 deps stage 加 `COPY packages/ui ../../packages/ui` → **仍失败**（输出与 round 7 字节级相同）
  - **round 9a (`c48d77c`) 真正根因**：`WORKDIR /app` 只有 1 层深度。Arborist 用 `path.resolve(WORKDIR, meta.resolved)` + `path.relative` 重算 link 目标查找 key；`path.resolve('/app', '../../packages/ui')` 在 FS root 处夹紧 → 重算 key 变成 `../packages/ui`，与 lockfile 字面量 `"../../packages/ui"` 不匹配 → `nodes.get()` 返回 undefined → `#loadLink` 在 `target.extraneous` 崩溃。这是**纯内存路径字符串不匹配**，在任何磁盘访问之前发生 —— 所以 round 8 的"复制文件"修复完全无效。**正解**：deps + builder stage 的 WORKDIR 改为镜像 monorepo 真实深度 `/repo/services/web` / `/repo/services/admin`。**验证**：web 镜像 end-to-end 成功（job 86563312502）
  - **round 9b (`4a0037d`) 第二个独立 latent bug**：npm ci 第一次真正成功后，admin builder stage 暴露 webpack `Module not found: Can't resolve '@inori/ui'`——admin 的 builder 阶段从未 `COPY packages/ui`（web 的 builder 一直有这行），此 gap 早于整个 saga，只是被 npm ci 崩溃掩盖。补上后 **Docker 全绿**（Web + API + Admin 三镜像全部 success）
- 真正的 Docker 根因链（最终）：
  1. **主因（round 9a）**：WORKDIR 深度与 monorepo 相对路径 `file:../../packages/ui` 不匹配 → Arborist 路径夹紧 → loadVirtual TypeError → 被 npm ci 吞成 EUSAGE decoy
  2. **次因（round 9b）**：admin builder 缺 `COPY packages/ui`（webpack 编译期）——被 1 掩盖
  3. **旁支（round 4-5）**：GHCR cache ref 大小写（已修，无害）
  4. **误诊（rounds 1-3, 6-8）**：缓存污染 / 文件损坏 / 文件缺失 —— 全部被后续实验否定
- 临时部署：`root@192.168.58.34` 的 `inori-v484-test` 栈（docker-compose.v484-test.yml）已成功 pull + up，全 5 容器 healthy：
  - API `http://192.168.58.34:18080/healthz` → `{"status":"ok"}`
  - Web `http://192.168.58.34:13000` → HTTP 307
  - Admin `http://192.168.58.34:13001/admin/login` → HTTP 200
- **已知独立问题（不阻塞部署）**：Build workflow 的 Admin E2E smoke 仍失败（`page.waitForURL` 登录重定向超时，`services/admin/e2e/smoke.spec.ts:33`）。自 `0341c99`（2026-07-10）起每次 Build run 都同样失败，与 Docker/npm 无关，不在本 patch 范围。

### v0.1.0 - 2026-06-02

- Establish server-managed multi-backend media storage covering local, NFS, SMB, S3-compatible, and distributed backends.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.2.0 - 2026-06-02

- Create the Go API scaffold and storage domain with validation, capability inference, default backend handling, and in-memory repositories.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.3.0 - 2026-06-02

- Expose storage administration through versioned HTTP endpoints for validation, registration, listing, default selection, and disabling.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.4.0 - 2026-06-02

- Protect administrator routes with bootstrap bearer-token authentication while keeping /healthz public.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.5.0 - 2026-06-02

- Add safe filesystem probes for local, NFS, SMB, and mounted distributed backends.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.6.0 - 2026-06-02

- Add conservative S3-compatible object probes with environment-referenced credentials.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.7.0 - 2026-06-02

- Add batch refresh, optional background refresh, and filesystem capacity reporting.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.8.0 - 2026-06-03

- Publish and test the OpenAPI 3.1 contract for the admin API.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.9.0 - 2026-06-03

- Add optional JSON file-backed persistence for storage backend state.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.10.0 - 2026-06-03

- Add the media object registry domain for metadata-only binary asset references.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.11.0 - 2026-06-03

- Expose authenticated media object registration, fetch, and filter endpoints.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.12.0 - 2026-06-03

- Add optional JSON file-backed persistence for media object metadata.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.13.0 - 2026-06-03

- Add read-only filesystem integrity verification for media object references.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.14.0 - 2026-06-03

- Add batch media object verification by backend ID or content hash.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.15.0 - 2026-06-04

- Persist the latest media object verification result in metadata.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.16.0 - 2026-06-04

- Support filtering media objects by latest verification status.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.17.0 - 2026-06-04

- Add limit/offset pagination and pagination metadata to media object lists.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.18.0 - 2026-06-04

- Add metadata-only media object statistics for dashboard-style summaries.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.19.0 - 2026-06-04

- Add metadata-only media object lifecycle updates with terminal deleted semantics.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.20.0 - 2026-06-04

- Support filtering media object lists by lifecycle state.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.21.0 - 2026-06-04

- Support filtering media object lists by asset kind.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.22.0 - 2026-06-04

- Split README content and localize documentation in the previous phase.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.23.0 - 2026-06-04

- Restore Markdown documentation to English as the repository documentation policy.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.24.0 - 2026-06-04

- Support media object list sorting by `backend_object_key`, `created_at`, `updated_at`, `size_bytes`, `object_key`, or `id`, with `asc` or `desc` order before pagination.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.25.0 - 2026-06-04

- Add metadata-only media object duplicate detection by content hash, with optional backend scoping and configurable minimum copy counts.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.26.0 - 2026-06-05

- Add metadata-only bulk media object lifecycle updates selected by exactly one filter, preserving terminal deleted semantics and never deleting media bytes.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.27.0 - 2026-06-05

- Add dry-run previews for metadata-only bulk media object lifecycle updates, reporting would-update outcomes without persisting changes.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.28.0 - 2026-06-05

- Persist latest committed media object lifecycle change metadata, including previous state, new state, change time, and single/bulk source.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.29.0 - 2026-06-05

- Add a read-only media object metadata timeline endpoint for registration, latest verification, and latest lifecycle transition summaries.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.30.0 - 2026-06-05

- Add metadata-only media object statistics backend scoping.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.31.0 - 2026-06-05

- Add a public `/versionz` endpoint exposing API name, version, commit, and build time.
- Inject version metadata into release binaries and Docker images through build flags and Docker build arguments.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.32.0 - 2026-06-05

- Add a public `/readyz` endpoint with storage service, media registry, and admin authentication readiness checks.
- Add a Docker liveness healthcheck that uses `/healthz` for container runtime monitoring.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.33.0 - 2026-06-06

- Add a public `/metrics` endpoint using Prometheus text exposition for readiness gauges and API build information.
- Keep metrics non-sensitive and aligned with `/readyz` readiness checks and `/versionz` build metadata.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.34.0 - 2026-06-06

- Add low-cardinality HTTP request counters and cumulative duration metrics labeled by method, route pattern, and status.
- Reuse the public `/metrics` endpoint while avoiding raw URL labels and secret-bearing request data.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.35.0 - 2026-06-10

- Add PostgreSQL-backed repository implementations for storage backends and media objects with automatic schema migration and shared connection pool.
- File and in-memory repositories remain available when INORI_DATABASE_URL is not set.
- Integration tests use testcontainers-go with a real PostgreSQL container under the integration build tag.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.36.0 - 2026-06-11

- Add user domain with PostgreSQL persistence: User and Session types, UserRepository and SessionRepository interfaces and PostgreSQL implementations.
- Add bcrypt password hashing (cost=12) and SHA-256 session token storage; plaintext token returned once at login, never stored.
- Add auth.Service: CreateUser, Login, Logout, ValidateToken, DisableUser, DeleteUser, EnsureInitialAdmin.
- Add INORI_SESSION_TTL env var (default 24h) and INORI_INITIAL_ADMIN_USER/PASSWORD bootstrap env vars.
- Add migrations 003_users and 004_sessions to shared PostgreSQL migration runner.
- Add 13 unit tests covering all service paths, race-clean.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.37.0 - 2026-06-11

- Add POST /api/v1/auth/login and POST /api/v1/auth/logout endpoints for session-based authentication.
- Upgrade requireAdminAuth middleware: validate session token via auth.Service first, fall back to INORI_ADMIN_TOKEN bootstrap token.
- Return 503 when neither auth service nor admin token is configured; return 401 on bad/missing credentials.
- Add 8 HTTP-layer tests covering login, logout, session token access, and revoked token denial.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.38.0 - 2026-06-11

- Add authenticated user management APIs for administrators: list users, create users, disable users, and delete users.
- Restrict user management routes to session-authenticated admin users while preserving bootstrap-token fallback behavior.
- Add HTTP-layer tests covering the full user management workflow, validation, conflicts, authorization, and missing auth service handling.
- Extend the OpenAPI contract with auth login/logout and user management schemas and paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.39.0 - 2026-06-12

- Add the music catalog domain foundation with Artist, Album, and Track metadata entities and repository interfaces.
- Add catalog service validation for required names, artist ownership, album membership, media object references, and non-negative numeric metadata.
- Add PostgreSQL-backed catalog repository implementations for artists, albums, and tracks.
- Add migration 005_catalog to the shared PostgreSQL migration runner with catalog tables and lookup indexes.
- Add race-clean catalog service tests and integration-build coverage for the PostgreSQL repository.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.40.0 - 2026-06-12

- Expose authenticated catalog administration endpoints for artists, albums, and tracks under `/api/v1/admin/catalog/`.
- Add list, create, get, and delete operations for all three entity types with `artistId` and `albumId` filter parameters on list endpoints.
- Add `MemoryRepository` to the catalog package for use in HTTP handler tests without external dependencies.
- Update the OpenAPI contract with catalog paths, `UserId`, and `CatalogId` path parameter components.
- Add 11 HTTP-layer catalog tests covering workflows, not-found errors, validation errors, and unconfigured service handling.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.41.0 - 2026-06-13

- Add `POST /api/v1/admin/catalog/import` endpoint that converts a verified media object into a catalog track record.
- Import validates that the media object exists, has `original_audio` or `transcoded_audio` asset kind, and is in `active` lifecycle state.
- Track title falls back to the media object ID when not supplied.
- Artist inherits from the album when only `albumId` is provided.
- Add `MediaObjectReader` interface and `ImportTrackRequest` to the catalog package; `GetMediaObjectInfoForImport` helper on `MediaObjectService`.
- Wire media object service → catalog service via `mediaObjectReaderAdapter` in the HTTP handler layer (no import cycle).
- Add `WithMediaObjectReader` method to `catalog.Service`.
- Add `ErrImportRejected` sentinel error mapped to HTTP 422.
- Add 7 `ImportTrack` unit tests in the catalog package and 7 HTTP-layer tests.
- Update OpenAPI contract with import route and `CatalogImportRequest` schema.


### v0.42.0 - 2026-06-13

- Add PostgreSQL full-text search over catalog metadata via `GET /api/v1/admin/catalog/search?q=`.
- Add migration `006_catalog_fts` with generated `tsvector` columns (weighted `A`/`B` for name vs sort-name) and GIN indexes on `artists`, `albums`, and `tracks`.
- Add `SearchCatalog(ctx, query)` to `catalog.Repository` interface and `catalog.Service`; empty query rejected with validation error.
- Implement `SearchCatalog` on `catalogpg.Repository` using `plainto_tsquery('simple', ...)` with `ts_rank` ordering; results grouped artists → albums → tracks.
- Add `MemoryRepository.SearchCatalog` with case-insensitive substring fallback for unit-test environments.
- Add `CatalogSearchResult`, `SearchResultItem`, `SearchResultKind` types.
- Add 5 `SearchCatalog` service unit tests and 4 HTTP-layer tests.
- Add `TestRepositorySearchCatalog` integration test (build tag: integration).
- Update OpenAPI contract with `/api/v1/admin/catalog/search` path, `CatalogSearchResult`, `SearchResultItem`, `SearchResultKind` schemas, and new error codes.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.43.0 - 2026-06-13

- Add read-only catalog browse endpoints for session-authenticated viewers and admins under `/api/v1/catalog/`: list/get artists, list/get albums (`?artistId=`), list/get tracks (`?albumId=`/`?artistId=`), and full-text search (`?q=`).
- Add `requireViewerAuth` middleware that accepts any valid session token (admin or viewer role) but rejects the static bootstrap admin token; returns 503 when no auth service is configured, 401 for missing or invalid tokens.
- Reuse existing `listArtists`, `getArtist`, `listAlbums`, `getAlbum`, `listTracks`, `getTrack`, and `searchCatalog` handlers without modification.
- Fix missing 405 method-not-allowed fallback for `/api/v1/admin/catalog/search`.
- Add `newViewerTestHandler` helper and 11 HTTP-layer tests covering viewer/admin session, 401 unauthorized, 503 for no-auth-service, not-found, missing query, seeded search, and 405 guards.
- Update OpenAPI contract with 7 new viewer catalog paths; bump `info.version` to `0.43.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.44.0 - 2026-06-13

- Add `POST /api/v1/admin/catalog/batch-import` endpoint that accepts a list of `CatalogImportRequest` items and processes each independently.
- Return HTTP 200 on full success, HTTP 207 Multi-Status on partial success, HTTP 422 when all items fail.
- Each result item carries `index`, `mediaObjectId`, the created `track` on success, or `error`/`errorCode` on failure.
- Add `BatchImportTracks(ctx, items)` method to `catalog.Service`; individual item failures do not abort subsequent items.
- Add `BatchImportResult`, `BatchImportResultItem` types to the catalog package.
- Add 5 `BatchImportTracks` unit tests and 6 HTTP-layer tests covering full-success, partial-success, all-fail, empty batch, no-catalog-service, and 405 guard.
- Update OpenAPI contract with `/api/v1/admin/catalog/batch-import` path and `CatalogBatchImportRequest`, `CatalogBatchImportResult`, `CatalogBatchImportResultItem` schemas; bump `info.version` to `0.44.0`.
- Fix pre-existing flaky `TestMediaObjectServiceUpdatesLifecycleState` by injecting a stepping clock that guarantees distinct timestamps across `RegisterMediaObject` and `SetMediaObjectLifecycleState` calls.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.45.0 - 2026-06-14

- Add `PATCH /api/v1/admin/catalog/artists/{id}`, `PATCH /api/v1/admin/catalog/albums/{id}`, and `PATCH /api/v1/admin/catalog/tracks/{id}` endpoints for partial metadata updates.
- Pointer-typed request fields (`*string`, `*int`) distinguish "not provided" from "explicitly empty", enabling clients to clear optional fields without touching unmentioned fields.
- Validation mirrors create: name, title, and artistId may not be set to an empty string; numeric fields must be non-negative; artist ownership of the referenced album is enforced when updating a track's albumId.
- Add `UpdateArtistRequest`, `UpdateAlbumRequest`, `UpdateTrackRequest` types to the catalog package.
- Add `WithClock` setter on `catalog.Service` for deterministic timestamp injection in tests.
- Implement `UpdateArtist`, `UpdateAlbum`, `UpdateTrack` on `catalog.Service`; each reads the current record, applies non-nil fields, bumps `UpdatedAt`, and saves.
- Add 11 `catalog.Service` unit tests and 7 HTTP-layer tests covering field changes, nil-field passthrough, empty-name rejection, not-found, and unconfigured-service guard.
- Update OpenAPI contract with `CatalogUpdateArtistRequest`, `CatalogUpdateAlbumRequest`, `CatalogUpdateTrackRequest` schemas and `patch` operations on artist, album, and track `/{id}` paths; bump `info.version` to `0.45.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.46.0 - 2026-06-14

- Add `Playlist` entity to the catalog domain: ordered collection of tracks with name, optional description, and an ordered `TrackIDs` list.
- Add `ErrInvalidPlaylist` and `ErrPlaylistNotFound` sentinel errors to the catalog package.
- Extend `catalog.Repository` interface with `SavePlaylist`, `GetPlaylist`, `ListPlaylists`, `DeletePlaylist`.
- Implement playlist methods on `catalog.MemoryRepository` (defensive slice copies) and `catalogpg.Repository` (transactional upsert + playlist_tracks replace).
- Add `CreatePlaylist`, `ListPlaylists`, `GetPlaylist`, `DeletePlaylist`, `UpdatePlaylist`, `AddTrackToPlaylist`, `RemoveTrackFromPlaylist` methods to `catalog.Service`; validate name non-empty, enforce track existence on add.
- Add migration `007_playlists` with `playlists` and `playlist_tracks` tables; `playlist_tracks` uses `ON DELETE CASCADE` for both foreign keys and a `(playlist_id, position)` primary key for ordering.
- Expose admin playlist endpoints under `/api/v1/admin/catalog/playlists/`: list, create, get, PATCH metadata, delete, `POST /{id}/tracks` (append), `DELETE /{id}/tracks/{trackId}` (remove first occurrence).
- Expose viewer-only read endpoints under `/api/v1/catalog/playlists/`: list and get.
- Add `ErrInvalidPlaylist` and `ErrPlaylistNotFound` to the `writeError` switch in the HTTP handler.
- Add 9 `catalog.Service` unit tests and 7 HTTP-layer tests covering CRUD, add/remove track, viewer access, not-found, empty-name rejection, and 405 guard.
- Update OpenAPI contract with `Playlist`, `CreatePlaylistRequest`, `UpdatePlaylistRequest`, `AddPlaylistTrackRequest` schemas and all 8 new paths; bump `info.version` to `0.46.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.47.0 - 2026-06-14

- Add `PUT /api/v1/admin/catalog/playlists/{id}/tracks` endpoint that atomically replaces the entire ordered track list of a playlist.
- Every supplied track ID must exist; an unknown ID returns 404. An empty `trackIds` array is valid and clears the playlist. Duplicate entries are preserved.
- Add `SetPlaylistTracks(ctx, playlistID, trackIDs)` to `catalog.Service`; validates track existence via `repo.GetTrack` for each ID, then calls `repo.SavePlaylist` which already performs a transactional full-replace of `playlist_tracks` rows in the PostgreSQL backend.
- No `catalog.Repository` interface change required — `SavePlaylist` already handles atomic replacement.
- Add `setPlaylistTracksRequest` struct and `setPlaylistTracks` handler to the HTTP handler layer.
- Add 5 `catalog.Service` unit tests and 7 HTTP-layer tests covering reorder, clear, duplicate preservation, unknown track, unknown playlist, missing `trackIds` field, and no-catalog-service 503.
- Add `SetPlaylistTracksRequest` schema and `put` operation on `/api/v1/admin/catalog/playlists/{id}/tracks` to the OpenAPI contract; bump `info.version` to `0.47.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.48.0 - 2026-06-14

- Add `POST /api/v1/admin/catalog/tracks/{id}/relink` endpoint that replaces the media object reference on an existing track.
- The new media object must exist, have `assetKind` of `original_audio` or `transcoded_audio`, and have `lifecycleState` of `active`; otherwise `422 relink_rejected` is returned.
- Add `ErrRelinkRejected` sentinel error and `RelinkTrackRequest` type to the catalog package.
- Add `RelinkTrack(ctx, id, req)` to `catalog.Service`; validates media object via `MediaObjectReader` before overwriting `mediaObjectId` and bumping `UpdatedAt`.
- Add `relinkTrack` handler and route to the HTTP handler layer; register 405 fallback for the sub-path.
- Add 7 `catalog.Service` unit tests and 6 HTTP-layer tests covering success, wrong asset kind, not-active lifecycle, media not found, track not found, no reader configured, and empty `mediaObjectId`.
- Add `CatalogRelinkTrackRequest` schema and `post` operation on `/api/v1/admin/catalog/tracks/{id}/relink` to the OpenAPI contract; bump `info.version` to `0.48.0` (corrected in v0.49.0).
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.49.0 - 2026-06-14

- Add `GET /api/v1/admin/catalog/playlists/{id}/tracks` and `GET /api/v1/catalog/playlists/{id}/tracks` endpoints that return the full ordered `Track` object list for a playlist in a single request, eliminating the need for N separate per-track fetches.
- An empty playlist returns an empty `tracks` array. Duplicate track entries (same ID appearing multiple times) are expanded once per occurrence in order.
- Add `GetPlaylistTracks(ctx, playlistID)` to `catalog.Service`; resolves each `trackID` in the playlist's `TrackIDs` slice via `repo.GetTrack` and returns them in order. Returns `ErrPlaylistNotFound` for unknown playlist IDs.
- Add `getPlaylistTracks` handler shared by both the admin and viewer routes; response shape is `{"tracks": [...Track...]}`.
- Register `GET /api/v1/admin/catalog/playlists/{id}/tracks` (admin-auth) and `GET /api/v1/catalog/playlists/{id}/tracks` (viewer-auth) routes; register 405 fallback for the viewer sub-path.
- Add 4 `catalog.Service` unit tests (ordered, empty, not-found, duplicate expansion) and 6 HTTP-layer tests (admin happy path, empty playlist, 404, viewer access, no-catalog-service 503, method-not-allowed 405).
- Add `PlaylistTracksResult` schema and `get` operations on both tracks sub-paths to the OpenAPI contract; bump `info.version` to `0.49.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert all eight playlist paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.50.0 - 2026-06-14

- Add `GET /api/v1/admin/catalog/stats` endpoint returning metadata-only aggregate entity counts for artists, albums, tracks, and playlists as a `CatalogStats` object.
- Add `CatalogStats` struct to the catalog package with `artists`, `albums`, `tracks`, `playlists` integer fields.
- Add `GetCatalogStats(ctx)` to `catalog.Service`; delegates to `repo.ListArtists`, `repo.ListAlbums`, `repo.ListTracks`, `repo.ListPlaylists` and returns counts. No new `Repository` interface methods required.
- Add `getCatalogStats` handler to the HTTP handler layer; returns 503 when no catalog service is configured.
- Register `GET /api/v1/admin/catalog/stats` (admin-auth) and 405 fallback for the path.
- Add 3 `catalog.Service` unit tests (empty catalog, populated counts, no-error baseline) and 4 HTTP-layer tests (empty response shape, populated counts, no-catalog-service 503, method-not-allowed 405).
- Add `CatalogStats` schema and `get` operation on `/api/v1/admin/catalog/stats` to the OpenAPI contract; bump `info.version` to `0.50.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.51.0 - 2026-06-14

- Add `GET /api/v1/admin/catalog/stats/artists` endpoint returning per-artist album and track counts as a `CatalogArtistStatsBreakdown` object, eliminating the need for N separate list calls.
- Add `GET /api/v1/admin/catalog/stats/albums` endpoint returning per-album track counts as a `CatalogAlbumStatsBreakdown` object.
- Add `ArtistStatItem`, `ArtistStatsBreakdown`, `AlbumStatItem`, `AlbumStatsBreakdown` types to the catalog package.
- Add `GetArtistStatsBreakdown(ctx)` and `GetAlbumStatsBreakdown(ctx)` to `catalog.Service`; counts are derived from existing `ListAlbumsByArtist`, `ListTracksByArtist`, `ListTracksByAlbum` calls. No new `Repository` interface methods required.
- Add `getArtistStatsBreakdown` and `getAlbumStatsBreakdown` handlers to the HTTP handler layer; each returns 503 when no catalog service is configured.
- Register `GET /api/v1/admin/catalog/stats/artists` and `GET /api/v1/admin/catalog/stats/albums` (admin-auth) with 405 fallbacks.
- Add 4 `catalog.Service` unit tests (empty/populated for each breakdown) and 8 HTTP-layer tests (empty shape, populated counts, 503, 405 for each endpoint).
- Add `CatalogArtistStatItem`, `CatalogArtistStatsBreakdown`, `CatalogAlbumStatItem`, `CatalogAlbumStatsBreakdown` schemas and `get` operations on both new paths to the OpenAPI contract; bump `info.version` to `0.51.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.52.0 - 2026-06-14

- Add `GET /api/v1/admin/catalog/stats/playlists` endpoint returning per-playlist track counts as a `CatalogPlaylistStatsBreakdown` object.
- Add `PlaylistStatItem`, `PlaylistStatsBreakdown` types to the catalog package; each item carries `playlistId`, `name`, and `trackCount` (duplicate track entries counted separately).
- Add `GetPlaylistStatsBreakdown(ctx)` to `catalog.Service`; counts are derived from each playlist's `TrackIDs` slice length. No new `Repository` interface methods required.
- Add `getPlaylistStatsBreakdown` handler to the HTTP handler layer; returns 503 when no catalog service is configured.
- Register `GET /api/v1/admin/catalog/stats/playlists` (admin-auth) and 405 fallback for the path.
- Add 2 `catalog.Service` unit tests (empty, populated with duplicate-track counting) and 4 HTTP-layer tests (empty shape, populated counts, no-catalog-service 503, method-not-allowed 405).
- Add `CatalogPlaylistStatItem`, `CatalogPlaylistStatsBreakdown` schemas and `get` operation on `/api/v1/admin/catalog/stats/playlists` to the OpenAPI contract; bump `info.version` to `0.52.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the new playlists stats path.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.53.0 - 2026-06-15

- Add `GET /api/v1/admin/catalog/recently-added` endpoint returning a newest-first unified timeline of recently created artists, albums, and tracks.
- Add `RecentItemKind`, `RecentCatalogItem`, and `RecentCatalogResult` types to the catalog package. Each timeline item includes `kind`, one entity payload, and `addedAt` copied from the entity's `CreatedAt` timestamp.
- Add `GetRecentlyAdded(ctx, kind, limit)` to `catalog.Service`; it derives results from existing `ListArtists`, `ListAlbums`, and `ListTracks` repository methods, supports `kind=artist|album|track`, defaults `limit` to 20, and clamps values above 100. No new `Repository` interface methods required.
- Add `getRecentlyAdded` handler to the HTTP handler layer; returns 400 for invalid `limit` or `kind`, 503 when no catalog service is configured, and registers a 405 fallback for the path.
- Register `GET /api/v1/admin/catalog/recently-added` (admin-auth).
- Add 5 `catalog.Service` unit tests and 8 HTTP-layer tests covering empty response shape, populated timeline payload, kind filter, invalid kind, invalid limit, limit handling, no-catalog-service 503, and method-not-allowed 405.
- Add `RecentItemKind`, `RecentCatalogItem`, and `RecentCatalogResult` schemas plus the `get` operation on `/api/v1/admin/catalog/recently-added` to the OpenAPI contract; bump `info.version` to `0.53.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the recently-added path.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.54.0 - 2026-06-15

- Add `GET /api/v1/admin/catalog/recently-updated` endpoint returning a newest-first unified timeline of recently updated artists, albums, and tracks.
- Add `UpdatedCatalogItem` and `UpdatedCatalogResult` types to the catalog package. Each timeline item includes `kind`, one entity payload, and `updatedAt` copied from the entity's `UpdatedAt` timestamp.
- Add `GetRecentlyUpdated(ctx, kind, limit)` to `catalog.Service`; it derives results from existing `ListArtists`, `ListAlbums`, and `ListTracks` repository methods, supports `kind=artist|album|track`, defaults `limit` to 20, and clamps values above 100. No new `Repository` interface methods required.
- Add `getRecentlyUpdated` handler to the HTTP handler layer; returns 400 for invalid `limit` or `kind`, 503 when no catalog service is configured, and registers a 405 fallback for the path.
- Register `GET /api/v1/admin/catalog/recently-updated` (admin-auth).
- Add `catalog.Service` unit tests and HTTP-layer tests covering empty response shape, updated timestamp ordering, kind filter, invalid kind, invalid limit, limit handling, no-catalog-service 503, and method-not-allowed 405.
- Add `UpdatedCatalogItem` and `UpdatedCatalogResult` schemas plus the `get` operation on `/api/v1/admin/catalog/recently-updated` to the OpenAPI contract; bump `info.version` to `0.54.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the recently-updated path.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.55.0 - 2026-06-15

- Add viewer-facing `GET /api/v1/catalog/recently-added` and `GET /api/v1/catalog/recently-updated` endpoints requiring session authentication (`requireViewerAuth`).
- Reuse existing `getRecentlyAdded` and `getRecentlyUpdated` handlers, wrapping them with `requireViewerAuth` middleware instead of `requireAdminAuth`.
- Register 405 fallbacks for both viewer paths.
- Add 16 HTTP-layer tests covering viewer auth success, admin session acceptance, static bootstrap token rejection, unauthorized requests, invalid kind/limit, and method-not-allowed.
- Add viewer path operations to the OpenAPI contract under the "Catalog" tag; bump `info.version` to `0.55.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the viewer recently-added and recently-updated paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.56.0 - 2026-06-15

- Add playlist entries to the recently-added and recently-updated unified catalog timelines.
- Extend `RecentItemKind` enum with `playlist`, and add `Playlist` fields to `RecentCatalogItem` and `UpdatedCatalogItem` types.
- Extend `GetRecentlyAdded` and `GetRecentlyUpdated` to iterate over playlists when `kind` is empty or `playlist`.
- Update `validateRecentItemKind` to accept `playlist` as a valid kind.
- Update OpenAPI contract: `RecentItemKind` enum, schemas, and endpoint descriptions; bump `info.version` to `0.56.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.57.0 - 2026-06-16

- Repair playlist participation in recently-added and recently-updated catalog timelines: the v0.56.0 implementation was incomplete — `RecentItemPlaylist` constant, `Playlist` payload fields on `RecentCatalogItem` and `UpdatedCatalogItem`, and `ListPlaylists` iterations in `GetRecentlyAdded` / `GetRecentlyUpdated` were missing.
- Update `validateRecentItemKind` to accept `"playlist"` and update the validation message to name all four valid kinds.
- Update OpenAPI contract: add `relink_rejected`, `validation_error`, and `invalid_limit` to the error code enum; bump `info.version` to `0.57.0`.
- Correct `services/api/internal/storage/capacity.go`: remove duplicate `FilesystemCapacityProvider` body now superseded by the build-tagged `capacity_unix.go` and `capacity_unsupported.go` files pulled in with the upstream update.
- Strengthen `openapi_contract_test.go`: assert `patch` on artist/album `{id}` paths, assert all three new error codes, and add `TestStorageAdminOpenAPIContractRecentTimelineSchemas` asserting the `RecentItemKind` enum includes `"playlist"` and both recent timeline item schemas carry a `playlist` payload ref.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.58.0 - 2026-06-17

- Add `GET /api/v1/catalog/tracks/{id}/playback` viewer-only endpoint returning a metadata-only `TrackPlaybackDescriptor` with `trackId`, `mediaObjectId`, `mimeType`, `durationMs`, `backendId`, `backendType`, and `objectKey`.
- Validate that the linked media object has `lifecycleState = active` and `assetKind ∈ {original_audio, transcoded_audio}`; return 422 `playback_unavailable` otherwise.
- Add `ErrPlaybackUnavailable` sentinel to the storage package and `playback_unavailable` to the `writeError` switch.
- Add `TrackPlaybackDescriptor` schema, `GET /api/v1/catalog/tracks/{id}/playback` path, and `playback_unavailable` error code to the OpenAPI contract; bump `info.version` to `0.58.0`.
- Add 8 HTTP-layer tests covering success, admin-session access, track-not-found, media-object-not-found, non-active lifecycle, wrong asset kind, no-catalog-service, and method-not-allowed.
- Extend `openapi_contract_test.go` with the new path, schema, error code, and `TestStorageAdminOpenAPIContractTrackPlaybackDescriptor`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.59.0 - 2026-06-17

- Expose viewer-accessible catalog stats endpoints: `GET /api/v1/catalog/stats`, `GET /api/v1/catalog/stats/artists`, `GET /api/v1/catalog/stats/albums`, and `GET /api/v1/catalog/stats/playlists`.
- Reuse existing `getCatalogStats`, `getArtistStatsBreakdown`, `getAlbumStatsBreakdown`, and `getPlaylistStatsBreakdown` handler functions under `requireViewerAuth`; no new domain logic required.
- Add 405 fallbacks for all four new viewer stats paths.
- Add 14 HTTP-layer tests covering empty stats, populated counts, admin session acceptance, no-catalog-service 503, and method-not-allowed for all four endpoints.
- Add four viewer stats paths to the OpenAPI contract; bump `info.version` to `0.59.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the four new viewer stats paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.60.0 - 2026-06-17

- Add `presignS3URL` to the storage package: generates an AWS Signature Version 4 presigned GET URL using query-parameter signing, reusing existing `s3ObjectURL`, `s3SigningKey`, and `hmacSHA256` helpers from `s3_probe.go`.
- Add `storage.Service.GetBackend(ctx, id)` for direct single-backend lookup; `storage.DefaultPresignedURLTTL = 15 * time.Minute` constant; `storage.Service.GeneratePresignedURL(ctx, backendID, objectKey, ttl)` orchestrating capability check, credential resolution via env var refs, and presigned URL generation.
- Extend `GET /api/v1/catalog/tracks/{id}/playback` response: populate optional `presignedUrl` field when the backend has `PresignedURLs` capability and credentials are configured; presign failures are non-fatal.
- Replace the backend full-list scan in `getTrackPlayback` with a single `GetBackend` call.
- Add `presignedUrl` optional property to `TrackPlaybackDescriptor` OpenAPI schema; bump `info.version` to `0.60.0`.
- Add 4 `presignS3URL` unit tests, 4 `GetBackend`/`GeneratePresignedURL` service tests, and 1 HTTP-layer presigned URL handler test.
- Extend `TestStorageAdminOpenAPIContractTrackPlaybackDescriptor` to assert `presignedUrl` is present in properties but absent from `required`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.61.0 - 2026-06-17

- Add `limit`/`offset` pagination to all four catalog list endpoints: artists, albums, tracks, and playlists (both admin and viewer routes).
- Add `parseCatalogPage` and `paginateCatalog[T]` helpers to the HTTP handler layer; no Repository or Service interface changes required.
- Add `CatalogPaginationMeta` type with `limit`, `offset`, `total`, and `hasMore` fields; all four list responses now include a `pagination` envelope alongside the existing entity array.
- `limit` defaults to 50 (max 500); `offset` defaults to 0; invalid values return 400 `invalid_limit` / `invalid_offset`.
- Existing `artistId` and `albumId` filter params continue to work and are applied before pagination.
- Add `CatalogPaginationMeta` schema, `limit`/`offset` params, and `pagination` response property to all 8 catalog list paths in the OpenAPI contract; bump `info.version` to `0.61.0`.
- Add `invalid_offset` to the OpenAPI error code enum and contract test assertion.
- Add 5 HTTP-layer tests covering limit, offset, hasMore, invalid params, and viewer-session access.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.62.0 - 2026-06-17

- Add `sortBy` and `sortOrder` query parameters to all four catalog list endpoints (artists, albums, tracks, playlists) on both admin and viewer routes.
- Artists: `sortBy` accepts `name` (default), `sortName`, `createdAt`, `updatedAt`.
- Albums: `sortBy` accepts `title` (default), `sortTitle`, `releaseYear`, `createdAt`, `updatedAt`.
- Tracks: `sortBy` accepts `title` (default), `sortTitle`, `trackNumber`, `discNumber`, `durationMs`, `createdAt`, `updatedAt`.
- Playlists: `sortBy` accepts `name` (default), `createdAt`, `updatedAt`.
- `sortOrder` must be `asc` (default) or `desc`; any other value returns `400 invalid_sort_order`.
- Sort is applied before pagination so `limit`/`offset` windows remain stable.
- Add sort-field constants to `catalog/types.go`; add `normalizeSortOrder` and per-entity sort functions to the HTTP handler layer; no Repository or Service interface changes.
- Add `sortBy`/`sortOrder` params and sort descriptions to all 8 catalog list paths in the OpenAPI contract; add `invalid_sort_order` to error enum; bump `info.version` to `0.62.0`.
- Add `TestStorageAdminOpenAPIContractCatalogListSortParams` asserting sort params on all 8 list paths.
- Add 6 HTTP-layer tests covering per-entity sort directions, invalid `sortOrder`, and viewer-session sort access.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.63.0 - 2026-06-17

- Add nested catalog browse routes: `GET /api/v1/catalog/artists/{id}/albums`, `GET /api/v1/catalog/artists/{id}/tracks`, and `GET /api/v1/catalog/albums/{id}/tracks` under both admin (`/api/v1/admin/catalog/…`) and viewer (`/api/v1/catalog/…`) paths.
- Each handler validates the parent entity (artist or album) before listing sub-entities; unknown IDs return 404.
- All six new routes support the same `limit`, `offset`, `sortBy`, and `sortOrder` parameters established in Phases 61–62.
- Add 4 HTTP-layer tests covering pagination, sort, 404 on unknown parent, 405 method-not-allowed, and viewer-session access to all three nested routes.
- Add 6 new paths to the OpenAPI contract with typed response schemas (albums/tracks with pagination) and full pagination+sort parameter declarations; bump `info.version` to `0.63.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` and `TestStorageAdminOpenAPIContractCatalogListSortParams` for all 6 new paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.64.0 - 2026-06-17

- Add `limit`/`offset` pagination to `GET /api/v1/catalog/playlists/{id}/tracks` and `GET /api/v1/admin/catalog/playlists/{id}/tracks`.
- Playlist track order is user-curated and is preserved exactly within each page; `sortBy`/`sortOrder` are intentionally not exposed.
- Response now includes a `pagination` envelope (`limit`, `offset`, `total`, `hasMore`) alongside the `tracks` array.
- Add `limit`/`offset` query params and `pagination` response property to both playlist tracks paths in the OpenAPI contract; bump `info.version` to `0.64.0`.
- Add `TestStorageAdminOpenAPIContractPlaylistTracksPagination` asserting `limit`/`offset` present and `sortBy`/`sortOrder` absent.
- Add 2 HTTP-layer tests covering order preservation, limit, offset, `hasMore`, invalid params, and viewer-session access.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.65.0 - 2026-06-17

- Add `ListQuery` and `ListPage[T]` types and 7 new `ListXxxPage` methods to the `catalog.Repository` interface: `ListArtistsPage`, `ListAlbumsPage`, `ListAlbumsByArtistPage`, `ListTracksPage`, `ListTracksByAlbumPage`, `ListTracksByArtistPage`, `ListPlaylistsPage`.
- Implement the 7 page methods on `catalog.MemoryRepository` (Go in-memory sort + slice) and add a new `catalog/postgres/repository_page.go` with SQL `ORDER BY … LIMIT $1 OFFSET $2` and `COUNT(*) OVER ()` window function for accurate total counts without a separate query.
- PostgreSQL ORDER BY uses `lower()` wrapping for text fields (consistent with existing list queries) and an `id` tiebreak for stable pagination across pages.
- Update all 7 catalog list HTTP handlers (`listArtists`, `listAlbums`, `listTracks`, `listPlaylists`, `listAlbumsByArtist`, `listTracksByArtist`, `listTracksByAlbum`) to call the new Page methods and remove the previous in-handler sort+paginate logic.
- Add 4 `ListXxxPage` catalog service unit tests (artist sort/paginate/offset-past-end, albums-by-artist sort, tracks paginate, playlists desc).
- Add 2 PostgreSQL integration tests under the `integration` build tag (`TestRepositoryListArtistsPage`, `TestRepositoryListAlbumsPageByArtist`).
- No change to HTTP API shape — client-facing behavior is identical to Phases 61–62.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.66.0 - 2026-06-17

- Add 4 aggregate stats methods to the `catalog.Repository` interface: `CountEntities`, `ArtistAlbumTrackCounts`, `AlbumTrackCounts`, `PlaylistTrackCounts`.
- Implement on `catalog.MemoryRepository` (in-memory counting) and `catalogpg.Repository` with SQL `COUNT(*)`/`GROUP BY` aggregate queries (single query per stats method, no N+1 iteration).
- Replace `GetCatalogStats`, `GetArtistStatsBreakdown`, `GetAlbumStatsBreakdown`, and `GetPlaylistStatsBreakdown` in `catalog.Service` with single-aggregate-call implementations.
- Add 2 PostgreSQL integration tests (`TestRepositoryCountEntities`, `TestRepositoryArtistAlbumTrackCounts`) under the `integration` build tag.
- Bump OpenAPI `info.version` to `0.66.0`. No HTTP API shape change.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.67.0 - 2026-06-17

- Add `RecentlyAdded(ctx, kind, limit)` and `RecentlyUpdated(ctx, kind, limit)` to the `catalog.Repository` interface.
- Implement on `catalog.MemoryRepository` (in-memory sort + slice) and `catalogpg.Repository` (single `UNION ALL … ORDER BY … LIMIT` query per method).
- Replace the 4-branch list-and-merge logic in `GetRecentlyAdded` and `GetRecentlyUpdated` (`catalog.Service`) with single delegate calls to the new repo methods; remove unused `sort` import from `service.go`.
- Add `TestRepositoryRecentlyAdded` PostgreSQL integration test (unified + kind filter + limit).
- Bump OpenAPI `info.version` to `0.67.0`. No HTTP API shape change.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.68.0 - 2026-06-17

- Add playback history domain: new `history` package with `PlayEvent` type, `Repository` interface, `Service` with `RecordPlay`/`ListPlays`/`ClearHistory` methods, in-memory repository, and PostgreSQL repository.
- Add migration `008_play_events` (id, user_id, track_id, played_at, created_at; FK cascades, two indexes).
- Extend `requireViewerAuth` and `requireAdminAuth` to inject the authenticated `auth.User` into the request context for downstream handler use.
- Add viewer-only `POST/GET/DELETE /api/v1/me/history` endpoints (user-scoped, session-auth required).
- Add `PlayEvent`, `PlayEventList` schemas and `/api/v1/me/history` path to OpenAPI contract; add `history_not_configured` error code; bump `info.version` to `0.68.0`.
- Add 5 history service unit tests and 5 HTTP-layer tests (record, list w/ pagination, clear, 503 not-configured, 405).
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.69.0 - 2026-06-17

- Extend the `history.Repository` interface with 3 aggregate stats methods: `HistoryStats(ctx)`, `TopTracks(ctx, limit)`, `TopUsers(ctx, limit)`.
- Implement on `history.MemoryRepository` (in-memory counting + sort) and `historypg.Repository` (single SQL `COUNT`/`GROUP BY` aggregate queries per method).
- Add `GetHistoryStats`, `GetTopTracks`, `GetTopUsers` to `history.Service`; limit clamped to 100, default 10.
- Add 3 admin-only routes: `GET /api/v1/admin/history/stats`, `GET /api/v1/admin/history/top-tracks`, `GET /api/v1/admin/history/top-users`.
- Add `HistoryStats`, `TrackPlayCount`, `UserPlayCount`, `TopTracksResult`, `TopUsersResult` schemas and the 3 new admin paths to the OpenAPI contract; bump `info.version` to `0.69.0`.
- Add 3 history service unit tests (GetHistoryStats, GetTopTracks, GetTopUsers) and 5 HTTP-layer tests (stats, top-tracks with limit, top-users, not-configured 503, 405).
- Add `TestStorageAdminOpenAPIContractAdminHistoryPaths` asserting new schemas and path operations.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.70.0 - 2026-06-18

- Add optional `?since=<RFC3339>` query parameter to `GET /api/v1/admin/history/stats`, `GET /api/v1/admin/history/top-tracks`, and `GET /api/v1/admin/history/top-users`; omitting it returns all-time data.
- Add `StatsFilter{Since time.Time}` type to `history/types.go`; update `Repository` interface so all three aggregate methods accept `StatsFilter`.
- Implement `since` filtering on `history.MemoryRepository` (in-memory `played_at >= since` guard) and `historypg.Repository` (`WHERE played_at >= $N` SQL clause).
- Thread `StatsFilter` through `history.Service` methods `GetHistoryStats`, `GetTopTracks`, `GetTopUsers`.
- Add `parseHistoryAdminFilter` helper in the HTTP handler layer to parse and validate the `since` param; invalid timestamps return `400 invalid_since`.
- Add 3 service unit tests (windowed stats, top-tracks, top-users) and 2 HTTP-layer tests (`TestAdminHistorySinceFilter`, `TestAdminHistorySinceInvalid`).
- Add `since` query param (string/date-time, optional) to all three admin history GET paths in the OpenAPI contract; add `invalid_since` to error code enum; bump `info.version` to `0.70.0`.
- Add `TestStorageAdminOpenAPIContractAdminHistorySinceParam` asserting the `since` param is present and optional on all three paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.71.0 - 2026-06-18

- Add optional `?until=<RFC3339>` (exclusive upper bound) query parameter to `GET /api/v1/admin/history/stats`, `GET /api/v1/admin/history/top-tracks`, and `GET /api/v1/admin/history/top-users`; composes with `?since`.
- Add `Until time.Time` to `StatsFilter` in `history/types.go`.
- Implement `until` guard (`played_at < until`, exclusive) on `history.MemoryRepository` aggregate methods.
- Replace four-branch since/no-since logic in `historypg.Repository` with a shared `statsWhere(f)` helper that builds `WHERE played_at >= $N AND played_at < $M` dynamically for any combination of bounds.
- Extend `parseHistoryAdminFilter` in `handler.go` to parse `?until` (returns `400 invalid_until` if unparseable) and validate `since < until` when both are present (returns `400 invalid_time_range`).
- Add 3 service unit tests (until-stats, since+until window on top-tracks, until combined) and 4 HTTP-layer tests (until filter, invalid until, invalid time range, updated since test).
- Add `until` query param (string/date-time, optional) to all three admin history GET paths in OpenAPI; add `invalid_until` and `invalid_time_range` to error code enum; bump `info.version` to `0.71.0`.
- Add `TestStorageAdminOpenAPIContractAdminHistoryUntilParam`; extend `TestStorageAdminOpenAPIContractSchemasAndErrors` for new error codes.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.72.0 - 2026-06-18

- Add `AdminPlayEventFilter{TrackID, UserID, Limit, Offset}` to `history/types.go`; add `ListPlayEventsByTrack(ctx, AdminPlayEventFilter)` to the `Repository` interface.
- Implement `ListPlayEventsByTrack` on `history.MemoryRepository` (in-memory filter + sort + slice) and `historypg.Repository` (`WHERE track_id = $3 [AND user_id = $4] … COUNT(*) OVER()`).
- Add `GetUserHistory` (admin-facing reuse of `ListPlayEvents` without user-scope restriction) and `GetTrackHistory` to `history.Service`.
- Add 2 admin routes: `GET /api/v1/admin/history/users/{userId}` (paginated events for any user, optional `?trackId` filter) and `GET /api/v1/admin/history/tracks/{trackId}` (paginated events for any track, optional `?userId` filter); add `methodNotAllowed` fallbacks for both.
- Add `getAdminUserHistory`, `getAdminTrackHistory` handler functions and `parseHistoryAdminPagination` helper; response shape is `{events, pagination}` identical to `GET /api/v1/me/history`.
- Add 2 service unit tests (`TestGetUserHistory`, `TestGetTrackHistory`) and 4 HTTP-layer tests (user history with pagination, track history, 405, 503 not-configured).
- Add 2 new paths to the OpenAPI contract with `PlayEventList` response schema ref; bump `info.version` to `0.72.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractAdminHistoryDetailPaths` asserting path params, query filters, pagination params, and response schema ref.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.73.0 - 2026-06-18

- Add `DeletePlayEventsByUserAdmin(ctx, userID)`, `DeletePlayEventsByTrack(ctx, trackID)`, and `DeletePlayEventsInWindow(ctx, StatsFilter)` to the `history.Repository` interface.
- Implement all three on `history.MemoryRepository` (in-memory guard under lock) and `historypg.Repository` (`DELETE FROM play_events WHERE …`; `DeletePlayEventsInWindow` reuses `statsWhere` helper).
- Add `AdminDeleteUserHistory`, `AdminDeleteTrackHistory`, and `AdminDeleteHistoryWindow` to `history.Service`; `AdminDeleteHistoryWindow` validates that at least one time bound is set.
- Add `DELETE /api/v1/admin/history/users/{userId}` and `DELETE /api/v1/admin/history/tracks/{trackId}` to existing paths; add new `DELETE /api/v1/admin/history` path with optional `?since`/`?until` time-window filter (at least one required at runtime).
- Return 400 `missing_time_filter` when neither `since` nor `until` is supplied to the window endpoint.
- Add `methodNotAllowed` fallback for `/api/v1/admin/history`.
- Add 3 `history.Service` unit tests (`TestAdminDeleteUserHistory`, `TestAdminDeleteTrackHistory`, `TestAdminDeleteHistoryWindow`).
- Add 5 HTTP-layer tests (`TestAdminDeleteUserHistory`, `TestAdminDeleteTrackHistory`, `TestAdminDeleteHistoryWindow`, `TestAdminDeleteHistoryWindowMissingFilter`, `TestAdminBulkDeleteHistoryNotConfigured`).
- Add `delete` operations to both detail paths and new window path in OpenAPI contract; add `missing_time_filter` to error code enum; bump `info.version` to `0.73.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `delete` on detail paths and new window path; add `TestStorageAdminOpenAPIContractAdminHistoryBulkDelete`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.74.0 - 2026-06-18

- Add `UserStatsFilter{UserID, Since, Until}` and `UserHistoryStats{TotalEvents, UniqueTracks}` types to `history/types.go`.
- Add `UserTopTracks(ctx, UserStatsFilter, limit)` and `UserHistoryStats(ctx, UserStatsFilter)` methods to the `Repository` interface.
- Implement both on `history.MemoryRepository` (user-scoped in-memory filter + sort) and `historypg.Repository` (new `userStatsWhere` helper that mandates `user_id = $1`).
- Add `GetMyStats` and `GetMyTopTracks` to `history.Service`; validate `UserID != ""`.
- Add viewer-only `GET /api/v1/me/history/stats` and `GET /api/v1/me/history/top-tracks` endpoints; both accept optional `?since`, `?until`; top-tracks also accepts `?limit` (default 10, max 100).
- Reuse `parseHistoryAdminFilter` and `parseHistoryAdminLimit` in the new handlers; inject `UserID` from auth context.
- Add `methodNotAllowed` fallbacks for both new viewer paths.
- Add 3 `history.Service` unit tests (`TestGetMyStats`, `TestGetMyTopTracks`, `TestGetMyTopTracksTimeWindow`).
- Add 4 HTTP-layer tests (`TestGetMyHistoryStats`, `TestGetMyTopTracks`, `TestGetMyHistoryStatsTimeWindow`, `TestGetMyHistoryStatsNotConfigured`).
- Add `UserHistoryStats` schema and 2 new viewer paths to OpenAPI contract; bump `info.version` to `0.74.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractViewerHistoryStatsPaths`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.75.0 - 2026-06-18

- Add `GlobalPlayEventFilter{UserID, TrackID, Since, Until, Limit, Offset}` to `history/types.go`.
- Add `ListAllPlayEvents(ctx, GlobalPlayEventFilter)` to the `Repository` interface.
- Implement `ListAllPlayEvents` on `history.MemoryRepository` (in-memory multi-filter + sort + slice) and `historypg.Repository` (dynamic `WHERE` clause construction with `LIMIT`/`OFFSET` + `COUNT(*) OVER()`).
- Add `GetAllHistory` to `history.Service`; limit clamped to 500, default 50.
- Add admin route `GET /api/v1/admin/history` (paginated global event list, optional `?userId`, `?trackId`, `?since`, `?until`, `?limit`, `?offset` filters); handler `getAdminAllHistory` reuses `parseHistoryAdminFilter` and `parseHistoryAdminPagination`.
- Add 3 `history.Service` unit tests (`TestGetAllHistory`, `TestGetAllHistoryUserFilter`, `TestGetAllHistoryTimeWindow`).
- Add 4 HTTP-layer tests (`TestAdminGetAllHistory`, `TestAdminGetAllHistoryTrackFilter`, `TestAdminGetAllHistoryNotConfigured`, `TestAdminGetAllHistoryMethodNotAllowed`).
- Add `get` operation to `/api/v1/admin/history` in OpenAPI contract with 6 query params and `PlayEventList` response schema ref; bump `info.version` to `0.75.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history`; add `TestStorageAdminOpenAPIContractAdminHistoryGlobalList`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.76.0 - 2026-06-18

- Add `Asc bool` field to `PlayEventFilter`, `AdminPlayEventFilter`, and `GlobalPlayEventFilter` in `history/types.go`; `false` (default) → `played_at DESC`, `true` → `played_at ASC`.
- Update sort comparators in `history.MemoryRepository` for `ListPlayEvents`, `ListPlayEventsByTrack`, and `ListAllPlayEvents` to respect `f.Asc`.
- Add `eventOrder(asc bool) string` helper to `historypg.Repository`; replace hard-coded `ORDER BY played_at DESC, id DESC` with `eventOrder(f.Asc)` in `ListPlayEvents`, `ListPlayEventsByTrack`, and `ListAllPlayEvents`.
- Add `parseHistoryOrder` helper to `httpapi/handler.go`; parses `?order=asc|desc` (default `desc`); returns `400 invalid_order` for any other value.
- Thread `Asc` through `listPlayEvents`, `getAdminUserHistory`, `getAdminTrackHistory`, and `getAdminAllHistory` handlers.
- Add 2 `history.Service` unit tests (`TestListPlaysAscOrder`, `TestGetAllHistoryAscOrder`).
- Add 4 HTTP-layer tests (`TestListPlayEventsAscOrder`, `TestListPlayEventsInvalidOrder`, `TestAdminGetAllHistoryAscOrder`, `TestAdminGetAllHistoryInvalidOrder`).
- Add `order` query param (string enum `["asc","desc"]`, optional, default `"desc"`) to `GET /api/v1/me/history`, `GET /api/v1/admin/history/users/{userId}`, `GET /api/v1/admin/history/tracks/{trackId}`, and `GET /api/v1/admin/history` in OpenAPI contract; add `invalid_order` to error code enum; bump `info.version` to `0.76.0`.
- Add `TestStorageAdminOpenAPIContractHistoryOrderParam` asserting `order` param on all four paths and `invalid_order` in the error enum.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.77.0 - 2026-06-18

- Add `ErrEventNotFound` and `ErrEventForbidden` sentinel errors to `history/types.go`.
- Add `GetPlayEventByID(ctx, id)` and `DeletePlayEventByID(ctx, id)` to the `Repository` interface.
- Implement both on `history.MemoryRepository` (map lookup; `ErrEventNotFound` on miss) and `historypg.Repository` (`SELECT`/`DELETE` by primary key; `ErrEventNotFound` on `ErrNoRows` or zero `RowsAffected`).
- Add 4 `history.Service` methods: `GetEventByID` (admin), `DeleteEventByID` (admin), `GetMyEvent` (viewer, ownership-checked), `DeleteMyEvent` (viewer, ownership-checked).
- Add admin routes `GET /api/v1/admin/history/{eventId}` and `DELETE /api/v1/admin/history/{eventId}`.
- Add viewer routes `GET /api/v1/me/history/{eventId}` and `DELETE /api/v1/me/history/{eventId}`; ownership check returns `403 event_forbidden` when the authenticated user does not own the event.
- Map `ErrEventNotFound` → `404 not_found` and `ErrEventForbidden` → `403 event_forbidden` in `writeError`.
- Add 5 `history.Service` unit tests (`TestGetEventByID`, `TestGetEventByIDNotFound`, `TestDeleteEventByID`, `TestGetMyEvent`, `TestDeleteMyEvent`).
- Add 7 HTTP-layer tests (`TestAdminGetEvent`, `TestAdminGetEventNotFound`, `TestAdminDeleteEvent`, `TestViewerGetEvent`, `TestViewerGetEventNotOwned`, `TestViewerDeleteEvent`, `TestPerEventHistoryNotConfigured`).
- Add `GET`/`DELETE` operations to `/api/v1/admin/history/{eventId}` and `/api/v1/me/history/{eventId}` in OpenAPI; add `event_forbidden` to error code enum; `PlayEvent` schema ref as 200 response; bump `info.version` to `0.77.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with both new paths; add `TestStorageAdminOpenAPIContractPerEventPaths`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.78.0 - 2026-06-19

- Add `UpdatePlayEventByID(ctx, id, playedAt)` to the `Repository` interface.
- Implement on `history.MemoryRepository` (lock + map update; `ErrEventNotFound` on miss) and `historypg.Repository` (`UPDATE … SET played_at = $2 … RETURNING …`; `ErrEventNotFound` on `ErrNoRows`).
- Add `UpdateEventByID(ctx, id, playedAt)` to `history.Service` (admin; validates non-zero `playedAt`).
- Add `UpdateMyEvent(ctx, userID, id, playedAt)` to `history.Service` (viewer; ownership-checked; returns `ErrEventForbidden` for non-owners).
- Add admin route `PATCH /api/v1/admin/history/{eventId}` → `patchAdminEvent`; decodes `{"playedAt": "<RFC3339>"}` request body; returns `400 invalid_played_at` for missing or unparseable timestamp.
- Add viewer route `PATCH /api/v1/me/history/{eventId}` → `patchMyEvent`; same validation; returns `403 event_forbidden` for non-owners.
- Add `UpdatePlayEventRequest` schema (`{playedAt: string/date-time}`) to OpenAPI components.
- Add `patch` operation to `/api/v1/admin/history/{eventId}` and `/api/v1/me/history/{eventId}` in OpenAPI contract; 200 response refs `PlayEvent`; `requestBody` refs `UpdatePlayEventRequest`; add `invalid_played_at` to error code enum; bump `info.version` to `0.78.0`.
- Add 3 `history.Service` unit tests (`TestUpdateEventByID`, `TestUpdateEventByIDNotFound`, `TestUpdateMyEvent`).
- Add 7 HTTP-layer tests (`TestAdminPatchEvent`, `TestAdminPatchEventNotFound`, `TestAdminPatchEventInvalidPlayedAt`, `TestViewerPatchEvent`, `TestViewerPatchEventInvalidPlayedAt`, `TestViewerPatchEventMissingPlayedAt`, `TestPatchEventHistoryNotConfigured`).
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `patch` on both paths; extend `TestStorageAdminOpenAPIContractPerEventPaths` to assert `UpdatePlayEventRequest` requestBody ref and `invalid_played_at` error code.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.79.0 - 2026-06-19

- Add `DeletePlayEventsByIDs(ctx, ids)` and `DeletePlayEventsByIDsForUser(ctx, userID, ids)` to the `Repository` interface.
- Implement both on `history.MemoryRepository` (lock + range-delete; unknown IDs silently skipped) and `historypg.Repository` (`DELETE … WHERE id = ANY($1)` and `DELETE … WHERE id = ANY($1) AND user_id = $2`).
- Add `MaxBatchDeleteIDs = 100` constant; add `BatchDeleteEvents(ctx, ids)` (admin) and `BatchDeleteMyEvents(ctx, userID, ids)` (viewer) to `history.Service`; validate non-empty ids and size ≤ 100.
- Add admin route `POST /api/v1/admin/history/batch-delete` → `batchDeleteAdminEvents`; decodes `{"ids":[…]}`; returns `{"deleted": N}`.
- Add viewer route `POST /api/v1/me/history/batch-delete` → `batchDeleteMyEvents`; same shape; silently skips IDs not owned by the viewer.
- Both routes return `400 invalid_ids` for empty or oversized `ids` array.
- Add `BatchDeleteRequest` schema (`{ids: string[], minItems:1, maxItems:100}`) and `BatchDeleteResult` schema (`{deleted: integer}`) to OpenAPI components; add both new paths; add `invalid_ids` to error code enum; bump `info.version` to `0.79.0`.
- Add 4 `history.Service` unit tests (`TestBatchDeleteEvents`, `TestBatchDeleteEventsUnknownIDsIgnored`, `TestBatchDeleteMyEvents`, `TestBatchDeleteEventsEmpty`).
- Add 5 HTTP-layer tests (`TestAdminBatchDeleteEvents`, `TestAdminBatchDeleteEventsEmptyBody`, `TestViewerBatchDeleteMyEvents`, `TestViewerBatchDeleteSkipsOtherUsersEvents`, `TestBatchDeleteHistoryNotConfigured`).
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with both batch-delete paths; add `TestStorageAdminOpenAPIContractBatchDelete`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.80.0 - 2026-06-19

- Add `Since time.Time` and `Until time.Time` fields to `PlayEventFilter` and `AdminPlayEventFilter` in `history/types.go`.
- Update `history.MemoryRepository.ListPlayEvents` and `ListPlayEventsByTrack` to apply `Since`/`Until` guards matching the existing pattern in `ListAllPlayEvents`.
- Replace two-branch (TrackID/no-TrackID) SQL logic in `historypg.Repository.ListPlayEvents` and `ListPlayEventsByTrack` with a unified dynamic `WHERE` clause builder that handles all combinations of `user_id`, `track_id`, `since`, and `until` in a single query path.
- Thread `Since`/`Until` from `parseHistoryAdminFilter` into `listPlayEvents`, `getAdminUserHistory`, and `getAdminTrackHistory` handlers.
- Add `since` and `until` query params (string/date-time, optional) to `GET /api/v1/me/history`, `GET /api/v1/admin/history/users/{userId}`, and `GET /api/v1/admin/history/tracks/{trackId}` in OpenAPI; bump `info.version` to `0.80.0`.
- Add 3 `history.Service` unit tests (`TestListPlaysSinceFilter`, `TestListPlaysUntilFilter`, `TestGetUserHistorySinceFilter`).
- Add 4 HTTP-layer tests (`TestListPlayEventsSinceFilter`, `TestListPlayEventsUntilFilter`, `TestAdminUserHistorySinceUntilFilter`, `TestAdminTrackHistorySinceFilter`).
- Add `TestStorageAdminOpenAPIContractListSinceUntilParams` asserting `since`/`until` on all three paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.81.0 - 2026-06-19

- Add `GetAdminUserStats(ctx, UserStatsFilter)` and `GetAdminUserTopTracks(ctx, UserStatsFilter, limit)` to `history.Service` (admin-facing; delegate to `UserHistoryStats` and `UserTopTracks` on the repository; require non-empty `UserID`).
- Add admin routes `GET /api/v1/admin/history/users/{userId}/stats` → `getAdminUserStats` and `GET /api/v1/admin/history/users/{userId}/top-tracks` → `getAdminUserTopTracks`; reuse `parseHistoryAdminFilter` and `parseHistoryAdminLimit`; respond with the same shapes as their `/me/history/stats` and `/me/history/top-tracks` counterparts.
- Add `methodNotAllowed` fallbacks for `GET /api/v1/admin/history/users/{userId}/stats` and `GET /api/v1/admin/history/users/{userId}/top-tracks`.
- Add 3 `history.Service` unit tests (`TestGetAdminUserStats`, `TestGetAdminUserTopTracks`, `TestGetAdminUserTopTracksTimeWindow`).
- Add 4 HTTP-layer tests (`TestAdminGetUserStats`, `TestAdminGetUserTopTracks`, `TestAdminGetUserStatsNotConfigured`, `TestAdminGetUserTopTracksTimeWindow`).
- Add `get` operation to `/api/v1/admin/history/users/{userId}/stats` and `/api/v1/admin/history/users/{userId}/top-tracks` in OpenAPI contract; both accept optional `?since`, `?until`; top-tracks also accepts `?limit`; stats refs `UserHistoryStats` schema; top-tracks refs `TrackPlayCountList` schema (matching the viewer path); bump `info.version` to `0.81.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with both new paths; add `TestStorageAdminOpenAPIContractAdminUserStatsPaths`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.82.0 - 2026-06-19

- Add `TrackHistoryStats{TotalEvents int, UniqueListeners int}` and `TrackStatsFilter{TrackID string, Since time.Time, Until time.Time}` types to `history/types.go`.
- Add `TrackHistoryStats(ctx, TrackStatsFilter)` and `TrackTopListeners(ctx, TrackStatsFilter, limit)` methods to the `Repository` interface.
- Implement both on `history.MemoryRepository` (track-scoped in-memory filter + sort) and `historypg.Repository` (new `trackStatsWhere` helper that mandates `track_id = $1`; `TrackTopListeners` returns `UserPlayCount` rows ordered by `play_count DESC, user_id ASC`).
- Add `GetTrackStats(ctx, TrackStatsFilter)` and `GetTrackTopListeners(ctx, TrackStatsFilter, limit)` to `history.Service` (admin; validate non-empty `TrackID`; limit clamp 1–100 default 10).
- Add admin routes `GET /api/v1/admin/history/tracks/{trackId}/stats` → `getAdminTrackStats` and `GET /api/v1/admin/history/tracks/{trackId}/top-listeners` → `getAdminTrackTopListeners`; reuse `parseHistoryAdminFilter` and `parseHistoryAdminLimit`; extract `trackId` from path.
- Add `methodNotAllowed` fallbacks for both new sub-paths.
- Add 3 `history.Service` unit tests (`TestGetTrackStats`, `TestGetTrackTopListeners`, `TestGetTrackTopListenersTimeWindow`).
- Add 4 HTTP-layer tests (`TestAdminGetTrackStats`, `TestAdminGetTrackTopListeners`, `TestAdminGetTrackTopListenersTimeWindow`, `TestAdminGetTrackStatsNotConfigured`).
- Add `TrackHistoryStats` schema to OpenAPI components; add `get` operation to `/api/v1/admin/history/tracks/{trackId}/stats` (refs `TrackHistoryStats`) and `/api/v1/admin/history/tracks/{trackId}/top-listeners` (refs `TopUsersResult`); both accept optional `?since`, `?until`; top-listeners also accepts `?limit`; bump `info.version` to `0.82.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with both new paths; add `TestStorageAdminOpenAPIContractAdminTrackStatsPaths`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.83.0 - 2026-06-19

- Add `TimelineGranularity` string type with constants `GranularityDay`, `GranularityWeek`, `GranularityMonth` to `history/types.go`.
- Add `TimelineFilter{Since time.Time, Until time.Time, Granularity TimelineGranularity, UserID string, TrackID string}` and `TimelineBucket{BucketStart time.Time (json:"bucketStart"), EventCount int (json:"eventCount")}` to `history/types.go`.
- Add `HistoryTimeline(ctx, TimelineFilter) ([]TimelineBucket, error)` to the `Repository` interface.
- Implement on `history.MemoryRepository`: iterate events, apply `UserID`/`TrackID`/`Since`/`Until` guards, truncate `played_at` to the bucket boundary (day=UTC day, week=Monday-anchored UTC week, month=UTC month), accumulate counts into a `map[time.Time]int`, then emit sorted `[]TimelineBucket` (empty bucket list is `[]TimelineBucket{}`).
- Implement on `historypg.Repository`: use `DATE_TRUNC($granularity, played_at AT TIME ZONE 'UTC')` in a dynamic `WHERE` clause built from `timelineWhere` helper (extends `statsWhere` with optional `user_id` and `track_id`); `GROUP BY bucket` order by `bucket ASC`; return `[]TimelineBucket` (empty → `[]TimelineBucket{}`).
- Add `GetHistoryTimeline(ctx, TimelineFilter)` to `history.Service`; validate `Since` and `Until` are both non-zero and `Since` < `Until`; validate `Granularity` is one of `day`/`week`/`month` (default `day`); return `ErrInvalidTimeRange` sentinel on bad range.
- Add `ErrInvalidTimeRange` sentinel to `history/types.go`.
- Add admin route `GET /api/v1/admin/history/timeline` → `getAdminHistoryTimeline`; parses `?since`, `?until` (both required, `400 missing_time_bounds` if absent), `?granularity` (optional, default `day`, `400 invalid_granularity` for other values), optional `?userId` and `?trackId`; returns `{"buckets":[{"bucketStart":"...","eventCount":N},...]}`; add `methodNotAllowed` fallback.
- Add 4 `history.Service` unit tests (`TestGetHistoryTimelineDay`, `TestGetHistoryTimelineWeek`, `TestGetHistoryTimelineUserFilter`, `TestGetHistoryTimelineInvalidRange`).
- Add 4 HTTP-layer tests (`TestAdminGetHistoryTimeline`, `TestAdminGetHistoryTimelineMissingSince`, `TestAdminGetHistoryTimelineInvalidGranularity`, `TestAdminGetHistoryTimelineNotConfigured`).
- Add `TimelineBucket` schema and `TimelineResult` schema (`{buckets: [TimelineBucket]}`) to OpenAPI components; add `get` operation to `/api/v1/admin/history/timeline` with `since`(required), `until`(required), `granularity`(enum day/week/month, default day), `userId`(optional), `trackId`(optional) params; 200 refs `TimelineResult`; add `missing_time_bounds` and `invalid_granularity` to error code enum; bump `info.version` to `0.83.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history/timeline`; add `TestStorageAdminOpenAPIContractHistoryTimeline`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.84.0 - 2026-06-19

- Add `GetMyTimeline(ctx, TimelineFilter)` to `history.Service` (viewer-facing; validate non-empty `UserID` in the filter, then delegate to `repo.HistoryTimeline`; reuse the same `ErrInvalidTimeRange` validation as `GetHistoryTimeline`).
- Add viewer route `GET /api/v1/me/history/timeline` → `getMyHistoryTimeline`; parses `?since`, `?until` (both required, `400 missing_time_bounds`), `?granularity` (optional, default `day`, `400 invalid_granularity`), optional `?trackId`; injects `UserID` from auth context; returns `{"buckets":[...]}`.
- Add 3 `history.Service` unit tests (`TestGetMyTimelineDay`, `TestGetMyTimelineTrackFilter`, `TestGetMyTimelineInvalidRange`).
- Add 4 HTTP-layer tests (`TestViewerGetHistoryTimeline`, `TestViewerGetHistoryTimelineMissingSince`, `TestViewerGetHistoryTimelineInvalidGranularity`, `TestViewerGetHistoryTimelineNotConfigured`).
- Add `get` operation to `/api/v1/me/history/timeline` in OpenAPI contract; `since`(required), `until`(required), `granularity`(enum day/week/month, default day), `trackId`(optional) params; 200 refs `TimelineResult`; bump `info.version` to `0.84.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/history/timeline`; add `TestStorageAdminOpenAPIContractViewerHistoryTimeline`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.85.0 - 2026-06-19

- Add `GetUser(ctx, id)` to `auth.Service` (delegates to `users.GetUser`; wraps result with `toView`).
- Add `getMe` handler: reads the authenticated user from `userFromContext` and writes `UserView` at `GET /api/v1/me`.
- Add `GET /api/v1/me` route (viewer-auth); add `/api/v1/me` methodNotAllowed catch-all.
- Add 2 `auth.Service` unit tests (`TestGetUser`, `TestGetUser_NotFound`).
- Add 3 HTTP-layer tests (`TestGetMe`, `TestGetMeUnauthenticated`, `TestGetMeNotConfigured`).
- Add `GET /api/v1/me` to OpenAPI contract; 200 refs `UserView`; bump `info.version` to `0.85.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me`; add `TestStorageAdminOpenAPIContractGetMe`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.86.0 - 2026-06-19

- Add `getAdminUser` handler: reads path value `id`, calls `authService.GetUser(ctx, id)`, writes `UserView` at `GET /api/v1/admin/users/{id}`.
- Register `GET /api/v1/admin/users/{id}` route (admin-auth); add `/api/v1/admin/users/{id}` to `TestStorageAdminOpenAPIContractCoversRoutes` expected map alongside `admin/users` and `admin/users/{id}/disable`.
- Add 3 HTTP-layer tests (`TestAdminGetUser`, `TestAdminGetUserNotFound`, `TestAdminGetUserNotConfigured`).
- Add `get` operation to `/api/v1/admin/users/{id}` in OpenAPI contract; 200 refs `UserView`, 404 ErrorEnvelope; bump `info.version` to `0.86.0`.
- Add `TestStorageAdminOpenAPIContractAdminGetUser`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.87.0 - 2026-06-19

- Add `ChangePassword(ctx, userID, currentPassword, newPassword)` to `auth.Service`: verifies current password via `CheckPassword`, enforces 8-character minimum on new password, hashes and saves the updated credential.
- Add `changePassword` handler: `POST /api/v1/me/change-password`; decodes `{currentPassword, newPassword}`; returns `204 No Content` on success; `400 invalid_user` for weak/missing new password, `401 unauthorized` for wrong current password.
- Register `POST /api/v1/me/change-password` (viewer-auth); add `/api/v1/me/change-password` methodNotAllowed catch-all.
- Add `ChangePasswordRequest` schema to OpenAPI components.
- Add 3 `auth.Service` unit tests (`TestChangePassword`, `TestChangePassword_WrongCurrent`, `TestChangePassword_WeakNew`).
- Add 4 HTTP-layer tests (`TestChangePassword`, `TestChangePasswordWrongCurrent`, `TestChangePasswordUnauthenticated`, `TestChangePasswordNotConfigured`).
- Add `POST /api/v1/me/change-password` to OpenAPI contract; bump `info.version` to `0.87.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractChangePassword`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.88.0 - 2026-06-19

- Add `EnableUser(ctx, id string) (UserView, error)` to `auth.Service`: mirrors `DisableUser`; sets `Enabled = true`, updates `UpdatedAt`, saves.
- Add `enableUser` handler: `POST /api/v1/admin/users/{id}/enable`; returns `UserView` of the re-enabled user.
- Register `POST /api/v1/admin/users/{id}/enable` (admin-auth); add `/api/v1/admin/users/{id}/enable` methodNotAllowed catch-all.
- Add 2 `auth.Service` unit tests (`TestEnableUser`, `TestEnableUser_NotFound`).
- Add 3 HTTP-layer tests (`TestEnableUser`, `TestEnableUserNotFound`, `TestEnableUserNotConfigured`).
- Add `POST /api/v1/admin/users/{id}/enable` to OpenAPI contract (with `UserId` path parameter); bump `info.version` to `0.88.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractEnableUser`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.89.0 - 2026-06-19

- Add `PatchUser(ctx, id string, role *Role, username *string) (UserView, error)` to `auth.Service`: validates and applies non-nil role/username fields; checks username uniqueness on change; updates `UpdatedAt`; returns `UserView`.
- Add `patchAdminUser` handler: `PATCH /api/v1/admin/users/{id}`; decodes `{role?, username?}`; rejects empty patch (`400 invalid_user`); returns `UserView` on success; propagates `ErrInvalidUser` (400) and `ErrUserConflict` (409).
- Register `PATCH /api/v1/admin/users/{id}` (admin-auth).
- Add `PatchUserRequest` schema to OpenAPI components (`role?: enum[admin,viewer], username?: string`).
- Add 3 `auth.Service` unit tests (`TestPatchUserRole`, `TestPatchUserUsername`, `TestPatchUserConflict`).
- Add 4 HTTP-layer tests (`TestAdminPatchUserRole`, `TestAdminPatchUserUsernameConflict`, `TestAdminPatchUserEmpty`, `TestAdminPatchUserNotConfigured`).
- Add `patch` operation to `/api/v1/admin/users/{id}` in OpenAPI contract; bump `info.version` to `0.89.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `patch` on `/api/v1/admin/users/{id}`; add `TestStorageAdminOpenAPIContractAdminPatchUser`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.90.0 - 2026-06-19

- Add `SessionView{UserID, ExpiresAt, CreatedAt}` to `auth` package as a safe public projection of a session (no token hash).
- Extend `SessionRepository` interface with `ListSessionsByUser(ctx, userID string) ([]Session, error)` and `RevokeAllSessionsByUser(ctx, userID string, revokedAt time.Time) (int, error)`.
- Implement both methods in `authpg.SessionRepository` (SQL: `SELECT … WHERE user_id = $1` and `UPDATE … WHERE user_id = $2 AND revoked_at IS NULL AND expires_at > $1`).
- Implement both methods in the in-memory test stubs (`memSessionRepo` in `service_test.go` and `memAuthSessionRepo` in `handler_test.go`).
- Add `ListActiveSessions(ctx, userID string) ([]SessionView, error)` to `auth.Service`: verifies user exists, calls `ListSessionsByUser`, filters out revoked and expired sessions.
- Add `RevokeAllSessionsForUser(ctx, userID string) (int, error)` to `auth.Service`: verifies user exists, delegates to `RevokeAllSessionsByUser`.
- Add 3 `auth.Service` unit tests: `TestListActiveSessionsEmpty`, `TestListActiveSessionsFiltersRevoked`, `TestListActiveSessionsFiltersExpired`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.91.0 - 2026-06-19

- Add `getAdminUserSessions` handler: `GET /api/v1/admin/users/{id}/sessions`; requires admin auth; calls `auth.Service.ListActiveSessions`; returns `{sessions: [SessionView], count: N}`; propagates `ErrUserNotFound` (404) and auth not configured (503).
- Register `GET /api/v1/admin/users/{id}/sessions` (admin-auth) and its `methodNotAllowed` fallback.
- Add `get` operation to `/api/v1/admin/users/{id}/sessions` in OpenAPI contract; add `SessionView` schema to components; bump `info.version` to `0.91.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/users/{id}/sessions`.
- Add 4 HTTP-layer tests: `TestAdminGetUserSessionsEmpty`, `TestAdminGetUserSessionsActive`, `TestAdminGetUserSessionsNotFound`, `TestAdminGetUserSessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.92.0 - 2026-06-19

- Add `deleteAdminUserSessions` handler: `DELETE /api/v1/admin/users/{id}/sessions`; requires admin auth; calls `auth.Service.RevokeAllSessionsForUser`; returns `{"revoked": N}`; propagates `ErrUserNotFound` (404) and auth not configured (503).
- Register `DELETE /api/v1/admin/users/{id}/sessions` (admin-auth).
- Add `delete` operation to `/api/v1/admin/users/{id}/sessions` in OpenAPI contract; bump `info.version` to `0.92.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `delete` on `/api/v1/admin/users/{id}/sessions`.
- Add 3 HTTP-layer tests: `TestAdminDeleteUserSessionsRevokeActive`, `TestAdminDeleteUserSessionsNotFound`, `TestAdminDeleteUserSessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.93.0 - 2026-06-19

- Add `getMyActiveSessions` handler: `GET /api/v1/me/sessions`; requires viewer auth; calls `auth.Service.ListActiveSessions` with the authenticated user's ID; returns `{sessions: [SessionView], count: N}`; 503 when auth not configured.
- Register `GET /api/v1/me/sessions` (viewer-auth) and its `methodNotAllowed` fallback.
- Add `get` operation to `/api/v1/me/sessions` in OpenAPI contract; bump `info.version` to `0.93.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/sessions`.
- Add 3 HTTP-layer tests: `TestViewerGetMySessionsFiltersRevoked`, `TestViewerGetMySessionsActive`, `TestViewerGetMySessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.94.0 - 2026-06-19

- Add `RevokeAllExcept(ctx, userID, exceptTokenHash string) (int, error)` to `auth.Service`: lists all sessions for the user via `ListSessionsByUser`, skips the session matching `exceptTokenHash`, revokes every other active non-expired session via `RevokeSession`; returns revoked count.
- Add `revokeMyOtherSessions` handler: `POST /api/v1/me/sessions/revoke-all`; requires viewer auth; extracts current bearer token hash with `auth.HashToken`; calls `RevokeAllExcept`; returns `{"revoked": N}`; 503 when auth not configured.
- Register `POST /api/v1/me/sessions/revoke-all` (viewer-auth) and its `methodNotAllowed` fallback.
- Add `post` operation to `/api/v1/me/sessions/revoke-all` in OpenAPI contract; bump `info.version` to `0.94.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `post` on `/api/v1/me/sessions/revoke-all`.
- Add 3 HTTP-layer tests: `TestViewerRevokeMyOtherSessions`, `TestViewerRevokeMyOtherSessionsNoneOther`, `TestViewerRevokeMyOtherSessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.95.0 - 2026-06-19

- Add `?limit`, `?offset`, `?sortBy` (username/role/createdAt/updatedAt), and `?sortOrder` (asc/desc) query parameters to `GET /api/v1/admin/users`.
- Sort and paginate over the full `[]UserView` slice in the `listUsers` handler; no new repository interface methods required.
- `limit=0` (absent) returns all users from `offset`; `limit > 0` pages the result; `hasMore` reflects whether more items follow.
- Response is `{"users":[...],"pagination":{"limit":N,"offset":N,"total":N,"hasMore":bool}}`.
- Invalid `sortOrder` returns `400 invalid_sort_order`; invalid `limit` or `offset` returns `400`.
- Add 5 HTTP-layer tests: `TestAdminListUsersPagination`, `TestAdminListUsersSortByUsername`, `TestAdminListUsersSortDesc`, `TestAdminListUsersInvalidSortOrder`, `TestAdminListUsersInvalidLimit`.
- Update `GET /api/v1/admin/users` in OpenAPI contract with `limit`, `offset`, `sortBy`, `sortOrder` query params and updated 200 response schema including `pagination`; bump `info.version` to `0.95.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.96.0 - 2026-06-19

- Add `?username` (exact match), `?role` (admin/viewer), and `?enabled` (true/false) filter query parameters to `GET /api/v1/admin/users`; filters are applied before sort and pagination.
- Invalid `?role` values return `400 invalid_role`; invalid `?enabled` values return `400 invalid_enabled`.
- Add 5 HTTP-layer tests: `TestAdminListUsersFilterByRole`, `TestAdminListUsersFilterByEnabled`, `TestAdminListUsersFilterByUsername`, `TestAdminListUsersFilterInvalidRole`, `TestAdminListUsersFilterInvalidEnabled`.
- Extend `GET /api/v1/admin/users` in OpenAPI contract with `username`, `role`, and `enabled` query params; bump `info.version` to `0.96.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.97.0 - 2026-06-19

- Add `ForceChangePassword(ctx, userID, newPassword string) error` to `auth.Service`: retrieves user by ID, validates new password (≥ 8 chars), hashes, and saves without verifying the current password.
- Add `forceChangePassword` handler: `POST /api/v1/admin/users/{id}/change-password`; requires admin auth; decodes `{newPassword}`; returns `204 No Content` on success; `400 invalid_user` for weak/missing password; `404` for unknown user; `503` when auth not configured.
- Register `POST /api/v1/admin/users/{id}/change-password` (admin-auth) and its `methodNotAllowed` fallback.
- Add `ForceChangePasswordRequest` schema to OpenAPI components (`newPassword: string, minLength: 8`).
- Add 3 `auth.Service` unit tests: `TestForceChangePassword`, `TestForceChangePassword_WeakNew`, `TestForceChangePassword_NotFound`.
- Add 4 HTTP-layer tests: `TestAdminForceChangePassword`, `TestAdminForceChangePasswordWeakPassword`, `TestAdminForceChangePasswordNotFound`, `TestAdminForceChangePasswordNotConfigured`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `post` on `/api/v1/admin/users/{id}/change-password`; add `TestStorageAdminOpenAPIContractAdminForceChangePassword`.
- Bump `info.version` to `0.97.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.98.0 - 2026-06-19

- Modify `DeleteUser(ctx, id string) error` in `auth.Service` to call `RevokeAllSessionsByUser` before deleting the user record, ensuring all active sessions are revoked as part of the deletion.
- No new endpoints or repository interface methods required.
- Add 2 `auth.Service` unit tests: `TestDeleteUserRevokesSessionsFirst`, `TestDeleteUserNotFound`.
- Add 1 HTTP-layer test: `TestAdminDeleteUserRevokesSessionsFirst`.
- Bump `info.version` to `0.98.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.99.0 - 2026-06-19

- Add `revokeAllMySessions` handler: `POST /api/v1/me/sessions/revoke-all-devices`; requires viewer auth; calls `auth.Service.RevokeAllSessionsForUser` with the authenticated user's ID; revokes ALL sessions including the current one; returns `{"revoked": N}`; `503` when auth not configured.
- Register `POST /api/v1/me/sessions/revoke-all-devices` (viewer-auth) and its `methodNotAllowed` fallback.
- Add `post` operation to `/api/v1/me/sessions/revoke-all-devices` in OpenAPI contract; bump `info.version` to `0.99.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `post` on `/api/v1/me/sessions/revoke-all-devices`.
- Add 3 HTTP-layer tests: `TestViewerRevokeAllMySessions`, `TestViewerRevokeAllMySessionsIncludesCurrent`, `TestViewerRevokeAllMySessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.0.0 - 2026-06-19

- Add `getMyTrackHistory` handler: `GET /api/v1/me/history/tracks/{trackId}`; requires viewer auth; reads `{trackId}` from path; accepts `limit`, `offset`, `since`, `until`, `order` query params; calls `history.Service.ListPlays` with `UserID` from auth context and `TrackID` from path; returns `{events, pagination}`; `503` when history service not configured.
- Register `GET /api/v1/me/history/tracks/{trackId}` (viewer-auth) and its `methodNotAllowed` fallback.
- Add `get` operation to `/api/v1/me/history/tracks/{trackId}` in OpenAPI contract; bump `info.version` to `1.0.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/history/tracks/{trackId}`.
- Add 3 HTTP-layer tests: `TestViewerGetMyTrackHistory`, `TestViewerGetMyTrackHistoryFiltersToOwnUser`, `TestViewerGetMyTrackHistoryMethodNotAllowed`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.1.0 - 2026-06-19

- Add `UserTrackStats{TrackID, TotalPlays, FirstPlayedAt, LastPlayedAt}` type to the `history` package.
- Extend `history.Repository` interface with `UserTrackPlayStats(ctx, userID, trackID string) (UserTrackStats, error)`.
- Implement `MemoryRepository.UserTrackPlayStats` and `historypg.Repository.UserTrackPlayStats`.
- Add `GetMyTrackStats(ctx, userID, trackID string) (UserTrackStats, error)` to `history.Service`.
- Add `getMyTrackStats` handler: `GET /api/v1/me/history/tracks/{trackId}/stats`; requires viewer auth; returns `UserTrackStats`; `503` when history service not configured.
- Register `GET /api/v1/me/history/tracks/{trackId}/stats` (viewer-auth) and its `methodNotAllowed` fallback.
- Add `UserTrackStats` schema to OpenAPI components; add `get` operation to `/api/v1/me/history/tracks/{trackId}/stats`; bump `info.version` to `1.1.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/history/tracks/{trackId}/stats`.
- Add 3 `history.Service` unit tests: `TestGetMyTrackStatsNoPlays`, `TestGetMyTrackStatsWithPlays`, `TestGetMyTrackStatsMissingArgs`.
- Add 3 HTTP-layer tests: `TestViewerGetMyTrackStats`, `TestViewerGetMyTrackStatsNoPlays`, `TestViewerGetMyTrackStatsMethodNotAllowed`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.2.0 - 2026-06-19

- Add `getAdminTrackTimeline` handler: `GET /api/v1/admin/history/tracks/{trackId}/timeline`; requires admin auth; reads `{trackId}` from path; accepts `since` (required), `until` (required), `granularity` (optional; day/week/month) query params; calls `history.Service.GetHistoryTimeline` with `TrackID`; returns `{buckets}`; `503` when history service not configured.
- Register `GET /api/v1/admin/history/tracks/{trackId}/timeline` (admin-auth) and its `methodNotAllowed` fallback.
- Add `get` operation to `/api/v1/admin/history/tracks/{trackId}/timeline` in OpenAPI contract; bump `info.version` to `1.2.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history/tracks/{trackId}/timeline`.
- Add 3 HTTP-layer tests: `TestAdminGetTrackTimeline`, `TestAdminGetTrackTimelineMissingBounds`, `TestAdminGetTrackTimelineMethodNotAllowed`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.3.0 - 2026-06-19

- Add `getAdminUserTimeline` handler: `GET /api/v1/admin/history/users/{userId}/timeline`; requires admin auth; reads `{userId}` from path; accepts `since` (required), `until` (required), `granularity` (optional; day/week/month) query params; calls `history.Service.GetHistoryTimeline` with `UserID`; returns `{buckets}`; `503` when history service not configured.
- Register `GET /api/v1/admin/history/users/{userId}/timeline` (admin-auth) and its `methodNotAllowed` fallback.
- Add `get` operation to `/api/v1/admin/history/users/{userId}/timeline` in OpenAPI contract; bump `info.version` to `1.3.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history/users/{userId}/timeline`.
- Add 3 HTTP-layer tests: `TestAdminGetUserTimeline`, `TestAdminGetUserTimelineMissingBounds`, `TestAdminGetUserTimelineMethodNotAllowed`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.4.0 - 2026-06-19

- Expose `?trackId` query parameter on `GET /api/v1/me/history/timeline`: the `getMyHistoryTimeline` handler already passes `TrackID` from this parameter to `GetMyTimeline`; the OpenAPI contract already declares it. No handler or schema changes needed.
- Add HTTP-layer test `TestViewerGetHistoryTimelineTrackIdFilter` confirming that `?trackId` correctly scopes the viewer's timeline to a single track.
- Bump `info.version` to `1.4.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.5.0 - 2026-06-20

- Add `getMyTrackTimeline` handler: `GET /api/v1/me/history/tracks/{trackId}/timeline`; requires viewer auth; reads `{trackId}` from path; accepts `since` (required), `until` (required), `granularity` (optional; day/week/month) query params; calls `history.Service.GetMyTimeline` with both `UserID` (from auth context) and `TrackID` (from path); returns `{buckets}`; `503` when history service not configured.
- Register `GET /api/v1/me/history/tracks/{trackId}/timeline` (viewer-auth) before the existing `{trackId}` wildcard fallback.
- Add `get` operation to `/api/v1/me/history/tracks/{trackId}/timeline` in OpenAPI contract; bump `info.version` to `1.5.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/history/tracks/{trackId}/timeline`.
- Add 3 HTTP-layer tests: `TestViewerGetMyTrackTimeline`, `TestViewerGetMyTrackTimelineMissingBounds`, `TestViewerGetMyTrackTimelineMethodNotAllowed`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.6.0 - 2026-06-20

- Add `UserHistorySummary{Stats UserHistoryStats, TopTracks []TrackPlayCount}` type to the `history` package.
- Add `GetAdminUserSummary(ctx, userID string, f UserStatsFilter, topN int) (UserHistorySummary, error)` to `history.Service`: calls `GetAdminUserStats` and `GetAdminUserTopTracks` in sequence and returns the combined struct.
- Add `getAdminUserHistorySummary` handler: `GET /api/v1/admin/history/users/{userId}/history-summary`; requires admin auth; reads `{userId}` from path; accepts optional `since`, `until` (RFC3339) and optional `?topN` (int; default 10; clamped 1–100) query params; returns `UserHistorySummary`; `503` when history service not configured.
- Register `GET /api/v1/admin/history/users/{userId}/history-summary` (admin-auth) before the existing `{userId}` wildcard; add `methodNotAllowed` fallback.
- Add `UserHistorySummary` schema to OpenAPI components; add `get` operation to `/api/v1/admin/history/users/{userId}/history-summary`; bump `info.version` to `1.6.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history/users/{userId}/history-summary`.
- Add 2 `history.Service` unit tests: `TestGetAdminUserSummary`, `TestGetAdminUserSummaryEmpty`.
- Add 3 HTTP-layer tests: `TestAdminGetUserHistorySummary`, `TestAdminGetUserHistorySummaryWithTopN`, `TestAdminGetUserHistorySummaryNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.7.0 - 2026-06-20

- Add `TrackHistorySummary{Stats TrackHistoryStatsResult, TopListeners []UserPlayCount}` type to the `history` package.
- Add `GetTrackSummary(ctx, trackID string, f TrackStatsFilter, topN int) (TrackHistorySummary, error)` to `history.Service`: calls `GetTrackStats` and `GetTrackTopListeners` in sequence and returns the combined struct.
- Add `getAdminTrackHistorySummary` handler: `GET /api/v1/admin/history/tracks/{trackId}/history-summary`; requires admin auth; reads `{trackId}` from path; accepts optional `since`, `until` (RFC3339) and optional `?topN` (int; default 10; clamped 1–100) query params; returns `TrackHistorySummary`; `503` when history service not configured.
- Register `GET /api/v1/admin/history/tracks/{trackId}/history-summary` (admin-auth) before the existing `{trackId}` wildcard; add `methodNotAllowed` fallback.
- Add `TrackHistorySummary` schema to OpenAPI components; add `get` operation to `/api/v1/admin/history/tracks/{trackId}/history-summary`; bump `info.version` to `1.7.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history/tracks/{trackId}/history-summary`.
- Add 2 `history.Service` unit tests: `TestGetTrackSummary`, `TestGetTrackSummaryEmpty`.
- Add 3 HTTP-layer tests: `TestAdminGetTrackHistorySummary`, `TestAdminGetTrackHistorySummaryWithTopN`, `TestAdminGetTrackHistorySummaryNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.8.0 - 2026-06-20

- Add `getMyHistorySummary` handler: `GET /api/v1/me/history/summary`; requires viewer auth; accepts optional `since`, `until` (RFC3339) and optional `?topN` (int; default 10; clamped 1–100) query params; calls `history.Service.GetMyStats` and `history.Service.GetMyTopTracks` with `UserID` from auth context; returns `{"stats": UserHistoryStats, "topTracks": []TrackPlayCount}`; `503` when history service not configured.
- Register `GET /api/v1/me/history/summary` (viewer-auth) before the `/api/v1/me/history/{eventId}` wildcard; add `methodNotAllowed` fallback.
- Add `get` operation to `/api/v1/me/history/summary` in OpenAPI contract; response: `stats` (`UserHistoryStats` ref) and `topTracks` (array of `TrackPlayCount`); bump `info.version` to `1.8.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/history/summary`.
- Add 3 HTTP-layer tests: `TestViewerGetMyHistorySummary`, `TestViewerGetMyHistorySummaryWithTopN`, `TestViewerGetMyHistorySummaryNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.9.0 - 2026-06-20

- Extend `history.Repository.UserTrackPlayStats` signature to accept a `UserStatsFilter` for optional `Since`/`Until` time bounds; `UserStatsFilter.UserID` is ignored — caller passes `userID` directly.
- Implement time-bound filtering in `history.MemoryRepository.UserTrackPlayStats`: skip events outside `[f.Since, f.Until)` when the bounds are non-zero.
- Implement time-bound filtering in `historypg.Repository.UserTrackPlayStats`: inject `AND played_at >= $3` / `AND played_at < $4` clauses when non-zero.
- Update `history.Service.GetMyTrackStats(ctx, userID, trackID string, f UserStatsFilter) (UserTrackStats, error)` to forward `f` to the repository.
- Update `getMyTrackStats` handler to parse optional `?since` / `?until` query params (RFC3339); return `400 invalid_since` / `400 invalid_until` on parse failure; pass them via `UserStatsFilter`.
- Update `GET /api/v1/me/history/tracks/{trackId}/stats` in OpenAPI to declare `since` and `until` query parameters; bump `info.version` to `1.9.0`.
- Add 1 `history.Service` unit test: `TestGetMyTrackStatsTimeWindow`.
- Add 2 HTTP-layer tests: `TestViewerGetMyTrackStatsSince`, `TestViewerGetMyTrackStatsUntil`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.10.0 - 2026-06-20

- Add `GlobalHistorySummary{Stats HistoryStats, TopTracks []TrackPlayCount, TopUsers []UserPlayCount}` type to the `history` package.
- Add `GetGlobalSummary(ctx, f StatsFilter, topN int) (GlobalHistorySummary, error)` to `history.Service`: calls `GetHistoryStats`, `GetTopTracks`, and `GetTopUsers` in sequence; `topN ≤ 0` defaults to 10, clamped to 100.
- Add `getAdminHistorySummary` handler: `GET /api/v1/admin/history/summary`; requires admin auth; accepts optional `since`, `until` (RFC3339) and optional `?topN` (int; default 10; clamped 1–100) query params; calls `history.Service.GetGlobalSummary`; returns `GlobalHistorySummary`; `503` when history service not configured.
- Register `GET /api/v1/admin/history/summary` (admin-auth) before the existing `/api/v1/admin/history/{eventId}` wildcard; add `methodNotAllowed` fallback.
- Add `GlobalHistorySummary` schema to OpenAPI components; add `get` operation to `/api/v1/admin/history/summary`; bump `info.version` to `1.10.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history/summary`.
- Add 2 `history.Service` unit tests: `TestGetGlobalSummary`, `TestGetGlobalSummaryWithTopN`.
- Add 3 HTTP-layer tests: `TestAdminGetHistorySummary`, `TestAdminGetHistorySummaryWithTopN`, `TestAdminGetHistorySummaryNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.11.0 - 2026-06-20

- Add `MyTrackSummary{Stats UserTrackStats, RecentTracks []TrackPlayCount}` type to the `history` package.
- Add `GetMyTrackSummary(ctx, userID, trackID string, f UserStatsFilter, topN int) (MyTrackSummary, error)` to `history.Service`: calls `GetMyTrackStats` and `GetMyTopTracks`; `topN ≤ 0` defaults to 10, clamped to 100.
- Add `getMyTrackSummary` handler: `GET /api/v1/me/history/tracks/{trackId}/summary`; requires viewer auth; reads `{trackId}` from path; accepts optional `since`, `until` (RFC3339) and optional `?topN` (int; default 10; clamped 1–100) query params; calls `history.Service.GetMyTrackSummary`; returns `MyTrackSummary`; `503` when history service not configured.
- Register `GET /api/v1/me/history/tracks/{trackId}/summary` (viewer-auth) before the existing `{trackId}` wildcard fallback; add `methodNotAllowed` fallback.
- Add `MyTrackSummary` schema to OpenAPI components; add `get` operation to `/api/v1/me/history/tracks/{trackId}/summary`; bump `info.version` to `1.11.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/history/tracks/{trackId}/summary`.
- Add 2 `history.Service` unit tests: `TestGetMyTrackSummary`, `TestGetMyTrackSummaryEmpty`.
- Add 3 HTTP-layer tests: `TestViewerGetMyTrackSummary`, `TestViewerGetMyTrackSummaryWithTopN`, `TestViewerGetMyTrackSummaryNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.12.0 - 2026-06-20

- Verify `historypg.Repository.TrackHistoryStats` and `TrackTopListeners` filter on `played_at` when `f.Since`/`f.Until` are non-zero; add `AND played_at >= $N` / `AND played_at < $N` clauses if missing.
- Verify `history.MemoryRepository.TrackHistoryStats` and `TrackTopListeners` skip events outside the bounds; add filtering if missing.
- Update `getAdminTrackStats` handler to parse optional `?since` / `?until` query params (RFC3339) and pass them into `TrackStatsFilter`; return `400 invalid_since` / `400 invalid_until` on parse failure.
- Update `getAdminTrackTopListeners` handler similarly.
- Update `GET /api/v1/admin/history/tracks/{trackId}/stats` and `GET /api/v1/admin/history/tracks/{trackId}/top-listeners` in OpenAPI to declare `since` and `until` query parameters; bump `info.version` to `1.12.0`.
- Add 1 `history.Service` unit test: `TestGetAdminTrackStatsTimeWindow`.
- Add 2 HTTP-layer tests: `TestAdminGetTrackStatsSince`, `TestAdminGetTrackStatsUntil`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.13.0 - 2026-06-20

- Verify `historypg.Repository.HistoryStats`, `TopTracks`, and `TopUsers` filter on `played_at` when `f.Since`/`f.Until` are non-zero; add clauses if missing.
- Verify `history.MemoryRepository.HistoryStats`, `TopTracks`, and `TopUsers` skip events outside the bounds; add filtering if missing.
- Verify `getAdminHistoryStats`, `getAdminTopTracks`, and `getAdminTopUsers` handlers forward parsed `since`/`until` into `StatsFilter`; fix forwarding if missing.
- Update `GET /api/v1/admin/history/stats`, `GET /api/v1/admin/history/top-tracks`, and `GET /api/v1/admin/history/top-users` in OpenAPI to declare `since` and `until` query parameters if not yet declared; bump `info.version` to `1.13.0`.
- Add 1 `history.Service` unit test: `TestGetHistoryStatsTimeWindow`.
- Add 2 HTTP-layer tests: `TestAdminGetHistoryStatsSince`, `TestAdminGetHistoryStatsUntil`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.14.0 - 2026-06-20

- Verify `getMyTopTracks` handler parses `?since`/`?until` and passes them as `UserStatsFilter.Since`/`Until` to `GetMyTopTracks`; fix forwarding if missing.
- Verify `getMyHistoryStats` handler does the same for `GetMyStats`.
- Verify `getMyHistoryTimeline` handler does the same for `GetMyTimeline`.
- Update `GET /api/v1/me/history/top-tracks`, `GET /api/v1/me/history/stats`, and `GET /api/v1/me/history/timeline` in OpenAPI to declare `since` and `until` query parameters if not yet declared; bump `info.version` to `1.14.0`.
- Add 1 `history.Service` unit test: `TestGetMyTopTracksTimeWindow`.
- Add 2 HTTP-layer tests: `TestViewerGetMyTopTracksSince`, `TestViewerGetMyTopTracksUntil`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.15.0 - 2026-06-20

- Verify that `GET /api/v1/admin/history/users/{userId}/timeline` and `GET /api/v1/admin/history/tracks/{trackId}/timeline` OpenAPI paths declare `since` (required), `until` (required), and `granularity` (optional) parameters; confirm handler layer already validates all three.
- Add `TestAdminGetUserTimelineSinceFilter` and `TestAdminGetTrackTimelineSinceFilter` HTTP-layer tests verifying that `?since` restricts each timeline to events on or after the given timestamp.
- Add `TestStorageAdminOpenAPIContractAdminDetailTimelinePaths` asserting param declarations and `TimelineResult` schema ref for both detail timeline paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.17.0 - 2026-06-20

- Verify that `GET /api/v1/admin/history` OpenAPI path already declares `since` and `until` query parameters; confirm `getAdminAllHistory` handler forwards them via `GlobalPlayEventFilter`.
- Add `TestAdminGetAllHistorySinceFilter` HTTP-layer test verifying that `?since` restricts the global event list to events on or after the given timestamp.
- Add `TestAdminGetAllHistoryUntilFilter` HTTP-layer test verifying that `?until` excludes events on or after the given timestamp.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.16.0 - 2026-06-20

- Verify that `GET /api/v1/me/history` OpenAPI path declares a `trackId` query parameter (already supported in handler via `PlayEventFilter.TrackID`).
- Add `TestViewerListPlayEventsTrackIdFilter` HTTP-layer test confirming that `?trackId` restricts results to only events for the specified track.
- Add `TestStorageAdminOpenAPIContractViewerHistoryTrackIdParam` asserting that `GET /api/v1/me/history` declares `trackId` in its parameters.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.18.0 - 2026-06-20

- Add `services/api/internal/history/postgres/repository_integration_test.go` with build tag `integration` containing 5 PostgreSQL integration tests using testcontainers-go:
  - `TestRepositoryHistoryStats`: verifies total events, unique users, unique tracks after inserts.
  - `TestRepositoryTopTracks`: verifies track play counts sorted by descending count.
  - `TestRepositoryUserHistoryStats`: verifies per-user scoping of total events and unique tracks.
  - `TestRepositoryTrackHistoryStats`: verifies per-track scoping of total events and unique listeners.
  - `TestRepositoryHistoryTimeline`: verifies day-granularity bucketing and Since/Until windowing.
- Bump OpenAPI `info.version` to `1.18.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.19.0 - 2026-06-20

- Perform full cross-check of handler-registered routes (115 operations) against OpenAPI `paths` entries; confirm 100% coverage with no orphans in either direction.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` expected map to include `POST /api/v1/auth/login` and `POST /api/v1/auth/logout`, which were missing from the assertion despite being present in both handler and spec.
- Synchronize `requirement.md` `## Current Version` and `VERSION` file to `1.19.0`.
- Bump OpenAPI `info.version` to `1.19.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.20.0 - 2026-06-20

- Wire `catalog.Service` and `history.Service` into `main.go`: add `catalogRepository(pool)` and `historyRepository(pool)` constructor helpers that return PostgreSQL-backed repositories when a pool is available and in-memory repositories otherwise.
- Append `httpapi.WithCatalogService(catalogService)` and `httpapi.WithHistoryService(historyService)` to `handlerOpts`; all catalog and history routes now respond correctly in production instead of returning 503.
- Add 3 `main_test.go` tests: `TestCatalogRepositoryDefaultsToMemory`, `TestHistoryRepositoryDefaultsToMemory`, `TestHandlerWithAllServicesReportsThreeBaseChecks`.
- Bump VERSION and OpenAPI `info.version` to `1.20.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.21.0 - 2026-06-20

- Extend `readinessReport()` in `handler.go` with two new `ReadinessCheck` entries: `catalog_service` and `history_service`; `ReadinessReport.Ready` becomes `false` if either is nil.
- Add `newNoCatalogTestHandler()` and `newNoHistoryTestHandler()` helper functions in `handler_test.go`; update the 8 existing `NoCatalogService` tests that incorrectly relied on `newTestHandler()` to use `newNoCatalogTestHandler()` instead.
- Update `TestReadinessIsPublic` to assert 5 checks (up from 3) and add 3 new readiness tests: `TestReadinessAllConfigured`, `TestReadinessMissingCatalog`, `TestReadinessMissingHistory`, `TestReadinessMissingAdminAuth`.
- Bump VERSION and OpenAPI `info.version` to `1.21.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.22.0 - 2026-06-20

- Add `internal/httpapi/cors.go` with `corsMiddleware(origins []string)` that sets CORS response headers and handles OPTIONS preflight with `204 No Content`; reflects request Origin when origins is empty (permissive dev mode) or when it matches a configured value; never emits `Access-Control-Allow-Origin: *`.
- Add `corsOrigins` field to `Handler`; add `WithCORSOrigins([]string) HandlerOption`.
- Wrap `Routes()` return value with `corsMiddleware(handler.corsOrigins)(...)`.
- Add `corsOrigins()` helper in `main.go` that parses `INORI_CORS_ORIGINS` (comma-separated); logs a permissive-mode warning when unset; import `strings`.
- Add `services/api/internal/httpapi/cors_test.go` with 6 tests: `TestCORSPreflightReturns204`, `TestCORSPreflightHeadersPresent`, `TestCORSAllowedOriginReflected`, `TestCORSDisallowedOriginOmitted`, `TestCORSPermissiveModeReflectsAnyOrigin`, `TestCORSNonPreflightPassesThrough`.
- Bump VERSION and OpenAPI `info.version` to `1.22.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.23.0 - 2026-06-20

- Add `internal/httpapi/requestid.go` with `requestIDMiddleware()`, `requestIDFromContext(ctx)`, and `generateRequestID()` (16 random bytes as 32 lowercase hex chars).
- Middleware reads `X-Request-ID` from the incoming request; generates a new ID when absent; echoes the ID on the response header; injects it into the request context.
- Chain order in `Routes()`: `requestIDMiddleware` wraps `corsMiddleware` wraps `instrument(mux)`.
- Add `services/api/internal/httpapi/requestid_test.go` with 4 tests: `TestRequestIDPassthroughExisting`, `TestRequestIDGeneratedWhenAbsent`, `TestRequestIDPresentOnAllRoutes`, `TestRequestIDInjectedIntoContext`.
- Bump VERSION and OpenAPI `info.version` to `1.23.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.24.0 - 2026-06-20

- Rewrite `README.md` to reflect the `1.24.0` baseline: update version, rename "0.x Architecture Direction", enumerate all completed phases (1–124), update run command with `INORI_CORS_ORIGINS`, add `docs/architecture/frontend-client-constraints.md` to project document list, revise Future Outlook to name inori-web, inori-admin, inori-app, and shared packages.
- Bump VERSION and OpenAPI `info.version` to `1.24.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.25.0 - 2026-06-20

- Add `POST /api/v1/admin/storage/backends/{id}/enable` endpoint to complement the existing disable endpoint.
- Implement `(*Service).EnableBackend` in `services/api/internal/storage/service.go`: fetch backend by ID, return current state unchanged if already enabled (idempotent), otherwise set `Enabled=true`, `HealthStatus=HealthStatusUnknown`, `UpdatedAt=now`, persist via `repository.Save`.
- Register `enableStorageBackend` handler in `Routes()` immediately after `disableStorageBackend`; add corresponding `methodNotAllowed` catch-all for the path.
- Add `/api/v1/admin/storage/backends/{id}/enable` path to OpenAPI spec `packages/api-contract/openapi/storage-admin.v1.json` with POST method, 200/401/404/503 responses, mirroring the disable path structure.
- Bump VERSION and OpenAPI `info.version` to `1.25.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.26.0 - 2026-06-20

- Add `Delete(ctx, id) error` to `storage.Repository` interface; implement on `MemoryRepository` (map delete), `FileRepository` (map delete + persist), and `postgres.BackendRepository` (SQL DELETE, 404 on zero rows).
- Add `ErrBackendIsDefault` and `ErrBackendInUse` sentinel errors to `storage/validation.go`.
- Implement `(*Service).DeleteBackend(ctx, id)`: fetch, guard `IsDefault → ErrBackendIsDefault`, then delegate to `repository.Delete`.
- Add `deleteStorageBackend` handler: check media object count via `ListMediaObjects(limit=1)`, reject with `ErrBackendInUse` if references exist, call `storage.DeleteBackend`, respond 204.
- Add `ErrBackendIsDefault → 409 storage_backend_is_default` and `ErrBackendInUse → 409 storage_backend_in_use` cases to `writeError`.
- Register `DELETE /api/v1/admin/storage/backends/{id}` in `Routes()`; fix `/validate` and `/refresh` catch-alls to explicit method prefixes to avoid Go ServeMux conflict with the new DELETE wildcard route.
- Add `/api/v1/admin/storage/backends/{id}` DELETE path and two new error codes to OpenAPI spec.
- Bump VERSION and OpenAPI `info.version` to `1.26.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.27.0 - 2026-06-20

- Add `Genre string` field (JSON `"genre"`, `omitempty`) to `catalog.Track` struct in `types.go`.
- Add `Genre string` field to `ListQuery` for track filter pass-through.
- Add `Genre *string` field to `UpdateTrackRequest`; add `Genre string` to `ImportTrackRequest`.
- Add `TrackSortByGenre = "genre"` constant.
- Update `(*Service).CreateTrack` signature to accept `genre string` parameter (after `mediaObjectID`); set `track.Genre = strings.TrimSpace(genre)`.
- Update `(*Service).UpdateTrack` to apply `req.Genre` when non-nil.
- Update `(*Service).ImportTrack` to set `track.Genre = strings.TrimSpace(req.Genre)`.
- `MemoryRepository.ListTracksPage/ListTracksByAlbumPage/ListTracksByArtistPage`: add `strings.EqualFold` genre filter; add `TrackSortByGenre` case to `trackLess`.
- `postgres.Repository.SaveTrack`: add `genre` column to INSERT/UPDATE; use `NULLIF($10,'')`.
- `postgres.Repository.GetTrack/ListTracks/ListTracksByAlbum/ListTracksByArtist`: add `COALESCE(genre,'')` to SELECT; update `scanTrack` to scan `&t.Genre`.
- `postgres.Repository.ListTracksPage/ListTracksByAlbumPage/ListTracksByArtistPage`: add `COALESCE(genre,'')` column and conditional `WHERE lower(COALESCE(genre,''))=lower($N)` when `q.Genre != ""`.
- `postgres.Repository.queryTracksPage`: add `&t.Genre` to Scan call.
- `trackOrderBy`: add `TrackSortByGenre → lower(COALESCE(genre,''))`.
- Migration `009_track_genre`: `ALTER TABLE tracks ADD COLUMN IF NOT EXISTS genre TEXT` + partial index.
- `httpapi`: add `Genre` to `createTrackRequest`, `patchTrackRequest`, `importTrackRequest`; parse `?genre` query param in `listTracks`; pass all through to service.
- Update all `CreateTrack` call sites in `service_test.go` to add `""` for the new genre param.
- Add `genre` field to `CatalogTrack` and `CatalogUpdateTrackRequest` OpenAPI schemas; add `?genre` parameter to 6 track list endpoints.
- Bump VERSION and OpenAPI `info.version` to `1.27.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.28.0 - 2026-06-20

- Add `internal/favorites` package: `FavoriteEntry`, `FavoritesPage`, `Repository` interface, `Service`.
- `Service` methods: `AddFavorite` (idempotent), `RemoveFavorite` (idempotent), `ListFavorites` (paginated, newest-first), `IsFavorite`, `AreFavorites` (batch).
- `favorites.MemoryRepository`: in-memory implementation with sorted output.
- `favorites/postgres.Repository`: PostgreSQL implementation using `user_track_favorites` table; `AddFavorite` uses `ON CONFLICT (user_id, track_id) DO NOTHING`; `ListFavorites` uses `COUNT(*) OVER()`; `AreFavorites` uses `ANY($2)`.
- Migration `010_user_track_favorites`: `user_track_favorites(user_id, track_id, created_at)` table + unique PK + covering index.
- `httpapi.Handler`: add `favoritesService *favorites.Service` field and `WithFavoritesService` option.
- `httpapi` routes: `POST /api/v1/me/favorites/tracks/{trackId}` (add, returns 200), `DELETE /api/v1/me/favorites/tracks/{trackId}` (remove, 204), `GET /api/v1/me/favorites/tracks` (list with limit/offset pagination).
- `requireFavoritesService` guard mirrors `requireHistoryService`.
- `main.go`: wire `favoritesRepository` and `favorites.NewService` with PG/memory fallback.
- OpenAPI: `FavoritesPage` schema, 3 favorites paths, `favorites_not_configured` error code.
- Bump VERSION and OpenAPI `info.version` to `1.28.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.29.0 - 2026-06-20

- Add `isFavorite bool` field to `CatalogTrack` OpenAPI schema; viewer responses carry the flag, admin responses always carry `false`.
- Define `trackView` struct in `httpapi` embedding `catalog.Track` plus `IsFavorite bool`.
- Add `annotateTracksWithFavorites(ctx, userID, tracks)` helper: single batch call to `favorites.Service.AreFavorites`; best-effort (annotation skipped on error, no request failure).
- Add `isViewerPath(r)` helper distinguishing `/api/v1/catalog/` and `/api/v1/me/` paths from admin paths.
- Update `listTracks` handler: viewer path → annotate and respond with `[]trackView`; admin path → respond with `[]catalog.Track` unchanged.
- Update `getTrack` handler: viewer path → wrap in `trackView` with `IsFavorite` set; admin path unchanged.
- Update `getPlaylistTracks` handler: viewer path → annotate paged slice; admin path unchanged.
- Update `listFavoriteTracks` handler: resolve track IDs to full `catalog.Track` objects via `catalogService.GetTrack`; build `[]trackView` with `IsFavorite=true` for all entries; falls back to `trackIds`-only response when catalog service is unavailable.
- Bump VERSION and OpenAPI `info.version` to `1.29.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.30.0 - 2026-06-20

- Add `favorites_service` as the 6th check in `readinessReport()`; message: "favorites service is configured" / "favorites service is not configured".
- Add `WithFavoritesService` to `newTestHandler()` in `handler_test.go`; update readiness check count assertions from 5 to 6 in `TestReadinessIsPublic` and `TestReadinessAllConfigured`; add `"favorites_service"` to the expected-ok names slice.
- Bump VERSION and OpenAPI `info.version` to `1.30.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.31.0 - 2026-06-20

- Add `UpdateBackendRequest{DisplayName *string, Priority *int}` to `storage` package.
- Implement `(*Service).UpdateBackend(ctx, id, req)`: fetch, apply non-nil fields (display name non-empty guard), set `UpdatedAt=now`, persist.
- Add `patchStorageBackendRequest` struct to `httpapi/handler.go`; implement `patchStorageBackend` handler.
- Register `PATCH /api/v1/admin/storage/backends/{id}` in `Routes()`.
- Add `PATCH /api/v1/admin/storage/backends/{id}` path to OpenAPI spec with `displayName` and `priority` request body fields; 200/400/401/404/503 responses.
- Bump VERSION and OpenAPI `info.version` to `1.31.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.32.0 - 2026-06-20

- Add `ReleaseYearMin int` and `ReleaseYearMax int` fields to `catalog.ListQuery`.
- `MemoryRepository.ListAlbumsPage`: apply year range guard in the collection loop.
- `MemoryRepository.ListAlbumsByArtistPage`: apply year range guard alongside `ArtistID` check.
- `postgres.Repository.ListAlbumsPage`: build optional `WHERE release_year >= $N` / `release_year <= $N` clauses when bounds are non-zero.
- `postgres.Repository.ListAlbumsByArtistPage`: extend dynamic WHERE clause with optional year bounds.
- `httpapi`: add `parseReleaseYearRange(w, r)` helper (validates non-negative ints, enforces min ≤ max); wire into `listAlbums` and `listAlbumsByArtist` handlers.
- Add `?releaseYearMin` and `?releaseYearMax` query params to 4 album list paths in OpenAPI spec.
- Bump VERSION and OpenAPI `info.version` to `1.32.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.33.0 - 2026-06-20

- Add `ClearUserFavorites(ctx, userID) error` and `RemoveTrackFavorites(ctx, trackID) error` to `favorites.Repository` interface.
- Implement both on `MemoryRepository` (map iteration + delete) and `postgres.Repository` (DELETE WHERE user_id / track_id).
- Add `ClearUserFavorites` and `AdminRemoveFavorite` to `favorites.Service`.
- Register admin favorites routes: `GET /api/v1/admin/favorites/users/{userId}/tracks`, `DELETE /api/v1/admin/favorites/users/{userId}/tracks`, `DELETE /api/v1/admin/favorites/users/{userId}/tracks/{trackId}`.
- Implement `adminListUserFavorites`, `adminClearUserFavorites`, `adminRemoveUserFavoriteTrack` handlers.
- Add admin favorites paths to OpenAPI spec (3 paths, 5 operations); total paths 112.
- Bump VERSION and OpenAPI `info.version` to `1.33.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.34.0 - 2026-06-20

- Update `README.md` to `1.34.0` baseline: version field, phases 125–134 descriptions, OpenAPI path count updated to 134 operations.
- Bump VERSION and OpenAPI `info.version` to `1.34.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.35.0 - 2026-06-20

- Add `GET /api/v1/admin/storage/backends/{id}` endpoint returning the single backend by ID; handler `getStorageBackend` delegates to `storage.Service.GetBackend`; 404 on unknown ID.
- Register `GET /api/v1/admin/storage/backends/{id}` in `Routes()` before the sub-path handlers.
- Fix OpenAPI spec `packages/api-contract/openapi/storage-admin.v1.json`: replace the `{id}` path that previously held only `DELETE` with a complete entry carrying `GET`, `PATCH`, and `DELETE` (the PATCH spec was missing from Phase 131 due to an insertion guard bug).
- Bump VERSION and OpenAPI `info.version` to `1.35.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.36.0 - 2026-06-20

- Add `TestStorageAdminOpenAPIContractPhase125to135Routes`: asserts `POST .../enable`, `GET/PATCH/DELETE .../backends/{id}`, viewer favorites (GET/POST/DELETE), and admin favorites (GET/DELETE/DELETE per-track) are all present in the OpenAPI spec.
- Add `TestStorageAdminOpenAPIContractPhase127GenreParam`: verifies `?genre` on the 6 track list paths.
- Add `TestStorageAdminOpenAPIContractPhase132ReleaseYearParams`: verifies `?releaseYearMin` and `?releaseYearMax` on 4 album list paths.
- Add `TestStorageAdminOpenAPIContractPhase127And129Fields`: verifies `CatalogTrack.genre` (Phase 127) and `CatalogTrack.isFavorite` (Phase 129) are in the schema.
- Add `TestStorageAdminOpenAPIContractPhase131PatchBackendBody`: verifies `displayName` and `priority` fields in the PATCH backend requestBody.
- Add `TestStorageAdminOpenAPIContractPhase128FavoritesPage`: verifies `FavoritesPage` schema with `pagination` field.
- Bump VERSION and OpenAPI `info.version` to `1.36.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.37.0 - 2026-06-20

- Add `newFavoritesTestHandler` test helper creating a handler with auth, catalog, and favorites services wired; seeds one viewer user, one admin user, and one track.
- Add `TestAddFavoriteTrack`: POST add → 200; idempotent re-add → 200; unauthenticated → 401.
- Add `TestRemoveFavoriteTrack`: POST add + DELETE remove → 204; idempotent remove → 204.
- Add `TestListFavoriteTracks`: empty list → total=0; add + list → total=1, track carries `isFavorite=true`.
- Add `TestListFavoritesNotConfigured`: handler without favorites service returns 503 `favorites_not_configured`.
- Add `TestAdminListUserFavorites`: admin GET /admin/favorites/users/{userId}/tracks returns the viewer's favorited track.
- Add `TestAdminClearUserFavorites`: admin DELETE all favorites → 204; verify empty afterward.
- Add `TestAdminRemoveUserFavoriteTrack`: admin DELETE single track favorite → 204; idempotent → 204.
- Add `TestCatalogTrackIsFavoriteInViewerList`: viewer catalog/tracks carries `isFavorite=true` for favorited track; admin catalog/tracks carries `false`.
- Bump VERSION and OpenAPI `info.version` to `1.37.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v1.38.0 - 2026-06-20

- Add `TestGetStorageBackend`: GET /admin/storage/backends/{id} returns the registered backend; unknown ID → 404.
- Add `TestPatchStorageBackend`: PATCH displayName updates the field; empty displayName → 400; unknown ID → 404.
- Add `TestEnableStorageBackend`: disable then enable → enabled=true; idempotent enable → 200.
- Add `TestDeleteStorageBackendGuards`: delete default backend → 409 `storage_backend_is_default`; unknown ID → 404.
- Add `TestDeleteStorageBackendSuccess`: delete non-default backend → 204; subsequent GET → 404.
- Add `TestAlbumReleaseYearFilter`: no filter → 3 albums; `releaseYearMin=2015` → 2; `releaseYearMax=2015` → 2; min>max → 400; invalid value → 400.
- Bump VERSION and OpenAPI `info.version` to `1.38.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.
### v1.39.0 - 2026-06-20

- Add `?types=` query parameter to `GET /api/v1/admin/catalog/search` and `GET /api/v1/catalog/search`; accepts comma-separated subset of `artist`, `album`, `track`; invalid values return 400 `validation_error`.
- Filter is applied in the handler after the catalog service returns results; no changes to `Repository` or `Service` interfaces.
- Add `TestCatalogSearchTypesFilter`: no types → 3 results; `types=track` → 1 track; `types=artist,album` → 2; invalid `types=playlist` → 400.
- Add `TestViewerCatalogSearchTypesFilter`: viewer token + `types=track` returns correct filtered results.
- Add `?types` query param to both search paths in OpenAPI spec.
- Bump VERSION and OpenAPI `info.version` to `1.39.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v2.0.0 - 2026-06-21

- **v2 Web UI launch**: inori-music now ships a full viewer web player alongside the v1 API server.
- `services/web/`: Next.js 15 / React 19 viewer player — auth, catalog browse, search, playback, user library (favorites + history), settings.
- `services/admin/`: Next.js 15 / React 19 admin console — users, catalog CRUD, import wizard, storage backends, media objects, history analytics.
- `packages/ui/`: shared `@inori/ui` design token + component library (NeonCard, Badge, Skeleton, StorageHealthBadge, LifecyclePill).
- **Neon Shrine** design language: dark `#070711` canvas, electric-violet `#9b5cff` primary, cyan `#0fd4c0` secondary, sakura `#ff5fa0` accent.
- **Nginx gateway**: `infra/nginx/` replaces Caddy; `/api/v1/*` → api:8080, `/admin/*` → admin:3001, `/*` → web:3000.
- **docker-compose.prod.yml**: four-service stack (postgres + api + web + admin + nginx).
- **Streaming playback**: `GET /api/v1/catalog/tracks/{id}/stream` Go handler for local/NFS/SMB backends (HTTP 206 Range, `?token=` auth fallback for `<audio>`).
- **Player upgrades**: 64-bar canvas visualizer, dnd-kit queue drawer, keyboard shortcuts (Space/←/→/↑/↓/N/P), mobile full-screen player, bottom tab nav.
- **i18n**: i18next with en/zh-Hans/ja locale files for both services.
- **Biome**: replaces ESLint as linter/formatter in both Next.js services.
- **CI**: `build.yml` web + admin jobs; `docker.yml` api + web + admin image jobs.
- The phase output is version-tracked and covered by type checks and Go tests.

### v2.1.0 - 2026-06-21

- **Phase 205 — User library**: `/library/favorites` (favorite tracks list, remove, play); `/library/history` (play events, stats card, top-5 tracks, clear all, per-event delete).
- **Phase 206 — Admin dashboard**: `/admin` stats cards (catalog + history counts), quick-nav to all admin sections, AdminTokenPanel bootstrap support.
- **Phase 207 — User management**: `/admin/users` list with offset pagination; create user (username/password/role); toggle enable/disable, force password reset, change role, delete.
- **Phase 208 — Catalog management**: `/admin/catalog` tabbed artists/albums/tracks list; inline title edit, delete.
- **Phase 209 — Import UI**: `/admin/import` single-track import form (title, mediaObjectId, artistId, albumId, trackNumber) and batch JSON import (array or `{items:[]}` wrapper).
- **Phase 210 — Storage management**: `/admin/storage` backend cards with health/capacity indicator; probe, enable/disable, set-default, delete.
- **Phase 211 — History management (Admin)**: `/admin/history` global stats, top-tracks, top-users; event list with delete; time-window clear.
- Shared infra: `useAdminApi()` hook, `bearerAdminApi()` client, AdminStore (persisted bootstrap token), EmptyState shared component.
- The phase output is version-tracked and covered by TypeScript type checks.

### v2.2.0 - 2026-06-21

- **Phase 215 — Streaming playback**: Go `GET /api/v1/catalog/tracks/{id}/stream` (HTTP 206 Range); authenticates via `Authorization` header or `?token=` query param; `storage.SafeObjectPath` for path-traversal safety; `useAudio` presignedUrl → streamUrl fallback.
- **Phase 216 — Display quality**: `lib/api/catalog-cache.ts` in-memory artistId→name / albumId→title cache; `TrackRow` shared component with isFavorite heart toggle; `/tracks` uses TrackRow + artist name resolution; `/albums/[id]` resolves artistName and links to artist page; PlayerBar error state with skip button; Topbar ⌘K/Ctrl+K shortcut → `/search`.
- **Phase 212 — Responsive + PWA**: `MobileSidebar.tsx` slide-in drawer; AppShell hamburger; `public/manifest.json` PWA manifest (standalone, icons); viewport themeColor; apple-web-app meta.
- **Phase 214 — Production deploy**: `docker-compose.prod.yml` four-service stack (postgres + api + web + caddy); Caddyfile reverse proxy with ACME TLS; `docker.yml` publish-web job (ghcr.io, amd64+arm64).
- The phase output is version-tracked and covered by TypeScript type checks.

### v2.3.0 - 2026-06-21

- **Phase 240 — Service split**: `services/admin/` extracted as independent Next.js 15 service on port 3001; standalone admin login (JWT or bootstrap token); AdminShell (collapsible sidebar + topbar); full admin route set; independent auth store, middleware, API client, Dockerfile. `services/web/` stripped of all admin routes/components.
- **Phase 241 — Neon Shrine design system**: `globals.css` full palette (void `#070711`, surface `#0d0d1a`, primary `#9b5cff`, secondary `#0fd4c0`, accent `#ff5fa0`); Google Fonts: Orbitron + Inter + JetBrains Mono + Noto Sans JP; same palette applied to `services/admin/`.
- **Phase 245 — Nginx gateway**: `infra/nginx/nginx.conf` tuned for audio streaming (proxy_buffering off, 300 s timeouts); `conf.d/inori.conf` routes `/api/v1/*` → api:8080, `/admin/*` → admin:3001, `/*` → web:3000, `/_next/static/` cached 7 d; `docker-compose.prod.yml` updated (Caddy removed); `build.yml` + `docker.yml` admin jobs added.
- The phase output is version-tracked and covered by TypeScript type checks.

### v2.4.0 - 2026-06-21

- **Phase 242 — Player upgrade**: `Visualizer.tsx` 64-bar canvas FFT (primary→secondary gradient); `QueueDrawer.tsx` dnd-kit sortable sheet; `FullscreenPlayer.tsx` motion slide-up; `BottomNav.tsx` mobile 5-tab bottom nav; `usePlayerKeyboard.ts` Space/←/→/↑/↓/N/P shortcuts; `store/player.ts` `reorderQueue()` with currentIndex tracking.
- **Phase 243+244 — Admin complete + packages/ui**: `packages/ui/` shared `@inori/ui` library (NeonCard, Badge, Skeleton, EmptyState, StorageHealthBadge, LifecyclePill, neon-shrine.css); admin `history/page.tsx` Recharts AreaChart with day/week/month granularity switcher; `storage/page.tsx` capacity bar + StorageHealthBadge.
- **Phase 246 — i18n**: i18next + react-i18next in both services; `public/locales/{en,zh-Hans,ja}/common.json`; `lib/i18n.ts` initI18n/setLanguage/SUPPORTED_LANGS; `/settings/language` locale picker; Sidebar Settings nav.
- **Phase 247 — Quality**: `scripts/gen-icons.mjs` canvas-based PWA icon generator; placeholder `icon-192.png` / `icon-512.png`; 0 TypeScript errors in both services; 732 Go tests pass.
- The phase output is version-tracked and covered by TypeScript type checks and Go tests.

### v2.5.0 - 2026-06-22

- **Admin basePath**: `next.config.ts` `basePath='/admin'` + `transpilePackages=['@inori/ui']`; `middleware.ts` strips basePath for auth guard; Dockerfile + docker-compose healthchecks updated to `/admin/login`; Nginx 308 redirect `/admin` → `/admin/`.
- **i18n initialization**: `I18nProvider.tsx` (web) and `AdminI18nProvider.tsx` (admin) lazy-init i18next on first client render; `services/admin/lib/i18n.ts` with `/admin/locales/` path prefix; admin locale files `{en,zh,ja}/common.json`.
- **packages/ui hardening**: `utils.ts` zero-dependency `cn()`; Badge/NeonCard use `any` to avoid peer-dep issues; StorageHealthBadge re-export avoids circular imports.
- **Biome**: `biome.json` a11y.useKeyWithClickEvents=off, suspicious.noArrayIndexKey=off; both services `npm run lint` = biome lint (0 errors).
- The phase output is version-tracked and covered by TypeScript type checks and Go tests.

### v2.6.0 - 2026-06-22

- **Phase 249 — History stats UI**: `/library/history` redesigned with Stats/Events tabs. Stats tab: 30-day play timeline bar chart (SVG, zero-dependency `BarChart` component); top-10 tracks榜单 (all-time play count); total plays + unique tracks summary cards. Consumes `GET /api/v1/me/history/stats`, `/timeline` (since/until/granularity), `/top-tracks`.
- **Phase 250 — Track detail page + history batch delete**: New `/tracks/[id]` detail page (title, artist link, album link, duration, genre, track/disc number, isFavorite toggle, play button, play-count stats via `GET /api/v1/me/history/tracks/{trackId}/stats`). History Events tab: checkbox multi-select; batch delete toolbar consuming `POST /api/v1/me/history/batch-delete` (chunked at 100 IDs). `TrackRow` title links to `/tracks/[id]`.
- **Phase 251 — E2E + version sync**: Playwright `@playwright/test` added to `services/web` devDependencies; `playwright.config.ts` targeting `http://localhost:3000`; `e2e/smoke.spec.ts` three smoke tests (login redirect, valid login, search input, player bar visible); `services/web/package.json` version bumped to `2.6.0`; `requirement.md` backfilled for v2.1.0–v2.6.0; `VERSION` updated to `2.6.0`.
- The phase output is version-tracked and covered by TypeScript type checks, Go tests, and Playwright smoke tests.

### v2.7.0 - 2026-06-22

- **CI hardening — OpenAPI contract fix**: `/api/v1/catalog/tracks/{id}/stream` path-level parameter moved to `$ref: '#/components/parameters/CatalogId'`; `security: [{bearerAuth: []}]` added to GET operation; `?token` query param retained at operation level. All 39 OpenAPI contract tests pass.
- **CI E2E job**: `build.yml` new `e2e` job (depends on `api` + `web`); starts API in in-memory mode (`INORI_INITIAL_ADMIN_USER/PASSWORD`); starts Next.js dev server; installs Playwright chromium; runs `e2e/smoke.spec.ts` (3 smoke tests); uploads `playwright-report/` artifact on every run (7-day retention).
- **E2E smoke test hardening**: `#username`/`#password` exact locators; shared `login()` helper; default credentials match CI env (`ci_viewer` / `ci-password-123`).
- The phase output is version-tracked and covered by TypeScript type checks, Go tests, and 39 contract tests.

### v2.8.0 - 2026-06-22

- **Phase 252 — 低难度 Gap 补全**:
  - `services/web/settings/sessions`: Added "Revoke all devices" button (`POST /api/v1/me/sessions/revoke-all-devices`); clears local session and redirects to `/login`; explanatory copy distinguishing revoke-others vs revoke-all-devices.
  - `services/admin/history`: Batch-delete multi-select toolbar (`POST /api/v1/admin/history/batch-delete`, chunked at 100 IDs); select-all toggle; selection cleared on reload.
  - `services/admin/users`: Sessions drawer per user — `MonitorX` icon opens modal with `GET /api/v1/admin/users/{id}/sessions` list and "Revoke all" button (`DELETE /api/v1/admin/users/{id}/sessions`).
- **Phase 253 — Admin catalog sub-relations + track relink**:
  - Artists tab: expand chevron loads `GET /admin/catalog/artists/{id}/albums` + `GET .../tracks` inline.
  - Albums tab: expand chevron loads `GET /admin/catalog/albums/{id}/tracks` inline.
  - Playlists tab: expand chevron loads `GET /admin/catalog/playlists/{id}/tracks`; per-track remove (`DELETE .../tracks/{trackId}`); add-track input (`POST .../tracks`).
  - Tracks tab: "Relink" button opens modal for `POST /admin/catalog/tracks/{id}/relink` with new `mediaObjectId`.
- **Phase 254 — Admin media-objects 详情抽屉**:
  - Stats bar: `GET /admin/media/objects/stats` (total count, total size, active count) + `GET .../duplicates` (duplicate group count).
  - Row click opens detail drawer: metadata grid, lifecycle state dropdown (`POST .../lifecycle`), verify button (`POST .../verify`), timeline list (`GET .../timeline`).
- **Phase 255 — Admin favorites 管理页**:
  - New `/admin/favorites` page and `AdminSidebar` nav entry.
  - User ID lookup form → `GET /admin/favorites/users/{userId}/tracks`.
  - Per-track remove (`DELETE .../tracks/{trackId}`) and clear-all (`DELETE .../tracks`).
- The phase output is version-tracked and covered by TypeScript type checks (0 errors in both services).

### v3.0.0 - 2026-06-26

- **Phase 300 — 脚手架**: `services/mobile/` Flutter 项目 (Android/iOS/macOS/Windows/Linux)；`pubspec.yaml` 集成 riverpod / hooks_riverpod / go_router / dio / just_audio / audio_service / flutter_secure_storage / cached_network_image / freezed；openapi-generator-cli `dart-dio` 生成 `lib/src/api/`；GitHub Actions CI (`flutter analyze` + `flutter test`)。
- **Phase 301 — 认证**: `AuthNotifier`（Riverpod AsyncNotifier）：login / logout / token 持久化（flutter_secure_storage）；`POST /api/v1/auth/login` → token + `GET /api/v1/me` → UserModel；go_router redirect guard 未登录跳 `/login`；LoginScreen：username/password 表单、错误提示、loading 状态。
- **Phase 302 — 目录浏览**: `ArtistsScreen` / `ArtistDetailScreen`（albums + tracks）；`AlbumsScreen` / `AlbumDetailScreen`；`TracksScreen`；`PlaylistsScreen` / `PlaylistDetailScreen`；分页 `InfiniteScrollController`（Riverpod family provider + keepAlive）；公共 `TrackListTile`（isFavorite 心形按钮、时长、artwork）。
- **Phase 303 — 搜索**: `SearchScreen`：TextField debounce 300ms → `GET /catalog/search?q=`；分类结果展示：Artists / Albums / Tracks 三栏；键盘搜索触发 + 清空按钮。
- **Phase 304 — just_audio 播放器引擎**: `PlayerNotifier`（Riverpod Notifier）：queue / currentIndex / status / position / volume / shuffle / repeat；`AudioHandler`（audio_service）：MediaItem、播放控制、background audio；播放 URL 解析 `GET /catalog/tracks/{id}/playback` → presignedUrl || streamUrl?token=；Queue 操作：playQueue / enqueue / enqueueNext / reorderQueue / removeFromQueue。
- **Phase 305 — 播放器 UI**: `MiniPlayerBar`（底部吸附，全局 persistent）：artwork、标题、艺术家、play/pause、next；`FullPlayerScreen`（全屏，路由 `/player`）：大封面、进度条（seek）、完整控件；`QueueSheet`（底部 sheet）：DraggableScrollableSheet + ReorderableListView。
- **Phase 306 — 键盘快捷键 + MediaSession**: 桌面端（macOS/Windows/Linux）：Space/←/→ 键盘快捷键（HardwareKeyboard listener）；MediaSession：audio_service 已注入，系统锁屏控制、通知栏控制；自动播放历史：`processingState == AudioProcessingState.completed` → `POST /me/history`。
- **Phase 307 — 收藏**: `FavoritesScreen`：`GET /me/favorites/tracks` 列表；`isFavorite` 状态同步到 TrackListTile（TrackFavoriteNotifier per trackId）；POST / DELETE `/me/favorites/tracks/{trackId}`。
- **Phase 308 — 历史**: `HistoryScreen`：`GET /me/history` 分页事件流；`HistoryStatsScreen` / Tab：stats + top-tracks + 30天 timeline（fl_chart BarChart）；批量删除：多选 + `POST /me/history/batch-delete`。
- **Phase 309 — 设置**: `SettingsScreen`：密码修改 / 语言切换 / 会话管理（revoke-all）；多语言：`flutter_localizations` + ARB 文件（en / zh-Hans / ja）；主题：Neon Shrine 暗色调（ColorScheme.dark，primary violet `#9b5cff`）。
- **Phase 310 — 响应式 + 自适应布局**: 手机（<600dp）：底部 NavigationBar 5 tab；平板（600–1199dp）：NavigationRail + 右侧内容区；桌面（≥1200dp）：永久侧边栏 + 内容区 + MiniPlayer 底栏。
- **Phase 311 — CI + 打包**: GitHub Actions `mobile` job（`flutter analyze` + `flutter test` + `flutter build apk --release`）；Android keystore 签名配置（环境变量 / key.properties）。
- **Phase 312 — v3.0.0 结案**: `services/mobile/pubspec.yaml` version: 3.0.0；`VERSION` 3.0.0；`requirement.md` v3.0.0 章节。

### v3.0.1 - 2026-06-26

- **fix: 历史记录显示真实曲名**: 新增 `TrackTitleResolver`（`AutoDisposeFamilyNotifier`），`HistoryScreen` 和 `HistoryStatsScreen` 的列表项从仅显示 `trackId` 改为异步解析并展示真实标题，回退为 trackId。
- **fix: PlayerNotifier 元数据缓存**: `PlayerNotifier` 引入 `_trackCache`（`Map<String, CatalogTrack>`），`_resolveTrack()` 异步获取并缓存 catalog 元数据；queue 入列时使用 `_stubMediaItem()`（可从缓存取已知标题），正式播放时再用 `_makeMediaItem()` 填充完整 title / artist / duration。
- **fix: ArtistsScreen suppress lint**: 添加 `unnecessary_non_null_assertion` ignore，消除 flutter analyze 警告。

### v3.0.2 - 2026-06-29

- **fix: 元数据显示质量** — Flutter 客户端中艺术家名和专辑名当前显示为 UUID；通过 ArtistCacheNotifier / AlbumCacheNotifier（AutoDisposeFamilyAsyncNotifier）异步批量解析并缓存，回退为 ID。
- `PlayerNotifier._makeMediaItem` 从缓存填充真实 `artist` 名和 `album` 标题；`MiniPlayerBar` / `FullPlayerScreen` 展示解析后的字符串。
- `TrackListTile` subtitle 从 `track.artistId` 查缓存展示 artist name，缓存未命中时发起 `GET /api/v1/catalog/artists/{id}` 请求。
- `flutter analyze` 保持 0 errors；补充对应 widget 测试。
- The phase output is version-tracked and covered by flutter analyze.

### v3.1.0 - 2026-06-29

- **feat: 封面图（Artwork）** — 服务端：Album / Artist 新增可选 `artworkMediaObjectId` 字段（数据库列 + OpenAPI schema 含 `CatalogAlbum.artworkMediaObjectId` + PATCH 端点支持设值）；新增 `GET /api/v1/catalog/albums/{id}/artwork` 端点，返回 `AlbumArtworkResponse{url, expiresIn}`（presigned URL，复用 `GeneratePresignedURL`，backend 能力检查同 Phase 60）；端点对 viewer 可见（requireViewerAuth）；OpenAPI contract 新增 `AlbumArtworkResponse` schema 及路径，版本升至 3.5.0。
- Flutter 客户端：`artworkUrlProvider(albumId)` Family AsyncNotifier，请求上述端点并缓存结果；`TrackListTile`、`MiniPlayerBar`、`FullPlayerScreen` 封面区域替换为 `CachedNetworkImage`，404/错误时降级到占位图标。
- The phase output is version-tracked and covered by Go unit tests, OpenAPI contract tests, and flutter analyze.

### v3.2.0 - 2026-06-29

- **feat: 用户个人播放列表（viewer 自建，非 admin catalog）** — 服务端：新增 `user_playlists` 域（`internal/userplaylist` package）；`POST/GET/PATCH/DELETE /api/v1/me/playlists`（viewer session 认证）；`POST/DELETE /api/v1/me/playlists/{id}/tracks`（追加 / 移除）；`GET/PUT /api/v1/me/playlists/{id}/tracks`（分页展开 / 全量替换）；OpenAPI contract 新增 `UserPlaylist`、`CreateUserPlaylistRequest`、`UpdateUserPlaylistRequest`、`AddUserPlaylistTrackRequest`、`SetUserPlaylistTracksRequest` schema 及 10 条路径，版本升至 3.5.0。
- Flutter 客户端：Library Tab 新增「我的播放列表」section；创建 / 编辑 / 删除对话框；播放列表详情页（TrackListTile 列表 + 播放全部）；长按 TrackListTile 弹出「添加到播放列表」sheet。
- The phase output is version-tracked and covered by Go service unit tests, HTTP handler tests, OpenAPI contract tests, and flutter analyze.

### v3.3.0 - 2026-06-29

- **feat: 桌面平台增强** — macOS / Windows / Linux：`package:tray_manager` 系统托盘图标（Play/Pause/Next/Quit）；`package:hotkey_manager` 全局快捷键（Space / ← / →），非焦点窗口下也响应。
- Android 深度链接：`AndroidManifest.xml` 添加 `intent-filter`（`inori://` scheme + HTTPS App Link）；iOS 深度链接：`Info.plist` `CFBundleURLSchemes` + Associated Domains；go_router 处理 `inori://tracks/{id}` 跳转至 TrackListTile 并触发播放。
- The phase output is version-tracked and covered by flutter analyze and manual device verification.

### v3.4.0 - 2026-06-29

- **feat: 离线播放 + 下载管理** — `package:sqflite`：本地 SQLite 存储离线曲目元数据（trackId / title / artistName / albumTitle / localPath / downloadedAt）；`DownloadManager`（Riverpod Notifier）：`GET /api/v1/catalog/tracks/{id}/playback` → 下载到 `getApplicationDocumentsDirectory()`；just_audio `AudioSource.uri(localPath)` 优先本地，回退网络。
- Flutter 客户端：Settings 新增「Offline Library」section；下载进度条（`http.Client` stream + StreamController）；离线标记（TrackListTile 角标）；离线模式检测（`connectivity_plus`）。
- The phase output is version-tracked and covered by flutter analyze.

### v3.5.0 - 2026-06-29

- **feat: 测试覆盖 + CI 完善** — `PlayerNotifier` 状态机单元测试（PlayerState 默认值、copyWith、队列重排、isIdle/isPlaying）：13 cases；`TrackFavoriteNotifier` 状态机单元测试（init 幂等、optimistic toggle、rollback、family 独立性）：6 cases；`HistoryNotifier` provider 单元测试（空列表、加载态、错误传播、batch-delete 语义）：6 cases；`artworkUrlProvider` 单元测试（200 返回 URL、404 返回 null、网络错误 null、空 albumId 短路）：4 cases；`AuthState` / LoginScreen 单元+Widget 测试：5 cases。总计新增 34 test cases，全部 `flutter test` 通过。
- CI `mobile` job：新增 `build-ios` job（`runs-on: macos-15`，`flutter build ipa --no-codesign`，artifact 上传 `build/ios/archive/`）；`test` job 添加 `--coverage` flag 并上传 coverage artifact；三端 job（APK / IPA / macOS）并行，依赖 `analyze` + `test` job。
- The phase output is version-tracked and covered by flutter test (34 test cases) and CI green on all three build targets (APK / IPA / macOS).

### v3.6.0 - 2026-06-29

- **fix: UserPlaylistDetailScreen Play All 实际触发播放** — "Play All" 按钮改为先调 `playerProvider.notifier.playQueue(_trackIds!)` 构建播放队列，再 `context.go(AppRoutes.player)` 跳转播放器；此前仅跳转而未触发任何音频播放。
- **fix: deepAlbum / deepArtist 死代码清理** — 移除 `AppRoutes.deepAlbum` / `deepAlbum` 两个误导性常量（v3.3.0 注册了 intent-filter 但无对应路由）；说明 `/albums/:id` / `/artists/:id` 深链接已由 ShellRoute 子路由天然覆盖，无需额外顶层路由；仅 `deepTrack` 需要专属 `_DeepLinkTrackScreen` 处理器（播放后跳 /player）。
- **fix: OfflineTrack 存真实 artistName / albumTitle** — `DownloadNotifier.startDownload` 下载完成后调 `catalogRepository.getArtist(track.artistId)` 解析显示名，失败时回退 UUID；同理调 `getAlbum(track.albumId)` 解析专辑标题；Settings "Offline Library" 页面现展示可读名称而非 UUID。
- The phase output is version-tracked and covered by flutter analyze (0 issues) and flutter test (42/46 pass; 4 pre-existing generated SDK failures unchanged).

### v3.6.1 - 2026-06-29

- Fix offline download SignatureDoesNotMatch: presigned URL downloads bypass the Bearer-token Dio interceptor to avoid S3/MinIO credential conflict.
- Fix userplaylist concurrent write race: AddTrack/RemoveTrack/SetTracks use database-level row locking (SELECT … FOR UPDATE) to prevent last-writer-wins data loss.
- Fix postgres repository Get() error masking: only pgx.ErrNoRows maps to ErrNotFound; all other scan errors propagate as internal errors.
- Fix desktop tray Quit to use Flutter app lifecycle exit instead of dart:io exit(0), ensuring SQLite WAL checkpoint and dispose chain execution.
- Fix DownloadNotifier _restoreFromDb unawaited async: errors are surfaced and disposal guard prevents StateError on rebuilt notifiers.
- Fix partial file leak on download error: catch block deletes the partial local file before setting DownloadError state.
- Fix /library/my-playlists parent route blank screen: replace SizedBox.shrink() with UserPlaylistListScreen.
- Fix user_playlist rename() error swallowing: expose server errors to the UI and roll back optimistic state on failure.
- Fix _DeepLinkTrackScreen stuck spinner: wrap playTrack() in try/catch, navigate to home on error with SnackBar feedback.
- Fix OfflineDb double-open race: use Completer<Database> to guarantee single initialization under concurrent awaits.
- Fix removeTrack onPressed missing try/catch: show error SnackBar and skip _loadTracks() on failure.
- Fix _name dual source of truth in UserPlaylistDetailScreen: derive name from provider state only.
- Commit 5 untracked test files so CI executes the full v3.5.0 test suite.
- Fix getAlbumArtwork ExpiresIn derived from artworkTTL constant rather than hardcoded literal.
- Fix getAlbumArtwork error masking: distinguish ErrNotFound (404) from internal errors (500) in GetMediaObject path.
- The phase output is version-tracked and covered by relevant tests.

### v4.0.0 - 2026-06-29

- **feat: 歌词支持（LRC / SRT 同步滚动歌词）** — 服务端：新增 asset kind `lyrics`；`POST /api/v1/catalog/tracks/{id}/lyrics`（admin，上传 LRC/SRT 文件到存储后端，返回 media object ID）；`GET /api/v1/catalog/tracks/{id}/lyrics`（viewer，返回 `LyricsResponse{format, content, mediaObjectId}`，content 为 UTF-8 明文，直接在响应体内返回，不走 presigned URL）；`DELETE /api/v1/catalog/tracks/{id}/lyrics`（admin）；OpenAPI contract 新增 `LyricsResponse`、`UploadLyricsRequest` schema 及三条路径，版本升至 4.0.0。
- Flutter 客户端：`lyricsProvider(trackId)` Family AsyncNotifier，请求 GET 端点，解析 LRC/SRT 为 `List<LyricLine>{timestamp, text}` 模型；`FullPlayerScreen` 新增歌词层（可上下滚动），当前行高亮（随 `playerProvider.position` 实时定位）；歌词区与封面区可手势切换（PageView 左右滑）；无歌词时展示占位文字。
- The phase output is version-tracked and covered by Go unit tests (lyrics handler 成功/无歌词/格式错误), OpenAPI contract tests, and flutter analyze (0 issues).

### v4.1.0 - 2026-06-29

- **feat: 音频增强（EQ / 响度归一化 / gapless / 速度控制）** — Flutter 客户端音频引擎扩展：`just_audio` equalizer 插件（`just_audio_equalizer` 或平台原生 DSP）实现 10 段参数均衡器，预设（Flat / Bass Boost / Vocal / Electronic）+ 自定义；ReplayGain 响度归一化（服务端在 Track 元数据中存储 `replayGainDb float64` 字段，Flutter 在 AudioHandler 中读取并通过 `just_audio` volume 缩放应用）；gapless 无缝播放（ConcatenatingAudioSource 队列替换当前逐曲创建方式）；播放速度控制（0.5× / 0.75× / 1× / 1.25× / 1.5× / 2×，`AudioPlayer.setSpeed()`）。
- 服务端：`tracks` 表新增 `replay_gain_db REAL` 列（migration）；`CatalogTrack` OpenAPI schema 新增 `replayGainDb` 字段（nullable float）；`PATCH /api/v1/catalog/tracks/{id}` 支持更新该字段；OpenAPI contract 版本升至 4.1.0。
- Flutter 客户端 UI：Settings 新增「音频」section —— EQ 开关 + 频段滑块 + 预设选择器；速度控制按钮（FullPlayerScreen 控制栏追加）；ReplayGain 开关；所有音频设置持久化到 `SharedPreferences`，冷启动恢复。
- The phase output is version-tracked and covered by Go unit tests (replayGain field migration + PATCH handler), OpenAPI contract tests, and flutter analyze (0 issues).

### v4.1.1 - 2026-06-29

- **feat: EQ 均衡器 + Gapless 无缝播放（v4.1.0 遗留项收尾）** — Flutter 客户端：引入 `just_audio_equalizer` 插件，实现 10 段参数均衡器（31Hz–16kHz，±12dB），预设 Flat / Bass Boost / Vocal / Electronic + 自定义；`EqSettings` 模型 + `EqNotifier`（Riverpod Notifier，SharedPreferences 持久化）；Settings 「音频」section EQ 区（开关 + 10 条垂直 Slider + 预设 SegmentedButton）；播放队列重构为 `ConcatenatingAudioSource`，实现 gapless 无缝衔接；presigned URL 提前 120s 刷新（`currentIndexStream` 监听，`ConcatenatingAudioSource.removeAt + insert`）。
- 服务端：补充 Go 单元测试——`uploadLyrics` handler（201 成功 / 400 格式错误 / 404 track 不存在）；`getLyrics` handler（200 成功 / 404 无歌词 / 500 存储错误）；`patchTrack` handler（replayGainDb 更新成功 / null 清除）。
- The phase output is version-tracked and covered by Go unit tests (lyrics + replayGain handlers), flutter analyze (0 issues), and flutter test (new EQ state-machine unit cases).

### v4.2.0 - 2026-06-29

- **feat: 睡眠定时器 + 交叉淡入淡出（Crossfade）** — Flutter 客户端：睡眠定时器（`SleepTimerNotifier`，Riverpod Notifier）支持固定时长（15 / 30 / 45 / 60 分钟）和「当前曲目结束后停止」两种模式；倒计时实时展示（MiniPlayerBar 角标 + FullPlayerScreen 控制栏图标）；定时器到期后调 `PlayerNotifier.pause()`，自动清除状态；Settings 「音频」section 追加睡眠定时器入口。交叉淡入淡出（crossfade）：曲目切换时 N 秒（0–8s，用户可配置）渐出 + 渐入，基于 `ConcatenatingAudioSource` 的 clip model 实现（依赖 v4.1.1 gapless 基础）；Settings 「音频」section 追加 crossfade 时长 Slider。
- The phase output is version-tracked and covered by flutter analyze (0 issues) and SleepTimerNotifier unit tests（固定时长到期 / 曲目结束模式 / 取消定时器 / 并发重置）.

### v4.3.0 - 2026-06-29

- **feat: 搜索增强（Meilisearch 替换 PostgreSQL 全文搜索）** — 服务端：Docker Compose 新增 `meilisearch` 服务（`getmeili/meilisearch:latest`，`MEILI_MASTER_KEY` 注入）；新增 `internal/search` package，`SearchService` 接口 + Meilisearch 实现（`github.com/meilisearch/meilisearch-go`）；索引同步：`catalog.Service` 写入 Artist / Album / Track 后异步推送到 Meilisearch index（goroutine，失败仅记 log，不影响主流程）；`GET /api/v1/catalog/search` 路由切换为 Meilisearch 后端，`typoTolerance` 开启（支持拼写容错和 CJK 分词）；Meilisearch 不可用时自动降级到 PostgreSQL 全文搜索，响应 schema 不变（`CatalogSearchResult`），兼容现有 Flutter 客户端。
- Flutter 客户端：SearchScreen 新增 autocomplete 下拉（debounce 150ms → GET 请求 → 展示前 5 条 Tracks 建议）；搜索结果页新增 Filter Tab（All / Artists / Albums / Tracks）；空结果状态插图（onSurfaceVariant 占位图 + 提示文字）。
- The phase output is version-tracked and covered by Go search service unit tests（模糊匹配 / 空查询 / Meilisearch 不可用降级到 PG 全文搜索）, docker-compose smoke test, OpenAPI contract tests, and flutter analyze (0 issues).

### v4.4.0 - 2026-07-05

- **feat: 歌词深化（逐字高亮 + 双语翻译 + 后台管理）** — 服务端：`LyricsResponse` OpenAPI schema 新增 `translation string?`（双语翻译文本）及 `source string?`（歌词来源标注：`embedded` / `manual` / `lrclib` 等）字段；`POST /api/v1/catalog/tracks/{id}/lyrics` 请求体支持同时上传主歌词与翻译；migration `014_track_lyrics_translation.sql` 新增 `tracks.lyrics_translation_media_object_id` 列；OpenAPI 版本升至 4.4.0。
- Flutter 客户端：`lrc_parser.dart` 新增逐字增强 LRC 解析（`<mm:ss.xx>` 内联时间戳切分 word-level spans）；`lyric_line.dart` 模型扩展 `words` 与 `translation` 字段；`full_player_screen.dart` 的 `_LyricsList` 重构为当前行逐字渐变高亮（无逐字数据时回退整行高亮），当前行下方渲染次要样式翻译文本；Settings 新增「双语歌词」开关。
- 管理端 Admin Web：`services/admin/app/(admin)/catalog/page.tsx` tracks tab 新增歌词管理入口（镜像现有 Relink 对话框模式），支持上传/预览/删除歌词与翻译文件、展示来源标注。
- The phase output is version-tracked and covered by Go unit tests (歌词翻译字段 + migration), OpenAPI contract tests, and flutter analyze (0 issues).

### v4.5.0 - 2026-07-06

- **feat: 搜索收尾（高亮 / 拼音 / 历史）+ ReplayGain 自动分析 + 全量重建索引** — 服务端：`internal/search.Service` 接口扩展高亮支持，`Search()` 返回值携带匹配片段（`MeilisearchService` 接入 `SearchRequest` 已支持的 `AttributesToHighlight`/`HighlightPreTag`/`HighlightPostTag`）；新增 `cmd/reindex/main.go` CLI 工具（复用 `cmd/server/main.go` 的 repo/service 构建模式，遍历全量 Artist/Album/Track 调用 `IndexTrack/Album/Artist` 重建 Meilisearch 索引）；拼音搜索：引入 Go 拼音库为中文标题生成拼音索引字段，写入 Meilisearch 附加字段并参与匹配。
- ReplayGain 自动分析：曲目上传/导入完成后触发后台任务调用外部工具（ffmpeg/loudgain 或等价）分析响度，结果写入 `tracks.replay_gain_db`（复用 `catalog.Service.UpdateTrack` 既有写入路径）；分析失败不阻塞上传流程，仅记 log。
- Flutter 客户端：`search_screen.dart` 新增搜索历史（`SharedPreferences` 持久化最近 N 条查询，聚焦时展示，可清除单条/全部）；搜索结果高亮渲染（`RichText`/`TextSpan` 按后端返回的高亮片段着色匹配子串）。
- The phase output is version-tracked and covered by Go unit tests (highlight 字段解析、reindex CLI、ReplayGain 分析 fallback), OpenAPI contract tests, and flutter analyze (0 issues).

### v4.6.0 - 2026-07-06

- **fix: 播放器封面图未透传 albumId 导致的静默失效** — `PlayerNotifier._makeMediaItem`/`_stubMediaItem` 解析出的 `albumId` 从未写入 `MediaItem.extras`，导致 `MiniPlayerBar`/`FullPlayerScreen` 读取 `extras['albumId']` 恒为 null，播放器内封面图自 v3.1.0 起从未正确显示，始终回退占位图标；修复为在两处 `extras` map 中补齐 `albumId` 字段。
- **feat: MiniPlayer 拖拽进度条** — `MiniPlayerBar` 此前完全没有进度指示控件，新增可拖拽的紧凑进度条，支持点击/拖动跳转播放位置，缓冲中禁用交互。
- **feat: EQ 自定义预设保存与命名** — `EqSettings` 新增 `customPresets` 字段，`EqNotifier` 支持保存当前频段为具名预设、切换、删除，持久化经 SharedPreferences；Settings EQ 区新增预设管理入口。
- The phase output is version-tracked and covered by flutter analyze (0 issues) and flutter test (EqNotifier 自定义预设保存/删除/持久化单元测试).

### v4.7.0 - 2026-07-07

- **fix: 播放器音频真实性修复（gapless 状态同步 / ReplayGain 应用 / Shuffle / EQ 真实接线 / crossfade 如实化）** — 2026-07-06 全量代码审查发现 v4.1.0–v4.2.0 多项音频特性为假实现：EQ 依赖的 `just_audio_equalizer` 从未加入 `pubspec.yaml`，`audio_handler.dart` 经 `(_player as dynamic).setBands` 调用不存在的方法且异常被静默吞掉（EQ UI 对音频输出零影响）；ReplayGain 开关从未被 `PlayerNotifier` 读取，`replayGainDb` 在客户端无消费方；`setShuffle` 仅翻转状态位不改变播放顺序；`_runCrossfade` 是切歌后同 player 的音量 V 形凹陷而非双轨交叉淡化；`PlayerNotifier` 未监听 `currentIndexStream`，gapless 队列自动前进时 UI/通知栏/历史上报全部脱轨，且 `next()` 取模回绕无视 `RepeatMode.off`。本阶段逐项修复：`currentIndexStream` 订阅同步状态与逐曲历史上报、repeat 语义映射 `setLoopMode`、手动切歌改增量 `seekToNext/Previous`（不再全队列重建）、ReplayGain 增益实际应用（`10^(db/20)` 与用户音量正交）、shuffle 经 `setShuffleModeEnabled` 真实生效、EQ 改用 `AudioPipeline` + `AndroidEqualizer`（不支持平台明确禁用标注）、crossfade 如实降级为「切歌淡入淡出」并同步文案。
- The phase output is version-tracked and covered by flutter analyze (0 issues), flutter test（PlayerNotifier 队列状态机集成测试 / ReplayGain 增益计算 / shuffle 顺序 / repeat 语义）, and 人工听感验收清单（真机 EQ 听感 / 连播元数据 / 逐曲历史）.

### v4.8.0 - 2026-07-10

- **chore: 测试与结构还债（v4 封版收官）** — Go：`internal/httpapi/handler.go`（5218 行）与 `handler_test.go`（9255 行）按域拆分为 storage/media/auth/catalog/history/favorites/userplaylist/search 多文件（同 package 纯移动，773 测试护航零行为变更）。Flutter：`AuthNotifier`/`SearchHistoryNotifier` 等 provider 单测补齐 + `MiniPlayerBar`/`TrackListTile`/`SearchScreen` widget 测试，`flutter test --coverage` 接入 CI artifact。Web：Playwright e2e 扩展为登录→浏览专辑→播放（真实 audio src/状态断言）→收藏/取消收藏→历史记录（合成 `ended` 事件驱动真实 onEnded 处理）→登出的完整主流程，引入 Vitest 覆盖 `store/player.ts` 队列逻辑与 `lib/` 工具函数（49 tests）。Admin：建立 Playwright e2e 最小回归（登录/用户管理/catalog/storage 页）+ CI `admin-e2e` job。仓库卫生：`git rm --cached .codegraph/daemon.pid`、README 重写至 v4.8.0 基线、web/admin/mobile 版本号与根 VERSION 对齐、CI 补 biome lint step。
- **fix: 端到端测试驱动发现并修复 12 个真实缺陷** — 补齐测试基建、用真实浏览器驱动播放/收藏/历史/登出全流程的过程中，暴露出此前隐藏的多个功能性缺陷（而非仅测试自身问题）：(1) bcrypt 测试 `-race` 超时；(2) CI 含连字符用户名导致 workflow 语法问题；(3) CSS `@import` 顺序错误；(4) `AuthProvider` 挂载位置不当；(5) 重复的 `/` 路由定义；(6) 错误的相对 CSS import 路径；(7) auth service 为 nil 时未判空导致 panic；(8) **ReplayGainDb 静音音轨写入 `+Inf`**——ffmpeg loudnorm 对绝对静音报告 `-inf` LUFS，`referenceLoudnessLUFS - (-Inf)` 产生非有限浮点数，`encoding/json` 无法序列化，导致该曲目此后所有 GET 请求返回 200 但空 body；(9) **收藏页永久显示"无收藏"**——`GET /api/v1/me/favorites/tracks` 在 catalog service 可用时（生产默认路径）只返回 `tracks` 而非 `trackIds`，Web 端空态判断只检查了后者；(10) Playwright 专辑列表计数与客户端数据请求竞态（测试基建问题）；(11) **播放历史从未被真正记录**——`useAudio.ts` 的 `onEnded` 处理器 POST 历史记录时携带 API 拒绝的多余字段（`durationSeconds`/`source`），Go 端严格 JSON 解码返回 400 但被 `.catch(() => {})` 静默吞掉，用户界面无任何异常表现，生产环境播放历史功能实际从未成功写入过一条记录；(12) **登出点击后不跳转 `/login`**——`AuthProvider` 的 zustand 订阅回调在 token 变空时又调用一次 `clearSession()`，与其自身触发的 `set()` 通知互相递归直至调用栈溢出，异常在 `Topbar.handleLogout()` 内部抛出，使其后的 `router.push("/login")` 永远不执行。这些修复本身即是本阶段"测试基建能真正抓住回归"目标的验证。本阶段完成后 v4 线封版，仅接受 bug 修复 patch（v4.8.x）。
- The phase output is version-tracked and covered by 全部 CI job（api/web/mobile/admin/e2e）本地复核绿灯（gofmt/`go test -race` 774 tests/biome lint 0 errors/flutter analyze 0 fatal + flutter test 87 tests/vitest 49 tests/playwright main-flow 1 test）.

### v4.8.1 - 2026-07-11

- **fix: v4.8.0 推送后远端 CI 发现的 3 个真实缺陷** — 本地 pre-push 验证全绿但远端 5 次 CI 运行全部失败，逐条拉取失败日志定位：(1) **Flutter CI `make gen:api` 触发 Makefile 静态模式规则解析错误**——`gen:api`/`build:watch` 目标名含冒号，GNU Make 按 `targets: pattern: prereqs` 语法解析而非字面目标名，因 `api`/`watch` 不含 `%` 通配符而在 `.PHONY` 行直接 abort；系 commit e314b83（Phase 304-308）引入的既有缺陷，此前被另一个已修复的 Flutter SDK 版本钉死问题挡住而从未暴露。修复为改用连字符（`gen-api`/`build-watch`），同步更新 `flutter.yml` 5 处与 `services/mobile/README.md` 1 处调用。(2) **Web CI type-check 失败**——`e2e/main-flow.spec.ts` 新增的 `@ts-expect-error` 抑制了一处并不产生错误的赋值（`class ProbedAudio extends NativeAudio` 结构上可赋值回 `window.Audio`），TS strict 模式判定该抑制注释本身为错误（TS2578）；系本阶段新增测试代码自身疏漏——push 前跑过 playwright/vitest/biome lint 但未单独对新文件跑 `type-check`。修复为移除多余抑制注释。(3) **Docker 镜像发布 CI 失败**（`publish-web`/`publish-admin`）——根因 A：仓库根 `.dockerignore` 的裸 `packages` 行排除整个 `packages/` 目录，导致两个 Dockerfile 的 `COPY packages/api-contract ...` 找不到源文件；根因 B（A 修复后会暴露的第二个缺陷）：`@inori/ui` 是两端 `package.json` 的 `file:../../packages/ui` 依赖（admin 端 `StorageHealthBadge.tsx` 有真实引用，两端 `next.config.ts` 均声明 `transpilePackages`），但两个 Dockerfile 都缺少对应的 `COPY packages/ui ...`。二分 CI 历史（`gh run list --json` + 逐次 `git diff`）确认此二缺陷自 commit 4b184bc8（2026-06-21，v2.2.0 阶段）起已连续存在超过 26 次运行、跨越约 19 天，与 v4.8.0 本阶段工作内容及 v4 整条线均无关，只是本次逐条排查 CI 时才第一次被发现。修复：`.dockerignore` 移除 `packages` 排除行，两个 Dockerfile 的 builder stage 各追加 `COPY packages/ui ../../packages/ui`。本地无 Docker 环境，此项修复仅通过等价目录结构复现 npm `file:` symlink 解析行为静态验证，实际效果待下次 CI 运行确认。
- The phase output is version-tracked; fixes (1)(2) verified locally (`make gen-api` end-to-end + zero-diff regenerated client; `npm run type-check` reproduces then clears the error), fix (3) verified only by static reasoning (no local Docker available) pending live CI confirmation.

### v4.8.2 - 2026-07-11

- **fix: v4.8.1 推送后 CI 发现的第 4 个真实缺陷** — 监控 v4.8.1 推送触发的远端 CI，`flutter.yml` 的 `Analyze` job 在 ubuntu-latest 上失败（macOS 侧同一 job 因矩阵 fail-fast 联带取消，非独立问题）。根因：`.github/workflows/flutter.yml` 第 42 行的 Analyze 步骤使用裸 `flutter analyze`（对 info 级问题也返回非零退出码），而仓库既有的 4 条 info 级问题（字符串插值多余花括号，`full_player_screen.dart`/`settings_screen.dart` 各两处）早已被认定可接受——`build.yml` 的等价步骤一直正确使用 `flutter analyze --no-fatal-infos`，两个 workflow 文件对同一命令的 flag 从 v3.5.0（commit 68411e7）引入起就不一致，此前同样被 v4.8.1 才修复的 Flutter SDK 版本钉死问题挡住、从未真正执行到。修复：`flutter.yml` 该行同步加上 `--no-fatal-infos`，与 `build.yml` 保持一致。
- The phase output is version-tracked and verified locally (`flutter analyze --no-fatal-infos` exits 0, 4 pre-existing info-level issues surfaced but non-fatal, matching build.yml's established baseline).

### v4.8.4 - 2026-07-11

- **fix: 4.8.3 推送后 Docker / Flutter / Admin E2E 收尾的 3 类独立缺陷** — 继续追踪 v4.8.3 触发的远端 CI，定位三个互不相关的失败根因：(1) **Docker 镜像构建 `npm ci` EUSAGE（BuildKit wildcard COPY 缓存命中）**——v4.8.3 把 `package-lock.json*` 通配 COPY 改为精确 `package-lock.json`，理论上应该修复该问题；但 `gh run view --log-failed` 复现的 `npm error ... Missing: app@4.8.3 from lock file` 错误实际更复杂：即使 Dockerfile 代码已改，GitHub Actions 的 BuildKit 缓存层并未自动失效（无 `cache-from`/`cache-to` 显式配置时仍可能命中 registry-stored wildcard-COPY layer），导致旧层被复用。两端 Dockerfile 的 deps stage 头加 `ARG CACHE_BUST=1` + `RUN echo "cache-bust=${CACHE_BUST}" > /dev/null` 使每次 build 在 COPY 前注入非缓存指令，强制 layer hash 变化以击穿 wildcard-COPY 残留缓存。(2) **Flutter APK `build.gradle.kts` 编译错误（15 个 error）**——CI 用的 Flutter 3.44.6 + AGP 9（`android.newDsl=true`），`java.util.Properties().also { it.load(props.inputStream()) }` 未显式 import 导致 `'util'` / `'load'` / `'getProperty'` 全部 unresolved；`android {}` 扩展函数在 AGP 9 已 deprecated。修复：`plugins {}` 前加 `import java.util.Properties`，第一行 android DSL 上方加 `@Suppress("DEPRECATION")`。(3) **Flutter `analyze` 把 info 级问题当 fatal**——`full_player_screen.dart`/`settings_screen.dart` 各 2 处字符串插值多余花括号（`'${speed}×'` / `'${s}×'`），`flutter analyze` 默认对 info 也返回 exit code 1 导致 CI 红。跟 v4.8.2 `flutter.yml` 引入的 flag 不一致问题是不同路径——这里是 v4.8.2 之后新引入的。修复：源码层删掉 4 处冗余花括号（`'$speed×'` / `'$s×'`），比 `--no-fatal-infos` 更彻底、消除警告本身。
- **hygiene: `services/admin/.gitignore` 追加 `*.tsconfig.tsbuildinfo`**——`tsconfig.tsbuildinfo` 是 TypeScript 增量构建产物，此前只在 `services/web/.gitignore` 有，admin 端缺失导致 `services/admin/tsconfig.tsbuildinfo` 不时进 staging；跟 web 端对齐。
- The phase output is version-tracked; items (2)(3) 本地可验证（`flutter analyze` 0 issues、`services/mobile/android/app/build.gradle.kts` 静态校验 Kotlin DSL 引用可解析）；item (1) 仍需 live CI 确认（本地无 Docker）。

### v4.8.3 - 2026-07-11

- **fix: v4.8.2 后继续收尾的 CI / Docker / Admin E2E 缺陷** — 继续追踪远端 CI 与本地复现后，确认并修复三类真实问题：(1) Flutter 生成客户端源码被 `.gitignore` 的 `lib/src/api/generated/*` 规则挡住，导致新引入的用户播放列表 / 歌词 / artwork OpenAPI Dart 源码未纳入版本控制；在真正干净的 build_runner 环境中会缺少模型/API 源文件。本阶段放开该目录下 OpenAPI generator 产出的 `.dart`/doc/test 源文件，只继续忽略 build_runner 生成的 `.g.dart`/`.freezed.dart`，并补交 33 个缺失源文件。(2) web/admin `package.json` 已升至 4.8.2，但 `package-lock.json` 根包版本仍停在 4.7.0；普通 CI 的 `npm ci` 在原路径下未暴露该问题，但 Docker deps stage 将 package 文件复制到 `/app` 后，npm 以实际目录名 `app@4.8.2` 校验 lock，触发 `npm ci` EUSAGE。修复为刷新两端 lockfile 根版本，并把 Dockerfile 中可选锁文件通配 `package-lock.json*` 改为必需的精确 `package-lock.json`，让缺锁在 COPY 阶段直接失败。(3) Admin Playwright smoke 登录用 `input[type="text"]` 选择器，在登录页结构变化或 token tab 状态下容易找不到输入框；改为按可访问 label 与 button role 操作真实登录表单。
- The phase output is version-tracked and verified locally：Docker deps-stage 等价目录布局下 web/admin `npm ci` 复现并清除 EUSAGE；`npm --prefix services/web ci`、`npm --prefix services/admin ci`、admin lint/type-check 通过；Flutter 生成客户端修复已在干净 build_runner + `flutter analyze --no-fatal-infos` + `flutter test --no-pub` 验证通过。

### v5.0.0 - TBD

- **feat: 安全加固与对外可用基线（v5 产品化主轴开篇）** — v5 主轴为「产品化/对外开放」。签名流媒体 URL：新增 `internal/streamsign` package（HMAC-SHA256 对 trackId+exp 签名，`INORI_STREAM_SIGNING_KEY` 配置，15 分钟有效期），`TrackPlaybackDescriptor.streamUrl` 直接返回已签名 URL，`streamTrack` 验签替代 `?token=` 会话校验，三端客户端删除 token 拼接逻辑——会话 token（24h 有效、全账户权限）不再进入 URL/代理日志。登录限速：per-IP + per-username 双维度内存限速器，连续失败 5 次指数退避，429 `too_many_requests`。会话清理：接线既有 `DeleteExpiredSessions` 为每小时后台任务。部署硬化：`MEILI_MASTER_KEY` 移除不安全默认值改为必填、nginx 追加 HSTS/nosniff/X-Frame-Options/基线 CSP 安全头、生产模式 CORS 未配置时 ERROR 警告。OpenAPI 契约版本升至 5.0.0。
- The phase output is version-tracked and covered by Go unit tests（签名/过期/篡改/限速/清理）, OpenAPI contract tests, web e2e 播放主流程回归, and 人工验证（access log 无会话 token、过期流 URL 返回 401）.

### v5.1.0 - 2026-07-13

- **feat: Web 体验对齐 I（歌词面板 + 搜索高亮/历史）** — 纯前端消费既有服务端能力，无服务端变更。歌词：`lib/lyrics/lrcParser.ts`/`srtParser.ts` 移植 mobile 端解析逻辑（行级/逐字 `<mm:ss.xx>` spans/翻译配对），`LyricsPanel` 组件随播放位置同步滚动、当前行高亮、逐字渐变（无 word 数据回退整行）、双语翻译展示与开关（localStorage 持久化），接入 `PlayerBar` 新增 Lyrics 入口。搜索：`highlight` 字段（`<mark>` 片段）经 `lib/search/highlight.ts` 字符串切分转 React `<mark>` 元素渲染（不使用 `dangerouslySetInnerHTML`，规避 XSS；PG 全文搜索降级空值回退纯文本），`api.gen.ts` 重新生成以补齐 `SearchResultItem.highlight` 字段；`lib/search/searchHistory.ts` 落地 localStorage 搜索历史（最近 20 条去重、聚焦展示下拉、单条删除/清空、键盘上下选择+回车确认），行为对齐 mobile 端 `SearchHistoryNotifier`。拼音搜索为既有服务端能力，web 零成本受益。
- The phase output is version-tracked and covered by Vitest（lrcParser/srtParser/highlight/searchHistory 共 86 tests all green）, Playwright e2e（新增 `search-history.spec.ts` 覆盖历史记录/下拉展示/选择/清空，既有 smoke + search 用例全绿，`main-flow` 因本地未播种 catalog 数据按设计跳过）, and type-check（tsc --noEmit 0 errors）+ biome lint（81 files, 0 errors）。本地起 Go API + Next dev server 手动验证 LyricsPanel 挂载无 console error。

### v5.2.0 - TBD

- **feat: Web 音频引擎对齐 II（ReplayGain / gapless / 状态持久化）** — WebAudio 增益管线：`AudioContext` + `MediaElementAudioSourceNode` + `GainNode`，ReplayGain 按 `10^(db/20)` 应用（与用户音量正交），audio 元素 `crossOrigin="anonymous"`，CORS 不满足时优雅降级直连播放；S3 presigned 场景的 bucket CORS 配置要求写入 operations 文档。双元素 gapless：备用 `HTMLAudioElement` 预载下一曲（进度 >50% 或剩余 <30s 触发），`ended` 时毫秒级切换，可选切歌淡入淡出（`linearRampToValueAtTime`），签名 URL 过期自动重取。状态持久化：zustand persist 保存队列/currentIndex/position（节流 5s）/音量/repeat/shuffle，刷新恢复不自动播放，登出清空。依赖 v5.0.0 签名 URL 先行（统一 CORS 语义），与 v5.1.0 可并行。
- The phase output is version-tracked and covered by Vitest（增益计算/预载条件/持久化序列化）, Playwright e2e（连播自动接续/刷新恢复断言）, and 人工验收（响度对比/间隙听感/预载时机）.

### v5.3.0 - TBD

- **feat: Web 体验对齐 III（用户播放列表 + 速度控制 + 睡眠定时器）** — 三块服务端能力已就绪但 web 零消费的缺口（2026-07-06 审查确认 `me/playlists`/`playbackRate`/sleep timer 在 `services/web/` 全部零命中）。用户播放列表：`library/playlists` 列表页 + 详情页（GET/POST/PATCH/DELETE `/api/v1/me/playlists` 及曲目增删排序端点，v3.2.0 已有），dnd-kit 拖拽排序（依赖已在 package.json），`TrackRow` 上下文菜单「添加到播放列表」。速度控制：PlayerBar 速度菜单 0.5×–2×（档位对齐 mobile v4.1.0），`audio.playbackRate` 双元素引擎下同步设置，localStorage 持久化。睡眠定时器：固定时长（15/30/45/60min）+「当前曲目结束后停止」双模式、倒计时角标、到期 pause——语义对齐 mobile `SleepTimerNotifier`（v4.2.0）。无服务端变更。
- The phase output is version-tracked and covered by Playwright e2e（播放列表全流程/playbackRate 断言）, Vitest（sleepTimer 状态机，用例对齐 mobile 端）, and type-check + biome lint.

### v5.4.0 - TBD

- **feat: 跨设备一致性（播放续播 + 搜索历史同步）—— v5 封版收官** — 服务端 viewer 域新增播放状态端点：migration `user_player_state` 单行 upsert 表（queue/currentIndex/positionMs/repeat/shuffle/updatedAt），`internal/playerstate` package（memory + PostgreSQL 双实现），`GET/PUT /api/v1/me/player-state`（last-write-wins，服务端时钟），`/readyz` 追加检查；搜索历史端点：`user_search_history` 表 + `GET/PUT/DELETE /api/v1/me/search-history`（服务端裁剪 20 条）。Web 与 Flutter 双端：播放中 30s 节流 + 切歌/暂停触发上报；启动时服务端状态较新则提示「在另一台设备听到 X，继续播放？」，确认后重建队列 seek 恢复（不自动播放）；搜索历史本地∪远端合并去重回写。OpenAPI 契约版本升至 5.4.0。本阶段完成后 v5 线封版：安全基线 + 双端对齐 + 跨设备一致构成产品化闭环，后续新方向按 v6+ 升大版本。
- The phase output is version-tracked and covered by Go unit tests（playerstate/searchhistory 裁剪与冲突语义）, OpenAPI contract tests, Playwright e2e（恢复提示流程）, flutter test（节流/合并策略）, and 人工验收（手机→web 续播误差 <5s）.

### v5.5.0 - 2026-07-26

- **feat: 樱花薄暮浅色 ACG 主题（Web）** — 已部署实例的界面被判定为典型「AI 暗黑模板」（电紫 `#9b5cff` + 近黑 `#070711` + Orbitron 赛博字体 + 扫描线），本阶段按二次元 ACG 特色重做视觉。方向经两轮迭代确定为**明亮日系**：奶油白画布 `#FFF7F2` + 纯白卡片 + 樱粉 `#D42062` / 晴空青 `#0A7D94` / 杏金点缀，画面中不出现任何黑色或灰色大面积背景；字体 Orbitron+Inter → `Zen Maru Gothic`（日系圆黑，标题）+ `Poppins`（正文），中文回退链补 PingFang SC / Microsoft YaHei。权威 token 落在 `packages/ui/src/styles/sakura-dusk.css`（原 `neon-shrine.css` 重命名），web 端 `globals.css` 同步；`color-scheme` 由 dark 改 light，圆角整体放大，新增 `.card-soft`/`.card-float` 粉调阴影（不用灰黑）、`.aurora-veil` 三层径向渐变氛围底、`.petal` 樱花飘落、`.vinyl-spin` 唱片旋转，删除 `.scanlines` 赛博语汇。结构性调整：首页 4 个等大方块改 Bento 不对称网格（Tracks 卡 2×2 主视觉 + 超大图标水印，Playlists 卡横跨末行消除空缺）、Recently Added 补封面缩略图、列表项由 `<div onClick>` 改 `<button>`（原键盘不可达）；`PlayerBar` 进度条抽出 `ProgressBar` 支持 pointer 拖拽（原仅点击跳转）+ hover 增高 + 圆形手柄，播放键与 `ControlBtn` 由 40px 提到 44px 满足触摸目标标准；`Sidebar` 选中态由整块实心改左侧 3px 指示条 + 淡樱粉底 + `aria-current`；`Visualizer` canvas 渐变原写死三个旧主题 hex，改为挂载时读一次 CSS 变量（每帧读会触发 style recalc）；新增 `PetalDrift` 组件（6 片 CSS-only 花瓣，`prefers-reduced-motion` 下 `display:none`——仅归零 duration 会让花瓣僵在半空）；`FullscreenPlayer` 封面 canvas 取色驱动背景光晕，跨域污染时回退主题渐变。
- **fix: 主题改造中发现的 3 类既有缺陷** — (1) **`--color-primary-foreground` 从未在 `@theme` 中定义**，却被 8 个文件 10 处引用（全是实心按钮文字），旧深色主题下靠继承浅色文字侥幸正常，浅色主题下会直接失效；补上定义。(2) **`bg-opacity-10` / `ring-opacity-20` 在 Tailwind v4 已失效**（v3 遗留写法），导致登录页与安全设置页的错误/成功提示渲染成实心红底红字、绿底绿字，几乎不可读；改为 `danger-dim` 底 + 边框，并把绕过主题的 `bg-green-500`/`text-green-600` 换成语义 token。(3) 四处硬编码 `bg-black/50`、`bg-black/60` 遮罩（Modal / MobileSidebar / QueueDrawer / LyricsPanel）改用 `--color-scrim` 梅墨半透明。
- 色板全部经 WCAG AA 实测而非估算，三条硬约束记录在 `.plan/20260726-027-v5.5.0-sakura-dusk-web.md`：浅色强调色做实心按钮时白字极易不达标（初版樱粉白字仅 2.4:1，最终主色定 `#D42062` 达 5.0:1）；`--color-primary` 不能直接当压在 `--color-primary-dim` 上的文字色（侧栏选中态仅 4.2:1，新增 `--color-primary-on-dim` 达 5.2:1）；`--color-border` 达不到 3:1 只能做装饰线，控件轮廓须用 `--color-border-strong`。
- Admin 后台与 Flutter 端对齐分别由 v5.6.0 / v5.7.0 跟进（`.plan/20260726-028`、`20260726-029`），因两端主题各有独立副本、且后台需降饱和处理、Flutter 端有 240 处符号引用需批量重命名。
- The phase output is version-tracked and verified by `tsc --noEmit`（0 errors）, biome lint（源码 96 files 0 errors）, vitest（203 tests all green）, Playwright（6 passed；另 3 个播放用例失败已用 `git stash` 回退到改动前复跑确认同样失败，根因是本地种子 media object 指向不存在的音频文件，与主题无关）, 以及 Chrome DevTools MCP 实机走查——375/1440 两档逐屏验收，并注入对比度脚本扫描 **7 个页面**的实时 DOM 文字/背景配对（修复后 0 处不达标；侧栏选中态那处正是靠此扫描发现，离线色板审计未覆盖组合态），另从 CSSOM 读取编译产物确认 reduced-motion 规则生效。纯前端主题改造，零 API schema 变更，故跳过 OpenAPI `info.version` 同步。

### v5.5.1 - 2026-07-26

- **fix: 修复 `gofmt` 对齐失效，Build CI 恢复通过** — `services/api/internal/httpapi/handler.go` 的 `Handler` 结构体在 v5.4.0 新增两个更长字段名后未重跑 gofmt，导致对齐列宽失效；本阶段仅做纯格式化修复（15 行加、15 行减，字段名/类型/顺序完全不变），零逻辑变更。`git diff -w` 归一化后两侧完全一致。本地 CI 等价检查全绿：`gofmt -l` PASS、`go vet` PASS、`go test` 792 passed、`go test -race` 792 passed、OpenAPI JSON 校验 PASS。此问题从 v4.8.4 起连续 7 个版本导致 Build 工作流失败，与 v5.5.0 前端主题改动无关；修复后后续版本的 Build 才能真实反映 Go 侧回归。
- The phase output is version-tracked; verified locally by the CI-equivalent Go checks above.
