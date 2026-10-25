import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
REPO = HERE.parent.parent
spec = importlib.util.spec_from_file_location('manifest', HERE / 'manifest.py')
m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)


class ManifestTests(unittest.TestCase):
    def inputs(self, base, error=None, unchanged=True):
        (base / 'fixture-before.json').write_bytes((REPO / 'testdata/workbench-native/manifest.json').read_bytes())
        (base / 'fixture-after.json').write_text(json.dumps(dict(unchanged=unchanged)))
        (base / 'binary').write_bytes(b'fixture-binary')
        (base / 'node').write_bytes(b'fixture-node')
        recipe = json.loads((REPO / 'testdata/workbench-native/cases.json').read_text())
        cases = []
        if not error:
            for case in recipe['cases']:
                path = base / case['image']; path.write_bytes(b'actual-native-pixels')
                cases.append(dict(image=case['image'], proof=dict(image=[1470, 922], client=[1470, 922], pngSha256=m.sha(path))))
        (base / 'capture-result.json').write_text(json.dumps(dict(cases=cases, error=error)))

    def build(self, base):
        with patch.object(m.subprocess, 'check_output', return_value='fixture-tool-version\n'):
            return m.build(base, 'go', REPO, base / 'binary', REPO, 'head', 'false', 'source', str(base / 'node'))

    def test_complete_captures_still_do_not_assert_semantics_or_pixels(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory); self.inputs(base)
            result = self.build(base)
            self.assertEqual(result['captureValidity'], 'READY_FOR_MANUAL_REVIEW')
            self.assertEqual(result['semanticState'], 'UNREVIEWED')
            self.assertEqual(result['pixelEquality'], 'NOT_COMPARED')
            self.assertIsNone(result['globalVisualIds'])
            self.assertIsNone(result['resolvedFontFaces'])
            self.assertEqual(len(result['actualImageSizes']), 8)

    def test_valid_metadata_append_is_separate_from_raw_equality_and_visual_acceptance(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory); self.inputs(base, unchanged=False)
            (base / 'fixture-after.json').write_text(json.dumps(dict(unchanged=False, integrityAccepted=True, semanticFixtureContentPreserved=True, sessionIntegrity=dict(appendedEntries=[{'type': 'custom', 'customType': 'orbit:workflow', 'data': {'mode': 'build'}}]))))
            result = self.build(base)
            self.assertEqual(result['captureValidity'], 'READY_FOR_MANUAL_REVIEW')
            self.assertFalse(result['fixtureRawByteEquality'])
            self.assertFalse(result['fixtureUnchangedAfterCapture'])
            self.assertTrue(result['fixtureSemanticContentPreserved'])
            self.assertEqual(result['semanticState'], 'UNREVIEWED')
            self.assertEqual(len(result['expectedVolatileRuntimeMetadata']), 1)
            self.assertTrue(result['normalizationGaps'])

    def test_startup_failure_yields_machine_readable_blocked_manifest(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory); self.inputs(base, error='STARTUP readiness failed')
            result = self.build(base)
            self.assertEqual(result['captureValidity'], 'BLOCKED')
            self.assertEqual(result['actualImageSizes'], {})
            self.assertIn('STARTUP', result['captureResult']['error'])
            self.assertTrue((base / 'manifest.json').is_file())

    def test_changed_fixture_or_png_blocks_validity(self):
        for changed in ('fixture', 'png'):
            with self.subTest(changed=changed), tempfile.TemporaryDirectory() as directory:
                base = Path(directory); self.inputs(base, unchanged=changed != 'fixture')
                if changed == 'png':
                    (base / 'wb01-transcript-dark.png').write_bytes(b'modified pixels')
                result = self.build(base)
                self.assertEqual(result['captureValidity'], 'BLOCKED')


if __name__ == '__main__': unittest.main()
