# Review file-header interaction (2026-10-26)

Isolated baseline: `fe034cd0cbfcd86a6d9f28b6d389d12cdb341237`, tree `506739e213841837c07987fbc002a71ba0f44d0a`. Must integrate against current lead tree; no branch/release changes made here.

## Source and actual failure
Pinned Rust `299a32bc`, `crates/orbit-pi/src/sidepane.rs:1628–1667` binds the entire normal/sticky file header to `toggle_file`. That function (`423–429`) changes collapse state and remeasures, without changing selected file or hunk cursor. `review_move_file` (`931–944`) uses the independently selected file as n/p origin.

Go previously only toggled from its chevron and additionally selected that file. Real mounted three-file App tests select alpha through the tree, use wheel to reveal bravo's normal/sticky header, then click actual measured coordinates. Four path/count cases did not collapse bravo. Two chevron cases changed selection alpha→bravo; actual `n` then selected charlie instead of bravo. Open-file child controls separately passed without collapsing their parent. Exact expected RED JSON is `/tmp/orbit-review-header-red-v3-1246.jsonl`; this is not a passing candidate.

## Narrow correction
The existing Row registers public `Clicked`, source pointer cursor and hover surface. Its click and the existing chevron's click combine into a single collapse toggle. Selection and hunk cursor remain unchanged. Existing Open-file control keeps its own hit. No List/scope/key/ancestor, layout, parser, global input or toolkit change. Arrow keyboard activation/focus remains an existing Go control, not a claim Rust has the same Tab stop.

Actual test coverage includes normal/sticky path/count/arrow, n navigation origin, Open-file child, Tab→Space→Enter with retained arrow focus, hover paint/pointer cursor, and below-reader collapse/expand preserving the semantic first row and offset. Existing sticky wheel, horizontal pan, stale held-press and source metadata tests remain unchanged. Collapse of the reader's own removed row and full cross-platform pixel equivalence are not established by the below-reader anchor test.

## Evidence boundaries
Initial fixture failures are preserved: `red-1244` used a nonexistent key constant; `red-1245` Open expected a relative path while production canonicalizes it; `red-v2-1246` encountered that test's own 12-byte generated bravo.go in the package working directory. The corrected fixture uses t.TempDir as real workspace and removed only that known generated file. No user files or real external editor were opened.

First correction run `green-1249` passed original matrix; `boundary-1250` passed focus/hover/anchor controls. Final explicit 10-top-name / 32-including-subtests race3x passed (174.074s), every name ran and passed three times, zero skips/fails; UI vet and Darwin ARM64 crossbuild exited0. Exact logs/hashes are in review-header-interaction-1251-evidence.json. These are not an aggregate or hosted native acceptance claim. No physical clipboard/model/auth/network is used by these tests.
