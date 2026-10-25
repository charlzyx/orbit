# W65 bounded terminal host recovery

Base: shipped .4 0af37538a766d57243f9170f28c726327d053127.
Patch: orbit-terminal-restart.patch (app.go and new terminal_host_test.go only).

## Contract and change
Rust crates/orbit-pi/src/terminal.rs:1806 restart retains active tab directory; :2035–2118 provides header restart, exited badge and recovery card. Go internal/terminal/panel.go:472 already implements active-only asynchronous restart with generation protection. Previously internal/ui/app.go:887–919 never consumed Exited/Error/Tooltip or called Restart.

Host now renders translated exited badge/card, restart header/card controls, persistent startup error, and tab tooltips. Existing Panel.Restart handles lifecycle. Controls disabled while active tab is loading. Startup failure title uses a fallback string (translation entry is not added here). No terminal domain or native grid changes.

## Test evidence
All tests use generated local fixtures/disposable POSIX PTYs, not user sessions.
- orbit-terminal-restart-red.log: baseline host fails both new tests for absent persistent exit/error cards.
- orbit-terminal-restart-final.log: new actual CPU host button tests pass with -race -p 2 (3.180s). Exit 7 → card Restart → fresh shell retains initial cwd/tab ID; independent tab remains same process/output; header restart; hide/reopen; restart followed by close does not resurrect tab. Missing executable → persistent error → generated safe shell → Retry succeeds.
- orbit-terminal-restart-vet.log: scripts/go-native.sh vet ./internal/ui, exit 0.
- orbit-terminal-restart-domain.log: existing TestPaneShortcutsPreserveShellKeys, TestLiveTabLifecycleAndNativeGrid, TestLiveHiddenPTYKeepsRunningWithoutWindowFrames all pass under -race -p 2 (1.537s).
- orbit-terminal-restart-captures/{exited,restarted,startup-error}.png: CPU host render; exited/error cards visually inspected. Fixture terminal palette is not representative of native app palette.

Invocation environment: ORBIT_GO=/workspace/shared/orbit-toolchains/go/bin/go; GOMODCACHE=/workspace/shared/orbit-toolchains/gopath/pkg/mod; GOCACHE=/tmp/orbit-pi-inheritance-cache; GOFLAGS=-buildvcs=false; ORBIT_TERMINAL_TEST=1; MYGO_GHOSTTY_VT points to pinned 9dcd8af0.../libghostty-vt.so; own writable TMPDIR and XDG_CACHE_HOME.
Host command: scripts/go-native.sh test -race -p 2 ./internal/ui -run '^TestTerminalHost' -count=1 -v.

## Limits
Not macOS native acceptance or total terminal parity. Horizontal tab-strip scrolling, exact restart spin animation, IME/clipboard, reflow/scrollback remain outside patch. Repeated actions are disabled during loading but no deterministic double-click stress test was added. This records the worker-stage proof; final coordinator source gates and feature-branch publication are recorded in functional-integration.md. No new release. All paths here are local evidence; captures may contain synthetic scratch paths.
