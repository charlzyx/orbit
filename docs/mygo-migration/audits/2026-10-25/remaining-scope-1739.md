# 持续迁移剩余范围（17:39 UTC 检查点）

Rust 基线仍为 `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`；用户安装包仍为 `.4` / `0af37538`，没有新发版。下列“已补”仅表示开发源码中的**特定子行为**，不是整个 scope 或完整 1:1 验收。

## 开发分支已补的范围

`ce4d49a828048aafb9aba006b1e3f2d59c011ca4` 的源文件哈希和同树 vet/race/arm64 编译见 [wave12 证据](continued-wave12-evidence.json)。

| Scope | 已补子行为 | 尚未由此证明 |
|---|---|---|
| C14 / C15 | retry/compaction 状态条、重试取消呈现、Context ring/details/Compact 与缺值区分 | 全部真实提供商事件；取消尾部状态和 exhausted error 本地化仍在后续批次 |
| W35 / W36 | dock Review 长行横滚、方向/Shift 上下文展开、sticky headers、可访问文本和 wheel 路由 | embedded Git 页面同等功能及完整原生像素对照 |
| W54 / W55 | Explorer 菜单动作、行内创建/重命名、重复与过期 owner 保护 | 原子文件系统 rename/editor autosave 竞争、所有原生焦点流程 |
| P17 | MCP Apply 已有前序 ACK 修正；本批补 parked runtime 配置指纹失效重启 | 真实 MCP 认证/端点、顶部 MCP tab 和精确启动 probe 时序 |
| S01 / S07 | windowless host 维护、工作区异步 native wait 和旧意图保护 | 用户 Mac 无响应根因和真实文件夹确认成功 |
| C05 / U04 / P19 | 附件初始读取 owner、附件/Usage 导出/背景图异步 native 对话框 | 全部真实拖放/粘贴、native Save overwrite/确认验收；Session HTML 导出另批 |

前序开发分支还补了 History 布局、Explorer 调宽/键盘、Usage Details 几何、手动清理入口和终端退出后的 Restart；证据分别在 [parity integration](parity-integration.md)、[functional integration](functional-integration.md)。这些同样不改变 `.4`。

## 正在集成、尚未作为同树最终验证或新安装包交付

- **S10 / S11**：会话菜单 path/reveal/state labels；删除与 metadata 的 canonical/SameFile 归属、运行中保护、精确 child lease、Restart cleanup handshake。只使用临时数据，外部任意文件替换不是内部 reservation 可以消除的竞态；Mac 大小写别名仍待验证。
- **Session HTML 导出**：捕获 task、异步 Save、真实 RPC 完成前保留 lease，避免同 child 新任务重定向；超时不是远端取消。Pi 1.4.0 临时 HOME 实测 HTML 内容有独立证据，不代表用户真实 session 已验收。
- **W08 / W17 / W35 / W36**：embedded diff 横滚、两级 sticky、方向/Shift gap；菜单绑定 repo/dataset identity；File history 从构建帧延迟导航，避免构建中改 tab。
- **P09 / P14 / P17**：隐藏供应商分组、真实 MCP 状态 tab、旧 binary probe intent 清理；启动 probe 时序及 binary 状态失效正在独立校正。
- **C14**：typed exhausted-retry error 所有权、本地化和 6 秒取消提示，避免猜测普通错误字符串。
- **C08**：消息导航轨道/hover preview/提示持久化和正确 Prev/Next turn；选择保留已发现 toolkit 边界，单独 opt-in 修正，不能把红测删掉算通过。
- **S13**：工作区菜单 copy/remove/icon/color 正在源对照实现；Remove 只移出应用侧栏，不删除目录或会话。S12 sticky 不顺带算入。

## 已确认但尚未关闭的源码差异

以下定位沿用 [Chat 原始核对](chat.md)、[Shell/Settings 原始核对](settings-lifecycle.md)，这些原始表格描述的是 `.4`，不覆盖后续源码状态。

| Scope | 剩余子行为 | Rust 定位线索 |
|---|---|---|
| C04 | 附件路径在 caret 插入而非追加；plus-menu subset | app/composer_ops.rs:9,113,471 |
| C09 | Markdown HTTP/data 图片；完整 table/list/code 布局 | transcript_view.rs:5201,5839,6177,6304,6537 |
| P03 | 独立 first-run/onboarding 与 Pi migration UI | app/view.rs:2664,2903 |
| P20 | 余下硬编码文字、完整语言/菜单一致性 | main.rs:626; theme.rs:1468 |
| S05 | 显式 surface traversal 对照，不能以默认 Tab 代替 | app/view.rs:3775 |
| S12 | workspace 分组 sticky header/marker 完整行为 | app/sidebar.rs:176,370 |
| S13 | 工作区图标/颜色/移除/复制路径的全交互（正在实现） | app/sidebar.rs:2373,2383 |

## 验证与部署剩余项（不能冒充代码缺失或代码已完成）

1. **当前 native gate 有真实失败，未标 PASS**：ce4 的 hosted Linux run `40216553377` 中 reference 和 aggregate 成功，native job 在 Details 的唯一 exact OCR 目标处停止，未发 click。原图是 `Loading local usage…`，可见 Details 但 OCR 未识别；artifact SHA256 `0eaaf126c26f656ffd8dde343c7bb0ec2bc4bffb3f002418834549a60c3fcd81`。正在修正 native 初始化前启动 heartbeat 的竞态，以及 capture readiness 条件，必须重新采集最终源的真实帧。
2. **S07 文件夹确认/用户无响应**：旧 `.4` AX 只有 11 nodes / 0 confirm buttons，没有发送确认。取消与 recent A→B→A 成功不能替代确认，源码异步修正也不是已复现用户问题。
3. **C01/C05/C10**：实际 Mac IME、拖放/粘贴、剪贴板/选择、window close/reopen，仍需按真实状态验证。
4. **真实 provider/auth/extensions**：未做真实模型推理、用户 OAuth/密钥/端点验证；fixture 的成功不扩展为真实账号通过。
5. **U08/U09/U14**：Go signed update feed/key、签名/公证、真实安装升级回滚未交付；不可以 Rust feed 替代。
6. **全量视觉验收**：既有 25 对图不是全状态覆盖，139 scope 是追踪目录，不是 139 PASS；loading/error/empty/busy/repeated/cancel 状态和跨 DPI/theme 仍需对照。

优先顺序是阻塞或错误目标的行为安全 → 可用功能缺口 → 原生同状态对照 → 几何细节；不编造百分比或“剩 N 个就全完成”。
