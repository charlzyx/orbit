# 987 Settings 五对原图实际复核 — 2026-10-25

## 证据范围

逐张 `view_image` 查看以下十张原 PNG，未裁切、缩放保存或拼图：
`{go/go-desktop,rust/reference-desktop}/settings-{appearance,agent,runtime,providers,general}-dark.png`。
根目录 `/workspace/shared/orbit-native-987-1846`。
两份 manifest 均记 clean source、1470×922：Go `987eacb79805b46a2152fc5003f8eb7bc59d0598`，Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`。
这是五个首屏的静态人工观察；未点击这些图中按钮，未确认滚动区下方、持久化、重启、provider认证或后台通知。非像素 PASS、非全部 state PASS。

| 原图 | 两端实际看见 | 差异与判定 |
|---|---|---|
| Appearance | Appearance/Dark、Orbit Azure、背景选择、English、双预览、IBM Plex Sans/Lilex、14/13/13px、100% 均可见；主内容左边与卡片边界基本同位 | Rust Light/Dark/System 组左缘约1129，Go约1138，组宽不同；正文/终端预览字宽与字形仍不同，部分字体下拉宽度有差别。不能因标签相同就断言实际 glyph face 相同；未打开字体菜单或验证修改后持久化。 |
| Agent | 同样的行为、session titles、session defaults三组；Queue follow-up / One at a time、auto-compaction、自动title、pi default均在 | **Auto-retry状态不同**：Rust为蓝色on开关；Go标题旁显示 `Unknown: this pi version has no getter`，右侧为下拉。Go不是缺少整行；这是显式unknown状态表达差异。其余卡片边界接近，文字宽度/微小边界仍不同。未从静态图推断Enable/Disable是否发送成功。 |
| Runtime | 两端显示Running，Pi路径、stdio、RPC patches on、package root、No output均在；Restart/Stop可见 | PID Rust4622 / Go6121，uptime12s /10s：跨进程/抓拍时间的volatile值，不计功能错误。主卡片底边约Rust501 / Go500，字号/字形/按钮边界略异。没有执行Stop/Restart，不能证明生命周期正确。 |
| Providers | 两端均0 active /0 connected、search、Refresh、Add provider以及真实provider卡片 | **内容导致整格错位**：Rust44 providers，首卡 `.manifest` /42 in catalog；Go43，首卡Amazon Bedrock，无 `.manifest`。Rust首排manifest/Bedrock/Ant Ling，Go首排Bedrock/Ant Ling/Anthropic，因此带OAuth多行的Anthropic把Go首排行高拉大；不能全归因layout bug。部分长endpoint说明截断样式不同（Go可见ellipsis，Rust原图文字在边缘截断）。未做登录、添加或刷新。 |
| General | 两端Connected、相同session store/workspace路径；三项notifications开关均on；两主卡片和分组齐全 | 文本字形/字宽和约1px的通知卡起点/分隔边界不同；无明显整组缺失。侧栏底部版本Rust `v0.2.12`、Go `v0.2.12-go-dev`属于构建身份差异。未实际触发通知、后台声音或开关存储。 |

共同：主内容约x399、右界约1311、侧栏分隔约239基本一致，不等于精确像素相等。侧栏若干图标轮廓（例如Privacy/Runtime）与文字栅格不同。上述近似坐标是看原图的几何观察，不是逐像素测量结果。

## 与后续 wave4 / wave5 的区别

只读对比 `987eacb` 到当前冻结 wave4 的 `settings_{appearance,general,providers,agent}.go`、`runtime_view.go` 与 `internal/workbench/providers.go`：无文件差异。因此不能把本批五页可见差异称为已被 wave4 修复。

- `internal/workbench/providers.go:321` 明确跳过 `.manifest.json`；`provider_catalog_gap_test.go` 防止其变成provider。保持真实catalog，不为凑44添加元数据伪provider；后续严格视觉比较应把两端输入的这一源行为差异明确记录。
- `internal/ui/settings_agent.go:100–135` 注明D04保留未观察到的值unknown，只有acknowledged值渲染on/off；不应为截图变蓝而伪造runtime getter结果。
- 当前未冻结 wave5 `settings_agent.go` diff仅把unknown/unsupported/updating/enabled/disabled与菜单动作本地化。它不消除这个英文原图里的unknown-vs-on语义差异，也不是新native抓拍证据。
- wave4的chat/context/rail/attachment等修正不在这五个首屏可见范围；不能借这些图验证，也不能用旧图反证修正。

## 下一步窄验收

1. 同一最新SHA保持实际model/auth缺省状态复拍，尤其记录retry unknown是预期状态而不是编造on。
2. Providers先隔离catalog内容差异再评布局；单独核真实provider卡的文字溢出/ellipsis，不恢复伪manifest。
3. Appearance核解析后的字体文件/字体fallback与控件宽度；不要以修改字号/缩放图片掩盖栅格差异。
4. 交互验收另用真实输入记录：字体/语言选择及重开、retry请求ack/error、Runtime Stop/Restart、通知开关持久化。此轮截图未执行这些操作。
