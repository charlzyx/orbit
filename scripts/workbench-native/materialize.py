#!/usr/bin/env python3
"""Create a new, local-only Git workspace and Pi v3 session, never user data."""
import argparse
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess

ROOT = Path('/tmp/orbit-workbench-fixture')
WORK = ROOT / 'workspace'
SESSION_ID = '01900000-0000-7000-8000-000000000008'
STAMP = '2026-10-01T12:00:00.000Z'
EPOCH = 1790856000
BASE_FILES = {
    'alpha.go': 'package fixture\n\nfunc fixtureAlpha() string {\n\treturn "base"\n}\n',
    'bravo.go': 'package fixture\n\nfunc fixtureBravo() string {\n\treturn "base"\n}\n',
    'docs/plan.md': '# Workbench plan\n\nA small local capture fixture.\n',
    'src/helper.go': 'package fixture\n\nconst StableHelper = 1\n',
}
FINAL_FILES = dict(BASE_FILES, **{
    'alpha.go': BASE_FILES['alpha.go'].replace('"base"', '"staged"'),
    'bravo.go': BASE_FILES['bravo.go'].replace('"base"', '"unstaged"'),
    'docs/plan.md': BASE_FILES['docs/plan.md'] + '\nHistory has two fixed commits.\n',
    'src/helper.go': BASE_FILES['src/helper.go'].replace('= 1', '= 2'),
    'notes.txt': 'Untracked fixture note.\n',
})


def digest(data):
    return hashlib.sha256(data).hexdigest()


def pretty(value):
    return json.dumps(value, indent=2, sort_keys=True) + '\n'


def home_files():
    entries = [dict(type='session', version=3, id=SESSION_ID, timestamp=STAMP, cwd=str(WORK))]
    parent = None
    def append(kind, **fields):
        nonlocal parent
        ident = f'workbench{len(entries):02d}'
        entries.append(dict(type=kind, id=ident, parentId=parent, timestamp=STAMP, **fields))
        parent = ident
    append('model_change', provider='anthropic', modelId='claude-sonnet-4-6')
    append('thinking_level_change', thinkingLevel='off')
    append('message', message=dict(role='user', content=[dict(type='text', text='Show the prepared local workbench fixture.')], timestamp=EPOCH * 1000))
    append('message', message=dict(role='assistant', api='anthropic-messages', provider='anthropic', model='claude-sonnet-4-6',
        content=[dict(type='text', text='Workbench fixture ready.\n\n- One staged change\n- One unstaged change\n- One untracked note\n\n```go\nconst CaptureMode = "local"\n```\n\nNo model request was made.')],
        usage=dict(input=40, output=30, cacheRead=0, cacheWrite=0, totalTokens=70, cost=dict(input=0, output=0, cacheRead=0, cacheWrite=0, total=0)), stopReason='stop', timestamp=EPOCH * 1000))
    append('session_info', name='Workbench fixture')
    encoded = '--' + re.sub(r'[/\\:]', '-', str(WORK).lstrip('/\\')) + '--'
    session = f'.pi/agent/sessions/{encoded}/{SESSION_ID}.jsonl'
    return {
        session: ''.join(json.dumps(item, separators=(',', ':')) + '\n' for item in entries),
        '.pi/agent/settings.json': pretty(dict(enableInstallTelemetry=False, enableAnalytics=False, defaultProvider='anthropic', defaultModel='claude-sonnet-4-6', defaultThinkingLevel='off')),
        '.orbit-pi/auto-title.json': pretty(dict(enabled=False, model=None)),
        '.orbit-pi/workspaces.json': pretty(dict(workspaces=[str(WORK)], group_by='one_list')),
        '.orbit-pi/ui.json': pretty(dict(language='en', ui_font_size=14, terminal_font_size=13, editor_font_size=13, spacing_density=100, reduce_motion=True)),
        '.orbit-pi/theme': pretty(dict(mode='dark', dark_theme='orbit-azure', light_theme='orbit-azure-light')),
        '.orbit-pi/fonts.json': pretty(dict(ui_font_family='.ZedSans', code_font_family='.ZedMono')),
        '.bash_profile': 'export PS1="orbit-fixture> "\nunset PROMPT_COMMAND\nexport HISTFILE=/dev/null\nprintf "Terminal fixture ready.\\n"\n',
    }, session


def git_env():
    return dict(PATH=os.environ['PATH'], LANG='C.UTF-8', LC_ALL='C.UTF-8', TZ='UTC',
                GIT_CONFIG_NOSYSTEM='1', GIT_CONFIG_GLOBAL=os.devnull, GIT_TERMINAL_PROMPT='0',
                GIT_AUTHOR_NAME='Orbit Fixture', GIT_AUTHOR_EMAIL='fixture@example.invalid',
                GIT_COMMITTER_NAME='Orbit Fixture', GIT_COMMITTER_EMAIL='fixture@example.invalid',
                GIT_AUTHOR_DATE='2026-10-01T12:00:00+00:00', GIT_COMMITTER_DATE='2026-10-01T12:00:00+00:00')


def git(work, *args):
    return subprocess.check_output(['git', '-c', 'core.hooksPath=/dev/null', '-c', 'commit.gpgsign=false', '-c', 'core.autocrlf=false', '-C', str(work), *args], env=git_env(), text=True)


def snapshot(root):
    work = root / 'workspace'
    owned, session = home_files()
    files = {f'home/{p}': digest((root / 'home' / p).read_bytes()) for p in sorted(owned)}
    files.update({f'workspace/{p}': digest((work / p).read_bytes()) for p in sorted(FINAL_FILES)})
    return dict(schema=1, name='workbench-native-fixed-october', credentials=False, inference=False,
                canonicalRoot=str(ROOT), workspace=str(WORK), session=f'home/{session}', sessions=1,
                files=files, branch=git(work, 'branch', '--show-current').strip(),
                commits=git(work, 'log', '--format=%H%x09%s').splitlines(),
                porcelain=git(work, 'status', '--porcelain=v1', '--untracked-files=all').splitlines(),
                stagedDiffSha256=digest(git(work, 'diff', '--cached', '--no-ext-diff', '--binary').encode()),
                unstagedDiffSha256=digest(git(work, 'diff', '--no-ext-diff', '--binary').encode()),
                remotes=git(work, 'remote').splitlines(),
                emptyDirectories=[str(path.relative_to(root)) for path in sorted(work.rglob('*'))
                                  if path.relative_to(work).parts[0] != '.git'
                                  and path.is_dir() and not path.is_symlink() and not any(path.iterdir())],
                sourceClockFrozen=False)


def materialize(root, manifest):
    # A wholly new root is mandatory. lexists also rejects dangling symlinks.
    root, manifest = Path(root), Path(manifest)
    if os.path.lexists(root) or os.path.lexists(manifest):
        raise ValueError('Refusing to reuse or overwrite a fixture root/manifest')
    if any(p.is_symlink() for p in (root.parent, *root.parents, manifest.parent, *manifest.parents)):
        raise ValueError('Refusing symlinked fixture/manifest parents')
    root.mkdir(mode=0o700)
    work = root / 'workspace'
    work.mkdir()
    owned, _ = home_files()
    for relative, content in owned.items():
        target = root / 'home' / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content)
        target.chmod(0o600)
    empty_template = root / 'empty-template'
    empty_template.mkdir()
    git(work, 'init', '--quiet', '--initial-branch=capture-main', '--object-format=sha1', '--template=' + str(empty_template))
    git(work, 'config', 'core.hooksPath', '/dev/null')
    git(work, 'config', 'core.autocrlf', 'false')
    git(work, 'config', 'commit.gpgsign', 'false')
    for relative, content in BASE_FILES.items():
        target = work / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content)
        target.chmod(0o644)
    git(work, 'add', '--', *sorted(BASE_FILES))
    git(work, 'commit', '--quiet', '-m', 'Initialize workbench fixture')
    for relative in ['docs/plan.md', 'src/helper.go']:
        (work / relative).write_text(FINAL_FILES[relative])
    git(work, 'add', '--', 'docs/plan.md', 'src/helper.go')
    git(work, 'commit', '--quiet', '-m', 'Add stable helper')
    for relative, content in FINAL_FILES.items():
        (work / relative).write_text(content)
        (work / relative).chmod(0o644)
    git(work, 'add', '--', 'alpha.go')
    (work / 'docs/empty').mkdir()
    for path in root.rglob('*'):
        if path.is_file() and '.git' not in path.parts:
            os.utime(path, (EPOCH, EPOCH))
    value = snapshot(root)
    manifest.parent.mkdir(parents=True, exist_ok=True)
    manifest.write_text(pretty(value))
    return value


SESSION_EVIDENCE_LIMIT = 1024 * 1024


def session_integrity(before, after):
    """Only the real Pi1.4 startup workflow append may vary, never transcript bytes.

    Source: contrib/orbit-workflow-extension/index.js::applyMode/session_start,
    policy.js::WORKFLOW_ENTRY_TYPE/DEFAULT_MODE, and Pi1.4 SessionManager's
    appendCustomEntry. A real metadata-only RPC regression proves the full diff.
    """
    value = dict(rawBytesEqual=before == after, originalEntryBytesPreserved=after.startswith(before),
                 semanticFixtureContentPreserved=False, accepted=False, appendedEntries=[],
                 expectedSha256=digest(before), actualSha256=digest(after),
                 rule='Unchanged bytes or exactly one additive orbit:workflow build entry; no other rewrite/append')
    if len(before) > SESSION_EVIDENCE_LIMIT or len(after) > SESSION_EVIDENCE_LIMIT:
        return dict(value, reason='Synthetic session exceeds the 1 MiB evidence limit')
    if before == after:
        return dict(value, semanticFixtureContentPreserved=True, accepted=True, reason='Byte-identical session')
    if not after.startswith(before):
        return dict(value, reason='Original fixed session bytes were rewritten or removed')
    suffix = after[len(before):]
    if not suffix.endswith(b'\n') or len(suffix.splitlines()) != 1:
        return dict(value, reason='Expected exactly one complete appended JSONL entry')
    try:
        entry = json.loads(suffix)
        original = [json.loads(line) for line in before.splitlines()]
    except (ValueError, UnicodeDecodeError):
        return dict(value, reason='Invalid session JSONL')
    value['appendedEntries'] = [entry]
    keys = {'type', 'customType', 'data', 'id', 'parentId', 'timestamp'}
    if not isinstance(entry, dict) or set(entry) != keys:
        return dict(value, reason='Unexpected appended entry fields')
    if entry['type'] != 'custom' or entry['customType'] != 'orbit:workflow' or entry['data'] != {'mode': 'build'}:
        return dict(value, reason='Unexpected appended entry content, including any message/usage/model/title change')
    if entry['parentId'] != original[-1]['id']:
        return dict(value, reason='Workflow metadata does not directly extend the original leaf')
    if not isinstance(entry['id'], str) or not re.fullmatch(r'[0-9a-f]{8}', entry['id']) or entry['id'] in {item.get('id') for item in original}:
        return dict(value, reason='Workflow entry ID is not the unique generated eight-hex shape')
    if not isinstance(entry['timestamp'], str) or not re.fullmatch(r'\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z', entry['timestamp']):
        return dict(value, reason='Workflow timestamp is not the UTC millisecond ISO shape')
    try:
        dt.datetime.fromisoformat(entry['timestamp'].replace('Z', '+00:00'))
    except ValueError:
        return dict(value, reason='Workflow timestamp is not a valid date/time')
    return dict(value, semanticFixtureContentPreserved=True, accepted=True,
                reason='Original six entries byte-preserved; one validated non-message startup workflow metadata append',
                volatileMetadataFields=['appendedEntries[0].id', 'appendedEntries[0].timestamp'])


def verify_snapshot(root, expected, actual, evidence_prefix=None):
    owned, session = home_files()
    before = owned[session].encode()
    with (root / 'home' / session).open('rb') as file:
        after = file.read(SESSION_EVIDENCE_LIMIT + 1)
    proof = session_integrity(before, after)
    session_key = 'home/' + session
    original_oracle_valid = digest(before) == expected['files'].get(session_key)
    # Every field/hash apart from the one session's raw hash remains strict.
    compared = dict(actual, files=dict(actual['files']))
    compared['files'][session_key] = expected['files'][session_key]
    other_unchanged = compared == expected
    accepted = original_oracle_valid and other_unchanged and proof['accepted']
    result = dict(unchanged=actual == expected, rawByteEquality=actual == expected,
                  integrityAccepted=accepted, semanticFixtureContentPreserved=accepted,
                  nonSessionFixtureUnchanged=other_unchanged, originalSessionOracleVerified=original_oracle_valid,
                  sessionIntegrity=proof, actual=actual)
    if evidence_prefix is not None:
        prefix = Path(evidence_prefix)
        targets = [prefix.with_name(prefix.stem + '-session-before.jsonl'), prefix.with_name(prefix.stem + '-session-after.jsonl')]
        if any(os.path.lexists(p) for p in targets):
            raise ValueError('Refusing to overwrite synthetic session evidence')
        if len(after) <= SESSION_EVIDENCE_LIMIT:
            for target, payload in zip(targets, (before, after)):
                target.write_bytes(payload)
                target.chmod(0o600)
            result['sessionEvidence'] = dict(before=targets[0].name, after=targets[1].name,
                                             limitBytes=SESSION_EVIDENCE_LIMIT, syntheticOnly=True)
        else:
            result['sessionEvidence'] = dict(omitted='Session exceeded the bounded evidence limit', limitBytes=SESSION_EVIDENCE_LIMIT)
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('--manifest', type=Path, required=True)
    parser.add_argument('--verify', type=Path)
    parser.add_argument('--session-evidence', action='store_true')
    args = parser.parse_args()
    if args.verify:
        expected = json.loads(args.verify.read_text())
        actual = snapshot(args.root)
        verification = verify_snapshot(args.root, expected, actual, args.manifest if args.session_evidence else None)
        args.manifest.write_text(pretty(verification))
        if not verification['integrityAccepted']:
            raise SystemExit('Fixture integrity failed: inspect the post-capture verification manifest and bounded session evidence')
    else:
        print(pretty(materialize(args.root, args.manifest)))


if __name__ == '__main__':
    main()
