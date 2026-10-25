# Native heartbeat startup ordering correction

The exact `ce4d49a` migration CI run `40216553377` passed its aggregate and pinned Rust reference jobs, but the Linux native capture failed closed before clicking Details. The downloaded 25-file artifact has SHA256 `0eaaf126c26f656ffd8dde343c7bb0ec2bc4bffb3f002418834549a60c3fcd81`. Its unmodified frame shows `Loading local usage…`; Details is visually present but absent from whole-window OCR. This is neither a successful Usage capture nor proof of the user's Mac picker symptom.

## Concrete source ordering defect

`internal/ui/run.go` started the retained heartbeat before `mygo.App.Run()`. The timer's first owner dispatch uses `mygo.RunOnMain` after 90 ms. In pinned MyGo `9c69cf829ece`:

- `internal/linux/app.go::Init` loads native function pointers before `initCallbacks()`;
- `Backend.Signal` sets its atomic `signaled` latch then calls `gIdleAddFull(... dispatchCallback ...)`;
- before callback initialization, `dispatchCallback` can still be zero;
- native job `orbit.log` has `GLib-CRITICAL ... idle_add_full: assertion 'function != NULL' failed` at startup.

That interleaving can leave a pending main-loop dispatch unsignaled. The source and log match this failure mechanism, but the CI artifact alone does not establish every causal step or reproduce the user's Mac hang.

## Narrow correction

Register the existing initial window/setup callback through `App.WhenReady`, and start the retained host only after that callback completes. A `sync.Once` guards the host lifetime, not individual windows. Reopening a window does not create another heartbeat. There is no MyGo core mutation, reduced compatibility guard, off-main widget update, or guessed native click.

## Evidence boundaries

- Virtual-time lifecycle regression: one second before readiness produces no heartbeat dispatch; initial setup precedes the first tick; duplicate readiness cannot duplicate the host; Quit stops it.
- Existing Controller window creation/close/reopen ownership and windowless maintenance regressions are retained.
- A source-wiring regression binds that lifecycle helper to the actual `Run` entrypoint and rejects a restored direct pre-initialization start.
- First test candidate used a plain diagnostic counter and race detection caught the **test** data race. The final test uses `atomic.Int32`; the production correction did not change to hide that failure.
- Final aggregate/vet/arm64 and native same-commit rerun results must be recorded separately. No native pass is inferred from these headless checks.

The capture harness also needs a completed semantic Usage frame before attempting the exact unique Details OCR target. Its no-input-on-failure rule must remain intact. `.4` is unchanged; this is unreleased development source.

## Final local gate (17:56 UTC)

The exact frozen correction source passed `scripts/go-native.sh test -race ./...` (UI 371.505s), `vet ./...`, and no-cgo darwin/arm64 build; the sequential chain exited 0. Native lifecycle focused race passed five repetitions, actual Run wiring passed ten. The integrated capture scripts passed 11 Usage and 53 workbench Python regressions (2 platform skips). Code whitespace checking passed with only raw OCR TSV fixtures excluded: their terminal tabs encode empty final fields and are preserved byte-for-byte with manifest hashes. This is not a relaxation of OCR confidence, readiness, or compatibility enforcement.

The prior ce4 native failure remains a recorded failure. The newly published correction still requires its own exact-commit CI/native captures; no new installer was produced and `.4` was not changed.
