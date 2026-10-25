#!/usr/bin/env python3
"""Portable packaging of committed, CI-built macOS binaries. Never signs/publishes."""
from pathlib import Path
import argparse,hashlib,json,os,plistlib,re,shutil,subprocess,urllib.request,zipfile
from audit_macho import inspect
HERE=Path(__file__).resolve().parent
VERSION_RE=re.compile(r'^0\.2\.12-go-test\.[0-9]{8}\.[1-9][0-9]*$')
COMMIT_RE=re.compile(r'^[0-9a-f]{40}$')
PINS={'arm64':'6f385d3221f8c776ba40530db9baf8c71478907897959849896eeb0ddbf16b41','amd64':'421e98d2683b39555209c59395655b4a217da0382e7669c92f6dff3949832664'}
def digest(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def write(p,s,mode=0o644):p.parent.mkdir(parents=True,exist_ok=True);p.write_text(s);p.chmod(mode)
def copy(a,b,mode=0o644):b.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(a,b);b.chmod(mode)
def git(root,*args):return subprocess.check_output(['git','-C',str(root),*args],text=True).strip()
def runtime_inputs(root):
 result={}
 for p in sorted(root.rglob('*')):
  if not p.is_file() or any(x in p.parts for x in ('.git','__pycache__','node_modules')):continue
  rel=p.relative_to(root).as_posix()
  if rel.endswith('.go') or rel in ('go.mod','go.sum','scripts/go-native.sh','scripts/go-native-overlay.py') or rel.startswith(('third_party/','contrib/','crates/orbit-pi/assets/','crates/orbit-pi/locales/')):
   if p.is_symlink():raise ValueError(f'Source symlink is not allowed: {rel}')
   result[rel]=digest(p)
 raw=''.join(f'{result[k]}  {k}\n' for k in sorted(result)).encode()
 return result,hashlib.sha256(raw).hexdigest()
def marker(root):
 m=json.loads((root/'.github/macos-test-prerelease.json').read_text())
 if m.get('enabled') is not True:raise ValueError('Explicit prerelease marker is disabled')
 if not VERSION_RE.fullmatch(m.get('version','')):raise ValueError('Invalid test-only version')
 if m.get('runtime_reference_tree')!='a7aa6cfde4e6f7ac487f88554deb979cbf0e3182':raise ValueError('Unexpected frozen reference')
 if m.get('minimum_macos')!='13.0' or m.get('native_mac_execution')!='NOT_RUN':raise ValueError('Unvalidated Mac claim')
 return m
class HTTPSOnly(urllib.request.HTTPRedirectHandler):
 def redirect_request(self,req,fp,code,msg,headers,newurl):
  if not newurl.startswith('https://'):raise ValueError('Refusing non-HTTPS native-runtime redirect')
  return super().redirect_request(req,fp,code,msg,headers,newurl)
def native_download(root,arch,out):
 manifest=json.loads((root/'third_party/mygo-terminal/mygo-plugin.json').read_text())
 libs=[x for x in manifest['libraries'] if x['name']=='libghostty-vt'];assert len(libs)==1
 f=libs[0]['files']['darwin-'+arch]
 url=f'https://github.com/egoist/mygo/releases/download/libghostty-vt-befcdfd2c3a1/libghostty-vt-darwin-{arch}.dylib'
 if f!={'name':'libghostty-vt.dylib','sha256':PINS[arch],'url':url}:raise ValueError('Native pin/URL drift')
 dest=out/f'libghostty-vt-{arch}.dylib';out.mkdir(parents=True,exist_ok=True)
 if not dest.exists():
  data=urllib.request.build_opener(HTTPSOnly()).open(url,timeout=90).read(16*1024*1024+1)
  if len(data)>16*1024*1024 or hashlib.sha256(data).hexdigest()!=PINS[arch]:raise ValueError('Runtime download hash/size mismatch')
  dest.write_bytes(data)
 if digest(dest)!=PINS[arch]:raise ValueError('Cached runtime hash mismatch; refusing overwrite')
 return dest,f

def build(root,binaries,out,commit,validation):
 if not COMMIT_RE.fullmatch(commit) or git(root,'rev-parse','HEAD')!=commit:raise ValueError('Commit does not match checkout')
 if git(root,'status','--porcelain','--untracked-files=no'):raise ValueError('Tracked checkout is dirty')
 m=marker(root);inputs,sourcehash=runtime_inputs(root)
 if sourcehash!=m['runtime_source_sha256']:raise ValueError('Default runtime differs from fixed checkpoint')
 v=json.loads(validation.read_text())
 if v.get('source_commit')!=commit or v.get('status')!='PASS' or v.get('native_macOS_execution')!='NOT_RUN':raise ValueError('Missing/mismatched aggregate validation')
 if set(v.get('passed_stages',[]))!={'root-race','vet','toolkit','terminal','wrapper','six-cross-builds','packaging-tests'}:raise ValueError('Incomplete validation gates')
 for rel,h in json.loads((HERE/'licenses-sha256.json').read_text()).items():
  if digest(HERE/'third-party-licenses'/rel)!=h:raise ValueError('License input drift: '+rel)
 out.mkdir(parents=True,exist_ok=True);results=[]
 for arch,label,machine in [('arm64','Apple Silicon','arm64'),('amd64','Intel','x86_64')]:
  binary=binaries/f'orbit-darwin-{arch}';lib,native=native_download(root,arch,out/'.downloads')
  if digest(binary)!=v['binary_sha256'][binary.name]:raise ValueError('Binary differs from validated output')
  appaudit,_=inspect(binary);libaudit,exports=inspect(lib)
  req=re.findall(r'\{"(ghostty_[a-z0-9_]+)",',(root/'third_party/mygo-terminal/internal/vt/lib.go').read_text())
  missing=[s for s in req if '_'+s not in exports]
  if missing or appaudit['architecture']!=arch or libaudit['architecture']!=arch:raise ValueError('Mach-O ABI/symbol/arch mismatch')
  if appaudit['minimum_os']!='13.0.0' or libaudit['minimum_os']!='13.0.0':raise ValueError('Deployment target changed')
  if not all(s.startswith(('/usr/lib/','/System/Library/')) for a in (appaudit,libaudit) for s in a['libraries']):raise ValueError('Unbundled non-system dylib')
  name=f'Orbit-Go-Test-{m["version"]}-{commit[:12]}-ci-macos-{arch}';pkg=out/'.staging'/name
  if pkg.exists():raise ValueError('Refusing to overwrite package staging: '+str(pkg))
  mac=pkg/'Orbit Go Test.app/Contents/MacOS';res=mac.parent/'Resources';prov=pkg/'provenance'
  copy(binary,mac/'orbit-bin',0o755);copy(lib,res/'libghostty-vt.dylib',0o755)
  replacements={'__ARCH__':arch,'__ARCHLABEL__':label,'__MACHINE__':machine,'__APPHASH__':digest(binary),'__LIBHASH__':digest(lib),'__SOURCE_COMMIT__':commit,'__VERSION__':m['version'],'__APP_SUFFIX__':commit[:12]}
  def template(file):
   text=(HERE/file).read_text()
   for a,b in replacements.items():text=text.replace(a,b)
   return text
  write(mac/'Orbit-Go-Test',template('launcher.sh.in'),0o755)
  write(pkg/'READ-ME-FIRST.txt',template('README.txt.in'));write(pkg/'开始测试.txt',template('Chinese.txt.in'))
  copy(HERE/'KNOWN-LIMITS.txt',pkg/'KNOWN-LIMITS.txt')
  copy(root/'assets/icons/alpha-logo.icns',res/'OrbitTest.icns')
  info={'CFBundleDevelopmentRegion':'en','CFBundleName':'Orbit Go Test','CFBundleDisplayName':'Orbit Go Test','CFBundleIdentifier':'dev.orbit.go.experimental.'+commit[:12],'CFBundleVersion':'20261025.1','CFBundleShortVersionString':'0.2.12','CFBundleExecutable':'Orbit-Go-Test','CFBundlePackageType':'APPL','CFBundleIconFile':'OrbitTest','LSMinimumSystemVersion':'13.0','NSHighResolutionCapable':True,'NSPrincipalClass':'NSApplication','NSSupportsAutomaticGraphicsSwitching':True}
  (mac.parent/'Info.plist').write_bytes(plistlib.dumps(info));write(mac.parent/'PkgInfo','APPL????')
  write(res/'sample-workspace/README.md','# Disposable Orbit test workspace\n\nSynthetic test files only. 中文输入测试。\nThis is not your real project. Editor autosave is enabled.\n')
  write(res/'sample-workspace/hello.go','package main\n\nimport "fmt"\n\nfunc main() { fmt.Println("Hello, isolated Orbit test") }\n')
  write(pkg/'Launch.command','#!/bin/bash\nset -euo pipefail\nHERE="$(cd -- "$(dirname "$0")" && pwd -P)"\nexec "$HERE/Orbit Go Test.app/Contents/MacOS/Orbit-Go-Test" "$@"\n',0o755)
  write(pkg/'Verify.command','#!/bin/bash\nset -euo pipefail\nexport PATH=/usr/bin:/bin:/usr/sbin:/sbin\ncd -- "$(dirname "$0")"\n/usr/bin/shasum -a 256 -c SHA256SUMS\n/usr/bin/file "Orbit Go Test.app/Contents/MacOS/orbit-bin" "Orbit Go Test.app/Contents/Resources/libghostty-vt.dylib"\nprintf "Integrity only; native startup and signature acceptance are NOT established.\\n"\n',0o755)
  shutil.copytree(HERE/'third-party-licenses',pkg/'licenses');copy(HERE/'licenses-sha256.json',prov/'licenses-sha256.json')
  copy(validation,prov/'ci-validation-summary.json');copy(root/'go.mod',prov/'go.mod.txt');copy(root/'go.sum',prov/'go.sum.txt')
  copy(root/'third_party/mygo-terminal/mygo-plugin.json',prov/'upstream-native-manifest.json')
  for part in ['terminal','composer','selection']:copy(root/f'third_party/mygo-{part}'/('ORBIT_PATCHES.md' if part=='terminal' else 'README.md'),prov/f'{part}-compatibility.md')
  write(prov/'runtime-inputs.json',json.dumps({'source_commit':commit,'reference_tree':m['runtime_reference_tree'],'runtime_sha256':sourcehash,'files':inputs},indent=2,sort_keys=True)+'\n')
  write(prov/'static-native-audit.json',json.dumps({'orbit':appaudit,'ghostty':libaudit,'required_symbols':len(req),'missing_symbols':missing,'native_macOS_execution':'NOT RUN'},indent=2)+'\n')
  buildinfo=subprocess.check_output(['go','version','-m',str(binary)],text=True)
  write(prov/'go-build-info.txt',buildinfo)
  provenance={'source_commit':commit,'version':m['version'],'runtime_source_sha256':sourcehash,'runtime_reference_tree':m['runtime_reference_tree'],'architecture':arch,'binary_sha256':digest(binary),'ghostty':native,'build':'GitHub Actions Linux, Go 1.27.3, canonical compatibility wrapper, CGO_ENABLED=0','native_macOS_execution':'NOT RUN','signing':'No Developer ID signing or notarization; toolchain/upstream arm64 bytes may contain code-signature commands','reproduction':'Recipe and pins are reproducible; temporary Go module replacement paths can change binary bytes. Actual hashes are authoritative. Old 0930 archive hashes are not reused.','pi':'Not bundled; optional Pi 1.4.0; Node-backed Pi requires Node >=22.19.0','ci_run':v['ci_run']}
  write(prov/'package-provenance.json',json.dumps(provenance,indent=2)+'\n')
  write(pkg/'SHA256SUMS',''.join(f'{digest(p)}  {p.relative_to(pkg).as_posix()}\n' for p in sorted(pkg.rglob('*')) if p.is_file()))
  for p in [mac/'Orbit-Go-Test',pkg/'Launch.command',pkg/'Verify.command']:subprocess.run(['bash','-n',str(p)],check=True)
  target=out/(name+'.zip')
  if target.exists():raise ValueError('Refusing archive overwrite: '+str(target))
  with zipfile.ZipFile(target,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
   for p in sorted(pkg.rglob('*')):
    if p.is_symlink():raise ValueError('Symlink in bundle')
    if p.is_file():
     zi=zipfile.ZipInfo(name+'/'+p.relative_to(pkg).as_posix(),(2026,10,25,9,30,0));zi.create_system=3;zi.external_attr=(0o100000|(p.stat().st_mode&0o777))<<16;zi.compress_type=zipfile.ZIP_DEFLATED;z.writestr(zi,p.read_bytes(),compresslevel=9)
  with zipfile.ZipFile(target) as z:
   if z.testzip() is not None:raise ValueError('ZIP corruption')
   for line in z.read(name+'/SHA256SUMS').decode().splitlines():
    h,rel=line.split('  ',1)
    if hashlib.sha256(z.read(name+'/'+rel)).hexdigest()!=h:raise ValueError('Archived checksum mismatch')
  results.append({'filename':target.name,'sha256':digest(target),'bytes':target.stat().st_size,'architecture':arch,'binary_sha256':digest(binary),'ghostty_sha256':digest(lib)})
 write(out/'release-manifest.json',json.dumps({'version':m['version'],'tag':'macos-test-v'+m['version'],'source_commit':commit,'runtime_source_sha256':sourcehash,'ci_run':v['ci_run'],'assets':results},indent=2)+'\n')
 copy(HERE/'KNOWN-LIMITS.txt',out/'KNOWN-LIMITS.txt')
 assetpaths=[out/x['filename'] for x in results]+[out/'KNOWN-LIMITS.txt',out/'release-manifest.json']
 write(out/'SHA256SUMS',''.join(f'{digest(p)}  {p.name}\n' for p in assetpaths))
 write(out/'RELEASE-NOTES.md',f'Experimental Go + MyGo macOS prerelease, not a complete replacement.\n\nSource commit: {commit}\nCI validation: {v["ci_run"]}\n\nDownload the arm64 ZIP for Apple Silicon or amd64 ZIP for Intel. macOS 13.0 minimum is declared in the binaries. Start with Launch.command for a fresh isolated HOME and synthetic workspace. Pi is separate; a missing-Pi state is expected by default.\n\nNo native macOS startup/acceptance has run. No Developer ID signing/notarization. If macOS blocks opening, stop and report the message; do not change security settings.\n\n'+(HERE/'KNOWN-LIMITS.txt').read_text())
 print(json.dumps({'version':m['version'],'source_commit':commit,'assets':results},indent=2))
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--source',type=Path,required=True);p.add_argument('--binaries',type=Path,required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--commit',required=True);p.add_argument('--validation',type=Path,required=True);a=p.parse_args();build(a.source.resolve(),a.binaries.resolve(),a.output.resolve(),a.commit,a.validation.resolve())
