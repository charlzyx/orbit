#!/usr/bin/env python3
"""Display-free candidate/observer invariance. Never a native-app pixel claim."""
from __future__ import annotations
import argparse
import importlib.util
import json
import os
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('specimens', ROOT.parent / 'mygo-compat/text-metrics/run-specimens.py')
reference = importlib.util.module_from_spec(spec)
spec.loader.exec_module(reference)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    for name in ('candidate-binary', 'observed-binary', 'assets', 'default-reference', 'enabled-reference', 'output'):
        p.add_argument('--' + name, required=True, type=Path)
    a = p.parse_args()
    refs = {'default': a.default_reference.resolve(), 'fractional': a.enabled_reference.resolve()}
    for mode, path in refs.items():
        if reference.sha(path) != reference.EXPECTED[mode]: p.error('unpinned reference: ' + mode)
        reference.check_roster(json.loads(path.read_text()))
    out = a.output.resolve(); out.mkdir()
    env = os.environ.copy()
    for key in ('DISPLAY', 'WAYLAND_DISPLAY', 'ORBIT_FONT_TRACE_DIR', 'ORBIT_FONT_TRACE_TEXTS', 'ORBIT_FONT_TRACE_CASE',
                'ORBIT_METRIC_EXPERIMENT', 'ORBIT_FONT_SPECIMEN_RESULT', 'ORBIT_COMPAT_SPECIMEN_ENABLE'):
        env.pop(key, None)
    env.update(LANG='C.UTF-8', LC_ALL='C.UTF-8', TZ='UTC', ORBIT_FONT_TRACE_ASSETS=str(a.assets.resolve()))
    manifest = {'schema':1, 'diagnostic':True, 'scope':'display-free native text engines; NO app paint or native present',
                'native_pixel_parity':'not evaluated', 'assets':{str(f.relative_to(a.assets)):reference.sha(f) for f in sorted(a.assets.rglob('*')) if f.is_file()},
                'binaries':{k:reference.sha(v) for k,v in [('candidate',a.candidate_binary),('observed',a.observed_binary)]}, 'runs':{}}
    cases = [('candidate', a.candidate_binary, False, '^TestOrbitTextMetricsSpecimen$'),
             ('observed_trace_off', a.observed_binary, False, '^TestOrbitFontSpecimen$'),
             ('observed_trace_on', a.observed_binary, True, '^TestOrbitFontSpecimen$')]
    for label, binary, traced, test in cases:
        for mode, path in refs.items():
            folder = out / (label + '-' + mode)
            home = folder / 'home'; home.mkdir(parents=True)
            trace = folder / 'trace'; trace.mkdir()
            current = dict(env, HOME=str(home), XDG_CACHE_HOME=str(home/'cache'), ORBIT_FONT_SPECIMEN_RESULT=str(folder/'result.json'))
            if mode == 'fractional': current['ORBIT_COMPAT_SPECIMEN_ENABLE'] = '1'
            if traced: current.update(ORBIT_FONT_TRACE_DIR=str(trace), ORBIT_FONT_TRACE_TEXTS=str(ROOT/'fixture-strings.json'), ORBIT_FONT_TRACE_CASE=folder.name)
            cmd = [str(binary.resolve()), '-test.run=' + test, '-test.v']
            run = subprocess.run(cmd, env=current, capture_output=True, text=True)
            (folder/'run.log').write_text(run.stdout+run.stderr)
            if run.returncode: raise SystemExit(run.stdout+run.stderr)
            result = folder/'result.json'; reference.check_roster(json.loads(result.read_text()))
            if result.read_bytes() != path.read_bytes(): raise SystemExit('exact specimen mismatch: '+folder.name)
            row = {'command':cmd,'exit_code':0,'cases':48,'sha256':reference.sha(result),'reference_sha256':reference.sha(path),'byte_identical':True,'trace_enabled':traced}
            if traced:
                batches = [json.loads(f.read_text()) for f in trace.glob('mygo-*.json')]
                if not batches or any(b['dropped_events'] != 0 or 'NO native present' not in b['presentation_semantics'] for b in batches):
                    raise SystemExit('missing, dropped or incorrectly labelled specimen trace: '+folder.name)
                shapes = [e for b in batches for e in b['events'] if e['kind']=='shape_run']
                if len(shapes) < 48: raise SystemExit('incomplete observed specimen shapes: '+folder.name)
                expected_round, expected_metrics = mode == 'default', 2 if mode == 'default' else 1
                if any(e.get('round_positions') is not expected_round or e.get('context_hint_metrics') != expected_metrics for e in shapes):
                    raise SystemExit('actual native context policy mismatch: '+folder.name)
                row.update(observed_shape_runs=len(shapes), actual_round_positions=expected_round, actual_context_hint_metrics=expected_metrics,
                           native_policy_getters_verified=True, trace_files={f.name:reference.sha(f) for f in trace.glob('mygo-*.json')})
            elif any(trace.iterdir()): raise SystemExit('unexpected trace with disabled configuration')
            manifest['runs'][folder.name] = row
            (out/'composed-specimen-equivalence.json').write_text(json.dumps(manifest,indent=2)+'\n')
            print(folder.name+': 48 cases exactly equal to the pinned reference',flush=True)

if __name__ == '__main__': main()
