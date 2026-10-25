#!/usr/bin/env python3
"""Verify one explicit candidate + observer composition, without writing sources."""
from __future__ import annotations
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
PIN = '9c69cf829ece196c5b3a3601f1c7bbe70526376f'
BASELINE = frozenset({'internal/text/pango_linux.go', 'ui/paint.go', 'ui/runtime.go'})
INPUTS = frozenset({'prepare.py', 'mygo_trace.go.txt', 'mygo_pango_trace.go.txt', 'mygo_specimen_test.go.txt'})
FILES = {
    'internal/text/pango_linux.go': 'combined_pango_linux.go.txt',
    'ui/paint.go': 'combined_paint.go.txt',
    'ui/runtime.go': 'combined_runtime.go.txt',
    'internal/text/orbit_mygo_pango_trace.go': 'combined_pango_trace.go.txt',
    'internal/text/orbit_mygo_trace.go': 'mygo_trace.go.txt',
    'internal/text/orbit_mygo_specimen_test.go': 'combined_specimen_test.go.txt',
}
CANDIDATE_FILES = frozenset({
    'internal/text/pango_linux.go', 'internal/text/orbit_text_metrics_linux.go',
    'internal/text/orbit_text_metrics_linux_test.go', 'internal/text/orbit_text_metrics_specimen_linux_test.go',
    'ui/orbit_text_metrics_linux.go', 'ui/orbit_text_metrics_other.go',
})


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify(source: Path, candidate: dict[str, Path], candidate_manifest: Path, assets: Path = ROOT) -> dict[str, Path]:
    m = json.loads((assets / 'combined-baseline.json').read_text())
    keys = {'schema', 'diagnostic', 'upstream_commit', 'candidate_manifest_sha256',
            'observer_inputs_sha256', 'baseline_sha256', 'intentional_collisions', 'overlays', 'patch_sha256'}
    if (set(m) != keys or m['schema'] != 1 or m['diagnostic'] is not True or m['upstream_commit'] != PIN):
        raise ValueError('unexpected combined diagnostic identity/schema')
    if set(candidate) != CANDIDATE_FILES or sha(candidate_manifest) != m['candidate_manifest_sha256']:
        raise ValueError('combined diagnostic requires the exact reviewed candidate')
    candidate_info = json.loads(candidate_manifest.read_text())
    for relative, path in candidate.items():
        item = candidate_info['overlays'][relative]
        expected = candidate_manifest.parent / item['source']
        if path.resolve() != expected.resolve() or sha(path) != item['sha256']:
            raise ValueError('combined diagnostic candidate mapping substitution: ' + relative)
    if set(m['baseline_sha256']) != BASELINE or set(m['overlays']) != set(FILES):
        raise ValueError('unexpected combined diagnostic paths')
    if set(m['observer_inputs_sha256']) != INPUTS:
        raise ValueError('unexpected combined diagnostic observer inputs')
    if m['intentional_collisions'] != ['internal/text/pango_linux.go']:
        raise ValueError('unexpected combined diagnostic collision permission')
    for relative, digest in m['baseline_sha256'].items():
        if not (source / relative).is_file() or sha(source / relative) != digest:
            raise ValueError('combined diagnostic upstream drift: ' + relative)
    for filename, digest in m['observer_inputs_sha256'].items():
        if (assets / filename).resolve().parent != assets.resolve() or sha(assets / filename) != digest:
            raise ValueError('combined diagnostic observer input drift: ' + filename)
    if sha(assets / 'combined-observer.patch') != m['patch_sha256']:
        raise ValueError('combined diagnostic patch hash mismatch')
    out = dict(candidate)
    for relative, filename in FILES.items():
        item = m['overlays'][relative]
        path = assets / filename
        if (set(item) != {'source', 'sha256'} or item['source'] != filename or
                path.resolve().parent != assets.resolve() or not path.is_file() or sha(path) != item['sha256']):
            raise ValueError('combined diagnostic overlay path/hash mismatch: ' + relative)
        if relative not in BASELINE and (source / relative).exists():
            raise ValueError('combined diagnostic additive source already exists: ' + relative)
        if relative in out and relative != 'internal/text/pango_linux.go':
            raise ValueError('unreviewed diagnostic collision: ' + relative)
        out[relative] = path
    return out
