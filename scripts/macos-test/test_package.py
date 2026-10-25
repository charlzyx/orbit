#!/usr/bin/env python3
"""Offline contract tests. No remote writes and no native macOS execution."""
import unittest,tempfile,json,hashlib,os,subprocess,importlib.util
from pathlib import Path
from unittest.mock import patch
import package
spec=importlib.util.spec_from_file_location('publisher',Path(__file__).with_name('publish-prerelease.py'));pub=importlib.util.module_from_spec(spec);spec.loader.exec_module(pub)
HERE=Path(__file__).resolve().parent
class PackagingTests(unittest.TestCase):
 def test_versions_test_only(self):
  for v in ['0.2.12-go-test.20261025.1','0.2.12-go-test.20261025.2']:self.assertTrue(package.VERSION_RE.fullmatch(v))
  for v in ['v1','0.2.12','latest','stable','0.2.12-go-test.20261025.1\nBAD=1','../../bad']:self.assertFalse(package.VERSION_RE.fullmatch(v))
 def test_fingerprint_tracks_runtime_not_docs(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);(r/'main.go').write_text('package main\n');(r/'README.md').write_text('one')
   _,a=package.runtime_inputs(r);(r/'README.md').write_text('two');_,b=package.runtime_inputs(r);self.assertEqual(a,b)
   (r/'main.go').write_text('package other\n');_,c=package.runtime_inputs(r);self.assertNotEqual(a,c)
 def test_license_manifest(self):
  for p,h in json.loads((HERE/'licenses-sha256.json').read_text()).items():self.assertEqual(package.digest(HERE/'third-party-licenses'/p),h)
 def test_shell_syntax(self):
  for p in [HERE/'launcher.sh.in',HERE/'ci-build.sh']:subprocess.run(['bash','-n',str(p)],check=True)
 def fixture_assets(self,r):
  commit='a'*40;version='0.2.12-go-test.20261025.1';assets=[]
  for a in ('arm64','amd64'):
   p=r/f'Orbit-Go-Test-{version}-{commit[:12]}-ci-macos-{a}.zip';p.write_bytes(b'synthetic archive fixture '+a.encode());assets.append({'filename':p.name,'bytes':p.stat().st_size,'sha256':package.digest(p),'architecture':a})
  (r/'KNOWN-LIMITS.txt').write_text('Native Mac NOT RUN')
  (r/'RELEASE-NOTES.md').write_text('Synthetic offline test only')
  (r/'release-manifest.json').write_text(json.dumps({'version':version,'tag':'macos-test-v'+version,'source_commit':commit,'assets':assets}))
  paths=[r/a['filename'] for a in assets]+[r/'KNOWN-LIMITS.txt',r/'release-manifest.json']
  (r/'SHA256SUMS').write_text(''.join(f'{package.digest(p)}  {p.name}\n' for p in paths));return commit,version
 def test_local_asset_integrity_fail_closed(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);commit,version=self.fixture_assets(r);good=pub.local_assets(r,commit,version);self.assertEqual(len(good),5)
   with self.assertRaises(ValueError):pub.local_assets(r,'b'*40,version)
   with self.assertRaises(ValueError):pub.local_assets(r,commit,'0.2.12-go-test.20261025.2')
   (r/'KNOWN-LIMITS.txt').write_text('tamper')
   with self.assertRaises(ValueError):pub.local_assets(r,commit,version)
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
  text=(HERE.parents[1]/'.github/workflows/macos-test-prerelease.yml').read_text()
  self.assertEqual(text.count('contents: write'),1);self.assertIn('paths: [.github/macos-test-prerelease.json]',text);self.assertNotIn('workflow_dispatch:',text);self.assertNotIn('pull_request:',text);self.assertIn('needs: validate-and-package',text);self.assertEqual(text.count('GH_TOKEN:'),1)
 def test_publisher_no_destructive_flags(self):
  text=(HERE/'publish-prerelease.py').read_text()
  for flag in ['--clobber','--force','--latest=true']:
   # Only command argument occurrences, not prose prohibitions.
   self.assertNotIn("'"+flag+"'",text)
  self.assertIn("'make_latest':'false'",text);self.assertIn("'prerelease':True",text)
 def test_simulated_launcher_isolation(self):
  subprocess.run([os.sys.executable,str(HERE/'test_launcher.py')],check=True)
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
   with self.assertRaisesRegex(RuntimeError,'64 MiB'):pub.download_asset('charlzyx/orbit',{'id':1,'name':'large.zip','size':64*1024*1024+1},Path('unused'),'0'*64)
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
if __name__=='__main__':unittest.main(verbosity=2)
