"""Build expected Pi bytes from pinned vendor tar + reviewed renderer patch sources."""
import base64
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import subprocess
import tarfile
import tempfile

VENDOR_INTEGRITY='sha512-Z41LpZ3yYHSC/UUps6Mr39UMFPwH8elhJBG9DY1P7kuzsE5qSK/6eMj2EHzO5evJRQWLbKSqUMw+4nCNtIyiQg=='
VENDOR_URL='https://registry.npmjs.org/@earendil-works/pi-coding-agent/-/pi-coding-agent-1.4.0.tgz'
MAX_BYTES=256*1024*1024

def content_manifest(root):
    result={};total=0
    for count,path in enumerate(Path(root).rglob('*')):
        if count>=8192:raise ValueError('Reference traversal bound')
        if path.is_symlink():raise ValueError('Reference symlink')
        if path.is_dir():continue
        with path.open("rb") as stream:data=stream.read(MAX_BYTES+1)
        total+=len(data)
        if len(result)>=4096 or total>MAX_BYTES:raise ValueError('Reference content bound')
        result[str(path.relative_to(root))]=hashlib.sha256(data).hexdigest()
    return result

def build(archive,node,source):
    archive,node,source=map(lambda p:Path(p).resolve(strict=True),(archive,node,source))
    if archive.stat().st_size>64*1024*1024:raise ValueError('Vendor archive bound')
    with archive.open("rb") as stream:raw=stream.read(64*1024*1024+1)
    if len(raw)>64*1024*1024:raise ValueError("Vendor archive bound")
    if 'sha512-'+base64.b64encode(hashlib.sha512(raw).digest()).decode()!=VENDOR_INTEGRITY:raise ValueError('Vendor integrity mismatch')
    if subprocess.check_output([node,'--version'],timeout=5,text=True).strip()!='v22.23.4':raise ValueError('Expected CI Node22.23.4')
    with tempfile.TemporaryDirectory(prefix='orbit-pi-reference-') as temp:
        root=Path(temp);package=root/'package';package.mkdir();total=0;seen=set()
        with tarfile.open(fileobj=__import__('io').BytesIO(raw),mode='r:gz') as tar:
            for index,member in enumerate(tar):
                if index>=8192:raise ValueError('Vendor member bound')
                name=PurePosixPath(member.name)
                if name.is_absolute() or '..' in name.parts or not name.parts or name.parts[0]!='package':raise ValueError('Unsafe vendor path')
                relative=Path(*name.parts[1:]);target=package/relative
                if member.isdir():target.mkdir(parents=True,exist_ok=True);continue
                if not member.isfile() or str(relative) in seen:raise ValueError('Unsupported/duplicate vendor member')
                seen.add(str(relative));total+=member.size
                if len(seen)>4096 or total>MAX_BYTES:raise ValueError('Vendor unpack bound')
                target.parent.mkdir(parents=True,exist_ok=True)
                with tar.extractfile(member) as src, target.open('xb') as dst:
                    data=src.read(member.size+1)
                    if len(data)!=member.size:raise ValueError('Vendor member length mismatch')
                    dst.write(data)
        package_meta=json.loads((package/'package.json').read_text())
        if (package_meta.get('name'),package_meta.get('version'))!=('@earendil-works/pi-coding-agent','1.4.0'):raise ValueError('Wrong vendor package')
        patch_sources={}
        for name in ('custom-ui','quota','auth'):
            folder=source/'contrib'/f'pi-{name}-rpc'
            for rel,sha in content_manifest(folder).items():patch_sources[f'contrib/pi-{name}-rpc/{rel}']=sha
            env=dict(HOME=str(root),PATH=str(node.parent)+':/usr/bin:/bin',PI_BIN=str(package/'dist/bundle/cli.js'),PI_PACKAGE_ROOT=str(package),PI_OFFLINE='1',PI_SKIP_VERSION_CHECK='1',LANG='C.UTF-8')
            subprocess.run([node,folder/'apply.mjs'],env=env,cwd=root,timeout=20,check=True,capture_output=True)
            if {f'contrib/pi-{name}-rpc/{key}':value for key,value in content_manifest(folder).items()}!={key:value for key,value in patch_sources.items() if key.startswith(f'contrib/pi-{name}-rpc/')}:raise ValueError('Patch source changed during reference build')
        return dict(vendorIntegrity=VENDOR_INTEGRITY,vendorArchiveSha256=hashlib.sha256(raw).hexdigest(),patchSources=patch_sources,sourceRoot=str(source),files=content_manifest(package))
