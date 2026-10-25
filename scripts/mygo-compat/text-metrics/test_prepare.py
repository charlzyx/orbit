#!/usr/bin/env python3
import importlib.util,json,pathlib,shutil,tempfile,unittest
ROOT=pathlib.Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('candidate_prepare',ROOT/'prepare.py');P=importlib.util.module_from_spec(spec);spec.loader.exec_module(P)

class CandidateVerificationTests(unittest.TestCase):
 def setUp(self):
  self.temp=tempfile.TemporaryDirectory();self.root=pathlib.Path(self.temp.name);self.assets=self.root/'assets';self.source=self.root/'source';shutil.copytree(ROOT,self.assets)
  # Reverse the exact review patch into a private fixture. Verification below
  # proves its reconstructed upstream file has the pinned hash; no mutable
  # checkout path or network source is needed by the unit suite.
  import subprocess
  for target,filename in P.FILES.items():
   file=self.source/target;file.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(ROOT/filename,file)
  subprocess.run(['git','apply','--reverse',str(ROOT/'text-metrics.patch')],cwd=self.source,check=True,capture_output=True)

 def tearDown(self): self.temp.cleanup()
 def manifest(self,mutate):
  p=self.assets/'baseline.json';m=json.loads(p.read_text());mutate(m);p.write_text(json.dumps(m))
 def test_exact_assets_verify(self): self.assertEqual(set(P.verify(self.source,self.assets)[1]),set(P.FILES))
 def test_upstream_hash_drift_rejected(self):
  p=self.source/'internal/text/pango_linux.go';p.write_text(p.read_text()+'\n')
  with self.assertRaises(ValueError):P.verify(self.source,self.assets)
 def test_overlay_hash_drift_rejected(self):
  p=self.assets/'orbit_text_metrics_linux.go.txt';p.write_text(p.read_text()+'\n')
  with self.assertRaises(ValueError):P.verify(self.source,self.assets)
 def test_unexpected_paths_rejected(self):
  self.manifest(lambda m:m['overlays'].update({'ui/extra.go':{'source':'extra','sha256':'0'*64}}))
  with self.assertRaises(ValueError):P.verify(self.source,self.assets)
 def test_path_substitution_rejected(self):
  self.manifest(lambda m:m['overlays']['ui/orbit_text_metrics_linux.go'].update(source='../other'))
  with self.assertRaises(ValueError):P.verify(self.source,self.assets)
 def test_module_identity_rejected(self):
  self.manifest(lambda m:m.update(module_sum='different'))
  with self.assertRaises(ValueError):P.verify(self.source,self.assets)
 def test_patch_drift_rejected(self):
  p=self.assets/'text-metrics.patch';p.write_text(p.read_text()+'\n')
  with self.assertRaises(ValueError):P.verify(self.source,self.assets)
 def test_existing_additive_file_rejected(self):
  p=self.source/'internal/text/orbit_text_metrics_linux.go';p.write_text('existing')
  with self.assertRaises(ValueError):P.verify(self.source,self.assets)
 def test_symlink_escape_rejected(self):
  p=self.assets/'orbit_text_metrics_linux.go.txt';outside=self.root/'outside.txt';outside.write_bytes(p.read_bytes());p.unlink();p.symlink_to(outside)
  with self.assertRaises(ValueError):P.verify(self.source,self.assets)
if __name__=='__main__':unittest.main(verbosity=2)
