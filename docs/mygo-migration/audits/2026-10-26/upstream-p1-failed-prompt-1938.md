# Go multi-request safety increment after P1 event candidate

Depends on event/status candidate 670335f (local freeze f63eabb), does not modify its frozen artifact. Root review found failed Prompt responses still unwound awaiting without checking ID. Exact Rust #87 handles failure before its successful handled-ID guard and shares that residual risk: this increment is explicit Go multi-request safety, not falsely attributed to an upstream fix.

Production delta: failed Prompt clears awaiting/request-ID only when it owns the current composer ID. Manager retires only that matching optimistic turnOpen and sets Ready only without actual running/retrying/compacting activity. It does NOT create a successful ending checkpoint for a rejected prompt, replay any prompt/tool, or suppress the foreign command's ordinary source-owned error banner.

Actual stdio JSONL child tests:
- A different request fails while composer is pending; composer must survive and matching success still finishes it.
- Old first prompt fails after a second accepted prompt; only matching current failure may retire awaiting/turnOpen, LatestTurn remains unchanged.
- Missing/empty response ID success and failure cannot claim the composer.
- Late success cannot consume newer request.
- Same actual Pi child is reused through NewSession A→B→A, then old request failure cannot cancel the new composer despite the repeated session ID.

External evidence `/workspace/shared/orbit-upstream-p1-evidence-1919/`:
- `failed-prompt-red.jsonl`: exact two new top-level tests each run/fail on candidate670, actual child not direct-helper invocation.
- `failed-prompt-green.jsonl` and `failed-prompt-green-hitset.json`: nine exact top/subtest names ×3 run/pass under race (27 runs, no skips), including previous real cancellation/handled cases.
- `failed-prompt-vet.log`: focused agent vet PASS; no full gate/native run/publication.

The separate missing/malformed message_end clearing audit remains source/fixture-only pending root review. This increment does not change that boundary. ANSI work remains display escape stripping, not a claim of sanitizing every control character.
