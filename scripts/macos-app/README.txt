Orbit Go Test — experimental native macOS app

1. Open the DMG for your Mac: arm64 = Apple Silicon, amd64 = Intel.
2. Drag Orbit Go Test.app to Applications, then eject the disk image.
3. Open Orbit Go Test from Applications. No Launch.command or Terminal setup.

Node 22.23.4, npm, Pi 1.4.0 and Ghostty are bundled. No separate installation,
credentials or user sessions are included. On first launch, public bundled Pi
files are copied once into a versioned private runtime directory. This keeps the
installed app unchanged and permits the existing Pi patch/configuration controls.

Test data is persistent and separate:
~/Library/Application Support/Orbit Go Test/Home
The first workspace contains only bundled synthetic examples, not your project. Opening a
real project later allows normal editing/autosave there. Configuration isolation
is not a filesystem/network sandbox. Provider login/configuration is still needed
for model inference; nothing signs into an account automatically.

This is not Developer ID signed or notarized. macOS may block a downloaded app.
If blocked, stop and report the exact message. Do not change security settings.
No signing/notarization identity is claimed. CI first-frame/Pi/Ghostty smoke on
hosted Macs does not establish Gatekeeper acceptance of a downloaded app or full
manual feature parity. See KNOWN-LIMITS.txt and Resources/Provenance inside app.

中文：下载对应架构的 DMG，打开后把 Orbit Go Test.app 拖到 Applications，
弹出磁盘映像后直接打开应用，不需要运行 Launch.command 或自行安装 Node/Pi。
首次启动会在独立测试目录生成可写的 Pi 运行副本，不导入你原有的账号、会话
或项目。模型对话仍需在应用中配置提供商。测试版未进行 Developer ID 签名和
公证；如果系统阻止打开，请停止并反馈提示，不要降低系统安全设置。
