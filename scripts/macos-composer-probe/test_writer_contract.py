import base64
import hashlib
import io
import pathlib
import re
import struct
import unittest
from unittest.mock import Mock
import zlib

from writer_contract import (ENCODED, PIXEL_SHA256, RGBA, TEXT, cleanup_decision,
                             expected_types, request_cleanup, validate_written)

TOKEN = "12345678-1234-1234-1234-123456789abc" * 2


def record(mode="png-text"):
    fixture = {"textSHA256": hashlib.sha256(TEXT.encode()).hexdigest()}
    if mode != "text-only":
        size, sha = ENCODED[mode]
        fixture.update(width=2, height=2, encodedBytes=size, encodedSHA256=sha,
                       rgbaSHA256=PIXEL_SHA256, rgbaBase64=base64.b64encode(RGBA).decode())
    return dict(schema=1, status="WRITTEN", pid=42, mode=mode, changeCount=10,
                types=expected_types(mode), fixture=fixture, cleanupToken=TOKEN,
                receipt="/private/tmp/orbit-composer-writer.Ab123Z/receipt.json")


class WriterContractTests(unittest.TestCase):
    def test_all_modes(self):
        for mode in ("png-text", "tiff-text", "text-only"):
            validate_written(record(mode), mode=mode, child_pid=42)

    def test_reject_unknown_modes(self):
        for mode in ("png", "tiff", "cleanup", "file:/tmp/image.png", "", "PNG-text"):
            with self.assertRaises(ValueError):
                expected_types(mode)

    def test_tiff_never_png(self):
        r = record("tiff-text")
        r["types"].append("public.png")
        with self.assertRaises(ValueError):
            validate_written(r, mode="tiff-text", child_pid=42)

    def test_reject_metadata_tampering(self):
        for key, value in (("schema", True), ("pid", 41), ("pid", True), ("mode", "text-only"),
                           ("changeCount", True), ("changeCount", -1),
                           ("types", []), ("cleanupToken", "guessed"),
                           ("receipt", "/tmp/arbitrary"),
                           ("receipt", "/private/tmp/orbit-composer-writer.../receipt.json"),
                           ("status", "CLEARED")):
            with self.subTest(key=key, value=value), self.assertRaises(ValueError):
                r = record(); r[key] = value
                validate_written(r, mode="png-text", child_pid=42)

    def test_negative_has_no_image(self):
        r = record("text-only"); r["fixture"]["width"] = 2
        with self.assertRaises(ValueError):
            validate_written(r, mode="text-only", child_pid=42)

    def test_wrong_raster_rejected(self):
        for field, value in (("width", 3), ("height", 1), ("encodedSHA256", "0"*64),
                             ("rgbaSHA256", "0"*64), ("rgbaBase64", "")):
            r = record(); r["fixture"][field] = value
            with self.assertRaises(ValueError):
                validate_written(r, mode="png-text", child_pid=42)

    def test_cleanup_owned_generation_only(self):
        self.assertEqual(cleanup_decision(live_token=TOKEN, owned_generation=10,
            command="cleanup " + TOKEN, current_generation=10), "clear")
        for generation in (9, 11, True):
            self.assertEqual(cleanup_decision(live_token=TOKEN, owned_generation=10,
                command="cleanup " + TOKEN, current_generation=generation), "generation-changed")

    def test_no_cleanup_with_forged_generation_path_or_token(self):
        for cmd in ("cleanup guessed", "cleanup " + TOKEN + " 10", "cleanup /tmp/receipt.json", "x"*129, ""):
            self.assertEqual(cleanup_decision(live_token=TOKEN, owned_generation=10,
                command=cmd, current_generation=10), "invalid-command-or-token")

    def test_mock_process_cleanup(self):
        process = Mock(pid=42, stdin=io.BytesIO())
        request_cleanup(process, record())
        self.assertEqual(process.stdin.getvalue(), ("cleanup " + TOKEN + "\n").encode())
        process = Mock(pid=43, stdin=io.BytesIO())
        with self.assertRaises(ValueError): request_cleanup(process, record())
        self.assertEqual(process.stdin.getvalue(), b"")

    def test_swift_embedded_encoded_hashes_and_pixels(self):
        source = pathlib.Path(__file__).with_name("clipboard_writer.swift").read_text()
        fixtures = {}
        for kind in ("png", "tiff"):
            blob = base64.b64decode(re.search(r'let ' + kind + r'Base64 = "([^"]+)"', source)[1], validate=True)
            self.assertEqual((len(blob), hashlib.sha256(blob).hexdigest()), ENCODED[kind + "-text"])
            fixtures[kind] = blob
        png = fixtures["png"]
        self.assertEqual(struct.unpack(">II", png[16:24]), (2, 2))
        offset = 8; compressed = b""
        while offset < len(png):
            size = struct.unpack(">I", png[offset:offset+4])[0]
            kind = png[offset+4:offset+8]; data = png[offset+8:offset+8+size]
            self.assertEqual(zlib.crc32(kind + data), struct.unpack(">I", png[offset+8+size:offset+12+size])[0])
            if kind == b"IDAT": compressed += data
            offset += 12 + size
        raw = zlib.decompress(compressed)
        self.assertEqual(raw, b"\0" + RGBA[:8] + b"\0" + RGBA[8:])
        tiff = fixtures["tiff"]
        self.assertEqual(tiff[:8], b"II*\0\x08\0\0\0")
        tags = {struct.unpack_from("<HHII", tiff, 10 + 12*i)[0]: struct.unpack_from("<HHII", tiff, 10 + 12*i)[1:] for i in range(10)}
        self.assertEqual(tags[256], (4, 1, 2)); self.assertEqual(tags[257], (4, 1, 2))
        self.assertEqual(tags[259], (3, 1, 1)); self.assertEqual(tags[262], (3, 1, 2))
        rgb = tiff[tags[273][2]:tags[273][2]+tags[279][2]]
        self.assertEqual(b"".join(rgb[i:i+3]+b"\xff" for i in range(0, len(rgb), 3)), RGBA)

    def test_swift_bounded_surface(self):
        source = pathlib.Path(__file__).with_name("clipboard_writer.swift").read_text()
        self.assertIn("O_EXCL | O_NOFOLLOW", source)
        self.assertIn("mode_t(0o600)", source)
        self.assertIn("board.changeCount == ownedGeneration", source)
        self.assertIn("command.count <= 128", source)
        self.assertLess(source.index("board.writeObjects([item])"), source.index("board.types"))
        self.assertEqual(source.count("board.types"), 1)
        for forbidden in ("data(forType:", "string(forType:", "readObjects(", "setDataProvider", "CommandLine.arguments[2]"):
            self.assertNotIn(forbidden, source)


if __name__ == "__main__":
    unittest.main()
