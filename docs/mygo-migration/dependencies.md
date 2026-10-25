# Go migration dependencies

All module versions and integrity checksums are pinned in `go.mod`/`go.sum`.
The original Orbit Apache-2.0 license, asset notices and bundled extension
attributions remain intact.

- MyGo, MIT: the requested native Go UI framework, with native windowing, text,
  markdown, editor, clipboard/dialog and terminal APIs. Its MIT notice is retained
  at `licenses/mygo-MIT.txt`. No webview path is used by the Go app.
- nativewebp v1.3.0, MIT: lossless alpha-preserving image attachment encoding with
  no cgo or Rust dependency. Required to preserve Orbit's existing WebP policy.
  Notice: `licenses/nativewebp-MIT.txt`.
- golang.org/x/image: native image decoding and Triangle/Bilinear downsampling.
  Resampling/codec output is behaviorally equivalent, not byte-identical to Rust's
  image crate; pixel acceptance must be measured rather than inferred.
- gopkg.in/yaml.v3: actual YAML frontmatter parsing for user `SKILL.md` files,
  including block scalars and quoted values. A line-split substitute would reject
  valid skills. Locale assets use their existing generated flat-key parser.
- fsnotify v1.10.1, BSD-3-Clause: native filesystem notification backends for
  Linux, macOS/BSD and Windows. Orbit adds recursive directory registration and
  trailing debounce; resource/backend failures use bounded metadata polling.
  Polling cannot detect same-size writes that preserve modification time.
  Native notifications on NFS/SMB/FUSE are not guaranteed by fsnotify.
  Notice: `licenses/fsnotify-BSD-3-Clause.txt`. Official versioned reference:
  https://pkg.go.dev/github.com/fsnotify/fsnotify@v1.10.1
- purego, go-text/typesetting, x/net and x/text are pinned transitive MyGo modules;
  x/sys is required by the native filesystem watcher.
- MyGo's terminal plugin downloads its pinned native Ghostty VT library on first
  use. This is the upstream framework's native terminal engine, not a Rust app
  sidecar. Tests distinguish its real PTY integration from domain-only tests.

Distribution packaging must retain the dependencies' license notices, original
asset notices and the native Ghostty runtime's own license. No release or binary
installer is published by this checkpoint.
