# Reconstructed focused composer turn navigation

Fresh reconstruction on exact published `cee007d9309815489bfaa11529fa3047a2b612f8`.
The old Wave11 full-gate result is historical, not evidence for these new bytes.

Pinned Rust299a32b `crates/orbit-pi/src/composer_send.rs:429–447` requires the
secondary Up/Down actions while Composer owns focus. `app/session.rs:2772–2792`
uses the same previous/next turn actions without an empty-draft requirement.
The Go text editor consumes modified arrows before global shortcuts, so the
composer OnTextKey now routes only exact Cmd-Up/Down to the existing commands.
Attachment path flushing and autocomplete keep precedence; Enter dispatch,
Home/End, other modifiers, global shortcut ordering and toolkit code are untouched.
Active composing input falls back to the original editor route rather than
intercepting IME. This is preservation of the existing Go composition behavior,
not a claim of physical platform IME parity.

The new real mounted full-App key regression initially failed on unchanged source:
Ctrl/Cmd-Up left the visible row at21 instead of preceding user row20 (1.845s).
After the small source hook, the same assertion passed. The five named tests cover
empty draft, middle/edge anchors, multiline Unicode selection followed by actual
replacement/Undo, IME and plain/modifier negative controls, real global keyboard
shortcut outside the mounted composer, and autocomplete plain-Down precedence.
The final user turn clamps at the real bottom; its boundary assertion compares
the same direct ScrollTo geometry rather than requiring impossible top alignment.

The earlier hosted Go CHAT02 navigation failure remains a failure. This worker
performed no fresh native capture or keyboard action, and focused mounted tests
are not native acceptance. Final focused evidence is recorded separately; final
integrated full-gate and any approved exact-SHA native recapture belong to lead.

The reconstruction conservatively leaves the key to the editor if the public Composing capability is unavailable, rather than consuming unknown composition state. The pinned required composer overlay supplies it, and the real pinned-interface tests cover the normal path. This is an explicit reconstruction fallback difference, not a toolkit change.
