#!/usr/bin/env python3
"""Prepare an architecture-pinned Node + Pi tree; never modify the source install.

Default use is on a matching macOS runner. --existing-pi-root is an audit-only
route for an already installed tree, and does not establish macOS execution.
Output must be new or empty. License gaps are explicit findings for release review
in runtime-manifest.json, not a claim that an SPDX field is a license notice.
"""
import argparse, hashlib, json, os, pathlib, platform, posixpath, shutil
import subprocess, sys, tarfile, tempfile, urllib.request

HERE = pathlib.Path(__file__).resolve().parent
MAX_ARCHIVE = 150 * 1024 * 1024
MAX_UNPACKED = 500 * 1024 * 1024
MAX_ENTRIES = 50000
PI_VERSION = '1.4.0'
NOTICE_WORDS = ('license', 'licence', 'notice', 'copying', 'copyright')

def digest(path):
    h = hashlib.sha256()
    with pathlib.Path(path).open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()

def write_json(path, value):
    pathlib.Path(path).write_text(json.dumps(value, indent=2, sort_keys=True) + '\n')

def checked_relative(name, root):
    p = pathlib.PurePosixPath(name)
    if p.is_absolute() or '\\' in name or '..' in p.parts or not p.parts or p.parts[0] != root:
        raise ValueError('Unsafe archive path: ' + name)
    return p

def safe_extract(archive, target, root):
    """Validate bounded tar, write files before symlinks; reject special/hard links."""
    target = pathlib.Path(target)
    if target.exists():
        raise ValueError('Extraction target must not exist')
    with tarfile.open(archive, 'r:gz') as tf:
        members, total = [], 0
        for member in tf:
            members.append(member)
            total += member.size
            if len(members) > MAX_ENTRIES or total > MAX_UNPACKED:
                raise ValueError('Node archive exceeds extraction budget')
        paths, symlinks = set(), set()
        for m in members:
            p = checked_relative(m.name, root)
            name = str(p)
            if name in paths:
                raise ValueError('Duplicate archive path: ' + name)
            paths.add(name)
            if not (m.isfile() or m.isdir() or m.issym()):
                raise ValueError('Unsupported archive entry: ' + name)
            if m.issym():
                if not m.linkname or m.linkname.startswith('/') or '\\' in m.linkname:
                    raise ValueError('Unsafe symlink: ' + name)
                resolved = posixpath.normpath(posixpath.join(posixpath.dirname(name), m.linkname))
                checked_relative(resolved, root)
                symlinks.add(name)
        for name in paths:
            if any(str(parent) in symlinks for parent in pathlib.PurePosixPath(name).parents):
                raise ValueError('Archive entry traverses a symlink: ' + name)
        target.mkdir(parents=True)
        for m in members:
            dest = target.joinpath(*pathlib.PurePosixPath(m.name).parts)
            if m.isdir():
                dest.mkdir(parents=True, exist_ok=True)
            elif m.isfile():
                dest.parent.mkdir(parents=True, exist_ok=True)
                source = tf.extractfile(m)
                if source is None:
                    raise ValueError('Unreadable archive member: ' + m.name)
                with source, dest.open('xb') as output:
                    shutil.copyfileobj(source, output)
                dest.chmod(0o755 if m.mode & 0o111 else 0o644)
        for m in members:
            if m.issym():
                dest = target.joinpath(*pathlib.PurePosixPath(m.name).parts)
                dest.parent.mkdir(parents=True, exist_ok=True)
                dest.symlink_to(m.linkname)
    for link in target.rglob('*'):
        if link.is_symlink() and not link.resolve().is_relative_to(target.resolve()):
            raise ValueError('Symlink escapes extraction root: ' + str(link))
    return {'entries': len(members), 'unpacked_bytes': sum(m.size for m in members)}

def download(url, target):
    req = urllib.request.Request(url, headers={'User-Agent': 'Orbit-runtime-builder/1'})
    count = 0
    with urllib.request.urlopen(req, timeout=90) as response, target.open('xb') as output:
        if response.geturl().split('/')[2] != 'nodejs.org':
            raise ValueError('Unexpected Node archive redirect host')
        while True:
            block = response.read(1024 * 1024)
            if not block:
                break
            count += len(block)
            if count > MAX_ARCHIVE:
                raise ValueError('Node archive exceeds download budget')
            output.write(block)

def file_inventory(root):
    result = []
    for p in sorted(root.rglob('*')):
        rel = p.relative_to(root).as_posix()
        if p.is_symlink():
            if os.path.isabs(os.readlink(p)) or not p.resolve().is_relative_to(root.resolve()):
                raise ValueError('Pi symlink escapes package tree: ' + rel)
            result.append({'path': rel, 'symlink': os.readlink(p)})
        elif p.is_file():
            result.append({'path': rel, 'bytes': p.stat().st_size, 'sha256': digest(p)})
    return result

def audit_pi(pi, licenses):
    lock = json.loads((pi / 'package-lock.json').read_text())
    if lock.get('lockfileVersion') != 3:
        raise ValueError('Expected lockfile v3')
    if lock['packages']['']['dependencies'].get('@earendil-works/pi-coding-agent') != PI_VERSION:
        raise ValueError('Unexpected Pi root pin')
    sources = json.loads((HERE / 'license-sources.json').read_text())
    for filename, source in sources['sources'].items():
        if digest(licenses / filename) != source['sha256']:
            raise ValueError('Supplemental license hash mismatch: ' + filename)
    summaries, issues = [], list(sources['issues'])
    for path, entry in sorted(lock['packages'].items()):
        if not path:
            continue
        pkg = pi / path
        metadata = json.loads((pkg / 'package.json').read_text())
        if metadata['version'] != entry['version'] or not entry.get('integrity'):
            raise ValueError('Installed version/integrity missing: ' + path)
        name = metadata['name']
        notices = []
        for f in sorted(pkg.rglob('*')):
            if not f.is_file() or f.is_symlink() or 'node_modules' in f.relative_to(pkg).parts:
                continue
            if any(word in f.name.lower() for word in NOTICE_WORDS):
                notices.append(f.relative_to(pkg).as_posix())
                copy = licenses / 'npm' / path.removeprefix('node_modules/') / f.relative_to(pkg)
                copy.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(f, copy)
        if name == 'data-uri-to-buffer':
            readme = pkg / 'README.md'
            if 'Permission is hereby granted' in readme.read_text():
                notices.append('README.md (contains full MIT notice)')
                copy = licenses / 'npm' / name / 'README.md'
                copy.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(readme, copy)
        supplement = sources['package_supplements'].get(name)
        if not notices and not supplement:
            issues.append({'package': name, 'version': entry['version'], 'issue': 'No shipped full license notice or verified scoped supplement'})
        summaries.append({'name': name, 'version': entry['version'], 'path': path,
                          'declared_license': metadata.get('license'), 'integrity': entry['integrity'],
                          'resolved': entry.get('resolved'), 'preserved_notices': notices,
                          'supplement': supplement})
    return summaries, issues

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', required=True, type=pathlib.Path)
    parser.add_argument('--arch', required=True, choices=['arm64'])
    parser.add_argument('--node-archive', type=pathlib.Path)
    parser.add_argument('--existing-pi-root', type=pathlib.Path,
                        help='Audit/copy an existing install without executing Node; not a macOS smoke test')
    args = parser.parse_args()
    pins = json.loads((HERE / 'node-pins.json').read_text())
    arch = args.arch
    pin = pins['archives'][arch]
    out = args.output.absolute()
    if out.is_symlink() or (out.exists() and any(out.iterdir())):
        raise ValueError('Output must be new or empty, not a symlink')
    out.mkdir(parents=True, exist_ok=True)
    if not args.existing_pi_root:
        expected_machine = 'arm64'
        if platform.system() != 'Darwin' or platform.machine() != expected_machine:
            raise ValueError('Install must run natively on the matching macOS architecture')
    with tempfile.TemporaryDirectory(prefix='.runtime-build-', dir=out) as tmp:
        tmp = pathlib.Path(tmp)
        archive = args.node_archive or tmp / pin['filename']
        if not args.node_archive:
            download(pin['url'], archive)
        actual_hash = digest(archive)
        if actual_hash != pin['sha256']:
            raise ValueError('Official Node archive SHA-256 mismatch')
        extracted = tmp / 'unpacked'
        extraction = safe_extract(archive, extracted, pin['filename'].removesuffix('.tar.gz'))
        shutil.move(str(next(extracted.iterdir())), out / 'node')
        node = out / 'node' / 'bin' / 'node'
        npm_package = out / 'node' / 'lib' / 'node_modules' / 'npm' / 'package.json'
        npm_version = json.loads(npm_package.read_text())['version']
        if npm_version != pins['npm_version']:
            raise ValueError('Unexpected bundled npm version')
        (out / 'pi').mkdir()
        for name in ('package.json', 'package-lock.json'):
            shutil.copy2(HERE / name, out / 'pi' / name)
        execution = 'not executed; existing-tree audit only'
        if args.existing_pi_root:
            src = args.existing_pi_root.resolve()
            if digest(src / 'package-lock.json') != digest(HERE / 'package-lock.json'):
                raise ValueError('Existing install lock differs from pinned lock')
            shutil.copytree(src / 'node_modules', out / 'pi' / 'node_modules', symlinks=True)
        else:
            got = subprocess.check_output([str(node), '--version'], text=True).strip()
            if got != 'v' + pins['version']:
                raise ValueError('Unexpected bundled Node execution version')
            env = dict(os.environ, PATH=str(node.parent) + os.pathsep + os.environ.get('PATH', ''))
            npm = [str(node), str(npm_package.parent / 'bin' / 'npm-cli.js')]
            npm_got = subprocess.check_output(npm + ['--version'], env=env, text=True).strip()
            if npm_got != pins['npm_version']:
                raise ValueError('Unexpected npm execution version')
            subprocess.run(npm + ['ci', '--ignore-scripts', '--omit=dev', '--no-audit', '--no-fund'],
                           cwd=out / 'pi', env=env, check=True, timeout=600)
            execution = 'matching macOS Node --version and npm ci completed; application smoke still required'
        if digest(out / 'pi' / 'package-lock.json') != digest(HERE / 'package-lock.json'):
            raise ValueError('npm changed pinned lock')
        licenses = out / 'licenses'
        shutil.copytree(HERE / 'licenses', licenses)
        shutil.copy2(HERE / 'license-sources.json', licenses / 'license-sources.json')
        shutil.copy2(out / 'node' / 'LICENSE', licenses / 'NODE-LICENSE.txt')
        original = file_inventory(out / 'pi')
        write_json(out / 'original-pi-files.json', original)
        packages, issues = audit_pi(out / 'pi', licenses)
        examples = out / 'pi' / 'node_modules' / '@earendil-works' / 'pi-coding-agent' / 'examples'
        pruned = [row for row in original if row['path'].startswith('node_modules/@earendil-works/pi-coding-agent/examples/')]
        if examples.is_dir():
            shutil.rmtree(examples)
        native = []
        for p in (out / 'pi').rglob('*'):
            if p.is_file() and p.suffix in ('.node', '.wasm', '.so', '.dylib'):
                with p.open('rb') as f:
                    magic = f.read(4)
                native.append({'path': p.relative_to(out).as_posix(), 'sha256': digest(p),
                               'format': 'WebAssembly' if magic == b'\x00asm' else 'platform binary'})
        manifest = {'node': dict(pins, selected_archive=pin, archive_sha256_verified=actual_hash,
                                extraction=extraction, license_sha256=digest(licenses / 'NODE-LICENSE.txt')),
                    'pi': {'version': PI_VERSION, 'lock_sha256': digest(HERE / 'package-lock.json'),
                           'packages': packages, 'original_file_inventory': 'original-pi-files.json',
                           'original_file_inventory_sha256': digest(out / 'original-pi-files.json')},
                    'execution_evidence': execution, 'pruned_files': pruned,
                    'pruning_note': 'Only Pi example resources omitted; docs and every installed runtime dependency retained. Examples are referenced as optional learning material by system prompt.',
                    'native_dependencies': native, 'license_status': 'incomplete' if issues else 'notice-coverage-audited',
                    'license_issues': issues,
                    'caveat': 'License inventory is a packaging audit, not a legal clearance. Pi upstream bundles strip legal comments and may contain dependency versions different from this npm graph.'}
        write_json(out / 'runtime-manifest.json', manifest)
        print(json.dumps({'output': str(out), 'node': pins['version'], 'npm': npm_version,
                          'pi': PI_VERSION, 'packages': len(packages), 'license_status': manifest['license_status'],
                          'license_issues': issues}, indent=2))
    return 0

if __name__ == '__main__':
    try:
        raise SystemExit(main())
    except Exception as exc:
        print('prepare-runtime: ' + str(exc), file=sys.stderr)
        raise SystemExit(1)
