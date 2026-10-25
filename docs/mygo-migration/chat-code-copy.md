# Chat code-copy source parity (copy-only)

Pinned Rust `transcript_view.rs:5841–5867,6549,6589–6597` accepts an
unclosed code fence and copies `lines.join("\n")`, regardless of live state.
Pinned MyGo `markdown/code.go:26–27,116–139` copies Document.Lines (an
artificial trailing newline per line), and disables its button for Open blocks.
Actual App.View RED fixtures proved added-newline and open-live/open-settled
copy failures; empty code already worked.

The small Orbit header/copy replacement uses the public Options.CodeBlock hook
and delegates the entire body to b.Body(c): no lexer, code layout, scroll,
selection or toolkit default changes. It removes exactly one artificial final
newline with TrimSuffix, preserving indentation, trailing spaces and genuine
final blank lines. Code() is called only for actual activated copies, not idle
frames or feedback repaint. The button remains active for open/live blocks.

`chatMarkdown` supplies immutable source identity to a node-local state holding
its own Options in place (never copies the internal sync.Map). Public rendering
options propagate individually. Caller slot identity is preserved outside the
owner-keyed subtree; multiple sibling prose runs cannot collide. Source changes
increment a small revision; only the header subtree is revision-keyed and its
activation checks revision plus real task/runtime/window ownership. The body
keeps its toolkit identity across source updates. New content/session cannot
retarget an old held copy press. Source strings are compared by value, never
hashed or stored in element keys. An unchanged frame does not build Code().

Both legacy/simple and logical RPC transcript rendering use this wrapper;
GitHub alert bodies explicitly recurse through it. Standalone Files markdown
preview is not globally switched to the new code header. No URL/image decoder,
loader or permission path is broadened. Existing hooks and per-view image state
remain intact. All clipboard tests use the native tester clipboard, never an
external desktop clipboard.

Fixtures cover the original six cases plus CRLF, Unicode/tabs, multiple blank
last lines; streaming append→close→settled, alert-contained code, logical RPC
projection; same-label changed code, same-ID runtime replacement, observed A→B→A
and Close; two sibling code views/keyboard focus; code/prose cross-message
selection and actual horizontal scroll; zero payload reads during idle and copy
feedback frames. Source text is never rewritten or evaluated.

This increment does not implement >28-line fold-to24, friendly language labels,
or declare code block pixel parity. It retains the raw language caption and
uses localized existing Copy/Copy code labels with checkmark feedback. Native
same-state capture remains separate. The additive Rust parser/join test change
is provided separately for review and execution using the existing reference CI
toolchain; no local Rust pass is claimed.
