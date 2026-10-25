#!/usr/bin/env python3
"""Fresh synthetic Pi v3 input for two static Chat states; Context blocked."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re

ROOT = Path('/tmp/orbit-chat-static-fixture')
WORK = ROOT / 'workspace'
SESSION_ID = '00000000-0000-4000-8000-000000001803'
STAMP = '2026-10-01T12:00:00.000Z'
EPOCH = 1790856000000
HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('workbench_materialize', HERE.parent / 'workbench-native/materialize.py')
wb = importlib.util.module_from_spec(spec); spec.loader.exec_module(wb)

def pretty(value): return json.dumps(value, indent=2, sort_keys=True) + '\n'
def digest(value): return hashlib.sha256(value).hexdigest()

def home_files():
    entries = [dict(type='session', version=3, id=SESSION_ID, timestamp=STAMP, cwd=str(WORK))]
    parent = None
    def append(kind, **fields):
        nonlocal parent
        ident = f'chat{len(entries):04d}'
        entries.append(dict(type=kind, id=ident, parentId=parent, timestamp=STAMP, **fields)); parent = ident
    append('model_change', provider='anthropic', modelId='claude-sonnet-4-6')
    append('thinking_level_change', thinkingLevel='off')
    for i in range(1, 13):
        marker = 'PROMPT ELEVEN' if i == 11 else f'PROMPT NUMBER {i:02d}'
        append('message', message=dict(role='user', content=[dict(type='text', text=marker)], timestamp=EPOCH))
        text = f'RESPONSE NUMBER {i:02d}\n\n' + '\n\n'.join(
            f'Paragraph {n:02d}. ' + 'This is persisted synthetic transcript text, not a model response. ' * 5
            for n in range(1, 13))
        text += '\n\n' + ('CHAT TAIL READY' if i == 12 else f'RESPONSE END {i:02d}')
        final = i == 12
        append('message', message=dict(role='assistant', api='anthropic-messages', provider='anthropic', model='claude-sonnet-4-6',
            content=[dict(type='text', text=text)], usage=dict(input=380000 if final else 900, output=20000 if final else 100,
            cacheRead=0, cacheWrite=0, totalTokens=400000 if final else 1000,
            cost=dict(input=0, output=0, cacheRead=0, cacheWrite=0, total=0)), stopReason='stop', timestamp=EPOCH))
    append('session_info', name='Static Chat fixture')
    encoded = '--' + re.sub(r'[/\\:]', '-', str(WORK).lstrip('/\\')) + '--'
    session = f'.pi/agent/sessions/{encoded}/{SESSION_ID}.jsonl'
    return {
        session: ''.join(json.dumps(e, separators=(',', ':')) + '\n' for e in entries),
        '.pi/agent/settings.json': pretty(dict(enableInstallTelemetry=False, enableAnalytics=False, defaultProvider='anthropic', defaultModel='claude-sonnet-4-6', defaultThinkingLevel='off')),
        '.orbit-pi/auto-title.json': pretty(dict(enabled=False, model=None)),
        '.orbit-pi/workspaces.json': pretty(dict(workspaces=[str(WORK)], group_by='one_list')),
        '.orbit-pi/ui.json': pretty(dict(language='en', ui_font_size=14, terminal_font_size=13, editor_font_size=13, spacing_density=100, reduce_motion=True)),
        '.orbit-pi/theme': pretty(dict(mode='dark', dark_theme='orbit-azure', light_theme='orbit-azure-light')),
        '.orbit-pi/fonts.json': pretty(dict(ui_font_family='.ZedSans', code_font_family='.ZedMono')),
    }, session

def materialize(root, manifest):
    root, manifest = Path(root), Path(manifest)
    if os.path.lexists(root) or os.path.lexists(manifest):
        raise ValueError('Refusing preexisting fixture root or manifest')
    if any(p.is_symlink() for p in (*root.parents, *manifest.parents)):
        raise ValueError('Refusing symlinked fixture/manifest parents')
    root.mkdir(mode=0o700); (root/'workspace').mkdir()
    files, session = home_files()
    for relative, content in files.items():
        target = root/'home'/relative; target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content); target.chmod(0o600)
    value = dict(schema=1, fixture='static-chat-v3', logicalWorkspace=str(WORK), session=session,
        userTurns=12, lastAssistantRecordedTokens=400000, expectedContextUsage='absent without configured provider auth', expectedModel='unknown',
        expectedSessionTokens=411000, files={f'home/{p}':digest(s.encode()) for p,s in sorted(files.items())},
        allowedUserInputs=['ctrl+1','ctrl+Up'], inferenceRequested=False)
    manifest.parent.mkdir(parents=True, exist_ok=True); manifest.write_text(pretty(value)); return value

def verify(root, expected):
    root=Path(root); files, session=home_files()
    if expected['files'] != {f'home/{p}':digest(s.encode()) for p,s in sorted(files.items())}:
        raise ValueError('Fixture oracle differs from generator')
    differences=[]
    for name, original in files.items():
        target=root/'home'/name
        if target.is_symlink() or not target.is_file(): raise ValueError('Owned fixture input missing or symlinked: '+name)
        if name != session and target.read_bytes()!=original.encode(): differences.append(name)
    target=root/'home'/session
    with target.open('rb') as stream: actual=stream.read(wb.SESSION_EVIDENCE_LIMIT+1)
    proof=wb.session_integrity(files[session].encode(),actual)
    proof['reason']=proof.get('reason','').replace('Original six entries','Original fixed entries')
    auth=root/'home/.pi/agent/auth.json'
    try: auth_empty=not os.path.lexists(auth) or (not auth.is_symlink() and json.loads(auth.read_text())=={})
    except (ValueError,OSError):auth_empty=False
    return dict(integrityAccepted=not differences and proof['accepted'] and auth_empty, credentialFileAbsentOrEmpty=auth_empty, changedOwnedInputs=differences, session=proof,
        rule='Fixed inputs byte-identical; session allows only existing validated one-entry startup workflow metadata append')

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('root',type=Path);p.add_argument('--manifest',required=True,type=Path);p.add_argument('--verify',type=Path)
    args=p.parse_args()
    if args.verify:
        if os.path.lexists(args.manifest): raise ValueError('Refusing existing verification output')
        value=verify(args.root,json.loads(args.verify.read_text()));args.manifest.write_text(pretty(value))
        if not value['integrityAccepted']: raise SystemExit('Static Chat fixture integrity failed')
    else: materialize(args.root,args.manifest)
if __name__=='__main__':main()
