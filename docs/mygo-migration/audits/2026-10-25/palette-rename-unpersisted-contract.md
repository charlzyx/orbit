# Rename follow-up: unpersisted live tasks are valid

The frozen `90ebb5fd...` UI/Controller target-binding candidate is not ready to integrate unchanged. Positive and negative credential-free tests on local pinned Pi 1.4.0 establish that a new task has a SessionPath before any file exists, and setting its name is legitimate without writing that file.

Pinned local Pi source (`@earendil-works/pi-coding-agent/dist`):
- `core/session-manager.js:785–805`: only user/assistant messages count as a conversation; before that, `_persist` returns without creating a file.
- `core/session-manager.js:914–923`: `appendSessionInfo` appends name metadata to the in-memory entries.
- `modes/rpc/rpc-mode.js:523–529`: `set_session_name` accepts a nonempty name and returns success.
- `core/agent-session.js:3241–3245`: naming emits `session_info_changed`.

Evidence in this isolated follow-up tree, with temp HOME/PI_CODING_AGENT_DIR/workspaces, no inherited credentials, offline, extensions/patches/tools/context files disabled, no inference:
- `TestPaletteRenameOfflineSavedFilePositiveController`: actual Controller prompt→explicit saved path→existing metadata lease→real Pi writes the synthetic saved JSONL name; ID unchanged, one original user fixture message, no run.
- `TestPaletteRenameOfflineFreshUnpersistedPiContract`: original live manager rename succeeds, name visible, identity unchanged, no messages/run, file still absent.
- Both positive tests passed race 2× in 6.790s, explicit `-v` log `/workspace/shared/orbit-toolchains-continued/palette-rename-offline-positive-2033.log`.
- `TestPaletteRenameOfflineFreshUnpersistedControllerParity` fails on frozen90ebb: Controller's mandatory file metadata path rejects `lstat ... no such file`. Log `palette-rename-offline-contract-2033.log`, three cases total13.688s, only this case fails.

Approved bounded correction implemented in `internal/agent/session_rename.go`, with UI capture in `palette_rename.go` and dispatch in `run.go`: capture an opaque live rename target while opening the prompt, checking the displayed runtime/session/path against the manager under its mutex. Submit revalidates that exact entry/task generation/client and reserves the existing SetSessionName sessionCommand lease before leaving the lock. Retain until actual completion/child death using existing callSessionCommand settlement behavior. Do not capture a replacement target at Save or fall back to the manager's current task. Sidebar saved rows continue through the existing file ownership metadata route. UI lifecycle/one-shot protections remain. This was approved as a separate follow-up; the original frozen90ebb artifact remains unchanged and must not be integrated alone. A capture failure is reported explicitly without falling back to current selection. The live target is captured under the same lock that checks displayed identity, and includes exact entry, workflow generation and client. Existing lease/protocol settlement code is reused unchanged. Prompt completion/cancel clears the retained target closure; unrelated generic prompts are unchanged.

## Verification boundaries and final gates

Agent synthetic tests verify displayed runtime/ID/path mismatch, zero target, same-child NewSession with the same path and changed ID, generation ABA, Stop/Close/replacement, duplicate/identity lease competition, retained lease after timeout, pending rename across a new task, and Close while a call is held. Initial focused agent2× passed16.199s; timeout-strengthened agent2× passed20.488s. UI initial full Rename2× passed48.722s and vet passed before final cold-row/closure-cleanup additions.

Real offline Pi tests cover live saved JSONL, fresh unpersisted task, stale UI at arm and at commit after same-child NewSession, and cold saved-row metadata rename without activating that task. All files and homes are synthetic temporary fixtures; no inference, provider credentials or user data are used. Earlier90ebb rejection and positive-original-Pi logs are retained rather than rewritten. Final combined focused results will be recorded below; complete package/native macOS parity remains separate.

Final combined focused gate: `/workspace/shared/orbit-toolchains-continued/palette-rename-safe-final-2042.log`, race2× agent20.141s + UI76.512s PASS, including SessionRename/PaletteRename/Delete/Pin/Abort and all real offline Pi cases above. Vet agent+UI PASS (`palette-rename-safe-vet-2042.log`). This gate includes the final cold-row test and prompt closure cleanup. Full package/integration gate is still pending and is not replaced by this result.
