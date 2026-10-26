import hashlib,json,plistlib,tempfile,unittest
from pathlib import Path
from test_bundle import derive,tree_hashes
class TestBundleDerivation(unittest.TestCase):
 def setUp(self):
  self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup);self.root=Path(self.tmp.name);self.app=self.root/'Baseline.app';(self.app/'Contents/MacOS').mkdir(parents=True);(self.app/'Contents/Resources/Provenance').mkdir(parents=True);(self.app/'Contents/MacOS/Orbit').write_bytes(b'baseline');(self.app/'Contents/Info.plist').write_bytes(plistlib.dumps({'CFBundleExecutable':'Orbit','CFBundleIdentifier':'fixture.keep'}));(self.app/'Contents/Resources/Provenance/bundle-provenance.json').write_text(json.dumps({'source_commit':'old-official-source'}));self.expected=self.root/'expected.json';self.expected.write_text(json.dumps(tree_hashes(self.app)));self.sha=hashlib.sha256(self.expected.read_bytes()).hexdigest();self.bin=self.root/'instrumented';self.bin.write_bytes(b'instrumented');self.binsha=hashlib.sha256(self.bin.read_bytes()).hexdigest();self.overlay=self.root/'overlay.json';self.overlay.write_text(json.dumps({'source_tree':'0476116687acf0838d52a0246210270000530530','source_remote_commit':'ee25813f9a2065ae8632d4c15b24e3c99e7bf469'}));self.out=self.root/'out'
 def call(self):return derive(self.app,self.expected,self.sha,self.bin,self.binsha,self.overlay,self.out)
 def test_only_binary_changed_original_retained(self):
  old=tree_hashes(self.app);app=self.call();self.assertEqual(tree_hashes(self.app),old);m=json.loads((self.out/'test-bundle-manifest.json').read_text());self.assertEqual(m['changed_paths'],['Contents/MacOS/Orbit']);self.assertEqual(m['baseline_source'],'old-official-source');self.assertEqual((app/'Contents/Info.plist').read_bytes(),(self.app/'Contents/Info.plist').read_bytes())
 def test_independent_manifest_mismatch_refused(self):
  self.sha='0'*64
  with self.assertRaises(ValueError):self.call()
  self.assertFalse(self.out.exists())
 def test_baseline_tamper_refused(self):
  (self.app/'Contents/MacOS/Orbit').write_bytes(b'tampered')
  with self.assertRaises(ValueError):self.call()
 def test_wrong_instrumented_hash_refused(self):
  self.binsha='0'*64
  with self.assertRaises(ValueError):self.call()
 def test_existing_output_refused(self):
  self.out.mkdir()
  with self.assertRaises(ValueError):self.call()
 def test_wrong_base_refused(self):
  self.overlay.write_text('{}')
  with self.assertRaises(ValueError):self.call()
if __name__=='__main__':unittest.main()
