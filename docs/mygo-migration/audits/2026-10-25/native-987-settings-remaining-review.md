# 987 剩余基础原图实际复核 — 2026-10-25

## 范围

继 `settings-five-actual-review.md` 后，逐张 `view_image` 看了同根目录下两端 **20张原图**：
`settings-{about,models,mcp,plugins,privacy,report-bug,shortcuts,skills,worktrees}-dark.png` 和 `new-task-dark.png`。
因此本批是 **9个Settings首屏+1个新任务首屏**，不是10个新增Settings状态。
Go `987eacb`、Rust `299a32b`，原始1470×922；未裁/拼/改图，没有额外原生输入或CI操作。
这里只记录看得到的内容，不能由空态、按钮存在或OCR_READY推断功能/像素验收通过。

| 原图 | 具体可见结果与差异 | 限制 / 后续源码关系 |
|---|---|---|
| About | Star链接、Orbit Pi、pi CLI Connected、Source/GitHub均在。Go把GPUI0.2.2行替换为MyGo9c69cf829ece及Go1.27.3，应用版本含-go-dev；卡片主要边界同位 | 真实迁移栈/版本标识差异，不应伪装成Rust。没有点击外链。`settings_misc.go:111+`仍对应当前实现。 |
| Models | 两端均 `No models reported by the runtime`，有search、Update model list、Favorites和同样空态说明；没有可见模型行 | 不证明模型清单缺失bug或模型功能通过；无凭据fixture下当前只是空态，未Refresh/收藏/选模型。空卡下界约Rust421、Go420，不夸大为功能漏项。 |
| MCP | 两端同agent/project路径、Checked1sago、空server提示和两个Add MCP Server入口；无错误/Loading；空态位置接近 | 未添加/探测/启停server。后续wave4改了manager生命周期、startup probe、quota pane，**这张987空态图不覆盖这些路径**；不能说旧图证实修复。 |
| Plugins | 两端search、npm/git/path输入、Global/Project、Install、Refresh、0package和No plugins installed齐全 | Rust Global文字/选中表现较亮，Go Global看上去更灰；仅记录选中对比度，不能由截图认定Go选错作用域。未安装或刷新，插件加载/restart行为未验证。 |
| Privacy | 自动清理off、Archive、30days、Clean up now、临时目录和Open folder齐全。两端显示3items，但Rust4742.1MB，Go82.7MB；Rust含orbit-reference4713.4MB目录，Go含orbit-go-native42.0MB文件，fixture分别1.4/13.4MB | **环境内容不相同**，不能用这组MB差异判定扫描/清理实现错误。AI scratch files只有标题/说明开头，下面未看；没有执行任何删除、打开目录或retention操作。 |
| Report bug | 类型切换、Context+Generate draft、Title、What happened、Steps、Add screenshot、Clear、Create issue均在且表单主结构接近。Go首尾文案改为review-before-sharing，底栏 `Diagnostics are off… Privacy…`；Rust称版本/OS/pi信息自动附加 | `docs/mygo-migration/bug-report-source-reconstruction.md`明确这些是保留的review/opt-in/redaction偏差，不是漏文案。没有生成draft/上传/提交；不能确认Privacy入口和review modal交互。本图未发现整项表单消失。 |
| Shortcuts | 当前首屏两端同Session六行、Agent三行、Composer六行至Stop Run，键名均可见：包括Ctrl+1–9、Ctrl+Tab、Ctrl+Shift+M/T、Alt+Enter/Ctrl+Shift+Enter等 | keycap宽度/间距不同；仅标签对照，不代表快捷键真的触发。下方滚动内容未验。`settings_shortcuts.go`987→wave4无变化。 |
| Skills | 两端三列结构、All skills0、No skills match、Select a skill、底栏0skills一致；未选任何skill | 只有空态；无法核安装路径、SKILL.md渲染/markdown、禁用/启用、搜索结果行为。列表/详情本体不能因空卡形状接近判PASS。 |
| Worktrees | 管理按钮、.wt、.orbit/worktree-setup.sh、run/open两个on开关均在。Go Default location说明是 `Relative to the repository root`，Rust是 `Relative to workspace` | 是具体文案/路径语义差异，需与当前effective-root语义一起验；现验收规格W40–W43明确主仓库根路径，不可只改文字掩盖行为。不执行setup脚本、不打开worktree；设置页源码987→wave4无变化。 |
| New task | 两端No projects yet、Connected、Start a task、相同workspace路径、Build、unknown model、composer等均在。Go中心workspace卡约比Rust低5px；composer顶约Go791/Rust782，下界约884/878，Go更短；不只是字体差异 | 这张未点workspace/附件/模式/模型/发送。后续wave4改chat/menu/attachment/owner/sidebar等，因此必须新SHA重新捕捉匹配状态；不能将该旧图直接当wave4表现，也未证实旧几何差已修。背景点阵相位/强度也不同，未做逐像素分析。 |

## 源码核对与优先项

- 只读diff `987eacb`→冻结wave4：`internal/ui/settings_misc.go`、`settings_shortcuts.go`、`internal/resourceui`、`internal/bugreportui`未变；MCP确有startup/state/quota相关增量。不把已改但未在图中呈现的流程勾成已验收。
- Report bug保留隐私偏差和About真实渲染栈标识应留在验收说明，不追求伪造相同文字。
- 最有价值的新原生状态不是重复空页：Models真实但无凭据的catalog/收藏fixture、Skills有真实本地SKILL.md、MCP可控无凭据server配置、Plugins可控本地包、Worktree无副作用配置编辑/取消。另行审批捕获recipe，本轮不扩动作。
- Privacy下一轮应先提供两端完全相同且明确可清理的temp fixture，再比较统计；当前runner构建目录不同，无法做相同输入的计数验收。
- New task目前可见几何偏差需在最新SHA原图核对；workspace picker卡住的用户报告仍不能由该未点击首屏证明修复。

本批无新增源码修改、CI触发、登录或原生动作。连同前五对，共人工看完此artifact的14个Settings首屏和1个New task；这不是14个功能完成或像素一致的声明。
