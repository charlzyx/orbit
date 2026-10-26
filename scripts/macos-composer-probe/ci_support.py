"""Offline-testable, bounded preparation/evidence helpers; no native actions."""
import argparse,hashlib,json,os,re,shutil,stat,zipfile
from pathlib import Path

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def copy_harness(source, target, manifest, expected):
    if sha(manifest)!=expected: raise ValueError('harness manifest hash')
    files=json.loads(manifest.read_text())['files']
    if source.is_symlink() or any(p.is_symlink() for p in source.rglob('*')):raise ValueError('harness symlink')
    actual={p.relative_to(source).as_posix() for p in source.rglob('*') if p.is_file() and '__pycache__' not in p.parts and p!=manifest}
    if actual!=set(files):raise ValueError('harness file set')
    for name,digest in files.items():
        p=Path(name)
        if p.is_absolute() or '..' in p.parts or source.joinpath(p).is_symlink() or sha(source/p)!=digest:raise ValueError('harness path/hash')
    if target.exists():raise ValueError('target exists')
    target.mkdir(parents=True)
    for name in files:
        out=target/name;out.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(source/name,out)
    shutil.copyfile(manifest,target/manifest.name)

def extract_known(archive, expected, output):
    if sha(archive)!=expected or output.exists():raise ValueError('archive hash/destination')
    with zipfile.ZipFile(archive) as z:
        entries=z.infolist()
        if len(entries)>10000 or sum(x.file_size for x in entries)>512*1024*1024:raise ValueError('archive budget')
        seen=set()
        for e in entries:
            p=Path(e.filename)
            mode=e.external_attr>>16
            if p.is_absolute() or '..' in p.parts or '\\' in e.filename or e.filename in seen or stat.S_ISLNK(mode) or e.file_size>256*1024*1024:raise ValueError('archive unsafe entry')
            seen.add(e.filename)
        output.mkdir(mode=0o700)
        for e in entries:
            p=output/e.filename
            if e.is_dir():p.mkdir(parents=True,exist_ok=True);continue
            p.parent.mkdir(parents=True,exist_ok=True)
            with z.open(e) as src,p.open('xb') as dst:shutil.copyfileobj(src,dst,1024*1024)

def collect(summary, output):
    data=json.loads(summary.read_text());cases=data.get('cases',[])
    if len(cases)>4:raise ValueError('case budget')
    output.mkdir(parents=True,exist_ok=True)
    for case in cases:
        root=Path(case['out_dir'])
        if not re.fullmatch('/private/tmp/orbit-composer-[0-9a-f]{64}',str(root)) or root.is_symlink():raise ValueError('unowned evidence root')
        rootstat=root.stat()
        if not stat.S_ISDIR(rootstat.st_mode) or rootstat.st_uid!=os.getuid() or stat.S_IMODE(rootstat.st_mode)!=0o700:raise ValueError('unsafe evidence root')
        names={'identity.json','runner-approved.json','config.json','report.json','runner-result.json','writer.json','writer-cleanup.json','native-process-proof.json','process-proof.stdout.raw','writer.stdout.raw','writer-cleanup.stdout.raw'}
        for phase in ('before','seed','final','undo'):
            names.update({phase+'.png',phase+'-os.png'})
            names.update(phase+'.os-'+x+'.json' for x in ('ready','before-request','before','after-request','after','proof','ack'))
        dest=output/root.name;dest.mkdir(exist_ok=False)
        for name in names:
            p=root/name
            if not p.exists():continue
            st=p.lstat()
            if not stat.S_ISREG(st.st_mode) or st.st_uid!=os.getuid() or st.st_size>16*1024*1024:raise ValueError('evidence unsafe')
            shutil.copyfile(p,dest/name)

def main():
    p=argparse.ArgumentParser();s=p.add_subparsers(dest='cmd',required=True)
    h=s.add_parser('copy-harness');h.add_argument('source',type=Path);h.add_argument('target',type=Path);h.add_argument('manifest',type=Path);h.add_argument('expected')
    z=s.add_parser('extract');z.add_argument('archive',type=Path);z.add_argument('expected');z.add_argument('output',type=Path)
    c=s.add_parser('collect');c.add_argument('summary',type=Path);c.add_argument('output',type=Path)
    a=p.parse_args()
    if a.cmd=='copy-harness':copy_harness(a.source,a.target,a.manifest,a.expected)
    elif a.cmd=='extract':extract_known(a.archive,a.expected,a.output)
    else:collect(a.summary,a.output)
if __name__=='__main__':main()
