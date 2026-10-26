# Commit-only model selection — bounded P2 #88 candidate

Source: upstream `118aa2ecac871df37e4fa56bf9a2be4fa853e70d` (#88),
`crates/orbit-pi/src/commit_model.rs`, `git_panel/mod.rs:2612–2940,7961`,
`commit_message.rs:49–96`. Based on public `ecaa79d581b4c712d0de98e9bd11a333e98e205a`
(tree `05882441c24ad879aac0b94b68633f31b5cad9f0`; equivalent local base `5dbc486`).
This candidate is not the complete upstream #88/#87 port or full UI parity.

## Behavior

- A native anchored, searchable Commit model picker reuses the workbench's
  existing Popover/TextInput/Scroll/Button primitives. It does not dispatch the
  chat model selector's action or change chat model preferences/recents.
- `commit-model.json` in Orbit's private preference root stores
  `{"model":{"provider":"provider","id":"model"}}`; null follows the active
  session. Atomic writes use mode 0600. Session/workspace changes retain this
  independent preference; workspace/tab/modal transitions close the picker.
- Resolve explicit identity first, otherwise active session, otherwise no forced
  model (`set_model` omitted). Explicit identity absent from catalog stays explicit
  with unknown context window; it never silently chooses the chat model.
- Like Rust `pick_commit_model`, selection takes effect in memory before persist.
  Unlike silent upstream best-effort persistence, failure is visibly reported as
  applying only for this run / not saved. Selection and dismissal return native
  focus to the commit message editor. Bad JSON returns a visible read error and
  follows the session, rather than treating corrupted bytes as a model identity.
- The generation admission captures resolved provider/id/context-window, scope,
  existing workspace generation and generator closure. Busy locks selection;
  cancellation/workspace change rejects stale results. Catalog refresh cannot
  rewrite a saved explicit identity. Issue/PR generation still uses its original
  active-session configuration, and chat receives no commit `set_model`.
- The existing fresh tool-free, no-session Pi JSONL child is retained. Model/auth/
  process failure remains an explicitly marked editable heuristic; cancellation
  does not become a successful draft. No commit/index/push action is added.

## Necessary prompt-budget dependency and limits

The selected context window now reaches commit prompt construction only. Unknown
window uses 256 KiB; known window uses saturating `tokens * 3 / 2`, clamped to
24 KiB–2 MiB. After metadata + 8 KiB reserve (with 12 KiB floor), shared diff and
original pools receive 50% / 30%; tiny original shares are transferred to diffs.
Per-file caps are 64 / 48 KiB, with UTF-8-safe truncation markers counted inside
allocations. Diffs consume one common pool, not 96 KiB multiplied by file count.

Additional Go guard: system + user including metadata and final reminder cannot
exceed the selected byte ceiling. For an extreme metadata/path list, that means
explicitly truncating metadata too, rather than claiming a complete file list.
This is an estimator, NOT a tokenizer/context-fit guarantee: the source 24 KiB
floor may exceed tiny model windows. No automatic alternate model or retry.

This narrow batch does **not** port upstream #87's deep-hunk source excerpts,
first/last-hunk preservation, response-body changes, regeneration previous/attempt
prompting, untracked 1 MiB policy or every source picker glyph/layout detail.
Original excerpts still begin at file start and untracked reading keeps the
existing Go limit. Native pixel comparison and actual provider execution are
NOT_RUN, not implied by mounted tests or cross-compilation.

## New evidence

Historical evidence label: `orbit-commit-model-evidence-2038`.
`budget-red.jsonl`: actual eight synthetic untracked-file prompt was 788,744 bytes.
`rpc-red-settled.jsonl`: actual offline JSONL child sent only prompt/get-last-text
although the private explicit model was persisted (missing set_model).
The earlier `rpc-red.jsonl` captured an unsettled startup fixture and is NOT the
behavioral RED; it is retained explicitly rather than reused as proof.

Focused tests use disposable homes/repos and a real local JSONL peer process,
not a real model/provider. They cover private persistence/errors, exact provider+
ID lookup, unknown/follow/default, child command ordering, unchanged active chat
and Issue/PR selection, shared budgets and UTF-8 boundaries, explicit fallback,
Busy/cancel/workspace stale-result guards, and mounted native search/keyboard/
selection/reveal/Escape/outside/repeated-open/focus/write-failure behavior.
Final exact executed hitsets and raw-log hashes are in the accompanying proof.
No full suite, hosted native action, release or remote write is authorized here.

## Rebuilt final review deltas (22:30 recovery)

The executor reset preserved 9194274 and historical logs, but not the complete
69f/8ed source objects. This candidate is a newly reconstructed tree, NOT a
byte-identical restoration of patch 66dcbf or tree edd914. Historical PASS remains
historical; new executed hitsets/hashes are recorded separately.

Successful preference load/selection no longer clears an unrelated Git error.
Preference failures still use the existing error channel; no new warning system
or popup was added. A follow-active model with nonempty ID but empty provider is
passed unchanged to the existing isolated child set_model path, rather than
silently selecting Pi default. The synthetic JSONL peer rejects that identity to
prove the normal explicit editable fallback is shown without prompting another
model. Persisted explicit choices still require both provider and ID.

Budget allocation difference: Go redistributes remaining shared bytes over
remaining files; Rust #88 allocates fixed initial per-file shares. Both impose a
shared total, but these are not identical excerpt-selection algorithms.

The only layout production delta is Wrap on the existing commit controls row.
Fonts, labels, model max width, action keys and handlers are unchanged. A fresh
full-App CPU fixture checks 1470/1200/1024 widths with default sidebar248 and
1470/1200 with sidebar400: complete labels, 2D non-overlap, actual action clicks,
held-generation picker disable, anchored popup, Escape/outside returning typing
to CommitMessage, and unchanged synthetic file/index/no HEAD. Initial fresh RED
reproduced Commit & push text41/104px and65/104px at the two narrow layouts.
No hosted native interaction or pixel equivalence is inferred from CPU images.
