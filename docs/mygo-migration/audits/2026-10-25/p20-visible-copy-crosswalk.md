# P20 visible copy crosswalk — bounded first batch

Reference: pinned Rust 299a32bc87fe3d4851bdb902d9c03dcf9bf99cac; Go clean
ade11f3879abd04f0e26630b236466cf9886787e plus read-only wave4-1808 working
snapshot inspected 2026-10-25 18:49 UTC. This is a prioritized source audit,
not an exhaustive string census or full P20 acceptance. Changes below are
unreleased, not present in delivered .4.

## Confirmed daily-use gaps

| Go location (inspected snapshot) | Visible hardcoded copy | Source/catalog contract | Disposition |
| --- | --- | --- | --- |
| chat.go:79–81 | Thinking / Thought process | transcript.thinking exists; distinct expanded wording requires source mapping | deferred; frozen wave4 |
| chat.go:117 | Jump to latest | message_scroller.jump_to_latest | deferred; frozen wave4 |
| chat.go:233–244 | Supervised / Auto-accept edits / Full access | access.supervised / auto_accept_edits / full_access | deferred; preserve protocol values |
| chat.go:356 | Build | workflow.build | deferred; distinguish workflow ID from title |
| chat.go:387 | Find in transcript… placeholder | search.find_in_transcript | deferred; label identity is not placeholder |
| composer_add_menu.go:20–26 | Attach files… / Paste image | composer.attach_file exists (singular); paste requires semantic check | deferred; frozen wave4 |
| workspace_menu.go:105–115,196–197 | Copy path / Icon & color / Remove from sidebar / Reset | sidebar.copy_path / icon_and_color / remove_from_sidebar / reset; Rust app/sidebar.rs:1546–1602 | deferred; frozen wave4 |
| workspace_menu.go:108,119,123 | Worktrees / New task here / Open in file manager | existing Go action extensions must retain target semantics, no blind key substitution | deferred |
| settings_agent.go:64–71 | Unknown | context_meter.unknown canonical generic status | first batch |
| settings_agent.go:87–98,139–141 | retry unknown/unsupported/enabled/disabled/updating and enable/disable | Rust app/settings.rs:5215–5250 uses localized labels but optimistic remembered retry switch; Go deliberately shows authoritative/unknown state | first batch; separate Go-only status keys, retain truthful semantics |
| session_menu_actions.go:19 | Rename… / Copy session ID | Go convenience additions, not Rust sidebar menu; usage.ctx_copy_session_id is exact canonical copy label | first batch; rename gets namespaced key preserving English ellipsis |
| settings.go:285–299 | Name / user / project / Cancel / Save | config dialog display separate from stored scope values; existing settings cancel/save keys | deferred |
| settings.go:319–325 | Report a bug / privacy warning / Copy diagnostics / Open issue tracker | bug_report/settings catalogs exist but Go diagnostics flow differs | deferred; do not translate or reinterpret diagnostic payload |
| settings_providers.go:340 | Connect pi to discover available sign-in methods. | Go runtime-discovery state, not provider-returned error | deferred; P03 setup author owns adjacent work |

Separate inherited catalog gap: `zh-CN.yml:486–487` contains literal English
`Archive session` / `Unarchive session`; those callsites already use Tr. This
is missing source translation, not a Go key-routing defect. The first batch
leaves pinned source catalogs unchanged, so this remains visible/open.

Non-targets: asset paths, action Kind, runtime protocol values, persisted values,
font/product names, arbitrary model/provider identifiers, user text, and raw
provider/RPC errors. A UI `.Label` can be test/accessibility identity, not painted
copy; this batch does not mechanically replace labels. Existing literal-label
selectors remain stable.

## Minimal implementation contract

Use original locale keys when semantics match exactly. Keep original YAML
catalogs byte-for-byte unchanged. Add a separate Go-only `go.*` supplement merged
by Locale after source parsing without overwriting source keys. English is the
canonical fallback; Chinese overrides the first batch. Other supported languages
explicitly receive English fallback for new Go-only keys, while existing source
keys retain their original translations. Unsupported locale loading still errors;
Translate(nil, English, key, nil) remains the UI fallback. No locale registry,
protocol, runtime behavior or confirmation changes.

## Verification and limits

Actual mounted Chinese unknown status and session menu fail before production
changes (record terminal log), then pass with exact command/value assertions.
Cover retry pending/known/unsupported, locale reload to English, unchanged raw
provider error, original catalog non-overwrite, unknown-locale fallback and all
supported-language key resolution. Existing English retry/menu tests must stay
valid. No physical Mac capture, native screen-reader/Tab review, or text-width
parity is claimed. Remaining table rows stay open for a later non-frozen batch.
