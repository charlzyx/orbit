# Reconstructed composer preview admission

This is a fresh reconstruction after the temporary executor reset. The old
Wave11 test result does not validate these new bytes. The three attachments
package files were recovered byte-for-byte from remote tree
`5f4b4838a41e52c7346ee0980f156ef9c7099f04`; this candidate adds the missing UI
preview helper, its ingress hook and new tests.

Preparation can preserve an oversized or undecodable input as its raw payload.
The UI bitmap preview now separately calls `image.DecodeConfig` and the recovered
`attachments.DecodeDimensionsAllowed` before `ui.DecodeBitmap`. The original
268435456-pixel threshold and division-based overflow-safe predicate are unchanged.
This may still admit about 1 GiB of RGBA pixels alone; it is not a total-memory cap,
a decoder cancellation mechanism, or a lower safety threshold. DecodeConfig itself
is not promised constant-cost. Existing SVG parsing remains a separate unchanged
branch. No remote/Markdown image pipeline or image-format policy is changed.

A 33-byte PNG IHDR fixture declares 32768² dimensions. Tests only parse its config
and exercise rejected-preview/raw-payload preservation, never decode giant pixels.
A temporary guard-disabled negative control ran only the injected fake-decoder
unit test: all five bad-header, zero, negative, over-limit and overflow configurations
were rejected by assertions when the fake decoder was reached. The production
guard was then restored byte-for-byte. Actual tiny PNG/JPEG/GIF/BMP/WebP/TIFF inputs
exercise the real AddImage pipeline and preserve MIME/base64 bytes; malformed and
unknown images retain their raw attachment with a nil preview. TIFF is registered
in the full App (`internal/ui/files.go`), and GIF remains the first-frame decoder.

Final focused results are recorded in the sibling recovery proof manifest. No
complete UI suite, native image dialog, model request or actual macOS validation
was performed by this worker; the lead must run the final integrated gate.
