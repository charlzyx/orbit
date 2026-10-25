# Settings, providers, skills, packages and MCP: P17–P20

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## P17 — MCP apply/reconnect lifecycle

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/mcp_ui.rs::tick_mcp](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L90); [crates/orbit-pi/src/app/mcp_ui.rs::mcp_after_settle](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L277); [crates/orbit-pi/src/app/mcp_ui.rs::mcp_force_apply](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L301); [crates/orbit-pi/src/app.rs::ParkedSession](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L150)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P17.behavior [UNTESTED; ROW_SPECIFIC]** Apply MCP changes and reconnect/probe selected servers while a fixture Pi runtime is active.
   - Expected: Detect external edits/fingerprint, pending apply notice, safe idle restart/resume, defer while running, force-apply warning path; stale parked clients respawn later.
2. **P17.edge [UNTESTED; ROW_SPECIFIC]** Delay or fail apply/reload and switch sessions; latest configuration must win, with retained actionable diagnostics and no duplicate restart.
   - Expected: Changes do not silently abort a live turn. Startup/probe loading, parse/connection/auth failure, reconnect outcome and updated fingerprints require independent states.
3. **P17.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Changes do not silently abort a live turn. Startup/probe loading, parse/connection/auth failure, reconnect outcome and updated fingerprints require independent states.
4. **P17.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P17.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P17.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P17.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P06 [PARTIAL]:** mcpui/workbench scoped race tests and19-case checklist pass. Host mount/deferred reload exists. Real OAuth and authoritative applied/readiness callback are not verified; Applied is deliberately not called on async restart acceptance. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../mcp-ui-implementation.md](../mcp-ui-implementation.md)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P18 — Appearance and typography

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/theme.rs::UiPrefs](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/theme.rs#L324); [crates/orbit-pi/src/theme.rs::FontPrefs](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/theme.rs#L482); [crates/orbit-pi/src/theme/appearance.rs::AppearancePrefs](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/theme/appearance.rs#L31); [crates/orbit-pi/src/app/settings.rs::appearance_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P18.behavior [UNTESTED; ROW_SPECIFIC]** Switch every bundled palette, font family/size, density and background setting; inspect preview and live surfaces.
   - Expected: System/light/dark, separate remembered light/dark themes, full bundled theme catalog and live previews; UI/code fonts, UI/terminal/editor sizes, density, sidebar and reduce-motion.
2. **P18.edge [UNTESTED; ROW_SPECIFIC]** Use missing font and invalid persisted values; relaunch and compare readability, clamps and palette-specific syntax colors.
   - Expected: Theme, ui.json and fonts.json retain choices; sizes 11/12/13/14/15/16/18/20, densities 80–120 in steps of 5. Invalid legacy prefs default/migrate. Theme changes repaint all surfaces consistently.
3. **P18.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Theme, ui.json and fonts.json retain choices; sizes 11/12/13/14/15/16/18/20, densities 80–120 in steps of 5. Invalid legacy prefs default/migrate. Theme changes repaint all surfaces consistently.
4. **P18.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P18.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P18.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P18.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P07 [PARTIAL]:** 10 palettes and55 binding parsing pass; many host labels remain English, backdrop integration is pending, and font/density/motion/OS contrast visual matrix remains open. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../background-acceptance.md](../background-acceptance.md)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P19 — Backdrop customization

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/settings.rs::background_controls](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6382); [crates/orbit-pi/src/app/settings.rs::background_choose](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6569); [crates/orbit-pi/src/app/view.rs::new_task_backdrop](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2271)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P19.behavior [UNTESTED; ROW_SPECIFIC]** Select bundled/custom backdrop and vary blur, fade, opacity and fit options.
   - Expected: Animated ordered-dither new-task field, optional background image, image picker/reset, blur/pixel-size/bottom-fade tuning and preview.
2. **P19.edge [UNTESTED; ROW_SPECIFIC]** Cancel file picker, remove selected asset and use extreme aspect ratios; fallback must keep composer usable and settings recoverable.
   - Expected: Decode/cache off paint path, bounded resize/blur; reduce motion freezes perpetual animation. Persisted backdrop state and image failures need faithful fallback.
3. **P19.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Decode/cache off paint path, bounded resize/blur; reduce motion freezes perpetual animation. Persisted backdrop state and image failures need faithful fallback.
4. **P19.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P19.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P19.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P19.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P07 [PARTIAL]:** 10 palettes and55 binding parsing pass; many host labels remain English, backdrop integration is pending, and font/density/motion/OS contrast visual matrix remains open. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../background-acceptance.md](../background-acceptance.md)

**Remaining:** Backdrop preferences/source audit exist; original ordered-dither field, image/blur/pixel/fade controls and preview are not yet reconstructed.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P20 — Localization and shortcuts reference

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/main.rs::set_app_menus](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L626); [crates/orbit-pi/src/theme.rs::set_ui_prefs](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/theme.rs#L1468); [crates/orbit-pi/src/app/settings.rs::shortcut_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4672)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P20.behavior [UNTESTED; ROW_SPECIFIC]** Change app language, inspect translated menus/dialogs and shortcut reference; search translated commands.
   - Expected: System locale plus English, zh-CN, Japanese, Korean, Spanish, French, German, pt-BR, Russian, Italian; live translated placeholder/menu rebuild; categorized shortcut help.
2. **P20.edge [UNTESTED; ROW_SPECIFIC]** Use missing translation keys and locale fallback; ensure displayed key labels match actual platform-scoped bindings.
   - Expected: Source en.yml + generated locale files. Native macOS menus follow language changes; keyboard label modifier names vary by OS. Accessibility is incomplete, not an implied guarantee.
3. **P20.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Source en.yml + generated locale files. Native macOS menus follow language changes; keyboard label modifier names vary by OS. Accessibility is incomplete, not an implied guarantee.
4. **P20.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P20.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P20.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P20.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P07 [PARTIAL]:** 10 palettes and55 binding parsing pass; many host labels remain English, backdrop integration is pending, and font/density/motion/OS contrast visual matrix remains open. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Locale assets and command registry exist, but hard-coded English labels and context-specific bindings prevent full localization/shortcut parity.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

