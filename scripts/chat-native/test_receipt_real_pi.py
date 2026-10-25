"""Explicit offline integration: actual pinned Node/Pi, no model or inference."""
import json
import os
from pathlib import Path
import selectors
import shutil
import subprocess
import tempfile
import threading
import unittest
import importlib.util
import launch_receipt as r
import vendor_reference
HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('receipt_capture',HERE/'capture.py');capture=importlib.util.module_from_spec(spec);spec.loader.exec_module(capture)

class RealReceiptPi(unittest.TestCase):
    @unittest.skipUnless(os.getenv('ORBIT_RECEIPT_NODE') and os.getenv('ORBIT_RECEIPT_PI_RUNTIME'),'explicit pinned actual Node/Pi fixture paths required')
    def test_both_source_argument_vectors_real_rpc_survive_title_rewrite(self):
        node=Path(os.environ['ORBIT_RECEIPT_NODE']).resolve()
        self.assertEqual(subprocess.check_output([node,'--version'],text=True).strip(),'v22.23.4')
        repo=HERE.parent.parent
        rust=Path(os.environ.get('ORBIT_RECEIPT_RUST_SOURCE','/workspace/shared/orbit-rust-reference-1628'))
        for kind in ('go','reference'):
            with self.subTest(kind=kind),tempfile.TemporaryDirectory() as temp:
                root=Path(temp);runtime=root/'runtime';shutil.copytree(os.environ['ORBIT_RECEIPT_PI_RUNTIME'],runtime,symlinks=True)
                package=runtime/'node_modules/@earendil-works/pi-coding-agent'
                fixture=root/'fixture';(fixture/'workspace').mkdir(parents=True);home=fixture/'home';home.mkdir()
                for name in ('quota','guard','title','workflow'):
                    shutil.copytree(repo/'contrib'/f'orbit-{name}-extension',home/'.orbit-pi'/f'{name}-extension')
                source=repo if kind=='go' else rust
                patch_env=dict(HOME=str(home),PATH=str(node.parent)+':/usr/bin:/bin',PI_BIN=str(package/'dist/bundle/cli.js'),PI_PACKAGE_ROOT=str(package),PI_OFFLINE='1')
                patch_scripts=[source/'contrib'/f'pi-{name}-rpc/apply.mjs' for name in ('custom-ui','quota','auth')]
                for script in patch_scripts:subprocess.run([node,script],env=patch_env,check=True,capture_output=True,timeout=20)
                approved=vendor_reference.build(os.environ['ORBIT_RECEIPT_VENDOR_ARCHIVE'],node,source)
                setup=r.prepare(package,node,fixture,root/'ledger',kind,approved);v=r.Verifier(setup['expected'],setup['expectedSha256'])
                # Same source patch scripts, but with the package-root launcher.
                # Omit PI_PACKAGE_ROOT to exercise each real JS resolver too.
                patch_env['PI_BIN']=setup['launcher'];del patch_env['PI_PACKAGE_ROOT']
                for script in patch_scripts:subprocess.run([node,script],env=patch_env,check=True,capture_output=True,timeout=20)
                v.check_files() # second patch pass may rewrite stats, never expected bytes

                env=dict(HOME=str(home),PATH=str(node.parent)+':/usr/bin:/bin',PI_CODING_AGENT_DIR=str(home/'.pi/agent'),PI_OFFLINE='1',PI_SKIP_VERSION_CHECK='1',LANG='C.UTF-8')
                ready=threading.Event();release=threading.Event();state={}
                def spawn_worker():
                    try:state['process']=subprocess.Popen([setup['launcher'],*v.cfg['args']],cwd=fixture/'workspace',env=env,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
                    except BaseException as error:state['error']=error
                    finally:ready.set()
                    release.wait(40)
                worker=threading.Thread(target=spawn_worker);worker.start()
                self.assertTrue(ready.wait(10))
                if 'error' in state:release.set();worker.join();raise state['error']
                p=state['process']
                try:
                    p.stdin.write(b'{"id":"receipt-proof","type":"get_state"}\n');p.stdin.flush()
                    selector=selectors.DefaultSelector();selector.register(p.stdout,selectors.EVENT_READ)
                    self.assertTrue(selector.select(20),'Real Pi did not respond')
                    reply=json.loads(p.stdout.readline());self.assertEqual(reply.get('command'),'get_state');self.assertTrue(reply.get('success'))
                    actual=[x.decode()for x in Path(f'/proc/{p.pid}/cmdline').read_bytes().split(b'\0')if x]
                    self.assertEqual(actual,['pi'])
                    self.assertEqual(capture.owned_rpc_children(os.getpid(),Path(v.cfg['entry'])),[])
                    proof=v.candidate(capture.process_identity(p.pid),capture.process_identity(os.getpid()))
                    self.assertIsNotNone(proof)
                    task_children=Path(f'/proc/{os.getpid()}/task/{os.getpid()}/children').exists()
                    if task_children:
                        self.assertTrue(worker.is_alive(), 'Spawn thread must remain live during ownership scan')
                        self.assertEqual(capture.owned_rpc_children(p.pid,Path(v.cfg['entry']),v),[])
                        accepted=capture.owned_rpc_children(os.getpid(),Path(v.cfg['entry']),v)
                        self.assertEqual([x['identity']['pid'] for x in accepted],[p.pid])
                        self.assertEqual(accepted[0]['argvDiagnostic'],['pi'])
                    elif os.environ.get('ORBIT_REQUIRE_PROC_CHILDREN')=='1':
                        self.fail('Hosted real receipt ancestry proof requires task children')
                    print(json.dumps(dict(renderer=kind,nodeVersion='v22.23.4',entrySha256=v.cfg['files']['dist/bundle/cli.js']['sha256'],actualCmdline=actual,rpcSuccess=True,receiptCandidateAccepted=True,ancestryVerified=task_children,patchScriptsIdempotent=True,vendorIntegrity=approved['vendorIntegrity'],vendorArchiveSha256=approved['vendorArchiveSha256'],patchSourcesSha256=r.digest(r.canonical(approved['patchSources'])),expectedPackageContentSha256=r.digest(r.canonical(approved['files'])),packageFiles=len(approved['files']),nodeSha256=v.cfg['nodeFile']['sha256'],launchArgs=v.cfg['args']),sort_keys=True),flush=True)
                finally:
                    p.terminate()
                    try:out,err=p.communicate(timeout=5)
                    except subprocess.TimeoutExpired:p.kill();out,err=p.communicate()
                    if p.returncode not in (0,-15):print(err.decode(errors='replace'),flush=True)
                    release.set();worker.join(timeout=5);self.assertFalse(worker.is_alive())
                self.assertEqual(capture.owned_rpc_children(os.getpid(),Path(v.cfg['entry']),v),[])

    def test_real_ancestry_platform_capability_required(self):
        if not Path(f'/proc/{os.getpid()}/task/{os.getpid()}/children').exists():
            if os.environ.get('ORBIT_REQUIRE_PROC_CHILDREN')=='1':self.fail('Hosted task children missing')
            self.skipTest('Local sandbox does not expose task children; complete ancestry remains unverified')
