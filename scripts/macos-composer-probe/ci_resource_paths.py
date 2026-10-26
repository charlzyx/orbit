"""Resolve only the two exact reviewed resources after safe ZIP extraction."""
import hashlib,json,shutil,sys
from pathlib import Path
DMG='b050706e51efa0b3dbce2c86557f9ba53176a875e779bd0389e8bdc29685ff7c'
EXPECTED='2d1cf296a34fff329f6ec7e0a9459b620ad038c259aa2c48fbfd40c6834aad77'
def prepare(root):
    record={'kind':'7bd-resource-baseline-not-ee258-bundle','source_run':40296602029,'artifact_ids':[12639059961,12638984059]}
    for subtree,pattern,name,digest in [('resources','*.dmg','baseline.dmg',DMG),('smoke','expected-app-arm64.json','expected-app-arm64.json',EXPECTED)]:
        found=list((root/subtree).rglob(pattern))
        if len(found)!=1 or found[0].is_symlink() or hashlib.sha256(found[0].read_bytes()).hexdigest()!=digest:raise ValueError('resource exact hash/count')
        with (root/name).open('xb') as dst,found[0].open('rb') as src:shutil.copyfileobj(src,dst)
        record[name]=digest
    (root/'resource-source.json').write_text(json.dumps(record,indent=2)+'\n')
if __name__=='__main__':prepare(Path(sys.argv[1]))
