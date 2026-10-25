import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('capture', HERE / 'capture.py')
c = importlib.util.module_from_spec(spec); spec.loader.exec_module(c)
FIXTURES = HERE / 'testdata'
PRIMARY = (FIXTURES / 'wb07-rust-psm11.tsv').read_text()
ALTERNATE = (FIXTURES / 'wb07-rust-psm3.tsv').read_text()
PROVENANCE = json.loads((FIXTURES / 'wb07-rust-ocr.json').read_text())
REQUIRED = PROVENANCE['requiredPhrases']


class ReadinessOCRTests(unittest.TestCase):
    def test_actual_synthetic_ocr_regression_requires_complete_original_titles(self):
        self.assertEqual(c.ready(c.words(PRIMARY), REQUIRED)[1]['missing'], ['Add stable helper'])
        self.assertIn('ada stable helper', c.plain(c.words(PRIMARY)))
        self.assertTrue(c.ready(c.words(ALTERNATE), REQUIRED)[0])
        for filename, digest in PROVENANCE['files'].items():
            self.assertEqual(c.sha(FIXTURES / filename), digest)
        self.assertEqual(PROVENANCE['imageTransformations'], [])

    def prepare(self, directory, primary=PRIMARY):
        cap = c.Capture('42', 123, directory, {})
        path = cap.evidence / 'WB07-settle.png'
        path.write_bytes(b'unchanged-whole-client-pixels')
        path.with_suffix('.tsv').write_text(primary)
        metadata = dict(pngSha256=c.sha(path), ocrSha256=c.sha(path.with_suffix('.tsv')))
        return cap, path, metadata

    def test_fallback_preserves_both_ocr_passes_and_same_full_image_hash(self):
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.prepare(directory)
            with patch.object(c, 'run', return_value=ALTERNATE) as run:
                accepted, detail, proof = cap.readiness(path, c.words(PRIMARY), metadata, REQUIRED, 1)
            self.assertTrue(accepted)
            self.assertEqual(detail, dict(missing=[], blocked=[]))
            self.assertEqual(proof['selectedReadinessOcrPSM'], 3)
            self.assertEqual([row['psm'] for row in proof['readinessOcrPasses']], [11, 3])
            source = cap.evidence / 'WB07-settle-fallback-001.png'
            run.assert_called_once_with('tesseract', str(source), 'stdout', '--psm', '3', 'tsv')
            self.assertEqual(source.read_bytes(), path.read_bytes())
            self.assertEqual(source.with_suffix('.tsv').read_text(), PRIMARY)
            for row in proof['readinessOcrPasses']:
                self.assertEqual(row['pngSha256'], metadata['pngSha256'])
                self.assertEqual(c.sha(Path(directory) / row['tsv']), row['tsvSha256'])
            self.assertEqual(cap.cases, [])

    def test_fallback_does_not_merge_partial_recognitions_or_correct_literals(self):
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.prepare(directory)
            # Alternate sees the previously missing title but loses the other.
            alternate = ALTERNATE.replace('Initialize', 'Wrong')
            with patch.object(c, 'run', return_value=alternate):
                accepted, detail, proof = cap.readiness(path, c.words(PRIMARY), metadata, REQUIRED, 1)
            self.assertFalse(accepted)
            self.assertIn('Initialize workbench fixture', detail['missing'])
            self.assertIsNone(proof['selectedReadinessOcrPSM'])

    def test_primary_error_is_never_overridden_by_fallback(self):
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.prepare(directory)
            rows = c.words(PRIMARY) + [dict(text='Loading...')]
            with patch.object(c, 'run') as run:
                accepted, detail, proof = cap.readiness(path, rows, metadata, REQUIRED, 1)
            self.assertFalse(accepted); self.assertEqual(detail['blocked'], ['loading...'])
            run.assert_not_called(); self.assertEqual(len(proof['readinessOcrPasses']), 1)

    def test_primary_success_needs_no_fallback_and_alternate_errors_still_block(self):
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.prepare(directory)
            with patch.object(c, 'run') as run:
                accepted, _, proof = cap.readiness(path, c.words(ALTERNATE), metadata, REQUIRED, 1)
            self.assertTrue(accepted); run.assert_not_called()
            self.assertEqual(proof['selectedReadinessOcrPSM'], 11)
            alternate = ALTERNATE.replace('Initialize', 'Loading... Initialize')
            with patch.object(c, 'run', return_value=alternate):
                accepted, detail, _ = cap.readiness(path, c.words(PRIMARY), metadata, REQUIRED, 2)
            self.assertFalse(accepted); self.assertEqual(detail['blocked'], ['loading...'])

    def test_fallback_rejects_changed_image_and_does_not_send_input(self):
        with tempfile.TemporaryDirectory() as directory:
            cap, path, metadata = self.prepare(directory)
            path.write_bytes(b'changed')
            with patch.object(c, 'run') as run:
                with self.assertRaisesRegex(ValueError, 'differs'):
                    cap.readiness(path, c.words(PRIMARY), metadata, REQUIRED, 1)
                run.assert_not_called()


if __name__ == '__main__': unittest.main()
