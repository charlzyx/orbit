#!/usr/bin/env python3
"""Hash-verified, temporary Go overlay for Orbit's pinned MyGo compatibility.

No source/module-cache writes. Forward every Go argument and ambient build env.
A caller overlay is merged only if it does not replace verified dependencies.
"""
from __future__ import annotations
import hashlib
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parent.parent
COMPAT = ROOT / "third_party" / "mygo-composer"
SELECTION = ROOT / "third_party" / "mygo-selection"


def fail(message: str) -> None:
    raise RuntimeError(message)


def split_flags(flags: str) -> list[str]:
    # Exact Go cmd/internal/quoted grammar: whole-token quotes; no shell
    # expansion, backslash unescaping, or embedded-quote processing.
    result = []
    offset = 0
    while offset < len(flags):
        while offset < len(flags) and flags[offset] in " \t\r\n":
            offset += 1
        if offset == len(flags):
            break
        if flags[offset] in ("'", '"'):
            quote = flags[offset]
            end = flags.find(quote, offset + 1)
            if end < 0:
                fail(f"Unterminated GOFLAGS {quote} string")
            result.append(flags[offset + 1:end])
            offset = end + 1
        else:
            end = offset
            while end < len(flags) and flags[end] not in " \t\r\n":
                end += 1
            result.append(flags[offset:end])
            offset = end
    return result


def quote_flag(flag: str) -> str:
    if flag == "":
        return '""'
    if not any(c.isspace() for c in flag) and not flag.startswith(("'", '"')):
        return flag
    if '"' not in flag:
        return '"' + flag + '"'
    if "'" not in flag:
        return "'" + flag + "'"
    fail("A GOFLAGS token contains whitespace and both quote types; Go cannot encode it")


def detach_overlay(flags: list[str]) -> tuple[list[str], str | None]:
    out, overlay, i = [], None, 0
    while i < len(flags):
        arg = flags[i]
        if arg == "-overlay":
            if i + 1 >= len(flags):
                fail("GOFLAGS -overlay requires a path")
            overlay = flags[i + 1]
            i += 2
        elif arg.startswith("-overlay="):
            overlay = arg.partition("=")[2]
            i += 1
        else:
            out.append(arg)
            i += 1
    return out, overlay


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


# Explicit reviewed paths, never manifest or filesystem glob discovery.
SELECTION_BASELINE = frozenset({"ui/context.go", "ui/textselection.go", "ui/input.go"})
SELECTION_ADDED = frozenset({"ui/orbit_selection.go", "ui/orbit_selection_test.go"})
SELECTION_IDENTITIES = ("upstream_commit", "module", "version", "module_sum", "go_mod_sum")


def verify_selection(module_dir: Path, compatibility: Path, expected: dict) -> dict[str, Path]:
    manifest = json.loads((compatibility / "baseline.json").read_text())
    if set(manifest) != set(SELECTION_IDENTITIES) | {"baseline_sha256", "overlays", "patch_sha256"}:
        fail("Unexpected selection compatibility manifest schema")
    if any(manifest.get(key) != expected.get(key) for key in SELECTION_IDENTITIES):
        fail("Selection compatibility identity differs from the reviewed MyGo pin")
    if set(manifest["baseline_sha256"]) != SELECTION_BASELINE:
        fail("Selection baseline paths differ from the explicit reviewed set")
    if set(manifest["overlays"]) != SELECTION_BASELINE | SELECTION_ADDED:
        fail("Selection overlay paths differ from the explicit reviewed set")
    if sha(compatibility / "selection-retention.patch") != manifest["patch_sha256"]:
        fail("Selection compatibility patch hash mismatch")
    for relative, digest in manifest["baseline_sha256"].items():
        original = module_dir / relative
        if not original.is_file() or sha(original) != digest:
            fail(f"Pinned MyGo selection baseline hash mismatch: {relative}")
    result = {}
    for relative, info in manifest["overlays"].items():
        if set(info) != {"source", "sha256", "added"}:
            fail(f"Unexpected selection overlay schema: {relative}")
        is_added = relative in SELECTION_ADDED
        if type(info["added"]) is not bool or info["added"] != is_added:
            fail(f"Selection additive/replacement kind mismatch: {relative}")
        if info["source"] != Path(relative).name + ".txt":
            fail(f"Selection source path differs from the reviewed file: {relative}")
        source = compatibility / info["source"]
        if source.resolve().parent != compatibility.resolve() or not source.is_file() or sha(source) != info["sha256"]:
            fail(f"Selection compatibility source hash/path mismatch: {relative}")
        if is_added and (module_dir / relative).exists():
            fail(f"Unexpected selection addition in module cache: {relative}")
        result[relative] = source
    return result


def main() -> int:
    if len(sys.argv) < 2:
        fail("Usage: scripts/go-native.sh <run|build|test|vet|other-go-command> [arguments...]")
    go = os.environ.get("ORBIT_GO") or shutil.which("go")
    if not go:
        fail("Go 1.27.3 is required; set ORBIT_GO to its go executable")
    env = os.environ.copy()
    args = sys.argv[1:]
    effective_cwd = Path.cwd()
    if args and args[0] == "-C":
        if len(args) < 3:
            fail("go -C requires a directory and command")
        effective_cwd = (effective_cwd / args[1]).resolve()
    elif args and args[0].startswith("-C="):
        effective_cwd = (effective_cwd / args[0][3:]).resolve()
    # Command-line -overlay would supersede our mandatory overlay. Explicitly
    # reject it, rather than silently building the unsafe fallback.
    for arg in sys.argv[2:]:
        if arg == "-overlay" or arg.startswith("-overlay="):
            fail("Pass a custom -overlay through GOFLAGS; the native wrapper merges it safely")
    inherited = env.get("GOFLAGS")
    if inherited is None:
        inherited = subprocess.check_output([go, "env", "GOFLAGS"], text=True).strip()
    flags, previous_overlay = detach_overlay(split_flags(inherited))
    if any(f == "-modfile" or f.startswith("-modfile=") for f in flags + args):
        fail("The native wrapper owns its temporary -modfile; remove the conflicting flag")
    env["GOFLAGS"] = " ".join(quote_flag(f) for f in flags)
    def check_output(argv):
        return subprocess.check_output([go, *argv], env=env, text=True, cwd=effective_cwd)
    selected = json.loads(check_output(["env", "-json", "GOMOD", "GOWORK"]))
    if selected.get("GOWORK") not in (None, "", "off"):
        fail("Native compatibility builds require GOWORK=off, not a workspace override")
    main_mod = Path(selected.get("GOMOD", ""))
    if not main_mod.is_file():
        fail("Run the native wrapper from the Orbit module or its terminal submodule")
    manifest = json.loads((COMPAT / "baseline.json").read_text())
    module = json.loads(check_output(["list", "-m", "-json", manifest["module"]]))
    if module.get("Path") != manifest["module"] or module.get("Version") != manifest["version"] or module.get("Replace"):
        fail("MyGo dependency is not the reviewed, unreplaced upstream pin")
    directory = module.get("Dir")
    if not directory or not Path(directory).is_dir():
        # A fresh module cache may have only its .mod metadata. Materialize the
        # exact pinned dependency through Go's authenticated module download.
        downloaded = json.loads(check_output(["mod", "download", "-json", f"{manifest['module']}@{manifest['version']}"]))
        if downloaded.get("Error"):
            fail(f"Cannot materialize pinned MyGo: {downloaded['Error']}")
        if downloaded.get("Path") != manifest["module"] or downloaded.get("Version") != manifest["version"]:
            fail("Downloaded MyGo identity differs from the reviewed pin")
        module.update(downloaded)
        directory = module.get("Dir")
    if not directory or not Path(directory).is_dir():
        fail("Go did not materialize the pinned MyGo source directory")
    module_dir = Path(directory)
    if module.get("Sum") != manifest["module_sum"] or module.get("GoModSum") != manifest["go_mod_sum"]:
        fail("Pinned MyGo go.sum identity mismatch")
    subprocess.run([go, "mod", "verify"], env=env, cwd=effective_cwd, check=True, stdout=subprocess.DEVNULL)
    for relative, digest in manifest["baseline_sha256"].items():
        path = module_dir / relative
        if not path.is_file() or sha(path) != digest:
            fail(f"Pinned MyGo baseline hash mismatch: {relative}")
    source = COMPAT / "orbit_composer.go.txt"
    if sha(source) != manifest["added_sha256"]:
        fail("Composer compatibility source hash mismatch; review/update the manifest")
    selection_overlays = verify_selection(module_dir, SELECTION, manifest)
    overlay: dict[str, str] = {}
    if previous_overlay:
        prior = json.loads((effective_cwd / previous_overlay).read_text())
        if set(prior) != {"Replace"} or not isinstance(prior["Replace"], dict):
            fail("Unsupported caller overlay; expected a Replace mapping")
        for original, replacement in prior["Replace"].items():
            original_path = (effective_cwd / original).absolute()
            # Disallow overriding any file in MyGo, not merely the design hashes.
            # Otherwise unreviewed code can supersede the verified dependency.
            if original_path.is_relative_to(module_dir.absolute()):
                fail(f"Caller overlay may not replace pinned MyGo source: {original}")
            overlay[str(original_path)] = str((effective_cwd / replacement).absolute()) if replacement else ""
    target = module_dir / manifest["added_file"]
    if target.exists():
        fail("Unexpected composer file already present in module cache; do not modify the cache")
    # Optional exact toolkit regression tests are additive and kept outside the
    # app's package discovery. Normal builds ignore the _test.go overlay entry.
    test_source = COMPAT / "orbit_composer_test.go.txt"
    if test_source.exists():
        if sha(test_source) != manifest.get("tests_sha256"):
            fail("Composer compatibility test hash mismatch")
    with tempfile.TemporaryDirectory(prefix="orbit-mygo-overlay-") as tmp:
        temp_root = Path(tmp)
        # Go 1.27 forbids overlays beneath GOMODCACHE. Isolate an unchanged,
        # verified source copy rather than relaxing/changing cache permissions.
        source_copy = temp_root / "mygo"
        shutil.copytree(module_dir, source_copy)
        for relative, replacement in selection_overlays.items():
            overlay[str(source_copy / relative)] = str(replacement)
        overlay[str(source_copy / manifest["added_file"])] = str(source)
        if test_source.exists():
            overlay[str(source_copy / "ui/orbit_composer_test.go")] = str(test_source)
        temp_mod = temp_root / "orbit.mod"
        shutil.copyfile(main_mod, temp_mod)
        main_sum = main_mod.with_suffix(".sum")
        if main_sum.exists():
            shutil.copyfile(main_sum, temp_mod.with_suffix(".sum"))
        mod_json = json.loads(check_output(["mod", "edit", "-json"]))
        edits = []
        for replacement in mod_json.get("Replace", []) or []:
            old, new = replacement["Old"], replacement["New"]
            if not new.get("Version"):
                old_id = old["Path"] + ("@" + old["Version"] if old.get("Version") else "")
                local = (main_mod.parent / new["Path"]).resolve()
                edits.append(f"-replace={old_id}={local}")
        edits.append(f"-replace={manifest['module']}={source_copy}")
        subprocess.run([go, "mod", "edit", f"-modfile={temp_mod}", *edits], env=env, cwd=effective_cwd, check=True)
        path = temp_root / "overlay.json"
        path.write_text(json.dumps({"Replace": overlay}, indent=2) + "\n")
        env["GOFLAGS"] = " ".join(quote_flag(f) for f in flags + [f"-modfile={temp_mod}", f"-overlay={path}"])
        print(f"go-native: verified {manifest['version']}; temporary source copy + composer range-edit and selection-retention overlays", file=sys.stderr)
        process = subprocess.Popen([go, *args], env=env)
        def forward(signum, _frame):
            if process.poll() is None:
                process.send_signal(signum)
        old = {}
        for sig in (signal.SIGINT, signal.SIGTERM):
            old[sig] = signal.signal(sig, forward)
        try:
            status = process.wait()
        finally:
            for sig, handler in old.items():
                signal.signal(sig, handler)
        return status if status >= 0 else 128 - status


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (RuntimeError, OSError, subprocess.CalledProcessError, json.JSONDecodeError) as exc:
        print(f"go-native: {exc}", file=sys.stderr)
        sys.exit(2)
