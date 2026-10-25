# Settings source assertion crosswalk

These are 23 named mounted/local-fixture assertions, not feature passes.
Pinned Rust: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`; MyGo: `9c69cf829ece`.
Evidence: [machine-readable source/step/expectation/test crosswalk](settings-source-crosswalk.json).
The scoped source manifest was unchanged before/after canonical race and vet.
Original source screenshots were inspected; updated native/OS sequences remain unverified.

Each case below also records its exact untested native/scroll/error boundaries in the JSON.

## P01 · TestSettingsSourceSearchIndexAndSectionNavigation

Source: [settings_search_hits](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L412), [settings_search_results](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L766)
- Step: Type pid into native Settings search; select Process ID.
- Expected: The55-entry source search index keeps category navigation visible, opens Runtime and clears query.
- Still untested: Full search keyboard traversal and non-Latin input on native OS.

## P01 · TestSettingsPageScrollRetained

Source: [settings_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- Step: Scroll Shortcuts, leave for eight frames, return.
- Expected: Per-page scroll offset survives view eviction.
- Still untested: All long pages, tiny-window clipping and source-native scroll physics.

## P01 · TestSettingsSourceHeaderAndGeneralRows

Source: [settings_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- Step: Inspect General and activate notification sound by pointer then Space.
- Expected: Fixed header and source General rows; unavailable updater adds no card; pointer/Space persist the preference.
- Still untested: OS notification permission and sound playback, all translations.

## P02 · TestNativeRuntimeUsesRealSnapshotAndStoppedUnknowns

Source: [runtime_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4931)
- Step: Mount real-shaped running then stopped process snapshots.
- Expected: PID/executable/uptime/stderr come from the snapshot; stopped fields never fabricate a connected process.
- Still untested: Real process Start/Stop/Restart, parked-process lifecycle and native fault injection.

## P02, U06, U07 · TestSourceRuntimePatchAndPrivacyControls

Source: [runtime_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4931), [privacy_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5618), [temporary_files_section](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/temp_ui.rs#L79)
- Step: Click RPC patches, repeat while save pending; render scoped temp entries; change retention by switch and filtered Enter.
- Expected: Dispatch actual patch setter without optimistic policy mutation; pending control blocks duplicate action; real temp metadata and mapped retention values persist.
- Still untested: This test does not write patch config or delete files; host lifecycle/cleanup safety tests and native picker captures remain separate.

## P04 · TestNativeAgentRetryIsUnknownUntilConfirmed

Source: [agent_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5196), [toggle_auto_retry](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5781)
- Step: Mount unknown retry policy, choose Enable, then inject a confirmed enabled state.
- Expected: Unknown is not advertised as enabled; only setter acknowledgement establishes the enabled value.
- Still untested: D04 intentionally differs from Rust initial enabled assumption; real provider transient-error retries and native menu pixels.

## P04 · TestSourceRetryStatusUsesObservedState

Source: [agent_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5196), [abort_retry](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5828)
- Step: Mount actual retry-in-progress and click Abort retry; then project idle.
- Expected: Dispatch AbortRetry without clearing observed state optimistically; idle removes the button.
- Still untested: Real RPC child acknowledgement, delayed retry and source-native error sequence.

## P18 · TestAppearanceSourceGroupsMappedChoicesAndPersistence

Source: [appearance_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926), [type_preview](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6241), [apply_settings_select](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7157)
- Step: Click Light, press Right; choose mapped active Theme and filter Interface font to Inter then Enter.
- Expected: Source groups/previews and human names; mode/font persist; changing active palette never overwrites inactive palette.
- Still untested: OS System light/dark events, actual font raster equivalence, every preset and failure/relaunch sequence.

## P18 · TestSourceSettingsPickerEmptyEscapeAndReopen

Source: [settings_select_popup](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907), [settings_select_step](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7104), [settings_select_confirm](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7124)
- Step: Open selector, filter no matches, press Enter/Escape, immediately reopen and select a filtered option.
- Expected: No-match Enter is a no-op; Escape closes; reopen clears stale filter; selected option uses original index.
- Still untested: Native OS focus restoration and all outside-dismiss/scroll placements.

## P18, P20, W43 · TestSettingsReconstructionMeasuredSourceCardGeometry

Source: [appearance_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926), [shortcut_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4672), [worktree_settings_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L870)
- Step: Mount1470x922 Appearance, Shortcuts and Worktrees fixtures.
- Expected: Named group bounds stay within2DIP of measured actual Rust screenshots; Composer lower edge is viewport-clipped.
- Still untested: Pixel equality, light theme, nondefault font/density, long scroll and matched new native captures. Geometry alone is not visual PASS.

## P19 · TestSourceBackgroundBlurGeometryPercentAndSteppedKeyboard

Source: [blur_slider](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6447)
- Step: Mount tuning with sigma3; click midpoint, Right, End while image worker is busy.
- Expected: 184px track/20px control and50% label; half-sigma keyboard step coalesces into pending tuning; End shows100%.
- Still untested: Actual native drag/cursor, background chooser OS dialog and renderer fidelity.

## P19 · TestBackgroundChooseCoalesceFadeReuseFailureAndClear

Source: [background_choose](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6569), [blur_slider](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6447)
- Step: Choose a real temporary PNG, coalesce tuning, change fade, inject failure and clear.
- Expected: Existing asynchronous photo store/cache preserves image/tuning semantics and reports failures; source-only control changes do not replace renderer.
- Still untested: Native file picker cancellation and unsupported real user image formats, GPU rendering.

## P20 · TestShortcutsSourceAdvertisedRowsAndComposerModes

Source: [shortcut_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4672), [shortcut_row](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4756)
- Step: Inspect advertised chords, compact session range and Composer section; change send mode.
- Expected: Primary source chord rather than aliases/palette-only commands; Composer labels follow actual preference.
- Still untested: Actual OS chord dispatch from terminal/editor/modal focus and non-English help layout.

## P05, P01 · TestSourceProviderGridCapabilitiesFilteringAndAboutIdentity

Source: [provider_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391), [provider_card](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3148), [settings_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- Step: Inspect four providers, click advertised login/API-key actions, filter, then visit About.
- Expected: Three296px columns with12px gaps; real capability dispatch; only custom entries get Configure; About says MyGo, not GPUI.
- Still untested: Long catalog scroll, complete quota popup, source About difference remains intentional, new native matched pixels.

## P05 · TestSourceProviderUnknownLiveCapabilitiesStayUnknown

Source: [provider_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391), [provider_card](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3148)
- Step: Expose a valid provider only through live ModelCatalog.
- Expected: Card remains visible without an invented API-key control.
- Still untested: Future Pi schema changes and true unknown providers from installed runtime.

## P05 · TestSourceProviderAdvertisedEmptyMethodsDoNotFallBackToStaticCapability

Source: [provider_card](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3148)
- Step: Expose an explicit runtime provider with no auth methods despite static API-key/OAuth metadata.
- Expected: Show source no-methods explanation; do not silently replace explicit empty capability with guessed auth action.
- Still untested: Runtime capability refresh races and real provider permissions.

## P05 · TestProviderCatalogOnlyFutureProviderKeepsUnknownCapabilities

Source: [provider_views](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2604), [read_catalog_counts](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/providers.rs#L587)
- Step: Load a fixture Pi package with valid catalog-only provider and separate manifest.
- Expected: Retain valid provider/count with unknown auth capabilities; never promote.manifest to a provider (D09 source-defect correction).
- Still untested: Real installed future package versions; exclusion is not pixel parity.

## P06 · TestSourceProviderEditorModalChipsKeyboardAndPersistence

Source: [provider_editor_layer](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145), [provider_editor_confirm](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7998), [provider_save](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L8027)
- Step: Open Add at800x560; type ID and Enter invalid form; choose API chip, fill fixture endpoint/models/key and save.
- Expected: Source header/body/footer and API chips; Enter validates instead of submitting composer; footer reachable; actual saved config has chosen API and two models; key never paints.
- Still untested: Same-run Rust modal screenshot, all native Tab orders,600px-height source clipping, every API family and relaunch.

## P06 · TestNativeProviderSaveFailureKeepsDraft

Source: [provider_save](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L8027)
- Step: Make Pi config parent a regular file and save a provider draft.
- Expected: Write failure retains editable fields and dummy secret with inline error, not false success.
- Still untested: Real permission-denied OS dialog and recovery/retry comparison with Rust.

## P06, P07 · TestSourceProviderCredentialModalSecretAndAlternativeMethod

Source: [provider_key_layer](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926), [provider_key_open](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7712), [apply_provider_action](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3730)
- Step: Open API key at800x560, type dummy key, switch to advertised browser login; reopen cookie input and Escape.
- Expected: 460px source dialog; display name/id; password masking and focus; alternative login uses actual advertised method and wipes draft; Escape wipes cookie.
- Still untested: Real OAuth issuer/credential transmission, device code OS clipboard, native screenshots and source terminal fallback.

## P06 · TestSourceCredentialSaveDoesNotDiscardReopenedDraft

Source: [provider_key_save](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7772)
- Step: Save dummy key asynchronously, dismiss and reopen same provider before UI applies completion; type a new draft.
- Expected: Completed persistence refreshes actual runtime configuration but cannot close or discard the later draft.
- Still untested: Rust synchronous save has no identical callback race; real OS crash/recovery and credential scope still separate.

## P07 · TestSourceProviderAuthenticationCardLifecycle

Source: [provider_auth_section](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2760)
- Step: Project waiting device session; open/copy/cancel; project failure, retry/dismiss; project new succeeded session.
- Expected: Status/actions stay inside matching card; current provider/method dispatch; cancel/dismiss never fabricate backend status; old dismissal cannot hide new session.
- Still untested: Real browser opening is host-tested separately; native clipboard, credentials, timeout timing and actual source-state screenshots remain unverified.

## W43 · TestSourceWorktreeInlinePreferencesNavigationAndFailure

Source: [worktree_settings_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L870)
- Step: Edit trimmed nonempty root, temporarily empty input, open Manage and inject config write failure through the setup-policy switch.
- Expected: Persist valid preferences; empty intermediate input preserves last valid policy; failed toggle does not adopt unpersisted consent; Manage navigates.
- Still untested: Setup-script input is not exercised by this test. Actual subsequent creation/setup requires independent host test; per-repository overrides and native focus/scroll.

No acceptance-builder or full-feature status was changed by this crosswalk.
The P02/U06/U07 and P06/P07 aliases reference the same named test, not extra independent proof.
About has no dedicated P row; its truthful identity check is an auxiliary P01 navigation assertion.
W43/U06/U07 stay under their real existing IDs, not relabeled as Settings P rows.
