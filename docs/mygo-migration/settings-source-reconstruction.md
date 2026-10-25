# Native Settings reconstruction, 2026-10-25

Source: Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, MyGo
`9c69cf829ece196c5b3a3601f1c7bbe70526376f`.

This is implementation and mounted-control evidence, not a native visual parity
PASS. The same-run 1470×922 screenshots in `probes/native-f61e00f` were inspected
as actual images. They are unchanged. The next hosted native capture must assess
the new implementation; font shaping/raster differences remain separately tracked
in `scripts/font-fidelity/findings.md`. No display-sandbox workaround or global
font/geometry scaling was used.

## Source structure restored

- `app/settings.rs:5926–6410`: Appearance now has Theme & Background, one
  mode-specific Theme selector, Language, conditional background preview/tuning,
  Type & Density with interface/terminal preview tiles, and Layout. Stored family,
  palette and locale identifiers are mapped to source human names. Segments support
  click, arrows, Home/End and native focus. Filterable native popovers support
  original-index selection, no matches, arrows, Enter, Escape and dismissal.
- `app/settings.rs:4672–4790` and `commands.rs`: Shortcuts shows the advertised
  primary chord, not every alias or palette-only command. Session slots use one
  range. Composer conventions are a separate board using the current send mode.
- `app/settings.rs:5196–5615`: Agent uses Behavior, Session Titles and Session
  Defaults. Existing model selectors and RPC actions remain real. Observed retry-in-progress exposes the real Abort Retry RPC; it is not
  cleared optimistically. Unknown retry
  remains explicit, not the source's invented initial enabled switch (D04).
- `app/worktrees.rs:870–978` and `app.rs:1446–1482`: Worktrees has Manage,
  Location and Creation boards. Manage navigates to the real worktree page; the
  two 240px inline fields persist trimmed, nonempty edits. Empty intermediate
  input does not replace the last valid preference. Failed persistence retains
  the editable input and does not adopt an unpersisted setup policy (D05).
- `app/settings.rs:4931–5192`: Runtime has Process, optional background processes
  and unboxed Recent stderr. RPC patches uses the persisted host action, disabled
  while saving, without pretending to restart/revert an installed patch. Extra
  migration-only dependency/installation cards no longer displace source content;
  their backend utilities remain intact.
- `app/settings.rs:5618–5750`, `app/temp_ui.rs`: Privacy uses the source retention
  labels and named presets, then real temporary and AI scratch file snapshots,
  entry metadata, counts and scoped cleanup actions. The existing confirmation
  and active/recent-file protections remain intact.
- `app/settings.rs:991–1055`, `helpers.rs:607–664`: About restores the app-icon
  header, project banner and headerless version group. It truthfully names MyGo
  and Go, never GPUI. Unconfigured updates do not add an unrelated card.
- `app/settings.rs:2275–2388,2391–2551,3148–3641`: Providers has a pinned search/
  count/refresh/add toolbar and three-column branded cards, source-sized badges,
  counts and endpoints. Direct API-key and advertised login controls reach the
  existing secret/RPC flows. Configure/Remove apply only to actual custom
  configuration; authenticated/credentialed cards expose Disconnect. Unknown
  provider capabilities do not invent authentication methods. The real session
  projects Connecting, browser/device waiting, success, error, timeout and
  cancellation into the matching card. Open/copy/cancel/retry reach existing
  host actions; terminal dismissal is presentation-only and cannot manufacture
  authentication state. Retry requires the method still be advertised.

## Concrete mounted regressions

`internal/ui/settings_pages_test.go` covers mapped persistence, source card bounds,
keyboard appearance selection, active/inactive palette independence, filtered
native selector lifecycle, primary shortcut references and composer conventions,
inline worktree persistence/navigation/failure, RPC patch pending semantics,
Privacy list projection, provider grid geometry and advertised auth dispatch,
truthful About identity, provider-card auth lifecycle and scoped retry actions.

`TestSettingsReconstructionMeasuredSourceCardGeometry` checks the source native
Appearance, Shortcuts and Worktrees boards against measured screenshots within
2 logical pixels. The bottom Composer board is clipped by the 922px viewport,
not asserted to have an invented visible lower edge. This is a useful geometry
regression, not pixel equivalence or native OS execution proof.

Canonical focused race passed using the mandatory source-copy/overlay wrapper:

```
scripts/go-native.sh test -race ./internal/ui -run 'TestAppearanceSource|TestShortcutsSource|TestSource|TestSettingsReconstruction|TestNativeAgentRetry|TestNativeRuntime|TestSettingsSource|TestGeneralMeasured|TestNativeProvider|TestBackground' -count=1
```

## Explicit remaining source discrepancies

- `.manifest` is not a provider. Rust `providers.rs:587–619` counts object-valued
  members in every `*.json`, including `.manifest.json`; `settings.rs:2635–2651`
  then promotes every count key to a built-in provider with an invented API-key
  capability. That is why the old native reference contains `.manifest` with
  “42 in catalog.” The Go live metadata join does not promote this manifest.
  Excluding it is a source-defect correction, not a matching-pixels claim.
- Catalog-only future providers are now merged without invented API-key or
  OAuth capabilities. `TestProviderCatalogOnlyFutureProviderKeepsUnknownCapabilities`
  constructs a real local package/catalog fixture and fails if a valid provider
  disappears, its model count is lost, or unreported capability is invented.
  Live-RPC unknown providers likewise remain visible without an API-key action
  (`TestSourceProviderUnknownLiveCapabilitiesStayUnknown`). This independent
  functional fix does not turn the manifest correction into visual parity.
- Retry remains D04. Auto-compaction/queue values are also never fabricated when
  the process has not reported them. About's rendering-engine identity is an
  intentional truthful implementation difference.
- Source-native browser/device/error/terminal auth state capture, all long-scroll
  pages, tiny-window clipping, and new same-state
  screenshots remain independent verification gates. Do not infer those passes
  from control presence, fixture data or initial-page geometry.

## Checkpoint ownership

Reconstruction source: `internal/ui/settings.go`, `background.go`,
`native_settings.go`, `runtime_view.go`, `providers.go`, and new
`settings_controls.go`, `settings_appearance.go`, `settings_shortcuts.go`,
`settings_agent.go`, `settings_misc.go`, `settings_providers.go`. Existing
`settings_general.go` / source layout helpers are reused unchanged. Tests:
`settings_pages_test.go`, adapted `providers_test.go`, and the catalog-only
`internal/workbench/provider_catalog_gap_test.go` regression. Controller snapshot,
RPC-patch persistence, auth-session reset, modal guards, worktree propagation and
system-locale integration are owned and verified separately.
