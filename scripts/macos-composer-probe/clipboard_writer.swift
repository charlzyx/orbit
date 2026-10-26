// TEST ONLY. Dedicated disposable macOS runner; never a user desktop.
// No ambient clipboard data is read. Cleanup authority remains in this process.
import AppKit
import CryptoKit
import Darwin
import Foundation

let marker = "clipboard-text"
let pngBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAEklEQVR42mP4z8DwHwyBNBgAAEnICfcD2WTxAAAAAElFTkSuQmCC"
let tiffBase64 = "SUkqAAgAAAAKAAABBAABAAAAAgAAAAEBBAABAAAAAgAAAAIBAwADAAAAhgAAAAMBAwABAAAAAQAAAAYBAwABAAAAAgAAABEBBAABAAAAjAAAABUBAwABAAAAAwAAABYBBAABAAAAAgAAABcBBAABAAAADAAAABwBAwABAAAAAQAAAAAAAAAIAAgACAD/AAAA/wAAAP////8="
let rgba = Data(base64Encoded: "/wAA/wD/AP8AAP///////w==")!
func digest(_ data: Data) -> String { SHA256.hash(data: data).map { String(format: "%02x", $0) }.joined() }
func emit(_ record: [String: Any]) {
    let data = try! JSONSerialization.data(withJSONObject: record, options: [.sortedKeys])
    FileHandle.standardOutput.write(data)
    FileHandle.standardOutput.write(Data([10]))
}
func fail(_ reason: String) -> Never { emit(["status": "BLOCKED", "reason": reason]); exit(2) }
let modes = ["png-text", "tiff-text", "text-only"]
guard CommandLine.arguments.count == 2, modes.contains(CommandLine.arguments[1]) else { fail("mode-not-allowed") }
guard ProcessInfo.processInfo.environment["ORBIT_COMPOSER_DISPOSABLE_RUNNER"] == "1" else { fail("disposable-runner-opt-in-required") }
let mode = CommandLine.arguments[1]
let token = UUID().uuidString.lowercased() + UUID().uuidString.lowercased()
// Fixed /private/tmp parent, no user-specified paths. mkdtemp creates mode 0700.
var template = Array("/private/tmp/orbit-composer-writer.XXXXXX".utf8CString)
guard let directory = template.withUnsafeMutableBufferPointer({ p -> String? in
    guard let created = mkdtemp(p.baseAddress!) else { return nil }; return String(cString: created)
}) else { fail("receipt-directory-create") }
let receipt = directory + "/receipt.json"
let fd = open(receipt, O_WRONLY | O_CREAT | O_EXCL | O_NOFOLLOW, mode_t(0o600))
guard fd >= 0 else { rmdir(directory); fail("receipt-exclusive-create") }
defer { close(fd); unlink(receipt); rmdir(directory) }
// Reserve receipt before touching the pasteboard. File is an audit artifact, NOT authority.
let item = NSPasteboardItem()
guard item.setString(marker, forType: .string) else { fail("set-text") }
var fixtures: [String: Any] = ["textSHA256": digest(Data(marker.utf8))]
if mode != "text-only" {
    let bytes = Data(base64Encoded: mode == "png-text" ? pngBase64 : tiffBase64)!
    let kind: NSPasteboard.PasteboardType = mode == "png-text" ? .png : .tiff
    guard item.setData(bytes, forType: kind) else { fail("set-image") }
    fixtures["encodedSHA256"] = digest(bytes)
    fixtures["encodedBytes"] = bytes.count
    fixtures["width"] = 2; fixtures["height"] = 2
    fixtures["rgbaSHA256"] = digest(rgba)
    fixtures["rgbaBase64"] = rgba.base64EncodedString()
}
let expectedTypes = ([NSPasteboard.PasteboardType.string.rawValue] + (mode == "text-only" ? [] : [mode == "png-text" ? NSPasteboard.PasteboardType.png.rawValue : NSPasteboard.PasteboardType.tiff.rawValue])).sorted()
guard item.types.map({ $0.rawValue }).sorted() == expectedTypes else { fail("item-types-mismatch") }
let board = NSPasteboard.general
let clearedGeneration = board.clearContents()
guard board.changeCount == clearedGeneration else { fail("publication-raced-before-write") }
guard board.writeObjects([item]) else { fail("write-failed") }
let ownedGeneration = board.changeCount
// Metadata only, checked against the just-written generation on both sides.
guard ownedGeneration > clearedGeneration else { fail("generation-did-not-advance") }
let actualTypes = (board.types ?? []).map { $0.rawValue }.sorted()
guard board.changeCount == ownedGeneration, actualTypes == expectedTypes else { fail("publication-raced-or-types-mismatch") }
let record: [String: Any] = ["schema": 1, "status": "WRITTEN", "pid": getpid(), "mode": mode,
    "changeCount": ownedGeneration, "types": actualTypes, "fixture": fixtures,
    "cleanupToken": token, "receipt": receipt]
let receiptData = try! JSONSerialization.data(withJSONObject: record, options: [.sortedKeys])
let receiptOK = receiptData.withUnsafeBytes { p in write(fd, p.baseAddress!, p.count) == p.count }
guard receiptOK, fsync(fd) == 0 else { fail("receipt-write-failed") }
emit(record)
// Exactly one bounded ASCII command. No caller generation, receipt path or arbitrary content.
var command = [UInt8]()
while command.count <= 128 {
    var byte: UInt8 = 0
    let count = read(STDIN_FILENO, &byte, 1)
    if count == 0 { emit(["status": "CLEANUP_REFUSED", "reason": "stdin-closed"]); exit(3) }
    if count < 0 { fail("stdin-read") }
    if byte == 10 { break }
    command.append(byte)
}
guard command.count <= 128, String(bytes: command, encoding: .utf8) == "cleanup " + token else {
    emit(["status": "CLEANUP_REFUSED", "reason": "invalid-command-or-token"]); exit(3)
}
guard board.changeCount == ownedGeneration else {
    emit(["status": "CLEANUP_REFUSED", "reason": "generation-changed"]); exit(3)
}
// AppKit has no compare-and-clear operation: a concurrent writer between this check
// and clearContents is not preventable. Admission requires a dedicated runner.
let after = board.clearContents()
emit(["status": "CLEARED", "pid": getpid(), "ownedChangeCount": ownedGeneration, "changeCount": after])
