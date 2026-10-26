# P1 event/request/status increment (not automatic recovery)

Base: Go 6a9c9937512387740f7a4e3cc4ca245e78d712df. Source contracts: upstream imrj05/orbit #87 `1a22457eb76dff9cdad2a2669213b7aaeea1d821` and #89 `64a23027f67189a412cebde982e4c1ba2925a4b4`. Complete Rust files were fetched read-only at exact commits and their Git blob hashes verified; manifest and full source are in the external evidence directory below. No whole upstream commit/cherry-pick, model credentials, native capture, or publication.

## Changes and source anchors

- `session.rs:653–666`, `events.rs:210–229,272,350`: record admitted abort only when a run/await/retry/compaction exists. Preserve across late agent_start, consume on retry_end, retire on settle/exit/Stop/session replacement. Explicit false after user cancel is not exhaustion; genuine false remains a typed retry failure. Both interactive Abort and direct Send(Abort) use the same intent.
- `session.rs:278`, `events.rs:783–810`: assign an explicit composer request ID before asynchronous transmission; only its own successful handled response clears awaiting/closes the checkpoint. An unrelated prompt RPC cannot close the live composer. State retirement additionally clears the outstanding ID; late cancelled/failed prompt workers cannot unwind a newer tracked prompt. This is protocol isolation evidence, not a claim that the old UI title action used the same runtime.
- `runtime.rs:16–61,73,107,137`, `events.rs:699–750,358–371`, `transcript.rs:701–715`: explicit error origin, independent of translated/provider text. Ordinary command success clears only its own banner; generic writers always clear ownership metadata. Successful/aborted/non-error message ends clear only the agent-message source. Nonempty assistant errorMessage is surfaced except stopReason=aborted, without requiring stopReason=error. Provider literal is preserved (the source's localized Agent error prefix is not newly reproduced here).
- `widgets.rs:160–225` at #89: status presentation discards CSI and OSC/APC/DCS/PM/SOS through BEL/ST, preserves literal Unicode and the source's remaining escape semantics. It executes no terminal or clipboard control and leaves the raw model status unchanged. Six-second quiet status behavior remains.

Error-source metadata is exported only for compatibility with the existing reflect-based Snapshot deep copy and excluded from JSON. No clone/reflection mechanism change. An initial private-field candidate caused a clone panic; that failed log is retained, not presented as a pass.

## Error writer ownership audit

Known ordinary response failures and the main Send worker get command ownership. NewSession's command failure, Compact's transport failure and interactive Abort's transport failure are likewise explicitly tagged. `parse` is an ordinary named wire command. Auth/quota/entries retain their existing dedicated handling and are excluded from generic same-source clearing; this increment does not redesign those channels or claim full upstream parity for all their error presentation.

Unowned writers intentionally remain: startup/bootstrap failure (mixed stages), extension_error, process/transport/protocol events, content-size rejection, one-way extension UI reply/input/resize failures, checkpoint precedence, manager lastError and setup/Stop reset. Their existing setConnectionError helper always retires older source and RetryFailure metadata. The retry-exhaustion typed setter also retires old command ownership. No text prefix or equal-message heuristic is used. Existing Snapshot redaction still redacts the displayed/provider error; source metadata contains only command/category, not error text.

## Evidence and limitations

External evidence: `/workspace/shared/orbit-upstream-p1-evidence-1919/`.

- `red.jsonl` + `red-hitset.json`: real stdio JSONL child; two top tests and two abort subcases all fail on unchanged production base. Cancel wrongly reports exhaustion; unrelated handled prompt clears awaiting/turn.
- `error-red.jsonl` + hitset: two exact JSONL reducer tests fail for same-source error retirement, including identical text from different origins.
- `status-red.jsonl` + hitset: actual mounted status, SGR and OSC52 BEL/ST, four exact names fail on base.
- `final-agent.jsonl`: retained failed intermediate private metadata/reflect clone candidate.
- `final-agent-v2.jsonl` + `final-agent-hitset.json`: final production, 39 exact top/subtest names ×3 run/pass, race, 117 runs, no skipped names. Includes original retry/redaction/checkpoint/Abort/new-session/parked/Stop/stream regressions.
- `final-real-protocol.jsonl`: final fixture refinement uses a status ACK barrier for handled composer (no synthetic retry on a handled-only path) and checks AnyBusy is false after matching reply. The cancel paths still use actual retry JSONL events.
- `final-ui.jsonl`: final production mounted status/parser/TTL focused race tests; exact hitset recorded separately.
- `vet.log`: focused agent/UI vet only.

Only focused tests and vet, no second full gate, no actual native-pixel acceptance. Automatic process recovery is a separate later candidate and is NOT implemented by this patch. No prompt/tool/queue replay is added. P2 Git/Mermaid/Composer and P3 are excluded.
