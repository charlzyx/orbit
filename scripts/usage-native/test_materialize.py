#!/usr/bin/env python3
import hashlib, importlib.util, json, pathlib, subprocess, sys, tempfile, unittest
sys.dont_write_bytecode=True
HERE=pathlib.Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('materialize',HERE/'materialize.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
class FixtureTest(unittest.TestCase):
    def test_committed_inputs_and_materialized_inputs_are_identical(self):
        files,manifest=module.fixture()
        root=HERE.parents[1]/'testdata/usage-native'
        self.assertEqual(json.loads((root/'manifest.json').read_text()),manifest)
        for name,text in files.items():self.assertEqual((root/'home'/name).read_text(),text)
        with tempfile.TemporaryDirectory(prefix='orbit-usage-materializer-') as tmp:
            home=pathlib.Path(tmp)/'home';out=pathlib.Path(tmp)/'manifest.json'
            command=[sys.executable,str(HERE/'materialize.py'),str(home),'--manifest',str(out)]
            subprocess.run(command,check=True,capture_output=True)
            for name,want in manifest['files'].items():self.assertEqual(hashlib.sha256((home/name).read_bytes()).hexdigest(),want)
            before={name:(home/name).read_bytes() for name in files}
            refused=subprocess.run(command,capture_output=True,text=True)
            self.assertNotEqual(refused.returncode,0);self.assertIn('Refusing to overwrite',refused.stderr)
            self.assertEqual(before,{name:(home/name).read_bytes() for name in files})
if __name__=='__main__':unittest.main()
