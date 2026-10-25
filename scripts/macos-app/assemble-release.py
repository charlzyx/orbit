#!/usr/bin/env python3
"""Assemble only two native-smoke-passed DMGs from this exact Actions run."""
from pathlib import Path
import argparse,json,shutil,os
from common import check_source,write,digest,HERE

def assemble(root,inputs,out,commit):
 m,_=check_source(root,commit)
 if out.exists():raise ValueError('Output already exists; refusing overwrite')
 out.mkdir(parents=True);assets=[]
 for arch in ('arm64','amd64'):
  p=inputs/arch/f'package-{arch}.json';a=json.loads(p.read_text())
  name=f'Orbit-Go-Test-{m["version"]}-{commit[:12]}-ci-macos-{arch}.dmg'
  if a.get('filename')!=name or a.get('source_commit')!=commit or a.get('version')!=m['version'] or a.get('architecture')!=arch or a.get('native_macos_smoke')!='PASS':raise ValueError('Wrong source/version or absent native smoke gate')
  expected_ci=f'https://github.com/{os.environ["GITHUB_REPOSITORY"]}/actions/runs/{os.environ["GITHUB_RUN_ID"]}'
  if a.get('ci_run')!=expected_ci:raise ValueError('Artifact not produced by this Actions run')
  source=p.parent/name
  if source.is_symlink() or source.stat().st_size!=a['bytes'] or not 0<a['bytes']<=512*1024*1024 or digest(source)!=a['sha256']:raise ValueError('DMG asset size/hash mismatch')
  shutil.copyfile(source,out/name);assets.append(a)
 manifest={'version':m['version'],'tag':'macos-test-v'+m['version'],'source_commit':commit,'runtime_source_sha256':m['runtime_source_sha256'],'native_macos_smoke':{'arm64':'PASS','amd64':'PASS'},'assets':assets}
 write(out/'release-manifest.json',json.dumps(manifest,indent=2)+'\n');shutil.copyfile(HERE/'KNOWN-LIMITS.txt',out/'KNOWN-LIMITS.txt')
 paths=[out/a['filename'] for a in assets]+[out/'release-manifest.json',out/'KNOWN-LIMITS.txt']
 write(out/'SHA256SUMS',''.join(f'{digest(p)}  {p.name}\n' for p in paths))
 write(out/'RELEASE-NOTES.md',f'''Experimental native Go + MyGo Finder app, source {commit}.

Download arm64 DMG for Apple Silicon or amd64 DMG for Intel. Open the disk image, drag Orbit Go Test.app to Applications, eject the image, then open the app. No Terminal command or separate Node/Pi installation is required. Bundled Node22.23.4 and Pi1.4.0 materialize into isolated test app data on first launch. Configure a provider in the app before model inference; no account or credential is bundled.

Both exact DMGs passed matching-architecture hosted macOS launch smoke: read-only mount, copy outside build directory, detach, native visible first frame with captured PNG, Ghostty load, Pi RPC get_state, second launch and unchanged app bytes. These are bounded smoke checks, not full feature or pixel parity, not downloaded-app Gatekeeper acceptance, and not successful model inference.

This is not Developer ID signed or notarized. macOS may block opening a downloaded build. If blocked, stop and report the message; no OS security changes or bypass commands are provided. The first experimental release/tag/assets remain unchanged.

'''+(HERE/'KNOWN-LIMITS.txt').read_text())
 print(json.dumps(manifest,indent=2))
if __name__=='__main__':
 p=argparse.ArgumentParser()
 for flag in ('source','inputs','output'):p.add_argument('--'+flag,type=Path,required=True)
 p.add_argument('--commit',required=True);a=p.parse_args();assemble(a.source.resolve(),a.inputs.resolve(),a.output.resolve(),a.commit)
