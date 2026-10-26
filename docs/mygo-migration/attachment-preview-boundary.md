# Composer bitmap preview: close raw-fallback dimension bypass

Scope: composer local attachment preparation/preview only, not Markdown/remote
message images, SVG parsing, file reads, native picker behavior or RPC admission.

Prepare previously returned raw bytes when DecodeConfig exceeded 268435456 pixels,
but the next composer step unconditionally passed those raw bytes to MyGo's
DecodeBitmap, which directly calls image.Decode. The preparation bypass was not
a preview allocation guard. Prepare's multiplication is also replaced by the
same positive-dimension/division predicate, avoiding signed-product overflow
without changing its existing threshold.

Before bitmap preview decoding the composer now independently checks config with
that shared predicate. PNG/JPEG/GIF/BMP/WebP/TIFF are registered in the actual App process (TIFF via
ui/files.go and background/photo.go blank imports). GIF preview remains
first-frame only, not DecodeAll. Unknown/unregistered, bad-header and
over-threshold inputs have no bitmap preview.
They retain exactly Prepare's output bytes/MIME/Base64, consistent with pinned
Rust attachments.rs compress_image raw passthrough and app.rs from_bytes payload
semantics. No fallback image is silently removed from the outgoing prompt.
SVG keeps its original ParseSVG path/budgets, untouched by this bitmap change.

**This is not an overall memory cap. The unchanged 268435456-pixel threshold can
still allow about 1 GiB of RGBA pixels alone, plus decoder/intermediate/preview
allocations.** DecodeConfig itself is not promised constant-cost. Neither this
check nor owner cancellation interrupts a decode that has already started. The
existing FIFO limits concurrent pipelines and owner result application, not total
decoded memory. A lower product safety threshold is a separate decision.

Safe evidence: an extracted preview seam's fake decoder RED was called for all
five rejected configs. New tests require zero calls for invalid/unsupported/
over-threshold/overflow configs and one fake call at the unchanged exact limit.
A 33-byte PNG IHDR advertises 32768² dimensions without pixel data; only config
and raw-fallback behavior are exercised, never a giant decoded allocation.
Actual tiny PNG/JPEG/GIF/BMP/WebP/TIFF attachments keep real previews and exact wire
payload; unsupported/malformed payloads remain byte-exact with preview=nil.
No giant image, model, network, real OS dialog or user filesystem was exercised.
