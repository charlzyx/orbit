#!/usr/bin/env python3
"""Verify and copy the metrics-only candidate. Canonical use is explicit diagnostic only."""
from __future__ import annotations
import argparse, hashlib, json, shutil
from pathlib import Path
ROOT=Path(__file__).resolve().parent
PIN='9c69cf829ece196c5b3a3601f1c7bbe70526376f'
IDENTITY={'upstream_commit':PIN,'module':'github.com/egoist/mygo','version':'v0.6.4-0.20261025001514-9c69cf829ece','module_sum':'h1:zKaW0m3ftywWzdR18dv/b+MKgS+iKTu7HF5twRZtSvs=','go_mod_sum':'h1:dVav53PwF2xEotjf4oA5lTY2Jp2exgzyjLhbSXhbMwE=','pristine_tree_sha256':'96dae6d68d612b648267f72fdf90f2211698612d29f3c89ca923a314c23af110'}
BASELINE={'internal/text/pango_linux.go'}
FILES={
 'internal/text/pango_linux.go':'pango_linux.go.txt',
 'internal/text/orbit_text_metrics_linux.go':'orbit_text_metrics_linux.go.txt',
 'internal/text/orbit_text_metrics_linux_test.go':'orbit_text_metrics_linux_test.go.txt',
 'internal/text/orbit_text_metrics_specimen_linux_test.go':'orbit_text_metrics_specimen_linux_test.go.txt',
 'ui/orbit_text_metrics_linux.go':'orbit_text_metrics_ui_linux.go.txt',
 'ui/orbit_text_metrics_other.go':'orbit_text_metrics_ui_other.go.txt',
}
def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()
def tree(root):
 rows={str(p.relative_to(root)):sha(p) for p in root.rglob('*') if p.is_file() and '.git' not in p.relative_to(root).parts and p.name!='.cargo-ok'}
 return hashlib.sha256(json.dumps(rows,sort_keys=True,separators=(',',':')).encode()).hexdigest()
def verify(source,assets=ROOT):
 m=json.loads((assets/'baseline.json').read_text())
 expected_keys={'schema','upstream_commit','module','version','module_sum','go_mod_sum','pristine_tree_sha256','baseline_sha256','overlays','patch_sha256','license_sha256'}
 if set(m)!=expected_keys or m['schema']!=1 or any(m[k]!=v for k,v in IDENTITY.items()):raise ValueError('unexpected candidate identity/schema')
 if set(m['baseline_sha256'])!=BASELINE or set(m['overlays'])!=set(FILES):raise ValueError('unexpected candidate paths')
 if sha(assets/'text-metrics.patch')!=m['patch_sha256'] or sha(assets/'LICENSE')!=m['license_sha256']:raise ValueError('patch/license hash mismatch')
 out={}
 for target,filename in FILES.items():
  item=m['overlays'][target]
  if set(item)!={'source','sha256'} or item['source']!=filename or (assets/filename).resolve().parent!=assets.resolve() or not (assets/filename).is_file() or sha(assets/filename)!=item['sha256']:raise ValueError('candidate overlay identity/hash mismatch: '+target)
  if target in BASELINE:
   if sha(source/target)!=m['baseline_sha256'][target]:raise ValueError('upstream source drift: '+target)
  elif (source/target).exists():raise ValueError('additive candidate path already exists: '+target)
  out[target]=assets/filename
 return m,out

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--source',type=Path,required=True);p.add_argument('--output',type=Path,required=True);a=p.parse_args();src=a.source.resolve();out=a.output.resolve();repo=ROOT.parents[2]
 if out.exists() or out.is_relative_to(src) or src.is_relative_to(out) or out.is_relative_to(repo) or repo.is_relative_to(out):p.error('new disjoint output outside repository required')
 m,overlays=verify(src)
 if tree(src)!=m['pristine_tree_sha256']:raise ValueError('pristine full source tree hash mismatch')
 shutil.copytree(src,out,ignore=shutil.ignore_patterns('.git'))
 for target,source in overlays.items():shutil.copyfile(source,out/target)
 (out.parent/(out.name+'-provenance.json')).write_text(json.dumps({'schema':1,'candidate':True,'production_enabled':False,'upstream_commit':PIN,'source_tree_sha256':m['pristine_tree_sha256'],'candidate_tree_sha256':tree(out),'manifest_sha256':sha(ROOT/'baseline.json'),'overlays':m['overlays']},indent=2)+'\n')
 print(out)
if __name__=='__main__':main()
