# P20 bounded accessibility name correction: workspace Reset

Only workspace appearance Reset is changed. This is not a complete a11y or P20
claim. Base is approved P20 second batch; integration trees remain untouched.

## Exact pinned MyGo source contract

MyGo v0.6.4-0.20261025001514-9c69cf829ece:
- ui/element.go:1293–1295: Label names the element for assistive technology and
  finding it in tests, not merely a private test ID.
- ui/access.go:295–297: AccessNode.ID comes from element ID; Label comes from
  element label. At :384, inner text is a fallback only when Label is empty.
- internal/darwin/access.go:245–253: the Label becomes AppKit accessibilityTitle
  for controls, accessibilityLabel otherwise. Therefore an English explicit
  Label overrides translated child text in the actual bridge.
- ui/element.go:647–660 and ui/context.go:254–286: Key supplies stable sibling
  state identity through parent+key, independently from Label. Use Context.Key
  before construction for controls whose state is initialized immediately.
- ui/api_gen.go:4072 exposes Element.ID; ui/ref.go:118–150 exposes Bind,
  Context.Resolve and Focused. No private state access is necessary.
- ui/headless.go:444–451 and :804–821: public Find/Click/Focused select labels or
  rendered text, not stable keys. They do not expose an accessibility-tree read.
- ui/element.go:954–959 explicitly limits Access callback to toolkit packages:
  its node type is internal platform.AccessNode. Toolkit tests access headless
  accessTree through private helpers; Orbit must not copy that privilege.

## Minimal change

The shared real Reset button helper uses Context.Key("workspace-appearance-reset")
and the translated sidebar.reset Label supplied by the normal button renderer.
Remove only the explicit English Label override. Existing English tests still
find Reset in English. The Chinese P20 test now selects its expected translated
name, rather than requiring the product to keep English indefinitely.

The helper is used by production and isolated element-state tests, not a dummy
parallel implementation. No toolkit patch, reflection, unsafe/private import,
AX bypass, new test-only production flag or global selector migration.

## Regression evidence and limits

Mounted Chinese popup before correction still exposed Find("Reset") despite
painting Chinese, and the new regression failed (`reset-a11y-red-1918.log`).
After correction, Chinese Reset is actionable, English Reset is absent from
public labels, and exact workspace icon/tint clears in the synthetic saved list
without changing other entries. A held press followed by A→B→A invalidation must
not mutate any saved workspace.

A real helper mount records public Element.ID across English→Chinese→English and
inserting/removing a preceding noninteractive sibling; focus and Enter action
must remain. This proves stable app element identity independently of text.
It does NOT directly inspect a native OS AX tree. Source bridge mapping plus
public semantic label/focus behavior are the available evidence; physical macOS
VoiceOver and OS AX acceptance remain unverified and must not be reported PASS.

Focused race (2x) passed 9.515s. Removing only the explicit Key caused the
identity/focus regression to fail on sibling insertion with different public IDs
(`reset-a11y-key-negative-1919.log`); the production Key was restored afterward.
Final full UI/assets race and vet run against the frozen source are recorded in
`reset-a11y-final-{full,vet}-1920.log`; inspect their terminal status before use.
