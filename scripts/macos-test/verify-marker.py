#!/usr/bin/env python3
from pathlib import Path
import os,subprocess,json
from package import marker,runtime_inputs,git
root=Path(__file__).resolve().parents[2]
if os.environ.get('GITHUB_REPOSITORY')!='charlzyx/orbit':raise SystemExit('Only charlzyx/orbit is authorized')
if os.environ.get('GITHUB_EVENT_NAME')!='push' or os.environ.get('GITHUB_REF')!='refs/heads/redesign/mygo-rpc':raise SystemExit('Only explicit redesign/mygo-rpc pushes may request this prerelease')
commit=os.environ['GITHUB_SHA'];before=os.environ.get('EVENT_BEFORE','')
if git(root,'rev-parse','HEAD')!=commit:raise SystemExit('Checkout mismatch')
if before and before!='0'*40:
 changed=git(root,'diff','--name-only',before,commit,'--','.github/macos-test-prerelease.json').splitlines()
else:changed=git(root,'show','--pretty=format:','--name-only',commit,'--','.github/macos-test-prerelease.json').splitlines()
if '.github/macos-test-prerelease.json' not in changed:raise SystemExit('Version marker did not change in this push; no release')
m=marker(root);_,actual=runtime_inputs(root)
if actual!=m['runtime_source_sha256']:raise SystemExit('Runtime-source fingerprint mismatch')
if before and before!='0'*40:
 old=subprocess.run(['git','-C',str(root),'show',before+':.github/macos-test-prerelease.json'],text=True,capture_output=True)
 if old.returncode==0 and json.loads(old.stdout).get('version')==m['version']:raise SystemExit('Version was not advanced; refusing repeat marker release')
with open(os.environ['GITHUB_OUTPUT'],'a') as f:f.write('version='+m['version']+'\ntag=macos-test-v'+m['version']+'\n')
print('Explicit marker accepted:',m['version'],'source',commit,'runtime',actual)
