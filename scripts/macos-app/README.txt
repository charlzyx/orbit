Orbit Go Test — experimental native macOS app

1. Open the arm64 DMG on your Apple Silicon Mac. New builds are arm64 only.
   Existing Intel releases and assets remain unchanged.
2. Drag Orbit Go Test.app to Applications, then eject the disk image.
3. Open Orbit Go Test from Applications. No Launch.command or Terminal setup.

Node 22.23.4, npm, Pi 1.4.0 and Ghostty are bundled. No separate installation,
credentials or user sessions are included. On first launch, public bundled Pi
files are copied once into a versioned private runtime directory. This keeps the
installed app unchanged and permits the existing Pi patch/configuration controls.

Orbit test preferences and the writable public-runtime copy are persistent and separate:
~/Library/Application Support/Orbit Go Test/Home

Pi itself uses your existing PI_CODING_AGENT_DIR (including ~ expansion), or
~/.pi/agent by default, with your normal HOME and launch environment. Existing
configured providers/models and their credentials are used in place, not copied.
Stock Pi may run normal migrations, refresh auth, and write settings/model caches
or sessions when launched. This is not a read-only Pi bridge. Orbit does not add
an automatic credential/configuration import or migration of its own.

The first workspace contains only bundled synthetic examples, not your project.
Opening a real project later allows normal editing/autosave there. --home is an
explicit isolated Orbit/Pi configuration override for disposable testing; it does
not remove inherited environment credentials or sandbox filesystem/network access.
The opt-in --smoke-report flow strips ambient credentials and forces offline mode.

Local diagnostics (this .4 build and later):
~/Library/Application Support/Orbit Go Test/Home/.orbit-pi/logs/app.log
The previous rotated file is app.log.1. Each is capped at 1 MiB for this app
process. Logs contain only fixed stage/result names, timing, version and a random
run ID. They do not contain prompts, messages, credentials or workspace paths,
and the app does not upload them. Logging failure never blocks startup. A last
"workspace_browse / started" entry can identify a pending native dialog; it is
not itself a diagnosis. The workspace-selection hang remains unconfirmed fixed.

This is not Developer ID signed or notarized. macOS may block a downloaded app.
If blocked, stop and report the exact message. Do not change security settings.
No signing/notarization identity is claimed. CI first-frame/Pi/Ghostty smoke on
hosted Macs does not establish Gatekeeper acceptance of a downloaded app or full
manual feature parity. See KNOWN-LIMITS.txt and Resources/Provenance inside app.

中文：后续仅构建 Apple Silicon（arm64）版本，已发布的 Intel 版本和资产保持不变。
下载 arm64 DMG，打开后把 Orbit Go Test.app 拖到 Applications，
弹出磁盘映像后直接打开应用，不需要运行 Launch.command 或自行安装 Node/Pi。
首次启动会在独立测试目录生成可写的公共 Pi 运行副本。模型、账号配置等直接
使用你已有的 PI_CODING_AGENT_DIR 或 ~/.pi/agent，不会复制到测试目录。
Pi 启动时仍可执行正常配置迁移、认证刷新及设置/模型缓存写入，并非只读模式。
--home 可显式指定隔离的测试配置目录，但不是文件系统或网络沙箱。
本地日志位于上述 .orbit-pi/logs/app.log，只记录固定阶段、结果和耗时，不自动上传。
选择工作区后无响应的问题尚未确认修复，整体迁移和像素对齐仍未完成。
测试版未进行 Developer ID 签名和公证；如果系统阻止打开，请停止并反馈提示，
不要降低系统安全设置。

Publisher safety preflight and diagnostics (offline-tested candidate)
-----------------------------------------------------------------
Before its first write, publish-prerelease.py reads the exact feature ref
refs/heads/redesign/mygo-rpc and requires a commit object whose SHA equals
GITHUB_SHA. A failed read, malformed ref or moved head stops without writing.
The check is repeated immediately before publishing the owned draft. An already
complete exact release still follows the existing read-only idempotent path.
These GETs are not an atomic lock: keep the feature branch frozen/serialized
through publication. A later head move or API error can leave a draft/uploads;
there is no automatic retry, resume, cleanup, permission change or ref force.

HTTP diagnostics report only the initiating request's fixed stage, method,
allowlisted host/path, HTTP status and fixed error category. Query strings,
headers, tokens, arbitrary response messages/errors/documentation URLs, raw
bodies and gh stderr are never echoed. HTTP error JSON is read at most4097bytes
and used only for a fixed-message category allowlist; oversize/malformed/unknown
content yields a generic category. gh stderr is captured, not inherited.

Immediately after a create response passes exact draft/prerelease/tag/source,
empty-assets and positive-integer-ID checks, a flushed publisher_owned_draft JSON
line records only repository, exact returned ID and the validated local tag/SHA.
This precedes uploads, so a later failure retains an inspectable owned ID in logs.
Malformed or identity-mismatched responses never claim ownership or log a remote
object; absence of this receipt does NOT prove the create request had no effect.
Inspect uncertain outcomes before another action. Original source/checksum,
asset ID, tag/ref, fingerprint and final readback checks remain in force.

This change does not establish why an earlier publish request returned403 and
does not grant or repair workflow/token permissions. A moved branch was only a
hypothesis. No live publishing operation is part of these tests.

Offline coverage: the existing CI entry test_app_package.py now executes all
39 tests (all previous31 retained, eight additions); discovery includes five
Go-icon tests for44 total. Separate legacy package tests21PASS. Exact recording
fixtures cover no-write head failures, moved head after verified uploads stopping
PATCH, immediate valid-ID receipt, invalid response rejection, and poisoned
HTTP/CLI/transport errors with no secret output. No new workflow trigger or
permission, credentials, marker, auxiliary ref, Go build or full suite is added.
