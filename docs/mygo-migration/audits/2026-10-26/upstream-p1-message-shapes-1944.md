# Message-end source boundary and combined P1 candidate

This increment follows frozen event/status candidate670335f and failed-prompt candidate5b8a00a. It is source equivalence, not broad input sanitization or automatic process recovery.

## Exact upstream evidence

At upstream #87 `1a22457eb76dff9cdad2a2669213b7aaeea1d821`:
- orbit-rpc/src/types.rs:885–898,974 maps any valid object with type=message_end to MessageEnd, without validating message shape. client.rs:488–516 has no second message validation.
- orbit-pi/src/transcript.rs:701–715 chooses `value.message` if present, otherwise the entire event; accepts assistant role, excludes explicit aborted stopReason, and trims a string errorMessage. Invalid unrelated fields do not veto a valid selected message/error.
- orbit-pi/src/app/events.rs:358–371 raises agent-source error on Some and clears only that source on None.

Thus missing/null/array/string/number/empty-object message may clear a prior agent-source error in Rust. This is deliberately retained and documented as tolerant source behavior, not labelled a safety improvement. It must not clear a command or unowned error. Invalid JSON or an invalid outer event cannot become a successful MessageEnd. Rust reports those as Unknown; Go retains its existing protocol-error reporting rather than silently erasing a failure.

The new projection receives `event.Raw` while the existing transcript body still receives `Fields["message"]`. Missing message can therefore use the bare envelope error; explicit null cannot fall back. Helper extraction is per raw field, not whole-struct decoding: wrong-type outer fields cannot invalidate a valid nested message, and wrong-type stopReason is not the string "aborted". Transcript storage/decoder is unchanged.

## Actual JSONL child evidence

External directory `/workspace/shared/orbit-upstream-p1-evidence-1919`:
- `message-shape-red.jsonl` and hitset: 13 exact names. Only bare-envelope case (and parent) fails before moving projection; the other shape/foreign-source/invalid-outer negative controls pass.
- `message-field-red.jsonl`: two additional exact child cases fail with whole-struct decoding, for malformed outer types and nonstring stopReason.
- `combined-agent-final.jsonl`: event ownership, failed-ID supplement, message shapes and existing lifecycle/checkpoint/abort/redaction tests in one race-focused invocation. Exact final hitset is recorded externally; no test-name-only or zero-match pass is counted.
- `combined-ui-final.jsonl`, `combined-vet-final.log`, `combined-arm64-final.log`: focused mounted ANSI/parser/TTL regression, vet and arm64 cross-build. The arm64 binary is compile evidence, not an installed/native-run or release artifact.

Original RED and intermediate Snapshot clone panic logs remain unchanged. No full suite, actual macOS interaction, publication, provider credentials, or prompt/tool/queue replay was performed. Limited process recovery remains a separate future patch.
