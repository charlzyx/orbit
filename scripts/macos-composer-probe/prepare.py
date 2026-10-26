#!/usr/bin/env python3
"""Create an explicit instrumentation overlay; never mutate production sources."""
import argparse,hashlib,json,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
BASE_TREE='0476116687acf0838d52a0246210270000530530'
BASE_REMOTE='ee25813f9a2065ae8632d4c15b24e3c99e7bf469'
def sha(b):return hashlib.sha256(b).hexdigest()
def replace_once(s,old,new):
 if s.count(old)!=1:raise ValueError('Changed/ambiguous native instrumentation anchor')
 return s.replace(old,new,1)
def prepare(out):
 tree=subprocess.check_output(['git','rev-parse','HEAD^{tree}'],cwd=ROOT,text=True).strip()
 if tree!=BASE_TREE:raise ValueError('Not the verified immutable Provider base tree')
 if subprocess.check_output(['git','diff','HEAD','--','cmd','internal','go.mod','go.sum','third_party','crates/orbit-pi/assets'],cwd=ROOT):raise ValueError('Production inputs changed')
 if subprocess.check_output(['git','ls-files','--others','--exclude-standard','--','cmd','internal','third_party','crates/orbit-pi/assets'],cwd=ROOT):raise ValueError('Untracked production inputs')
 if not out.is_absolute() or out.exists():raise ValueError('Overlay output must be fresh absolute directory')
 run=ROOT/'internal/ui/run.go';source=run.read_text();changed=replace_once(source,'\tx.UI.View(c)\n','\tx.UI.View(c)\n\tcomposerProofObserve(x,c)\n');changed=replace_once(changed,'\t\tw.OnReadyToShow(func() {','\t\tw.OnReadyToShow(func() {\n\t\t\tcomposerProofStart(x,w)')
 virtual=ROOT/'internal/ui/composer_probe_darwin.go'
 if virtual.exists():raise ValueError('Would shadow real source')
 out.mkdir();(out/'run.go').write_text(changed);helper=HERE/'composer_probe_darwin.go.txt';(out/'composer_probe_darwin.go').write_bytes(helper.read_bytes());(out/'overlay.json').write_text(json.dumps({'Replace':{str(run):str(out/'run.go'),str(virtual):str(out/'composer_probe_darwin.go')}},indent=2)+'\n')
 manifest={'source_remote_commit':BASE_REMOTE,'source_tree':tree,'local_head':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'original_run_sha256':sha(run.read_bytes()),'instrumented_run_sha256':sha((out/'run.go').read_bytes()),'helper_sha256':sha(helper.read_bytes()),'probe_sources':{str(p.relative_to(ROOT)):sha(p.read_bytes()) for p in sorted((HERE/'menuproof').glob('*.go'))},'native_execution':'NOT_RUN','limitations':['Instrumented source, not byte-identical release binary','Observer only reads business state; exact finite event driver adds native inputs only when explicitly opted in','No workflow is changed or invoked by preparation']}
 (out/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n');return out/'overlay.json'
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--output',type=Path,required=True);a=p.parse_args();print(prepare(a.output))
