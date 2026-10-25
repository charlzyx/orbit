#!/usr/bin/env python3
"""Build a separately identified Linux diagnostic binary and immutable provenance."""
from __future__ import annotations
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

ROOT = Path(__file__).resolve().parents[2]
MODES = {'fractional':'--diagnostic-text-metrics', 'fractional-observed':'--diagnostic-text-metrics-observed'}


def sha(path): return hashlib.sha256(path.read_bytes()).hexdigest()


def files(root):
    return {str(p.relative_to(root)):sha(p) for p in sorted(root.rglob('*'))
            if p.is_file() and '.git' not in p.relative_to(root).parts and '__pycache__' not in p.parts}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--diagnostic', action='store_true', required=True)
    parser.add_argument('--mode', choices=MODES, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--base-git-tree', help='Verified base Git tree, if this is a diagnostic export')
    args = parser.parse_args()
    out = args.output.resolve()
    if out.exists() or out.is_relative_to(ROOT) or ROOT.is_relative_to(out): parser.error('new output disjoint from source required')
    go = os.environ.get('ORBIT_GO') or shutil.which('go')
    if not go: parser.error('Go 1.27.3 is required')
    env = dict(os.environ, CGO_ENABLED='0', GOWORK='off')
    go_version = subprocess.check_output([go, 'version'], text=True, env=env).strip()
    if not go_version.startswith('go version go1.27.3 '): parser.error('exact Go 1.27.3 required')
    target = json.loads(subprocess.check_output([go, 'env', '-json', 'GOOS', 'GOARCH'], env=env, text=True))
    if target['GOOS'] != 'linux' or target['GOARCH'] not in ('amd64','arm64'): parser.error('diagnostic target must be Linux amd64/arm64')
    before = files(ROOT)
    out.mkdir()
    binary = out / ('orbit-go-' + args.mode + '-diagnostic')
    cmd = [str(ROOT/'scripts/go-native.sh'), MODES[args.mode], 'build', '-buildvcs=false', '-p=2', '-o', str(binary), './cmd/orbit']
    run = subprocess.run(cmd, cwd=ROOT, env=env, capture_output=True, text=True)
    (out/'build.log').write_text(run.stdout+run.stderr)
    if run.returncode: raise SystemExit(run.stdout+run.stderr)
    if before != files(ROOT): raise SystemExit('source changed during diagnostic build; refuse provenance')
    info = subprocess.check_output([go, 'version', '-m', str(binary)], text=True, env=env)
    (out/'go-build-info.txt').write_text(info)
    manifest = {'schema':1, 'diagnostic':True, 'production':False, 'native_pixel_parity':'not evaluated',
                'mode':args.mode, 'wrapper_mode':MODES[args.mode], 'startup_build_tag':'orbit_diagnostic_fractional_text_linux',
                'startup_policy':{'pango_round_positions':False,'cairo_hint_metrics':1,'configured_before':'LoadAssets'},
                'observations':'native Pango/Cairo getters, not policy assumptions' if args.mode.endswith('observed') else 'not compiled',
                'base_git_tree':args.base_git_tree, 'source_tree_sha256':hashlib.sha256(json.dumps(before,sort_keys=True,separators=(',',':')).encode()).hexdigest(),
                'source_files_sha256':before, 'binary':binary.name, 'binary_sha256':sha(binary), 'binary_bytes':binary.stat().st_size,
                'go_version':go_version, 'target':target, 'CGO_ENABLED':'0', 'GOWORK':'off', 'GOFLAGS':env.get('GOFLAGS',''),
                'command':cmd, 'build_log_sha256':sha(out/'build.log'), 'go_build_info_sha256':sha(out/'go-build-info.txt'),
                'mygo_pin':'9c69cf829ece196c5b3a3601f1c7bbe70526376f',
                'candidate_manifest_sha256':sha(ROOT/'scripts/mygo-compat/text-metrics/baseline.json'),
                'combined_manifest_sha256':sha(ROOT/'scripts/font-fidelity/combined-baseline.json') if args.mode.endswith('observed') else None,
                'hosted_capture':'not executed; use capture-diagnostic.sh with this provenance'}
    (out/'build-provenance.json').write_text(json.dumps(manifest,indent=2)+'\n')
    print(out/'build-provenance.json')

if __name__ == '__main__': main()
