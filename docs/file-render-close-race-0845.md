# Stale file render snapshot after async close — .3 failure correction

Base local0c5efc82e445bcb73ba94ac1a077db134bed22f7, tree28630c2a98f0cdca5cf28cc5c20d0fa1008afce6. Coordinator verified this full tree equals the public c212 release source. No unapproved transaction/namespace features included.

Official failure retained at `/workspace/shared/orbit-release3-failure-0839/artifact/root-race.jsonl`: TestMountedFilePrimaryWAndShiftWHaveDistinctScopes panicked at files.go783 st.Font during real Cmd-W key frame after330+s of aggregate UI execution. This is a runtime nil dereference, not a detached-ID fixture error.

## Source cause and targeted correction

fileEditorContent captured Active, ignored both bindFileNative boolean results, called syncDocument if path state was absent, then unconditionally dereferenced st.Font. A background close may invalidate that snapshot; liveID validation correctly refuses binding/seeding, leaving nil. Merely adding a nil guard is insufficient: a reopened oldpath can have nonnil state belonging to a new ID, which must not be rendered for the stale document.

- Behavior-preserving extraction `fileEditorDocument(c,docs,d)` isolates the existing snapshot render body for a deterministic boundary regression.
- Both failed bind checks stop rendering the stale body and request a fresh frame; no path fallback and no relaxation of liveID validation.
- A separate nil check after syncDocument handles closure after a successful bind check but before native seeding.
- The next normal frame calls fileEditorContent and takes a fresh Active snapshot. This is not a loop that retains/retries the stale document; fresh empty/live rendering remains normal. Image/markdown/readonly branches otherwise unchanged.

## Actual evidence and limits

Original exact case -race count10 PASS3.030s locally: does NOT reproduce or explain the hosted scheduler's exact timing.

`controlled-render-red.jsonl` FAIL1.782s: capture actual Active document, run actual closeFile async worker and real completion/pruning, then mount the unchanged production snapshot body with the earlier document. It reaches the same st.Font nil dereference (line790 after extraction). This deterministic snapshot/close boundary proves the source contract bug, not the identical hosted event schedule.

`close-render-green.jsonl`: original Cmd-W/Shift-W case and controlled close-snapshot regression, each ten repetitions,20PASS race3.116s.

Final `render-final-focused.jsonl`: 8 top-level /10 PASS test events, zero skip/fail, race3.136s. Includes closed snapshot no resurrection; reopened samepath/newID must not render File editor for stale snapshot, then fresh frame renders live state; original Cmd-W/Shift-W; queued reopen/new selection; subtree rename native caret/Undo; readonly sync failure; closed/re-renamed stale snapshots; actual truncated document. UIvet/diffcheckPASS.

No full suite, native capture, publication or release action. Exact hitset and hashes in local render-proof.json. Root independent diff review and parent aggregate are still required.

## Integrated candidate result

Root independently approved the bounded production diff. Integrated into release4 candidate 29bc26e on the exact c212 source tree, without transaction/menu changes. Full root race3374PASS/19 conditional skips, UI368.423s; toolkit19PASS, terminal115PASS/2 skips, vet and Darwin arm64 build passed. Final1824 file hashes matched. Hosted release and actual installation remain separate pending gates; historical RED and focused-only evidence above are retained.
