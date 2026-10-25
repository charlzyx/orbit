# Wave 6 frozen integration — 2026-10-25

This is feature-branch source, not a new installer. Released `.4` is unchanged. Source tests and static captures do not establish full Rust functional or pixel parity.

## Included bounded corrections

- P20 Chat/workspace visible copy; Reset's localized accessible label with a separate stable key. Other English semantic labels remain an explicit gap.
- Appearance grouped controls preserve the source pointer/focus behavior and one group Tab stop.
- Review source row geometry and old/new split-column mapping. Native font minimum/wrapping and exact Nerd glyphs remain unverified.
- Dependency setup replaces only the main column; the sidebar remains usable as in Rust.
- An **unexposed**, fixture-tested installer coordinator and bounded quit settlement. No installer, migration, authentication, or configuration operation was executed; this batch exposes no installer button.
- Inline image owner isolation clears owner-bound caches/preview and suppresses late decode/error completion. Current buttons use owner-epoch keys and row-entry ownership checks; this batch does not claim a separate inner activation check or same-slot content revision protection. The latter is a separately reviewed next increment.
- New Task clears a draft only after synchronous host admission for the unchanged presented intent. Admission is not remote RPC success.
- Terminal per-tab close targets stable IDs; toolbar close hides the panel and retains PTY state.
- Test-only retry receipt timestamp correction preserves the six-second production TTL and existing assertions.

## Frozen combined evidence

570 Go/module files were hashed before the final gate and rechecked unchanged afterward.

- Whole-repository `scripts/go-native.sh vet ./...`: PASS.
- Whole-repository `scripts/go-native.sh test -race ./...`: PASS, including agent 148.865 s, native 70.552 s, UI 511.059 s.
- Darwin arm64 cross-build: PASS; combined command exit 0.
- Local logs: `wave6-{vet,race,arm64}-1940.log` in the retained cloud toolchain evidence directory.
- Chat harness: 29 tests, 28 PASS and one explicit local `/proc/task/children` capability skip; the required hosted regression passed separately.
- Native exact-final-source capture remains pending publication/CI. CPU or mounted fixtures are not physical macOS input acceptance.

## Static Chat guard evidence remains separate

The second opt-in attempt at `3d170d83fe4a72cd8ea0748c4680f0ca6b89d8b9` / run `40227121094` passed aggregate and required real threaded-child ownership regression, but both native capture jobs failed. Go's STARTUP guard accepted no pinned Pi RPC child and dispatched zero inputs. Raw evidence is retained; no retry or weaker process-name matching is authorized by this result. The earlier main-thread-only traversal flaw is real but was not a sufficient diagnosis of the capture failure. This normal source publication removes the temporary one-shot trigger; default Chat capture remains off.

## Separately pending, not included

Palette Delete confirmation target binding, palette Pin current-path correction, per-slot inline-image content revision, npm-only explicit migration UI binding, and Review header glyph/count refinement remain separate next-wave candidates. Real user Mac picker hang reproduction, physical IME/native dialog confirmation, broad settings/auth/MCP acceptance and remaining visual states remain open.
