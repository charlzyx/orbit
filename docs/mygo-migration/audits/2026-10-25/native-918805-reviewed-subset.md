# Hosted native subset actually reviewed: 918805

Source `918805045b7e9c77ab252f72a70652cfb8438920`, run `40228444291`, Go artifact `12612676312`; downloaded ZIP SHA256 `b06650db029f55305e12f1d19e6c5bf257c3a1fc750ef7d77c95e35d3d810fa2`. Artifact manifest confirms this exact source, 1470×922 Xvfb/Openbox Linux software graphics, isolated credential-free fixture. This is not macOS physical-input acceptance.

All four jobs (required proc regression, aggregate, Go native and Rust reference native) passed, verified at 20:09 UTC. Actual manual image review in this checkpoint covered **two** images, not every emitted state:

- **WB05 Review** (`f2673656607211b8b0a42344fc480e3b52ac06d3a97e6eb4586fbee178c63b1b`): source-style single gutter and old/new row numbering are visible. Notes and the added return line still wrap where the pinned Rust fixture fits; overall layout/font parity is NOT established. Current Go file headers still lack the separately staged next-wave glyph/count refinement. Both the old same-fixture Rust image and the current same-run Rust raw image were actually viewed, without cropping or rescaling. Current reference artifact `12612726257` ZIP SHA256 `70ca6035360a3ea0a6841f409d1ac63f4479447467676a722eb4b66b9dad3410`, manifest source `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, confirms the same 1470×922 viewport. That comparison is not a claim that every volatile runtime variable was frozen. Exact available-width/font/wrapping cause is under source investigation, not assumed fixed by changing a declared font size.
- **WB08 Terminal** (`be3159262ede6ce88e67cd083fdb300374c539e1f59726c3ca8e34280f0808fd`): the workspace tab now visibly has its own close control; the toolbar separately shows add/restart/hide. The fixture prompt is visible. This screenshot does not prove close/hide input behavior; disposable multi-PTY mounted tests provide that separate evidence.

The app log contains the expected software-renderer no-GPU notice, not the earlier pre-init null callback symptom. No conclusion about the user's macOS picker hang follows from this log.

Raw local archive: `/workspace/shared/orbit-native-918805-2001/go.zip`. Static Chat opt-in failure at the previous commit is separate and remains blocked before input; this normal source run does not enable that opt-in.
