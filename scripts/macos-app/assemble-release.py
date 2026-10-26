#!/usr/bin/env python3
"""Assemble only the arm64 native-smoke-passed DMG from this exact Actions run."""
from pathlib import Path
import argparse,json,shutil,os
from common import check_source,write,digest,HERE,ARCHITECTURES

def assemble(root,inputs,out,commit):
 m,_=check_source(root,commit)
 if {p.name for p in inputs.iterdir()}!=set(ARCHITECTURES):raise ValueError('Wrong architecture input set; arm64 only')
 if out.exists():raise ValueError('Output already exists; refusing overwrite')
 out.mkdir(parents=True);assets=[]
 for arch in ARCHITECTURES:
  p=inputs/arch/f'package-{arch}.json';a=json.loads(p.read_text())
  name=f'Orbit-Go-Test-{m["version"]}-{commit[:12]}-ci-macos-{arch}.dmg'
  if a.get('filename')!=name or a.get('source_commit')!=commit or a.get('version')!=m['version'] or a.get('architecture')!=arch or a.get('native_macos_smoke')!='PASS':raise ValueError('Wrong source/version or absent native smoke gate')
  expected_ci=f'https://github.com/{os.environ["GITHUB_REPOSITORY"]}/actions/runs/{os.environ["GITHUB_RUN_ID"]}'
  if a.get('ci_run')!=expected_ci:raise ValueError('Artifact not produced by this Actions run')
  source=p.parent/name
  if source.is_symlink() or source.stat().st_size!=a['bytes'] or not 0<a['bytes']<=512*1024*1024 or digest(source)!=a['sha256']:raise ValueError('DMG asset size/hash mismatch')
  shutil.copyfile(source,out/name);assets.append(a)
 manifest={'version':m['version'],'tag':'macos-test-v'+m['version'],'source_commit':commit,'runtime_source_sha256':m['runtime_source_sha256'],'native_macos_smoke':{arch:'PASS' for arch in ARCHITECTURES},'assets':assets}
 write(out/'release-manifest.json',json.dumps(manifest,indent=2)+'\n');shutil.copyfile(HERE/'KNOWN-LIMITS.txt',out/'KNOWN-LIMITS.txt')
 paths=[out/a['filename'] for a in assets]+[out/'release-manifest.json',out/'KNOWN-LIMITS.txt']
 write(out/'SHA256SUMS',''.join(f'{digest(p)}  {p.name}\n' for p in paths))
 write(out/'RELEASE-NOTES.md',f'''Experimental native Go + MyGo Finder app, source {commit}.

This test package contains the user-approved vector GOPI sidebar wordmark and pixel PI Go app icon. Bundle identity and private Orbit Home paths are unchanged. It includes bounded selected-process recovery without replaying prompts/tools, an independent commit-message model preference, request/error ownership fixes, and the tested editor rename/identity protections. Composer navigation, menus, folding, attachment admission, normal Pi configuration inheritance and bounded private diagnostics remain. The user reported .20261026.2 usable, but full functionality, feel and visual parity remain unfinished; the earlier specific workspace-picker hang is not independently confirmed fixed. All earlier tags and assets remain unchanged.

Download the arm64 DMG for Apple Silicon. Future builds are arm64 only; existing Intel releases and assets remain unchanged. Open the disk image, drag Orbit Go Test.app to Applications, eject the image, then open the app. No Terminal command or separate Node/Pi installation is required. Bundled Node22.23.4 and Pi1.4.0 materialize into private app runtime data on first launch. Pi reads the launch user's normal ~/.pi/agent or PI_CODING_AGENT_DIR, including existing model/provider configuration. Stock Pi may normally migrate settings and refresh OAuth credentials; Orbit does not copy credentials. If no provider is configured, configure one in the app before model inference. No account or credential is bundled.

This exact arm64 DMG passed matching-architecture hosted macOS launch smoke: read-only mount, copy outside build directory, detach, native visible first frame with an in-process content-render PNG (not OS screen pixels), Ghostty load, Pi RPC get_state, second launch and unchanged app bytes. These are bounded smoke checks, not full feature or pixel parity, not downloaded-app Gatekeeper acceptance, and not successful model inference.

This is a new version, not a retry or replacement of .3 or .4. Earlier validation/publication failures are preserved. The dedicated cloud Mac composer proof had no assigned runner in either attempt and remains NOT_RUN; ordinary hosted launch smoke does not establish clipboard, IME or full interaction parity.

This is not Developer ID signed or notarized. macOS may block opening a downloaded build. If blocked, stop and report the message; no OS security changes or bypass commands are provided. The first experimental release/tag/assets remain unchanged.

'''+(HERE/'KNOWN-LIMITS.txt').read_text())
 print(json.dumps(manifest,indent=2))
if __name__=='__main__':
 p=argparse.ArgumentParser()
 for flag in ('source','inputs','output'):p.add_argument('--'+flag,type=Path,required=True)
 p.add_argument('--commit',required=True);a=p.parse_args();assemble(a.source.resolve(),a.inputs.resolve(),a.output.resolve(),a.commit)
