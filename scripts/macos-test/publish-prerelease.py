#!/usr/bin/env python3
"""The ONLY remote mutation entry point. Must run only in guarded release job.
No replacement, tag force, asset clobber, deletion, stable/latest publication,
credential creation, settings changes, signing, or automatic retry is supported.
"""
from pathlib import Path
import argparse,hashlib,json,os,re,subprocess,tempfile,urllib.request,urllib.parse
from package import COMMIT_RE,VERSION_RE,digest,marker,git,ARCHITECTURES

def command(*args):return subprocess.check_output(['gh',*args],text=True)
def api(path):return json.loads(command('api',path))
class SafeRedirect(urllib.request.HTTPRedirectHandler):
 def redirect_request(self,req,fp,code,msg,headers,newurl):
  if not newurl.startswith('https://'):raise RuntimeError('Non-HTTPS GitHub asset redirect rejected')
  redirected=super().redirect_request(req,fp,code,msg,headers,newurl)
  if urllib.parse.urlparse(newurl).netloc!=urllib.parse.urlparse(req.full_url).netloc:
   redirected.remove_header('Authorization');redirected.remove_header('Cookie')
  return redirected

def request(method,url,data=None,accept='application/vnd.github+json',content_type='application/json'):
 host=urllib.parse.urlparse(url).hostname
 if host not in ('api.github.com','uploads.github.com') or not url.startswith('https://'):raise RuntimeError('Unexpected GitHub API destination')
 token=os.environ.get('GH_TOKEN')
 if not token:raise RuntimeError('Missing ephemeral Actions token')
 req=urllib.request.Request(url,data=data,method=method,headers={'Authorization':'Bearer '+token,'Accept':accept,'Content-Type':content_type,'X-GitHub-Api-Version':'2022-11-28','User-Agent':'orbit-test-prerelease'})
 return urllib.request.build_opener(SafeRedirect()).open(req,timeout=180)

def rest_json(method,path,payload):
 with request(method,'https://api.github.com/'+path,json.dumps(payload).encode()) as response:return json.load(response)

def upload_asset(repo,release_id,path):
 url=f'https://uploads.github.com/repos/{repo}/releases/{release_id}/assets?name='+urllib.parse.quote(path.name,safe='')
 with request('POST',url,path.read_bytes(),content_type='application/octet-stream') as response:return json.load(response)

def download_asset(repo,asset,path,expected):
 # Asset ID, not mutable tag/name resolution. Cross-origin redirects strip auth.
 if not isinstance(asset.get('size'),int) or not 0 <= asset['size'] <= 64*1024*1024:raise RuntimeError('Remote asset size exceeds the 64 MiB verification bound')
 with request('GET',f'https://api.github.com/repos/{repo}/releases/assets/{asset["id"]}',accept='application/octet-stream') as response:
  data=response.read(asset['size']+1)
 if len(data)!=asset['size'] or hashlib.sha256(data).hexdigest()!=expected:raise RuntimeError('Remote asset hash/size mismatch; no overwrite: '+asset['name'])
 path.write_bytes(data)

def optional_ref(repo,tag):
 r=subprocess.run(['gh','api',f'repos/{repo}/git/ref/tags/{tag}'],text=True,capture_output=True)
 if r.returncode==0:return json.loads(r.stdout)
 if 'HTTP 404' in r.stderr:return None
 raise RuntimeError('Tag lookup failed; no write is safe: '+r.stderr)
def resolve_ref(repo,ref):
 obj=ref['object']
 for _ in range(5):
  if obj['type']=='commit':return obj['sha']
  if obj['type']!='tag':raise RuntimeError('Unexpected tag object type')
  obj=api(f'repos/{repo}/git/tags/{obj["sha"]}')['object']
 raise RuntimeError('Excessively nested tag')
def find_release(repo,tag):
 pages=json.loads(command('api','--paginate','--slurp',f'repos/{repo}/releases?per_page=100'))
 matches=[r for page in pages for r in page if r['tag_name']==tag]
 if len(matches)>1:raise RuntimeError('Duplicate release tag records')
 return matches[0] if matches else None

def local_assets(folder,commit,version):
 m=json.loads((folder/'release-manifest.json').read_text())
 if m.get('source_commit')!=commit or m.get('version')!=version or m.get('tag')!='macos-test-v'+version:raise ValueError('Artifact/source/version mismatch')
 expected_zips={f'Orbit-Go-Test-{version}-{commit[:12]}-ci-macos-{a}.zip' for a in ARCHITECTURES}
 if {x['filename'] for x in m['assets']}!=expected_zips or len(m['assets'])!=len(ARCHITECTURES) or {x.get('architecture') for x in m['assets']}!=set(ARCHITECTURES):raise ValueError('Wrong architecture asset set')
 for a in m['assets']:
  p=folder/a['filename']
  if p.is_symlink() or p.stat().st_size!=a['bytes'] or digest(p)!=a['sha256']:raise ValueError('ZIP mismatch')
 expected=expected_zips|{'KNOWN-LIMITS.txt','release-manifest.json'}
 rows={}
 for line in (folder/'SHA256SUMS').read_text().splitlines():
  h,name=line.split('  ',1)
  if name in rows or name not in expected or not re.fullmatch('[0-9a-f]{64}',h):raise ValueError('Bad checksum manifest')
  rows[name]=h
 if set(rows)!=expected:raise ValueError('Incomplete checksum manifest')
 for name,h in rows.items():
  p=folder/name
  if p.is_symlink() or digest(p)!=h:raise ValueError('Asset checksum mismatch')
 rows['SHA256SUMS']=digest(folder/'SHA256SUMS')
 return rows

def verify_release(repo,tag,release,expected,commit,require_published):
 if not release.get('prerelease') or release.get('target_commitish')!=commit:raise RuntimeError('Existing release is not the exact requested prerelease/commit')
 if require_published and release.get('draft'):raise RuntimeError('Existing draft is incomplete; never overwrite/resume automatically')
 assets=release.get('assets',[])
 if len(assets)!=len(expected) or {a['name'] for a in assets}!=set(expected):raise RuntimeError('Release asset set differs; never overwrite')
 with tempfile.TemporaryDirectory(prefix='orbit-release-verify-') as temp:
  for name,h in expected.items():
   asset=next(a for a in assets if a['name']==name)
   download_asset(repo,asset,Path(temp)/name,h)

def main(root,folder):
 if os.environ.get('GITHUB_EVENT_NAME')!='push' or os.environ.get('GITHUB_REF')!='refs/heads/redesign/mygo-rpc':raise RuntimeError('Wrong event/branch')
 repo=os.environ.get('GITHUB_REPOSITORY','');commit=os.environ.get('GITHUB_SHA','')
 if repo!='charlzyx/orbit' or not COMMIT_RE.fullmatch(commit):raise RuntimeError('Only the authorized charlzyx/orbit repository and an exact commit are allowed')
 if git(root,'rev-parse','HEAD')!=commit or git(root,'status','--porcelain','--untracked-files=no'):raise RuntimeError('Checkout mismatch/dirty')
 m=marker(root);version=m['version'];tag='macos-test-v'+version;expected=local_assets(folder,commit,version)
 existing=find_release(repo,tag);ref=optional_ref(repo,tag)
 if existing:
  if not ref or resolve_ref(repo,ref)!=commit:raise RuntimeError('Existing tag/commit mismatch')
  verify_release(repo,tag,existing,expected,commit,True)
  print('Exact prerelease and hashes already exist; no writes:',existing['html_url']);return
 if ref:raise RuntimeError('Tag already exists without matching complete release; refusing to modify it')
 # Create once and retain the exact ID returned by GitHub. Never resolve
 # a mutable tag to choose the release to upload into or publish.
 created=rest_json('POST',f'repos/{repo}/releases',{'tag_name':tag,'target_commitish':commit,'name':'Orbit Go macOS test '+version,'body':(folder/'RELEASE-NOTES.md').read_text(),'draft':True,'prerelease':True,'make_latest':'false'})
 if not created.get('draft') or not created.get('prerelease') or created.get('tag_name')!=tag or created.get('target_commitish')!=commit or created.get('assets'):raise RuntimeError('New empty exact-commit draft could not be established; stop without retry')
 created_id=created['id'];uploaded_ids={}
 for name in expected:
  asset=upload_asset(repo,created_id,folder/name)
  if asset.get('name')!=name or asset.get('size')!=(folder/name).stat().st_size or asset.get('state')!='uploaded':raise RuntimeError('Upload outcome mismatch; inspect exact draft before retry')
  uploaded_ids[name]=asset['id']
 created=api(f'repos/{repo}/releases/{created_id}')
 if {a['name']:a['id'] for a in created['assets']}!=uploaded_ids:raise RuntimeError('Draft asset IDs changed')
 verify_release(repo,tag,created,expected,commit,False)
 check=api(f'repos/{repo}/releases/{created_id}')
 if check['id']!=created_id or not check.get('draft') or not check.get('prerelease') or check.get('tag_name')!=tag or check.get('target_commitish')!=commit:raise RuntimeError('Owned draft changed concurrently; stop')
 fingerprints=lambda r:sorted((a['id'],a['name'],a['size'],a['updated_at']) for a in r['assets'])
 if fingerprints(check)!=fingerprints(created):raise RuntimeError('Draft assets changed concurrently; stop')
 draft_ref=optional_ref(repo,tag)
 if draft_ref and resolve_ref(repo,draft_ref)!=commit:raise RuntimeError('Tag changed concurrently before publication; stop')
 # GitHub has no transaction spanning every check. Publish this same owned ID,
 # never another draft discovered by tag, and verify exact state afterward.
 rest_json('PATCH',f'repos/{repo}/releases/{created_id}',{'tag_name':tag,'target_commitish':commit,'draft':False,'prerelease':True,'make_latest':'false'})
 final=api(f'repos/{repo}/releases/{created_id}');ref=optional_ref(repo,tag)
 if final['id']!=created_id or final.get('tag_name')!=tag or final.get('draft') or not ref or resolve_ref(repo,ref)!=commit:raise RuntimeError('Publication outcome uncertain; inspect before retry')
 if {a['name']:a['id'] for a in final['assets']}!=uploaded_ids:raise RuntimeError('Published asset IDs changed; inspect outcome')
 verify_release(repo,tag,final,expected,commit,True)
 print('Verified requested prerelease:',final['html_url'])
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--source',type=Path,required=True);p.add_argument('--assets',type=Path,required=True);a=p.parse_args()
 try:main(a.source.resolve(),a.assets.resolve())
 except Exception as e:raise SystemExit(str(e)+'\nNo destructive cleanup or retry was performed. A newly created draft may remain after an upload/publication failure; inspect its exact state before any new action.')
