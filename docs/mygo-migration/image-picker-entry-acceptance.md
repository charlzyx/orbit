# C04 source image-only picker entry

Rust `composer_ops.rs:435–466` opens an asynchronous single image picker with
exact filters `png,jpg,jpeg,webp,gif,bmp`, rejects admission at eight existing
attachments, and uses the existing image attachment pipeline. The Go plus menu
previously exposed only the general multi-file chooser.

The new Attach image entry dispatches `BrowseImage`; native options preserve
parent/current directory, single selection and the source extension filter.
The controller uses the already-reviewed shared native dialog singleton and
opaque task/workspace/window/close guard. Cancellation is a no-op, current-owner
errors are surfaced, and obsolete errors are discarded. Unexpected non-image
paths never become prompt text. Only the first result can enter the image
pipeline; admission is checked before opening and existing pipeline checks
preserve the eight-image cap if another attachment arrives while selecting.

No decode/compression/preview/Markdown changes are included. In particular this
increment does not establish new guarantees for malformed image contents in
the pre-existing attachment decoder. It does not claim physical native panel
confirmation, user Mac IME, or the original reported hang has been verified.

Tests use the real menu dispatch, exact option assertions, a real temporary PNG
and existing async decode path, held native wait fixtures (nonblocking call and
singleton), cancellation/error/reopen, stale task/window/queued-close results,
unsupported extension rejection, and both admission/completion cap checks.

Callback validation uses the same six suffixes (case insensitive), not the
broader workspace preview-image predicate. Rust `app.rs:995` only admits those
formats from files; SVG/TIFF passthrough belongs to already-supplied clipboard
images. An unexpected SVG/TIFF first result is rejected even if a later result
looks supported. General file-drop and clipboard/decode policies are unchanged.
