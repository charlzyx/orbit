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
        read=Path.read_bytes
        def changing(path):
            value=read(path)
            if path==self.proc/'200/cmdline':self.process(200,100,ticks='new-owner')
            return value
        with patch.object(Path,'read_bytes',changing):self.assertEqual(c.owned_rpc_children(100,self.pi),[])
    def test_thread_and_process_bounds_are_not_silently_truncated(self):
        self.task(100,101,' '.join(str(x)for x in range(200,265)))
        with self.assertRaisesRegex(ValueError,'process fanout'):c.owned_rpc_children(100,self.pi)
        self.task(100,101,'200')
        for tid in range(300,1325):self.task(100,tid,'')
        with self.assertRaisesRegex(ValueError,'thread fanout'):c.owned_rpc_children(100,self.pi)
