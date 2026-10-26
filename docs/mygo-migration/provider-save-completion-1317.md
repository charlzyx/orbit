# Provider save completion versus modal lifetime

Isolated baseline d05facafbbceddff535b764054925ea4a7195165, tree d0005488893adc8854a0d3724904aeb10db65083 (lead Review integration). No branch, release, or aggregate execution here.

## Actual contract and RED
Pinned Rust299a32, app/settings.rs:7984–8024 cancel/scrim close the editor. provider_save:8027–8100 writes configuration synchronously, then always reloads custom providers and catalogs on success. Go makes the write asynchronous. Its providers.go:66–98 formerly discarded all completion effects if the modal was no longer the same editor, even after models.json had actually been written.

The new mounted tests use real Add provider, Save, Escape/scrim, then Add provider again. They hold only the already-completed UI callback from App.updates; SaveProvider really writes the temporary PiDir/models.json. The captured completion is then released against the new draft. Both dismissal cases had zero RuntimeConfigChanged/RefreshProviders instead of one. Same-editor success, actual write failure, and Cancel/Escape/scrim without Save passed. Exact RED is /tmp/orbit-provider-save-red-1314.jsonl (2.919s), not a passing gate.

## Narrow fix and ownership
Successful completion clears its own old d.key promptly. Different workspace remains an early return. In the same workspace, successful persisted mutation dispatches RuntimeConfigChanged and refreshes providers regardless of whether the old modal is still open. Editor pointer identity only gates consuming the draft or displaying its write error. It cannot close/overwrite a replacement editor. Failed writes do not emit successful effects. No new global callback/target protocol.

SaveProvider writes global Store.PiDir/models.json, not a session file (internal/workbench/providers.go:57). Existing Controller.RuntimeConfigChanged schedules active-only config apply at completion (config_apply.go:28), capturing runtime/path/session and revalidating after the500ms delay and before RestartIdle. Busy/loading defer, mismatched target refuses, parked sessions retain their running configuration. Same-workspace task changes still use the same global provider file; this candidate retains existing normal-save completion semantics, not a command to resurrect the original task. Different-workspace suppression is intentionally retained. No claim that every running process reloads or that prior missed notification made every runtime permanently stale.

## Evidence boundaries
Final focused includes the original dismissal cases, new-draft preservation, same-editor save, unsuccessful save/error retention, no-save cancellation, different workspace suppression, and same-workspace task-change projection. Tests use synthetic IDs, example.invalid URLs and a nonexistent fixture Pi binary; no model/auth/network. One old-object cleanup assertion inserts an unsent marker after the write completed solely to observe key clearing; it never writes that marker as a credential. Existing credential regression uses only its pre-existing disposable fixture data.

First corrected run passed4.428s. Final race3x exact8top / 14-including-subtests names each ran and passed3 times, zero skip/fail,7.916s; UI vet and Darwin ARM64 exited0. Exact JSON/status/log hashes are in provider-save-completion-1317-evidence.json. These are not full aggregate/native acceptance. No files outside test temporary homes are modified by the tests.
