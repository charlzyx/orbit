"""Deterministic proc fixtures supplement, not replace, the real thread test."""
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('proc_capture',HERE/'capture.py');c=importlib.util.module_from_spec(spec);spec.loader.exec_module(c)

class ProcScanner(unittest.TestCase):
    def setUp(self):
        self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup)
        self.root=Path(self.tmp.name);self.pi=self.root/'pi.py';self.pi.write_text('fixture')
        self.exe=self.root/'node';self.exe.write_text('fixture executable')
        self.proc=self.root/'proc';self.proc.mkdir()
        self.process(100,1);self.process(200,100)
        (self.proc/'100/cmdline').write_bytes(b'owned-native-app\0')
        self.task(100,100,'');self.task(100,101,'200');self.task(200,200,'')
        def path(value):
            value=str(value)
            return self.proc/value.removeprefix('/proc').lstrip('/') if value=='/proc' or value.startswith('/proc/') else Path(value)
        self.patch=patch.object(c,'Path',side_effect=path);self.patch.start();self.addCleanup(self.patch.stop)
    def process(self,pid,parent,ticks='10',cli=None):
        p=self.proc/str(pid);p.mkdir(exist_ok=True)
        fields=['S',str(parent)]+['0']*17+[ticks]
        (p/'stat').write_text(f'{pid} (fixture) '+' '.join(fields))
        if not (p/'exe').exists():(p/'exe').symlink_to(self.exe)
        (p/'cmdline').write_bytes(('node\0'+str(cli or self.pi)+'\0--mode\0rpc\0').encode())
    def task(self,pid,tid,children):
        p=self.proc/str(pid)/'task'/str(tid);p.mkdir(parents=True,exist_ok=True);(p/'children').write_text(children)
    def test_nonleader_thread_and_ancestry_with_exact_pinned_cli(self):
        out=c.owned_rpc_children(100,self.pi)
        self.assertEqual([x['identity']['pid']for x in out],[200])
        self.assertEqual([x['pid']for x in out[0]['ancestry']],[100,200])
        self.assertEqual(c.owned_rpc_children(100,self.root/'wrong_pi'),[])
    def test_exited_thread_child_and_wrong_parent_fail_closed(self):
        (self.proc/'100/task/102').mkdir();self.task(100,103,'999')
        self.process(200,777)
        self.assertEqual(c.owned_rpc_children(100,self.pi),[])
    def test_child_identity_changed_during_cmdline_read_is_rejected(self):
        read=c.read_cmdline
        def changing(pid):
            value=read(pid)
            if pid==200:self.process(200,100,ticks='new-owner')
            return value
        with patch.object(c,'read_cmdline',changing):self.assertEqual(c.owned_rpc_children(100,self.pi),[])
    def test_thread_and_process_bounds_are_not_silently_truncated(self):
        self.task(100,101,' '.join(str(x)for x in range(200,265)))
        with self.assertRaisesRegex(ValueError,'process fanout'):c.owned_rpc_children(100,self.pi)
        self.task(100,101,'200')
        for tid in range(300,1325):self.task(100,tid,'')
        with self.assertRaisesRegex(ValueError,'thread fanout'):c.owned_rpc_children(100,self.pi)

    def test_oversized_owned_argv_diagnostics_refused(self):
        (self.proc/'200/cmdline').write_bytes(b'x'*32769)
        with self.assertRaisesRegex(ValueError,'diagnostic bound'):c.owned_rpc_children(100,self.pi)
    def test_receipt_accepts_rewritten_title_only_with_full_owned_identity(self):
        import json
        r=c.receipt_module
        package=self.root/'package';(package/'dist/bundle').mkdir(parents=True)
        (package/'package.json').write_text('{"name":"@earendil-works/pi-coding-agent","version":"1.4.0"}')
        (package/'dist/bundle/cli.js').write_text('fixture')
        fixture=self.root/'fixture';(fixture/'workspace').mkdir(parents=True)
        setup=r.prepare(package,self.exe,fixture,self.root/'ledger','go',dict(files={k:v['sha256'] for k,v in r.inventory(package).items()}));verifier=r.Verifier(setup['expected'],setup['expectedSha256']);cfg=verifier.cfg
        (self.proc/'200/cmdline').write_bytes(b'pi\0')
        rejected=[]
        self.assertEqual(c.owned_rpc_children(100,self.pi,verifier,rejected),[])
        self.assertEqual(rejected[0]['reason'],'no matching launch receipt')
        receipt=dict(schema=1,nonce=cfg['nonce'],expectedSha256=setup['expectedSha256'],identity=dict(pid=200,startTicks='10'),parent=c.process_identity(100),args=cfg['args'],node=cfg['node'],entry=cfg['entry'],createdMonotonic=r.time.monotonic())
        target=self.root/'ledger/200-10.json';target.write_bytes(r.canonical(receipt));target.chmod(0o600)
        accepted=c.owned_rpc_children(100,self.pi,verifier)
        self.assertEqual([x['identity']['pid']for x in accepted],[200]);self.assertEqual(accepted[0]['argvDiagnostic'],['pi'])
        self.process(200,777)
        self.assertEqual(c.owned_rpc_children(100,self.pi,verifier),[])
        self.process(200,100);receipt['nonce']='foreign';target.write_bytes(r.canonical(receipt))
        with self.assertRaisesRegex(ValueError,'Foreign receipt'):c.owned_rpc_children(100,self.pi,verifier)
