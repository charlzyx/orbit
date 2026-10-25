#!/usr/bin/env python3
"""Record capture readiness and input provenance, never a visual parity PASS."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def build(out, kind, source, binary, harness, head, dirty, commit, node_binary):
    out, source, binary, harness = map(Path, (out, source, binary, harness))
    recipe_path = harness / 'testdata/workbench-native/cases.json'
    recipe = json.loads(recipe_path.read_text())
    result = json.loads((out / 'capture-result.json').read_text())
    fixture = json.loads((out / 'fixture-before.json').read_text())
    after = json.loads((out / 'fixture-after.json').read_text())
    fonts = source / 'crates/orbit-pi/assets/fonts'
    images = {case['image']: case['proof']['image'] for case in result['cases']}
    integrity = []
    if len(result['cases']) != len(recipe['cases']):
        integrity.append('Not all eight requested cases were captured')
    for case in result['cases']:
        path = out / case['image']
        if not path.is_file() or sha(path) != case['proof']['pngSha256']:
            integrity.append('Missing or modified screenshot: ' + case['image'])
        if case['proof']['image'] != recipe['viewport'] or case['proof']['client'] != recipe['viewport']:
            integrity.append('Image/client geometry mismatch: ' + case['image'])
    gaps = ['Actual selected states and loaded content require manual image review', 'Resolved glyph faces unverified', 'Clock/animation/caret phases not frozen']
    if not after['unchanged']:
        gaps.append('Owned fixture files, Git commits/status/diffs changed during capture')
    manifest = dict(source=commit, application=kind, sourceRootGitHead=head, sourceRootDirty=dirty == 'true',
                    sourceCleanlinessRecorded='Before the original app launch, including untracked source files',
                    binarySha256=sha(binary), runtimeNode=dict(path=node_binary, version=subprocess.check_output([node_binary, '--version'], text=True).strip(), sha256=sha(Path(node_binary))), harnessGitHead=subprocess.check_output(['git', '-C', str(harness), 'rev-parse', 'HEAD'], text=True).strip(),
                    suite=recipe['suite'], recipeSha256=sha(recipe_path), recipe=recipe,
                    requestedViewport=recipe['viewport'], actualImageSizes=images,
                    palette='orbit-azure', language='en', fontSize=14, density=100, reduceMotion=True,
                    state='isolated fixed synthetic workbench Git and Pi v3 session, no credentials or inference',
                    display='Xvfb/Openbox Linux software graphics', dpi=96, uiScale=1, timezone='UTC',
                    clockFrozen=False, animationPhaseFrozen=False, focusAndCaretPhaseVerified=False,
                    requestedFontPreferences=recipe['fontPreferences'], resolvedFontFaces=None,
                    bundledFontInputs={str(p.relative_to(fonts)): sha(p) for p in sorted(fonts.rglob('*')) if p.suffix.lower() in ('.ttf', '.otf')},
                    fixtureSha256=sha(out / 'fixture-before.json'), fixture=fixture,
                    fixturePreflightVerified=True, fixtureUnchangedAfterCapture=after['unchanged'],
                    captureValidity=('BLOCKED' if result['error'] or integrity or not after['unchanged'] else 'READY_FOR_MANUAL_REVIEW'),
                    captureIntegrityGaps=integrity,
                    semanticState='UNREVIEWED', pixelEquality='NOT_COMPARED', captureResult=result,
                    normalizationGaps=gaps, volatileCases={case['image']: case['volatile'] for case in recipe['cases']},
                    globalVisualIds=None, note='WB IDs are local recipes, not global coverage. Full native client pixels only; no resizing, crop alignment, masking, or tolerance. Capture readiness is independent from pixel equality.')
    (out / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
    return manifest


if __name__ == '__main__':
    build(*sys.argv[1:])
