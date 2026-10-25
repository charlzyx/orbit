# Visual inventory: mcp (V234–V475)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V234 · Settings mcp baseline
- Features: P01, P14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: Server rows;  Scope/transport/exposure/status;  Add and refresh.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Settings mcp baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V234/rust.png` ↔ `../probes/visual/V234/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V235 · Mcp empty
- Features: P14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_empty_card:L1627](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1627)
- Expected: No configured servers;  Add action.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp empty. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V235/rust.png` ↔ `../probes/visual/V235/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V236 · Mcp search filtered
- Features: P14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: Search input;  Matching servers.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp search filtered. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V236/rust.png` ↔ `../probes/visual/V236/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V237 · Mcp search no matches
- Features: P14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: No matching servers.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp search no matches. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V237/rust.png` ↔ `../probes/visual/V237/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V238 · Mcp status gallery
- Features: P14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: Connected/disabled/error/auth-required statuses;  Tool counts.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp status gallery. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V238/rust.png` ↔ `../probes/visual/V238/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V239 · Mcp add stdio editor
- Features: P15, P16. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_editor_layer:L1823](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1823)
- Expected: Name;  Command;  Arguments;  Environment;  Scope;  Exposure.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp add stdio editor. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V239/rust.png` ↔ `../probes/visual/V239/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V240 · Mcp add http editor
- Features: P15, P16. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_editor_layer:L1823](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1823)
- Expected: URL;  Headers;  Transport;  Secret fields.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp add http editor. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V240/rust.png` ↔ `../probes/visual/V240/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V241 · Mcp edit existing server
- Features: P15, P16. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_editor_layer:L1823](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1823)
- Expected: Existing server values;  Secret placeholders;  Save/Cancel.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp edit existing server. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V241/rust.png` ↔ `../probes/visual/V241/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V242 · Mcp editor exposure radio choices
- Features: P15. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_editor_layer:L1823](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1823)
- Expected: Codemode / CodemodeDeferred / Deferred / Direct / Hidden radio rows; Selected radio styling; no invented dropdown.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp editor exposure radio choices. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V242/rust.png` ↔ `../probes/visual/V242/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V243 · Mcp invalid editor
- Features: P15. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_editor_layer:L1823](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1823)
- Expected: Validation error;  Draft preserved;  No false save.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp invalid editor. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V243/rust.png` ↔ `../probes/visual/V243/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V244 · Mcp remove confirmation
- Features: P14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: Destructive server removal confirmation;  Cancel.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp remove confirmation. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V244/rust.png` ↔ `../probes/visual/V244/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V245 · Mcp test in progress
- Features: P17. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: Testing state;  Duplicate action unavailable.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp test in progress. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V245/rust.png` ↔ `../probes/visual/V245/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V246 · Mcp test failure
- Features: P17. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: Failed test detail;  Retained server configuration.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp test failure. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V246/rust.png` ↔ `../probes/visual/V246/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V247 · Mcp auth required
- Features: P16. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: Sign-in required;  Auth action.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp auth required. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V247/rust.png` ↔ `../probes/visual/V247/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V248 · Mcp auth in progress
- Features: P16. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows:L1135](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135)
- Expected: Device/browser authorization state;  Cancel.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp auth in progress. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V248/rust.png` ↔ `../probes/visual/V248/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Authentication gate: Real authentication remains unverified; only deterministic noncredential visual fixtures are permitted without the relevant authorization.

## V249 · Mcp config external change
- Features: P17. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_external_change_card:L1724](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1724)
- Expected: External change notice;  Reload/apply action.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp config external change. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V249/rust.png` ↔ `../probes/visual/V249/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V250 · Mcp restart pending agent busy
- Features: P17. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_restart_card:L1772](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1772)
- Expected: Pending changes;  Agent-busy restriction;  Force/apply affordance.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp restart pending agent busy. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V250/rust.png` ↔ `../probes/visual/V250/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V251 · Mcp status popup
- Features: P14, P17. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_popup_panel:L2560](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L2560)
- Expected: Status-chip anchored MCP list;  Server actions;  Open Settings.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp status popup. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V251/rust.png` ↔ `../probes/visual/V251/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V252 · Mcp status popup empty
- Features: P14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_popup_empty:L2589](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L2589)
- Expected: No configured server status;  Settings link.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp status popup empty. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V252/rust.png` ↔ `../probes/visual/V252/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V473 · Mcp config parse error
- Features: P14, P17. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_notice_card:L1696](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1696)
- Expected: Configuration parse/probe error;  No false connected status.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp config parse error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V473/rust.png` ↔ `../probes/visual/V473/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V474 · Mcp project untrusted
- Features: P14, P16. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_notice_card:L1696](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1696)
- Expected: Project untrusted note;  No automatic credential exposure.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp project untrusted. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V474/rust.png` ↔ `../probes/visual/V474/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V475 · Mcp codemode disabled
- Features: P14, P15. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/mcp_ui.rs::mcp_notice_card:L1696](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1696)
- Expected: Codemode disabled notice;  Configured server tools not falsely available.
- Setup: Open the isolated mcp surface in fixture.json. Configure the named source state: Mcp codemode disabled. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V475/rust.png` ↔ `../probes/visual/V475/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
