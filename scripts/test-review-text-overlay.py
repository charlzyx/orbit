#!/usr/bin/env python3
import importlib.util,json,os,shutil,subprocess,tempfile,unittest
from pathlib import Path
ROOT=Path(__file__).resolve().parent.parent
spec=importlib.util.spec_from_file_location('wrapper',ROOT/'scripts/go-native-overlay.py')
w=importlib.util.module_from_spec(spec);spec.loader.exec_module(w)
class ReviewOverlay(unittest.TestCase):
 @classmethod
 def setUpClass(cls):
  cls.module=Path(json.loads(subprocess.check_output([os.environ.get('ORBIT_GO','go'),'list','-m','-json','github.com/egoist/mygo'],cwd=ROOT))['Dir'])
  cls.assets=ROOT/'third_party/mygo-review-text'
  cls.identity=json.loads((ROOT/'third_party/mygo-composer/baseline.json').read_text())
 def test_exact_composition_and_no_cache_writes(self):
  before={str(p):w.sha(p) for p in self.module.rglob('*') if p.is_file()}
  review=w.verify_review_text(self.module,self.identity,None)
  selection=w.verify_selection(self.module,ROOT/'third_party/mygo-selection',self.identity)
  joined=w.merge_reviewed([review,selection,{self.identity['added_file']:ROOT/'third_party/mygo-composer/orbit_composer.go.txt'}])
  self.assertEqual(len(review),11);self.assertEqual(len(joined),17)
  self.assertEqual(before,{str(p):w.sha(p) for p in self.module.rglob('*') if p.is_file()})
  with self.assertRaises(RuntimeError):w.merge_reviewed([review,{'ui/layout.go':Path('unreviewed')}])
 def test_global_diagnostics_rejected_before_file_access(self):
  for mode in w.DIAGNOSTIC_MODES:
   with self.assertRaisesRegex(RuntimeError,'conflict'):w.verify_review_text(Path('/not-a-module'),{},mode)
 def test_asset_and_manifest_drift_fail_closed(self):
  for name in ['baseline.json','review-fractional.patch','ui__layout.go.txt','LICENSE']:
   with tempfile.TemporaryDirectory() as d:
    dest=Path(d)/'assets';shutil.copytree(self.assets,dest)
    p=dest/name;p.write_bytes(p.read_bytes()+b'\n')
    with self.assertRaises((RuntimeError,ValueError)):w.verify_review_text(self.module,self.identity,None,dest)
 def test_upstream_drift_and_added_collision(self):
  verifier=w.import_verifier(self.assets/'prepare.py')
  for altered in ['ui/layout.go','ui/orbit_fractional.go']:
   with tempfile.TemporaryDirectory() as d:
    module=Path(d)
    for rel in verifier.BASELINE:
     p=module/rel;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(self.module/rel,p)
    (module/altered).write_text('unexpected')
    with self.assertRaises(ValueError):w.verify_review_text(module,self.identity,None)
 def test_identity_mismatch(self):
  expected=dict(self.identity,version='wrong')
  with self.assertRaisesRegex(RuntimeError,'identity'):w.verify_review_text(self.module,expected,None)
if __name__=='__main__':unittest.main()
