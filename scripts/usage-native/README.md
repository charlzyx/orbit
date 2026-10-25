# Native Usage paired-capture input

These are recorded, synthetic pi JSONL files, not display-only aggregates. No credentials, network or model inference is involved. Both collectors read the same four files from a direct workspace group under `$HOME/.pi/agent/sessions`. The Rust collector uses HOME rather than PI_CODING_AGENT_DIR, so set both consistently as the main capture harness already does.

## Install before launching either application

Keep `scripts/capture-native-reference.sh` viewport, theme, font, scale, runtime, off-line and graphics settings unchanged. Immediately after it creates the fresh isolated HOME and writes ui.json/theme, and before the binary is launched, run:

    python3 "$repo_root/scripts/usage-native/materialize.py" "$HOME" --manifest "$output/usage-fixture.json"
    mkdir -p /tmp/orbit-visual-fixture/workspace/{alpha,beta}

The materializer refuses to replace any existing target. Use the harness's new isolated home, never a real home. The committed `testdata/usage-native/home` contains the same deterministic bytes for collector regression tests; do not add these fixture workspaces to the app's project list. This preserves a matched empty project sidebar. The two apps use the same absolute fake cwd strings.

Data oracle: 33 requests, 33 prompts, 4 sessions, 2 providers, 4 provider/model pairs (two duplicate model IDs at different providers), 55,360 tokens, $0.912 actual total, 4 tool runs with 1 tool error, 1 failed request, 1 aborted request and 1 additional unbilled provider error. All 33 requests have measurable gaps and their models have positive price evidence. The custom range persisted in usage.json is October 1–14, 2026 inclusive; this is independent of the run's current date. Mode starts Simple, Tokens, Average, Models, Sessions, tokens-desc and 25 session rows. First chronological prompt supplies each session title.

## Actual UI capture sequence

After the main harness sizes the actual client to 1470×922, use its native window ID and display. Do not use CPU fixture captures here.

1. `bash scripts/usage-native/capture-state.sh simple "$window" "$output"` sends Ctrl+U, the pinned global ToggleUsage command (`platform.rs:762`, `commands.rs:894`, Go commands.go). Inspect usage-simple-dark.png: selected Simple, custom Oct 1–14 chip, populated Summary, Usage over time and real data. Confirm scan is complete, not skeleton. If still loading, wait and capture the same window again without another Ctrl+U toggle.
2. From that actual renderer's inspected image, find the Details segment in the filter bar. `bash scripts/usage-native/capture-state.sh details "$window" "$output" X Y` clicks that verified point. Inspect selected Details and the Breakdown/Records cards. Do not assume Go and Rust coordinates match before measuring them. A wrong click is a failed state setup, never a valid pair.
3. From the inspected Details image, locate All models. `bash scripts/usage-native/capture-state.sh models "$window" "$output" X Y` opens the real model menu. Inspect Search, All, Select all and the four provider-scoped model options. Do not select an option; the intended state is the unfiltered menu-open view. Escape closes it after capture.

Each script invocation records its pointer setup and checks actual image size, but explicitly does NOT mark the state visually verified. Record the inspected visible labels, fixture manifest SHA-256, binary/source hashes, exact UI actions and selected mode in the paired capture manifest. Regenerate both sides from the same committed fixture bytes. No resize, blur, recolor or crop may conceal layout differences.

## Clock and other open gates

Process clock is NOT frozen. The scan-relative header age is volatile; collect both within the same stated interval and report any age difference rather than masking it. The activity heatmap ends at the scan's calendar day, so same fixture data alone does not make captures from different UTC dates comparable. The fixture manifest lists these exceptions. Per-renderer resolved fonts, hover/focus phase, native execution and actual matched pixel comparison still need their own evidence. This script/fixture preparation is not an executed native Usage pair.
