#!/usr/bin/env python3
"""Create a new, local-only Git workspace and Pi v3 session, never user data."""
import argparse
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


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('--manifest', type=Path, required=True)
    parser.add_argument('--verify', type=Path)
    args = parser.parse_args()
    if args.verify:
        expected = json.loads(args.verify.read_text())
        actual = snapshot(args.root)
        args.manifest.write_text(pretty(dict(unchanged=actual == expected, actual=actual)))
        if actual != expected:
            raise SystemExit('Fixture changed: inspect the post-capture verification manifest')
    else:
        print(pretty(materialize(args.root, args.manifest)))


if __name__ == '__main__':
    main()
