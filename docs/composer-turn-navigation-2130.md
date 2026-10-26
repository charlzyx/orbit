# Focused composer turn-navigation correction

Base: 0bf7cb693a5f120d9f424c6793de399af45fae75. Isolated source-only candidate; no native retry authorized or performed.

## Source contract
Pinned Rust299a32b composer_send.rs429–447 explicitly verifies Composer context resolves Ctrl/Cmd-Up/Down to PrevTurn/NextTurn. main.rs370–406 separately binds plain arrows and home/end, plus modified left/right editing. app/session.rs2772–2792 does not condition turn navigation on empty draft or selection; transcript.rs2278–2315 defines preceding user turn from live edge. No source-based transcript navigation behavior is inferred for secondary-home/end; this patch leaves these untouched. Plain arrow autocomplete precedence remains.

Go's TextArea owns modified arrows (MyGo input.go703–722, editor_navigation.go26–50), so a global Shortcut cannot recover the consumed key. Existing navigation test directly called shortcutCommand and missed this event route. The patch only routes exact Cmd-Up/Down through the existing command inside chat composer's OnTextKey, after existing attachment and autocomplete handling. No toolkit/global ordering/harness changes.

Active IME is explicitly left to the existing editor path using the existing compatibility Composing query. Rust's app handler itself has no IME condition; this conservative preservation of Go's active-IME behavior is not a claim that all platform IME dispatch matches Rust. Existing editor arrow handling may end preedit; the negative test checks no transcript interception, then starts a fresh preedit to independently verify Enter does not submit. No default IME policy changes are introduced.

## Observed failure and validation
Authorized one-shot bfc1579/run40234988331: Rust CHAT01/02 raw screenshots show tail then PROMPT ELEVEN/RESPONSE NUMBER11. Go CHAT01 tail loaded, CHAT02 before and final settle images byte-identical, still tail. Evidence /workspace/shared/orbit-chat-third-2122/verification.md. Keep that capture FAIL even after this source correction.

Mounted RED (red.log): full App1470x922, native click Message,12 turns following tail, real Tester.Key(Cmd,Up) remained row21 instead of preceding user row20. GREEN after13-line OnTextKey addition reached row20. No direct shortcut invocation in new tests.

Mounted matrix: empty draft; multiline nonempty draft with caret/selection preserved via subsequent replacement and Undo; Up/Down middle anchors; first/last boundaries; global shortcut outside composer; ordinary arrows and home/end, modified home/end left in editor; autocomplete ordinary Down consumes popup while modified Up navigates; active-preedit negative control plus existing submit IME coverage.

The final user turn is close enough to document end for ScrollTo(Start) to clamp; tests assert exact middle-row targets and separately compare true bottom bounds rather than requiring an impossible top-aligned final row. This is test geometry, not relaxed native expected text. No new native or pixel pass is claimed.

Final scoped verification: new matrix plus existing IME race count2 PASS23.036s (`final-race.log`); broader Navigation/NativeComposerCompletion set race count2 PASS35.208s (`race.log`); UI package vet PASS (`vet.log`); git diff --check PASS. No complete final-tree UI suite or native recapture claimed.
