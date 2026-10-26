# Jump-only viewport clipping (2026-10-26 11:25 UTC)

Base d4b33caf0cdc5af2b6db8cf684e061ccea89da4a already includes the reviewed fade. Pinned Rust 299a32bc `crates/orbit-pi/src/message_scroller.rs:292–300` wraps the viewport in `overflow_hidden`; jump remains absolute bottom8 with raw32px button (346 onward). No Tab contract change: prior source audit remains authoritative.

Actual mounted baseline RED: a.messages in a 24px viewport at Y60 produced Jump raw Y44/H32. Pixel Y58 was gray, and an actual click there activated Jump outside its viewport. The original failing log is retained; this is not a production crash or whole-app minimum-window claim.

Only new production node is a full-viewport absolute Clip/PassThrough wrapper around the existing jump layer. List/selectable scope/fade ancestry and keys are untouched; raw button32/bottom8, tooltip, label, focus/Space and handler are unchanged. Tiny viewport clips the upper16px rather than scaling; only the remaining16px can activate. The normal shadow is clipped at the viewport edge, matching the Rust shell boundary. No toolkit change, global clip, scroll algorithm or fade change.

Public mechanism tests separately prove raw/visible geometry at24/80. Actual production tiny light/dark tests prove outside paint/input refusal and remaining visible button activation. Existing real wheel, continuous-selection, exact context-menu target, pointer-focus/Space, normal light/dark circle/pill and fade-order regressions ran unchanged. Images in /tmp/orbit-jump-clip-captures-1124 were generated from real test frames; tiny light/dark were visually inspected. The red exterior is a deliberate clipping sentinel.

Exact final JSON evidence: 7 top-level names, each3run/3pass; all subtests each3pass, zero skip/fail,18.288s. UI vet exit0, diffcheck clean. Full suite/native OS input NOT_RUN here; lead owns aggregate. No publication or release.

Final additive1129: actual tiny light/dark hover verifies exact Tooltip text multiplicity (one accessible button Label before/after outside hover, two after visible hover). Initial HasText fixture matched Label and falsely failed; original1127log retained. Corrected3 complete names each3run/3pass, zero skip/fail,3.253s. Production unchanged. Darwin ARM64 crossbuild explicitEXIT0 (1126status).
