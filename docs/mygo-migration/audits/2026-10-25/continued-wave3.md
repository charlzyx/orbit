# 持续迁移第三波：冻结范围与验证边界

本批源冻结于 2026-10-25 17:59 UTC，基于已公开的 `305a9ef617934929e3db6c072e6bd297a73ed602`（native 初始化/Usage readiness 修正），**未生成新安装包，不改变 `.4`**。Rust 参照仍为 `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`。

## 冻结范围

| 原始 scope | 实际源码变化 | 关键限制 |
|---|---|---|
| S10/S11 | source session 菜单、复制 ID/path/reveal、状态化 pin/archive；canonical/SameFile 归属、metadata/delete reservation、精确 child lease 与 Restart 清理顺序 | 仅临时数据测试；外部任意文件系统替换不是内部 fence 的原子保证；Mac 大小写别名未验 |
| Session HTML | 异步 Save 捕获 task，实际 RPC 完成/child 退出前持有 lease，超时只报告而不假装取消；已知错误 command 响应仍保持不确定性保护 | 原有 Go 导出入口的安全修正，不虚构 Rust 独立入口；扩展自行改变 child 会话不受宿主命令序列化完全控制 |
| W08/W17/W35/W36 | embedded Git diff 长行/横滚/两级 sticky/方向与 Shift context；菜单 repo/dataset 归属；File history 延后到下一帧导航；dock gap 两个按钮精确 16+16px | mounted 几何不是全原生像素验收；原来文件 rename/editor autosave 原子性问题不在本批 |
| C14 | typed retry failure 本地化、普通错误不靠猜字符串分类、取消反馈 6 秒 TTL | 真实 provider 错误/重试流仍需实际输入验证 |
| C08 | source 消息导航轨道、预览、hint、正确 Prev/Next；仅 rail opt-in 保留当前 scope 选择；hint 持久化移出 UI 帧 | retained selection 补丁保持普通 control/outside clearing；跨窗口句柄拒绝；原生 paired Chat capture 尚未完成 |
| P09/P14/P17 | 隐藏供应商分组、真实 MCP tab、精确配置加载后 3 秒 startup probe、binary 切换清旧 intent/状态 | 不伪造 connected；新 binary 不继承旧自动 probe；真实 OAuth/MCP endpoints 未测 |
| S13 | workspace copy/remove/icon/color、状态 marker、shared Store revision 防同秒 remove/re-add ABA | Remove 只移出侧栏；Worktrees 按 Rust 打开当前 workspace，不偷偷切到后台 header 的目录 |

## 组合验证策略与诊断记录

独立补丁由不同源码副本交付，按冻结 hash 整合，原始源码保留。Manager 合并明确保留 `manualCompactPending` busy、compaction nonce、typed ConnectionError setter；MCP 合并保留 host visibility 维护。没有重复套用早期 export agent patch 覆盖最终 lease 改动。

组合第一次 core gate 的 `TestStoppedMetadataCompletionCannotPolluteReplacementTask` 在**初次连接**阶段被测试的 50ms RPC timeout 打断，还没进入 held metadata 场景。仅测试改用正常 bootstrap 预算；专门的 past-timeout 场景改等待真实 `lease.notified` 而非 sleep。原 held rename、Restart、清理、旧目标错误不污染新任务断言保留。随后组合 agent race 155.244s、MCP race 4.494s 通过。

第一次组合 UI gate 的新 Session Export 测试仍 drain 旧 App queue，而生产 completion 已由 Controller mailbox 持有，因此 5 个非 Close case 等不到回调。测试改用已有真实 Controller drain，另加实际 `Controller.View` 帧完成导出的回归；原 confirm/cancel/error/new-task/window/close 断言保留，专项 race 3 次 32.153s 通过。没有通过改生产 timeout、跳断言或放弃关闭/取消保护来“修绿”。

C08 新 overlay 在整合树通过原 composer/selection toolkit race（1.431s）及 canonical wrapper 的 21 个边界测试；其错误跨窗口句柄与 hint 同步文件锁阻塞均先有真实红测再修正。组合 C08/导出重点回归 23.289s 通过。原始 raw diff patch 的上下文空白是补丁格式，不重写 hash-listed evidence 来迎合通用 whitespace checker。

**最终同树 aggregate/vet/arm64 与 exact-commit CI/native 必须另记终态。** 上述局部与中间通过不能替代最终门禁，既有失败日志也不删除。

## 独立第四波，不计入这里

C04 caret 文件插入、C09 有界 HTTP/data raster 图片、S05 surface focus traversal、S12 workspace sticky headers 正在隔离实现/验证。新的 native Chat 只读 fixture/capture 也单独审查。它们不会悄悄进入本冻结批次或已有 `.4` 安装包。

## 追踪入口

- [会话 lease 与所有权证据](session-file-lease-evidence.md)
- [workspace header 行为](workspace-header-actions.md)
- [Chat 导航与选择证据](chat-navigation-evidence.md)
- [embedded Review](embedded-review-scroll-sticky.md) / [菜单 owner](git-menu-workspace-identity.md)
- [剩余 scope 清单](remaining-scope-1739.md)

139 scope 是原版核对目录，不是 139 PASS；25 对旧图不是全状态像素接受。没有声称用户 Mac 的 picker 无响应已被复现或修复。

### 18:07 UTC 最终本地同树门禁

冻结源码 fingerprint `f25b36547b889a527da59c438dc9ed5bada6b8a36823102223781ad1d3ed53a3`（Go 源/测试、selection overlay、go.mod/go.sum 和 canonical wrapper 共 471 个文件）复验未变化。全仓 vet、aggregate race（UI 389.957s；其他包亦通过）、no-cgo darwin/arm64 build 顺序链 exit 0。后续只采用已发布 `987eacb` 的 harness metadata 修正，未改变上述 Go/overlay 字节。此结果不是远端 CI 或 native paired 全量验收；公开提交后的 exact SHA 结果另行跟踪。
