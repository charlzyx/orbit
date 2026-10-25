# Visual inventory: transcript (V044–V168)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V044 · Transcript streaming answer follow tail
- Features: C07, C08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render:L8669](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L8669)
- Expected: Transcript streaming answer follow tail;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript streaming answer follow tail. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V044/rust.png` ↔ `../probes/visual/V044/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V045 · Transcript user scrolled up jump to bottom
- Features: C07, C08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render:L8669](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L8669)
- Expected: Transcript user scrolled up jump to bottom;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript user scrolled up jump to bottom. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V045/rust.png` ↔ `../probes/visual/V045/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V046 · Transcript prompt navigation rail active prompt
- Features: C07, C08. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render:L8669](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L8669)
- Expected: Transcript prompt navigation rail active prompt;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript prompt navigation rail active prompt. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V046/rust.png` ↔ `../probes/visual/V046/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V047 · Transcript markdown headings links lists tasklists quote alert
- Features: C09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render:L8669](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L8669)
- Expected: Transcript markdown headings links lists tasklists quote alert;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript markdown headings links lists tasklists quote alert. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V047/rust.png` ↔ `../probes/visual/V047/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Typed transcript owner supplies exact source path/symbol and tool/detail variants.

## V048 · Transcript table long cell horizontal scroll
- Features: C09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render:L8669](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L8669)
- Expected: Transcript table long cell horizontal scroll;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript table long cell horizontal scroll. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V048/rust.png` ↔ `../probes/visual/V048/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Typed transcript owner supplies exact source path/symbol and tool/detail variants.

## V049 · Transcript find multiple results current highlight
- Features: C13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Transcript find multiple results current highlight;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response; Active band and other matching-row wash use different alpha.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript find multiple results current highlight. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V049/rust.png` ↔ `../probes/visual/V049/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V050 · Transcript find zero results
- Features: C13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Transcript find zero results;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript find zero results. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V050/rust.png` ↔ `../probes/visual/V050/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V051 · Transcript retry countdown
- Features: C14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::run_status_strip:L3882](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3882)
- Expected: Transcript retry countdown;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript retry countdown. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`
- Planned pair: `../probes/visual/V051/rust.png` ↔ `../probes/visual/V051/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V052 · Transcript compacting
- Features: C14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::run_status_strip:L3882](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3882)
- Expected: Transcript compacting;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript compacting. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`
- Planned pair: `../probes/visual/V052/rust.png` ↔ `../probes/visual/V052/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V153 · Transcript user text and images
- Features: C05, C07, C09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: User prose;  Multiple image attachment thumbnails;  User timestamp.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript user text and images. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V153/rust.png` ↔ `../probes/visual/V153/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V154 · Transcript assistant markdown image loaded
- Features: C09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Assistant Markdown image loaded;  Alt text and surrounding prose.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript assistant markdown image loaded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V154/rust.png` ↔ `../probes/visual/V154/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V155 · Transcript markdown image fallback
- Features: C09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Unavailable image fallback;  Alt text;  No fabricated bitmap.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript markdown image fallback. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V155/rust.png` ↔ `../probes/visual/V155/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V156 · Transcript working before prose
- Features: C11, C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_working_indicator:L4820](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L4820)
- Expected: Working label;  Fixed elapsed time;  Wave dots at frozen phase.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript working before prose. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V156/rust.png` ↔ `../probes/visual/V156/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V157 · Transcript working activity verb
- Features: C11, C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_working_indicator:L4820](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L4820)
- Expected: Activity-specific working verb;  Real active tool context.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript working activity verb. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V157/rust.png` ↔ `../probes/visual/V157/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V158 · Transcript turn work folded
- Features: C11, C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Settled work summary collapsed;  Answer retained;  Later work ordered after answer.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript turn work folded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V158/rust.png` ↔ `../probes/visual/V158/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V159 · Transcript turn work expanded
- Features: C11, C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Pre-answer work expanded;  Post-answer work and later prose ordered.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript turn work expanded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V159/rust.png` ↔ `../probes/visual/V159/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V160 · Transcript thinking streaming
- Features: C11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_thinking_body:L2626](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L2626)
- Expected: Live thinking body;  Streaming affordance.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript thinking streaming. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V160/rust.png` ↔ `../probes/visual/V160/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V161 · Transcript thinking settled open
- Features: C11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_thinking_body:L2626](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L2626)
- Expected: Settled expanded thinking;  Collapse control.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript thinking settled open. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V161/rust.png` ↔ `../probes/visual/V161/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V162 · Transcript thinking folded
- Features: C11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Collapsed thinking header;  No body rows.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript thinking folded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V162/rust.png` ↔ `../probes/visual/V162/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V163 · Transcript thinking inner scroll detached
- Features: C08, C11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_thinking_body:L2626](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L2626)
- Expected: Long thinking body;  Inner scrollbar;  Detached position.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript thinking inner scroll detached. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V163/rust.png` ↔ `../probes/visual/V163/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V164 · Transcript provider error
- Features: C14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_assistant_error:L2240](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L2240)
- Expected: Provider failure card;  Exact bounded failure text.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript provider error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`
- Planned pair: `../probes/visual/V164/rust.png` ↔ `../probes/visual/V164/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V165 · Transcript stopped partial answer
- Features: C14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_assistant_error:L2240](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L2240)
- Expected: Partial answer retained;  Stopped label.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript stopped partial answer. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`
- Planned pair: `../probes/visual/V165/rust.png` ↔ `../probes/visual/V165/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V166 · Transcript tool shell
- Features: C11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_tool_detail:L3551](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L3551)
- Expected: Shell command;  Expanded stdout;  Path/command styling.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript tool shell. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V166/rust.png` ↔ `../probes/visual/V166/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V167 · Transcript tool read view
- Features: C11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_tool_detail:L3551](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L3551)
- Expected: Read/view path;  Bounded content preview.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript tool read view. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V167/rust.png` ↔ `../probes/visual/V167/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V168 · Transcript tool write
- Features: C11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_tool_detail:L3551](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L3551)
- Expected: Written file path;  Addition diff;  Line gutter.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript tool write. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V168/rust.png` ↔ `../probes/visual/V168/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
