# 未发布批次：History / Explorer / Usage / 手动清理

2026-10-25，基于已发布 .4 源码 0af37538a766d57243f9170f28c726327d053127，将先前暂停的 17 个文件逐一核对哈希后隔离集成。原工作目录、其他待处理文档与终端历史实验均未覆盖。此记录不修改 .4 安装包、不请求新 release。

## 本批改变

- History/Graph：短列表从筛选区下方开始，空/加载/长列表/展开与 Load more 保持源行为。
- Explorer：增加宽度拖拽与记忆、键盘/选中路径和相关布局修正。创建/重命名仍为弹窗；菜单仍有已列缺项，不能称完整 Explorer 对齐。
- Usage Details：按原版 token/自然内容高度调整 Breakdown、表头、分页等布局；已验证两种宽度的挂载几何，但未取得新的同状态原生截图验收。
- Retention：手动清理不再被自动清理开关禁止；不会顺带开启自动清理。永久删除仍保留确认，测试仅使用虚构临时会话。

## 冻结源码检查

- 全仓 race：2,344 个 test/subtest pass event，14 skip，24 package pass，无失败。这不是 2,344 项产品验收。
- vet、composer/selection toolkit、terminal module、wrapper 验证通过。
- Linux/Windows amd64+arm64、macOS arm64 五个 cgo-free 编译通过。编译不是目标平台运行。
- 新原生 Go/Rust 同状态截图：等待本次提交的云端流水线，未标 PASS。
- 没有真实模型推理或用户 Mac 验收。

第一次本地整体验证受临时磁盘容量与只读字体/native 缓存影响失败；保留该次记录，纠正独立 scratch/cache 后完整重跑。没有把基础设施失败解释为产品缺陷，也没有略过失败直接称全绿。

具体文件哈希、日志哈希与编译哈希见 [parity-integration-evidence.json](parity-integration-evidence.json)。本批不修复工作区确认无响应、不修复 MCP 完成确认或终端退出后的宿主 Restart 入口；这些独立问题仍在 [验收清单](README.md) 中保留。
