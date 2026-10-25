#!/usr/bin/env python3
"""Run only a locally built diagnostic app against disposable native Mac fixtures.

This never changes accessibility/security permissions or reads user Pi state.
The app executable has an explicit generated instrumentation overlay; it is not
claimed to be byte-identical to the downloaded baseline executable.
"""
from pathlib import Path
import argparse, hashlib, json, os, platform, shutil, subprocess, time, signal


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()


def run(app, output):
    if platform.system() != 'Darwin' or platform.machine() != 'arm64':
        raise RuntimeError('Native Apple Silicon runner required')
    app, output = app.resolve(), output.resolve()
    if output.exists():
        raise RuntimeError('Use a fresh evidence directory, never overwrite a prior run')
    output.mkdir(parents=True)
    home = output / 'synthetic-home'
    home.mkdir()
    (home / '.orbit-workspace-probe-synthetic-home').write_text('Synthetic CI fixture; no user data.\n')
    workspaces = [home / name for name in ('Initial Workspace', 'Selected Workspace A', 'Selected Workspace B 中文')]
    for i, directory in enumerate(workspaces):
        directory.mkdir()
        (directory / 'fixture.txt').write_text(f'Isolated workspace fixture {i}. No private content.\n')
    # Seed normal persisted recents, not an injected controller action. These
    # allow real recent-row lifecycle coverage before native panel acceptance.
    preferences = home / 'Library/Application Support/Orbit Go Test/Home/.orbit-pi'
    preferences.mkdir(parents=True)
    (preferences / 'workspaces.json').write_text(json.dumps({'workspaces': [{'path': str(p), 'added_at': i + 1} for i, p in enumerate(workspaces)]}, indent=2) + '\n')
    # The recent picker derives rows from saved Pi sessions, not merely the
    # workspace list. These are inert synthetic history records, never prompts
    # sent to a provider. Seed both the shipped .3 private home and the corrected
    # normal HOME fixture; each application version reads only its selected one.
    for profile, pi_home in (('private', preferences.parent), ('normal', home)):
        sessions = pi_home / '.pi/agent/sessions/native-regression'
        sessions.mkdir(parents=True)
        for i, directory in enumerate(workspaces):
            records = [
                {'type': 'session', 'version': 3, 'id': f'synthetic-{profile}-{i}', 'cwd': str(directory), 'timestamp': '2026-01-01T00:00:00Z'},
                {'type': 'message', 'id': f'fixture-{i}', 'parentId': None, 'timestamp': '2026-01-01T00:00:01Z', 'message': {'role': 'user', 'content': 'Native regression fixture, never submitted to a model', 'timestamp': 1767225601000}},
            ]
            (sessions / f'fixture-{i}.jsonl').write_text(''.join(json.dumps(record) + '\n' for record in records))
    (output / 'config.json').write_text(json.dumps({'workspaces': [str(p) for p in workspaces], 'invalid': str(home / 'Missing Workspace')}, indent=2) + '\n')
    executable = app / 'Contents/MacOS/Orbit'
    if not executable.is_file():
        raise RuntimeError('Installed diagnostic app executable missing')
    env = {'HOME': str(home), 'PATH': '/usr/bin:/bin:/usr/sbin:/sbin', 'TMPDIR': str(output / 'temp'), 'LANG': 'en_US.UTF-8', 'ORBIT_WORKSPACE_PROBE_DIR': str(output), 'GOTRACEBACK': 'crash'}
    Path(env['TMPDIR']).mkdir()
    start = time.time()
    launch = ['/usr/bin/open', '-n', '-W', '--stdout', str(output / 'app.stdout.log'), '--stderr', str(output / 'app.stderr.log')]
    for k, v in env.items():
        launch += ['--env', k + '=' + v]
    launch += [str(app), '--args', '--offline', '--workspace', str(workspaces[0])]
    meta = {'started_at_unix': start, 'runner_macos': platform.mac_ver()[0], 'architecture': platform.machine(), 'instrumented_executable_sha256': sha(executable), 'app': str(app), 'synthetic_home': str(home), 'native_interaction': 'AppKit events and actual NSOpenPanel; no external accessibility or security-setting changes', 'claim_limit': 'Generated diagnostic overlay; not byte-identical to the released executable. Baseline production source remains unchanged.'}
    (output / 'launch.json').write_text(json.dumps(meta, indent=2) + '\n')
    timed_out = False
    with (output / 'launch.log').open('w') as log:
        proc = subprocess.Popen(launch, cwd='/', env=env, stdout=log, stderr=subprocess.STDOUT)
        try:
            code = proc.wait(timeout=210)
        except subprocess.TimeoutExpired:
            timed_out = True
            stages = output / 'stages.jsonl'
            for line in stages.read_text().splitlines() if stages.is_file() else []:
                try:
                    pid = json.loads(line).get('pid')
                except ValueError:
                    continue
                if not isinstance(pid, int) or pid <= 0:
                    continue
                command = subprocess.run(['/bin/ps', '-p', str(pid), '-o', 'command='], text=True, stdout=subprocess.PIPE).stdout.strip()
                if command == str(executable) or command.startswith(str(executable) + ' '):
                    # Sample only this verified disposable app, never another
                    # runner process. Sampling failure cannot hide the timeout.
                    with (output / 'timeout-sample.log').open('w') as sample_log:
                        try:
                            subprocess.run(['/usr/bin/sample', str(pid), '1', '1'], stdout=sample_log, stderr=subprocess.STDOUT, timeout=8)
                        except subprocess.TimeoutExpired:
                            sample_log.write('\nSampling timed out.\n')
                    try: os.kill(pid, signal.SIGTERM)
                    except ProcessLookupError: pass
                    time.sleep(2)
                    remaining = subprocess.run(['/bin/ps', '-p', str(pid), '-o', 'command='], text=True, stdout=subprocess.PIPE).stdout.strip()
                    if remaining == str(executable) or remaining.startswith(str(executable) + ' '):
                        # MyGo's first SIGTERM requests graceful Quit. If a
                        # nested modal await prevents that, stop this exact test
                        # process rather than leaking it on the hosted runner.
                        try: os.kill(pid, signal.SIGKILL)
                        except ProcessLookupError: pass
                    break
            proc.terminate()
            try:
                proc.wait(timeout=10)
            except subprocess.TimeoutExpired:
                proc.kill(); proc.wait()
            code = -1
    # Collect only this freshly launched synthetic process's diagnostics. Never
    # upload an arbitrary runner-wide diagnostic directory.
    report = {}
    if (output / 'report.json').is_file():
        report = json.loads((output / 'report.json').read_text())
    pid = report.get('pid')
    if not isinstance(pid, int):
        for line in (output / 'stages.jsonl').read_text().splitlines() if (output / 'stages.jsonl').is_file() else []:
            try:
                entry = json.loads(line)
                if isinstance(entry.get('pid'), int): pid = entry['pid']; break
            except ValueError:
                continue
    diagnostics = []
    if isinstance(pid, int):
        import pwd
        for base in {home, Path(pwd.getpwuid(os.getuid()).pw_dir)}:
            directory = base / 'Library/Logs/DiagnosticReports'
            if not directory.is_dir(): continue
            for source in directory.glob('Orbit*.ips'):
                if source.stat().st_mtime < start: continue
                raw = source.read_text(errors='replace')
                # Modern .ips has a separate metadata first line plus JSON body.
                body = raw.split('\n', 1)[-1]
                try: data = json.loads(body)
                except ValueError: continue
                if data.get('pid') != pid: continue
                dest = output / 'crash' / source.name
                dest.parent.mkdir(exist_ok=True)
                shutil.copyfile(source, dest)
                diagnostics.append({'file': str(dest.relative_to(output)), 'sha256': sha(dest)})
    result = {'launch_returncode': code, 'timed_out': timed_out, 'pid': pid, 'app_report_status': report.get('status', 'MISSING'), 'crash_files': diagnostics, 'elapsed_seconds': round(time.time()-start, 2)}
    (output / 'runner-result.json').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result, indent=2))
    if timed_out or code != 0 or report.get('status') != 'PASS':
        raise RuntimeError('Native workspace probe failed or did not complete; inspect preserved stage/stderr evidence')


if __name__ == '__main__':
    p = argparse.ArgumentParser()
    p.add_argument('--app', type=Path, required=True)
    p.add_argument('--output', type=Path, required=True)
    a = p.parse_args()
    run(a.app, a.output)
