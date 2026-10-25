#!/usr/bin/env python3
"""Trusted isolated-fixture provenance, NOT authentication against hostile same UID."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import secrets
import stat
import sys
import time

MAX_FILES=4096
MAX_BYTES=256*1024*1024
MAX_RECORDS=16
MAX_JSON=2*1024*1024
LAUNCHER='.orbit-chat-launcher.py'

def digest(data): return hashlib.sha256(data).hexdigest()
def canonical(value): return json.dumps(value,sort_keys=True,separators=(',',':')).encode()
def identity(pid):
    p=Path('/proc')/str(pid)
    return dict(pid=pid,startTicks=(p/'stat').read_text().rsplit(')',1)[1].split()[19],exe=str((p/'exe').resolve(strict=True)))
def stamp(p):
    s=p.lstat()
    if not stat.S_ISREG(s.st_mode): raise ValueError('Not a regular pinned file')
    return [s.st_dev,s.st_ino,s.st_size,s.st_mtime_ns,s.st_ctime_ns]
def directory_identity(p):
    s=Path(p).lstat()
    if not stat.S_ISDIR(s.st_mode):raise ValueError("Pinned directory replaced")
    return [s.st_dev,s.st_ino,s.st_uid]

def file_record(p):
    before=stamp(p)
    if before[2]>MAX_BYTES: raise ValueError('Pinned file too large')
    with p.open("rb") as stream:data=stream.read(MAX_BYTES+1)
    if len(data)>MAX_BYTES:raise ValueError("Pinned file too large")
    if stamp(p)!=before: raise ValueError('Pinned file changed during hash')
    return dict(stat=before,sha256=digest(data))
def inventory(package):
    result={};total=0
    # Reject links rather than follow arbitrary trees; installed Pi package has
    # plain files. Dependencies are not recursively traversed by this inventory.
    for count,p in enumerate(package.rglob('*')):
        if count>=MAX_FILES*2:raise ValueError('Package traversal bound')
        if p.name==LAUNCHER and p.parent==package: continue
        if p.is_symlink(): raise ValueError('Symlink inside pinned package')
        if p.is_dir(): continue
        total+=p.stat().st_size
        if len(result)>=MAX_FILES or total>MAX_BYTES: raise ValueError('Package inventory exceeded bound')
        result[str(p.relative_to(package))]=file_record(p)
    return result

def read_json(path,limit=MAX_JSON):
    fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW)
    try:
        s=os.fstat(fd)
        if not stat.S_ISREG(s.st_mode) or s.st_uid!=os.getuid() or s.st_mode&0o077 or s.st_size>limit: raise ValueError('Unsafe receipt/config file')
        data=os.read(fd,limit+1)
        if len(data)>limit: raise ValueError('Oversized receipt/config')
        return json.loads(data),data
    finally: os.close(fd)

def verify_expected(path,expected_hash):
    cfg,data=read_json(path)
    if digest(data)!=expected_hash or cfg.get('schema')!=1: raise ValueError('Expected manifest changed')
    return cfg

def expected_args(kind,fixture):
    home=fixture/'home'
    args=['--mode','rpc','--approve']
    if kind=='go': args+=['--offline','--session-dir',str(home/'.pi/agent/sessions')]
    elif kind!='reference': raise ValueError('Unknown renderer')
    for name in ('quota','guard','title','workflow'):
        args+=['--extension',str(home/'.orbit-pi'/f'{name}-extension/index.js')]
    return args

def launcher_source(module,expected,expected_hash):
    return f'#!{sys.executable}\nimport runpy,sys\nsys.argv=[{module!r},"launch",{str(expected)!r},{expected_hash!r},*sys.argv[1:]]\nrunpy.run_path({module!r},run_name="__main__")\n'

def prepare(package,node,fixture,ledger,kind,approved):
    package,node,fixture=map(lambda p:Path(p).resolve(strict=True),(package,node,fixture))
    ledger=Path(ledger)
    if ledger.exists() or ledger.is_symlink(): raise ValueError('Ledger already exists')
    if (package/LAUNCHER).exists() or (package/LAUNCHER).is_symlink(): raise ValueError('Launcher already exists')
    metadata=json.loads((package/'package.json').read_text())
    if metadata.get('name')!='@earendil-works/pi-coding-agent' or metadata.get('version')!='1.4.0': raise ValueError('Not pinned Pi package/version')
    entry=package/'dist/bundle/cli.js'
    records=inventory(package)
    if {key:value['sha256'] for key,value in records.items()}!=approved['files']:raise ValueError('Installed package differs from independent vendor + source patch reference')
    ledger.mkdir(mode=0o700)
    cfg=dict(schema=1,nonce=secrets.token_hex(32),createdMonotonic=time.monotonic(),package=str(package),packageDirectory=directory_identity(package),ledgerDirectory=directory_identity(ledger),node=str(node),entry=str(entry),cwd=str(fixture/'workspace'),args=expected_args(kind,fixture),files=records,provenance={key:value for key,value in approved.items() if key!='files'},nodeFile=file_record(node),ledger=str(ledger.resolve()),kind=kind,module=str(Path(__file__).resolve()),moduleFile=file_record(Path(__file__).resolve()))
    # Expected content is prepared by the harness, independently of any child.
    raw=canonical(cfg); expected=ledger/'expected.json'
    fd=os.open(expected,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW,0o600)
    with os.fdopen(fd,'wb') as f: f.write(raw)
    module=str(Path(__file__).resolve())
    launcher=launcher_source(module,expected,digest(raw))
    fd=os.open(package/LAUNCHER,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW,0o700)
    with os.fdopen(fd,'w') as f: f.write(launcher)
    return dict(expected=str(expected),expectedSha256=digest(raw),launcher=str(package/LAUNCHER))

class Verifier:
    def __init__(self,path,expected_hash):
        self.path=Path(path);self.expected_hash=expected_hash
        self.cfg=verify_expected(path,expected_hash)
        self.checked=False
        self.package_stamps={}
        self.hashed_bytes=0
        self.changed_files=0
    def check_files(self):
        cfg=self.cfg
        verify_expected(self.path,self.expected_hash)
        if directory_identity(cfg['package'])!=cfg['packageDirectory'] or directory_identity(cfg['ledger'])!=cfg['ledgerDirectory']:raise ValueError('Pinned directory identity changed')
        directory=Path(cfg['ledger']).lstat()
        if not stat.S_ISDIR(directory.st_mode) or directory.st_uid!=os.getuid() or directory.st_mode&0o077: raise ValueError('Unsafe ledger')
        for count,_ in enumerate(Path(cfg['ledger']).iterdir()):
            if count>=MAX_RECORDS+1:raise ValueError('Receipt fanout bound')
        # Content hash once per verifier; thereafter exact inode/ctime/mtime/size
        # changes get a bounded rehash against the ORIGINAL expected digest.
        # Same-byte patch rewrites are allowed; new content is never blessed.
        package=Path(cfg['package'])
        actual=set()
        for count,path in enumerate(package.rglob('*')):
            if count>=MAX_FILES*2:raise ValueError('Package traversal bound')
            if path.name==LAUNCHER and path.parent==package:continue
            if path.is_symlink():raise ValueError('Pinned package gained symlink')
            if not path.is_dir():actual.add(str(path.relative_to(package)))
            if len(actual)>MAX_FILES:raise ValueError('Package fanout bound')
        if actual!=set(cfg['files']):raise ValueError('Pinned package file set changed')
        launcher=package/LAUNCHER
        if launcher.is_symlink() or launcher.read_text()!=launcher_source(cfg['module'],self.path,self.expected_hash):raise ValueError('Launcher changed')
        module=Path(cfg['module'])
        if stamp(module)!=cfg['moduleFile']['stat']:raise ValueError('Receipt module changed')
        if not self.checked and file_record(module)!=cfg['moduleFile']:raise ValueError('Receipt module hash changed')
        for rel,record in cfg['files'].items():
            path=Path(cfg['package'])/rel
            current=stamp(path)
            if self.package_stamps.get(rel)!=current:
                if rel in self.package_stamps or current!=record['stat']:self.changed_files+=1
                self.hashed_bytes+=current[2]
                if self.changed_files>32 or self.hashed_bytes>2*MAX_BYTES:raise ValueError('Pinned package rehash budget exceeded')
                verified=file_record(path)
                if verified['sha256']!=record['sha256']:raise ValueError('Pinned package hash changed')
                self.package_stamps[rel]=verified['stat']
        node=Path(cfg['node'])
        if stamp(node)!=cfg['nodeFile']['stat']: raise ValueError('Pinned Node changed')
        if not self.checked and file_record(node)!=cfg['nodeFile']: raise ValueError('Pinned Node hash changed')
        self.checked=True
    def candidate(self,ident,parent):
        self.check_files();cfg=self.cfg
        if ident['exe']!=cfg['node']: return None
        path=Path(cfg['ledger'])/f"{ident['pid']}-{ident['startTicks']}.json"
        try: rec,_=read_json(path,32768)
        except FileNotFoundError:return None
        if rec.get('schema')!=1 or rec.get('nonce')!=cfg['nonce'] or rec.get('expectedSha256')!=self.expected_hash: raise ValueError('Foreign receipt')
        if rec.get('identity')!={'pid':ident['pid'],'startTicks':ident['startTicks']} or rec.get('parent')!=parent: raise ValueError('Receipt lineage changed')
        if rec.get('args')!=cfg['args'] or rec.get('node')!=cfg['node'] or rec.get('entry')!=cfg['entry']: raise ValueError('Receipt launch mismatch')
        when=rec.get('createdMonotonic')
        if not isinstance(when,(float,int)) or not cfg['createdMonotonic']<=when<=time.monotonic(): raise ValueError('Receipt time invalid')
        return dict(receipt=path.name,expectedSha256=self.expected_hash)

def launch(path,expected_hash,args):
    verifier=Verifier(path,expected_hash);verifier.check_files();cfg=verifier.cfg
    if args==['--version']:
        os.execve(cfg['node'],[cfg['node'],cfg['entry'],*args],os.environ.copy())
    if args!=cfg['args'] or os.getcwd()!=cfg['cwd']: raise ValueError('Unapproved runtime arguments/workspace')
    ident=identity(os.getpid());parent=identity(os.getppid())
    rec=dict(schema=1,nonce=cfg['nonce'],expectedSha256=expected_hash,identity=dict(pid=ident['pid'],startTicks=ident['startTicks']),parent=parent,args=args,node=cfg['node'],entry=cfg['entry'],createdMonotonic=time.monotonic())
    final=Path(cfg['ledger'])/f"{ident['pid']}-{ident['startTicks']}.json"
    temp=final.with_suffix('.pending')
    fd=os.open(temp,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW,0o600)
    with os.fdopen(fd,'wb') as f:f.write(canonical(rec));f.flush();os.fsync(f.fileno())
    os.link(temp,final,follow_symlinks=False);temp.unlink()
    verifier.check_files()
    os.execve(cfg['node'],[cfg['node'],cfg['entry'],*args],os.environ.copy())

def main():
    if len(sys.argv)>1 and sys.argv[1]=='launch':
        launch(sys.argv[2],sys.argv[3],sys.argv[4:]);return
    p=argparse.ArgumentParser();p.add_argument('package');p.add_argument('node');p.add_argument('fixture');p.add_argument('ledger');p.add_argument('kind',choices=['go','reference']);p.add_argument('vendor_archive');p.add_argument('source_root');a=p.parse_args()
    import vendor_reference
    approved=vendor_reference.build(a.vendor_archive,a.node,a.source_root)
    print(json.dumps(prepare(a.package,a.node,a.fixture,a.ledger,a.kind,approved)))
if __name__=='__main__': main()
