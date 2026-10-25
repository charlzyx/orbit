import copy
import importlib.util
import json
import os
from pathlib import Path
import selectors
import signal
import subprocess
import tempfile
import time
import unittest

HERE = Path(__file__).resolve().parent
REPO = HERE.parent.parent
spec = importlib.util.spec_from_file_location('materialize', HERE / 'materialize.py')
m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)


def original():
    files, path = m.home_files()
    return files[path].encode()


def workflow():
    return dict(type='custom', customType='orbit:workflow', data={'mode': 'build'}, id='1234abcd', parentId='workbench05', timestamp='2026-10-25T11:57:09.559Z')


def append(before, entry):
    return before + json.dumps(entry, separators=(',', ':')).encode() + b'\n'


class SessionIntegrityTests(unittest.TestCase):
    def test_only_one_proven_workflow_append_is_accepted_not_called_unchanged(self):
        before = original(); proof = m.session_integrity(before, append(before, workflow()))
        self.assertTrue(proof['accepted'])
        self.assertTrue(proof['semanticFixtureContentPreserved'])
        self.assertTrue(proof['originalEntryBytesPreserved'])
        self.assertFalse(proof['rawBytesEqual'])
        self.assertEqual(proof['appendedEntries'], [workflow()])

    def test_unchanged_session_remains_accepted(self):
        self.assertTrue(m.session_integrity(original(), original())['rawBytesEqual'])
        self.assertTrue(m.session_integrity(original(), original())['accepted'])

    def test_every_unproved_append_shape_is_rejected(self):
        variants = []
        for key, value in [('type', 'message'), ('customType', 'orbit:quota'), ('data', {'mode': 'plan'}), ('parentId', 'wrong'), ('id', 'not-hex'), ('timestamp', '2026-99-25T11:57:09.559Z')]:
            entry = workflow(); entry[key] = value; variants.append(entry)
        extra = workflow(); extra['usage'] = {}; variants.append(extra)
        extra = workflow(); extra['data']['extra'] = 'anything'; variants.append(extra)
        for entry in variants:
            with self.subTest(entry=entry): self.assertFalse(m.session_integrity(original(), append(original(), entry))['accepted'])

    def test_rewrites_extra_entries_truncation_and_oversize_rejected(self):
        before = original()
        values = [before.replace(b'Workbench fixture ready.', b'Changed response.'), append(append(before, workflow()), workflow()), before[:-1], append(before, workflow()) + b'\n', before + b'{' + b'x' * m.SESSION_EVIDENCE_LIMIT]
        for value in values:
            with self.subTest(size=len(value)): self.assertFalse(m.session_integrity(before, value)['accepted'])

    def test_verification_saves_bounded_synthetic_bytes_and_keeps_other_hashes_strict(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / 'fixture'; expected = m.materialize(root, Path(directory) / 'before.json')
            session = root / expected['session']; session.write_bytes(append(session.read_bytes(), workflow()))
            proof = m.verify_snapshot(root, expected, m.snapshot(root), Path(directory) / 'after.json')
            self.assertFalse(proof['unchanged']); self.assertTrue(proof['integrityAccepted'])
            self.assertTrue(proof['nonSessionFixtureUnchanged'])
            self.assertEqual((Path(directory) / proof['sessionEvidence']['before']).read_bytes(), original())
            self.assertEqual((Path(directory) / proof['sessionEvidence']['after']).read_bytes(), session.read_bytes())
            (root / 'workspace/alpha.go').write_text('real change\n')
            self.assertFalse(m.verify_snapshot(root, expected, m.snapshot(root))['integrityAccepted'])

    def test_real_pi_metadata_only_switch_proves_complete_additive_diff(self):
        module = Path(os.environ.get('ORBIT_PI_SESSION_MANAGER', '/tmp/orbit-pi-runtime/node_modules/@earendil-works/pi-coding-agent/dist/core/session-manager.js'))
        cli = module.parent.parent / 'bundle/cli.js'
        if not cli.is_file(): self.skipTest('Pinned Pi1.4 CLI not installed')
        with tempfile.TemporaryDirectory(prefix='orbit-workbench-rpc-proof-') as directory:
            base = Path(directory); root = base / 'fixture'; expected = m.materialize(root, base / 'before.json')
            session = root / expected['session']
            entries = [json.loads(line) for line in session.read_text().splitlines()]
            # Only the local probe cwd differs from the canonical hosted fixture.
            entries[0]['cwd'] = str(root / 'workspace')
            session.write_text(''.join(json.dumps(entry, separators=(',', ':')) + '\n' for entry in entries))
            before = session.read_bytes()
            args = ['node', str(cli), '--mode', 'rpc', '--offline', '--approve', '--session-dir', str(root / 'home/.pi/agent/sessions'), '--no-skills', '--no-prompt-templates', '--no-context-files']
            for ext in ('quota', 'guard', 'title', 'workflow'):
                args.extend(['--extension', str(REPO / f'contrib/orbit-{ext}-extension/index.js')])
            env = dict(PATH=os.environ['PATH'], HOME=str(root / 'home'), PI_CODING_AGENT_DIR=str(root / 'home/.pi/agent'), PI_OFFLINE='1', PI_SKIP_VERSION_CHECK='1', LANG='C.UTF-8', LC_ALL='C.UTF-8', TZ='UTC')
            with (base / 'stderr.log').open('w') as error:
                process = subprocess.Popen(args, cwd=root / 'workspace', env=env, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=error, text=True, bufsize=1, start_new_session=True)
                selector = selectors.DefaultSelector(); selector.register(process.stdout, selectors.EVENT_READ)
                events = []
                try:
                    for kind, fields in [('switch_session', {'sessionPath': str(session)}), ('get_state', {}), ('get_messages', {}), ('get_entries', {})]:
                        ident = 'proof-' + kind
                        process.stdin.write(json.dumps(dict(id=ident, type=kind, **fields)) + '\n'); process.stdin.flush()
                        deadline = time.monotonic() + 20; response = None
                        while time.monotonic() < deadline and response is None:
                            self.assertIsNone(process.poll(), 'Pi exited unexpectedly')
                            for key, _ in selector.select(.1):
                                line = key.fileobj.readline()
                                if line:
                                    entry = json.loads(line); events.append(entry)
                                    if entry.get('id') == ident: response = entry; break
                        self.assertIsNotNone(response, 'Pi metadata response timed out')
                        self.assertTrue(response.get('success'), response)
                finally:
                    if process.poll() is None: os.killpg(process.pid, signal.SIGTERM)
                    process.wait(timeout=5); selector.close(); process.stdin.close(); process.stdout.close()
            proof = m.session_integrity(before, session.read_bytes())
            self.assertTrue(proof['accepted'], proof)
            self.assertFalse(proof['rawBytesEqual'])
            self.assertEqual(len(proof['appendedEntries']), 1)
            self.assertFalse(any(event.get('type') in ('agent_start', 'agent_end', 'turn_start', 'turn_end', 'message_start', 'message_update', 'message_end') for event in events))


if __name__ == '__main__': unittest.main()
