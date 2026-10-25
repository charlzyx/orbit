# Native screenshot review: f61e00f

Pinned Rust299a32b vs Go f61e00f; actual hosted Linux windows,1470×922 client pixels,100% scale,96dpi, English, Orbit Azure dark. No image resizing, cropping, alignment or masks were used.

**15 observed page states,30 original PNGs,15 raw diagnostic comparisons,0 exact pixel matches,0 fully normalized pairs,0 visual passes.** All30 original images were inspected individually. This is not completion of490 candidate cases. Four empty pages map to their empty-state IDs, not their populated baseline cases.

Actual glyph-face resolution, caret/clock/animation phase remain unverified. Privacy scratch contents, runtime PIDs/uptime and MCP timestamps differ between jobs; their raw metrics are diagnostic only. Below-fold controls and alternate states remain unexecuted.

[Hosted run](https://github.com/charlzyx/orbit/actions/runs/40170731635)

| Candidate | Screenshot | Changed pixels | Raw percentage | Manual finding |
|---|---|---:|---:|---|
| V001 | new-task-dark.png | 64096 | 4.7291% | Empty new task, connected real offline Pi. Sidebar/header alignment improved; composer is 3px shorter and hero/text positions differ; caret phase differs. |
| V078 | settings-about-dark.png | 354960 | 26.1897% | FAIL: source icon, star banner and version/component row structure absent. Go/framework/development identity should remain truthful, not falsely claim GPUI. |
| V075 | settings-agent-dark.png | 671508 | 49.5454% | FAIL: source Behavior/Titles/Defaults ordering and segmented controls are absent. Unknown retry state is a separately documented deliberate difference, not a pass. |
| V071 | settings-appearance-dark.png | 717972 | 52.9736% | FAIL: generic separate theme rows replace source segmented appearance/active theme; missing source preview tiles, source language placement, labels and ordering. |
| V070 | settings-general-dark.png | 28686 | 2.1165% | Expected General groups, paths and switches are visible. Remaining baseline/row geometry and glyph differences; development version differs. |
| V235 | settings-mcp-dark.png | 57671 | 4.2551% | Empty MCP, NOT populated V234. Source icon and unboxed centered state replaced by bordered card; header checked-time/path formatting differs and clock is volatile. |
| V125 | settings-models-dark.png | 234810 | 17.3248% | Empty catalog, NOT populated V124. FAIL: source centered icon/large empty panel replaced by compact left-aligned text; toolbar geometry differs. |
| V225 | settings-plugins-dark.png | 246543 | 18.1905% | Empty plugins, NOT populated V222. Source icon tile, larger filled empty card, input icon and toolbar geometry differ. |
| V073 | settings-privacy-dark.png | 678459 | 50.0582% | FAIL: retention labels/control layout and scratch-file inventory differ. Actual temporary build trees differ between jobs, so dynamic inventory is not a normalized fixture. |
| V077 | settings-providers-dark.png | 738620 | 54.4970% | FAIL: source responsive provider card grid/icons/auth actions is replaced by a vertical Configure list. Runtime catalog parsing also differs; no credentials used. |
| V253 | settings-report-bug-dark.png | 659900 | 48.6889% | FAIL: form width/grouping, segmented Issue type, Generate draft placement and footer differ; Go adds privacy/diagnostic controls. No issue submission or external communication executed. |
| V074 | settings-runtime-dark.png | 635548 | 46.8921% | FAIL: source RPC-patch switch absent from this Go surface; status/transport labels, extra dependency rows and stderr card differ. PID/uptime are volatile. |
| V072 | settings-shortcuts-dark.png | 713123 | 52.6158% | FAIL: Go lists palette-only commands and expanded 1–9 chords where source uses shorter reference rows; wrong row heights/dividers and chip case. |
| V216 | settings-skills-dark.png | 17960 | 1.3251% | Empty skills, NOT populated V213. Source no-match search icon missing, search row/empty-state offsets differ. Master/detail split is present. |
| V076 | settings-worktrees-dark.png | 379681 | 28.0137% | FAIL: missing Manage worktrees navigation row, inline editable location fields, source group structure and labels. |
