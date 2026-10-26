import collections,json,tempfile,unittest,os
from pathlib import Path
from unittest.mock import patch
import ci_check_gates as g
import ci_resource_paths as r
class CIPolicyTests(unittest.TestCase):
 def setUp(self):self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup);self.p=Path(self.tmp.name)
 def test_complete_exact_gate_and_missing_skip_refused(self):
  manifest=self.p/'v';manifest.write_text(json.dumps({'python_tests':['mod.C.test_a'],'go_policy_counts':{'TestA':{}}}));py=self.p/'py';py.write_text('test_a (mod.C.test_a) ... ok\nRan 1 tests\nOK\n');go=self.p/'go';events=[{'Action':a,'Test':'TestA'} for _ in range(3) for a in ('run','pass')];go.write_text('\n'.join(map(json.dumps,events)));g.check(py,go,manifest)
  for changed in [events[:-1],events+[{'Action':'skip','Test':'TestA'}],events+[{'Action':'pass','Test':'TestOther'}]]:
   go.write_text('\n'.join(map(json.dumps,changed)))
   with self.assertRaises(ValueError):g.check(py,go,manifest)
 def test_resource_missing_duplicate_wrong_hash_refused(self):
  (self.p/'resources').mkdir();(self.p/'smoke').mkdir()
  with self.assertRaises(ValueError):r.prepare(self.p)
  (self.p/'resources/a.dmg').write_bytes(b'wrong')
  with self.assertRaises(ValueError):r.prepare(self.p)
  (self.p/'resources/b.dmg').write_bytes(b'wrong')
  with self.assertRaises(ValueError):r.prepare(self.p)
 def test_workflow_scope_and_failclosed_chain(self):
  p=Path(os.environ.get('ORBIT_COMPOSER_WORKFLOW_SOURCE',str(Path(__file__).resolve().parents[2]/'.github/workflows/macos-composer-proof.yml')))
  s=p.resolve().read_text()
  for required in ['contents: read','actions: read',"github.event.before == 'ee25813f9a2065ae8632d4c15b24e3c99e7bf469'",'Run isolated macOS composer proof once (2026-10-26)','runs-on: macos-15','DEVELOPER_DIR: /Applications/Xcode_16.4.app/Contents/Developer','Build version 16F6','--require-hashes','--only-binary=:all:','hdiutil attach -readonly -nobrowse','ci_check_gates.py','ci_support.py collect']:
   self.assertIn(required,s)
  for forbidden in ['workflow_dispatch:','contents: write','actions: write','xattr -','codesign --force','CGRequest','xcode-select --switch','sudo ']:self.assertNotIn(forbidden,s)
  self.assertLess(s.index('ci_check_gates.py'),s.index('runner.py --execute-native'))
 def test_actual_workflow_predicate_event_matrix(self):
  p=Path(os.environ.get('ORBIT_COMPOSER_WORKFLOW_SOURCE',str(Path(__file__).resolve().parents[2]/'.github/workflows/macos-composer-proof.yml')))
  expression=next(line.strip()[4:] for line in p.read_text().splitlines() if line.startswith('    if: '))
  good={'github.repository':'charlzyx/orbit','github.event_name':'push','github.ref':'refs/heads/redesign/mygo-rpc','github.event.before':'ee25813f9a2065ae8632d4c15b24e3c99e7bf469','github.event.head_commit.message':'Run isolated macOS composer proof once (2026-10-26)'}
  def accepted(values):
   e=expression
   for key in sorted(values,key=len,reverse=True):e=e.replace(key,repr(values[key]))
   return eval(e.replace('&&',' and '),{'__builtins__':{}},{})
  self.assertTrue(accepted(good))
  for key in good:
   bad=dict(good);bad[key]='wrong';self.assertFalse(accepted(bad),key)
  for event in ('pull_request','workflow_dispatch','schedule'):
   bad=dict(good);bad['github.event_name']=event;self.assertFalse(accepted(bad))
