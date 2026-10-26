#!/usr/bin/env python3
"""Offline admission contracts for a FUTURE four-case macOS composer runner.

Default invocation emits BLOCKED and performs no native actions. Native execution
requires --execute-native, an explicit local app/writer/process-proof executable,
Darwin arm64, and disposable-runner opt-in. Implementation has NOT_RUN_NATIVE
status; only offline mocked tests have run. App capture belongs to the overlay.
"""
import argparse
import math
import hashlib
import struct
import zlib
import os
import platform
import plistlib
import secrets
import selectors
import signal
import stat
import subprocess
import time
from dataclasses import dataclass
import json
import re

from writer_contract import PIXEL_SHA256, validate_written

CASES = ("png-text", "tiff-text", "text-only", "menu-cancel")
PRODUCTION_BASE = "ee25813f9a2065ae8632d4c15b24e3c99e7bf469"
PRODUCTION_TREE = "0476116687acf0838d52a0246210270000530530"
MAX_WRITER_LINE = 16 * 1024
MAX_REPORT = 128 * 1024
NONCE = re.compile(r"[0-9a-f]{64}\Z")


class Blocked(ValueError):
    pass


def strict_int(value, name, minimum=1):
    if type(value) is not int or value < minimum:
        raise Blocked("invalid-" + name)
    return value


def canonical_absolute(path):
    if not isinstance(path, str) or not path.startswith("/") or "\x00" in path:
        raise Blocked("nonabsolute-path")
    if any(part in (".", "..", "") for part in path.split("/")[1:]):
        raise Blocked("noncanonical-path")
    return path


def unique_json(raw, max_bytes):
    if not isinstance(raw, bytes) or len(raw) > max_bytes:
        raise Blocked("oversized-or-nonbytes-json")
    def unique_pairs(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise Blocked("duplicate-json-key")
            result[key] = value
        return result
    def reject_constant(_):
        raise Blocked("nonfinite-json-number")
    try:
        obj = json.loads(raw.decode("utf-8"), object_pairs_hook=unique_pairs,
                         parse_constant=reject_constant)
    except (UnicodeError, json.JSONDecodeError, RecursionError) as error:
        raise Blocked("invalid-json") from error
    if not isinstance(obj, dict):
        raise Blocked("json-not-object")
    return obj


def validate_writer_line(raw, *, mode, child_pid):
    """Input is already deadline-bounded by a FUTURE runner (not implemented)."""
    if not isinstance(raw, bytes) or not raw.endswith(b"\n") or raw.count(b"\n") != 1:
        raise Blocked("writer-not-one-complete-line")
    if mode not in CASES[:3]:
        raise Blocked("clipboard-not-permitted-in-case")
    value = unique_json(raw, MAX_WRITER_LINE)
    try:
        return validate_written(value, mode=mode, child_pid=child_pid)
    except (ValueError, KeyError, TypeError) as error:
        raise Blocked("invalid-writer-metadata") from error


@dataclass(frozen=True)
class ProcessObservation:
    """Expected libproc output. Untrusted synthetic fixtures in offline tests.

    Native observation obtains path via proc_pidpath and start time via
    x/sys/unix.SysctlKinfoProc, not report JSON, ps text or a PID alone.
    Constructing this object is NOT evidence that any observation happened.
    """
    pid: int
    start_us: int
    executable: str


@dataclass(frozen=True)
class CaseAdmission:
    mode: str
    nonce: str
    executable: str
    launch_before_us: int
    launch_after_us: int


def validate_process(observed, admission):
    """Pure comparison only; caller cannot use this as an observation API."""
    if not isinstance(observed, ProcessObservation) or not isinstance(admission, CaseAdmission):
        raise Blocked("missing-independent-process-observation")
    if admission.mode not in CASES:
        raise Blocked("unknown-case")
    if not isinstance(admission.nonce, str) or not NONCE.fullmatch(admission.nonce):
        raise Blocked("invalid-case-nonce")
    strict_int(observed.pid, "pid")
    strict_int(observed.start_us, "process-start")
    strict_int(admission.launch_before_us, "launch-before")
    strict_int(admission.launch_after_us, "launch-after")
    if admission.launch_after_us < admission.launch_before_us:
        raise Blocked("invalid-launch-window")
    if not admission.launch_before_us <= observed.start_us <= admission.launch_after_us:
        raise Blocked("process-not-fresh")
    if canonical_absolute(observed.executable) != canonical_absolute(admission.executable):
        raise Blocked("wrong-executable")
    if ".app/Contents/MacOS/" not in observed.executable:
        raise Blocked("not-app-executable")
    return observed


def validate_report_envelope(raw, *, admission, observed, report_mtime_us, observed_at_us):
    """Validates identity only, never returns PASS for UI/native behavior.

    Future reader must open a fresh private out_dir/report.json using O_NOFOLLOW,
    fstat its regular-file ownership, read with size cap and retain raw bytes.
    Current pure API does not read files or establish mtime authenticity.
    """
    validate_process(observed, admission)
    strict_int(report_mtime_us, "report-mtime")
    strict_int(observed_at_us, "observed-at")
    if not observed.start_us <= report_mtime_us <= observed_at_us:
        raise Blocked("report-not-fresh")
    report = unique_json(raw, MAX_REPORT)
    if (report.get("nonce"), report.get("mode"), report.get("executable")) != (admission.nonce, admission.mode, observed.executable):
        raise Blocked("report-identity-mismatch")
    if strict_int(report.get("pid"), "report-pid") != observed.pid or strict_int(report.get("start_us"), "report-start") != observed.start_us:
        raise Blocked("report-process-mismatch")
    if report.get("status") not in ("PASS", "FAIL", "BLOCKED"):
        raise Blocked("report-status-unknown")
    if not isinstance(report.get("inputlog"), list) or len(report["inputlog"]) > 16:
        raise Blocked("invalid-input-log")
    # Behavioral oracle is separate. Never treat a self-reported PASS as proof.
    return {"status": "ENVELOPE_VALIDATED_ONLY", "reported_status": report["status"],
            "native_acceptance": "NOT_IMPLEMENTED", "report": report}



def validate_native_target(event):
    if event.get("action") == "key:Escape":
        if event.get("target_source") != "matching-native-menu-begin" or "target" in event:
            raise Blocked("escape-not-owned-menu-notification")
        return
    if event.get("target_source") != "fresh-unique-AX":
        raise Blocked("missing-fresh-AX-evidence")
    target=event.get("target")
    if not isinstance(target,dict) or target.get("identity") != {k:event[k] for k in ("pid","start_us","window")}:
        raise Blocked("AX-owner-mismatch")
    at=strict_int(target.get("at_unix_ns"),"AX-time")
    if not 0 <= event["at_unix_ns"]-at <= 250_000_000:
        raise Blocked("AX-target-stale")
    label="Add attachment" if event["action"]=="click:Add attachment" else "Message"
    if target.get("label")!=label or type(target.get("matches")) is not int or target["matches"]!=1 or type(target.get("nodes")) is not int or not 1<=target["nodes"]<=2048:
        raise Blocked("AX-target-ambiguous-or-unbounded")
    def rect(value):
        if not isinstance(value,dict) or set(value)!={"x","y","w","h"} or any(type(v) not in (int,float) or not math.isfinite(v) for v in value.values()):
            raise Blocked("AX-nonfinite-rectangle")
        x,y,w,h=(value[k] for k in ("x","y","w","h"))
        if w<=0 or h<=0 or not math.isfinite(x+w) or not math.isfinite(y+h):raise Blocked("AX-invalid-rectangle")
        return x,y,w,h
    x,y,w,h=rect(target.get("content"));a,b,c,d=rect(target.get("rect"))
    if a<x or b<y or a+c>x+w or b+d>y+h:raise Blocked("AX-target-outside-content")


def validate_behavior(report, *, observed_at_us, expected_owner):
    """Pure schema-1 oracle. Match is NOT independently observed native success.

    First validate_report_envelope with an independent process observation. The
    caller must supply the independently known owner identity, not copy it from
    the report. Native orchestration separately binds this oracle to the process-proof CLI; offline fixtures are not native evidence.
    """
    if type(report.get("schema")) is not int or report["schema"] != 1:
        raise Blocked("wrong-report-schema")
    if report.get("inputlog_complete") is not True:
        raise Blocked("native-observer-finalization-incomplete")
    if report.get("status") != "PASS":
        raise Blocked("native-report-not-pass")
    mode = report.get("mode")
    if mode not in CASES:
        raise Blocked("unknown-case")
    pid = strict_int(report.get("pid"), "report-pid")
    start_us = strict_int(report.get("start_us"), "report-start")
    window = strict_int(report.get("window"), "window")
    strict_int(observed_at_us, "observed-at")
    if expected_owner is None or type(expected_owner) not in (str, int) or not expected_owner:
        raise Blocked("missing-independent-owner")
    seed_actions = ["click:Message", "key:s", "key:e", "key:e", "key:d"]
    actions = seed_actions + (["click:Add attachment", "key:Escape", "key:x"] if mode == "menu-cancel" else ["key:CmdV"] + (["key:CmdZ"] if mode == "text-only" else []))
    log = report.get("inputlog")
    if not isinstance(log, list) or len(log) != len(actions):
        raise Blocked("wrong-action-count")
    times = []
    for entry, action in zip(log, actions):
        if not isinstance(entry, dict) or entry.get("action") != action:
            raise Blocked("unexpected-action")
        if (strict_int(entry.get("pid"), "event-pid"), strict_int(entry.get("start_us"), "event-start"), strict_int(entry.get("window"), "event-window")) != (pid, start_us, window):
            raise Blocked("event-owner-mismatch")
        stamp = strict_int(entry.get("at_unix_ns"), "event-time")
        if not start_us * 1000 <= stamp <= observed_at_us * 1000 or (times and stamp <= times[-1]):
            raise Blocked("event-time-not-fresh-or-ordered")
        times.append(stamp)
    for event in log: validate_native_target(event)
    states = report.get("states")
    if isinstance(states, list) and states and isinstance(states[0], dict) and states[0].get("phase") == "before":
        before = states[0]
        stamp = strict_int(before.get("at_unix_ns"), "before-time")
        if (before.get("draft") != "" or before.get("attachments") != [] or before.get("pending") is not False
            or type(before.get("messages")) is not int or before["messages"] != 0
            or type(before.get("focused")) is not bool or type(before.get("ready")) is not bool
            or type(before.get("owner")) is not type(expected_owner) or before["owner"] != expected_owner
            or not start_us * 1000 < stamp < times[0]):
            raise Blocked("unsafe-before-state")
        states = states[1:]
    phases = ["seed", "final", "undo"] if mode == "text-only" else ["seed", "final"]
    if not isinstance(states, list) or len(states) != len(phases):
        raise Blocked("wrong-state-count")
    previous = 0
    for state, phase in zip(states, phases):
        if not isinstance(state, dict) or state.get("phase") != phase:
            raise Blocked("wrong-state-phase")
        if type(state.get("owner")) is not type(expected_owner) or state["owner"] != expected_owner:
            raise Blocked("state-owner-mismatch")
        if state.get("pending") is not False or state.get("focused") is not True or state.get("ready") is not True or type(state.get("messages")) is not int or state["messages"] != 0:
            raise Blocked("state-not-ready-or-submitted")
        stamp = strict_int(state.get("at_unix_ns"), "state-time")
        lower = times[4] if phase == "seed" else times[-1]
        upper = observed_at_us * 1000
        if phase == "seed":
            upper = times[5]
        if mode == "text-only" and phase == "final":
            lower, upper = times[5], times[6]
        if not lower < stamp < upper or stamp <= previous:
            raise Blocked("state-time-not-fresh-or-ordered")
        previous = stamp
        expected_draft = "seed"
        if phase == "final" and mode == "text-only": expected_draft += "clipboard-text"
        if phase == "final" and mode == "menu-cancel": expected_draft += "x"
        if state.get("draft") != expected_draft:
            raise Blocked("wrong-draft")
        expected_attachments = []
        if phase == "final" and mode in ("png-text", "tiff-text"):
            expected_attachments = [{"mime": "image/png", "width": 2, "height": 2, "pixel_sha256": PIXEL_SHA256}]
        if state.get("attachments") != expected_attachments:
            raise Blocked("wrong-attachments")
    if mode == "menu-cancel":
        menu = report.get("menu")
        if not isinstance(menu, dict) or set(menu) != {"begins", "ends", "actions", "escape_requests"} or any(type(value) is not int for value in menu.values()) or menu != {"begins": 1, "ends": 1, "actions": 0, "escape_requests": 1}:
            raise Blocked("menu-cancel-not-proved")
    elif "menu" in report:
        raise Blocked("unexpected-menu-evidence")
    return {"status": "OFFLINE_ORACLE_MATCH", "native_acceptance": "NOT_IMPLEMENTED"}

def same_owned_process(initial, current, admission):
    """Pure no-PID-reuse predicate. DOES NOT authorize or perform signaling."""
    validate_process(initial, admission)
    validate_process(current, admission)
    if current != initial:
        raise Blocked("process-identity-changed")
    return True


def native_process_observation(_pid):
    raise Blocked("NOT_IMPLEMENTED: verified proc_pidpath + PROC_PIDTBSDINFO binding required")


def offline_manifest():
    return {"status": "BLOCKED", "native_launch": "IMPLEMENTED_NOT_RUN",
            "native_execution": "NOT_RUN", "production_base": PRODUCTION_BASE,
            "production_tree": PRODUCTION_TREE, "cases": list(CASES),
            "required_launch_environment": ["HOME", "XDG_CONFIG_HOME", "PI_OFFLINE", "PI_SKIP_VERSION_CHECK", "ORBIT_COMPOSER_PROOF_CONFIG"],
            "pending": ["independent review", "compile Swift and process-proof on macOS", "fresh dedicated runner admission", "actual four-case native execution"],
            "note": "Default is offline only. Native implementation is unexecuted; explicit gated execution required."}


# Native orchestration is opt-in and must never run during offline unit tests.
def _regular(path, executable=False):
    path = canonical_absolute(path)
    if os.path.realpath(path) != path:
        raise Blocked("symlink-or-noncanonical-local-path")
    st = os.stat(path, follow_symlinks=False)
    if not stat.S_ISREG(st.st_mode) or st.st_uid != os.getuid() or st.st_mode & 0o022:
        raise Blocked("unsafe-local-file")
    if executable and not os.access(path, os.X_OK):
        raise Blocked("not-executable")
    return path


def _read_private_json(path, limit):
    fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
    try:
        st = os.fstat(fd)
        if not stat.S_ISREG(st.st_mode) or st.st_uid != os.getuid() or st.st_size > limit:
            raise Blocked("unsafe-or-oversized-artifact")
        raw = os.read(fd, limit + 1)
        return unique_json(raw, limit), raw, st.st_mtime_ns // 1000
    finally:
        os.close(fd)


def _publish_json(path, value):
    _publish_raw(path, json.dumps(value, sort_keys=True).encode() + b"\n")


def _publish_raw(path, raw):
    temporary = path + ".tmp"
    fd = os.open(temporary, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600)
    try:
        with os.fdopen(fd, "wb") as stream:
            stream.write(raw); stream.flush(); os.fsync(stream.fileno())
        # Never overwrite an existing approval or config artifact.
        os.link(temporary, path, follow_symlinks=False)
    finally:
        os.unlink(temporary)


def _line(process, deadline):
    raw = bytearray()
    with selectors.DefaultSelector() as selector:
        selector.register(process.stdout, selectors.EVENT_READ)
        while len(raw) < MAX_WRITER_LINE:
            remaining = deadline - time.monotonic()
            if remaining <= 0 or not selector.select(remaining):
                raise Blocked("child-output-deadline")
            byte = os.read(process.stdout.fileno(), 1)
            if not byte:
                raise Blocked("child-output-incomplete")
            raw.extend(byte)
            if byte == b"\n": return bytes(raw)
    raise Blocked("child-output-oversized")


def _inspect(verifier, pid, deadline):
    strict_int(pid, "inspection-pid")
    process = subprocess.Popen([verifier, "process-proof", str(pid)], stdin=subprocess.DEVNULL,
        stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, bufsize=0, env={"PATH": "/usr/bin:/bin"})
    try:
        raw = _line(process, min(deadline, time.monotonic()+5))
        remaining = min(5, deadline-time.monotonic())
        if remaining <= 0 or process.wait(timeout=remaining) != 0:
            raise Blocked("process-proof-failed")
        # Reject extra output, including a second JSON identity.
        os.set_blocking(process.stdout.fileno(), False)
        try:
            if os.read(process.stdout.fileno(), 1): raise Blocked("process-proof-extra-output")
        except BlockingIOError as error:
            raise Blocked("process-proof-pipe-not-closed") from error
        data = unique_json(raw, MAX_WRITER_LINE)
        if set(data) != {"pid", "start_us", "executable"}:
            raise Blocked("process-proof-schema")
        return ProcessObservation(data["pid"], data["start_us"], data["executable"]), raw
    finally:
        if process.poll() is None: process.terminate()
        process.stdout.close()


def _bundle_executable(app):
    canonical_absolute(app)
    if not app.endswith(".app") or os.path.realpath(app) != app or not os.path.isdir(app):
        raise Blocked("invalid-app-bundle")
    info_path = _regular(app + "/Contents/Info.plist")
    if os.stat(info_path).st_size > MAX_REPORT: raise Blocked("oversized-info-plist")
    with open(info_path, "rb") as stream: info = plistlib.load(stream)
    executable = info.get("CFBundleExecutable")
    if not isinstance(executable, str) or not executable or "/" in executable or executable in (".", ".."):
        raise Blocked("invalid-bundle-executable")
    return _regular(app + "/Contents/MacOS/" + executable, executable=True)


def _wait_artifact(path, deadline, launcher):
    while time.monotonic() < deadline:
        try: return _read_private_json(path, MAX_REPORT)
        except FileNotFoundError: pass
        if launcher.poll() is not None: raise Blocked("app-exited-before-artifact")
        time.sleep(min(0.05, max(0, deadline-time.monotonic())))
    raise Blocked("app-artifact-deadline")


def _terminate_verified(verifier, observed, admission):
    # There is no pidfd/atomic compare-and-signal on this path. Dedicated runner
    # isolation remains required; never signal an unverified/reused PID.
    current, _ = _inspect(verifier, observed.pid, time.monotonic()+5)
    same_owned_process(observed, current, admission)
    os.kill(observed.pid, signal.SIGTERM)



MAX_CAPTURE = 16 * 1024 * 1024


def _read_capture(path):
    fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
    try:
        before = os.fstat(fd)
        if not stat.S_ISREG(before.st_mode) or before.st_uid != os.getuid() or not 33 <= before.st_size <= MAX_CAPTURE:
            raise Blocked("unsafe-or-oversized-capture")
        with os.fdopen(fd, "rb", closefd=False) as stream:
            raw = stream.read(MAX_CAPTURE + 1)
        after = os.fstat(fd)
        if (before.st_ino, before.st_size, before.st_mtime_ns, before.st_ctime_ns) != (after.st_ino, after.st_size, after.st_mtime_ns, after.st_ctime_ns) or len(raw) != before.st_size:
            raise Blocked("capture-changed-during-read")
        return raw
    finally:
        os.close(fd)


def validate_captures(report, *, out_dir, admission, observed, observed_at_us):
    """Verify untransformed in-process content-rerender PNGs, not OS window pixels."""
    validate_process(observed, admission)
    if report.get("capture_kind")!="content-rerender":raise Blocked("capture-kind-not-declared")
    strict_int(observed_at_us, "capture-observed-at")
    if out_dir != "/private/tmp/orbit-composer-" + admission.nonce:
        raise Blocked("capture-output-directory-mismatch")
    expected = ["before", "seed", "final"] + (["undo"] if admission.mode == "text-only" else [])
    captures = report.get("captures")
    if not isinstance(captures, list) or len(captures) != len(expected):
        raise Blocked("capture-count-mismatch")
    previous = observed.start_us * 1000
    for capture, phase in zip(captures, expected):
        if not isinstance(capture, dict) or set(capture) != {"phase", "sha256", "bytes", "at_unix_ns"} or capture.get("phase") != phase:
            raise Blocked("capture-phase-or-schema-mismatch")
        stamp = strict_int(capture.get("at_unix_ns"), "capture-time")
        if not previous < stamp <= observed_at_us * 1000:
            raise Blocked("capture-time-not-fresh-or-ordered")
        previous = stamp
        count = strict_int(capture.get("bytes"), "capture-bytes")
        digest = capture.get("sha256")
        if count > MAX_CAPTURE or not isinstance(digest, str) or not re.fullmatch(r"[0-9a-f]{64}", digest):
            raise Blocked("capture-size-or-hash-invalid")
        # No caller path or filename: phase comes from this fixed local allowlist.
        raw = _read_capture(out_dir + "/" + phase + ".png")
        if len(raw) != count or hashlib.sha256(raw).hexdigest() != digest:
            raise Blocked("capture-size-or-hash-mismatch")
        if len(raw) < 33 or raw[:8] != b"\x89PNG\r\n\x1a\n" or raw[8:16] != b"\0\0\0\rIHDR":
            raise Blocked("capture-not-png-ihdr")
        width, height = struct.unpack(">II", raw[16:24])
        if not 1 <= width <= 4096 or not 1 <= height <= 4096:
            raise Blocked("capture-dimensions-out-of-bounds")
        if zlib.crc32(raw[12:29]) != struct.unpack(">I", raw[29:33])[0]:
            raise Blocked("capture-ihdr-crc-mismatch")
        offset=8; chunks=0; image_data=False; ended=False
        while offset<len(raw):
            chunks+=1
            if chunks>8192 or offset+12>len(raw):raise Blocked("capture-png-chunk-bound")
            length=struct.unpack(">I",raw[offset:offset+4])[0];end=offset+12+length
            if end>len(raw):raise Blocked("capture-png-truncated-chunk")
            kind=raw[offset+4:offset+8]
            if zlib.crc32(raw[offset+4:end-4])!=struct.unpack(">I",raw[end-4:end])[0]:raise Blocked("capture-png-chunk-crc")
            if kind==b"IDAT" and length>0:image_data=True
            if kind==b"IEND":
                if length!=0 or end!=len(raw):raise Blocked("capture-png-bad-end")
                ended=True
            offset=end
        if not image_data or not ended:raise Blocked("capture-png-incomplete")
    return {"status": "CAPTURE_ARTIFACTS_VALIDATED", "phases": expected}

def _os_phases(mode):
    if mode not in CASES: raise Blocked("OS phase mode")
    return ["before", "seed", "final"] + (["undo"] if mode == "text-only" else [])


def _capture_os_phases(mode, directory, nonce, observed, verifier, deadline, launcher):
    import window_capture
    completed = []
    fixed_window = None
    for phase in _os_phases(mode):
        ready, _, _ = _wait_artifact(directory+"/"+phase+".os-ready.json", deadline, launcher)
        phase_deadline = min(deadline, time.monotonic()+10)
        number = strict_int(ready.get("window_number"), "window-number")
        if number > 0xffffffff: raise Blocked("window-number-overflow")
        if fixed_window is not None and number != fixed_window: raise Blocked("OS-window-changed-between-phases")
        fixed_window = number
        token = dict(nonce=nonce, pid=observed.pid, start_us=observed.start_us, window_number=number, phase=phase)
        if ready != token: raise Blocked("OS-ready-owner-or-phase")
        identity = window_capture.Identity(nonce, observed.pid, observed.start_us, observed.executable, number)
        requested = []
        def receipt(stage, requested_phase, operation_deadline):
            if requested_phase != phase or stage not in ("before", "after") or requested != ([] if stage == "before" else ["before"]):
                raise Blocked("OS-receipt-order")
            requested.append(stage)
            _publish_json(directory+"/"+phase+".os-"+stage+"-request.json", token)
            value, _, _ = _wait_artifact(directory+"/"+phase+".os-"+stage+".json", min(phase_deadline, operation_deadline), launcher)
            return value
        try:
            proof = window_capture.capture_owned_window(identity=identity, phase=phase, verifier=verifier, own_receipt=receipt, deadline=phase_deadline, execute_native=True)
        except window_capture.Blocked as error: raise Blocked(str(error)) from error
        if requested != ["before", "after"] or proof.get("phase") != phase or proof.get("status") != "OS_WINDOW_CAPTURE_CANDIDATE":
            raise Blocked("OS-incomplete-capture")
        _publish_json(directory+"/"+phase+".os-proof.json", proof)
        _publish_json(directory+"/"+phase+".os-ack.json", dict(token, status="PASS"))
        completed.append(phase)
    return completed


def _native_case(mode, app, executable, writer, verifier):
    nonce = secrets.token_hex(32)
    directory = "/private/tmp/orbit-composer-" + nonce
    os.mkdir(directory, 0o700)  # fresh exclusive directory; never adopt an old run
    home = directory + "/home"; pi = directory + "/pi"; config_home = directory + "/config"
    os.mkdir(home, 0o700); os.mkdir(pi, 0o700); os.mkdir(config_home, 0o700)
    deadline = time.monotonic() + 90
    helper = launcher = observed = admission = writer_record = None
    result = {"mode": mode, "out_dir": directory, "status": "BLOCKED"}
    try:
        generation = 0
        if mode != "menu-cancel":
            helper = subprocess.Popen([writer, mode], stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                stderr=subprocess.DEVNULL, bufsize=0, env={"PATH": "/usr/bin:/bin", "HOME": home, "TMPDIR": "/private/tmp", "ORBIT_COMPOSER_DISPOSABLE_RUNNER": "1"})
            raw = _line(helper, min(deadline, time.monotonic()+10))
            _publish_raw(directory + "/writer.stdout.raw", raw)
            writer_record = validate_writer_line(raw, mode=mode, child_pid=helper.pid)
            _publish_json(directory + "/writer.json", writer_record)
            generation = writer_record["changeCount"]
        config = dict(mode=mode, out_dir=directory, nonce=nonce, clipboard_generation=generation)
        _publish_json(directory + "/config.json", config)
        environment = {"ORBIT_COMPOSER_DISPOSABLE_RUNNER": "1", "ORBIT_COMPOSER_PROOF_CONFIG": directory+"/config.json",
                       "HOME": home, "XDG_CONFIG_HOME": config_home, "PI_CODING_AGENT_DIR": pi, "PI_OFFLINE": "1", "PI_SKIP_VERSION_CHECK": "1"}
        argv = ["/usr/bin/open", "-n", "-W"]
        for key, value in environment.items(): argv.extend(["--env", key+"="+value])
        argv.append(app)
        launched_us = time.time_ns() // 1000
        launcher = subprocess.Popen(argv, stdin=subprocess.DEVNULL, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, env={"PATH": "/usr/bin:/bin", "HOME": home, "TMPDIR": "/private/tmp"})
        identity, raw_identity, identity_mtime = _wait_artifact(directory+"/identity.json", deadline, launcher)
        if identity.get("nonce") != nonce or identity.get("mode") != mode:
            raise Blocked("identity-nonce-or-case-mismatch")
        claimed_pid = strict_int(identity.get("pid"), "identity-pid")
        candidate, proof_raw = _inspect(verifier, claimed_pid, deadline)
        _publish_raw(directory+"/process-proof.stdout.raw", proof_raw)
        admission = CaseAdmission(mode, nonce, executable, launched_us, time.time_ns()//1000)
        validate_process(candidate, admission)
        if candidate.pid != claimed_pid or strict_int(identity.get("start_us"), "identity-start") != candidate.start_us or identity.get("executable") != candidate.executable:
            raise Blocked("identity-native-proof-mismatch")
        if not candidate.start_us <= identity_mtime <= admission.launch_after_us:
            raise Blocked("identity-artifact-not-fresh")
        observed = candidate  # only independently proved processes may be signaled
        _publish_json(directory+"/native-process-proof.json", unique_json(proof_raw, MAX_WRITER_LINE))
        _publish_json(directory+"/runner-approved.json", identity)
        os_phases = _capture_os_phases(mode, directory, nonce, observed, verifier, deadline, launcher)
        report, raw_report, report_mtime = _wait_artifact(directory+"/report.json", deadline, launcher)
        if report.get("os_capture_phases") != os_phases: raise Blocked("OS-phase-report-mismatch")
        now_us = time.time_ns()//1000
        validate_report_envelope(raw_report, admission=admission, observed=observed,
                                 report_mtime_us=report_mtime, observed_at_us=now_us)
        validate_behavior(report, observed_at_us=now_us, expected_owner=nonce+":1")
        validate_captures(report, out_dir=directory, admission=admission, observed=observed, observed_at_us=now_us)
        remaining = deadline-time.monotonic()
        if remaining <= 0 or launcher.wait(timeout=remaining) != 0:
            raise Blocked("app-did-not-exit-cleanly")
        result["status"] = "PASS"
    except (Blocked, OSError, ValueError, subprocess.TimeoutExpired) as error:
        result["reason"] = str(error)
        if observed is not None and admission is not None and launcher is not None and launcher.poll() is None:
            try: _terminate_verified(verifier, observed, admission)
            except (Blocked, OSError, ValueError, subprocess.TimeoutExpired) as cleanup_error:
                result["app_cleanup"] = "REFUSED: " + str(cleanup_error)
    finally:
        if helper is not None:
            try:
                if writer_record is None or helper.poll() is not None:
                    raise Blocked("live-validated-writer-unavailable")
                helper.stdin.write(("cleanup "+writer_record["cleanupToken"]+"\n").encode("ascii")); helper.stdin.flush()
                raw = _line(helper, time.monotonic()+5)
                _publish_raw(directory+"/writer-cleanup.stdout.raw", raw)
                cleanup = unique_json(raw, MAX_WRITER_LINE)
                _publish_json(directory+"/writer-cleanup.json", cleanup)
                if cleanup.get("status") != "CLEARED" or cleanup.get("pid") != helper.pid or cleanup.get("ownedChangeCount") != writer_record["changeCount"]:
                    raise Blocked("writer-cleanup-refused-or-invalid")
                if helper.wait(timeout=5) != 0: raise Blocked("writer-cleanup-exit")
            except (Blocked, OSError, ValueError, subprocess.TimeoutExpired) as error:
                result["status"] = "BLOCKED"; result["clipboard_cleanup"] = str(error)
            finally:
                # Never clear via another process, never restore ambient data.
                if helper.poll() is None: helper.terminate()
                helper.stdin.close(); helper.stdout.close()
        _publish_json(directory+"/runner-result.json", result)
    return result


def run_native_cases(*, app, writer, process_proof, execute_native=False):
    if execute_native is not True or platform.system() != "Darwin" or platform.machine() != "arm64" or os.environ.get("ORBIT_COMPOSER_DISPOSABLE_RUNNER") != "1":
        raise Blocked("native-admission-requires-explicit-Darwin-arm64-disposable-runner")
    executable = _bundle_executable(app)
    writer = _regular(writer, executable=True)
    verifier = _regular(process_proof, executable=True)
    results = []
    for mode in CASES:
        result = _native_case(mode, app, executable, writer, verifier)
        results.append(result)
        if result["status"] != "PASS": break  # no retries or later inputs after failure
    return {"status": "PASS" if len(results) == 4 and all(r["status"] == "PASS" for r in results) else "BLOCKED", "cases": results}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--execute-native", action="store_true")
    parser.add_argument("--app"); parser.add_argument("--writer"); parser.add_argument("--process-proof")
    args = parser.parse_args()
    if not args.execute_native:
        print(json.dumps(offline_manifest(), sort_keys=True)); return 2
    if not all((args.app, args.writer, args.process_proof)):
        parser.error("native execution requires explicit --app --writer --process-proof")
    try: result = run_native_cases(app=args.app, writer=args.writer, process_proof=args.process_proof, execute_native=True)
    except (Blocked, OSError, ValueError) as error: result = {"status": "BLOCKED", "reason": str(error)}
    print(json.dumps(result, sort_keys=True))
    return 0 if result["status"] == "PASS" else 2


if __name__ == "__main__":
    raise SystemExit(main())
