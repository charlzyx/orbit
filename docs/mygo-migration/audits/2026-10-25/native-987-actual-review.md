# 987 native 原图人工复核（2026-10-25）

成功运行：[40219606106](https://github.com/charlzyx/orbit/actions/runs/40219606106)。Artifacts：`go-native-desktop`（12609551194）、`pinned-reference-desktop`（12609152659）。

## 范围和来源

实际逐张 `view_image` 看了 WB01–WB08 两端共 16 张，以及 Usage Simple/Details 两端 4 张。未调整、裁切、拼接或覆盖原图。其余 15 个基础页未在本报告复核，不能据此称 25 对全部通过。

- Go：`987eacb79805b46a2152fc5003f8eb7bc59d0598`，binary SHA `91bea0d4cb1f7675966e27b281dd90a9baeaf69f828bc20765e48e595256db78`。
- Rust：`299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`。
- Go ZIP SHA256：`fc52b84ec4ba65328dad68dde002577bf5e1b26935b5dd959365e464212fda11`。
- Rust ZIP SHA256：`0fa6030231c148ad7519ebe98cc3b8a8be500b115a8002bb83e68d6a21f5b1f2`。
- 每张 1470×922，Linux 原生客户端捕获，不是 macOS。相同字体偏好不证明实际 glyph face 相同。
- 工作区原始 fixture 字节相同，staged alpha / unstaged bravo / untracked notes 和两条 commit 相同。捕获后两端各追加一条允许的 `orbit:workflow build` metadata，原六条 session entries 保留；**不能称捕获前后 session 字节完全不变**。

## 按张实际观察

图路径前缀：`go/go-desktop/workbench/` 与 `rust/reference-desktop/workbench/`（相对于解包后的双 artifact 根目录）。以下是可见静态状态复核，不是交互功能全 PASS。

| 状态／原图文件 | 看得见的语义 | 实际差异与尚未验证 |
|---|---|---|
| WB01 `wb01-transcript-dark.png` | 两端同一会话、user 请求、三项列表、Go code block、No model request、空闲 composer 均在 | Go user metadata 在左、Rust 在右；列表行距/代码块头/正文整体高度不同，user bubble 顶约 y453 vs y377。Go sidebar 无 Rust 的 current、workspace 行、24d 样式。没有测试复制、选区或发送。 |
| WB02 `wb02-explorer-dark.png` | 两端右侧同一约236px Explorer，docs/src收起、alpha/bravo M、notes U、底部5files | GO/TXT badge 当前在图标内，无此前严重纵向错位；不能推广为全部 Explorer 功能通过。图标/字形/选区细节仍不同；目录展开、重命名、删除、隐藏文件未测。 |
| WB03 `wb03-file-alpha-dark.png` | alpha 内容为 `fixtureAlpha` / `return "staged"`，单tab和Explorer当前文件正确 | Go top header 为 Files、全绝对路径、Save+开关；Rust保留会话title、文件名、65B/5lines/Saved。Go底部与gutter为6行，Rust摘要5lines但gutter也有空的第6行：**首先是行数摘要口径差异，不能断言文件多写一行**。Go文字行高约20 vs Rust18、tab无文件badge、文件面板边框/内边距不同。未编辑/保存，不能从Save字样判断保存失败。 |
| WB04 `wb04-file-tabs-dark.png` | alpha/bravo双tab，bravo active；内容为 `fixtureBravo` / `"unstaged"`，Explorer选中bravo | 同 WB03；Rust67B/5lines，Go6lines。不能从静态双tab证明dirty关闭确认或切回定位正确。 |
| WB05 `wb05-review-uncommitted-dark.png` | Uncommitted、alpha/bravo/notes三文件、总+3/-2、文件树显示且Explorer隐藏 | 同一右dock起点约x1022，但Go多独立+/− gutter，文本可用宽更窄，新增return字符串、notes会多折一行；Rust notes一行、Go两行。Rust文件头有文件图标及每文件+/−统计，Go有open箭头而无该统计。wrap-on状态可见，不证明wrap-off长行、横滚、sticky、gap expansion或选择复制。 |
| WB06 `wb06-git-changes-dark.png` | alpha staged1；bravo+notes unstaged2；+/−数值与原fixture吻合 | Go旧大按钮/搜索控制栏、文件名居中、行高/列表背景；Rust紧凑条目、左文件名、section行内操作、页容器边框。两端当前都折叠diff；**没有证据断言Go缺diff内容**。没有stage/discard/commit写操作验证。 |
| WB07 `wb07-git-history-dark.png` | 相同两commit消息、短hash2605b2a/2657390、capture-main和作者存在 | Go搜索/作者/path/Follow控制常驻；Rust只有All branches按钮。Go作者居中、branch右侧；Rust作者在消息下左侧、branch在消息前、行尾箭头；全页容器与icon不同。未打开commit、过滤、Follow rename或菜单。 |
| WB08 `wb08-terminal-dark.png` | 两端真正终端区域均显示 `Terminal fixture ready.` 和 `orbit-fixture>`；底边dock起点约y662相同 | Rust tab内有close x，Go该tab内无x但工具栏有全局x/其他按钮；tab高度、背景和cursor表现不同（cursor相位属volatile）。未键入任何命令，**不证明退出/restart、PTY resize、选择复制、Tab focus正确**。 |

## Usage 两对补充

- `usage-simple-dark.png`：两端实际完成态，无 Loading skeleton；33 requests、55.4K tokens、4 sessions、9.5% cache、5.4s均可见，heatmap/Summary/Usage over time区块与主要坐标接近。Go summary分隔线、字体与间距仍异；不能称像素相等，也未测filter/export。
- `usage-details-dark.png`：两端四行Models breakdown及四项数值、pagination 1/1、Records 4sessions顶部均在；自然高度改善成立（底部Records均约y865开始），不是记录明细全部可见。窄列标题两端都有截断，不把共同现象单列为Go新缺陷。

## 987 与后续源码必须区分

本图 **不含** 发布的 wave3 `41ba2ad`。已直接对照 `git diff --name-only 987eacb 41ba2ad`：后续包含 embedded Git diff横向滚动/双层sticky、context方向展开、Git/GitHub菜单owner保护，以及chat导航/重试等；应在最新SHA再捕并用相应展开/滚动/菜单recipe验证。不能用WB06折叠静态图反证这些补丁，也不能因为源码/unit绿就把图上差异勾掉。

`internal/ui/files.go`、terminal与基本Git外壳布局不在该范围的修复清单内；WB03/04行数/标题和WB06/07外壳差异依旧是明确待办，除非更新source另证。WB05 review gap后续修复不是本图可见的普通diff gutter/文件头视觉对齐。

## 下一轮最小验收

优先核对 WB03/04 trailing newline 的行数摘要定义并保持原字节/可编辑空末行；不为凑5行删除文件末尾换行。随后最新SHA复拍WB03/04。独立增加Review wrap-off横滚/vertical scrollbar、expanded Git双sticky/context、真实terminal退出/restart及tabs操作的有状态recipe；现有OCR_READY只负责采集准备，不能替代这些证据。
