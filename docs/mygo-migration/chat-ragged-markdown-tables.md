# Chat ragged Markdown tables — bounded content recovery

Pinned Rust `299a32b` stores every body cell (`transcript_view.rs:5899–5923`),
and renders the maximum header/body column count (`:6863–6886,6951–6979`).
Pinned MyGo fixes columns at the separator and drops additional cells in
`parser/block.go:1328–1353`. This is a source-compatibility difference on ragged
tables, not a claim that MyGo violates GFM.

Orbit's `markdownRaggedTable` intercepts only an existing Table whose raw body
has additional columns. The parser's Table.End ends at the last retained cell;
the hook extends only to the end of that same physical line, never the next
block. `rustTableCells` follows Rust's naive pipe split, deliberately not a new
GFM parser. It builds a presentation-only table with empty added header cells
and additional separators; original messages, search and Copy Message remain
unchanged. Ordinary tables still use the untouched default renderer.

The projection is rendered with the public default table renderer: cell text
cannot become headings, lists, fences or arbitrary nested blocks. Block and
resource loaders are absent. The existing default table itself renders images
as `[alt]` via spanBuilder, not image loading, so this does not disable an
existing table-image feature. Rust instead interprets an inline image as `!`
and a linked label: that pre-existing inline difference remains, not claimed
fixed here. Links dispatch the same OpenLink action to the existing controller
scheme/path rules; no URL launcher, loader or permission relaxation is added.
The callback and keyed subtree use current task/runtime/window ownership to
reject old held clicks after same-ID owner transitions.

ShareSelection keeps the actual transcript's selection/copy owner. It disables
MyGo's separate CopyData/Copy as Markdown path (`markdown/view.go:169–184`), so
synthetic separator/header syntax is not exported by a nested markdown owner.
Mounted tests cover cross-cell and cross-message copy with no synthetic syntax,
link actions, stream/settle, following paragraph preservation, original message
identity, stale held pointer and fresh activation, inline-only cells, and scroll.

Known remaining scope: exact table geometry/width breakout, complete Rust inline
syntax parity, code-copy newline/live availability, and code folding. No native
pixel acceptance or full C09 completion is implied.

## Pinned Rust executable contract gate (pending hosted execution)

`scripts/markdown-parity/check-rust-table.py /tmp/orbit-reference` validates the
reference file SHA and verbatim extracted parser segments before invoking the
already-installed Rust compiler on the stdlib-only fixture. No Cargo packages,
network, installs, GUI actions or credentials. Proposed insertion: existing
reference-desktop job after pinned reference worktree creation and Rust1.94
setup. Workflow not changed by this patch. Local environment lacks rustc;
Python syntax check is not a Rust pass. Hosted gate must execute before claiming
runtime-proven Rust parser equivalence.

## Resource and content-revision guards (review revision 2)

Projection is capped at 16,384 padded cells and a conservative 1 MiB output
budget. Division checks precede rows×columns multiplication; an eight-byte
per-cell allowance bounds generated separators/row framing/delimiters. Oversize
ragged tables do **not** fall back to the truncating parser: the complete raw
table source is shown as selectable text with an unselectable layout-limit
notice. No rows/cells are silently omitted; original message storage is intact.
The notice is currently English (new local diagnostic, not claimed full locale
parity). Its raw text fallback has no clickable interpretation/resource loading.

A real held-press RED demonstrated that owner-only keys were insufficient when
same-task/same-label link content changed. A node-local raw-content value and
monotonic revision now key the inner subtree and guard link dispatch. Equal
content retains the cached projection; no per-frame payload hashing or large
content-derived element key is used. The allocation-free shape scan still runs
to determine whether a table is ragged; normal tables return before creating a
custom subtree. Native press/release across URL replacement must be rejected,
while the fresh replacement link remains usable.

Review revision 3 also budgets **every raw output row**, including header and
separator, while only extra body cells activate the hook. A real pinned-parser
fixture proved the distinction matters: an escaped-pipe header still parses as
two MyGo columns, but Rust's naive split can produce more than16,384 raw header
slots. That case now uses complete-source fallback. Actual mismatched unescaped
header/delimiter counts and escaped invalid delimiters do not produce a MyGo
Table; normal escaped headers with ordinary body widths remain untouched.
