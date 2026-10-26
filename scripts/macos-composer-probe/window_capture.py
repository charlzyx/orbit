"""OS-window capture candidate. Local validation is NOT_RUN_NATIVE.

No permission request, screen capture fallback, window enumeration, or content
capture fallback. A future phase handshake must hold the owned window stable and
supply fresh before/after own-NSWindow receipts. Receipt callbacks must honor the
provided monotonic deadline. Native admission and independent review are still
required; a returned artifact is not a four-case acceptance verdict.

CGWindow bounds use top-left coordinates; NSWindow frames use bottom-left.
Compare each source before/after exactly, but cross-source compare ONLY w/h.
Matching a receipt's pid/start/window with independently collected window-proof
is mandatory; neither a report nor an arbitrary window number proves ownership.
"""
import copy
import hashlib
import io
import json
import math
import os
import platform
import re
import selectors
import stat
import struct
import subprocess
import time
import zlib
import warnings
from dataclasses import dataclass

PHASES = ("before", "seed", "final", "undo")
MAX_JSON = 16384
MAX_PNG = 16 * 1024 * 1024


class Blocked(ValueError):
    pass


@dataclass(frozen=True)
class Identity:
    nonce: str
    pid: int
    start_us: int
    executable: str
    window_number: int


def _positive_int(value):
    return type(value) is int and value > 0


def validate_identity(identity):
    if not isinstance(identity, Identity) or not isinstance(identity.nonce, str) or not re.fullmatch(r"[0-9a-f]{64}", identity.nonce):
        raise Blocked("identity-nonce")
    if not all(_positive_int(value) for value in (identity.pid, identity.start_us, identity.window_number)):
        raise Blocked("identity-process-or-window")
    if not isinstance(identity.executable, str) or not identity.executable.startswith("/") or "\x00" in identity.executable or any(part in ("", ".", "..") for part in identity.executable.split("/")[1:]):
        raise Blocked("identity-executable")


def bounds(value):
    if not isinstance(value, dict) or set(value) != {"x", "y", "w", "h"}:
        raise Blocked("bounds-schema")
    if any(type(v) not in (int, float) or not math.isfinite(v) for v in value.values()):
        raise Blocked("bounds-not-finite")
    if not 0 < value["w"] <= 4096 or not 0 < value["h"] <= 4096:
        raise Blocked("bounds-size")
    return value


def validate_receipt(receipt, identity, phase, now_ns, *, earliest_ns=None):
    validate_identity(identity)
    if phase not in PHASES or not isinstance(receipt, dict): raise Blocked("receipt-phase")
    if any(type(receipt.get(key)) is not int for key in ("pid", "start_us", "window_number")) or (receipt.get("nonce"), receipt.get("pid"), receipt.get("start_us"), receipt.get("window_number"), receipt.get("phase")) != (identity.nonce, identity.pid, identity.start_us, identity.window_number, phase):
        raise Blocked("receipt-owner-mismatch")
    if receipt.get("visible") is not True: raise Blocked("receipt-not-visible")
    stamp = receipt.get("at_unix_ns")
    if not _positive_int(stamp) or not max(identity.start_us * 1000, now_ns-5_000_000_000) <= stamp <= now_ns or (earliest_ns is not None and stamp <= earliest_ns):
        raise Blocked("receipt-stale-or-out-of-order")
    bounds(receipt.get("bounds"))
    return receipt


def validate_proof(proof, identity):
    validate_identity(identity)
    if not isinstance(proof, dict) or any(type(proof.get(key)) is not int for key in ("pid", "start_us", "window_number")) or (proof.get("pid"), proof.get("start_us"), proof.get("executable"), proof.get("window_number")) != (identity.pid, identity.start_us, identity.executable, identity.window_number):
        raise Blocked("independent-window-owner-mismatch")
    if proof.get("onscreen") is not True or proof.get("preflight") is not True:
        raise Blocked("window-not-onscreen-or-permission-unavailable")
    bounds(proof.get("bounds"))
    return proof


def compare_pair(receipt, proof):
    if (receipt["bounds"]["w"], receipt["bounds"]["h"]) != (proof["bounds"]["w"], proof["bounds"]["h"]):
        raise Blocked("cross-source-window-size-mismatch")


def validate_continuity(before_receipt, after_receipt, before_proof, after_proof):
    compare_pair(before_receipt, before_proof); compare_pair(after_receipt, after_proof)
    if before_receipt["bounds"] != after_receipt["bounds"] or before_proof["bounds"] != after_proof["bounds"]:
        raise Blocked("window-moved-or-resized")


def _remaining(deadline):
    remaining = deadline-time.monotonic()
    if not math.isfinite(remaining) or remaining <= 0: raise Blocked("capture-deadline")
    return remaining


def _local_executable(path):
    if not isinstance(path, str) or not path.startswith("/") or os.path.realpath(path) != path:
        raise Blocked("verifier-path")
    st = os.stat(path, follow_symlinks=False)
    if not stat.S_ISREG(st.st_mode) or st.st_uid != os.getuid() or st.st_mode & 0o022 or not os.access(path, os.X_OK):
        raise Blocked("unsafe-verifier")
    return path


def private_output(identity, phase):
    validate_identity(identity)
    if phase not in PHASES: raise Blocked("capture-phase")
    directory = "/private/tmp/orbit-composer-" + identity.nonce
    if os.path.realpath(directory) != directory: raise Blocked("symlink-output-directory")
    st = os.stat(directory, follow_symlinks=False)
    if not stat.S_ISDIR(st.st_mode) or st.st_uid != os.getuid() or stat.S_IMODE(st.st_mode) != 0o700:
        raise Blocked("unsafe-output-directory")
    path = directory + "/" + phase + "-os.png"
    if os.path.lexists(path): raise Blocked("capture-output-already-exists")
    return path


def _json(raw):
    if not raw or len(raw) > MAX_JSON: raise Blocked("window-proof-empty-or-oversized")
    def pairs(values):
        result = {}
        for key, value in values:
            if key in result: raise Blocked("window-proof-duplicate-key")
            result[key] = value
        return result
    def constant(_): raise Blocked("window-proof-nonfinite")
    try: value = json.loads(raw.decode(), object_pairs_hook=pairs, parse_constant=constant)
    except (ValueError, UnicodeError, RecursionError) as error: raise Blocked("window-proof-invalid-json") from error
    if not isinstance(value, dict): raise Blocked("window-proof-not-object")
    return value


def _window_proof(verifier, identity, deadline):
    _remaining(deadline)
    process = subprocess.Popen([verifier, "window-proof", str(identity.pid), str(identity.start_us), str(identity.window_number)], stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, bufsize=0, env={"PATH": "/usr/bin:/bin"})
    try:
        raw = bytearray()
        with selectors.DefaultSelector() as selector:
            selector.register(process.stdout, selectors.EVENT_READ)
            while True:
                if not selector.select(_remaining(deadline)): raise Blocked("window-proof-deadline")
                part = os.read(process.stdout.fileno(), min(4096, MAX_JSON+1-len(raw)))
                if not part: break
                raw.extend(part)
                if len(raw) > MAX_JSON: raise Blocked("window-proof-oversized")
        if process.wait(timeout=_remaining(deadline)) != 0: raise Blocked("window-proof-command-failed")
        return validate_proof(_json(bytes(raw)), identity)
    finally:
        try:
            if process.poll() is None:
                try: process.terminate()  # only this direct Popen child
                except ProcessLookupError: pass
            try: process.wait(timeout=2)
            except subprocess.TimeoutExpired as error:
                # No SIGKILL or signal of a report-supplied/unknown PID.
                raise Blocked(f"window-proof-child-not-reaped pid={process.pid}") from error
        finally:
            process.stdout.close()



# Legal PNG bit-depth/color combinations. Adam7 is deliberately unsupported:
# this guard counts noninterlaced scanline bytes, never implements pixel filters.
PNG_DEPTHS = {0: (1, 2, 4, 8, 16), 2: (8, 16), 3: (1, 2, 4, 8), 4: (8, 16), 6: (8, 16)}
PNG_CHANNELS = {0: 1, 2: 3, 3: 1, 4: 2, 6: 4}
MAX_INFLATED = 128 * 1024 * 1024 + 4096


def _validate_idat_length(parts, width, height, header):
    depth, color, compression, filtering, interlace = header
    if interlace != 0: raise Blocked("capture-adam7-unsupported")
    if depth not in PNG_DEPTHS.get(color, ()) or compression != 0 or filtering != 0:
        raise Blocked("capture-png-header-unsupported")
    expected = height * (1 + (width * PNG_CHANNELS[color] * depth + 7) // 8)
    if width * height > 4096 * 4096 or expected > MAX_INFLATED:
        raise Blocked("capture-inflate-budget")
    decoder = zlib.decompressobj(); total = 0
    try:
        for part in parts:
            pending = part
            while pending:
                # At most 64 KiB transient output, one extra byte detects overflow.
                decoded = decoder.decompress(pending, min(65536, expected-total+1))
                total += len(decoded)
                if total > expected: raise Blocked("capture-scanline-length-excess")
                if decoder.unused_data: raise Blocked("capture-zlib-trailing-data")
                pending = decoder.unconsumed_tail
        if not decoder.eof or decoder.unused_data or decoder.unconsumed_tail or total != expected:
            raise Blocked("capture-zlib-incomplete-or-scanline-length")
    except zlib.error as error:
        raise Blocked("capture-zlib-invalid") from error

def _verify_png_decode(raw, width, height):
    # Mature decoder, read-only. Capture corroborates appearance; never state truth.
    # Both dimensions and total decoded-pixel budget are checked BEFORE decoding.
    if width * height > 4096 * 4096: raise Blocked("capture-pixel-budget")
    try:
        import PIL
        from PIL import Image, ImageFile
    except ImportError as error:
        raise Blocked("capture-decoder-unavailable: Pillow required") from error
    if ImageFile.LOAD_TRUNCATED_IMAGES:
        raise Blocked("capture-decoder-truncated-mode-enabled")
    try:
        with warnings.catch_warnings():
            warnings.simplefilter("error", Image.DecompressionBombWarning)
            with Image.open(io.BytesIO(raw)) as image:
                if image.format != "PNG" or image.size != (width, height) or getattr(image, "n_frames", 1) != 1:
                    raise Blocked("capture-decoder-format-size-or-animation")
                image.verify()
            # verify() intentionally does not decode pixels; reopen and load them.
            with Image.open(io.BytesIO(raw)) as image:
                if image.format != "PNG" or image.size != (width, height):
                    raise Blocked("capture-decoder-reopen-mismatch")
                image.load()
    except (OSError, ValueError, SyntaxError, Image.DecompressionBombError, Image.DecompressionBombWarning) as error:
        raise Blocked("capture-png-decode-failed: " + str(error)) from error
    return "Pillow " + PIL.__version__


def _png(path):
    fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
    try:
        st = os.fstat(fd)
        if not stat.S_ISREG(st.st_mode) or st.st_uid != os.getuid() or not 33 <= st.st_size <= MAX_PNG: raise Blocked("capture-empty-or-unsafe")
        with os.fdopen(fd, "rb", closefd=False) as stream: raw = stream.read(MAX_PNG+1)
        after = os.fstat(fd)
        if (st.st_size, st.st_mtime_ns, st.st_ctime_ns) != (after.st_size, after.st_mtime_ns, after.st_ctime_ns) or len(raw) != st.st_size: raise Blocked("capture-file-changed")
        if raw[:16] != b"\x89PNG\r\n\x1a\n\0\0\0\rIHDR": raise Blocked("capture-not-png")
        width, height = struct.unpack(">II", raw[16:24])
        if not 1 <= width <= 4096 or not 1 <= height <= 4096: raise Blocked("capture-png-size")
        # Bounded structural integrity only; no decompression, transforms or OCR.
        # Same complete-chunk/CRC/nonempty-IDAT/terminal-IEND policy as runner.
        offset = 8; chunks = 0; image_data = False; ended = False; idat_parts = []
        while offset < len(raw):
            chunks += 1
            if chunks > 8192 or offset + 12 > len(raw):
                raise Blocked("capture-png-chunk-bound")
            length = struct.unpack(">I", raw[offset:offset+4])[0]
            end = offset + 12 + length
            if end > len(raw): raise Blocked("capture-png-truncated-chunk")
            kind = raw[offset+4:offset+8]
            if zlib.crc32(raw[offset+4:end-4]) != struct.unpack(">I", raw[end-4:end])[0]:
                raise Blocked("capture-png-chunk-crc")
            if kind == b"IDAT":
                idat_parts.append(raw[offset+8:end-4])
                if length > 0: image_data = True
            if kind == b"IEND":
                if length != 0 or end != len(raw): raise Blocked("capture-png-bad-end")
                ended = True
            offset = end
        if not image_data or not ended: raise Blocked("capture-png-incomplete")
        _validate_idat_length(idat_parts, width, height, raw[24:29])
        decoder = _verify_png_decode(raw, width, height)
        return dict(sha256=hashlib.sha256(raw).hexdigest(), bytes=len(raw), width=width, height=height, decoder=decoder)
    finally: os.close(fd)


def capture_owned_window(*, identity, phase, verifier, own_receipt, deadline, execute_native=False):
    """Future bounded operation; own_receipt(stage, phase, deadline) must be bounded.

    Explicit executable is the reviewed Go window-proof tool. No alternate
    screencapture executable/options are accepted. Only the reviewed fixed phase handshake calls this candidate.
    """
    if execute_native is not True or platform.system() != "Darwin" or platform.machine() != "arm64" or os.environ.get("ORBIT_COMPOSER_DISPOSABLE_RUNNER") != "1":
        raise Blocked("native-capture-not-admitted")
    try:
        validate_identity(identity); _remaining(deadline)
        verifier = _local_executable(verifier)
        path = private_output(identity, phase)
        before = copy.deepcopy(own_receipt("before", phase, deadline))
        validate_receipt(before, identity, phase, time.time_ns())
        proof_before = copy.deepcopy(_window_proof(verifier, identity, deadline))
        compare_pair(before, proof_before)
        validate_receipt(before, identity, phase, time.time_ns())
        if private_output(identity, phase) != path: raise Blocked("output-path-changed")
        started_ns = time.time_ns()
        result = subprocess.run(["/usr/sbin/screencapture", "-x", "-l", str(identity.window_number), path], stdin=subprocess.DEVNULL, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=_remaining(deadline), check=False, env={"PATH": "/usr/bin:/bin"})
        if result.returncode != 0: raise Blocked("screencapture-command-failed")
        captured_ns = time.time_ns()
        proof_after = copy.deepcopy(_window_proof(verifier, identity, deadline))
        after = copy.deepcopy(own_receipt("after", phase, deadline))
        validate_receipt(after, identity, phase, time.time_ns(), earliest_ns=captured_ns)
        _remaining(deadline)
        validate_continuity(before, after, proof_before, proof_after)
        artifact = _png(path)
        _remaining(deadline)
        return dict(status="OS_WINDOW_CAPTURE_CANDIDATE", integration="phase-handshake-candidate", phase=phase, path=path, at_unix_ns=started_ns, before_receipt=before, after_receipt=after, before_proof=proof_before, after_proof=proof_after, **artifact)
    except (OSError, subprocess.SubprocessError) as error:
        raise Blocked("native-capture-operation-failed: " + str(error)) from error
