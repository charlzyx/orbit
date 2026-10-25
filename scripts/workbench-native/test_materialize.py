import importlib.util
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

HERE = Path(__file__).resolve().parent
REPO = HERE.parent.parent
spec = importlib.util.spec_from_file_location('materialize', HERE / 'materialize.py')
m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)


class FixtureTests(unittest.TestCase):
    def test_reproducible_git_and_owned_files_match_oracle(self):
        expected = json.loads((REPO / 'testdata/workbench-native/manifest.json').read_text())
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            first = m.materialize(base / 'one', base / 'one.json')
            second = m.materialize(base / 'two', base / 'two.json')
            self.assertEqual(first, second)
            self.assertEqual(first, expected)
            self.assertEqual(first['porcelain'], ['M  alpha.go', ' M bravo.go', '?? notes.txt'])
            self.assertEqual(first['remotes'], [])
            self.assertEqual(len(first['commits']), 2)
            self.assertEqual(m.snapshot(base / 'one'), first)
            self.assertEqual(m.git(base / 'one/workspace', 'config', 'core.hooksPath').strip(), '/dev/null')
            self.assertTrue((base / 'one/workspace/docs/empty').is_dir())

    def test_new_root_no_overwrite_or_partial_writes(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            root = base / 'existing'; root.mkdir()
            protected = root / 'keep'; protected.write_text('untouched')
            with self.assertRaisesRegex(ValueError, 'Refusing'):
                m.materialize(root, base / 'manifest.json')
            self.assertEqual(protected.read_text(), 'untouched')
            self.assertFalse((base / 'manifest.json').exists())
            (base / 'manifest.json').write_text('untouched manifest')
            with self.assertRaises(ValueError):
                m.materialize(base / 'unused', base / 'manifest.json')
            self.assertFalse((base / 'unused').exists())

    def test_symlink_root_and_parent_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            (base / 'dangling').symlink_to(base / 'absent')
            with self.assertRaises(ValueError):
                m.materialize(base / 'dangling', base / 'manifest.json')
            (base / 'real').mkdir(); (base / 'link').symlink_to(base / 'real', target_is_directory=True)
            with self.assertRaises(ValueError):
                m.materialize(base / 'link' / 'fixture', base / 'manifest.json')
            self.assertFalse((base / 'real/fixture').exists())

    def test_session_schema_parent_chain_membership_and_no_inference(self):
        files, path = m.home_files()
        entries = [json.loads(line) for line in files[path].splitlines()]
        self.assertEqual(entries[0]['type'], 'session')
        self.assertEqual(entries[0]['version'], 3)
        self.assertEqual(entries[0]['cwd'], str(m.WORK))
        parent = None
        for entry in entries[1:]:
            self.assertEqual(entry['parentId'], parent)
            parent = entry['id']
        messages = [e['message'] for e in entries if e['type'] == 'message']
        self.assertEqual([e['role'] for e in messages], ['user', 'assistant'])
        self.assertIn('Workbench fixture ready.', messages[1]['content'][0]['text'])
        self.assertEqual(entries[-1]['name'], 'Workbench fixture')
        self.assertEqual(json.loads(files['.orbit-pi/workspaces.json'])['workspaces'], [str(m.WORK)])
        self.assertFalse(json.loads(files['.orbit-pi/auto-title.json'])['enabled'])
        self.assertNotIn('auth.json', '\n'.join(files))
        self.assertEqual(sum(p.endswith('.jsonl') for p in files), 1)

    def test_deleted_empty_directory_changes_snapshot(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            expected = m.materialize(base / 'root', base / 'manifest.json')
            (base / 'root/workspace/docs/empty').rmdir()
            actual = m.snapshot(base / 'root')
            self.assertEqual(actual['emptyDirectories'], [])
            self.assertNotEqual(actual, expected)

    def test_unexpected_content_means_directory_is_not_empty(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            expected = m.materialize(base / 'root', base / 'manifest.json')
            (base / 'root/workspace/docs/empty/unexpected.txt').write_text('unexpected fixture mutation\n')
            actual = m.snapshot(base / 'root')
            self.assertEqual(actual['emptyDirectories'], [])
            self.assertIn('?? docs/empty/unexpected.txt', actual['porcelain'])
            self.assertNotEqual(actual, expected)

    def test_unexpected_empty_directory_changes_snapshot(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            expected = m.materialize(base / 'root', base / 'manifest.json')
            (base / 'root/workspace/unexpected').mkdir()
            actual = m.snapshot(base / 'root')
            self.assertEqual(actual['emptyDirectories'], ['workspace/docs/empty', 'workspace/unexpected'])
            self.assertNotEqual(actual, expected)

    def test_owned_change_detected(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            expected = m.materialize(base / 'root', base / 'manifest.json')
            (base / 'root/workspace/alpha.go').write_text('mutated\n')
            self.assertNotEqual(m.snapshot(base / 'root'), expected)

    def test_pinned_pi_parser_if_installed(self):
        module = Path(os.environ.get('ORBIT_PI_SESSION_MANAGER', '/tmp/orbit-pi-runtime/node_modules/@earendil-works/pi-coding-agent/dist/core/session-manager.js'))
        if not module.is_file():
            self.skipTest('Pinned Pi is not installed; fixture schema/chain tests still run')
        files, path = m.home_files()
        source = 'import { buildSessionContext } from ' + json.dumps(module.as_uri()) + ';\n'
        source += 'const entries = ' + json.dumps([json.loads(line) for line in files[path].splitlines()][1:]) + ';\n'
        source += 'console.log(JSON.stringify(buildSessionContext(entries)));'
        value = json.loads(subprocess.check_output(['node', '--input-type=module', '-e', source], text=True))
        self.assertEqual(len(value['messages']), 2)
        self.assertEqual(value['model'], {'provider': 'anthropic', 'modelId': 'claude-sonnet-4-6'})


if __name__ == '__main__': unittest.main()
