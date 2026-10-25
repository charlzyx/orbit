# C08 conversation navigation: scoped evidence (2026-10-25)

## Original source and concrete corrections
Baseline Rust 299a32bc87fe3d4851bdb902d9c03dcf9bf99cac at `/workspace/shared/orbit-rust-reference-1628`.
- `transcript_view.rs:80–95,1109–1156,1396–1734`: actual width threshold **1040**, not stale 872 comment; ≥2 user turns and scrollable transcript. Left16, width44, 12px tick pitch, 2px bar height, base width32 with hover scales1/.68/.44/.25; at-rest width8; viewport cap80%. Independent wheel scroll, active tick autoscroll only when active changes, edge fades20px.
- `transcript_view.rs:7437–7471`: active user turn follows viewport once reader leaves tail, including during streaming. Go keeps existing virtualized List, stable row keys, FollowEnd and Jump to latest; new rail never changes list layout width.
- `transcript.rs:2278–2317` and `app/session.rs:2772–2790`: keyboard previous/next moves relative to current USER TURN, not nearest user row before firstVisible. At tail anchor newest user turn. Original Go incorrectly landed back in current user turn from assistant rows (red expected4/got6). Keyboard use dismisses hint even at edge.
- `transcript_view.rs:1140–1156,1635–1750`: materialize only hovered prompt/first nonempty response before next user; normalized Unicode grapheme limits100/240. Preview width320/max126, gap12, clamp within rail; hint216px; never show hint and preview together.
- `transcript.rs:845,857,897–906,2196–2210`:10sec hint, per-install rail_hint_seen, retain sibling hints.json keys; click/tick/keyboard/TTL dismiss. Hover temporarily replaces hint but does not persist dismissal.
- `transcript_view.rs:827–904,1340–1383`: canvas hover governs selection start, outside panel bounds governs clear. Rail should preserve selected text. Original canonical MyGo selection overlay cleared any other control/outside press, exposing an actual source mismatch when rail was introduced.

## Frozen source chain
1. `/workspace/shared/orbit-chat-navigation-1728`, based on frozen feedback-v2. `orbit-chat-navigation-v1.patch`, SHA256 `2ff8fe60a6462cb6033e9f882b2f7f45ffb0f6bf33167dddc749f5d3b642fd15`. This intentionally retains failing rail-selection regression; **not independently acceptable**.
2. `/workspace/shared/orbit-chat-navigation-selection-1738`: independent reviewed candidate for the narrow selection interaction. Only opted-in rail shell can preserve a selection from its exact live scope and build; no general outside/control relaxation. Canonical overlay remains pinned/hash verified; source module cache unchanged. Final candidate patch/hash to be appended after review/race.

## Evidence obtained
- `orbit-chat-navigation-red.log`: original implementation missing rail; keyboard from assistant row expected4/got6.
- `orbit-chat-navigation-first.log`: first actual mounted tick/hover and keyboard regressions pass.
- `orbit-chat-navigation-count3.log`: rail geometry/visibility/threshold resize, hint TTL+persistence+sibling preservation, hover fan, independent wheel, tail/follow, stream history-reading, session reset/same-length replacement, grapheme/snippet boundaries, search/focus/Chinese preview, existing find/follow regressions pass (UI3.085sec).
- `orbit-chat-navigation-freeze-race.log`: V1 UI fails exactly `TestNavigationRailKeepsRetainedTextSelection` (original selected prompt becomes clipboard sentinel); existing retained toolkit fixture passes. This failure was not skipped or weakened.
- `orbit-chat-navigation-selection-optin-first.log`: same unmodified actual mounted app selection regression now passes; toolkit positive/narrow-negatives first pass (UI1.696/toolkit0.104sec).
- `orbit-chat-navigation-selection-optin-count3.log`: all navigation tests including original red selection, existing retained copy and search/follow plus toolkit retained regressions PASS (UI5.082/transcript0.065/toolkit0.577sec).
- `orbit-chat-navigation-overlay-verifier.log`:21 canonical wrapper verification tests PASS. No hash/pin guard bypass.
- `orbit-chat-navigation-headless.png`: actual MyGo full App raster inspected; rail fan/active marker and one hovered card visible. This is headless CPU evidence only, not a paired hosted macOS/Rust capture.

## Remaining acceptance boundaries
No provider credentials or real-user session files used. No publication/main merge/release. No hosted macOS input, native occlusion/drag/menu/window tests, identical frozen Rust/Go visual pair or full long-session performance acceptance established here. Existing unrelated generic error translations, remote markdown image loading and larger migration gaps are outside this batch. Do not relabel C08/full migration completed from focused passes.

## Final candidate freeze

- Increment over navigation-v1: `/workspace/shared/orbit-chat-navigation-selection-v2.patch`; SHA256 `32fa47935baf240989392a54672d227a9a8f126d37adf9240ec3df5038b766af`. Same-prefix manifest records each file. Source copy remains isolated.
- `orbit-chat-navigation-selection-race.log`: final code race PASS UI12.925s/transcript1.149s/toolkit1.780s, including navigation, original follow/find, retained selection and composer compatibility tests.
- `orbit-chat-navigation-vet.log`: scoped UI/transcript vet PASS.
- Earlier v1 red remains archived unchanged; final increment fixes the existing failing test rather than deleting/skipping/changing its requirement. Narrow API remains subject to parent review before integration.

## Review corrections and final stacking order

- Supersede selection-v2 with **selection-v3**, still based directly on navigation-v1. `orbit-chat-navigation-selection-v3.patch`, SHA256 `f1e6d9f2254c67f7bcebca79c817444b6067f5d761297ff8b6a986b96cfcbcde`. `orbit-chat-navigation-crosswindow-red.log` proves old cross-window expired Element lookup panicked; nodeFor now rejects foreign owner before lookup. Double-Tester regression and all composer/selection/navigation race tests PASS: UI13.508/transcript1.170/toolkit2.119s in `orbit-chat-navigation-selection-v3-race.log`.
- Parent review found synchronous hint preference Update on native frame/click/TTL path. `orbit-chat-navigation-hint-lock-red.log` proves a deliberately held Store mutex blocked the actual mounted hint click (bounded guard released it after1sec; no hang inference).
- Independent final source `/workspace/shared/orbit-chat-navigation-hint-1748` adds captured-store App.Async write; UI optimistically dismisses globally and sets saving; worker callback only clears saving and reports a global error toast. No session/model mutation, no UI-thread file access from dismiss. Repeated dismiss is idempotent while pending/after failure. Store.Update still reloads and preserves concurrent sibling keys; malformed JSON is not overwritten.
- Final additive **hint-async** patch after selection-v3: `orbit-chat-navigation-hint-async.patch`, SHA256 `1fbba7c70a3e0bfade7a6fdafdb1fe7b4c81516d56033a79b0b885ce3d497369`. Only3files (App field, worker change, regressions).
- `orbit-chat-navigation-hint-count5.log`: all navigation + original follow/find tests PASS UI9.464s. Existing TTL/persistence assertions await the actual hint saving completion instead of assuming synchronous I/O. New tests hold store lock, click hint, repeat dismissal10times, switch session before callback, prove sibling-key preservation; malformed JSON failure appears via UI completion without changing owner/presentation.
- Final exact-code `orbit-chat-navigation-hint-final-race.log`: PASS UI11.370/transcript1.253/toolkit1.912s including original native-toolkit composer and retained selection regressions. No hosted OS acceptance claim.

Apply order: frozen feedback-v2 → navigation-v1 → navigation-selection-v3 (NOT v2 then v3) → navigation-hint-async. Each remains a separate reviewable immutable artifact; do not apply intermediate known-red v1 alone.
