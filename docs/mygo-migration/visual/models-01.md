# Audited visual candidates: models

Numbered candidates, not validated screenshot recipes. State validity, fixture execution, capture, comparison and parity are separate gates.

## V124 · Settings models baseline
- Features: P01, P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_rows:L1433](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1433)
- State validity: **VALID_SOURCE_BRANCH**. Source-render branch exists. Concrete fixture specified; actual paired capture still unverified. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Search models…; 5 model(s); All providers dropdown; Update model list; Favorites toggle; Active model board with Second · Second Same; First provider heading and three cards; Second provider heading and two cards; Current card Active badge; Optional context badges; no context badge on First Same.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Settings models baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V124. Overrides: `{"surface": "Settings > Models"}`.
- Planned pair: `../probes/visual/V124/rust.png` ↔ `../probes/visual/V124/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V125 · Models empty catalog
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_rows:L1433](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1433)
- State validity: **VALID_SOURCE_BRANCH**. There is NO active-model board and NO provider dropdown for empty catalog. Source does not show a loading spinner on this page unless catalog update is running; Go distinct loading/disconnected messaging is additional error/readiness handling. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Search models…; No models reported count and empty card; No models reported hint; Update model list; Favorites.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models empty catalog. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V125. Overrides: `{"surface": "Settings > Models", "catalog": []}`.
- Planned pair: `../probes/visual/V125/rust.png` ↔ `../probes/visual/V125/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V126 · Models query filtered
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_rows:L1433](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1433)
- State validity: **VALID_SOURCE_BRANCH**. Source-render branch exists. Concrete fixture specified; actual paired capture still unverified. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Typed query; 1 of 5 count; Active model board unchanged; First heading with one Beta card.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models query filtered. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V126. Overrides: `{"surface": "Settings > Models", "models_query": "  BeTa  "}`.
- Planned pair: `../probes/visual/V126/rust.png` ↔ `../probes/visual/V126/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V127 · Models favorites only
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_rows:L1433](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1433)
- State validity: **VALID_SOURCE_BRANCH**. Source-render branch exists. Concrete fixture specified; actual paired capture still unverified. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Selected Favorites toggle; 2 of 5 count and two favorites; First/Beta and Second/Second Same cards only; Active board.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models favorites only. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V127. Overrides: `{"surface": "Settings > Models", "favorites": [["first", "beta"], ["second", "same"]], "favorites_only": true}`.
- Planned pair: `../probes/visual/V127/rust.png` ↔ `../probes/visual/V127/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V128 · Models provider filtered
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_rows:L1433](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1433)
- State validity: **VALID_SOURCE_BRANCH**. Source-render branch exists. Concrete fixture specified; actual paired capture still unverified. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Second provider dropdown label; 2 of 5 count; Second heading and two cards; Active board.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models provider filtered. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V128. Overrides: `{"surface": "Settings > Models", "provider_filter": "second"}`.
- Planned pair: `../probes/visual/V128/rust.png` ↔ `../probes/visual/V128/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V129 · Models combined no matches
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_rows:L1433](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1433)
- State validity: **VALID_SOURCE_BRANCH**. Source sections starts with active_model_row, so sections.is_empty is unreachable: do not invent a Rust no-matches card. Go intentionally adds a clearer no-matches notice, which is an explicit visual delta. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Query and active filter controls; 0 of 5 count; Active model board only below toolbar.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models combined no matches. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V129. Overrides: `{"surface": "Settings > Models", "models_query": "zzz-no-match", "favorites": [["first", "beta"]], "favorites_only": true, "provider_filter": "second"}`.
- Planned pair: `../probes/visual/V129/rust.png` ↔ `../probes/visual/V129/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V130 · Models updating
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_toolbar:L1082](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1082)
- State validity: **VALID_SOURCE_BRANCH**. Duplicate requests are ignored in handler, not a distinct disabled-button render branch; original pointer/handler remain. Do not label it Updating or claim one screenshot verifies duplicate gating. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Unchanged Update model list label; Accent refresh SVG rotating at a700ms period, captured at frozen phase; Update button opacity 0.6; Unchanged catalog.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models updating. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V130. Overrides: `{"surface": "Settings > Models", "refreshing": true, "animation_phase": "accent refresh SVG rotation fixed at 0 degrees; reduce-motion freezes at this phase"}`.
- Planned pair: `../probes/visual/V130/rust.png` ↔ `../probes/visual/V130/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
- Note: Independent source challenge corrected first-pass expectations; no image comparison has run.

## V131 · Models update error
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_toolbar:L1082](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1082)
- State validity: **VALID_SOURCE_BRANCH**. Source error is a toast, not an inline catalog error card. Go also retains error detail inline; document that as a robustness delta, not source pixels. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Could not update the model list: fixture catalog update failure error toast; Update model list returned to normal opacity; Last catalog still visible.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models update error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V131. Overrides: `{"surface": "Settings > Models", "catalog_update_result": {"exit_code": 1, "stderr": "fixture catalog update failure"}, "toast_elapsed_ms": 0}`.
- Planned pair: `../probes/visual/V131/rust.png` ↔ `../probes/visual/V131/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V132 · Models card hover
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_card:L1561](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1561)
- State validity: **VALID_SOURCE_BRANCH**. Remove invented input/output/cost metadata. These are not rendered by model_card. Context-window presence is the only card capability metadata. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Alpha name; alpha ID; Favorite star; 128k context badge; Set active footer revealed; Current Second Same retains Active badge.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models card hover. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V132. Overrides: `{"surface": "Settings > Models", "pointer": "center of inactive First/Alpha card"}`.
- Planned pair: `../probes/visual/V132/rust.png` ↔ `../probes/visual/V132/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
- Note: Independent source challenge corrected first-pass expectations; no image comparison has run.

## V133 · Models provider filter dropdown
- Features: P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::model_provider_select:L1282](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1282)
- State validity: **VALID_SOURCE_BRANCH**. Use enough providers to reach the scroll limit. All providers sentinel is first but may be outside the centered visible viewport; do not claim it is simultaneously visible with every provider. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Search filter with focus; Flat provider-name options with chosen Provider 05 check; Keyboard wash on selected row; List viewport capped at 220px.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Models provider filter dropdown. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V133. Overrides: `{"surface": "Settings > Models", "catalog": "Ten one-model providers provider-00 through provider-09, name Model 00 through Model 09, id model-00 through model-09, in numeric order", "provider_filter": "provider-05", "open_control": "All providers"}`.
- Planned pair: `../probes/visual/V133/rust.png` ↔ `../probes/visual/V133/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V134 · Settings active model dropdown
- Features: P04, P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::active_model_select:L1317](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1317)
- State validity: **VALID_SOURCE_BRANCH**. No provider-group headers in settings selects. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Search filter; Five flat Provider · Name choices in catalog order; Check on Second · Second Same; Selected and keyboard wash.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Settings active model dropdown. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V134. Overrides: `{"surface": "Settings > Models", "open_control": "Active model"}`.
- Planned pair: `../probes/visual/V134/rust.png` ↔ `../probes/visual/V134/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V135 · Settings default model dropdown
- Features: P04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- State validity: **VALID_SOURCE_BRANCH**. Remove invented provider-grouped options. Current selection means saved default, not live current model. List is flat. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: pi default sentinel; Five flat Provider · Name options in catalog order; Check on First · Beta.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Settings default model dropdown. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V135. Overrides: `{"surface": "Settings > Agent", "default_model": ["first", "beta"], "open_control": "Default model"}`.
- Planned pair: `../probes/visual/V135/rust.png` ↔ `../probes/visual/V135/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
- Note: Independent source challenge corrected first-pass expectations; no image comparison has run.

## V136 · Settings default thinking disabled
- Features: P04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **VALID_SOURCE_BRANCH**. No popup, caret, or fabricated thinking ladder. A stale unmatched default model follows the same render branch. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Default model chip pi default; Default thinking noninteractive pi default chip.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Settings default thinking disabled. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V136. Overrides: `{"surface": "Settings > Agent", "default_model": null}`.
- Planned pair: `../probes/visual/V136/rust.png` ↔ `../probes/visual/V136/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V137 · Settings default thinking dropdown
- Features: P04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- State validity: **VALID_SOURCE_BRANCH**. Use Beta explicit null map to suppress off/minimal/medium; max is present because explicitly mapped. Do not substitute the live session ladder. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: pi default sentinel; Low, High, Max rows from Beta capability map; Check on High.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Settings default thinking dropdown. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V137. Overrides: `{"surface": "Settings > Agent", "default_model": ["first", "beta"], "default_thinking": "high", "open_control": "Default thinking"}`.
- Planned pair: `../probes/visual/V137/rust.png` ↔ `../probes/visual/V137/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V138 · Settings title model dropdown
- Features: P04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- State validity: **VALID_SOURCE_BRANCH**. Replace inaccurate Session default with exact Active session model; it uses live session model, not session birth default. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Active session model sentinel; Five flat Provider · Name options; Check on First · Alpha.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Settings title model dropdown. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V138. Overrides: `{"surface": "Settings > Agent", "title_model": ["first", "alpha"], "open_control": "Title model"}`.
- Planned pair: `../probes/visual/V138/rust.png` ↔ `../probes/visual/V138/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
- Note: Independent source challenge corrected first-pass expectations; no image comparison has run.

## V139 · Settings model dropdown filtered
- Features: P04, P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- State validity: **VALID_SOURCE_BRANCH**. Search retains original catalog indices; choose a saved default outside the filter to distinguish keyboard highlight from selected check. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Typed First query; Three First provider options only; Keyboard-highlighted Alpha row; No sentinel because it does not match.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Settings model dropdown filtered. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V139. Overrides: `{"surface": "Settings > Agent", "open_control": "Default model", "dropdown_query": "First", "keys": ["ArrowDown"]}`.
- Planned pair: `../probes/visual/V139/rust.png` ↔ `../probes/visual/V139/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V140 · Settings model dropdown no matches
- Features: P04, P08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- State validity: **VALID_SOURCE_BRANCH**. The explicit No matches text is rendered; an empty blank popup is not source parity. Enter sends nothing. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Typed unmatched query; One muted No matches entry.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Settings model dropdown no matches. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V140. Overrides: `{"surface": "Settings > Models", "open_control": "Active model", "dropdown_query": "zzz-no-match"}`.
- Planned pair: `../probes/visual/V140/rust.png` ↔ `../probes/visual/V140/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V141 · Composer model all groups
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Three models make all three sections fit simultaneously inside six-row-equivalent viewport. Larger catalogs require separate scrolled captures to claim all groups visible. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Recents heading and Beta; Favorites heading and Alpha; Second provider heading and current Second Same check; Search, scope row and footer count3.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer model all groups. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V141. Overrides: `{"surface": "Composer", "catalog_subset": [0, 1, 2], "recents": [["first", "beta"]], "favorites": [["first", "alpha"]], "picker": "Model", "scope": "All"}`.
- Planned pair: `../probes/visual/V141/rust.png` ↔ `../probes/visual/V141/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V142 · Composer model recents
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. No Recents header in dedicated Recents scope. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Selected Recents scope; Flat Beta, Second Same, Alpha rows in MRU order; Current check; 3 recent footer.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer model recents. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V142. Overrides: `{"surface": "Composer", "recents": [["first", "beta"], ["second", "same"], ["first", "alpha"]], "picker": "Model", "scope": "Recents"}`.
- Planned pair: `../probes/visual/V142/rust.png` ↔ `../probes/visual/V142/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V143 · Composer model favorites
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Order is provider-grouped catalog traversal, not order in favorites.json. No Favorites group header in this dedicated scope. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Selected Favorites scope; Flat Alpha, Beta, Second Same rows; Accent stars; Current check; 3 favorites footer.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer model favorites. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V143. Overrides: `{"surface": "Composer", "favorites": [["second", "same"], ["first", "beta"], ["first", "alpha"]], "picker": "Model", "scope": "Favorites"}`.
- Planned pair: `../probes/visual/V143/rust.png` ↔ `../probes/visual/V143/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V144 · Composer model provider
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Provider-scope recents differ from MRU ordering; no redundant provider group heading. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Selected First scope; Recents heading with Alpha then Beta in catalog order; First Same below recent rows without provider header; 3 models footer.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer model provider. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V144. Overrides: `{"surface": "Composer", "recents": [["first", "beta"], ["first", "alpha"]], "picker": "Model", "scope": "provider:first"}`.
- Planned pair: `../probes/visual/V144/rust.png` ↔ `../probes/visual/V144/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V145 · Composer model filtered
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Use both equal-ID providers, not a single matching row, to visibly prove provider identity and independent keyboard highlight. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Typed same query; Second Same with actual current check; First Same row; After ArrowDown keyboard highlight on First Same, distinct from current check.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer model filtered. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V145. Overrides: `{"surface": "Composer", "picker": "Model", "picker_query": "same"}`.
- Planned pair: `../probes/visual/V145/rust.png` ↔ `../probes/visual/V145/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V146 · Composer model no matches
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Explicit two-line no-match entry is rendered, not blank empty results. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: No matching models title; Try a provider, model or id hint; Zero model count footer; Search/scope controls remain.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer model no matches. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V146. Overrides: `{"surface": "Composer", "picker": "Model", "picker_query": "zzz-no-match"}`.
- Planned pair: `../probes/visual/V146/rust.png` ↔ `../probes/visual/V146/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V147 · Composer model favorite current chip
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Use one short provider so full label fits. For larger catalogs this is a horizontal-scroll row; clipping at right does not justify replacing label with an icon-only action. Separate scroll regression proves reachability. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: All scope; Second scope; Full Favorite current star/text chip; Current model row.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer model favorite current chip. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V147. Overrides: `{"surface": "Composer", "catalog_subset": [1], "favorites": [], "recents": [], "picker": "Model", "scope_scroll_x": 0}`.
- Planned pair: `../probes/visual/V147/rust.png` ↔ `../probes/visual/V147/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V148 · Composer model row hover
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Hover must follow guard expiration and real pointer movement; stationary pointer after scroll is not a highlight event. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Unfavorited Alpha star revealed; Second Same favorite star remains accent; Second Same current check remains.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer model row hover. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V148. Overrides: `{"surface": "Composer", "catalog_subset": [0, 1], "favorites": [["second", "same"]], "picker": "Model", "pointer": "First/Alpha row after 400ms hover guard"}`.
- Planned pair: `../probes/visual/V148/rust.png` ↔ `../probes/visual/V148/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
