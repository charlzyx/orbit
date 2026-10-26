# Source bottom fade: narrow production candidate

Base13eea694145d80d861a0276c8845bab0cdd7744f. Only chat.go gains a five-line decorative sibling before Jump. Original transcript scope/List ancestry/keys and scroll algorithm are untouched. No toolkit extension, network/image renderer, fade bitmap or global Clip.

While away from tail, a full-viewport absolute Clip+PassThrough wrapper contains the fixed48 DIP bottom0 gradient (same background RGB, alpha0→original alpha,180°). It is clipped within its own wrapper, including viewports below48, not rescaled. Jump is the next sibling and retains its fill/foreground and input. Source refs are pinned Rust299 message_scroller292–329; official GPUI0.2.2 window1215 default rem16 and absence of Orbit rem overrides. MyGo public Gradient180 uses the same axis. Header/sidebar/composer are not ancestors of the fade clip.

The first test-only after-App injection probe is archived verbatim in fade-after-App-probe-archived-1110.go.txt. It drew over Jump and is not used as final production ordering evidence. Its prior success/failure logs remain in fade-feasibility-1108.md/evidence. The standalone fixed48/tiny24 crop mechanism test remains; production tests now inspect actual messageTranscript paint/bounds and actual App controls.

Actual production light/dark CPU images at /tmp/orbit-fade-production-1112/production-fade-{light,dark}.png show faded lower text while the Jump fill remains its exact opaque surface color (pixel assertion). The initial current-production focused matrix passed5.595s, including normal/tiny fade bounds and existing real wheel/selection/context menu/Jump pointer and Space controls. Final exact-name regression evidence is separate.

The tiny production-render fixture constrains the real messageTranscript to24 DIP (not a complete useful24px-high app). It proves the fade itself stays within that viewport. It also reveals the previously present32px Jump circle extending above such a tiny viewport; Rust clips its whole shell. This candidate does NOT change Jump's clipping or claim all tiny-viewport chrome now matches. That independent geometry boundary remains explicit, avoiding a global selection/scope Clip change.

No full suite, hosted-native pixel acceptance, publication or release is performed here. CPU pixels and public measured geometry are local evidence only.

Final focused gates: 20 exact source-selected top-level names / 32 including subtests, each run/pass3, no skip/fail, 36.309s; scoped UI vet and Darwin ARM64 crossbuild explicit exit0. Final actual-production image hashes and hitsets are in fade-production-1115-evidence.json. Full aggregate remains lead-owned.
