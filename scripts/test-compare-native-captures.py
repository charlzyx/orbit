import importlib.util
import json
import tempfile
import unittest
from pathlib import Path
from PIL import Image

spec=importlib.util.spec_from_file_location('compare',Path(__file__).with_name('compare-native-captures.py')); mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
class TestNativeComparison(unittest.TestCase):
 def fixture(self,p,size=(4,3),color=(1,2,3,255),state='fixed'):
  p.mkdir();Image.new('RGBA',size,color).save(p/'state.png');(p/'manifest.json').write_text(json.dumps({'source':'fixture','actualImageSizes':{'state.png':list(size)},'requestedViewport':[4,3],'state':state}))
 def test_exact_and_one_changed_pixel(self):
  with tempfile.TemporaryDirectory() as t:
   a,b,o=[Path(t)/n for n in ['a','b','out']];self.fixture(a);self.fixture(b)
   self.assertEqual(mod.compare(a,b,o)['cases'][0]['status'],'EXACT_PIXELS')
   im=Image.open(b/'state.png');im.putpixel((0,0),(2,2,3,255));im.save(b/'state.png')
   r=mod.compare(a,b,o)['cases'][0];self.assertEqual(r['status'],'DIFFERENT');self.assertEqual(r['changed_pixels'],1);self.assertEqual(r['manual_semantic_review'],'UNREVIEWED')
 def test_size_metadata_and_missing_refuse_metrics(self):
  for variant in ['size','state','missing']:
   with self.subTest(variant=variant),tempfile.TemporaryDirectory() as t:
    a,b,o=[Path(t)/n for n in ['a','b','out']];self.fixture(a);self.fixture(b,size=(5,3) if variant=='size' else (4,3),state='other' if variant=='state' else 'fixed')
    if variant=='missing':(b/'state.png').unlink()
    r=mod.compare(a,b,o)['cases'][0];self.assertIn(r['status'],['MISSING','RESHOOT']);self.assertNotIn('normalized_rmse',r)
 def test_exact_pixels_do_not_validate_missing_fixture_metadata(self):
  with tempfile.TemporaryDirectory() as t:
   a,b,o=[Path(t)/n for n in ['a','b','out']];self.fixture(a);self.fixture(b)
   r=mod.compare(a,b,o)
   self.assertEqual(r['cases'][0]['status'],'EXACT_PIXELS')
   self.assertEqual(r['normalized_fixture_gate'],'BLOCKED')
   self.assertTrue(r['normalization_gaps'])
 def test_complete_metadata_still_requires_semantic_review(self):
  with tempfile.TemporaryDirectory() as t:
   a,b,o=[Path(t)/n for n in ['a','b','out']];self.fixture(a);self.fixture(b)
   for p in [a,b]:
    m=json.loads((p/'manifest.json').read_text())
    m.update(palette='azure',language='en',fontSize=14,density=100,reduceMotion=True,display='x11',dpi=96,uiScale=1,timezone='UTC',resolvedFontFaces=[{'sha256':'fixture','faceIndex':0}],binarySha256='fixture',sourceRootGitHead='fixture',clockFrozen=True,animationPhaseFrozen=True,focusAndCaretPhaseVerified=True,sourceRootDirty=False)
    (p/'manifest.json').write_text(json.dumps(m))
   r=mod.compare(a,b,o)
   self.assertEqual(r['normalized_fixture_gate'],'METADATA_READY_MANUAL_REVIEW_REQUIRED')
   self.assertEqual(r['cases'][0]['manual_semantic_review'],'UNREVIEWED')
   m['resolvedFontFaces']=[{'sha256':'different','faceIndex':0}];(b/'manifest.json').write_text(json.dumps(m))
   self.assertEqual(mod.compare(a,b,o)['normalized_fixture_gate'],'BLOCKED')
 def test_zero_declared_cases_is_not_acceptance(self):
  with tempfile.TemporaryDirectory() as t:
   a,b,o=[Path(t)/n for n in ['a','b','out']];self.fixture(a);self.fixture(b)
   for p in [a,b]:
    m=json.loads((p/'manifest.json').read_text());m['actualImageSizes']={};(p/'manifest.json').write_text(json.dumps(m))
   with self.assertRaisesRegex(ValueError,'zero comparisons'):
    mod.compare(a,b,o)
if __name__=='__main__':unittest.main()
