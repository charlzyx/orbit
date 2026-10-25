import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('capture', HERE / 'capture.py')
c = importlib.util.module_from_spec(spec); spec.loader.exec_module(c)
PRIMARY = (HERE / 'testdata/wb03-go-psm11.tsv').read_text()
ALTERNATE = (HERE / 'testdata/wb03-go-psm3.tsv').read_text()
PROVENANCE = json.loads((HERE / 'testdata/wb03-go-target-ocr.json').read_text())
REGION = PROVENANCE['region']


class TargetOCRTests(unittest.TestCase):
    def setup_frame(self, directory):
        cap = c.Capture('42', 123, directory, {})
        path = cap.evidence / 'WB03-action-1-before.png'
        path.write_bytes(b'unchanged-full-native-window')
        path.with_suffix('.tsv').write_text(PRIMARY)
        metadata = dict(pngSha256=c.sha(path), ocrSha256=c.sha(path.with_suffix('.tsv')))
        return cap, path, metadata

    def test_actual_frozen_artifact_low_confidence_is_not_a_primary_click(self):
        for filename, digest in PROVENANCE['files'].items():
            self.assertEqual(c.sha(HERE / 'testdata' / filename), digest)
        self.assertEqual(PROVENANCE['imageTransformations'], [])
        with self.assertRaisesRegex(ValueError, 'no click sent'):
            c.exact_target(c.words(PRIMARY), 'alpha.go', REGION)
        target = c.exact_target(c.words(ALTERNATE), 'alpha.go', REGION)
        self.assertEqual(target['bbox'], [1286, 190, 44, 11])
        self.assertGreaterEqual(target['confidence'], 60)

    def test_independent_whole_frame_read_confirms_same_exact_literal_and_bounds(self):
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.setup_frame(directory)
            with patch.object(c, 'run', return_value=ALTERNATE) as run:
                target, proof = cap.target(path, c.words(PRIMARY), metadata, 'alpha.go', REGION)
            run.assert_called_once_with('tesseract', str(path), 'stdout', '--psm', '3', 'tsv')
            self.assertEqual(target['bbox'], [1286, 190, 44, 11])
            self.assertEqual(proof['targetOcrPSM'], 3)
            self.assertEqual(proof['pngSha256'], c.sha(path))
            for item in proof['targetOcrPasses']:
                self.assertEqual(item['tsvSha256'], c.sha(Path(directory) / item['tsv']))
            self.assertLess(proof['targetOcrPasses'][0]['confidence'], 60)
            self.assertGreaterEqual(proof['targetOcrPasses'][1]['confidence'], 60)

    def test_missing_duplicate_or_wrong_region_is_never_resolved_by_alternate(self):
        rows = c.words(PRIMARY)
        alpha = next(r for r in rows if r['text'] == 'alpha.go')
        for candidates, region in [([], REGION), (rows + [alpha], REGION), (rows, [0, 0, 100, 100])]:
            with self.subTest(candidates=candidates, region=region), tempfile.TemporaryDirectory() as directory:
                cap, path, metadata = self.setup_frame(directory)
                with patch.object(c, 'run') as run, self.assertRaises(ValueError):
                    cap.target(path, candidates, metadata, 'alpha.go', region)
                run.assert_not_called()

    def test_alternate_must_be_confident_exact_unique_and_same_bounds(self):
        variants = [ALTERNATE.replace('alpha.go', 'alpha.goM'),
                    ALTERNATE.replace('80.610741', '59.999'),
                    ALTERNATE.replace('1286\t190\t44\t11', '1287\t190\t44\t11'),
                    ALTERNATE + next(l for l in ALTERNATE.splitlines(True) if l.rstrip().endswith('\talpha.go'))]
        for alternate in variants:
            with self.subTest(alternate=alternate), tempfile.TemporaryDirectory() as directory:
                cap, path, metadata = self.setup_frame(directory)
                with patch.object(c, 'run', return_value=alternate), self.assertRaises(ValueError):
                    cap.target(path, c.words(PRIMARY), metadata, 'alpha.go', REGION)

    def test_frame_modification_before_or_during_read_never_returns_target(self):
        for during in [False, True]:
            with self.subTest(during=during), tempfile.TemporaryDirectory() as directory:
                cap, path, metadata = self.setup_frame(directory)
                if not during:
                    path.write_bytes(b'changed')
                def read(*args):
                    path.write_bytes(b'changed')
                    return ALTERNATE
                with patch.object(c, 'run', side_effect=read) as run, self.assertRaises(ValueError):
                    cap.target(path, c.words(PRIMARY), metadata, 'alpha.go', REGION)
                if not during:
                    run.assert_not_called()

    def test_primary_confident_target_does_not_run_alternate(self):
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.setup_frame(directory)
            with patch.object(c, 'run') as run:
                _, proof = cap.target(path, c.words(PRIMARY), metadata, 'bravo.go', REGION)
            run.assert_not_called()
            self.assertEqual(proof['targetOcrPSM'], 11)

    def test_failed_confirmation_never_dispatches_pointer_input(self):
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.setup_frame(directory)
            with patch.object(cap, 'capture', return_value=(c.words(PRIMARY), metadata)), patch.object(c, 'run', return_value=ALTERNATE.replace('alpha.go', 'wrong')) as run:
                with self.assertRaises(ValueError):
                    cap.action('WB03', 1, dict(clickExactLabel='alpha.go', region=REGION))
            self.assertFalse(any('mousemove' in call.args or 'click' in call.args for call in run.call_args_list))
            self.assertEqual(cap.actions, [])

    def test_mixed_confidence_duplicate_literal_is_ambiguous_in_each_pass(self):
        high = next(r for r in c.words(ALTERNATE) if r['text'] == 'alpha.go')
        low = dict(high, left='1350', conf='20')
        with self.assertRaises(ValueError):
            c.exact_target([high, low], 'alpha.go', REGION)
        extra = next(l for l in ALTERNATE.splitlines(True) if l.rstrip().endswith('\talpha.go')).replace('1286\t190', '1350\t190').replace('80.610741', '20')
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.setup_frame(directory)
            with patch.object(c, 'run', return_value=ALTERNATE + extra), self.assertRaises(ValueError):
                cap.target(path, c.words(PRIMARY), metadata, 'alpha.go', REGION)

    def test_nonfinite_or_out_of_range_confidence_is_not_clickable(self):
        source = next(r for r in c.words(ALTERNATE) if r['text'] == 'alpha.go')
        for confidence in ['nan', 'inf', '101', '-1']:
            with self.subTest(confidence=confidence), self.assertRaises(ValueError):
                c.exact_target([dict(source, conf=confidence)], 'alpha.go', REGION)
