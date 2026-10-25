import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('verify_target', HERE / 'verify-target.py')
v = importlib.util.module_from_spec(spec); spec.loader.exec_module(v)


def row():
    return dict(level='5', text='Details', conf='97', left='1388', top='97', width='60', height='28')


class TargetTests(unittest.TestCase):
    def test_details_uses_fresh_pixels_records_hash_box_and_sends_no_input(self):
        with tempfile.TemporaryDirectory() as directory:
            out = Path(directory)
            with patch.object(v.capture.Capture, 'capture', return_value=([row()], {'pngSha256': 'fresh-frame', 'ocrSha256': 'fresh-ocr', 'client': [1470, 922]})) as capture, patch.object(v.capture, 'run') as run:
                coordinates = v.verify('42', 123, out, 'Details', [1000, 70, 1470, 160])
                capture.assert_called_once_with(out / 'details-before.png')
                run.assert_not_called()
            self.assertEqual(coordinates, [1418, 111])
            proof = json.loads((out / 'details-target.json').read_text())
            self.assertEqual(proof['frame']['pngSha256'], 'fresh-frame')
            self.assertEqual(proof['target']['bbox'], [1388, 97, 60, 28])
            self.assertEqual(proof['action'], 'NO_INPUT_SENT')
            self.assertEqual(proof['semanticState'], 'UNREVIEWED')

    def test_missing_or_ambiguous_details_fails_without_input(self):
        for rows in ([], [row(), row()]):
            with self.subTest(rows=rows), tempfile.TemporaryDirectory() as directory:
                with patch.object(v.capture.Capture, 'capture', return_value=(rows, {'pngSha256': 'fresh-frame'})), patch.object(v.capture, 'run') as run:
                    with self.assertRaises(ValueError):
                        v.verify('42', 123, directory, 'Details', [1000, 70, 1470, 160])
                    run.assert_not_called()
                self.assertEqual(json.loads((Path(directory) / 'details-target.json').read_text())['status'], 'BLOCKED')

    def test_previous_evidence_cannot_authorize_new_pointer_action(self):
        with tempfile.TemporaryDirectory() as directory:
            out = Path(directory); (out / 'details-before.png').write_bytes(b'old')
            with patch.object(v.capture.Capture, 'capture') as capture:
                with self.assertRaisesRegex(ValueError, 'previous action'):
                    v.verify('42', 123, out, 'Details', [1000, 70, 1470, 160])
                capture.assert_not_called()

    def test_main_helper_receives_verified_coordinates_and_clicks_once(self):
        main = (HERE.parent / 'capture-native-reference.sh').read_text()
        self.assertNotIn('"$output" 1420 111', main)
        self.assertIn('"$app_pid" "$output/usage-evidence" Details 1000 70 1470 160', main)
        self.assertIn('details "$window" "$output" "$details_x" "$details_y"', main)
        helper = (HERE.parent / 'usage-native/capture-state.sh').read_text()
        self.assertEqual(helper.count('click 1'), 1)
        self.assertNotIn('click', (HERE / 'verify-target.py').read_text().split('def verify(', 1)[1].split('def main(', 1)[0])


if __name__ == '__main__': unittest.main()
