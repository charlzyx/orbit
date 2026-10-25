# Chat code captions (independent of folding)

`markdown_code_caption.go` mirrors pinned299a32b `highlight.rs:123–164`
`language_label` aliases exactly. ASCII lowercase and surrounding whitespace
normalization apply to known aliases; an unknown tag is returned literally,
not normalized or replaced. In particular Scala/Objective-C tokenizer aliases
are not invented as caption mappings. Unicode case-fold equivalents do not
expand Rust's ASCII-only normalization.

`transcript_view.rs:6738–6748` supplies the remaining rules: bare fences use
existing `transcript_view.plain_text`; Mermaid case-insensitive uses existing
`transcript_view.mermaid_source` (English: “Mermaid · source (preview unavailable)”).
These locale keys are reused, not new hardcoded translations. Only the visible
header text changes. Parser/tokenizer language, accessibility group identity,
body/highlighting/selection, and code-copy payload remain untouched. Unknown
languages retain their raw label. No Mermaid renderer or fold behavior added.

Mounted regressions cover ts/rs/bare/Mermaid/unknown, code copy unchanged, and
live English→Simplified-Chinese captions for the localized labels. The initial
focused command did not run these caption tests; its caption PASS claim is
withdrawn. Corrected exact-name JSON evidence is recorded below; UI vet passes.
This is not native pixel acceptance or full code parity.

## Locale test evidence correction

The initial focused pattern anchored shortened names rather than these tests'
complete names, so its successful exit only covered the code-copy regression.
The earlier caption-locale PASS claim is withdrawn; the original command/log
is preserved. Exact tests then reproduced the same missing catalog reload in
both the original isolated caption tree and the combined tree: assigning
Prefs.Language alone is not the Appearance language setter.

The corrected mounted fixture uses the real Settings → Appearance language
selector, chooses 简体中文, verifies the temporary Store persisted zh-CN, and
returns to chat before asserting both the non-English translation and rendered
caption. No production locale/catalog behavior or Chinese assertion was changed.
The corrected `-race -count=5 -json` run explicitly selects
`TestSourceCodeCaptionsPreserveUnknownAndTokenizerTag` and
`TestMountedCodeSourceCaptionsAndLocale`; all seven expected top-level/subtest
names have exactly five run and five pass events, with zero skips/failures.
Evidence: orbit-caption-locale-corrected-json-2131.log and
orbit-caption-locale-hitset-2131.json (UI6.756s). Focused results must be checked
against the intended test-name set, not inferred from exit0.
