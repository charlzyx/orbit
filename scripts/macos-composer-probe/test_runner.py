import json
import base64
import hashlib
import struct
import zlib
import io
import os
import tempfile
from contextlib import ExitStack
import unittest
from dataclasses import replace
from unittest.mock import Mock, patch

from runner import (Blocked, CASES, CaseAdmission, ProcessObservation,
                    MAX_REPORT, MAX_WRITER_LINE, native_process_observation,
                    offline_manifest, run_native_cases, same_owned_process, validate_behavior,
                    unique_json, validate_process, validate_report_envelope,
                    validate_writer_line)
from test_writer_contract import record
from writer_contract import PIXEL_SHA256


class RunnerContractTests(unittest.TestCase):
    def setUp(self):
        self.admission = CaseAdmission("png-text", "a"*64,
            "/private/tmp/fresh/Orbit.app/Contents/MacOS/Orbit", 100, 200)
        self.observed = ProcessObservation(42, 150, self.admission.executable)
        self.report = dict(nonce=self.admission.nonce, pid=42, start_us=150,
            executable=self.admission.executable, mode="png-text", status="PASS", inputlog=[])

    def envelope(self, report=None, **kwargs):
        options = dict(admission=self.admission, observed=self.observed,
                       report_mtime_us=160, observed_at_us=170)
        options.update(kwargs)
        return validate_report_envelope(json.dumps(self.report if report is None else report).encode(), **options)

    def test_exact_four_cases_and_no_native_claim(self):
        self.assertEqual(CASES, ("png-text", "tiff-text", "text-only", "menu-cancel"))
        manifest = offline_manifest()
        self.assertEqual(manifest["status"], "BLOCKED")
        self.assertEqual(manifest["native_execution"], "NOT_RUN")

    def test_native_unimplemented_not_fallback(self):
        with patch("subprocess.Popen", side_effect=AssertionError("must not spawn")), patch("os.kill", side_effect=AssertionError("must not kill")):
            with self.assertRaisesRegex(Blocked, "NOT_IMPLEMENTED"):
                native_process_observation(42)
            with self.assertRaises(Blocked):
                run_native_cases(app="/a.app", writer="/writer", process_proof="/proof")

    def test_writer_valid_lines(self):
        for mode in CASES[:3]:
            validate_writer_line(json.dumps(record(mode)).encode()+b"\n", mode=mode, child_pid=42)

    def test_writer_reject_partial_multiple_oversized_or_menu(self):
        for raw, mode in ((b"{}", "png-text"), (b"{}\n{}\n", "png-text"),
                          (b" "*MAX_WRITER_LINE+b"\n", "png-text"),
                          (json.dumps(record()).encode()+b"\n", "menu-cancel")):
            with self.assertRaises(Blocked):
                validate_writer_line(raw, mode=mode, child_pid=42)

    def test_json_reject_duplicate_and_nonfinite(self):
        for raw in (b'{"pid":1,"pid":42}', b'{"x":NaN}', b'{"x":Infinity}', b'[]', b'\xff'):
            with self.assertRaises(Blocked): unique_json(raw, MAX_REPORT)

    def test_valid_envelope_never_native_pass(self):
        result = self.envelope()
        self.assertEqual(result["status"], "ENVELOPE_VALIDATED_ONLY")
        self.assertEqual(result["native_acceptance"], "NOT_IMPLEMENTED")

    def test_report_cannot_claim_other_process(self):
        for key, value in (("pid", 43), ("pid", True), ("start_us", 149),
                           ("nonce", "b"*64), ("executable", "/tmp/other"),
                           ("mode", "menu-cancel"), ("status", "SUCCESS"),
                           ("inputlog", ["click"]*17)):
            report = dict(self.report); report[key] = value
            with self.subTest(key=key), self.assertRaises(Blocked): self.envelope(report)

    def test_report_stale_future_or_oversized(self):
        for mtime in (149, 171, True):
            with self.assertRaises(Blocked): self.envelope(report_mtime_us=mtime)
        with self.assertRaises(Blocked):
            validate_report_envelope(b" "*(MAX_REPORT+1), admission=self.admission,
                observed=self.observed, report_mtime_us=160, observed_at_us=170)

    def test_process_preexisting_wrong_path_noncanonical(self):
        for observed in (replace(self.observed, start_us=99), replace(self.observed, start_us=201),
                         replace(self.observed, pid=True), replace(self.observed, executable="/tmp/other"),
                         replace(self.observed, executable="/private/tmp/fresh/../fresh/Orbit.app/Contents/MacOS/Orbit")):
            with self.assertRaises(Blocked): validate_process(observed, self.admission)

    def test_pid_reuse_refused(self):
        self.assertTrue(same_owned_process(self.observed, self.observed, self.admission))
        with self.assertRaisesRegex(Blocked, "identity-changed"):
            same_owned_process(self.observed, replace(self.observed, start_us=151), self.admission)
        with self.assertRaises(Blocked):
            same_owned_process(self.observed, replace(self.observed, pid=43), self.admission)

    def test_missing_independent_observation(self):
        with self.assertRaises(Blocked): self.envelope(observed=None)
        with self.assertRaises(Blocked): self.envelope(observed=self.report)


    def behavior_report(self, mode):
        report = dict(self.report, schema=1, mode=mode, window=7)
        actions = ["click:Message", "key:s", "key:e", "key:e", "key:d"]
        actions += ["click:Add attachment", "key:Escape", "key:x"] if mode == "menu-cancel" else ["key:CmdV"] + (["key:CmdZ"] if mode == "text-only" else [])
        report["inputlog_complete"] = True
        report["inputlog"] = [dict(action=action, at_unix_ns=151000+i*1000, pid=42, start_us=150, window=7) for i, action in enumerate(actions)]
        for event in report["inputlog"]:
            event["target_source"]="matching-native-menu-begin" if event["action"]=="key:Escape" else "fresh-unique-AX"
            if event["action"]!="key:Escape":
                event["target"]={"identity":{"pid":42,"start_us":150,"window":7},"at_unix_ns":event["at_unix_ns"]-1,"label":"Add attachment" if event["action"]=="click:Add attachment" else "Message","matches":1,"nodes":12,"content":{"x":0,"y":0,"w":100,"h":100},"rect":{"x":1,"y":1,"w":10,"h":10}}

        def state(phase, stamp, draft, attachments):
            return dict(phase=phase, at_unix_ns=stamp, draft=draft, attachments=attachments,
                        pending=False, focused=True, ready=True, owner=self.admission.nonce+":1", messages=0)
        report["states"] = [state("seed", 155500, "seed", [])]
        if mode == "text-only":
            report["states"] += [state("final", 156500, "seedclipboard-text", []), state("undo", 158000, "seed", [])]
        elif mode == "menu-cancel":
            report["states"] += [state("final", 159000, "seedx", [])]
            report["menu"] = dict(begins=1, ends=1, actions=0, escape_requests=1)
        else:
            report["states"] += [state("final", 158000, "seed", [dict(mime="image/png", width=2, height=2, pixel_sha256=PIXEL_SHA256)])]
        return report

    def test_incomplete_native_finalization_refused(self):
        for value in (False,None,"true"):
            report=self.behavior_report("png-text"); report["inputlog_complete"]=value
            with self.assertRaises(Blocked): self.oracle(report)

    def test_AX_freshness_clip_unique_owner_negative(self):
        for mutation in (lambda t:t.update(matches=2),lambda t:t.update(nodes=2049),lambda t:t["identity"].update(window=8),lambda t:t["rect"].update(x=200),lambda t:t["rect"].update(w=float("nan")),lambda t:t.update(at_unix_ns=999999999)):
            report=self.behavior_report("png-text");mutation(report["inputlog"][0]["target"])
            with self.assertRaises(Blocked):self.oracle(report)

    def oracle(self, report):
        return validate_behavior(report, observed_at_us=170, expected_owner=self.admission.nonce+":1")

    def test_four_behavior_oracles_still_not_native_evidence(self):
        for mode in CASES:
            result = self.oracle(self.behavior_report(mode))
            self.assertEqual(result["status"], "OFFLINE_ORACLE_MATCH")
            self.assertEqual(result["native_acceptance"], "NOT_IMPLEMENTED")

    def test_missing_repeated_and_forbidden_actions(self):
        for mode in CASES:
            for action in ("key:Enter", "key:CmdV", "click:Add attachment"):
                report = self.behavior_report(mode)
                report["inputlog"].append(dict(report["inputlog"][-1], action=action))
                with self.assertRaises(Blocked): self.oracle(report)
            report = self.behavior_report(mode); report["inputlog"].pop()
            with self.assertRaises(Blocked): self.oracle(report)
            report = self.behavior_report(mode); report["inputlog"][0]["action"] = "performAction"
            with self.assertRaises(Blocked): self.oracle(report)

    def test_state_semantic_negative_cases(self):
        for key, value in (("draft", "seedclipboard-text"), ("pending", True),
                           ("focused", False), ("ready", False), ("messages", 1),
                           ("messages", False), ("owner", "other"), ("attachments", [])):
            report = self.behavior_report("png-text")
            report["states"][-1][key] = value
            with self.subTest(key=key), self.assertRaises(Blocked): self.oracle(report)
        report = self.behavior_report("text-only"); report["states"].pop()
        with self.assertRaises(Blocked): self.oracle(report)

    def test_event_and_state_timing_identity_negative(self):
        for key, value in (("pid", 43), ("start_us", 151), ("window", 8), ("at_unix_ns", 149000)):
            report = self.behavior_report("png-text"); report["inputlog"][0][key] = value
            with self.assertRaises(Blocked): self.oracle(report)
        for index, stamp in ((0, 155000), (0, 156000), (1, 155500), (1, 170000)):
            report = self.behavior_report("png-text"); report["states"][index]["at_unix_ns"] = stamp
            with self.assertRaises(Blocked): self.oracle(report)

    def test_menu_must_open_close_no_action_one_escape(self):
        for key, value in (("begins", 0), ("ends", 0), ("actions", 1),
                           ("escape_requests", 2), ("begins", True)):
            report = self.behavior_report("menu-cancel"); report["menu"][key] = value
            with self.assertRaises(Blocked): self.oracle(report)

    def test_optional_before_state_strict_and_only_first(self):
        report = self.behavior_report("png-text")
        before = dict(report["states"][0], phase="before", draft="", at_unix_ns=150500, focused=False)
        report["states"].insert(0, before)
        self.oracle(report)
        for field, value in (("draft", "ambient"), ("attachments", [{}]), ("messages", 1), ("pending", True), ("at_unix_ns", 151000)):
            modified = dict(before); modified[field] = value
            report["states"][0] = modified
            with self.assertRaises(Blocked): self.oracle(report)
        report["states"] = report["states"][1:] + [before]
        with self.assertRaises(Blocked): self.oracle(report)

    def test_native_admission_never_spawns_on_wrong_platform_or_gate(self):
        with patch("runner.subprocess.Popen", side_effect=AssertionError("no spawn")), patch("runner.platform.system", return_value="Linux"), patch("runner.platform.machine", return_value="arm64"), patch.dict("os.environ", {"ORBIT_COMPOSER_DISPOSABLE_RUNNER": "1"}):
            with self.assertRaises(Blocked): run_native_cases(app="/a.app", writer="/w", process_proof="/p", execute_native=True)
        with patch("runner.subprocess.Popen", side_effect=AssertionError("no spawn")), patch("runner.platform.system", return_value="Darwin"), patch("runner.platform.machine", return_value="arm64"), patch.dict("os.environ", {}, clear=True):
            with self.assertRaises(Blocked): run_native_cases(app="/a.app", writer="/w", process_proof="/p", execute_native=True)

    def test_native_matrix_stops_on_first_failure_no_retry(self):
        with patch("runner.platform.system", return_value="Darwin"), patch("runner.platform.machine", return_value="arm64"), patch.dict("os.environ", {"ORBIT_COMPOSER_DISPOSABLE_RUNNER": "1"}), patch("runner._bundle_executable", return_value="/a.app/Contents/MacOS/a"), patch("runner._regular", side_effect=lambda p, **kw: p), patch("runner._native_case", side_effect=[{"status":"PASS"},{"status":"BLOCKED"}]) as case:
            result = run_native_cases(app="/a.app", writer="/w", process_proof="/p", execute_native=True)
            self.assertEqual(result["status"], "BLOCKED")
            self.assertEqual([call.args[0] for call in case.call_args_list], list(CASES[:2]))

    def test_terminate_requires_fresh_same_process_inspection(self):
        from runner import _terminate_verified
        with patch("runner._inspect", return_value=(self.observed, b"{}")), patch("runner.os.kill") as kill:
            _terminate_verified("/proof", self.observed, self.admission)
            self.assertEqual(kill.call_count, 1)
            self.assertEqual(kill.call_args.args[0], self.observed.pid)
        with patch("runner._inspect", return_value=(replace(self.observed, start_us=151), b"{}")), patch("runner.os.kill") as kill:
            with self.assertRaises(Blocked): _terminate_verified("/proof", self.observed, self.admission)
            kill.assert_not_called()

    def test_mocked_native_handshake_launch_and_cleanup(self):
        from runner import _native_case
        helper = Mock(pid=42, stdin=io.BytesIO(), stdout=io.BytesIO())
        helper.poll.return_value = None; helper.wait.return_value = 0
        launcher = Mock(); launcher.poll.return_value = None; launcher.wait.return_value = 0
        identity = dict(pid=42, start_us=150, executable=self.admission.executable, nonce=self.admission.nonce, mode="png-text")
        report = self.behavior_report("png-text")
        report["os_capture_phases"] = ["before", "seed", "final"]
        cleanup = dict(status="CLEARED", pid=42, ownedChangeCount=10, changeCount=11)
        with ExitStack() as stack:
            stack.enter_context(patch("runner.os.mkdir"))
            stack.enter_context(patch("runner.secrets.token_hex", return_value=self.admission.nonce))
            stack.enter_context(patch("runner.time.time_ns", side_effect=[100000,170000,170000]))
            publish = stack.enter_context(patch("runner._publish_json"))
            stack.enter_context(patch("runner._publish_raw"))
            stack.enter_context(patch("runner.validate_captures"))
            os_capture = stack.enter_context(patch("runner._capture_os_phases", return_value=["before", "seed", "final"]))
            popen = stack.enter_context(patch("runner.subprocess.Popen", side_effect=[helper, launcher]))
            stack.enter_context(patch("runner._line", side_effect=[json.dumps(record()).encode()+b"\n", json.dumps(cleanup).encode()+b"\n"]))
            stack.enter_context(patch("runner._inspect", return_value=(self.observed, json.dumps({"pid":42,"start_us":150,"executable":self.admission.executable}).encode()+b"\n")))
            stack.enter_context(patch("runner._wait_artifact", side_effect=[(identity,json.dumps(identity).encode(),155),(report,json.dumps(report).encode(),160)]))
            result = _native_case("png-text", "/a.app", self.admission.executable, "/writer", "/proof")
            self.assertEqual(result["status"], "PASS")
            os_capture.assert_called_once()
            self.assertEqual(os_capture.call_args.args[0], "png-text")
            self.assertEqual(os_capture.call_args.args[2], self.admission.nonce)
            self.assertEqual(os_capture.call_args.args[3], self.observed)
            argv = popen.call_args_list[1].args[0]
            self.assertEqual(argv[:3], ["/usr/bin/open", "-n", "-W"])
            self.assertIn("PI_OFFLINE=1", argv); self.assertIn("PI_SKIP_VERSION_CHECK=1", argv)
            self.assertIn("ORBIT_COMPOSER_DISPOSABLE_RUNNER=1", argv)
            paths = [call.args[0] for call in publish.call_args_list]
            self.assertTrue(any(path.endswith("/runner-approved.json") for path in paths))
            self.assertTrue(any(path.endswith("/writer-cleanup.json") for path in paths))

    def test_identity_timeout_never_signals_unverified_process(self):
        from runner import _native_case
        launcher = Mock(); launcher.poll.return_value = None
        with patch("runner.os.mkdir"), patch("runner._publish_json") as publish, patch("runner.subprocess.Popen", return_value=launcher), patch("runner._wait_artifact", side_effect=Blocked("deadline")), patch("runner._inspect") as inspect, patch("runner.os.kill") as kill:
            result = _native_case("menu-cancel", "/a.app", self.admission.executable, "/writer", "/proof")
            self.assertEqual(result["status"], "BLOCKED")
            kill.assert_not_called(); inspect.assert_not_called()
            self.assertFalse(any(call.args[0].endswith("runner-approved.json") for call in publish.call_args_list))

    def test_private_artifact_rejects_symlink_fifo_oversized(self):
        from runner import _read_private_json
        with tempfile.TemporaryDirectory(prefix="composer-contract-") as directory:
            path = directory + "/report.json"
            with open(path, "wb") as stream: stream.write(b"{}")
            self.assertEqual(_read_private_json(path, 16)[0], {})
            link = directory + "/link"; os.symlink(path, link)
            with self.assertRaises(OSError): _read_private_json(link, 16)
            fifo = directory + "/fifo"; os.mkfifo(fifo)
            with self.assertRaises(Blocked): _read_private_json(fifo, 16)
            with self.assertRaises(Blocked): _read_private_json(path, 1)

    def capture_fixture(self):
        return base64.b64decode("iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAEklEQVR42mP4z8DwHwyBNBgAAEnICfcD2WTxAAAAAElFTkSuQmCC")

    def capture_report(self, mode="png-text", raw=None):
        raw = self.capture_fixture() if raw is None else raw
        phases = ["before", "seed", "final"] + (["undo"] if mode == "text-only" else [])
        return {"capture_kind":"content-rerender","captures": [dict(phase=phase, sha256=hashlib.sha256(raw).hexdigest(), bytes=len(raw), at_unix_ns=151000+i*1000) for i, phase in enumerate(phases)]}

    def captures(self, report, mode="png-text", out_dir=None):
        from runner import validate_captures
        return validate_captures(report, out_dir=out_dir or "/private/tmp/orbit-composer-"+self.admission.nonce,
            admission=replace(self.admission, mode=mode), observed=self.observed, observed_at_us=170)

    def test_capture_exact_phases_and_fixed_paths(self):
        with patch("runner._read_capture", return_value=self.capture_fixture()) as read:
            for mode in CASES:
                self.captures(self.capture_report(mode), mode)
            self.assertTrue(all(call.args[0].startswith("/private/tmp/orbit-composer-"+self.admission.nonce+"/") for call in read.call_args_list))
            self.assertTrue(all(call.args[0].endswith(("/before.png", "/seed.png", "/final.png", "/undo.png")) for call in read.call_args_list))

    def test_capture_refuses_phases_paths_hash_bytes_and_times(self):
        with patch("runner._read_capture", return_value=self.capture_fixture()):
            for field, value in (("phase", "../other"), ("phase", "seed"), ("sha256", "0"*64), ("bytes", 74), ("bytes", True), ("at_unix_ns", 149000), ("at_unix_ns", 171000)):
                report = self.capture_report(); report["captures"][0][field] = value
                with self.subTest(field=field), self.assertRaises(Blocked): self.captures(report)
            report = self.capture_report(); report["captures"].pop()
            with self.assertRaises(Blocked): self.captures(report)
            report = self.capture_report(); report["captures"][0]["path"] = "/tmp/ambient.png"
            with self.assertRaises(Blocked): self.captures(report)
            with self.assertRaises(Blocked): self.captures(self.capture_report(), out_dir="/tmp/other")

    def test_capture_refuses_png_header_or_dimensions(self):
        originals = [b"x"*75, self.capture_fixture()[:32],self.capture_fixture()[:33],self.capture_fixture()[:-12],self.capture_fixture()+b"trailing"]
        for dimensions in ((0, 2), (4097, 2), (2, 4097)):
            raw = bytearray(self.capture_fixture()); raw[16:24] = struct.pack(">II", *dimensions)
            raw[29:33] = struct.pack(">I", zlib.crc32(raw[12:29])); originals.append(bytes(raw))
        raw = bytearray(self.capture_fixture()); raw[29] ^= 1; originals.append(bytes(raw))
        for raw in originals:
            with patch("runner._read_capture", return_value=raw), self.assertRaises(Blocked): self.captures(self.capture_report(raw=raw))

    def test_capture_local_file_rejects_symlink_fifo_and_oversize(self):
        from runner import _read_capture, MAX_CAPTURE
        with tempfile.TemporaryDirectory(prefix="composer-capture-contract-") as directory:
            path = directory + "/before.png"
            with open(path, "wb") as stream: stream.write(self.capture_fixture())
            self.assertEqual(_read_capture(path), self.capture_fixture())
            link = directory + "/link"; os.symlink(path, link)
            with self.assertRaises(OSError): _read_capture(link)
            fifo = directory + "/fifo"; os.mkfifo(fifo)
            with self.assertRaises(Blocked): _read_capture(fifo)
            with open(path, "wb") as stream: stream.truncate(MAX_CAPTURE+1)
            with self.assertRaises(Blocked): _read_capture(path)

    def test_invalid_admission(self):
        for admission in (replace(self.admission, nonce="short"),
                          replace(self.admission, mode="paste"),
                          replace(self.admission, launch_after_us=99)):
            with self.assertRaises(Blocked): validate_process(self.observed, admission)


if __name__ == "__main__":
    unittest.main()
