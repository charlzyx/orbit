import importlib.util
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch
SPEC=importlib.util.spec_from_file_location('probe_prepare',Path(__file__).with_name('prepare.py'));m=importlib.util.module_from_spec(SPEC);SPEC.loader.exec_module(m)
class PrepareTests(unittest.TestCase):
 def setUp(self):
  self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup);self.root=Path(self.tmp.name);(self.root/'internal/ui').mkdir(parents=True);self.here=self.root/'scripts/macos-composer-probe';(self.here/'menuproof').mkdir(parents=True);(self.here/'composer_probe_darwin.go.txt').write_text('fixture helper');(self.here/'menuproof/policy.go').write_text('fixture policy');self.source='func View() {\n\tx.UI.View(c)\n}\nfunc Ready() {\n\t\tw.OnReadyToShow(func() {\n})\n}\n';(self.root/'internal/ui/run.go').write_text(self.source)
  self.addCleanup(patch.stopall);patch.object(m,'ROOT',self.root).start();patch.object(m,'HERE',self.here).start();self.calls=[]
  def git(args,**kw):
   self.calls.append(args)
   if args[1] in ('diff','ls-files'):return b''
   return m.BASE_TREE+'\n' if args[-1]=='HEAD^{tree}' else 'local-exact-tree-head\n'
  self.git=patch.object(m.subprocess,'check_output',side_effect=git).start()
 def test_exact_overlay_preserves_original_and_only_two_anchors(self):
  out=self.root/'output';result=m.prepare(out);self.assertTrue(result.is_file());self.assertEqual((self.root/'internal/ui/run.go').read_text(),self.source);s=(out/'run.go').read_text();self.assertEqual(s.count('composerProofObserve(x,c)'),1);self.assertEqual(s.count('composerProofStart(x,w)'),1)
 def test_wrong_base_no_output(self):
  self.git.side_effect=lambda *a,**k:'wrong\n';out=self.root/'out'
  with self.assertRaises(ValueError):m.prepare(out)
  self.assertFalse(out.exists())
 def test_dirty_production_no_output(self):
  self.git.side_effect=lambda args,**k:b'dirty' if args[1]=='diff' else m.BASE_TREE+'\n';out=self.root/'out'
  with self.assertRaises(ValueError):m.prepare(out)
  self.assertFalse(out.exists())
 def test_untracked_production_refused(self):
  self.git.side_effect=lambda args,**k:b'rogue.go' if args[1]=='ls-files' else (b'' if args[1]=='diff' else m.BASE_TREE+'\n');out=self.root/'out'
  with self.assertRaises(ValueError):m.prepare(out)
  self.assertFalse(out.exists())
 def test_missing_or_duplicate_anchor_no_output(self):
  for s in ['missing',self.source+self.source]:
   (self.root/'internal/ui/run.go').write_text(s);out=self.root/'out'
   with self.assertRaises(ValueError):m.prepare(out)
   self.assertFalse(out.exists())
 def test_existing_or_relative_output_refused(self):
  out=self.root/'out';out.mkdir()
  for p in [out,Path('relative')]:
   with self.assertRaises(ValueError):m.prepare(p)
 def test_real_helper_shadow_refused(self):
  (self.root/'internal/ui/composer_probe_darwin.go').write_text('existing');out=self.root/'out'
  with self.assertRaises(ValueError):m.prepare(out)
  self.assertFalse(out.exists())
if __name__=='__main__':unittest.main()
