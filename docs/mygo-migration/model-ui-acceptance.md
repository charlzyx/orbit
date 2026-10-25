# Native model catalog and picker acceptance

Source-first procedure written before implementation. Frozen Rust reference:
`299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Scope: P08, C06 and relevant
S03/S04 model-selection paths. This supplement never claims whole-row parity,
matched pixels or actual OS keyboard/IME execution from CPU fixture tests.

## Source-derived procedures

1. **Catalog truth, loading and errors.** Start without a catalog, then supply an
   actual Pi `get_available_models` result. Show no invented model/provider rows;
   loading, disconnected and returned-empty states remain distinct. Refresh errors
   leave the last returned catalog intact. Source: app/events.rs catalog reducer,
   app/settings.rs model_rows/model_toolbar. Catalog capabilities and context come
   from the RPC payload, not a hard-coded model list.
2. **Models toolbar and grid.** Supply interleaved providers; group in first
   appearance order, preserving catalog order within groups. Three cards per row,
   provider glyph/name/count header, active-model board first. Search is trimmed,
   case-insensitive substring on name/id/provider. Provider and favorite filters
   combine with search. Count includes catalog total and catalog-present favorites,
   and says shown-of-total when any filter is active. Clearing each filter restores
   the unchanged catalog. Provider control is omitted with fewer than two providers.
3. **Identity and selection.** Two providers may expose the same model id. Cards and
   active-model menu identify exact provider+id; neither preemptively changes the
   selected state before Pi reports it. A card click and the active dropdown send
   one set_model command for the rendered row. They do not record Recents. Failure
   remains visible. Repeated clicks are discrete requests, not duplicate dispatch
   from a single native click. Source: settings.rs model_card/ActiveModel.
4. **Favorite isolation and persistence.** Clicking a card or picker star toggles
   exactly provider+id without selecting the model, changing the current session,
   recording recents or closing the popup. Repeat to unstar; reopen/relaunch to
   verify favorites.json compatibility. On persistence failure retain old confirmed
   state and report failure. Source: favorites.rs and model_selector.rs favorite_button.
5. **Picker grouping/scopes.** All scope pins Recents in MRU order, then Favorites
   in catalog order, then remaining provider groups, with no duplicate rows. Recents
   scope is flat in MRU order. Provider scope pins that provider's recents in catalog
   order, then remaining models, no provider group header. Favorites scope is flat
   catalog order. Search still applies in every scope; missing catalog entries never
   materialize. All, Recents/Favorites when stored entries exist, and provider chips
   share a single horizontal scroll. Favorite-current chip is present only when no
   favorites exist and the current model resolves. Source: ModelSelector::rows_for.
6. **Picker current identity.** The model picker compares ASCII-alphanumeric
   normalized id/name against state id/name, with case-insensitive provider match
   when a provider is reported. A different provider cannot appear selected. The
   settings cards deliberately retain exact identity semantics. Source:
   model_selector_match.rs is_model_selected.
7. **Picker keyboard/focus.** Open focuses its search and pins/scrolls current
   selection; up/down clamp at endpoints and skip headers. A nonempty changed query
   chooses its first row, zero results never submit, clearing query re-pins current.
   Scope/catalog change re-pins, but an unchanged snapshot never resets a deliberate
   arrow choice. Escape/outside-click dismiss without an RPC action. Host restores
   composer focus and suppresses the same outside-down/trigger-up reopening for
   200ms. Mouse moves own hover highlighting only after the 400ms suppression; a
   stationary pointer over a row during scrolling must not hijack the keyboard.
8. **Recents and repeat confirmation.** A composer model pick records provider+id
   before dispatch and caps MRU at five. Reselecting the leading model performs no
   persistence rewrite. Selecting closes the picker; subsequent Enter cannot send
   a second action. Source: pickers.rs toggle_model_selector/recent_models.rs.
9. **Thinking.** Open only when actual get_available_thinking_levels has more than
   one option. Render only reported levels, with source icons/display labels/hints;
   filtering includes localized hints. Choose the exact raw level, never normalized
   display text. Zero results and cancellation send nothing. Selection reflects Pi.
10. **Settings dropdowns.** Provider, active model, default model, default thinking
    and title model use searchable 360px-minimum popovers. Open pins the current
    original index, filter retains original indices, arrows clamp over visible rows,
    Enter chooses highlighted or first match, zero results do nothing. Escape and
    outside press close without applying. Active and provider options are catalog
    ordered; defaults/title start with the relevant Pi/session-default sentinel.
11. **Session defaults and title model.** Defaults write only Orbit's
    session-defaults.json, never Pi settings or the live session. Non-reasoning model
    supports off only. Reasoning base ladder is off/minimal/low/medium/high; explicit
    null map values remove levels, xhigh/max require explicit non-null mappings.
    Selecting another model retains thinking only if supported; Pi default clears
    model/provider/thinking together. Default-thinking is disabled when no catalog
    default resolves. Title model preserves auto-title enabled state and writes
    auto-title.json without selecting the live model. Source: app.rs
    catalog_thinking_levels and settings.rs apply_settings_select. Spawn/new-session
    application, stale-default validation and manual-choice precedence require
    separate manager integration evidence.
12. **Catalog update lifecycle.** One click runs the configured Pi's documented
    update-model command off-thread, not merely get_available_models. Busy/streaming
    refuses restart, duplicates are gated, failure remains actionable, successful
    completion safely restarts/resumes and refreshes catalog. Closing a picker/page
    does not fabricate cancellation of a launched external operation. Module shows
    host-provided status only. Source: settings.rs models_refresh/restart_pi_for_catalog.

## Explicit deviations / separate gates

- The Rust filtered-empty branch is unreachable because model_rows initializes the
  active-model board before checking sections.is_empty(). Go should show the intended
  no-matches notice while preserving that board, rather than silently render a blank grid.
- Source ignores preference write failures. Go retains confirmed preference state,
  reports write failures and does not silently overwrite malformed JSON.
- CPU layout/key/click tests are not matched native Rust pixels, real OS IME,
  accessibility, platform lifecycle or live authenticated inference evidence.
- Native placement/host focus restoration and guarded catalog-update runtime behavior
  require host integration tests; isolated projection tests do not prove those gates.

## Implemented evidence (2026-10-25)

`internal/modelui` now supplies the Models toolbar/body, searchable provider/active/
default/title selects, source-grouped composer model popup and reported-level thinking
popup. The host mounts these through `internal/ui/models.go`; it owns placement,
composer focus, catalog queries and configured update/restart. Shared capabilities
come from `agent.CatalogThinkingLevels`, not a second UI-only model rules table.

| Procedures | Executed coverage |
| --- | --- |
| 1–3 | `TestCatalogGroupsFiltersCountsAndExactIdentity`, `TestNativeModelsThreeColumnCardGeometry`, `TestNativeModelLoadingErrorAndEmptyHaveNoFabricatedRows`, `TestNativeStarsNeverSelectOrCloseAndCardsDispatchOnce` |
| 4 | `TestFavoritesPersistenceFailureAndNoSelection`, native star isolation tests |
| 5–6 | `TestPickerSourceGroupingScopesAndDedup`, `TestPickerMatchingNormalizationAndProviderGuard`, `TestPickerIdentityOnlyFoldsASCII`, `TestNativeHorizontalScopesKeepFavoriteCurrentReachable` |
| 7–8 | `TestPickerKeyboardClampsSkipsHeadersAndUnchangedSnapshot`, `TestSourceCenteredScrollAccountsForHeadersAndClamps`, `TestModelDispatchAndRecentsDoNotFabricateSelection`, `TestRecentsCapStaleModelsAndCancellation`, native search/scope/long-list/stationary-pointer tests |
| 9 | `TestThinkingPickerReportedLevelsAndHintSearch`, `TestNativePickerLongListKeyboardRevealAndThinkingGate` |
| 10 | `TestSearchableSettingsOriginalIndicesAndClamping`, native default-menu/original-index/no-match/outside/repeated-trigger tests |
| 11 | `TestThinkingCapabilitiesAndSourceDefaults`, `TestDefaultsPersistenceFailureAndTitleSelection` plus scoped host/manager default integration tests |
| 12 | `TestRefreshLifecycleGateAndSnapshotCopies`, `TestRunningLocksOnlyComposerPickerAndReloadCapsRecents`, `TestNativeRefreshGlyphAndReducedMotion`, scoped host configured-refresh tests |

- The native refresh control now uses the original accent refresh SVG and 700ms
  rotation, respecting both Orbit and desktop reduced motion. Its label stays
  `Update model list`; duplicate requests are ignored rather than reporting success.
- `TestLiveCredentialFreeCatalogProjection` starts pinned Pi 1.4.0 in disposable,
  offline state and sends only the model/thinking getters. Its real catalog was
  empty. The rendered empty state and no-selectability are verified; no prompt,
  authentication, inference, tool execution or provider catalog update is sent.
- `probes/modelui/verification.json` records final file/source hashes and test
  results. The six component PNGs are 1000×760 synthetic native CPU fixtures, not
  the canonical 1470×922 host/source pairing, and do not establish pixel parity.
- `model-visual-source-review.json` independently reviews all 29 inventory IDs
  V124–V152 against concrete rendered children and supplies concrete fixture
  inputs. It corrects invented card metadata, grouped settings options and title
  sentinel wording. It also lists uncovered render branches; 29 is not asserted
  to be an exhaustive count of all possible visual states.

Additional deliberate robustness differences: Go distinguishes loading/disconnected
readiness and retains actionable inline catalog/persistence errors; Rust's catalog
update failure is a transient toast. Neither difference may be counted as matched
source pixels. The original filtered-empty defect remains explicitly identified.

**Verdict:** scoped functional tests pass. Whole-feature parity, matching source
pixels, all platform lifecycle/IME/accessibility behavior and authenticated live
provider behavior remain separate, unestablished gates.
