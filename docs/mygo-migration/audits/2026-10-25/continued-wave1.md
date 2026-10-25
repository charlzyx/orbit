# 持续迁移第一波：源码与验证边界

固定 Rust：`299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`。开发分支基线：`5e4e1425ff94346066fa35c95d005e904e4b5437`。这是未发版的迁移源码，不改变已下载的 `.4`。

## 本波行为

- [MCP parked 配置](../../mcp-parked-fingerprint.md)：读取目标项目/全局配置原始字节，闲置重开时异步确认旧进程配置；忙碌会话不被换掉，旧完成不能抢新选择。发送前再次核对客户端身份。
- [工作区选择](workspace-dialog-async.md)：按 Rust 的异步原生 sheet 设计把 native 等待移出 View，UI owner 消费结果；重复点击和旧会话/窗口/工作区意图被隔离。不是用户无响应已复现或已修复的声明。
- [Review](review-long-lines.md)、[上下文与固定标题](review-context-sticky.md)：长行尾部可横向访问，unified/split 保留竖向虚拟化与滚动锚点，方向/Shift 上下文展开、真实 sticky 标题、可视右缘竖向滚动条。独立 embedded Git renderer 不在此波。
- [Chat 状态条](chat-status-strip.md)：重试细节及压缩状态、实际重试取消 RPC、中文默认错误。取消乐观呈现、耗尽错误本地化仍有差异，不能称 C14 全闭环。
- [Explorer 菜单](explorer-menu-actions.md)：目录展开/收起与默认应用打开、旧工作区菜单回调失效。行内创建/重命名属于下一波。

## 真实发现，而非修改测试掩盖

Review 的首个 sticky 实现让 MyGo 自动加 header 高度，破坏旧的行锚点和 hunk 跳转；撤回该用法并用原生独立标题层后，原断言不变通过。竖向滚动条在长行横滚下移出可视区的问题，通过真实右缘拖动红绿验证修正。Explorer A→B→A 后旧菜单回调复活有独立失败回归；新增工作区 menu epoch 修复。

## 门禁记录

初步整合树（未含最后 Review sticky 增量）的全 vet、agent race（75.523 秒）、完整 UI race（308.742 秒）通过。Review 作者最后冻结源码的全部 Review race（193.161 秒）与 vet 通过。最后 Review 增量已于 16:58 UTC 整合，**最终同树 aggregate/native 门禁仍运行中**，不能把之前结果冒充最后源码通过。

兼容层 wrapper 的 21 项 Python 回归通过；workbench fixture 52 项执行结果通过、其中 2 项跳过，跳过不计验收。Composer/selection 工具包专项通过。完整结果会在同一证据页更新。

云端 mounted/JSONL fixture 与跨编译都不能替代真实 Mac 文件夹确认、真实模型推理和像素级原生对照。旧 `.4` 确认探测仍为 BLOCKED；25 对旧截图不代表本波新状态或全部功能已覆盖。原生确认、完整迁移和视觉 1:1 均未完成。

## 合并后的最终本地门禁

17:19 UTC 将已审第一、第二波冻结为一个源码树，包含 Review a11y/标题滚轮修正及 terminal 测试帧同步回归。17:25 UTC 最后全 vet、aggregate race（完整 UI 314.623 秒）、darwin/arm64 跨编译全部通过；见[同树证据](continued-wave12-evidence.json)。之前的失败/中间通过仍是诊断历史，不冒充最后门禁。尚待同提交远端 CI/native 与实图比较；未发新版安装包。
