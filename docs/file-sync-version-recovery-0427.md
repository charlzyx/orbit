# Native editor version acknowledgement on SetText failure

Dependency:0864f5bcf68e30e55df4aea897c5e73d8b6520d3 (new rename-preflight candidate based on ea98724). This is an independent narrow increment; no public/frozen capture checkout was changed.

fileEditor previously called Editor.SetText, displayed its error, then unconditionally recorded the native Buffer version as synchronized. The unchanged native buffer could therefore remain different from the domain forever, even after the reason for rejection disappeared. The production hunk only moves version acknowledgement into the success branch. Existing error notification remains. No forced Save, background retry loop/timer, stable-ID architecture, rename reservation or watcher change is added.

## New real failure and recovery fixture
TestFileEditorFailedSyncRetainsVersionAndRetries mounts the actual editor and changes its public native Buffer before domain synchronization. A disposable file is then made invalid UTF8; real Editor.Reload sets Content.Lossy, and real SetText rejects the pending edit with ErrReadOnly during actual rendering. The test verifies the real error toast, native text/Undo retained, domain text still the lossy disk snapshot, and invalid bytes unchanged. Restoring valid UTF8 through real Reload makes the next normal render a legitimate retry. It must synchronize the original pending native text and only then acknowledge the version.

Baseline red-states.jsonl records the wrong acknowledgement and the failed retry: domain remained `restored\n` while native Buffer retained `pending native\n`. After the hunk, exact race count2 passes (2.616s). Initial fixture compilation typo used Toast.Message rather than its actual Body field; fixture-build-error.jsonl is retained separately and is not counted as a behavioral RED.

This is a controlled renderer/domain synchronization seam, not an OS key-order reproduction. The fixture deliberately does not run the watcher's syncDocument callback; whether that separate path can overwrite an unobserved local draft remains unverified here. Normal debounce autosave after successful synchronization is unchanged. If a document remains read-only, subsequent ordinary frames may still report the same error; this patch does not add a deduplication policy and does not promise persistent-error UI behavior was exhaustively tested.

Source contracts: Editor.SetText returns ErrReadOnly without mutation for noneditable content (internal/workspace/editor.go164 onward); FileContent.Editable excludes Lossy (files.go138); LoadFile detects invalid UTF8 (files.go239 onward). The native versions entry is only an acknowledgement and must not advance on that real rejection.

Final same-source focused race:4 exact top-level tests PASS,0 skip/failure,2.967s; UI vet and git diff --check PASS. proof.json records actual target hitsets and log SHA256s. No full UI suite, hosted capture, publication or integration has been performed for this candidate.
