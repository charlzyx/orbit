#!/usr/bin/env python3
"""Offline contract tests. No remote writes and no native macOS execution."""
import unittest,tempfile,json,hashlib,os,subprocess,importlib.util
from pathlib import Path
from common import LEGACY
import bundle,smoke
from unittest.mock import patch
import common as package
spec=importlib.util.spec_from_file_location('publisher',Path(__file__).with_name('publish-prerelease.py'));pub=importlib.util.module_from_spec(spec);spec.loader.exec_module(pub)
spec=importlib.util.spec_from_file_location('assembler',Path(__file__).with_name('assemble-release.py'));assembler=importlib.util.module_from_spec(spec);spec.loader.exec_module(assembler)
HERE=Path(__file__).resolve().parent
class PackagingTests(unittest.TestCase):
 def test_diagnostics_require_fresh_private_allowlisted_stages(self):
  with tempfile.TemporaryDirectory() as d:
   p=Path(d)/'app.log';version='0.2.12-go-test.20261025.4'
   rows=[{'timestamp':'2026-10-25T00:00:00Z','run_id':'a'*32,'version':version,'event':event,'result':result,'duration_ms':0} for event,result in [('app_startup','started'),('controller_init','ok'),('native_run','ok'),('app_shutdown','ok')]]
   def save(data):p.write_text(''.join(json.dumps(row)+'\n' for row in data))
   save(rows);self.assertEqual(len(smoke.verify_diagnostics(p,version,1)),4)
   with self.assertRaises(ValueError):smoke.verify_diagnostics(p,version,2)
   save(rows[:-1])
   with self.assertRaises(ValueError):smoke.verify_diagnostics(p,version,1)
   for field,value in [('private_path','/synthetic-secret'),('event','private text'),('version','wrong'),('run_id','bad')]:
    bad=[dict(row) for row in rows];bad[0][field]=value;save(bad)
    with self.assertRaises(ValueError):smoke.verify_diagnostics(p,version,1)
   save(rows);link=Path(d)/'linked';link.symlink_to(p)
   with self.assertRaises(ValueError):smoke.verify_diagnostics(link,version,1)
 def test_versions_test_only(self):
  for v in ['0.2.12-go-test.20261025.1','0.2.12-go-test.20261025.2']:self.assertTrue(package.VERSION_RE.fullmatch(v))
  for v in ['v1','0.2.12','latest','stable','0.2.12-go-test.20261025.1\nBAD=1','../../bad']:self.assertFalse(package.VERSION_RE.fullmatch(v))
 def test_fingerprint_tracks_runtime_not_docs(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);(r/'main.go').write_text('package main\n');(r/'README.md').write_text('one')
   _,a=package.runtime_inputs(r);(r/'README.md').write_text('two');_,b=package.runtime_inputs(r);self.assertEqual(a,b)
   (r/'main.go').write_text('package other\n');_,c=package.runtime_inputs(r);self.assertNotEqual(a,c)
 def test_license_manifest(self):
  for p,h in json.loads((LEGACY/'licenses-sha256.json').read_text()).items():self.assertEqual(package.digest(LEGACY/'third-party-licenses'/p),h)
 def test_shell_syntax(self):
  for p in [HERE/'ci-validate.sh']:subprocess.run(['bash','-n',str(p)],check=True)
 def fixture_assets(self,r):
  commit='a'*40;version='0.2.12-go-test.20261025.2';assets=[]
  for a in ('arm64',):
   p=r/f'Orbit-Go-Test-{version}-{commit[:12]}-ci-macos-{a}.dmg';p.write_bytes(b'synthetic archive fixture '+a.encode());assets.append({'filename':p.name,'bytes':p.stat().st_size,'sha256':package.digest(p),'architecture':a})
  (r/'KNOWN-LIMITS.txt').write_text('Native Mac NOT RUN')
  (r/'RELEASE-NOTES.md').write_text('Synthetic offline test only')
  (r/'release-manifest.json').write_text(json.dumps({'version':version,'tag':'macos-test-v'+version,'source_commit':commit,'assets':assets,'native_macos_smoke':{'arm64':'PASS'}}))
  paths=[r/a['filename'] for a in assets]+[r/'KNOWN-LIMITS.txt',r/'release-manifest.json']
  (r/'SHA256SUMS').write_text(''.join(f'{package.digest(p)}  {p.name}\n' for p in paths));return commit,version
 def test_local_asset_integrity_fail_closed(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);commit,version=self.fixture_assets(r);good=pub.local_assets(r,commit,version);self.assertEqual(len(good),4)
   with self.assertRaises(ValueError):pub.local_assets(r,'b'*40,version)
   with self.assertRaises(ValueError):pub.local_assets(r,commit,'0.2.12-go-test.20261025.3')
   (r/'KNOWN-LIMITS.txt').write_text('tamper')
   with self.assertRaises(ValueError):pub.local_assets(r,commit,version)
 def test_arm64_only_assembly_accepts_one_native_passed_dmg(self):
  import contextlib,io
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);fixture=r/'fixture';fixture.mkdir();commit,version=self.fixture_assets(fixture)
   manifest=json.loads((fixture/'release-manifest.json').read_text());asset=manifest['assets'][0]
   asset.update({'source_commit':commit,'version':version,'native_macos_smoke':'PASS','ci_run':'https://github.com/charlzyx/orbit/actions/runs/123'})
   inputs=r/'inputs';arch=inputs/'arm64';arch.mkdir(parents=True)
   (arch/asset['filename']).write_bytes((fixture/asset['filename']).read_bytes())
   package_json=arch/'package-arm64.json';package_json.write_text(json.dumps(asset))
   marker={'version':version,'runtime_source_sha256':'b'*64}
   with patch.object(assembler,'check_source',return_value=(marker,{})),patch.dict(os.environ,{'GITHUB_REPOSITORY':'charlzyx/orbit','GITHUB_RUN_ID':'123'}),contextlib.redirect_stdout(io.StringIO()):
    output=r/'release';assembler.assemble(r,inputs,output,commit)
    self.assertEqual(len(pub.local_assets(output,commit,version)),4)
    result=json.loads((output/'release-manifest.json').read_text())
    self.assertEqual(result['native_macos_smoke'],{'arm64':'PASS'})
    self.assertEqual([a['architecture'] for a in result['assets']],['arm64'])
    self.assertEqual(len(list(output.glob('*.dmg'))),1)
    self.assertIn('arm64 only',(output/'RELEASE-NOTES.md').read_text())
    for key,value in [('native_macos_smoke','NOT_RUN'),('source_commit','c'*40),('ci_run','https://github.com/charlzyx/orbit/actions/runs/456')]:
     bad=dict(asset);bad[key]=value;package_json.write_text(json.dumps(bad))
     with self.assertRaises(ValueError):assembler.assemble(r,inputs,r/key,commit)
    package_json.write_text(json.dumps(asset));(inputs/'amd64').mkdir()
    with self.assertRaisesRegex(ValueError,'architecture input'):assembler.assemble(r,inputs,r/'intel',commit)
 def test_publisher_rejects_missing_extra_and_mislabeled_architectures(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);commit,version=self.fixture_assets(r);p=r/'release-manifest.json';m=json.loads(p.read_text());arm=m['assets'][0]
   for assets in [[],[arm,dict(arm,architecture='amd64',filename=arm['filename'].replace('arm64','amd64'))],[dict(arm,architecture='amd64')]]:
    bad=dict(m,assets=assets);p.write_text(json.dumps(bad))
    with self.assertRaisesRegex(ValueError,'architecture asset'):pub.local_assets(r,commit,version)
 def test_future_build_entrypoints_reject_intel_before_work(self):
  self.assertEqual(package.ARCHITECTURES,('arm64',))
  with patch.object(bundle,'check_source',side_effect=AssertionError('Must not read source')):
   with self.assertRaisesRegex(ValueError,'arm64 only'):bundle.build(None,None,None,None,None,'amd64',None)
  with self.assertRaisesRegex(ValueError,'arm64 only'):smoke.smoke(Path('unused'),'amd64')
  with tempfile.TemporaryDirectory() as d:
   output=Path(d)/'output'
   result=subprocess.run([os.sys.executable,str(HERE/'runtime/prepare-runtime.py'),'--arch','amd64','--output',str(output)],text=True,capture_output=True)
   self.assertEqual(result.returncode,2);self.assertIn('invalid choice',result.stderr);self.assertFalse(output.exists())
 def test_five_cross_build_validation_gate(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);validation=r/'validation.json';commit='a'*40
   stages=['root-race','vet','toolkit','terminal','wrapper','packaging-tests']
   with patch.object(bundle.platform,'system',return_value='Darwin'),patch.object(bundle.platform,'machine',return_value='arm64'),patch.object(bundle,'check_source',return_value=({},{})),patch.object(bundle,'verify_runtime',return_value={}):
    for stage,error in [('six-cross-builds','Exact aggregate'),('five-cross-builds','Fresh output')]:
     validation.write_text(json.dumps({'source_commit':commit,'status':'PASS','passed_stages':stages+[stage]}))
     with self.assertRaisesRegex(ValueError,error):bundle.build(r,r,r,r,commit,'arm64',validation)
 def test_remote_mismatch_requires_no_download(self):
  with patch.object(pub.subprocess,'run',side_effect=AssertionError('No commands permitted')):
   for release in [{'prerelease':False,'target_commitish':'a'*40},{'prerelease':True,'target_commitish':'b'*40},{'prerelease':True,'target_commitish':'a'*40,'draft':True}]:
    with self.assertRaises(RuntimeError):pub.verify_release('charlzyx/orbit','test',release,{},'a'*40,True)
 def test_existing_tag_never_mutated(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);commit,version=self.fixture_assets(r)
   with patch.dict(os.environ,{'GITHUB_EVENT_NAME':'push','GITHUB_REF':'refs/heads/redesign/mygo-rpc','GITHUB_REPOSITORY':'charlzyx/orbit','GITHUB_SHA':commit}),patch.object(pub,'git',side_effect=[commit,'']),patch.object(pub,'marker',return_value={'version':version}),patch.object(pub,'find_release',return_value=None),patch.object(pub,'optional_ref',return_value={'object':{'type':'commit','sha':commit}}),patch.object(pub.subprocess,'run',side_effect=AssertionError('Must not mutate')):
    with self.assertRaisesRegex(RuntimeError,'Tag already exists'):pub.main(r,r)
 def test_wrong_event_never_reads_or_mutates_remote(self):
  with patch.dict(os.environ,{'GITHUB_EVENT_NAME':'pull_request','GITHUB_REF':'refs/heads/main'}),patch.object(pub,'command',side_effect=AssertionError('Must not query')):
   with self.assertRaisesRegex(RuntimeError,'Wrong event'):pub.main(Path('.'),Path('.'))
 def test_completed_exact_release_is_read_only(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);commit,version=self.fixture_assets(r);release={'id':1,'html_url':'https://github.com/charlzyx/orbit/releases/tag/test'}
   with patch.dict(os.environ,{'GITHUB_EVENT_NAME':'push','GITHUB_REF':'refs/heads/redesign/mygo-rpc','GITHUB_REPOSITORY':'charlzyx/orbit','GITHUB_SHA':commit}),patch.object(pub,'git',side_effect=[commit,'']),patch.object(pub,'marker',return_value={'version':version}),patch.object(pub,'find_release',return_value=release),patch.object(pub,'optional_ref',return_value={'object':{'type':'commit','sha':commit}}),patch.object(pub,'verify_release') as verify,patch.object(pub.subprocess,'run',side_effect=AssertionError('Must not mutate')):
    pub.main(r,r);verify.assert_called_once()
 def test_existing_release_wrong_commit_fails(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);commit,version=self.fixture_assets(r)
   with patch.dict(os.environ,{'GITHUB_EVENT_NAME':'push','GITHUB_REF':'refs/heads/redesign/mygo-rpc','GITHUB_REPOSITORY':'charlzyx/orbit','GITHUB_SHA':commit}),patch.object(pub,'git',side_effect=[commit,'']),patch.object(pub,'marker',return_value={'version':version}),patch.object(pub,'find_release',return_value={'id':1}),patch.object(pub,'optional_ref',return_value={'object':{'type':'commit','sha':'b'*40}}),patch.object(pub.subprocess,'run',side_effect=AssertionError('Must not mutate')):
    with self.assertRaisesRegex(RuntimeError,'tag/commit mismatch'):pub.main(r,r)
 def test_workflow_permissions_and_trigger(self):
  text=(HERE.parents[1]/'.github/workflows/macos-app-prerelease.yml').read_text()
  self.assertEqual(text.count('contents: write'),1);self.assertIn('paths: [.github/macos-app-prerelease.json]',text);self.assertNotIn('workflow_dispatch:',text);self.assertNotIn('pull_request:',text);self.assertIn('needs: [validate, native-app]',text);self.assertEqual(text.count('GH_TOKEN:'),1)
  self.assertIn('arch: arm64',text);self.assertNotIn('arch: amd64',text);self.assertNotIn('macos-15-intel',text);self.assertNotIn('macos-app-amd64-',text)
  script=(HERE/'ci-validate.sh').read_text()
  self.assertIn('for target in linux/amd64 linux/arm64 darwin/arm64 windows/amd64 windows/arm64; do',script)
  self.assertNotIn('darwin/amd64',script);self.assertIn("'five-cross-builds'",script)
 def test_publisher_no_destructive_flags(self):
  text=(HERE/'publish-prerelease.py').read_text()
  for flag in ['--clobber','--force','--latest=true']:
   # Only command argument occurrences, not prose prohibitions.
   self.assertNotIn("'"+flag+"'",text)
  self.assertIn("'make_latest':'false'",text);self.assertIn("'prerelease':True",text)
 def test_download_hash_and_redirect_auth(self):
  import io,urllib.request
  with tempfile.TemporaryDirectory() as d:
   data=b'synthetic';asset={'id':123,'name':'asset.zip','size':len(data)};h=hashlib.sha256(data).hexdigest()
   with patch.object(pub,'request',return_value=io.BytesIO(data)):pub.download_asset('charlzyx/orbit',asset,Path(d)/'asset.zip',h)
   with patch.object(pub,'request',return_value=io.BytesIO(b'tampered!')):
    with self.assertRaisesRegex(RuntimeError,'hash/size'):pub.download_asset('charlzyx/orbit',asset,Path(d)/'bad.zip',h)
   req=urllib.request.Request('https://api.github.com/repos/x/y/releases/assets/1',headers={'Authorization':'Bearer SYNTHETIC','Cookie':'SYNTHETIC'})
   moved=pub.SafeRedirect().redirect_request(req,None,302,'redirect',{},'https://release-assets.githubusercontent.com/asset')
   self.assertIsNone(moved.get_header('Authorization'));self.assertIsNone(moved.get_header('Cookie'))
 def test_oversized_remote_asset_rejected_before_network(self):
  with patch.object(pub,'request',side_effect=AssertionError('Must not download')):
   with self.assertRaisesRegex(RuntimeError,'512 MiB'):pub.download_asset('charlzyx/orbit',{'id':1,'name':'large.zip','size':512*1024*1024+1},Path('unused'),'0'*64)
 def lifecycle(self,mode):
  import contextlib,copy
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);commit,version=self.fixture_assets(r);expected=pub.local_assets(r,commit,version);tag='macos-test-v'+version;calls=[];uploaded=[];api_reads=0
   empty={'id':17,'tag_name':tag,'target_commitish':commit,'draft':True,'prerelease':True,'assets':[],'html_url':'https://github.com/charlzyx/orbit/releases/tag/'+tag}
   def mutation(method,path,payload):
    calls.append((method,path,payload))
    if method=='POST':return copy.deepcopy(empty)
    self.assertEqual(method,'PATCH');self.assertTrue(path.endswith('/releases/17'));self.assertEqual(payload,{'tag_name':tag,'target_commitish':commit,'draft':False,'prerelease':True,'make_latest':'false'});return {}
   def upload(repo,rid,path):
    self.assertEqual(rid,17)
    if mode=='upload-failure' and uploaded:raise RuntimeError('Synthetic upload failed')
    a={'id':100+len(uploaded),'name':path.name,'size':path.stat().st_size,'updated_at':'fixed','state':'uploaded'};uploaded.append(a);return dict(a)
   def read(path):
    nonlocal api_reads
    self.assertTrue(path.endswith('/releases/17'));api_reads+=1;snapshot=copy.deepcopy(empty);snapshot['assets']=copy.deepcopy(uploaded)
    if api_reads==2 and mode=='replacement':snapshot['id']=99
    if api_reads==2 and mode=='asset-tamper':snapshot['assets'][0]['id']=999
    if api_reads>=3:snapshot['draft']=False
    return snapshot
   def download(repo,asset,path,h):self.assertEqual(package.digest(r/asset['name']),h);path.write_bytes((r/asset['name']).read_bytes())
   refs=[None,None,{'object':{'type':'commit','sha':commit}}]
   with contextlib.ExitStack() as stack:
    for ctx in [patch.dict(os.environ,{'GITHUB_EVENT_NAME':'push','GITHUB_REF':'refs/heads/redesign/mygo-rpc','GITHUB_REPOSITORY':'charlzyx/orbit','GITHUB_SHA':commit}),patch.object(pub,'git',side_effect=[commit,'']),patch.object(pub,'marker',return_value={'version':version}),patch.object(pub,'find_release',return_value=None),patch.object(pub,'optional_ref',side_effect=refs),patch.object(pub,'rest_json',side_effect=mutation),patch.object(pub,'upload_asset',side_effect=upload),patch.object(pub,'api',side_effect=read),patch.object(pub,'download_asset',side_effect=download)]:stack.enter_context(ctx)
    if mode=='success':pub.main(r,r)
    else:
     with self.assertRaises(RuntimeError):pub.main(r,r)
   self.assertEqual([c[0] for c in calls],['POST','PATCH'] if mode=='success' else ['POST'])
   self.assertTrue(all('/releases/99' not in c[1] for c in calls))
 def test_new_release_success_by_owned_id(self):self.lifecycle('success')
 def test_replacement_draft_never_published(self):self.lifecycle('replacement')
 def test_asset_tamper_stops_before_publish(self):self.lifecycle('asset-tamper')
 def test_upload_failure_leaves_draft_without_retry(self):self.lifecycle('upload-failure')
 def test_native_gates_fail_closed(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);commit,version=self.fixture_assets(r)
   m=json.loads((r/'release-manifest.json').read_text());m['native_macos_smoke']['arm64']='NOT_RUN';(r/'release-manifest.json').write_text(json.dumps(m))
   with self.assertRaisesRegex(ValueError,'native macOS'):pub.local_assets(r,commit,version)
 def test_symlinks_are_relocation_safe(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);(r/'file').write_text('fixture');(r/'link').symlink_to('file');self.assertEqual(bundle.tree_hashes(r)['link'],{'symlink':'file'})
   (r/'link').unlink();(r/'link').symlink_to(r/'file')
   with self.assertRaisesRegex(ValueError,'absolute'):bundle.tree_hashes(r)
   (r/'link').unlink();(r/'link').symlink_to('../escape')
   with self.assertRaises(ValueError):bundle.tree_hashes(r)
 def test_marker_rejects_first_version_and_placeholder(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);(r/'.github').mkdir();p=r/package.MARKER
   m={'enabled':True,'version':'0.2.12-go-test.20261025.2','runtime_source_sha256':'a'*64,'minimum_macos':'13.0','required_native_smoke':True};p.write_text(json.dumps(m));self.assertEqual(package.marker(r),m)
   for key,value in [('version','0.2.12-go-test.20261025.1'),('runtime_source_sha256','PENDING'),('required_native_smoke',False)]:
    bad=dict(m);bad[key]=value;p.write_text(json.dumps(bad))
    with self.assertRaises(ValueError):package.marker(r)
 def test_native_smoke_report_requires_real_events_and_installed_paths(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);app=r/'Installed App.app';home=r/'User';private=home/'Library/Application Support/Orbit Go Test/Home'
   report={'status':'PASS','architecture':'arm64','pid':123,'started_at':__import__('datetime').datetime.now(__import__('datetime').timezone.utc).isoformat(),'session_id':'unique','home':str(private),'user_home':str(home),'pi_agent_dir':str(home/'.pi/agent'),'workspace':str(private/'Workspace'),'pi_path':str(private/'runtime/version/pi/node_modules/.bin/pi'),'checks':{k:True for k in ['native_window_ready','first_frame','ghostty_load','pi_rpc_get_state','native_capture']}}
   report_file=r/'report.json';capture=Path(str(report_file)+'.png');capture.write_bytes(b'\x89PNG\r\n\x1a\n'+__import__('struct').pack('>I',13)+b'IHDR'+__import__('struct').pack('>II',640,480)+bytes(9));report.update({'capture_kind':'native-content','capture_path':str(capture),'capture_width':640,'capture_height':480,'capture_sha256':package.digest(capture)})
   for k,v in {'executable':'Contents/MacOS/Orbit','ghostty_path':'Contents/Frameworks/libghostty-vt.dylib','node_path':'Contents/Resources/node/bin/node','bundled_pi_path':'Contents/Resources/bin/pi'}.items():report[k]=str(app/v)
   smoke.verify_report(report,app,home,'arm64',0,report_file)
   for k in report['checks']:
    report['checks'][k]=False
    with self.assertRaises(ValueError):smoke.verify_report(report,app,home,'arm64',0,report_file)
    report['checks'][k]=True
   for key in ['home','user_home','pi_agent_dir']:
    old=report[key];report[key]='/tmp/wrong/config'
    with self.assertRaises(ValueError):smoke.verify_report(report,app,home,'arm64',0,report_file)
    report[key]=old
   report['executable']='/tmp/wrong/Orbit'
   with self.assertRaises(ValueError):smoke.verify_report(report,app,home,'arm64',0,report_file)
 def test_default_finder_launch_contract(self):
  source=(HERE/'smoke.py').read_text()
  self.assertIn("'/usr/bin/open','-n','-W'",source);self.assertIn("'/usr/bin/hdiutil','detach'",source)
  self.assertNotIn("'--home'",source);self.assertNotIn("'--workspace'",source);self.assertNotIn("'--pi'",source)
  self.assertIn('for i in (1,2):',source);self.assertIn('tree_hashes(installed)!=expected',source)
  for prohibited in ['spctl --master-disable','xattr -d','codesign --force','launchctl setenv']:self.assertNotIn(prohibited,source)
 def test_deployment_minimum_rejects_newer_dependency(self):
  import struct
  with tempfile.TemporaryDirectory() as d:
   p=Path(d)/'native';p.write_bytes(struct.pack('<8I',0xfeedfacf,0x100000c,0,2,1,24,0,0)+struct.pack('<6I',0x32,24,1,(14<<16),0,0))
   with self.assertRaisesRegex(ValueError,'macOS13'):bundle.require_arch(p,'arm64')
   p.write_bytes(struct.pack('<8I',0xfeedfacf,0x100000c,0,2,1,24,0,0)+struct.pack('<6I',0x32,24,2,(11<<16),0,0))
   with self.assertRaisesRegex(ValueError,'Non-macOS'):bundle.require_arch(p,'arm64')
   p.write_bytes(struct.pack('<8I',0xfeedfacf,0x100000c,0,2,1,24,0,0)+struct.pack('<6I',0x32,24,1,(11<<16),0,0))
   self.assertEqual(bundle.require_arch(p,'arm64')['minimum_os'],'11.0.0')
 def test_audit_only_runtime_rejected(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);(r/'runtime-manifest.json').write_text(json.dumps({'execution_evidence':'not executed; existing-tree audit only'}))
   with self.assertRaisesRegex(ValueError,'Audit-only'):bundle.verify_runtime(r,'arm64')
 def test_runtime_safe_extraction(self):
  subprocess.run([os.sys.executable,str(HERE/'runtime/test_prepare_runtime.py')],check=True)
if __name__=='__main__':unittest.main(verbosity=2)

