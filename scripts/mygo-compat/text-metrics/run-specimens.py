#!/usr/bin/env python3
"""Run the display-free candidate in fresh processes and require exact results."""
from __future__ import annotations

import argparse
import hashlib
import itertools
import json
import os
from pathlib import Path
import subprocess

EXPECTED = {
    "default": "c4fd4e0aa650873aa03c9a189d401f6a845f366e88467bc19bfaf82a5a231953",
    "fractional": "fd68942581716de92ce723c8fec08fb084c01bd4f77b1eb9bc07d5ead7ae219f",
}


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def check_roster(records: list) -> None:
    want = set(itertools.product(
        (".ZedSans", ".ZedMono"), (False, True), (400, 500, 600, 700),
        ("General", "This machine", "AV office fi 0123456789 → 中文 🙂")))
    got = [(r["style"]["Family"], r["style"]["Italic"], r["style"]["Weight"], r["text"])
           for r in records]
    if len(got) != 48 or set(got) != want or any(r["style"]["Size"] != 14 for r in records):
        raise ValueError("specimen roster is not the exact 48-case fixture")


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__)
    for name in ("binary", "assets", "default-reference", "enabled-reference", "output"):
        p.add_argument("--" + name, type=Path, required=True)
    args = p.parse_args()
    binary, assets, output = args.binary.resolve(), args.assets.resolve(), args.output.resolve()
    references = {"default": args.default_reference.resolve(), "fractional": args.enabled_reference.resolve()}
    for mode, reference in references.items():
        if sha(reference) != EXPECTED[mode]:
            p.error("reference hash does not match the validated specimen: " + mode)
        check_roster(json.loads(reference.read_text()))
    output.mkdir()  # Never overwrite a previous run.
    env = os.environ.copy()
    for key in ("DISPLAY", "WAYLAND_DISPLAY", "ORBIT_FONT_TRACE_DIR", "ORBIT_FONT_TRACE_TEXTS",
                "ORBIT_FONT_TRACE_CASE", "ORBIT_METRIC_EXPERIMENT", "ORBIT_FONT_SPECIMEN_RESULT",
                "ORBIT_COMPAT_SPECIMEN_ENABLE"):
        env.pop(key, None)
    env["ORBIT_FONT_TRACE_ASSETS"] = str(assets)
    result = {"schema": 1, "scope": "display-free shaping, no actual-app paint or trace instrumentation",
              "binary_sha256": sha(binary), "assets": {str(f.relative_to(assets)): sha(f)
              for f in sorted(assets.rglob("*")) if f.is_file()}, "runs": {}}
    for mode, reference in references.items():
        folder = output / mode
        (folder / "home").mkdir(parents=True)
        native_env = dict(env, HOME=str(folder / "home"), XDG_CACHE_HOME=str(folder / "home/cache"),
                          ORBIT_FONT_SPECIMEN_RESULT=str(folder / "result.json"))
        if mode == "fractional":
            native_env["ORBIT_COMPAT_SPECIMEN_ENABLE"] = "1"
        command = [str(binary), "-test.run=^TestOrbitTextMetricsSpecimen$", "-test.v"]
        run = subprocess.run(command, env=native_env, capture_output=True, text=True)
        (folder / "run.log").write_text(run.stdout + run.stderr)
        if run.returncode:
            raise SystemExit(run.stdout + run.stderr)
        generated = folder / "result.json"
        check_roster(json.loads(generated.read_text()))
        exact = generated.read_bytes() == reference.read_bytes()
        result["runs"][mode] = {"command": command, "exit_code": run.returncode, "cases": 48,
            "candidate_sha256": sha(generated), "reference": str(reference),
            "reference_sha256": sha(reference), "byte_identical": exact}
        (output / "specimen-equivalence.json").write_text(json.dumps(result, indent=2) + "\n")
        if not exact:
            raise SystemExit("candidate/reference byte difference: " + mode)
        print(mode + ": 48 cases byte-identical to the pinned reference", flush=True)


if __name__ == "__main__":
    main()
