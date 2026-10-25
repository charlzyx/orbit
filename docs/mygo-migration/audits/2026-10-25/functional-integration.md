# 未发布功能补齐：MCP 完成状态、终端恢复与回归门禁

本批基于 `.4` → 暂停的17文件批次 `32083a3` → OCR脚本修复 `e577c31`。这不是新的安装包；用户已下载的 `.4` 不包含以下修复。原始脏工作树未覆盖，功能补丁在隔离源中合并，旧诊断脚本与审计文档保留。

## 修正与限定

- **MCP**：等待精确 replacement runtime 完成 bootstrap、恢复同一 saved conversation，并匹配配置版本后才确认；不再永远留在未确认状态。活动会话忙则延迟，不因无关 parked 会话忙而阻止活动会话应用。失败保持可显式重试，新编辑/切会话不能被旧完成事件清除。Reconnect 中断只读 probe，写操作繁忙时明确要求稍后重试。见[源合同与回归](../../mcp-reload-acknowledgment.md)。仍缺 Rust 的闲置 parked 会话 MCP 指纹失效重建，不代表真实 MCP 服务连接或全部行为验收。
- **终端**：接通现有 domain Restart，显示退出/启动错误和恢复入口；保留活动 tab/cwd，不停止其他 tab。测试使用真实临时 PTY 和真实 CPU 宿主按钮，覆盖 exit、启动失败后重试、隐藏重开及重启后立即关闭。见[证据](terminal-recovery.md)。不是 macOS、IME、滚动回溯或像素验收。
- **Explorer 新 badge 缺陷**：原版 helpers.rs:968–990 使用8DIP行高；MyGo `LineHeight(8)` 却是8倍字体高度。改用 `FixedLineHeight`，不改变菜单或图标整体样式。负向挂载测试得到文字高度60、badge高度17；修正后默认/缩放字体、行选中与刷新范围回归通过。新native截图仍须核验。
- **Abort/Clear 测试同步**：原等待条件把本地 awaiting 当成已收到 provider agent_start，release 信号可早于 fixture 暂存回复。仅改三处测试前置同步，不改runtime、不延长超时，三组各重复100次通过。见[因果和边界](abort-test-synchronization.md)。

## 质量门禁的失败也保留

MCP初稿独立复审发现三处组合边界：新编辑覆盖后bootstrap失败不能重试；历史操作错误混入连接错误；忙时Reconnect无声丢失。均补回归修正后再复审，不以初稿focused测试通过作为交付。

`32083a3` aggregate通过但native Go因截断OCR TSV失败；`e577c31`两端native采集通过但aggregate在上述Abort测试失败。详见[实图核对](native-image-review.md)。这些流水线不能拼接成一次全绿。

本地后一次full-race命令返回0，但/tmp容量耗尽使JSON日志截断，因此不作为完整门禁证据。冻结源和生成缓存均校验所有文件哈希后保留式迁移、旧路径symlink兼容；临时构建与日志分盘后重新执行。没有删源代码、真实会话或降低断言。

最终合并源门禁与提交后CI结果在本报告后续记录；执行中的、跳过的与未测试的流程不标为通过。

## 最终本地冻结源门禁（14:34 UTC）

全仓 `-race -count=1` 完整结束，24包、2361个test/subtest pass event、14 skip、0 fail；vet通过；workbench Python43项、2 skip通过。完整JSON日志每行可解析、所有已启动package均结束，源文件哈希与运行前冻结值一致。数量是测试事件，不是产品功能验收。具体哈希见 [functional-integration-evidence.json](functional-integration-evidence.json)。

提交后的同SHA CI与新native采集仍待完成；本地通过不替代它们。之前17文件批次的toolkit/terminal-module/wrapper和5平台编译证据保持独立，不冒充本次新功能源码的目标平台运行。
