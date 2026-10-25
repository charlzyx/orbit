# 第四波：已整合源码与尚未接受的范围

基于第三波最终树 `2f3dc3c20e496506de8ccd150fc0708f5c913c78`（公开提交 `41ba2ad5d9763b9bde8627f1d7d8293d425390ce`）。`.4` 安装包不变，无 main 合并。Rust 参照仍为 `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`。

## 有界修正

- **C04**：非图片路径按 Rust 在 active caret 插入，不替换选择，每路径独立 undo；IME 延迟、旧 task/隐藏 composer 队列丢弃。Mention file 插入 `@` 并触发补全；单图片选择器及回调均限定原版六种后缀，最多八张。原生菜单打开时恢复 composer 焦点。真实 OS 菜单/输入法/选择确认未因此算通过。
- **C09（部分）**：有界 HTTP/data 栅格图片、显式不支持状态、每跳/实际拨号地址校验，无 ambient proxy/cookie/auth。workspace 服务入口先失效旧 owner，修复 A→B→A 的旧图片点击和预览滞留。SVG、私有主机、动画仍不支持；不是完整 Rust 图片能力。
- **S12**：expanded workspace sticky header 在原 List children 内保持滚动/row key；下一 header 推离、唯一可交互副本、resize grip 与 scrollbar hit gutter 分离。
- **S05（部分）**：chat/files 已知 surface 的 Tab/ShiftTab 遍历，保留 live terminal/editor/autocomplete 自己的按键语义。退出 terminal 改为可操作 Restart shell，后台退出不抢焦点。此退出交接是 Go 可访问性修正，不冒称 Rust 已有同样交接；其他页面仍待核对。
- **静态 Chat 取证**：默认关闭的手动输入，仅同一 synthetic persisted session 的 Tail / PrevTurn 两状态。Context 因无认证 Pi 返回 unknown model 而明确 BLOCKED；没有 fake key、set_model、推理、猜坐标或 hover/selection 扩权。

## 整合与验证

独立补丁均经 scoped review。app.go 手工合并仅保留双方字段/owner hooks：caret prune 位于 mailbox drain 后；C09 owner reset 位于 workspace Model 改写前；S12 List hook 保留第三波 session-row 菜单。S05 与 C09 不替换第三波 manager、MCP 或 selection overlay。

组合重点 UI race 35.689s 通过；此筛选没有覆盖 loader/native 无匹配用例的包，不能称那些包被该次测试验证。最终同树全仓 vet、aggregate race（agent 154.518s、UI 451.939s，其余包亦通过）、darwin/arm64 build 顺序链 EXIT 0。Go 源文件 hash 与门禁启动时一致。静态 Chat harness 21 项测试包括实际 Capture metadata 链与 own-process-group 清理，不代表已执行 hosted capture。

## 成功恢复的旧基线证据

`987eacb` 的 migration run `40219606106` 三项全部成功。下载的 Go/Rust artifacts SHA256：

- Go `fc52b84ec4ba65328dad68dde002577bf5e1b26935b5dd959365e464212fda11`
- Rust `0fa6030231c148ad7519ebe98cc3b8a8be500b115a8002bb83e68d6a21f5b1f2`

各含 17 基础状态与 8 workbench 状态。实际查看 Go Usage Details 已显示完整 33 请求、4 模型表格，日志无之前的 GLib 空 callback 错误；这支持 Linux 启动修正，不证明用户 Mac picker 卡住已修复。OCR_READY 仍只表示可人工审阅，不能提升成行为验收或像素相同。这些图不含第三/四波新源码，必须重捕才能评价新效果。

## 下一步与限制

全 scope 仍未完成。P03 setup、P20 文案与剩余焦点/编辑器差异独立继续；139 scope 是核对目录，不是通过数，25 对图也不是全状态。真实认证/provider/MCP、macOS 原生输入法/对话框及完整逐状态像素验收尚有空缺。无新安装包。
