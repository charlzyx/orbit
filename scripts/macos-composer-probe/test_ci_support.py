import hashlib,json,tempfile,unittest,zipfile,warnings
from pathlib import Path
import ci_support as c
class CISupportTests(unittest.TestCase):
 def setUp(self):self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup);self.r=Path(self.tmp.name)
 def test_exact_harness_copy_and_extra_hash_rejection(self):
  src=self.r/'src';src.mkdir();(src/'a').write_text('a');m=src/'harness-manifest.json';m.write_text(json.dumps({'files':{'a':c.sha(src/'a')}}));c.copy_harness(src,self.r/'copy',m,c.sha(m));self.assertEqual((self.r/'copy/a').read_text(),'a')
  (src/'extra').write_text('x')
  with self.assertRaises(ValueError):c.copy_harness(src,self.r/'bad',m,c.sha(m))
 def test_archive_hash_traversal_symlink_duplicate_refused(self):
  for i,kind in enumerate(('hash','traversal','symlink','duplicate')):
   z=self.r/f'{i}.zip'
   with zipfile.ZipFile(z,'w') as f:
    if kind=='traversal':f.writestr('../x','x')
    elif kind=='symlink':
     n=zipfile.ZipInfo('link');n.external_attr=0o120777<<16;f.writestr(n,'target')
    else:
     f.writestr('safe','x')
     if kind=='duplicate':
      with warnings.catch_warnings():
       warnings.simplefilter('ignore',UserWarning);f.writestr('safe','y')
   with self.assertRaises(ValueError):c.extract_known(z,'0'*64 if kind=='hash' else c.sha(z),self.r/f'out{i}')
 def test_exact_archive_extract(self):
  z=self.r/'a.zip'
  with zipfile.ZipFile(z,'w') as f:f.writestr('dir/file','data')
  c.extract_known(z,c.sha(z),self.r/'out');self.assertEqual((self.r/'out/dir/file').read_text(),'data')
 def test_collect_refuses_arbitrary_or_excess_cases(self):
  p=self.r/'summary';p.write_text(json.dumps({'cases':[{'out_dir':'/tmp/user'}]}))
  with self.assertRaises(ValueError):c.collect(p,self.r/'out')
  p.write_text(json.dumps({'cases':[{}]*5}))
  with self.assertRaises(ValueError):c.collect(p,self.r/'out2')
