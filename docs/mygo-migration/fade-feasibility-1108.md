# Bottom-fade public-API feasibility (historical test-only stage)

This records the earlier feasibility stage. Current production candidate and final ordering tests are described in fade-production-1113.md; the after-App probe is archived, not used as production proof.

Base13eea694145d80d861a0276c8845bab0cdd7744f. No production call site, global scope Clip, toolkit patch, full suite or release. Source audit and hashes: /tmp/orbit-fade-audit-1101/plan.md and manifest.json. Pinned Rust299 message_scroller320–329: bottom0 fixed3rem gradient180, alpha0→background; default GPUI0.2.2 rem16 yields48DIP, with no Orbit override found.

Test-only helper: separate absolute full-viewport wrapper Clip/PassThrough, containing its own fixed48-high bottom0 PassThrough gradient. It never reparents the List or gives the original scope a clip. At viewport24, raw Element.Bounds stays48 while visible Tester.Find bounds is24. Pixel crop starts in the lower half of the same gradient, not a newly rescaled24px gradient. Outside the wrapper remains untouched; equal-y samples across x are identical (vertical axis).

Initial /tmp/orbit-fade-mechanism-1104.jsonl incorrectly expected Tester.Find to return raw48 despite clipping; tiny case failed. Public raw Element.Bounds and clipped visible bounds were separated; /tmp/orbit-fade-mechanism-v2-1104.jsonl passed. This was fixture correction, not a production fix.

Actual App paint/input probe uses a fresh same-sized wrapper placed over the measured transcript viewport **after App.View in the test**. Light/dark before/after CPU images demonstrate the currently missing bottom fade and that the public primitive can create it. Because the probe is drawn after the App it also fades Jump: these are explicitly not final production ordering screenshots. Any production candidate must insert fade before the existing jump sibling, then verify Jump remains unfaded/topmost. The current actual App probes verify wheel reaches the transcript and exact message-menu targeting and selection remain available beneath the decorative layer. No raw user data or real clipboard is involved (Tester synthetic messages only).

Evidence /tmp/orbit-fade-input-paint-1107.jsonl:2 top-level tests/6 including subcases, each run/pass1, no skip/fail,6.221s. This is mechanism feasibility, not production RED→GREEN or hosted-native acceptance. Four actual-App comparison images plus24/80 crop images are under /tmp/orbit-fade-captures-1107. Future native/reference matching remains separate.
