# Embedded diff retained-context expansion

Pinned Rust source `sidepane.rs:1750–1854` defines the same diff gap controls as
Review: leading → end, trailing → start; between ≤100 → both, larger between →
start/end. The label reveals both edges; Shift on label or directional gutter
reveals all retained lines. Missing retained context must remain inert. Source
`2251–2258` maps start/both to down-chevron and end to up-chevron.

Previously the embedded renderer exposed all four above/below/both/all buttons
regardless of gap position, and had no source-equivalent label/Shift action.
The replacement uses actual native buttons, a 32px row and 46px bordered gutter.
The two-control first hit target is fixed to the source's 16px height rather than
letting native flex growth split border-inclusive heights to 16.5/15.5. Interior
button width is 45px because the outer gutter includes its 1px right border.

Actions re-resolve the stable gap ID and file inside the staged/unstaged snapshot;
there is no reread of changing disk contents. Element identity includes epoch,
snapshot and gap identity, so a held press cannot expand a replacement snapshot.
The existing long-line width calculation still includes retained hidden context.
No selectable text/copy handler, virtual-list anchor or workspace data loading is
replaced by this patch.

Real mounted tests cover staged/unstaged × unified/split × edge direction;
label both/Shift behavior; small-gap single control; gutter Shift and repeated
expansion; unavailable context; exact gutter/row/hit-target geometry; and held
press across snapshot replacement followed by a normal fresh click. Initial
`embedded-gap-red.log` shows missing source action targets before implementation.
Geometry tests caught native border/flex interactions rather than merely counting
controls. Final package race/vet results are recorded with the handoff.

This is toolkit input/layout evidence, not a native macOS image comparison or
completion of all Review/Git parity. Existing font/theme differences remain
outside this bounded behavior correction.
