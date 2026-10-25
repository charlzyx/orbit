import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('analyze',HERE/'analyze.py');analyze=importlib.util.module_from_spec(spec);spec.loader.exec_module(analyze)
ps=importlib.util.spec_from_file_location('prepare',HERE/'prepare.py');prepare=importlib.util.module_from_spec(ps);ps.loader.exec_module(prepare)
class TraceTests(unittest.TestCase):
    def setUp(self):self.tmp=tempfile.TemporaryDirectory();self.root=Path(self.tmp.name)
    def tearDown(self):self.tmp.cleanup()
    def batch(self,events=None,**fields):
        d={'schema':1,'renderer':'mygo-pango','diagnostic':True,'case_id':'synthetic','completed_present_call':1,'presentation_semantics':'headless specimen; no native present','dropped_events':0,'events':events if events is not None else [{'kind':'face','font_id':'f1','face_index':0,'loaded_bytes_sha256':'a'*64,'font_file_sha256':'a'*64},{'kind':'shape_run','font_id':'f1','text':'General'}]};d.update(fields);(self.root/'batch.json').write_text(json.dumps(d));return d
    def test_empty_rejected(self):
        with self.assertRaises(ValueError):analyze.read_trace(self.root)
    def test_empty_events_rejected(self):
        self.batch([])
        with self.assertRaises(ValueError):analyze.read_trace(self.root)
    def test_missing_drops_rejected(self):
        d=self.batch();d.pop('dropped_events');(self.root/'batch.json').write_text(json.dumps(d))
        with self.assertRaises(ValueError):analyze.read_trace(self.root)
    def test_missing_loaded_not_equal(self):
        self.batch([{'kind':'face','font_id':'f1','face_index':0},{'kind':'shape_run','font_id':'f1','text':'General'}]);r=analyze.read_trace(self.root)
        self.assertEqual(analyze.compare(r,r)['shared_texts'][0]['observed_face_set_status'],'unknown')
    def test_same_faces_never_pixel_pass(self):
        self.batch();r=analyze.read_trace(self.root);c=analyze.compare(r,r)
        self.assertEqual(c['shared_texts'][0]['observed_face_set_status'],'same');self.assertEqual(c['native_pixel_parity'],'not evaluated');self.assertFalse(r['paint_observed'])
    def test_drops_reported(self):
        self.batch(dropped_events=1);self.assertTrue(analyze.read_trace(self.root)['problems'])
    def test_conflicting_face_rejected(self):
        d=self.batch();d['events'].append({'kind':'face','font_id':'f1','face_index':1,'loaded_bytes_sha256':'b'*64});self.batch(d['events'])
        with self.assertRaises(ValueError):analyze.read_trace(self.root)
    def test_invalid_font_hash_rejected(self):
        self.batch([{'kind':'face','font_id':'f1','loaded_bytes_sha256':'no'}])
        with self.assertRaises(ValueError):analyze.read_trace(self.root)
    def test_anchor_drift_refused(self):
        p=self.root/'x';p.write_text('abc')
        with self.assertRaises(ValueError):prepare.replace(self.root,'x','missing','new')
        self.assertEqual(p.read_text(),'abc')
    def test_source_hash_drift_refused(self):
        with self.assertRaises(ValueError):prepare.verify(self.root,'mygo')
    def test_no_production_integration(self):
        repo=HERE.parent.parent
        self.assertNotIn('font-fidelity',(repo/'scripts/go-native.sh').read_text())
        self.assertNotIn('font-fidelity',(repo/'scripts/go-native-overlay.py').read_text())
        for path in (repo/'crates/orbit-pi/src').rglob('*.rs'):
            self.assertNotIn('orbit_font_trace',path.read_text())
if __name__=='__main__':unittest.main()
