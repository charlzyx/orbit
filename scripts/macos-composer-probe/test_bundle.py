#!/usr/bin/env python3
"""Derive a clearly instrumented test app; never change baseline or release gates."""
import argparse,hashlib,json,os,plistlib,shutil,stat,sys
from pathlib import Path
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(HERE.parent/'macos-app'))
from bundle import tree_hashes

def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def derive(baseline,expected_path,expected_sha,binary,binary_sha,overlay_manifest,output):
 for p in (baseline,expected_path,binary,overlay_manifest):
  if not p.is_absolute() or p.is_symlink():raise ValueError('Only explicit absolute non-symlink inputs')
 if not baseline.is_dir() or not baseline.name.endswith('.app') or not binary.is_file():raise ValueError('App/binary inputs required')
 if not output.is_absolute() or output.exists() or output.is_relative_to(baseline):raise ValueError('Fresh independent output required')
 if digest(expected_path)!=expected_sha or digest(binary)!=binary_sha:raise ValueError('Independent expected input hash mismatch')
 expected=json.loads(expected_path.read_text());actual=tree_hashes(baseline)
 if actual!=expected:raise ValueError('Baseline differs from exact official expected-app manifest')
 info=plistlib.loads((baseline/'Contents/Info.plist').read_bytes());name=info.get('CFBundleExecutable')
 if not isinstance(name,str) or name in ('','.','..') or '/' in name or '\\' in name:raise ValueError('Unsafe executable name')
 rel='Contents/MacOS/'+name
 if rel not in expected or 'sha256' not in expected[rel]:raise ValueError('Baseline executable missing')
 overlay=json.loads(overlay_manifest.read_text())
 if overlay.get('source_tree')!='0476116687acf0838d52a0246210270000530530' or overlay.get('source_remote_commit')!='ee25813f9a2065ae8632d4c15b24e3c99e7bf469':raise ValueError('Unreviewed instrumentation base')
 source=json.loads((baseline/'Contents/Resources/Provenance/bundle-provenance.json').read_text())
 output.mkdir(mode=0o700);app=output/'Orbit Composer Proof.app';shutil.copytree(baseline,app,symlinks=True)
 target=app/rel;shutil.copyfile(binary,target);os.chmod(target,expected[rel]['mode'])
 derived=tree_hashes(app);changes=sorted(k for k in set(expected)|set(derived) if expected.get(k)!=derived.get(k))
 if changes!=[rel] or derived[rel]['sha256']!=binary_sha:raise ValueError('Unexpected test bundle changes')
 if tree_hashes(baseline)!=expected:raise ValueError('Baseline mutated')
 manifest={'kind':'instrumented-test-app-not-release','baseline_source':source.get('source_commit'),'baseline_expected_manifest_sha256':expected_sha,'instrumented_source':overlay['source_remote_commit'],'instrumented_tree':overlay['source_tree'],'overlay_manifest_sha256':digest(overlay_manifest),'binary_sha256':binary_sha,'changed_paths':changes,'original_binary_sha256':expected[rel]['sha256'],'derived_files':derived,'limitations':['Baseline resources may predate Provider; their source and exact hashes are retained, not relabeled Provider validation','Info.plist BundleID/Home routing unchanged','No signature/quarantine/security changes; loader rejection must block native proof']}
 (output/'test-bundle-manifest.json').write_text(json.dumps(manifest,indent=2,sort_keys=True)+'\n');return app
if __name__=='__main__':
 p=argparse.ArgumentParser();
 for n in ('baseline','expected','binary','overlay-manifest','output'):p.add_argument('--'+n,type=Path,required=True)
 p.add_argument('--expected-sha256',required=True);p.add_argument('--binary-sha256',required=True);a=p.parse_args();print(derive(a.baseline,a.expected,a.expected_sha256,a.binary,a.binary_sha256,a.overlay_manifest,a.output))
