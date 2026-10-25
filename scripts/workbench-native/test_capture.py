import importlib.util
import json
import os
import subprocess
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
REPO = HERE.parent.parent
spec = importlib.util.spec_from_file_location('capture', HERE / 'capture.py')
c = importlib.util.module_from_spec(spec); spec.loader.exec_module(c)
SUITE = json.loads((REPO / 'testdata/workbench-native/cases.json').read_text())


def row(text='alpha.go', x=1250, y=155, confidence=95):
    return dict(level='5', text=text, conf=str(confidence), left=str(x), top=str(y), width='70', height='15')


class CaptureTests(unittest.TestCase):
    def test_truncated_tsv_word_has_no_text_and_cannot_satisfy_readiness_or_click(self):
        header = 'level\tleft\ttop\twidth\theight\tconf\ttext\n'
        for truncated in ['5\t1250\t155\t70\t15\t95', '5\t1250\t155\t70\t15\t95\t']:
            with self.subTest(truncated=truncated):
                rows = c.words(header + truncated)
                self.assertEqual(rows, [])
                self.assertFalse(c.ready(rows, ['alpha.go'])[0])
                with self.assertRaisesRegex(ValueError, 'no click sent'):
                    c.exact_target(rows, 'alpha.go', [1100, 90, 1470, 850])
        valid = '5\t1250\t155\t70\t15\t95\talpha.go\n'
        rows = c.words(header + valid + '5\t1250\t155\t70\t15\t95')
        self.assertEqual(len(rows), 1)
        self.assertEqual(c.exact_target(rows, 'alpha.go', [1100, 90, 1470, 850])['center'], [1285, 162])

    def test_startup_and_case_driver_share_owned_evidence_directory(self):
        with tempfile.TemporaryDirectory() as directory:
            first = c.Capture('42', 123, directory, SUITE)
            (first.evidence / 'startup.json').write_text('preserved')
            second = c.Capture('42', 123, directory, SUITE)
            self.assertEqual((second.evidence / 'startup.json').read_text(), 'preserved')

    def test_actual_hook_resolves_root_from_non_repository_cwd(self):
        # Deliberately fail the Node preflight before creating a fixture or
        # invoking any native command. A cwd-dependent git lookup fails first.
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            mockbin = base / 'bin'; mockbin.mkdir()
            node = mockbin / 'node'; node.write_text('#!/bin/sh\nprintf "mock-node-before-native\\n"\n'); node.chmod(0o755)
            cwd = base / 'not-a-repository'; cwd.mkdir()
            env = dict(os.environ, PATH=str(mockbin) + os.pathsep + os.environ['PATH'], GITHUB_ACTIONS='true', RUNNER_OS='Linux', DISPLAY=':92')
            result = subprocess.run(['bash', str(HERE / 'capture.sh'), '/bin/true', str(base / 'output'), 'go', str(REPO), 'head', 'false', 'source'], cwd=cwd, env=env, text=True, capture_output=True)
            self.assertEqual(result.returncode, 1)
            self.assertIn('Expected pinned Node v22.23.4, got mock-node-before-native', result.stdout)
            self.assertNotIn('not a git repository', result.stderr)
            self.assertFalse((base / 'output').exists())

    def test_clean_environment_preserves_pinned_node_and_script_root(self):
        shell = (HERE / 'capture.sh').read_text()
        self.assertIn('node_directory=$(dirname "$node_binary")', shell)
        self.assertIn('node_version" == v22.23.4', shell)
        self.assertIn('node_modules/.bin:$node_directory:', shell)
        self.assertIn('${BASH_SOURCE[0]}', shell)
        self.assertNotIn('repo_root=$(git rev-parse', shell)
        self.assertIn('env -i', shell)
        self.assertIn('GITHUB_ACTIONS', shell)

    def test_exact_target_requires_unique_confident_region_match(self):
        self.assertEqual(c.exact_target([row()], 'alpha.go', [1100, 90, 1470, 850])['center'], [1285, 162])
        for rows in [[], [row('alpha.goM')], [row(confidence=59)], [row(x=200)], [row(), row()]]:
            with self.subTest(rows=rows), self.assertRaises(ValueError):
                c.exact_target(rows, 'alpha.go', [1100, 90, 1470, 850])
        # A duplicate name elsewhere in the editor must not redirect the dock click.
        self.assertEqual(c.exact_target([row(), row(x=500)], 'alpha.go', [1100, 90, 1470, 850])['bbox'], [1250, 155, 70, 15])

    def test_readiness_is_not_pixel_equality_and_rejects_loading(self):
        rows = [row('Workbench'), row('fixture'), row('ready.')]
        self.assertTrue(c.ready(rows, ['Workbench fixture ready'])[0])
        self.assertFalse(c.ready(rows, ['alpha.go'])[0])
        self.assertFalse(c.ready(rows + [row('Loading...')], ['Workbench fixture ready'])[0])

    def test_recipes_are_eight_distinct_bounded_actual_inputs(self):
        self.assertEqual([case['id'] for case in SUITE['cases']], [f'WB{n:02}' for n in range(1, 9)])
        self.assertEqual(len(set(case['image'] for case in SUITE['cases'])), 8)
        self.assertIsNone(SUITE['globalVisualIds'])
        self.assertEqual(SUITE['viewport'], [1470, 922])
        self.assertEqual(SUITE['cases'][2]['actions'][0]['clickExactLabel'], 'alpha.go')
        self.assertEqual(SUITE['cases'][3]['actions'][0]['clickExactLabel'], 'bravo.go')
        self.assertEqual(SUITE['cases'][7]['actions'][-1], {'key': 'ctrl+j'})
        self.assertEqual(SUITE['cases'][7]['actions'][0]['clickExactLabel'], 'Workbench')
        self.assertEqual(SUITE['cases'][4]['actions'][2]['clickExactLabel'], 'Workbench')
        self.assertEqual(SUITE['cases'][6]['actions'][0]['clickExactLabel'], 'message')
        self.assertTrue(all(case['source'] and case['require'] for case in SUITE['cases']))
        self.assertFalse(any('type' in action for case in SUITE['cases'] for action in case['actions']))

    def test_ambiguous_target_sends_no_pointer_input(self):
        with tempfile.TemporaryDirectory() as directory:
            cap = c.Capture('42', 123, directory, SUITE)
            with patch.object(cap, 'capture', return_value=([row(), row()], {})), patch.object(c, 'run') as run:
                with self.assertRaises(ValueError):
                    cap.action('WB03', 1, SUITE['cases'][2]['actions'][0])
                self.assertEqual(run.call_args_list[0].args, ('xdotool', 'windowactivate', '--sync', '42'))
                self.assertEqual(run.call_count, 1)

    def test_click_logs_whole_frame_hash_and_target_before_input(self):
        with tempfile.TemporaryDirectory() as directory:
            cap = c.Capture('42', 123, directory, SUITE)
            events = []
            def command(*args):
                if 'mousemove' in args:
                    stored = json.loads((Path(directory) / 'actions.json').read_text())
                    self.assertEqual(stored[-1]['outcome'], 'TARGET_VERIFIED_CLICK_PENDING')
                    self.assertEqual(stored[-1]['before']['pngSha256'], 'frame-hash')
                events.append(args)
                return ''
            with patch.object(cap, 'capture', return_value=([row()], {'pngSha256': 'frame-hash'})), patch.object(c, 'run', side_effect=command), patch.object(c.time, 'sleep'):
                cap.action('WB03', 1, SUITE['cases'][2]['actions'][0])
            self.assertIn(('xdotool', 'mousemove', '--window', '42', '1285', '162', 'click', '1'), events)
            self.assertEqual(cap.actions[-1]['target']['bbox'], [1250, 155, 70, 15])

    def test_intermediate_readiness_is_recorded_without_an_extra_case(self):
        with tempfile.TemporaryDirectory() as directory:
            cap = c.Capture('42', 123, directory, SUITE)
            with patch.object(c, 'run'), patch.object(c.time, 'sleep'), patch.object(cap, 'settle', return_value={'consecutiveReadySamples': 3}) as settle:
                cap.action('WB05', 2, dict(key='ctrl+shift+g', afterRequire=['Staged', 'Changes']))
            self.assertEqual(settle.call_args.args[0]['image'], 'evidence/WB05-action-2-ready.png')
            self.assertEqual(cap.actions[-1]['afterReadiness']['consecutiveReadySamples'], 3)
            self.assertEqual(cap.cases, [])

    def test_settle_resets_after_missing_sample_and_preserves_bytes(self):
        with tempfile.TemporaryDirectory() as directory:
            cap = c.Capture('42', 123, directory, SUITE)
            observations = iter([True, True, False, True, True, True])
            def capture(path):
                path.write_bytes(b'actual-image-bytes')
                return [row('Ready' if next(observations) else 'Loading...')], {'image': [1470, 922]}
            with patch.object(cap, 'capture', side_effect=capture), patch.object(c.time, 'sleep'):
                proof = cap.settle(dict(id='UNIT', image='unit.png', require=['Ready']))
            self.assertEqual(proof['samples'], 6)
            self.assertEqual(proof['consecutiveReadySamples'], 3)
            self.assertEqual((Path(directory) / 'unit.png').read_bytes(), b'actual-image-bytes')

    def test_bad_client_size_refuses_screenshot_instead_of_rescaling(self):
        with tempfile.TemporaryDirectory() as directory:
            cap = c.Capture('42', 123, directory, SUITE)
            with patch.object(cap, 'alive'), patch.object(c, 'run', return_value='WIDTH=800\nHEIGHT=600') as run:
                with self.assertRaisesRegex(ValueError, 'no resizing'):
                    cap.capture(Path(directory) / 'wrong.png')
                self.assertEqual(run.call_count, 1)

    def test_mocked_sequence_records_manual_unreviewed_not_pass(self):
        with tempfile.TemporaryDirectory() as directory:
            cap = c.Capture('42', 123, directory, SUITE)
            with patch.object(cap, 'action') as action, patch.object(cap, 'settle', return_value={'image': [1470, 922]}), patch.object(c, 'run'):
                cap.execute()
            self.assertEqual(action.call_count, sum(len(case['actions']) for case in SUITE['cases']))
            result = json.loads((Path(directory) / 'capture-result.json').read_text())
            self.assertEqual(result['capturedCases'], 8)
            self.assertEqual(result['semanticState'], 'UNREVIEWED')
            self.assertEqual(result['pixelEquality'], 'NOT_COMPARED')
            self.assertTrue(all(row['semanticState'] == 'UNREVIEWED' for row in result['cases']))

    def test_failed_case_keeps_prior_cases_and_blocks_not_count_inflation(self):
        with tempfile.TemporaryDirectory() as directory:
            cap = c.Capture('42', 123, directory, SUITE)
            with patch.object(cap, 'action'), patch.object(cap, 'settle', side_effect=[{}, RuntimeError('not ready')]), patch.object(c, 'run'):
                with self.assertRaises(RuntimeError): cap.execute()
            result = json.loads((Path(directory) / 'capture-result.json').read_text())
            self.assertEqual(result['capturedCases'], 1)
            self.assertEqual(result['captureValidity'], 'BLOCKED')


if __name__ == '__main__': unittest.main()
