import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest
from unittest.mock import patch
import launch_receipt as r

class ReceiptGuards(unittest.TestCase):
    def setUp(self):
        self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup);self.root=Path(self.tmp.name)
        self.package=self.root/'package';self.package.mkdir();(self.package/'dist/bundle').mkdir(parents=True)
        (self.package/'package.json').write_text('{"name":"@earendil-works/pi-coding-agent","version":"1.4.0"}')
        (self.package/'dist/bundle/cli.js').write_text('fixture')
        self.node=self.root/'node';self.node.write_text('node fixture')
        self.fixture=self.root/'fixture';(self.fixture/'workspace').mkdir(parents=True)
        self.setup=r.prepare(self.package,self.node,self.fixture,self.root/'ledger','go',dict(files={k:v['sha256'] for k,v in r.inventory(self.package).items()}))
        self.verifier=r.Verifier(self.setup['expected'],self.setup['expectedSha256']);self.cfg=self.verifier.cfg
        self.ident=dict(pid=12,startTicks='42',exe=str(self.node));self.parent=dict(pid=1,startTicks='1',exe='/owned/app')
        self.rec=dict(schema=1,nonce=self.cfg['nonce'],expectedSha256=self.setup['expectedSha256'],identity=dict(pid=12,startTicks='42'),parent=self.parent,args=self.cfg['args'],node=str(self.node),entry=self.cfg['entry'],createdMonotonic=r.time.monotonic())
    def write(self,rec=None):
        p=self.root/'ledger/12-42.json';p.write_bytes(r.canonical(self.rec if rec is None else rec));p.chmod(0o600);return p
    def test_valid_record_and_missing_record(self):
        self.assertIsNone(self.verifier.candidate(self.ident,self.parent));self.write()
        self.assertIsNotNone(self.verifier.candidate(self.ident,self.parent))
        self.assertIsNone(self.verifier.candidate(dict(self.ident,exe='/wrapper'),self.parent))
        self.assertIsNone(self.verifier.candidate(dict(self.ident,startTicks='43'),self.parent))
    def test_foreign_partial_wrong_arguments_and_parent_rejected(self):
        for key,bad in [('nonce','foreign'),('expectedSha256','wrong'),('args',['--mode','rpc','--model','foreign']),('node','/other'),('entry','/other'),('parent',dict(self.parent,pid=2)),('identity',dict(pid=12,startTicks='old')),('createdMonotonic',float('nan')),('createdMonotonic',1e100)]:
            with self.subTest(key=key):
                self.write(dict(self.rec,**{key:bad}))
                with self.assertRaises(ValueError):self.verifier.candidate(self.ident,self.parent)
        self.write().write_text('{')
        with self.assertRaises(ValueError):self.verifier.candidate(self.ident,self.parent)
    def test_wrong_hash_and_mutated_pinned_file_never_reblessed(self):
        with self.assertRaises(ValueError):r.Verifier(self.setup['expected'],'wrong')
        self.write();self.verifier.check_files()
        (self.package/'dist/bundle/cli.js').write_text('changed')
        for _ in range(2):
            with self.assertRaises(ValueError):self.verifier.candidate(self.ident,self.parent)
    def test_symlink_oversized_and_fanout(self):
        p=self.write();p.unlink();p.symlink_to(self.package/'package.json')
        with self.assertRaises(OSError):self.verifier.candidate(self.ident,self.parent)
        p.unlink();self.write().write_text(' '*32769)
        with self.assertRaises(ValueError):self.verifier.candidate(self.ident,self.parent)
        p.unlink()
        for i in range(r.MAX_RECORDS+1):(self.root/'ledger'/str(i)).write_text('x')
        with self.assertRaises(ValueError):self.verifier.check_files()
    def test_no_reuse_and_exact_args_before_exec(self):
        with self.assertRaises(ValueError):r.prepare(self.package,self.node,self.fixture,self.root/'ledger','go',dict(files={k:v['sha256'] for k,v in r.inventory(self.package).items()}))
        with patch.object(r.os,'execve') as call:
            with self.assertRaises(ValueError):r.launch(self.setup['expected'],self.setup['expectedSha256'],['--mode','rpc'])
            call.assert_not_called()
    def test_version_probe_is_forwarded_without_rpc_record(self):
        with patch.object(r.os,'execve',side_effect=SystemExit) as call:
            with self.assertRaises(SystemExit):r.launch(self.setup['expected'],self.setup['expectedSha256'],['--version'])
        self.assertEqual(call.call_args.args[1],[str(self.node),self.cfg['entry'],'--version'])
        self.assertEqual(len(list((self.root/'ledger').glob('*.json'))),1)

    def test_same_bytes_stat_rewrite_is_bounded_not_new_content(self):
        self.write();self.verifier.check_files();path=self.package/'dist/bundle/cli.js'
        original=path.read_bytes();path.write_bytes(original)
        self.assertIsNotNone(self.verifier.candidate(self.ident,self.parent))
        self.assertEqual(self.verifier.changed_files,1)
        self.verifier.changed_files=32;path.write_bytes(original)
        with self.assertRaises(ValueError):self.verifier.candidate(self.ident,self.parent)
    def test_new_file_or_changed_launcher_refused(self):
        extra=self.package/'unexpected.js';extra.write_text('x')
        with self.assertRaises(ValueError):self.verifier.check_files()
        extra.unlink();(self.package/r.LAUNCHER).write_text('replacement')
        with self.assertRaises(ValueError):self.verifier.check_files()
    def test_exec_failure_cannot_create_accepted_live_node(self):
        with patch.object(r.os,'getcwd',return_value=self.cfg['cwd']),patch.object(r.os,'execve',side_effect=OSError('exec failed')):
            with self.assertRaises(OSError):r.launch(self.setup['expected'],self.setup['expectedSha256'],self.cfg['args'])
        self.assertIsNone(self.verifier.candidate(r.identity(os.getpid()),r.identity(os.getppid())))
    def test_package_and_ledger_directory_replacements_rejected(self):
        replacement=self.root/'replacement';shutil.copytree(self.package,replacement)
        self.package.rename(self.root/'old-package');replacement.rename(self.package)
        with self.assertRaisesRegex(ValueError,'directory identity'):self.verifier.check_files()
    def test_node_and_module_identity_mutations_rejected(self):
        self.node.write_text('wrong Node')
        with self.assertRaisesRegex(ValueError,'Node changed'):self.verifier.check_files()
        # Module is production source, never mutate it: substitute the recorded
        # expected stamp in this unit verifier and exercise the real check.
        self.verifier.cfg['nodeFile']=r.file_record(self.node)
        self.verifier.cfg['moduleFile']['stat']=[0,0,0,0,0]
        with self.assertRaisesRegex(ValueError,'module changed'):self.verifier.check_files()
    def test_independent_reference_rejects_same_version_altered_package(self):
        approved=dict(files={key:value['sha256'] for key,value in self.cfg['files'].items()})
        (self.package/r.LAUNCHER).unlink()
        (self.package/'dist/bundle/cli.js').write_text('tampered but still version 1.4.0')
        with self.assertRaisesRegex(ValueError,'independent vendor'):
            r.prepare(self.package,self.node,self.fixture,self.root/'new-ledger','go',approved)
    def test_ledger_directory_replacement_rejected(self):
        ledger=self.root/'ledger';replacement=self.root/'ledger-copy';shutil.copytree(ledger,replacement)
        ledger.rename(self.root/'old-ledger');replacement.rename(ledger)
        with self.assertRaisesRegex(ValueError,'directory identity'):self.verifier.check_files()
