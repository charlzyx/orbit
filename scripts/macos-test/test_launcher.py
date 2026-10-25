#!/usr/bin/env python3
"""Linux-only mocked launcher tests, explicitly NOT native application execution."""
from pathlib import Path
import tempfile, subprocess, hashlib, json, shutil, os
BASE=Path(__file__).resolve().parent
results=[]
for arch,label,machine in [('arm64','Apple Silicon','arm64'),('amd64','Intel','x86_64')]:
 with tempfile.TemporaryDirectory(prefix='orbit-launcher-test-') as tmp:
  root=Path(tmp); mac=root/'Package with spaces/Orbit Go Test.app/Contents/MacOS'; res=mac.parent/'Resources'; mac.mkdir(parents=True); (res/'sample-workspace').mkdir(parents=True)
  exe=mac/'orbit-bin'; exe.write_text('#!/bin/bash\nprintf "ARG:%s\\n" "$@"\n/usr/bin/env | /usr/bin/sort\n');exe.chmod(0o755)
  lib=res/'libghostty-vt.dylib';lib.write_text('not a dylib; mock input for shell isolation checks only\n')
  (res/'sample-workspace/sample.txt').write_text('synthetic\n')
  uname=root/'uname';uname.write_text(f'#!/bin/sh\ncase "$1" in -s) echo Darwin;; -m) echo {machine};; esac\n');uname.chmod(0o755)
  sw=root/'sw_vers';sw.write_text('#!/bin/sh\necho 13.0.0\n');sw.chmod(0o755)
  script=(BASE/'launcher.sh.in').read_text().replace('/usr/bin/uname',f'"{uname}"').replace('/usr/bin/sw_vers',f'"{sw}"')
  values={'__ARCH__':arch,'__ARCHLABEL__':label,'__MACHINE__':machine,'__APPHASH__':hashlib.sha256(exe.read_bytes()).hexdigest(),'__LIBHASH__':hashlib.sha256(lib.read_bytes()).hexdigest()}
  for k,v in values.items():script=script.replace(k,v)
  launch=mac/'Orbit-Go-Test';launch.write_text(script);launch.chmod(0o755)
  env=dict(os.environ);env.update({'OPENAI_API_KEY':'SYNTHETIC-SENTINEL','SSH_AUTH_SOCK':'SYNTHETIC-SENTINEL','MYGO_DEV_URL':'SYNTHETIC-SENTINEL','PI_BIN':'SYNTHETIC-SENTINEL','PI_CODING_AGENT_DIR':'SYNTHETIC-SENTINEL','ZDOTDIR':'SYNTHETIC-SENTINEL'})
  def run(*args):return subprocess.run([str(launch),*args],env=env,text=True,capture_output=True)
  r=run('--help');assert r.returncode==0
  for args in [('--pi',),('--pi','relative'),('--unknown',),('--workspace','relative')]:assert run(*args).returncode==2,args
  r=run(); assert r.returncode==0,r.stderr
  state=Path(next(x.split(': ',1)[1] for x in r.stdout.splitlines() if x.startswith('Test data:')))
  log=(state/'launch.log').read_text(); assert 'SYNTHETIC-SENTINEL' not in log
  assert 'ARG:--offline' in log and f'HOME={state}/home' in log and f'PI_BIN={state}/no-pi-selected' in log and f'MYGO_GHOSTTY_VT={lib}' in log
  assert state.stat().st_mode & 0o777 == 0o700
  assert (state/'workspace/sample.txt').read_text()=='synthetic\n'
  assert not (res/'sample-workspace/.pi').exists()
  shutil.rmtree(state)
  pi=root/'pi executable';pi.write_text('#!/bin/sh\nexit 0\n');pi.chmod(0o755)
  workspace=root/'disposable project';workspace.mkdir();(workspace/'untouched').write_text('sentinel')
  r=run('--pi',str(pi),'--workspace',str(workspace));assert r.returncode==0,r.stderr
  state=Path(next(x.split(': ',1)[1] for x in r.stdout.splitlines() if x.startswith('Test data:')))
  log=(state/'launch.log').read_text();assert f'ARG:{workspace}' in log and f'ARG:{pi}' in log and 'SYNTHETIC-SENTINEL' not in log
  assert (workspace/'untouched').read_text()=='sentinel';shutil.rmtree(state)
  lib.write_text('tampered');r=run();assert r.returncode==1 and 'Integrity mismatch:' in r.stderr
  sw.write_text('#!/bin/sh\necho 12.7.6\n');r=run();assert r.returncode==1 and 'macOS 13.0' in r.stderr
  sw.write_text('#!/bin/sh\necho 13.0.0\n');uname.write_text('#!/bin/sh\ncase "$1" in -s) echo Darwin;; -m) echo wrong;; esac\n');r=run();assert r.returncode==1 and 'Wrong architecture:' in r.stderr
  uname.write_text('#!/bin/sh\necho Linux\n');r=run();assert r.returncode==1 and 'for macOS' in r.stderr
  results.append({'arch':arch,'mocked_contract_test':'PASS','checks':['argument validation','fresh private test HOME/workspace','ambient secret/environment exclusion','explicit Pi and disposable workspace with spaces','no synthetic-source modification','file hash fail closed','minimum OS guard','architecture guard','platform guard'],'native_execution':'NOT RUN; uname/sw_vers and orbit binary replaced in test-only copies'})

print(json.dumps(results,indent=2))
