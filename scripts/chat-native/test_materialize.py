import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('chat_materialize',HERE/'materialize.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)

class FixtureTests(unittest.TestCase):
    def test_same_bytes_fixed_chain_and_no_credentials(self):
        files,path=m.home_files();entries=[json.loads(line)for line in files[path].splitlines()]
        self.assertEqual(entries[0]['version'],3);parent=None
        for entry in entries[1:]:self.assertEqual(entry['parentId'],parent);parent=entry['id']
        messages=[e['message']for e in entries if e['type']=='message'];self.assertEqual(len(messages),24)
        self.assertEqual(messages[-2]['role'],'user');self.assertIn('CHAT TAIL READY',messages[-1]['content'][0]['text'])
        self.assertEqual(sum(x['usage']['totalTokens']for x in messages if x['role']=='assistant'),411000)
        self.assertNotIn('auth.json',' '.join(files));self.assertFalse(json.loads(files['.orbit-pi/auto-title.json'])['enabled'])
        with tempfile.TemporaryDirectory() as d:
            p=Path(d);a=m.materialize(p/'one',p/'one.json');b=m.materialize(p/'two',p/'two.json');self.assertEqual(a,b)
            self.assertEqual(a,json.loads((HERE.parent.parent/'testdata/chat-native/manifest.json').read_text()))
            self.assertTrue(m.verify(p/'one',a)['integrityAccepted'])
    def test_existing_and_symlink_roots_refused(self):
        with tempfile.TemporaryDirectory()as d:
            p=Path(d);(p/'old').mkdir();(p/'old/keep').write_text('keep')
            with self.assertRaises(ValueError):m.materialize(p/'old',p/'out.json')
            self.assertEqual((p/'old/keep').read_text(),'keep')
            (p/'link').symlink_to(p/'old',target_is_directory=True)
            with self.assertRaises(ValueError):m.materialize(p/'link/new',p/'out.json')
    def test_only_validated_startup_metadata_may_append(self):
        with tempfile.TemporaryDirectory()as d:
            p=Path(d);before=m.materialize(p/'root',p/'before.json');_,relative=m.home_files();session=p/'root/home'/relative
            original=session.read_bytes();leaf=json.loads(original.splitlines()[-1])
            entry=dict(type='custom',customType='orbit:workflow',data={'mode':'build'},id='abcde123',parentId=leaf['id'],timestamp=m.STAMP)
            session.write_bytes(original+(json.dumps(entry)+'\n').encode());self.assertTrue(m.verify(p/'root',before)['integrityAccepted'])
            entry['type']='message';session.write_bytes(original+(json.dumps(entry)+'\n').encode());self.assertFalse(m.verify(p/'root',before)['integrityAccepted'])
    def test_preferences_or_oracle_mutation_rejected(self):
        with tempfile.TemporaryDirectory()as d:
            p=Path(d);before=m.materialize(p/'root',p/'before.json');(p/'root/home/.orbit-pi/ui.json').write_text('{}')
            self.assertFalse(m.verify(p/'root',before)['integrityAccepted'])
            before['files']={}
            with self.assertRaises(ValueError):m.verify(p/'root',before)

class CredentialIsolation(unittest.TestCase):
    def test_new_empty_pi_auth_file_allowed_nonempty_rejected(self):
        with tempfile.TemporaryDirectory()as d:
            p=Path(d);before=m.materialize(p/'root',p/'before.json');auth=p/'root/home/.pi/agent/auth.json'
            auth.write_text('{}');self.assertTrue(m.verify(p/'root',before)['integrityAccepted'])
            auth.write_text('{"fixture_marker":true}');self.assertFalse(m.verify(p/'root',before)['integrityAccepted'])
            auth.unlink();auth.symlink_to(p/'missing');self.assertFalse(m.verify(p/'root',before)['integrityAccepted'])

if __name__=='__main__':unittest.main()
