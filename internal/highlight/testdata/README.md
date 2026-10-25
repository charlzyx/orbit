# Pinned Orbit lexical behavior oracle

These 1,204 cases are recorded outputs from `crates/orbit-pi/src/highlight.rs` at
Orbit commit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, SHA-256
`d4c97a77caa49da5d628be3d7657aa523865c3fc65a43a539859b09e4200e65a`.
They are not produced by the Go implementation. The source's 15 built-in Rust
unit tests also passed in the cloud with rustc 1.94.0, edition 2021.

Run `python3 internal/highlight/testdata/generate.py` with `rustc` on PATH
(or set `RUSTC` to its executable path). The script refuses a changed source hash,
compiles a small independent stdin/stdout harness, runs each case against the
pinned Rust module, and writes one fixture per language. Python and Rust are only
needed to regenerate fixtures; ordinary Go tests do not invoke either tool.

Coverage: all 16 language vocabularies and classification precedence, comments,
all quote slots and invalid carried slots, numeric prefixes/exponents/suffixes,
Unicode byte offsets, shell command/flag/path/assignment/operator/variable
roles, diff whole-line classes, unfinished streaming input and empty lines.
The Go tests additionally check every valid Unicode prefix, lossless paint runs,
invalid external spans, malformed file bytes, concurrent cache access and bounded
eviction. This is lexical compatibility, not proof of whole-GUI pixel parity.

The Go scanner is independently written against observed Orbit behavior. Orbit's
source header says it was adapted from MIT-licensed Waku, but Waku currently
publishes GPL-3.0-only and its first root LICENSE commit found in the public history
is GPL-3.0. That historical MIT provenance remains unverified. No current Waku
implementation is copied into the Go package; the inherited Orbit NOTICE has not
been altered. See `docs/mygo-migration/highlight-parity.md` for the audit caveat.
