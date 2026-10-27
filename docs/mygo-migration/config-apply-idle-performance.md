# Config-apply idle observation fast path

`Controller.syncAt` already obtains the frame's snapshot; its later
`tickConfigApply` previously made another full Manager snapshot and runtime
snapshot even when no config deadline or replacement flight existed. In that
state the old code had no cleanup, ACK, or notification to perform. The new
three-line guard runs after the existing stop check; scheduled work (including
future deadlines) and flights retain the complete old state machine.

Synthetic real JSONL child histories, Go 1.27.3, canonical wrapper, non-race
Linux amd64 EPYC 9V74 (not macOS frame rate):

| Idle history | Before | After | Allocations before → after |
|---:|---:|---:|---:|
| 0 messages | 6.17 us | 7.76 ns | 42 → 0 |
| 1,000 approximately 1 KiB messages | 42.81 ms | 7.33 ns | 8,044 → 0 |

This is measured on the original byte-copy baseline, independent of the separate
snapshot byte-copy candidate. The remaining first snapshot and transcript
projection are not optimized here; this is not a claim that typing is fixed.

A zero-value real Manager RED measured 26 unwanted allocations per idle call.
The green requires zero plus a nil-runtime access sentinel. Future deadlines,
flight observations and stopped owners are negative controls. Existing real-child
MCP/config reload tests cover completion ACK, deferred busy targets, failures,
retry, new edits, session replacement, and generic-vs-MCP acknowledgment.
No manager test hook, unsafe lock access, replay, or scheduling changes are used.
