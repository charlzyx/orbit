# 验证清单与剩余范围（19:59 UTC）

**安装包仍是 `.4`。** 以下开发源码改进并不在该安装包里；只有重新发布且明确告知后才能按新版验收。当前已公开开发源码为 `918805045b7e9c77ab252f72a70652cfb8438920`，其最终本地 vet、全仓 race、arm64 编译通过；同 SHA migration run `40228444291` 的四项 job（proc、aggregate、Go native、Rust reference native）均通过；这仍不代表所有状态已逐张验收。Rust 基线仍为 `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`。

## 明天优先验证（按风险/日常行为）

1. **工作区切换与窗口生命周期**：打开、取消文件夹选择；最近目录 A→B→A；关闭窗口后重新打开。记录是否无响应及日志时间。开发版已把对话框等待移出 UI 构建，但没有证明用户 Mac 的原问题已修复，真实确认按钮仍未完成自动验收。
2. **会话/草稿不串台**：切任务、重开、New Task 被拒时草稿保留；旧窗口/旧菜单不能操作新任务。开发版 New Task 仅在宿主接受意图后清草稿，接受不等于 RPC 成功。删除、重命名只在一次性测试会话中验证，并仔细看确认目标。
3. **输入/附件**：中文 IME、Shift+Enter、光标中间插文件、每个路径单独 Undo、Mention file、图片菜单取消。开发版已有 caret 队列、六格式单图选择器与菜单焦点恢复；物理 macOS IME/文件对话框仍需实测。
4. **聊天内容与状态**：切任务后不能显示旧图片/预览；retry/compact/context 缺值不应伪装为零；长消息滚动、Prev/Next、选择、搜索、复制。真实模型/扩展事件与完整原生状态尚未验收。
5. **日常项目操作**：终端各 tab 单独关闭，右上关闭只隐藏并保留进程；Explorer 创建/重命名；Review 长行横滚、上下文展开、左右列及 sticky header。开发版已有对应窄修复，整体 Git 页面和字体/像素仍有差异。
6. **设置/启动**：缺依赖时主区显示 setup，侧栏保持可用；语言切换、Reset、Appearance 键盘分组。已公开版本只有只读依赖提示；未暴露安装器，未执行任何真实迁移/授权操作。

## 已公开的局部改进，不等于整个 Scope 完成

| Scope | 已公开子行为 | 仍需区分的缺口 |
|---|---|---|
| C04/C05 | caret 路径、Mention、六格式单图、菜单焦点、附件 owner、异步对话框 | 真实拖放/剪贴板/IME/确认和完整混合内容行为 |
| C08 | 消息导航轨道、预览、选择保留窄 overlay | 静态 Chat 两态捕获仍被身份门禁阻止，物理选择未验收 |
| C09 | 有边界的网络/data raster、owner 缓存/预览隔离 | 同 slot 内容修订下一批；SVG/private hosts 有意不支持，table/list/code 尚在源核对；不是完整 C09 |
| C14/C15 | typed retry、compaction、Context details/缺值、目标归属 | 真实提供商事件、真实取消/压缩、状态全覆盖未验收 |
| P03 | 只读依赖 setup、exact-owner 恢复、main-only 布局 | npm 确认迁移 UI 下一批；任何真实安装/迁移均未执行 |
| P17 | MCP ACK、parked 指纹、启动 probe/状态 tab | 真实端点/认证与扩展配置组合未验收 |
| P20/S05 | 一部分可见文字、Reset localized label、surface traversal/Appearance focus | 其余 accessible names、菜单/全页语言、原生键盘/AX 未完整覆盖 |
| S10/S11/S12/S13 | 目标绑定 metadata/export、删除 reservation、workspace 菜单/sticky | palette Delete/Pin 下一批；外部文件替换竞态和 Mac 大小写别名残余 |
| W35/W36/W54/W55 | Review/embedded diff pan/sticky/context、Explorer inline/menus | Git 外层结构、字形/换行/像素、filesystem rename/autosave 原子性残余 |

## 正在下一批，尚未公开

Palette Delete 确认固定目标、Pin 当前路径、inline image 内容 revision、Review 文件标题/计数、npm-only setup 显式确认卡片。每项有独立 source review 与临时 fixture 回归，仍需组合全仓门禁后才能算已发布源码。

## 原生证据边界

- `a5c40cc2` 的正常 migration run `40226143789`：proc、aggregate、Go native、Rust native 全通过。
- 第二次静态 Chat opt-in `3d170d83` / `40227121094`：aggregate/proc 通过，两端 STARTUP 因未识别 pinned Pi 子进程拒绝输入，**0 输入/0 Chat case**。自有离线 Pi 已证实成功 RPC 子进程的 argv 被改成 `pi`，不满足原 exact-argv 接受条件；本地还缺 `/proc/task/children`，因此 scanner 空列表不能唯一归因于标题。原 CI 未保存被拒 argv，完整 ancestry/receipt 链仍待 hosted 验证；不把局部复现当成全部 CI 原因证明，也不改成按进程名放行。
- 已实际查看的 25 对旧界面不是 25 项像素 PASS，更不是完整功能数量。原始 139 条 scope/490 个候选断言只是目录，不是完成率。

历史 `.4` 审计请看 [chat.md](chat.md)、[settings-lifecycle.md](settings-lifecycle.md)；它们的“缺失”描述不能覆盖后续修复。本页也不是所有未知差异的穷尽列表。
