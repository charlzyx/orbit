# Thought-only residual wheel routing — 2026-10-26 11:54 UTC

Base a1fb7be9fe9b86bac7ae7c1409f797842780fa6a (tiny Jump clip). This isolated candidate does not alter the current integration gate.

## Source contract and reachability
Pinned Rust299a32bc transcript_view.rs2778–2809 renders260px Thought overflow scroll with occlusion and chain callback. 2858–2891 splits the event into inner consumption and residual;2900–2904 clamp equation and7521–7542 source tests specify5px inner+5px outer from a10px overshoot. MessageScroller.scroll_by253–271 uses scroll_to_end on tail, clearing unread239; ordinary List wheel48–51 deliberately does not clear sticky unread.
Go actual chat logicalMessage→transcriptui.Turn→turn.go Thought Scroll is reachable for projected assistant Thinking. Pinned MyGo input.go503–520 returns after first scrollBy success;559–578 clamps with a bool result, so same-event residual is lost. Existing standalone Thought follow test lacked the parent List.

## Minimal opt-in
Options.ThoughtScroll is optional; nil preserves old MyGo default. Only Thought body installs HandleInput. InputScroll uses delivered local DIP DY, or DX if DY0; finite checks precede mutation. It computes clamped inner target and residual, invokes captured host callback even for0 to validate owner, then commits inner offset only after acceptance. All other event kinds returnfalse. No toolkit/global scroll changes.
Host captures the existing inline owner (runtime/session/workspace/opaque task generation/window lifetime); stale session/runtime/ABA/closed callbacks reject before changing either tracked scroll or unread. Accepted nonzero residual updates existing transcriptScroll. At saturated tail it uses List.ScrollToEnd rather than persisting an estimated MaxY; only this Thought-origin path clears unread. Ordinary outer wheel unchanged.

Both precise and notch input are immediate inside Thought, matching the source chain; other MyGo notch glides remain unchanged. Public InputEvent is alreadyDIP. Darwin surface.go882–892 normalizes signs and converts nonprecise lines to40DIP before delivery; no second rescaling here. Shift is validated at callback level plus native source, not physical input: public Tester.Scroll/Wheel omit Mods even after HoldModifiers. This remains an explicit native verification limit.

## Evidence and corrections
Real Turn/List baseline RED: inner650→655(max655) on10px, outer0→0 instead of5. Inner-only and exact-boundary no-outer controls, already-saturated next-event bubbling pass. Corrected actual App baseline RED uses visible full Thinking text bounds X285/Y414.35/W878/H255 and point335/541.85, not guessed header offset. Real wheel away + snapshot change establishes unread; nested tail-return then ordinary wheel away incorrectly retains it. Ordinary outer-wheel control correctly retains sticky.
Earlier1141 App header+30 test hit ordinary area, so its nested-path claim is withdrawn. Original logs retained. Attempted1151 overlay failed wrapper invocation then compile, not acceptance runs; clean exact-baseline worktree1152 provided corrected RED. Public mechanism first proof also exposed stale virtual Max estimate and Tester Shift omission; original failures retained, not masked by tolerance/timeouts.

Initial v1 production focused JSON9top/15inclsub each3run/3pass (superseded below), zero skips/failures (manifest exactnames). Includes actualApp owner changes, live Thinking growth/detach/rearm/reverse, real continuous body selection+Copy, native menu/wheel/selection and keyboard controls, default no-opt-in follow, precise residual and ordinary sticky control. Callback unit checks non-scroll key/compose/pointer/select/command/zoom returnfalse, stale admission, zero-residual validation, Shift and NaN.
Separate test-only public mechanism9cases remains visibly test-only: precise/notch, horizontal fallback, short inner, no outer range, two edges and virtualList saturation, owner rejection. Production full suite and hosted native input NOT_RUN here. UI/transcriptui vet and DarwinARM64 crossbuild explicitEXIT0. Lead owns aggregate and publication; no real RPC/auth/installer or user files used.

## Final consume-stale correction (11:59 UTC)
Independent review found that callback rejection returningfalse allowed the default native scroll chain to consume the SAME stale event. Actual committed-old-Thought tests reproduced old inner655→645 after owner reset and actual App new-runtime outer1638.171→1688.171 before the next frame. Direct callback tests alone had not covered fallthrough. Original1158 RED retained.

Only rejected **valid opted-in scroll** now returns true (consume) without modifying inner/outer/unread; nil opt-in and all unrelated events remainfalse. The Options contract explicitly distinguishes rejection of state change from event consumption. The host callback still returnsfalse for stale owners; only the input handler prevents unsafe fallback. Both real old-hit-tree REDs now pass. No blanket native-event interception.

Final v2 JSON11top/17inclsub, every3run/3pass, zero skips/failures; UI7.617s, explicit focused/vet/ARM64 EXIT0. Complete final manifest preserves v1 evidence as superseded rather than reusing it. No full suite, publication or hosted-native claim.
