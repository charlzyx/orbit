# Wave 9: exact rename ownership and asynchronous completion safeguards

Status: source/test integration validated, based on public `99866d8611aac61fcef9267e6d7cfc534c050944`. The shipped `.4` installer is unchanged. This is not full migration or native acceptance.

## Included scope

- Rename captures the displayed runtime/session/path under the manager lock before prompting, then revalidates its opaque entry/generation/client and reserves the existing session-command lease at Save. Fresh Pi tasks with a path but no persisted JSONL remain renamable. Cold saved rows retain file/path ownership checks. There is no fallback to whichever task is current at Save. The earlier saved-only intermediate fix was rejected because it broke valid fresh-task rename.
- Automatic Pi updater shutdown tracks the actual worker settlement channel. Close cancels and waits within one shared three-second deadline for explicit and automatic workers, without waiting for a UI acknowledgement or holding a manager lock. Timeout means uncertainty, not claimed cleanup. No actual updater or installer was executed in validation.
- Typed in-memory deferral reasons distinguish the existing busy-timeout event from missing-verifier or admission failure. Only confirmed busy timeout receives the original next-launch notice; worker scheduling and business deadlines are unchanged.
- Pending image read/decode work is owner-scoped and blocks Submit until settled. Both mounted submission and Controller admission preserve the draft rather than sending without the image and allowing its late callback to attach to the next draft. No automatic send is added. Task changes and Close discard old ownership. Capacity, ordering and origin-specific failure behavior are separate pending work.

The automatic target-home/PiDir guard is already in Wave 8 (`30aa6142…`), not reapplied here. Fractional Review glyph policy, dynamic palette titles, code-copy fixes and attachment capacity/order are excluded from this freeze.

## Evidence and limits

The combined focused race run passed: agent 10.760s, native 1.055s, UI 32.749s. Diagnosticlog compiled in that filtered run but matched no tests; it is not counted as a diagnostic test pass. The first complete race gate failed solely at `TestModalTransitionRestoresComposerFocus` (UI package 498.406s); arm64 was not reached by the chained command. A focused rerun reproduced the same failure in 3.445s. Its generic App fixture had no SessionPath but expected Rename to open a prompt, conflicting with the new explicit-target contract. The test-only correction supplies a disposable target path (no file is written), retaining every original prompt/focus/Escape assertion and the separate pathless-rename refusal test. Production code is unchanged. The complete freeze-v2 gate passed: vet, all 25 race-tested packages (UI 558.443s), and Darwin arm64 cross-build, with all 607 Go/module hashes rechecked unchanged. Corrected focus plus Rename tests also passed three repetitions in 58.582s. The earlier failed gate is not relabeled as passing. The source freeze records 607 Go/module hashes; the final gate used those exact bytes.

Tests use controlled sessions, synthetic RPC children or offline pinned Pi metadata operations; no user session deletion, credentials, model inference or real installation. UI prompt ownership and process lifecycle tests do not imply native macOS physical-input acceptance. The user's picker hang remains unconfirmed.
