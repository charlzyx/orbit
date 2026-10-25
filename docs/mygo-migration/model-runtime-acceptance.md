# Model catalog refresh and session-birth defaults

Functional **PARTIAL**; visual/native OS **UNTESTED**. These assertions close
specific source mismatches, not the entire Models/settings feature.

Pinned Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`:
- `app/settings.rs::models_refresh` and `restart_pi_for_catalog` (1369–1430):
  invoke `pi update --models`, reject busy/streaming/duplicates, and only on
  success restart/resume the active session so its catalog reloads.
- `app/runtime.rs::default_model_entry` / `apply_default_model` (193–304):
  provider-less default uses the first matching catalog entry; missing catalog
  stays armed; unknown model fails soft; push model once, wait until live, then
  apply only a supported thinking level. Manual model/thinking choice disarms.
- `app.rs::catalog_thinking_levels` (1093): reasoning=false supports only Off;
  null explicitly excludes a level; XHigh/Max require explicit non-null mapping.
- `session_defaults.rs::cli_args`: fresh process spawn still passes defaults
  through source-compatible CLI arguments; resumed sessions are not overwritten.

## Corrected behavior and tests

Native configured refresh captures executable/home/Pi directory/workspace and
child environment before I/O. An explicit environment never inherits ambient
credentials. It augments GUI PATH using that same environment, executes literal
arguments without a shell, bounds each output to1MiB, redacts known environment
secrets, cancels the owned process group, and imposes a2minute deadline.

`TestModelCatalogConfiguredExecutableAndIsolatedEnvironment`,
`TestModelCatalogConfiguredFailureRedactsEnvironmentSecrets`,
`TestModelCatalogConfiguredTimeoutAndInvalidExecutable`, and
`TestModelCatalogConfiguredOutputIsBounded` use disposable shell fixtures, not
real update downloads or provider calls. Host tests separately cover exclusion,
failure/cancel/duplicate and successful child restart/resume/catalog requery.

The output-bound test first reproduced2.1MB on both streams. Embedded
`bytes.Buffer.ReadFrom` let `io.Copy` bypass the custom `Write` limit. Explicit
`ReadFrom` now routes through bounded `Write`; the focused regression
`TestRuntimeOutputReaderFromCannotBypassLimit` keeps that optimization covered.

The agent's I/O-free birth-default state machine waits for session ID/catalog,
tracks one-shot model/thinking pushes, and skips unsupported values. Manual
intent synchronously invalidates queued default commands; worker serialization
prevents an older default overtaking the later manual choice.

Tests: `TestBirthDefaultWaitsForIdentityCatalogAndLiveModel`,
`TestBirthDefaultSkipsRemovedUnsupportedAlreadyLiveAndUnset`,
`TestBirthDefaultProviderlessSelectsFirstSourceCatalogMatch`,
`TestCatalogThinkingLevelsSourceCapabilitySemantics`, and
`TestManualModelIntentDisarmsPendingDefaultWhileNotReady`.
`TestBirthDefaultRealChildOrdersThinkingAfterModelConfirmation` launches a real
local protocol fixture process, verifies exactly one model and thinking command,
and requires an intervening authoritative state query. It is not provider inference.

[Current file hash manifest and executed commands](model-runtime-evidence.json).
Full agent race also ran the credential-free official Pi1.4 protocol tests.
No actual authenticated model-catalog update, provider inference or native
Windows/macOS execution is represented by these local fixture results.
