# Wave10 image gate: same-workspace refresh is not a new owner

## Causal evidence, not a timeout/flakiness workaround

Exact unchanged `TestMountedMarkdownImageServiceRoundtripInvalidatesHeldOwner`:
wave9 baseline 5 run/5 pass; wave10 5 run/5 fail at original fresh-preview assertion.
Logs `/workspace/shared/orbit-image-wave{9,10}-exact-2214.jsonl`; checked hitset
`orbit-image-exact-comparison-2216.json`. Full failing aggregate remains preserved.

Isolated stage instrumentation (`orbit-image-wave10-diagnostic-2215.jsonl`) showed
stable button bounds {272,541.34753,162,82} and Markdown cache owner 3, but inline
owner generation changed 4→5 after `changeWorkspaceServices(initial)` despite the
same service/model workspace. The owner-keyed copy wrapper correctly changed the
button subtree identity, so release could not activate the prior press. Thus the
new wrapper exposed a genuine conflation of idempotent service refresh and user
intent. No owner guard, click check, original assertion or timing limit is removed.

## Minimal correction and all production entry points

`run.go::changeWorkspaceServices` advances the shared epoch only when the requested
path differs from either controller workspace or Model.Workspace. It retains both
A→B and B→A boundaries and repairs model/service mismatch. Same-owner refresh still
performs the existing service refresh work, without changing ownership.

`run.go::SetWorkspace` explicitly advances when both workspaces already equal the
requested path, before invoking the helper. Consequently every admitted nonempty
SetWorkspace intent advances exactly once: the explicit branch for equal paths,
or the helper for changed/mismatched paths. Cancel/busy paths remain unchanged.

Production callers (entire internal/ui source search):
- Helper: `run.go:303`, controller tick reconciliation, only if the active session
  reports a nonempty workspace different from the current controller workspace.
- Helper: `run.go:370`, admitted SetWorkspace as described above.
- SetWorkspace: `run.go:180`, Worktree OpenWorkspace callback.
- SetWorkspace: `run.go:535–536`, dispatched SetWorkspace action.
- SetWorkspace: `workspace_dialog.go:76`, already-owned successful native sheet.
No other Controller helper/SetWorkspace call sites were found. Terminal/Git/MCP/
Resources SetWorkspace methods are separate service APIs, not these entry points.

NewSession action and OpenSession retain their independent unconditional intent
bumps (`run.go` respective handlers); no changes there. AttachmentContext also
continues to include Agent.CaptureWorkflowTarget, not only the workspace epoch.
`agent/workflow.go:11–20` captures opaque live-entry key + workflowGeneration;
manager resets/accepted new-session paths advance that generation (`manager.go:444,
621`). Actual manager-created same-workspace empty-ID task and service roundtrip
regressions pass independently of view-supplied session IDs. Close remains guarded.

## Tests and limits

Original image test file is byte-identical, SHA256
`d5e801c2e2d1534ae75c67d47b50cb63e3b033ba7909d47d7888a0c31d8836c0`.
Core original test + two new boundary tests: exact 3 names x5 = 15 run/pass,
zero skip/fail, UI 3.505s (`orbit-image-owner-fix-core-2217.jsonl` + hitset).

Picker/dialog/attachment/inline/Markdown matrix: exact 41 top/subtest names x3 =
123 run/pass, zero skip/fail, UI 13.693s (`orbit-image-owner-matrix-2217.jsonl`,
`orbit-image-owner-matrix-hitset-2218.json`). Includes actual Controller-frame
completion, cancellation/error, old sheet after New/OpenSession/workspace/ABA,
shutdown, file-dialog stale task/window, empty-ID attachment worker, identity
hydration, actual manager task generation and held/fresh image clicks.

Targeted UI vet passed (`orbit-image-owner-vet-2218.log`). No second full UI, native Mac run, new capture, publication
or pixel-parity claim. Parent owns the revised combined aggregate.
