# MyGo migration: verified baseline

This directory records the source baseline and acceptance contract for the
confirmed Go + MyGo native reconstruction of Orbit. Implementation is in progress;
this directory is not evidence of completed feature parity. The existing Rust
application is preserved only as a comparison reference, not the Go deliverable.

## Pinned inputs

- Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac` (v0.2.12), Apache-2.0.
- MyGo `9c69cf829ece196c5b3a3601f1c7bbe70526376f`, MIT; requires Go 1.27.3.
- Pi `@earendil-works/pi-coding-agent@1.4.0`, official tag
  `4a080252f06caf63a525d7873c5d94d3c23887e0`; see the RPC contract report.
- `baselines.json` records reference screenshot checksums and dimensions.

Orbit already speaks Pi's JSON-lines RPC over a child process's standard input and
output. MyGo is a desktop application framework, not a chat application's theme.
Replacing GPUI with MyGo's native UI is a renderer/controller migration, not a
stylesheet change. No webview or mock service is an equivalent implementation.

## Acceptance principles

1. Preserve all functionality actually implemented at the pinned Orbit commit,
   including error, loading, cancellation, persistence, keyboard and restart paths.
2. Do not count controls, screenshots, mocks, or source-preserved Rust features as
   implemented MyGo features.
3. Keep Orbit's planned/unshipped functionality separate from shipped behavior.
4. Keep Pi responsible for its agent state, credentials, model catalog, tools and
   sessions. The Go/native layer must consume real protocol and filesystem state.
5. Record each ported feature's implementation and executable evidence. Distinguish
   automated, manually exercised, platform-blocked, and not-run validation.
6. Compare visuals at the same viewport, scale, fonts, palette, locale, data and
   selected state. Record geometry and pixel diffs; do not declare pixel-perfect
   parity from visual impression alone.
7. The marketing screenshots are older than the code baseline: the General image
   shows v0.2.5. Treat them as reference images, not current-runtime captures.
8. Preserve original copyright/license notices. Retain MyGo's MIT notice for any
   incorporated source. The fork's identity, update feeds and signing must not
   impersonate upstream or silently replace a fork with an upstream update.

## Verification already performed

These are **upstream baseline checks**, not migration completion checks:

- All 10 `contrib/**/*.test.mjs` files: 122 tests passed, none skipped.
- `cargo test --locked -p orbit-rpc --lib --test auth_fake_pi --test quota_fake_pi`:
  39 unit tests and 3 scripted child-process integration tests passed.
- `python3 scripts/check_i18n.py`: 1,599 used keys, 2,039 definitions, zero missing
  keys and zero placeholder mismatches.
- Real Pi 1.4.0 was run with an isolated empty home/config, offline and without
  model inference. State/catalog/abort/agent-control/session-read commands were
  exercised. Stock and Orbit-patched results are distinguished in the RPC report.
- Upstream MyGo's gallery was rendered with `ui.Tester` using Linux text shaping
  and the CPU rasterizer. Those images prove native headless rendering, not native
  desktop GPU, input-method or operating-system integration parity.

The cloud has not accessed the user's Mac, browser, credentials or server.
The user created `https://github.com/charlzyx/orbit` after the available connector
was found not to expose repository creation. GitHub now verifies `fork: true`,
with both parent and source equal to `imrj05/orbit`, and the authenticated account
has push permission. The initial source checkpoint below has now been remotely verified.

## Reports

- `feature-parity.md`: source-derived inventory and acceptance requirements.
- `pi-rpc-contract.md`: real protocol, extension and compatibility findings.
- `native-ui-feasibility.md`: MyGo APIs, platform constraints and visual baselines.

The native Go framework choice is confirmed. A skin-only implementation, Rust
backend sidecar, or browser/webview replacement would not satisfy the request.

## Published Go checkpoint

The initial native source checkpoint is commit
`80580ee3cffe8eab19b9399d7813fa6be61ec6d9` on
[`charlzyx/orbit: redesign/mygo-rpc`](https://github.com/charlzyx/orbit/tree/redesign/mygo-rpc).
The entire frozen Go source passed race tests and vet, including three real
credential-free Pi integration cases and real disposable PTY/native-grid tests.
The actual `cmd/orbit` executable cross-built with cgo disabled for all six
Linux/macOS/Windows amd64/arm64 targets. These checks do not prove native OS
integration or completed feature/pixel parity. Hosted Go CI also passed for
that exact checkpoint.

The first hosted reference capture failed because a shallow checkout lacked
the pinned source SHA, before compiling or opening an app. Commit `ae3d5c8`
fixes that fixture and captures both original and Go native desktop windows.
Its pixel/OS outcome must be recorded from actual artifacts, not inferred from
the workflow's presence.
