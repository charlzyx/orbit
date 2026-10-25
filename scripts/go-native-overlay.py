#!/usr/bin/env python3
"""Hash-verified, temporary Go overlay for Orbit's pinned MyGo compatibility.

No source/module-cache writes. Forward every Go argument and ambient build env.
A caller overlay is merged only if it does not replace verified dependencies.
"""
from __future__ import annotations
import hashlib
import importlib.util
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


def default_trimpath(flags: list[str], arguments: list[str]) -> list[str]:
    # Go's build action hashes absolute source directories unless -trimpath is
    # set. Our verified temporary dependency copy gets a fresh directory each
    # invocation; embedding it would defeat dependency-cache reuse and retain
    # paths to already-deleted files in debug information. Preserve an explicit
    # caller choice (including -trimpath=false for local debugging).
    if any(f == "-trimpath" or f.startswith("-trimpath=") for f in flags + arguments):
        return list(flags)
    return [*flags, "-trimpath"]


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



DIAGNOSTIC_MODES = frozenset({"--diagnostic-text-metrics", "--diagnostic-text-metrics-observed"})
DIAGNOSTIC_TAG = "orbit_diagnostic_fractional_text_linux"


def diagnostic_mode(arguments: list[str]) -> tuple[str | None, list[str]]:
    args = list(arguments)
    mode = args.pop(0) if args and args[0] in DIAGNOSTIC_MODES else None
    if any(arg.startswith("--diagnostic-") for arg in args):
        fail("Use exactly one recognized diagnostic mode before the Go arguments")
    if not args:
        fail("A Go command is required after the diagnostic mode")
    return mode, args


def diagnostic_flags(flags: list[str], arguments: list[str], mode: str | None) -> tuple[list[str], list[str]]:
    # Preserve Go's last command-line -tags override of GOFLAGS, rather than
    # silently dropping caller tags or permitting them to drop our startup tag.
    flags, arguments = list(flags), list(arguments)
    locations = []
    for tokens in (flags, arguments):
        for index, token in enumerate(tokens):
            if token == "-tags":
                if index + 1 == len(tokens):
                    fail("-tags requires a value")
                locations.append((tokens, index + 1, "", tokens[index + 1]))
            elif token.startswith("-tags="):
                locations.append((tokens, index, "-tags=", token.partition("=")[2]))
    for _, _, _, value in locations:
        if DIAGNOSTIC_TAG in value.replace(",", " ").split():
            fail("The diagnostic startup build tag is owned by the explicit wrapper mode")
    if mode:
        if mode not in DIAGNOSTIC_MODES:
            fail("Unknown diagnostic mode")
        if locations:
            tokens, index, prefix, value = locations[-1]
            tokens[index] = prefix + value + ("," if value else "") + DIAGNOSTIC_TAG
        else:
            flags.append("-tags=" + DIAGNOSTIC_TAG)
    return flags, arguments


def merge_reviewed(mappings: list[dict[str, Path]]) -> dict[str, Path]:
    out = {}
    for mapping in mappings:
        for relative, source in mapping.items():
            path = Path(relative)
            if path.is_absolute() or ".." in path.parts or relative in out:
                fail("Unreviewed compatibility path/collision: " + relative)
            out[relative] = source
    return out


def import_verifier(path: Path):
    spec = importlib.util.spec_from_file_location("orbit_" + path.parent.name + "_" + path.stem, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def verify_diagnostic(module_dir: Path, mode: str | None) -> dict[str, Path]:
    if mode is None:
        return {}  # Normal builds never load or apply candidate/observer assets.
    if mode not in DIAGNOSTIC_MODES:
        fail("Unknown diagnostic mode")
    candidate_root = ROOT / "scripts/mygo-compat/text-metrics"
    _, mapping = import_verifier(candidate_root / "prepare.py").verify(module_dir)
    if mode == "--diagnostic-text-metrics-observed":
        observer = ROOT / "scripts/font-fidelity/combined-verify.py"
        mapping = import_verifier(observer).verify(module_dir, mapping, candidate_root / "baseline.json")
    return mapping


def verify_review_text(module_dir: Path, expected: dict, mode: str | None, assets: Path | None = None) -> dict[str, Path]:
    if mode:
        fail("Global text-metrics diagnostics conflict with local Review text policy")
    assets = assets or ROOT / "third_party/mygo-review-text"
    if sha(assets / "baseline.json") != "9b8d49dcb3f95af6aef6e2e2d34fe7e6afcb3f2f16e8b8353cc1cef39dadd86c":
        fail("Review text manifest differs from reviewed hash")
    verifier = import_verifier(ROOT / "third_party/mygo-review-text/prepare.py")
    manifest, mappings = verifier.verify(module_dir, assets)
    if any(manifest[key] != expected[key] for key in SELECTION_IDENTITIES):
        fail("Review text compatibility identity differs from pinned MyGo")
    return mappings


def main() -> int:
    if len(sys.argv) < 2:
        fail("Usage: scripts/go-native.sh <run|build|test|vet|other-go-command> [arguments...]")
    go = os.environ.get("ORBIT_GO") or shutil.which("go")
    if not go:
        fail("Go 1.27.3 is required; set ORBIT_GO to its go executable")
    env = os.environ.copy()
    mode, args = diagnostic_mode(sys.argv[1:])
    effective_cwd = Path.cwd()
    if args and args[0] == "-C":
        if len(args) < 3:
            fail("go -C requires a directory and command")
        effective_cwd = (effective_cwd / args[1]).resolve()
    elif args and args[0].startswith("-C="):
        effective_cwd = (effective_cwd / args[0][3:]).resolve()
    # Command-line -overlay would supersede our mandatory overlay. Explicitly
    # reject it, rather than silently building the unsafe fallback.
    for arg in args[1:]:
        if arg == "-overlay" or arg.startswith("-overlay="):
            fail("Pass a custom -overlay through GOFLAGS; the native wrapper merges it safely")
    inherited = env.get("GOFLAGS")
    if inherited is None:
        inherited = subprocess.check_output([go, "env", "GOFLAGS"], text=True).strip()
    flags, previous_overlay = detach_overlay(split_flags(inherited))
    flags = default_trimpath(flags, args)
    flags, args = diagnostic_flags(flags, args, mode)
    if any(f == "-modfile" or f.startswith("-modfile=") for f in flags + args):
        fail("The native wrapper owns its temporary -modfile; remove the conflicting flag")
    env["GOFLAGS"] = " ".join(quote_flag(f) for f in flags)
    def check_output(argv):
        return subprocess.check_output([go, *argv], env=env, text=True, cwd=effective_cwd)
    selected = json.loads(check_output(["env", "-json", "GOMOD", "GOWORK", "GOOS", "GOARCH"]))
    if mode and (selected.get("GOOS") != "linux" or selected.get("GOARCH") not in ("amd64", "arm64")):
        fail("DIAGNOSTIC fractional text metrics require Linux amd64 or arm64; no fallback")
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
    review_overlays = verify_review_text(module_dir, manifest, mode)
    diagnostic_overlays = {}
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
    reviewed_overlays = merge_reviewed([
        selection_overlays, {manifest["added_file"]: source},
        {"ui/orbit_composer_test.go": test_source} if test_source.exists() else {},
        review_overlays, diagnostic_overlays,
    ])
    with tempfile.TemporaryDirectory(prefix="orbit-mygo-overlay-") as tmp:
        temp_root = Path(tmp)
        # Go 1.27 forbids overlays beneath GOMODCACHE. Isolate an unchanged,
        # verified source copy rather than relaxing/changing cache permissions.
        source_copy = temp_root / "mygo"
        shutil.copytree(module_dir, source_copy)
        for relative, replacement in reviewed_overlays.items():
            if str(source_copy / relative) in overlay:
                fail("Caller overlay collides with a verified compatibility source")
            overlay[str(source_copy / relative)] = str(replacement)
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
        print(f"go-native: verified {manifest['version']}; temporary source copy + composer range-edit, selection-retention and local Review text overlays", file=sys.stderr)
        if mode:
            print(f"go-native: DIAGNOSTIC ONLY {mode}; startup fractional metrics enabled; not a parity claim", file=sys.stderr)
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
    except (RuntimeError, ValueError, OSError, subprocess.CalledProcessError, json.JSONDecodeError) as exc:
        print(f"go-native: {exc}", file=sys.stderr)
        sys.exit(2)
