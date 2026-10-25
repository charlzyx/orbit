#!/usr/bin/env python3
"""Record two captures and the Context blocker, never a parity verdict."""
import json
from pathlib import Path
import subprocess
import sys
import materialize

def build(out,kind,source,binary,harness,head,dirty,commit,node_binary):
    out,source,binary,harness=map(Path,(out,source,binary,harness))
    suite=json.loads((harness/'testdata/chat-native/cases.json').read_text())
    result=json.loads((out/'capture-result.json').read_text())
    before=json.loads((out/'fixture-before.json').read_text());after=json.loads((out/'fixture-after.json').read_text())
    images={c['image']:c['proof']['image']for c in result['cases']};errors=[]
    if len(images)!=2:errors.append('Not all two approved static cases captured')
    if any(v!=[1470,922]for v in images.values()):errors.append('Native client image dimensions differ')
    if not after['integrityAccepted']:errors.append('Fixture integrity or empty-credential condition failed')
    if subprocess.check_output(['git','-C',str(source),'rev-parse','HEAD'],text=True).strip()!=head:errors.append('Source HEAD changed since capture preflight')
    value=dict(schema=1,suite=suite,application=kind,source=commit,sourceRootGitHead=head,sourceRootDirty=dirty=='true',
        binarySha256=materialize.digest(binary.read_bytes()),harnessGitHead=subprocess.check_output(['git','-C',str(harness),'rev-parse','HEAD'],text=True).strip(),
        runtimeNode=dict(path=node_binary,version=subprocess.check_output([node_binary,'--version'],text=True).strip()),
        fixture=before,fixtureIntegrity=after,preflight=json.loads((out/'pi-preflight.json').read_text()),
        requestedCaptureCases=2,capturedCases=len(images),actualImageSizes=images,blockedCases=suite['blockedCases'],
        semanticState='UNREVIEWED',pixelEquality='NOT_COMPARED',captureValidity='BLOCKED'if errors else 'READY_FOR_MANUAL_REVIEW',errors=errors,
        platform='hosted Linux Xvfb/Openbox',viewport=[1470,922],dpi=96,uiScale=1,palette='orbit-azure',language='en',fontSize=14,density=100,
        fonts=dict(ui='.ZedSans',code='.ZedMono',resolvedGlyphFaces='UNVERIFIED'),clockFrozen=False,animationPhaseFrozen=False,
        selectionHoverRPCStatusCoverage='NOT_AUTHORIZED_NOT_CAPTURED',contextCoverage='BLOCKED_NO_AUTH_MODEL_UNKNOWN',
        rpcChildEvidence='Observed pinned Pi RPC descendants; not proof of provider inference or request acknowledgement',
        actions=json.loads((out/'actions.json').read_text())if(out/'actions.json').exists()else [])
    (out/'manifest.json').write_text(materialize.pretty(value));return value
if __name__=='__main__':
    value=build(*sys.argv[1:]);raise SystemExit(0 if not value['errors']else 1)
