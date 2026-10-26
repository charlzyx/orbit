#!/usr/bin/env python3
"""Build a native Finder .app and DMG on a matching hosted Mac; never publish."""
from pathlib import Path
import argparse,json,os,platform,plistlib,re,shutil,subprocess,struct
from common import HERE,LEGACY,check_source,copy,write,run,digest,native_download,inspect,ARCHITECTURES
APP_NAME='Orbit Go Test.app'
SYSTEM_PATH='/usr/bin:/bin:/usr/sbin:/sbin'

def tree_hashes(root):
 out={}
 for p in sorted(root.rglob('*')):
  rel=p.relative_to(root).as_posix()
  if p.is_symlink():
   if os.path.isabs(os.readlink(p)) or not p.resolve().is_relative_to(root.resolve()):raise ValueError('External or absolute bundle symlink: '+rel)
   out[rel]={'symlink':os.readlink(p)}
  elif p.is_file():out[rel]={'sha256':digest(p),'mode':p.stat().st_mode&0o777}
 return out

def require_arch(path,arch):
 a,_=inspect(path)
 if a['architecture']!=arch:raise ValueError('Wrong Mach-O architecture: '+str(path))
 if not a['minimum_os'] or tuple(map(int,a['minimum_os'].split('.')))>(13,0,0):raise ValueError('Dependency exceeds declared macOS13 minimum: '+str(path))
 data=Path(path).read_bytes();commands=struct.unpack_from('<I',data,16)[0];off=32;macos=False
 for _ in range(commands):
  cmd,size=struct.unpack_from('<II',data,off)
  if cmd==0x24:macos=True
  if cmd==0x32:
   if struct.unpack_from('<I',data,off+8)[0]!=1:raise ValueError('Non-macOS Mach-O platform')
   macos=True
  off+=size
 if not macos:raise ValueError('No macOS deployment metadata: '+str(path))
 a['platform']='macOS' 
 if not all(x.startswith(('/usr/lib/','/System/Library/')) for x in a['libraries']):raise ValueError('Unexpected non-system dylib chain: '+str(path))
 return a

def verify_runtime(runtime,arch):
 m=json.loads((runtime/'runtime-manifest.json').read_text());pins=json.loads((HERE/'runtime/node-pins.json').read_text());selected=pins['archives'][arch]
 if m.get('execution_evidence')!='matching macOS Node --version and npm ci completed; application smoke still required':raise ValueError('Audit-only runtime is not a native build input')
 if m.get('node',{}).get('selected_archive')!=selected or m['node'].get('archive_sha256_verified')!=selected['sha256'] or m['node'].get('version')!='22.23.4':raise ValueError('Bundled Node provenance mismatch')
 if m.get('pi',{}).get('version')!='1.4.0' or m['pi'].get('lock_sha256')!=digest(HERE/'runtime/package-lock.json') or digest(runtime/'pi/package-lock.json')!=m['pi']['lock_sha256']:raise ValueError('Bundled Pi lock provenance mismatch')
 if digest(runtime/'original-pi-files.json')!=m['pi']['original_file_inventory_sha256']:raise ValueError('Pi inventory drift')
 expected={x['path']:x for x in json.loads((runtime/'original-pi-files.json').read_text())}
 for row in m['pruned_files']:
  if not row['path'].startswith('node_modules/@earendil-works/pi-coding-agent/examples/') or expected.get(row['path'])!=row:raise ValueError('Unexpected runtime pruning')
  del expected[row['path']]
 actual=tree_hashes(runtime/'pi')
 if set(expected)!=set(actual):raise ValueError('Pi runtime file set differs from inventory')
 for path,row in expected.items():
  key='symlink' if 'symlink' in row else 'sha256'
  if actual[path].get(key)!=row[key]:raise ValueError('Pi runtime inventory mismatch: '+path)
 return m

def verify_go_icon(root):
 manifest=json.loads((root/'assets/icons/go/manifest.json').read_text())
 source='assets/icons/go/orbit-go.svg'
 outputs=('assets/icons/go/orbit-go.icns','crates/orbit-pi/assets/go-app-icon.png')
 if manifest.get('source')!=source or manifest.get('source_sha256')!=digest(root/source):raise ValueError('Go icon vector source mismatch')
 if set(manifest.get('outputs',{}))!=set(outputs):raise ValueError('Unexpected Go icon output paths')
 for rel in outputs:
  p=root/rel;row=manifest['outputs'][rel]
  if row.get('sha256')!=digest(p) or row.get('bytes')!=p.stat().st_size:raise ValueError('Go icon generated asset mismatch')
 data=(root/outputs[0]).read_bytes()
 if len(data)<8 or data[:4]!=b'icns' or struct.unpack('>I',data[4:8])[0]!=len(data):raise ValueError('Invalid Go ICNS container')
 expected={'icp4':16,'icp5':32,'icp6':64,'ic07':128,'ic08':256,'ic09':512,'ic10':1024,'ic11':32,'ic12':64,'ic13':256,'ic14':512};seen=set();off=8
 while off<len(data):
  if off+8>len(data):raise ValueError('Truncated Go ICNS chunk')
  kind=data[off:off+4].decode('ascii');size=struct.unpack('>I',data[off+4:off+8])[0]
  if kind not in expected or kind in seen or size<32 or off+size>len(data):raise ValueError('Invalid Go ICNS slot')
  png=data[off+8:off+size]
  if png[:8]!=b'\x89PNG\r\n\x1a\n' or png[12:16]!=b'IHDR' or struct.unpack('>II',png[16:24])!=(expected[kind],expected[kind]):raise ValueError('Invalid Go ICNS PNG dimensions')
  seen.add(kind);off+=size
 if seen!=set(expected) or manifest.get('icns_slots')!=expected:raise ValueError('Missing Go ICNS size/retina slots')
 return manifest

def install_go_icon(root,res):
 verify_go_icon(root)
 copy(root/'assets/icons/go/orbit-go.icns',res/'OrbitTest.icns')
 if digest(res/'OrbitTest.icns')!=digest(root/'assets/icons/go/orbit-go.icns'):raise ValueError('Packaged Go icon changed')
 copy(root/'assets/icons/go/manifest.json',res/'Provenance/go-icon-manifest.json')

def build(root,binaries,runtime,out,commit,arch,validation):
 if arch not in ARCHITECTURES:raise ValueError('Future macOS app builds support arm64 only')
 if platform.system()!='Darwin' or platform.machine()!='arm64':raise ValueError('Matching native macOS runner required')
 m,inputs=check_source(root,commit)
 runtime_evidence=verify_runtime(runtime,arch)
 v=json.loads(validation.read_text())
 if v.get('source_commit')!=commit or v.get('status')!='PASS' or set(v.get('passed_stages',[]))!={'root-race','vet','toolkit','terminal','wrapper','five-cross-builds','packaging-tests'}:raise ValueError('Exact aggregate validation required')
 if out.exists():raise ValueError('Fresh output required; no archive overwrite')
 out.mkdir(parents=True);volume=out/'.volume';app=volume/APP_NAME;mac=app/'Contents/MacOS';res=app/'Contents/Resources';frameworks=app/'Contents/Frameworks';prov=res/'Provenance'
 binary=binaries/f'orbit-darwin-{arch}'
 if digest(binary)!=v['binary_sha256'][binary.name]:raise ValueError('Binary checksum differs from tested build')
 lib,libpin=native_download(root,arch,out/'.downloads')
 appaudit=require_arch(binary,arch);libaudit=require_arch(lib,arch)
 req=re.findall(r'\{"(ghostty_[a-z0-9_]+)",',(root/'third_party/mygo-terminal/internal/vt/lib.go').read_text());_,exports=inspect(lib)
 if any('_'+s not in exports for s in req):raise ValueError('Missing Ghostty ABI symbols')
 copy(binary,mac/'Orbit',0o755);copy(lib,frameworks/'libghostty-vt.dylib',0o755)
 for part in ('node','pi'):shutil.copytree(runtime/part,res/part,symlinks=True)
 nodeaudit=require_arch(res/'node/bin/node',arch)
 if not (res/'pi/node_modules/@earendil-works/pi-coding-agent/dist/bundle/cli.js').is_file():raise ValueError('Pinned Pi CLI missing')
 install_go_icon(root,res)
 info={'CFBundleDevelopmentRegion':'en','CFBundleName':'Orbit Go Test','CFBundleDisplayName':'Orbit Go Test','CFBundleIdentifier':'dev.orbit.go.experimental','CFBundleVersion':m['version'].split('go-test.')[1],'CFBundleShortVersionString':'0.2.12','CFBundleExecutable':'Orbit','CFBundlePackageType':'APPL','CFBundleIconFile':'OrbitTest','LSMinimumSystemVersion':'13.0','NSHighResolutionCapable':True,'NSPrincipalClass':'NSApplication','NSSupportsAutomaticGraphicsSwitching':True}
 (app/'Contents/Info.plist').write_bytes(plistlib.dumps(info));write(app/'Contents/PkgInfo','APPL????')
 write(res/'bin/pi','#!/bin/sh\nset -eu\nRES="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)"\nexec "$RES/node/bin/node" "$RES/pi/node_modules/@earendil-works/pi-coding-agent/dist/bundle/cli.js" "$@"\n',0o755)
 write(res/'sample-workspace/README.md','# Orbit Go Test workspace\n\nSynthetic starter project. 中文输入测试。\nChanges are saved in your experimental app data directory, never in the application bundle.\n')
 write(res/'sample-workspace/hello.go','package main\n\nimport "fmt"\nfunc main() { fmt.Println("Hello, Orbit") }\n')
 for rel,h in json.loads((LEGACY/'licenses-sha256.json').read_text()).items():
  if digest(LEGACY/'third-party-licenses'/rel)!=h:raise ValueError('Core license drift: '+rel)
 shutil.copytree(LEGACY/'third-party-licenses',res/'Licenses/core')
 shutil.copytree(runtime/'licenses',res/'Licenses/node-pi')
 copy(runtime/'runtime-manifest.json',prov/'runtime-manifest.json')
 copy(runtime/'original-pi-files.json',prov/'original-pi-files.json')
 copy(HERE/'runtime/pi-upstream-package-lock.json',prov/'pi-upstream-package-lock.json')
 copy(res/'node/LICENSE',res/'Licenses/node-pi/Node-LICENSE')
 # Pristine bundled Pi is materialized atomically into isolated writable app data
 # by the native bootstrap. Existing runtime patch preferences apply to that copy.
 pi_hashes=tree_hashes(res/'pi')
 import hashlib
 pi_tree_sha256=hashlib.sha256(json.dumps(pi_hashes,sort_keys=True,separators=(',',':')).encode()).hexdigest()
 write(res/'orbit-bundle.json',json.dumps({'schema':1,'version':m['version'],'source_commit':commit,'pi_version':'1.4.0','pi_rpc_patches':[],'pi_tree_sha256':pi_tree_sha256},indent=2)+'\n')
 env={'HOME':str(out/'.version-home'),'PATH':str(res/'node/bin')+':'+SYSTEM_PATH,'PI_OFFLINE':'1','PI_SKIP_VERSION_CHECK':'1'}
 Path(env['HOME']).mkdir()
 # Verify official bundled Node is runnable on this runner, Pi resolves without user's PATH.
 nodeversion=subprocess.check_output([str(res/'node/bin/node'),'--version'],env=env,text=True).strip()
 piversion=subprocess.check_output([str(res/'bin/pi'),'--version'],env=env,text=True,timeout=30).strip()
 if nodeversion!='v22.23.4' or piversion!='1.4.0':raise ValueError('Bundled Node/Pi version mismatch')
 selected_native=res/'pi/node_modules/@earendil-works/pi-tui/native/darwin/prebuilds'/'darwin-arm64'/'darwin-platform.node'
 addon_audit=require_arch(selected_native,arch)
 run(res/'node/bin/node','-e','require(process.argv[1]); console.log("Pi native addon loaded")',selected_native,env=env,timeout=30)
 write(prov/'pi-native-addon.json',json.dumps({'path':str(selected_native.relative_to(res)),'sha256':digest(selected_native),'audit':addon_audit,'native_load':'PASS'},indent=2)+'\n')
 copy(validation,prov/'aggregate-validation.json');copy(root/'go.mod',prov/'go.mod.txt');copy(root/'go.sum',prov/'go.sum.txt')
 write(prov/'go-build-info.txt',subprocess.check_output(['go','version','-m',str(binary)],text=True))
 write(prov/'runtime-inputs.json',json.dumps({'source_commit':commit,'runtime_sha256':m['runtime_source_sha256'],'files':inputs},indent=2,sort_keys=True)+'\n')
 write(prov/'bundle-provenance.json',json.dumps({'source_commit':commit,'version':m['version'],'architecture':arch,'binary_sha256':digest(binary),'ghostty':libpin,'node_version':nodeversion,'pi_version':piversion,'static_native_audit':{'app':appaudit,'ghostty':libaudit,'node':nodeaudit},'build':'Exact committed source, canonical compatibility wrapper, CGO_ENABLED=0; native macOS smoke gates release separately','signing':'No Developer ID signature or notarization; upstream/toolchain signatures preserved. No Gatekeeper bypass.','ci_run':v['ci_run']},indent=2)+'\n')
 copy(HERE/'KNOWN-LIMITS.txt',res/'KNOWN-LIMITS.txt');copy(HERE/'README.txt',res/'README.txt')
 hashes=tree_hashes(app);write(prov/'bundle-files.json',json.dumps(hashes,indent=2,sort_keys=True)+'\n')
 copy(HERE/'README.txt',volume/'READ-ME-FIRST.txt');copy(HERE/'KNOWN-LIMITS.txt',volume/'KNOWN-LIMITS.txt');os.symlink('/Applications',volume/'Applications')
 name=f'Orbit-Go-Test-{m["version"]}-{commit[:12]}-ci-macos-{arch}.dmg';target=out/name
 run('/usr/bin/hdiutil','create','-volname','Orbit Go Test','-srcfolder',volume,'-format','UDZO','-fs','HFS+',target)
 run('/usr/bin/hdiutil','verify',target)
 # Preserve expected app bytes outside the DMG; no payload edits after creation.
 write(out/f'expected-app-{arch}.json',json.dumps(tree_hashes(app),sort_keys=True)+'\n')
 shutil.rmtree(volume)
 write(out/f'package-{arch}.json',json.dumps({'filename':name,'sha256':digest(target),'bytes':target.stat().st_size,'architecture':arch,'source_commit':commit,'version':m['version'],'binary_sha256':digest(binary),'ghostty_sha256':digest(lib),'ci_run':v['ci_run']},indent=2)+'\n')
 print(target)
if __name__=='__main__':
 p=argparse.ArgumentParser()
 for flag in ('source','binaries','runtime','output','validation'):p.add_argument('--'+flag,type=Path,required=True)
 p.add_argument('--commit',required=True);p.add_argument('--arch',choices=ARCHITECTURES,required=True);a=p.parse_args()
 build(a.source.resolve(),a.binaries.resolve(),a.runtime.resolve(),a.output.resolve(),a.commit,a.arch,a.validation.resolve())
