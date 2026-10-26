# Single-image mixed clipboard precedence (2026-10-26 12:27 UTC)

Exact base fe034cd0cbfcd86a6d9f28b6d389d12cdb341237 (tree506739e213841837c07987fbc002a71ba0f44d0a). Only production chat.go changed. No native/toolkit/Controller/+menu/FIFO/decoder changes.

## Source and behavior
Rust299 composer.rs744–765 first collects image entries and returns without inserting accompanying text; app/events.rs1108–1125 drains images in order to attachment capacity. Previous Go OnPaste only dispatched PasteImage for empty text, so PNG+Text replaced the selected draft with text and attached nothing.

Current candidate captures the displayed attachment owner before building input. Old owner paste is consumed without dispatch or default text insertion. Current owner text-empty preserves the old PasteImage route. Current owner text-nonempty with canonicalPNG advertised also dispatches PasteImage and consumes. All other text remains the native editor's original replace-selection/undo path. No image representation is read by this new metadata check; existing Controller.ReadImage/AddImage admits the image through unchanged FIFO/pending/cap.

## Native formats and honest limits
Pinned MyGo9c69cf: Darwin clipboard.go153–179 mapsTIFF to canonicalPNG; original system.go46–61 readsPNG/TIFF. Linux clipboard.go146–147 advertisesPNG forimage/* and existing GTK ReadImage performs conversion. Windows clipboard.go100–101 mapsDIB/DIBV5 toPNG; original ReadImage keepsPNG/DIB conversion. This does not promise JPEG-only or generalNSImage conversion onMac that the old path lacked.

Metadata discovery and ReadImage are not one atomic snapshot. Clipboard changes/busy metadata remains an existing limitation. Formats cannot expose typed error: unavailable metadata+nonempty text keeps normal text fallback, not silent consume. Emptytext keeps the legacyPasteImage error route. The candidate does not fetch all representations or materialize all items. No total native allocation bound is claimed.

Multiple clipboard images remain a separate known source gap: globalReadImage yields one, whileRust iteratesall. The original two-image failing fixture is archived in composer-clipboard-original-red-1216.go.txt, not converted to a false parityPASS. A strict cap-before-native-read multiitem snapshot would need anotherAPI design; post-read truncation is not that guarantee. +menuPasteImage staysunchanged.

## Evidence
Actual mounted CmdV with controlled2×2PNG+text, sameitem and separateitems, produced originalRED: selected乙 replaced byCLIP in甲😀CLIP尾, zeroPasteImage/zeroattachments, changedselection/undo. Imageonly actualAddImagepipeline and textonlyreplace/Undo controls passed. Testbackend replaces only globalnativeReadImage with Tester's controlled firstPNG; actual OnPaste/Dispatch/pipeline remain. No systemclipboard/userdata accessed.

Final9top/19inclsub each3run/3pass, zeroskip/fail,23.204s; UIvet and DarwinARM64 explicitEXIT0. The metadata test uses a lazyPNGprovider and asserts zero read calls in the new probe. Stale committed input→actualSetSnapshot session restoration preserves its saved新草稿; runtime/generation/closed retain existingdraft and reject oldimage dispatch/text insertion. Existing FIFO safety/pending-submit/IME tests unchanged.

Initial stalefixture directly assigned a differenta.Draft without a frame. Same-session native editor wroteoldbuffer back on an ordinaryLeftkey in exactbaseline too; baselinePaste additionally insertedclipboardtext, candidatePaste doesnot. This independent programmatic-buffer synchronization residual is NOT fixed or hidden here. Original tests/logs retained in the supersededfixture and baselineprobe .go.txt archives. The finalguard only claims rejection of oldpaste intent, not generic editor synchronization.

Fullsuite/hostedOS clipboard/physicalIME NOT_RUN by this worker. Lead aggregate and independent review required beforepublication.
