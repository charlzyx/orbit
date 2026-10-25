# 持续迁移第二波：独立整合中

基于持续迁移第一波的冻结源码。**未发布安装包，不改变 `.4`；本页不是整体验收 PASS。** 独立补丁审查/专项通过之后，仍须同树 aggregate 与原生验证。

## 已进入本整合树

1. **无窗口后台维护**：真正 owner-thread 回调处理完成队列、设置/MCP 应用完成、快照与过期提示；Pi runtime drain 仍在 timer worker。Controller mailbox 改为不阻塞 owner 的队列、每批 256 条，明确无界队列内存与已有 App.Async shutdown 等待边界。
2. **保留的 Settings→MCP 页面**：不用“最近 2 秒画过”代替页面状态；保留 60 秒外部变更轮询，切走页面停止新轮询但消费已有完成；外部修改仍须手动 Apply。
3. **Context usage**：环形占用/hover/详情与手动 Compact 接线，真实 RPC 去重、拒绝/超时未知状态、旧 session 响应隔离。缺失/错误 token bucket 显示未知而不编成 0，部分字段可单独显示；这比 Rust 缺失 bucket 默认 0 更保守。尚未获得 session ID 时禁用 Compact，ID 到达后可用。Rust 确实允许 streaming/retry/awaiting 中手动 compact，按源码保留。
4. **Retry Cancel 呈现**：点击即隐藏当前事件的状态条/设置项，真实 Manager retry/busy 不被伪造为空闲；后续相同 attempt/error 的新 retry revision 重新显示。旧 session action 不得隐藏新会话。耗尽错误本地化与源取消提示位置另在后续增量。
5. **Explorer 行内操作**：顶部新建行、原行重命名，非法/重名输入保留并可修正、Escape/点击取消、重复提交/dirty guard、工作区 epoch 及旧完成不抢焦点。非 Explorer 通用 Prompt 不变。现有 filesystem rename 与 editor remap 之间的 autosave 原子性需独立后续审查。
6. **附件与 Usage 导出对话框**：异步 native 等待、UI-owned 完成，附件 owner 用稳定 opaque runtime generation + intent/window，正常空 ID→hydrated ID 不误丢图；文件读取开始前即捕获 owner。Usage 导出先冻结 index/filter/options，取消/关闭不写文件，保留原生覆盖确认。
7. **背景图选择**：源 Rust 异步对话框；close 与背景 generation/reset 隔离，不把全局外观偏好错误绑定某一聊天会话。

## 尚未声称完成

Session HTML 导出仍保留原同步路径：简单“先检查 target 再 Send”存在同 client `new_session`/fork/clone 重定向竞态，超时也不能证明服务端命令结束，安全序列化另审。Pi binary 等其他同步对话框未被一并改写。真实原生确认、权限提示/覆盖确认交互、模型推理、逐状态像素对照仍未覆盖。

本整合不含正在独立审查的 session 删除/路径别名保护、embedded Git diff 两层固定标题及 quota 隐藏/MCP tab。不要将这些 pending 工作混入此波通过范围。

## 最后整合门禁（17:19 UTC 冻结）

为避免重复 native CI，已审的第一、第二波合为一次最终源码冻结；[机器可核对的源文件清单](continued-wave12-evidence.json)记录固定基线、57 个变化源码文件及 SHA-256。第三波源码未进入此冻结。

整合测试曾暴露 Controller mailbox 替换后旧测试仍只消费 App 队列的问题。已改用真实 owner drain，并增加实际 `Controller.View` 消费异步选择结果的回归；workspace/attachment/usage/background 三次 race 通过。原“捕获队列回调后关闭应用，再消费应不修改状态”的断言保留。

另一轮 terminal 测试在画帧后才观察到 worker 已完成的错误状态，随后误把旧帧当新状态检查。确定性新测试在旧 helper 上失败；明确补画已观察到的完成帧后，所有 TerminalHost 测试 race 十次通过，原 persistent error 文本断言不改。这是测试时序修正，不是声称修了新终端运行时故障。

早期第二波 aggregate 已通过（完整 UI 363.103 秒），但之后加入 Review 标题滚轮修正和上述 test barrier，所以它不作为最后源码的通过证明。最后同树 `vet ./...`、`test -race ./...`（完整 UI 314.623 秒）与无 cgo 的 darwin/arm64 跨编译于 17:25 UTC 全部通过，命令链退出 0。源文件 fingerprint 与 17:19 冻结一致。feature 分支发布后的同 SHA CI/native 仍待验证，不能把本地全绿等同原生验收。
