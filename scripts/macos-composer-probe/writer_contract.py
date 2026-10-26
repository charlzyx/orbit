"""Offline contract/oracle; does not launch the native helper or access clipboard.

A future runner must bound stdout reads and deadlines itself; validate one JSON
record before input. This module is not evidence of Swift compilation/execution.
"""
import hashlib
import re

MODES = frozenset(("png-text", "tiff-text", "text-only"))
TEXT = "clipboard-text"
RGBA = bytes.fromhex("ff0000ff00ff00ff0000ffffffffffff")
PIXEL_SHA256 = hashlib.sha256(RGBA).hexdigest()
ENCODED = {
    "png-text": (75, "9ccfc2abaa3984dc34c93aee16be0afa8a5e1395f25492b3df67897e6d00df10"),
    "tiff-text": (152, "d5d3c0c0582c73ceed866caea23fc528dcbdf25ddf732a097d589179dd2a2e91"),
}
TOKEN = re.compile(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\Z")
RECEIPT = re.compile(r"/private/tmp/orbit-composer-writer\.[A-Za-z0-9]{6}/receipt\.json\Z")


def expected_types(mode):
    if mode not in MODES:
        raise ValueError("mode-not-allowed")
    return sorted(["public.utf8-plain-text"] + ({"png-text": ["public.png"], "tiff-text": ["public.tiff"]}.get(mode, [])))


def validate_written(record, *, mode, child_pid):
    """Fail closed; child_pid comes from the spawned process, not its JSON."""
    types = expected_types(mode)
    if type(child_pid) is not int or child_pid <= 0:
        raise ValueError("invalid-child-pid")
    if type(record.get("schema")) is not int or record["schema"] != 1 or record.get("status") != "WRITTEN":
        raise ValueError("not-written")
    if type(record.get("pid")) is not int or record["pid"] != child_pid or record.get("mode") != mode:
        raise ValueError("wrong-process-or-mode")
    if type(record.get("changeCount")) is not int or record["changeCount"] < 1:
        raise ValueError("invalid-generation")
    if record.get("types") != types:
        raise ValueError("wrong-types")
    if not isinstance(record.get("cleanupToken"), str) or not TOKEN.fullmatch(record["cleanupToken"]):
        raise ValueError("invalid-token")
    if not isinstance(record.get("receipt"), str) or not RECEIPT.fullmatch(record["receipt"]):
        raise ValueError("invalid-receipt-path")
    fixture = record.get("fixture")
    if not isinstance(fixture, dict) or fixture.get("textSHA256") != hashlib.sha256(TEXT.encode()).hexdigest():
        raise ValueError("wrong-text")
    if mode == "text-only":
        if set(fixture) != {"textSHA256"}:
            raise ValueError("image-in-negative-fixture")
    else:
        size, sha = ENCODED[mode]
        if (fixture.get("width"), fixture.get("height"), fixture.get("encodedBytes"), fixture.get("encodedSHA256"), fixture.get("rgbaSHA256"), fixture.get("rgbaBase64")) != (2, 2, size, sha, PIXEL_SHA256, "/wAA/wD/AP8AAP///////w=="):
            raise ValueError("wrong-raster")
    return record


def cleanup_decision(*, live_token, owned_generation, command, current_generation):
    """Model of helper's in-memory policy, not a filesystem authorization API."""
    if not TOKEN.fullmatch(live_token) or len(command) > 128 or command != "cleanup " + live_token:
        return "invalid-command-or-token"
    if type(owned_generation) is not int or owned_generation < 1 or type(current_generation) is not int or owned_generation != current_generation:
        return "generation-changed"
    return "clear"


def request_cleanup(process, record):
    """Only forwards the live writer's token; no paths/generation supplied.

    Caller must first validate_written with process.pid and retain that exact
    record/process pair. Tests use a mock only; this function does not spawn.
    """
    validate_written(record, mode=record["mode"], child_pid=process.pid)
    process.stdin.write(("cleanup " + record["cleanupToken"] + "\n").encode("ascii"))
    process.stdin.flush()
