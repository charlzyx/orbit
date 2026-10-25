#!/usr/bin/env python3
"""Verify Pi 1.4.0 RPC without credentials, model inference, or external tools.

Uses only Python's standard library. Launches the supplied official Pi executable
in a temporary empty home/workspace with no inherited provider environment.
The output is evidence, not a protocol fixture with deterministic session IDs.
"""

import argparse
import json
import os
from pathlib import Path
import queue
import shutil
import subprocess
import sys
import tempfile
import threading
import time

COMMANDS = (
    "get_state", "get_available_models", "abort",
    "get_available_thinking_levels", "get_messages", "get_session_stats",
    "get_entries", "get_commands", "auth.list", "quota.list",
)
ARGS = (
    "--mode", "rpc", "--no-session", "--no-approve", "--offline",
    "--no-extensions", "--no-skills", "--no-prompt-templates",
    "--no-context-files", "--no-tools",
)
START_EVENTS = {"agent_start", "message_start", "tool_execution_start"}


def check(condition, message):
    if not condition:
        raise RuntimeError(message)


def resolve_executable(value):
    found = shutil.which(value)
    if found:
        return str(Path(found).absolute())
    candidate = Path(value).expanduser().absolute()
    check(candidate.is_file(), "Pi executable was not found")
    return str(candidate)


def isolated_environment(home, pi):
    node = shutil.which("node")
    path_dirs = [str(Path(pi).parent)]
    if node:
        path_dirs.append(str(Path(node).parent))
    path_dirs.extend(os.defpath.split(os.pathsep))
    env = {
        "PATH": os.pathsep.join(dict.fromkeys(path_dirs)),
        "HOME": str(home),
        "USERPROFILE": str(home),
        "XDG_CONFIG_HOME": str(home / "config"),
        "XDG_CACHE_HOME": str(home / "cache"),
        "PI_CODING_AGENT_DIR": str(home / ".pi" / "agent"),
        "PI_OFFLINE": "1",
        "PI_SKIP_VERSION_CHECK": "1",
        "LANG": "C.UTF-8",
    }
    # Windows needs these OS paths; they are not credential/config variables.
    for key in ("SystemRoot", "WINDIR"):
        if key in os.environ:
            env[key] = os.environ[key]
    return env, node


def read_pipe(pipe, kind, events):
    try:
        # Binary readline splits only on LF, including on Windows. It does not
        # treat U+2028/U+2029 or UTF-8 multibyte sequences as record boundaries.
        while True:
            line = pipe.readline()
            if not line:
                break
            events.put((kind, line))
    finally:
        events.put((kind + "_eof", b""))


def sanitize(value, replacements):
    if isinstance(value, dict):
        return {k: sanitize(v, replacements) for k, v in value.items()}
    if isinstance(value, list):
        return [sanitize(v, replacements) for v in value]
    if isinstance(value, str):
        for original, placeholder in replacements:
            value = value.replace(original, placeholder)
    return value


def probe(pi, expectation, timeout):
    records, diagnostics = [], []
    with tempfile.TemporaryDirectory(prefix="orbit-pi-rpc-probe-") as temp:
        home = Path(temp)
        agent = home / ".pi" / "agent"
        workspace = home / "workspace"
        agent.mkdir(parents=True)
        workspace.mkdir()
        (agent / "settings.json").write_text(json.dumps({
            "enableInstallTelemetry": False, "enableAnalytics": False,
        }), encoding="utf-8")
        env, node = isolated_environment(home, pi)
        version = subprocess.run(
            [pi, "--version"], cwd=workspace, env=env,
            stdout=subprocess.PIPE, stderr=subprocess.PIPE,
            text=True, check=True, timeout=timeout,
        ).stdout.strip()
        check(version == "1.4.0", "This probe is pinned to Pi 1.4.0; refusing a different version")
        node_version = None
        if node:
            node_version = subprocess.run(
                [node, "--version"], cwd=workspace, env=env, check=True,
                stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                text=True, timeout=timeout,
            ).stdout.strip()

        process = subprocess.Popen(
            [pi, *ARGS], cwd=workspace, env=env,
            stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
        )
        events = queue.Queue()
        readers = [threading.Thread(target=read_pipe, args=(pipe, name, events), daemon=True)
                   for name, pipe in (("stdout", process.stdout), ("stderr", process.stderr))]
        for reader in readers:
            reader.start()
        request_ids = {f"contract-{index + 1}": command
                       for index, command in enumerate(COMMANDS)}
        seen = set()
        stdout_eof = False

        def consume(kind, line):
            nonlocal stdout_eof
            if kind == "stdout_eof":
                stdout_eof = True
                return
            if kind == "stderr_eof":
                return
            if kind == "stderr":
                diagnostics.append(line.decode("utf-8", errors="replace").rstrip("\r\n"))
                return
            if not line.strip():
                return
            record = json.loads(line)
            check(isinstance(record, dict), "Non-object protocol record")
            check(record.get("type") not in START_EVENTS, "Unexpected agent/message/tool run")
            records.append(record)
            if record.get("type") == "response" and record.get("id") in request_ids:
                check(record.get("command") == request_ids[record["id"]], "Response correlation mismatch")
                check(record["id"] not in seen, "Duplicate response ID")
                seen.add(record["id"])

        try:
            for request_id, command in request_ids.items():
                process.stdin.write((json.dumps({"id": request_id, "type": command}) + "\n").encode())
            process.stdin.flush()
            deadline = time.monotonic() + timeout
            while len(seen) < len(COMMANDS):
                remaining = deadline - time.monotonic()
                check(remaining > 0, "Timed out waiting for RPC responses")
                try:
                    consume(*events.get(timeout=min(remaining, 0.25)))
                except queue.Empty:
                    check(not stdout_eof, "Pi closed stdout before all responses arrived")
            process.stdin.close()
            exit_code = process.wait(timeout=timeout)
            for reader in readers:
                reader.join(timeout=timeout)
                check(not reader.is_alive(), "Pipe reader did not finish after process shutdown")
            while not events.empty():
                consume(*events.get_nowait())
            check(exit_code == 0, "Pi did not exit successfully after stdin close")
        finally:
            if process.poll() is None:
                process.kill()
                process.wait(timeout=timeout)
            for pipe in (process.stdin, process.stdout, process.stderr):
                pipe.close()

        responses = {record["command"]: record for record in records
                     if record.get("type") == "response"}
        for command in COMMANDS[:8]:
            check(responses[command].get("success") is True, command + " failed")
        check(responses["get_available_models"]["data"]["models"] == [], "Unexpected configured models")
        check(responses["get_messages"]["data"]["messages"] == [], "Unexpected existing messages")
        stats = responses["get_session_stats"]["data"]
        check(stats["tokens"]["total"] == 0 and stats["cost"] == 0, "Unexpected model usage")
        check(stats["toolCalls"] == 0, "Unexpected tool execution")
        if expectation == "stock":
            for command in ("auth.list", "quota.list"):
                check(responses[command].get("success") is False, "Unexpected patched capability: " + command)
                check("Unknown command" in responses[command].get("error", ""), "Unexpected unsupported-command error")
        elif expectation == "patched":
            for command in ("auth.list", "quota.list"):
                check(responses[command].get("success") is True, command + " patch unavailable")
            providers = responses["auth.list"]["data"]["providers"]
            check(all(not provider["authenticated"] for provider in providers), "Unexpected authenticated provider")
            check(responses["quota.list"]["data"]["providers"] == [], "Unexpected configured quota provider")
            capabilities = responses["get_state"]["data"].get("capabilities", {}).get("extension_ui", [])
            check("custom" in capabilities, "Custom UI patch capability missing")
        result = {
            "runtime": "@earendil-works/pi-coding-agent@" + version,
            "expectation": expectation, "node": node_version,
            "arguments": list(ARGS), "credential_environment_inherited": False,
            "external_automatic_network_disabled": True,
            "commands": list(COMMANDS), "responses": records,
            "stderr": diagnostics, "exit_code": exit_code,
            "model_or_tool_start_events": 0,
        }
        replacements = [(str(home), "<isolated-home>"),
                        (str(Path(pi).parent), "<pi-bin-dir>")]
        if node:
            replacements.append((str(Path(node).parent), "<node-bin-dir>"))
        return sanitize(result, replacements)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pi", required=True, help="Path to the official Pi 1.4.0 executable")
    parser.add_argument("--output", required=True, type=Path, help="JSON evidence destination")
    parser.add_argument("--expect", choices=("stock", "patched", "either"), default="either")
    parser.add_argument("--timeout", type=float, default=30, help="Positive per-stage timeout in seconds")
    args = parser.parse_args()
    check(args.timeout > 0, "Timeout must be positive")
    result = probe(resolve_executable(args.pi), args.expect, args.timeout)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    for record in result["responses"]:
        if record.get("type") == "response":
            print(record["command"] + ": " + ("success" if record["success"] else record.get("error", "failed")))
    print("Clean exit 0; no model/tool start events; evidence saved")


if __name__ == "__main__":
    try:
        main()
    except (OSError, RuntimeError, ValueError, subprocess.SubprocessError) as error:
        print("Probe failed: " + str(error), file=sys.stderr)
        sys.exit(1)
