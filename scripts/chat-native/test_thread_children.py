"""Real proc/thread fixture; skips honestly on sandboxes without task children."""
import importlib.util
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import threading
import unittest

HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('thread_capture',HERE/'capture.py');c=importlib.util.module_from_spec(spec);spec.loader.exec_module(c)

class RealThreadChildren(unittest.TestCase):
    def test_native_jobs_require_real_thread_regression_before_capture(self):
        source=(HERE.parent.parent/'.github/workflows/go-native.yml').read_text()
        self.assertIn('  proc-guard-regression:',source)
        self.assertIn('ORBIT_REQUIRE_PROC_CHILDREN: "1"',source)
        self.assertIn("-p 'test_thread_children.py' -v",source)
        for job in ('reference-desktop','native-desktop'):
            self.assertIn('  '+job+':\n    needs: [proc-guard-regression, receipt-guard-regression]\n',source)

    def test_worker_thread_owned_rpc_child_is_found_but_unrelated_root_is_not(self):
        parent=os.getpid()
        if not Path(f'/proc/{parent}/task/{parent}/children').exists():
            if os.environ.get('GITHUB_ACTIONS')=='true' or os.environ.get('ORBIT_REQUIRE_PROC_CHILDREN')=='1':
                self.fail('Hosted regression requires /proc/PID/task/TID/children; cannot skip ownership evidence')
            self.skipTest('Kernel does not expose /proc/PID/task/TID/children; hosted Linux execution required')
        with tempfile.TemporaryDirectory() as d:
            script=Path(d)/'controlled_cli.py'
            script.write_text('import time\ntime.sleep(30)\n')
            ready=threading.Event();release=threading.Event();state={}
            def worker():
                try:
                    child=subprocess.Popen([sys.executable,str(script),'--mode','rpc'])
                    state.update(child=child,tid=threading.get_native_id())
                    ready.set();release.wait(10)
                except BaseException as error:state['error']=error;ready.set();return
                finally:
                    if 'child' in state:
                        child.terminate()
                        try:child.wait(timeout=3)
                        except subprocess.TimeoutExpired:child.kill();child.wait()
            thread=threading.Thread(target=worker);thread.start()
            try:
                self.assertTrue(ready.wait(5));self.assertNotIn('error',state)
                child=state['child'];self.assertNotEqual(state['tid'],parent)
                leader=Path(f'/proc/{parent}/task/{parent}/children').read_text().split()
                worker_ids=Path(f'/proc/{parent}/task/{state["tid"]}/children').read_text().split()
                self.assertNotIn(str(child.pid),leader,'Fixture must prove old main-thread-only walker misses the child')
                self.assertIn(str(child.pid),worker_ids)
                found=c.owned_rpc_children(parent,script.resolve())
                self.assertEqual([x['identity']['pid']for x in found],[child.pid])
                self.assertEqual(c.owned_rpc_children(child.pid,script.resolve()),[], 'The root itself is not its own descendant')
                self.assertEqual(c.owned_rpc_children(parent,Path(d)/'different_cli.py'),[])
            finally:release.set();thread.join(5)
            self.assertFalse(thread.is_alive())
            self.assertEqual(c.owned_rpc_children(parent,script.resolve()),[])

if __name__=='__main__':unittest.main()
