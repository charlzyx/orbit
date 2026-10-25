# 第五波：只读 setup、真实状态与有界本地化

基于已发布第四波 `00e685d1c0b13eebba068ab5119be0ef258ca59c` 的同字节本地树。安装包仍为 `.4`；这里不是新 release，不合 main。

## 本批源码范围

- **P03 部分**：原本不可达的缺依赖 setup 只读入口；检查实际配置 binary、必需 Pi/Node 和可选 Git，复制安装提示、刷新、完成后仅 exact target/binary 的 never-live failed entry 恢复。singleflight/generation/关闭守卫，任何 installer、npm migration、认证写入都不在此批执行或暴露。
- **P20 第一批**：Agent 真实 retry 状态、会话菜单接原 catalog；Go 独有七键使用英文/简中补充，原 catalog 优先，其余语言明确英文 fallback。普通 provider 错误原文不猜测翻译。不是全部中文、不是全 accessibility 本地化。
- **C04 文件分类**：通用文件添加按 Rust 六种图片后缀；SVG/TIFF 文件在 caret 插路径，clipboard 的原始 MIME/字节仍不变。未修改读取失败策略、附件限额或 decoder。
- **文件 metadata**：loaded/saved 内容行数与可编辑 buffer 行数分开，保留末尾 LF 和空末行；保存后更新 disk metadata。Dirty 不再一律假称 Saving，明确 conflict/error/saving/unsaved/saved。Rust 图片 `0 lines` 的现有行为亦保留并测试，不假装原版显示 dimensions。
- **S05 只读补证**：Settings、Usage、Worktrees 四条真实 mounted 键盘路径未复现 trap。保留 tests，不因此把 surface ring 扩展到所有页。Appearance 单组 Tab stop 差异在另一隔离批，未混入这里。

## 取证与边界

987 基线的 25 对图现在均已有实际逐张人工复核，报告分别记录 Workbench/Usage、五项 Settings、其余基础页。**这不是 25 对通过**：图形差异仍多，空态不证明 MCP/plugin/provider/模型可用；真实 runner temp 容量、构建技术栈、隐私保护及 truthful unknown 必须与意外缺口分开。新的 runtime 字节还需要 exact-SHA 重捕。

静态 Chat 单次 test-only 提交 `1305b509b4973e5b80fdb7323e1705d3a670a270` 的 run `40224305711` 确认真正启用输入 1，但 Go 与 Rust 分别在 STARTUP guard 阶段阻断，均 0/2 case、尚未发 Ctrl1。原图显示 Connected，启动源码调用 Agent.Open；不能把缺少 RPC descendant 简单解释成 lazy Pi。正在检查 harness 只枚举主线程 children 的归属判断，不降低所有权/实时图像 guard，不把框架失败说成 app crash。

此批已移除临时单次 push 条件，保留原 manual opt-in 与所有默认-off负向矩阵。任何纠正后的 capture 重试需新的已审确切提交，不能不加区分重跑。

## 门禁记录

独立最终门禁与 scoped reviews 已留档。组合重点 race：UI 15.131s、agent 3.254s；native 包此次筛选无匹配测试，不算 native 通过。19:07 冻结 553 个 Go/module 文件后，全仓 vet→aggregate race→darwin/arm64 顺序链 EXIT 0；agent 153.835s、UI 460.125s、native 79.581s，其余包亦通过。553 个源文件 hash 复验不变。后续只增取证文档，不移动源码门禁。

下一批独立进行 P20 chat/workspace、Appearance focus、source Review 行布局和安装协调器的 fixture-only 合约；未据此宣称这些已经交付或原版全部功能完成。

### Harness 修正的独立证据门槛

有界 all-task children scanner 已独立审阅，RPC/窗口/图像/当前child identity guard 全保留。组合 Python 29 项：28 PASS、1 因本地 `/proc` 不提供 task children 而明确 SKIP。新的 Ubuntu fast regression job 必须通过真实 worker-thread child 用例，两个 renderer 才可开始捕获；普通提交 Chat 仍 off。未把确定性 fixture 或 Linux 文档当 hosted 因果证明。

P03 当前只读入口的 sidebar/main-column 源布局差异另有已隔离修正，未混入本冻结源码；P20 第二批、Appearance 单组 Tab 与安装协调器也留待下一批。
