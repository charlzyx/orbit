"""Shared, fail-closed helpers for the separately versioned Finder app release."""
from pathlib import Path
import hashlib, importlib.util, json, re, subprocess, sys
HERE=Path(__file__).resolve().parent
LEGACY=HERE.parent/'macos-test'
sys.path.insert(0,str(LEGACY))
from package import runtime_inputs, native_download, COMMIT_RE, VERSION_RE
from macho import inspect
MARKER='.github/macos-app-prerelease.json'
def digest(path):
 h=hashlib.sha256()
 with Path(path).open('rb') as f:
  for chunk in iter(lambda:f.read(1024*1024),b''):h.update(chunk)
 return h.hexdigest()
def git(root,*args):return subprocess.check_output(['git','-C',str(root),*args],text=True).strip()
def marker(root):
 m=json.loads((root/MARKER).read_text())
 if m.get('enabled') is not True or not VERSION_RE.fullmatch(m.get('version','')):raise ValueError('Invalid explicit experimental version')
 if m['version']=='0.2.12-go-test.20261025.1':raise ValueError('First release is immutable; require a new version')
 if not re.fullmatch('[0-9a-f]{64}',m.get('runtime_source_sha256','')):raise ValueError('Source fingerprint must be finalized before publication')
 if m.get('minimum_macos')!='13.0' or m.get('required_native_smoke') is not True:raise ValueError('Native macOS smoke must gate publication')
 return m

def write(p,s,mode=0o644):p.parent.mkdir(parents=True,exist_ok=True);p.write_text(s);p.chmod(mode)
def run(*args,**kwargs):return subprocess.run([str(a) for a in args],check=True,**kwargs)
def check_source(root,commit):
 if not COMMIT_RE.fullmatch(commit) or git(root,'rev-parse','HEAD')!=commit or git(root,'status','--porcelain','--untracked-files=no'):raise ValueError('Exact clean committed source required')
 m=marker(root);inputs,h=runtime_inputs(root)
 if h!=m['runtime_source_sha256']:raise ValueError('Runtime fingerprint mismatch')
 return m,inputs
def copy(a,b,mode=0o644):
 import shutil
 b.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(a,b);b.chmod(mode)
