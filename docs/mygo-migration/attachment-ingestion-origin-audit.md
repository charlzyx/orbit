# Attachment ingestion: source origins and RED evidence

Read-only production baseline: `2092a6efa3f6eb57efb8713e1df6e6099ec3232b`,
frozen wave8 snapshot plus reviewed pending-submit ownership. This tree currently
adds only regression tests and this audit; no production queue fix is included.

## Exact current callsites

| Entry | Current route | Required source policy |
| --- | --- | --- |
| External drop | App.View root.DroppedFiles → App.AddPaths | Ordered drop batch; before every path, max-eight stops the remaining batch |
| General file picker | composerAddMenu BrowseAttachments → Controller.Action → browseAttachmentsDialog → App.AddPaths | Image or caret-path reference; ordinary paths still work at eight images |
| Image-only picker | composerAddMenu BrowseImage → Controller.Action → browseImageDialog (six suffixes, first result) → App.AddPaths(paths[:1]) | Image only; failed/unsupported selection reports error, never prompt-path insertion |
| Clipboard image | composer OnPaste / Paste image → PasteImage → AddImage | Existing image ingest/cap, no path fallback |

Current native BrowseAttachments uses a multi-select dialog; pinned Rust
attach_file uses pick_file. This audit does not change that separate chooser
surface or silently reinterpret it as an external drop.

## Pinned Rust policies

* `app/composer_ops.rs:405–433` processes ExternalPaths sequentially. The cap
  check is before every path, not only images; after reaching eight, subsequent
  ordinary paths are not inserted.
* `app.rs:993–1010` accepts exactly png/jpg/jpeg/webp/gif/bmp for file images.
  Failed file reading returns None. A successful read remains Some even if
  compression/preview must preserve raw bytes; decode fallback is not a failed
  file read.
* on_file_drop handles None by inserting that exact path plus a space at the
  caret, focuses the composer and continues without using an image slot.
* `attach_file`, composer_ops.rs:468–496, has the same None→path behavior but no
  drop-loop pre-cap: a plain file can still be referenced at eight images.
* `attach_image`, composer_ops.rs:436–464, reports unsupported_image_format for
  None, and does not insert the path. Its existing pre-dialog max-eight check
  remains independent of external-drop truncation.

Therefore a single generic failure boolean/cap rule on AddPaths is incorrect.
An eventual queue needs explicit typed origin, with the original caret/IME queue
retained for all genuine path-reference operations.

## Controlled regression evidence

`/tmp/orbit-attachment-order-origin-final-red.log`: focused race repeated three
times, 11.857 seconds, intentionally FAIL against unchanged production:

1. Hold slow.png's read; allow fast.png's actual owner completion, then release
   slow: attachments become fast,slow instead of input order (all three runs).
2. Actual mounted root DropFiles at eight images inserts tail.txt.
3. Seven images + first.png + second.png + tail.txt reads both images and inserts
   tail, instead of accepting first then stopping the batch.
4. Seven images + missing.png (read error) + good.png + after.txt drops the failed
   path and inserts after, instead of inserting missing then stopping after good.
5. Injected real general-picker completion with an unreadable image loses its
   required path fallback. Controls pass: image-only failure does not insert a
   path; general file picker at eight images still inserts an ordinary path.

No native dialog, actual clipboard, model, network, installer or user file was
used. Reads use injected fixture bytes/errors or missing paths in a test tempdir.
The queue/busy-admission design is pending separate review; these RED tests must
not be integrated as a green production patch by themselves.

## Isolated implementation (not OS acceptance)

`attachment_ingress.go` uses explicit Drop/FilePicker/ImagePicker/Clipboard
origins and one owner FIFO. Reads and decode stages are serial within that FIFO;
clipboard and picker arrivals cannot overtake a held drop. A pending-submit token
covers the entire admitted request, including the read-to-decode handoff and
queued result. Existing caret/IME/undo operations remain in attachment_path_edit.
Only Drop stops all remaining paths at eight images. General picker read failures
insert paths; strict image-picker failures retain their existing error feedback.
The pre-existing pending-error test now explicitly uses the strict image origin,
rather than requiring the incorrect dropped-file error policy.

Deliberate Go safety limits, NOT Rust/OS limits: at most 256 paths and 256 KiB of
UTF-8 path bytes per request, eight outstanding ingress requests, two retained IO
pipelines, 64 MiB per clipboard item and 128 MiB total retained clipboard input.
Count/byte checks precede copying, pending-token reservation and draft mutation;
byte comparisons subtract from remaining capacity to avoid integer overflow.
Admission failure rejects the entire new request with one fixed busy/limit hint,
never silently truncates it or automatically retries/sends it. The existing file
read and decoder limits remain unchanged. The native general picker remains its
existing multi-selection surface (Rust attach_file uses single selection).

An IO slot spans read, decode and owner consumption, bounding ready-result buffers
as well as active workers. Cancellation detaches the old draft immediately, drops
queued (not running) requests, and keeps the active request's slot/raw accounting
until its actual completion is consumed. Its result can only release its own
resources, not append to/unlock a newer owner. Two retained old pipelines reject
another image request rather than spawning unbounded blocked reads. This does not
claim arbitrary filesystem IO is forcibly cancelable. No global mutex or owner
wait is introduced. No real clipboard, dialog, model or user files are exercised.

The 128 MiB budget is clipboard raw input only, not a total-memory guarantee:
file read results, Base64 and decoded previews are not charged to it. The existing
Stat-before-ReadFile check is not an authoritative 64 MiB bound for a growing or
nonregular file; an actual bounded reader is a separate follow-up. Two IO slots
bound simultaneous pipelines, not their total resident bytes. No new decoder,
SVG or pixel-bound guarantee is claimed by this queue change.
