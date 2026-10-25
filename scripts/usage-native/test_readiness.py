import importlib.util
import json
import os
import subprocess
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('usage_readiness', HERE / 'wait-ready.py')
u = importlib.util.module_from_spec(spec); spec.loader.exec_module(u)


def rows(text):
    return [dict(text=word) for word in text.split()]


class UsageReadinessTests(unittest.TestCase):
    def test_real_capture_schema_keeps_image_dimensions_and_path(self):
        # Exercise Capture.capture itself, not a fake result dictionary. Only
        # the native subprocess boundary is stubbed with raw recorded OCR.
        with tempfile.TemporaryDirectory() as directory:
            tsv = (HERE / 'testdata' / 'completed-go-simple.tsv').read_text()
            def native(*args):
                if args[:2] == ('xdotool', 'getwindowgeometry'):
                    return 'WIDTH=1470\nHEIGHT=922'
                if args[0] == 'import':
                    Path(args[-1]).write_bytes(b'unchanged-whole-native-frame')
                    return ''
                if args[0] == 'identify': return '1470 922'
                if args[0] == 'tesseract': return tsv
                if args[:2] == ('xdotool', 'getwindowname'): return 'Orbit'
                self.fail(f'Unexpected native command {args}')
            with patch.object(u.capture, 'run', side_effect=native), patch.object(u.capture.os, 'kill'), patch.object(u.time, 'sleep'):
                result = u.wait_ready('42', 123, directory, 'simple')
            self.assertEqual(result['samples'], 3)
            for sample in result['observations']:
                self.assertEqual(sample['image'], [1470, 922])
                self.assertEqual(sample['client'], [1470, 922])
                self.assertTrue((Path(directory) / sample['imagePath']).is_file())
                self.assertEqual(sample['pngSha256'], u.capture.sha(Path(directory) / sample['imagePath']))

    def test_loading_usage_with_titles_is_never_completed(self):
        required = u.REQUIRED['simple']
        for marker in ('Loading local usage…', 'Loading local usage...', 'Unable to load usage data', 'Unable to read usage data', 'Refreshing…', 'Refreshing.', 'permission denied', 'No usage data yet'):
            with self.subTest(marker=marker):
                self.assertFalse(u.capture.ready(rows(' '.join(required) + ' ' + marker), required)[0])

    def test_completed_simple_and_details_have_distinct_body_anchors(self):
        for phase in ('simple', 'details'):
            self.assertTrue(u.capture.ready(rows(' '.join(u.REQUIRED[phase])), u.REQUIRED[phase])[0])
        self.assertFalse(u.capture.ready(rows('Usage Simple Details All models'), u.REQUIRED['simple'])[0])
        self.assertFalse(u.capture.ready(rows(' '.join(u.REQUIRED['simple'])), u.REQUIRED['details'])[0])

    def test_observations_preserved_and_loading_resets_consecutive_gate(self):
        with tempfile.TemporaryDirectory() as directory:
            observations = iter([True, True, False, True, True, True])
            def frame(path):
                good = next(observations)
                path.write_bytes(b'whole-window-' + str(good).encode())
                path.with_suffix('.tsv').write_text('unmodified-ocr')
                return rows(' '.join(u.REQUIRED['simple']) + ('' if good else ' Loading local usage…')), {'image': [1470, 922], 'client': [1470, 922], 'window': '42', 'title': 'Orbit', 'pngSha256': u.capture.sha(path), 'ocrSha256': u.capture.sha(path.with_suffix('.tsv'))}
            with patch.object(u.capture.Capture, 'capture', side_effect=frame), patch.object(u.time, 'sleep'), patch.object(u.capture, 'run') as command:
                result = u.wait_ready('42', 123, directory, 'simple')
            command.assert_not_called()
            self.assertEqual(result['consecutiveReadySamples'], 3)
            self.assertEqual(result['samples'], 6)
            self.assertEqual(len(list((Path(directory) / 'usage-readiness').glob('simple-*.png'))), 6)
            self.assertEqual((Path(directory) / 'usage-simple-dark.png').read_bytes(), b'whole-window-True')

    def test_timeout_saves_blocked_proof_without_publishing_skeleton(self):
        with tempfile.TemporaryDirectory() as directory:
            def frame(path):
                path.write_bytes(b'skeleton'); path.with_suffix('.tsv').write_text('loading')
                return rows('Loading local usage…'), {'image': [1470, 922], 'client': [1470, 922], 'window': '42', 'title': 'Orbit'}
            with patch.object(u.capture.Capture, 'capture', side_effect=frame), patch.object(u.time, 'monotonic', side_effect=[0, 0, 2]), patch.object(u.time, 'sleep'), patch.object(u.capture, 'run') as command:
                with self.assertRaisesRegex(RuntimeError, 'did not become ready'):
                    u.wait_ready('42', 123, directory, 'simple', timeout=1)
            command.assert_not_called()
            self.assertFalse((Path(directory) / 'usage-simple-dark.png').exists())
            self.assertIn('BLOCKED', (Path(directory) / 'usage-readiness' / 'simple-ready.json').read_text())

    def test_previous_result_or_image_cannot_be_reused(self):
        with tempfile.TemporaryDirectory() as directory:
            (Path(directory) / 'usage-simple-dark.png').write_bytes(b'previous')
            with patch.object(u.capture.Capture, 'capture') as frame:
                with self.assertRaisesRegex(ValueError, 'previous'):
                    u.wait_ready('42', 123, directory, 'simple')
                frame.assert_not_called()

    def test_shell_loading_timeout_stops_sequence_without_pointer_click(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory); tools = root / 'tools'; tools.mkdir()
            log = root / 'commands.log'
            for name, body in {
                'xdotool': 'printf "%s\\n" "$*" >> "$COMMAND_LOG"\n',
                'python3': 'echo readiness-blocked >> "$COMMAND_LOG"; exit 1\n',
            }.items():
                path = tools / name; path.write_text('#!/bin/sh\n' + body); path.chmod(0o755)
            env = dict(os.environ, PATH=str(tools) + os.pathsep + os.environ['PATH'],
                       COMMAND_LOG=str(log), ORBIT_CAPTURE_APP_PID='123')
            run = subprocess.run(['bash', str(HERE / 'capture-state.sh'), 'simple', '42', str(root / 'out')], env=env, capture_output=True)
            self.assertNotEqual(run.returncode, 0)
            commands = log.read_text()
            self.assertIn('key ctrl+u', commands)
            self.assertIn('readiness-blocked', commands)
            self.assertNotIn('click', commands)
            self.assertFalse((root / 'out' / 'usage-actions.log').exists())
            # A stale final image is rejected even before window activation.
            (root / 'out' / 'usage-simple-dark.png').write_bytes(b'previous')
            log.write_text('')
            run = subprocess.run(['bash', str(HERE / 'capture-state.sh'), 'simple', '42', str(root / 'out')], env=env, capture_output=True)
            self.assertNotEqual(run.returncode, 0)
            self.assertEqual(log.read_text(), '')

    def test_prior_completed_go_and_rust_whole_images_satisfy_anchors(self):
        provenance = json.loads((HERE / 'testdata' / 'completed-usage.json').read_text())
        self.assertEqual(provenance['imageTransformations'], [])
        for frame in provenance['frames']:
            path = HERE / 'testdata' / frame['tsv']
            self.assertEqual(u.capture.sha(path), frame['tsvSha256'])
            self.assertEqual(frame['required'], u.REQUIRED[frame['phase']])
            self.assertTrue(u.capture.ready(u.capture.words(path.read_text()), frame['required'])[0], frame)

    def test_expired_deadline_cannot_publish_a_late_third_ready_frame(self):
        with tempfile.TemporaryDirectory() as directory:
            def frame(path):
                path.write_bytes(b'whole-window'); path.with_suffix('.tsv').write_text('raw')
                return rows(' '.join(u.REQUIRED['simple'])), {'image': [1470, 922], 'client': [1470, 922], 'window': '42', 'title': 'Orbit'}
            with patch.object(u.capture.Capture, 'capture', side_effect=frame), patch.object(u.time, 'monotonic', side_effect=[0, 0, 0, 0, 2]), patch.object(u.time, 'sleep'):
                with self.assertRaisesRegex(RuntimeError, 'deadline'):
                    u.wait_ready('42', 123, directory, 'simple', timeout=1)
            self.assertFalse((Path(directory) / 'usage-simple-dark.png').exists())

    def test_native_capture_tools_have_individual_timeouts(self):
        with patch.object(u.capture.subprocess, 'check_output', return_value='result') as command:
            self.assertEqual(u.capture.run('tesseract', 'image.png', 'stdout'), 'result')
        self.assertEqual(command.call_args.kwargs['timeout'], 15)

    def test_actual_ce4_loading_ocr_remains_blocked(self):
        path = HERE / 'testdata' / 'ce4-loading.tsv'
        ready, detail = u.capture.ready(u.capture.words(path.read_text()), u.REQUIRED['simple'])
        self.assertFalse(ready)
        self.assertTrue(detail['blocked'])
        provenance = json.loads((HERE / 'testdata' / 'ce4-loading.json').read_text())
        self.assertEqual(u.capture.sha(path), provenance['passes'][0]['tsvSha256'])
        self.assertEqual(provenance['imageTransformations'], [])


if __name__ == '__main__': unittest.main()
