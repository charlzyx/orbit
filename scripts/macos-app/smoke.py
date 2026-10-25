#!/usr/bin/env python3
"""Launch the final-DMG copy via LaunchServices twice; never bypass OS security."""
from pathlib import Path
import argparse,json,os,platform,plistlib,shutil,subprocess,tempfile,time,struct,re
from datetime import datetime
from common import digest,write,run
from bundle import APP_NAME,tree_hashes,require_arch,SYSTEM_PATH,ARCHITECTURES

def verify_diagnostics(path,version,launches):
 if not path.is_file() or path.is_symlink() or not 0<path.stat().st_size<=1024*1024:raise ValueError('Missing or oversized private app diagnostics')
 rows=[json.loads(line) for line in path.read_text().splitlines()]
 allowed_events={'app_startup','app_shutdown','native_run','controller_init'}
 allowed_results={'started','ok','cancelled','error','skipped','busy','timeout'}
 runs={}
 for row in rows:
  if set(row)!={'timestamp','run_id','version','event','result','duration_ms'}:raise ValueError('Unexpected diagnostic fields')
  if row['event'] not in allowed_events or row['result'] not in allowed_results or row['version']!=version:raise ValueError('Unexpected diagnostic event/result/version')
  if not re.fullmatch('[0-9a-f]{32}',row['run_id']) or not isinstance(row['duration_ms'],int) or not 0<=row['duration_ms']<=86400000:raise ValueError('Invalid diagnostic run or duration')
  datetime.fromisoformat(row['timestamp'].replace('Z','+00:00'))
  runs.setdefault(row['run_id'],set()).add((row['event'],row['result']))
 required={('app_startup','started'),('controller_init','ok'),('native_run','ok'),('app_shutdown','ok')}
 if len(runs)!=launches or any(not required<=events for events in runs.values()):raise ValueError('Missing fresh startup/window/shutdown diagnostic stages')
 return rows

def verify_report(report,app,home,arch,started,report_file):
 if report.get('status')!='PASS' or report.get('architecture')!=arch:raise ValueError('Native smoke did not pass on expected architecture')
 if not isinstance(report.get('pid'),int) or report['pid']<=0:raise ValueError('Missing live app PID')
 checks=report.get('checks',{})
 if not all(checks.get(k) is True for k in ('native_window_ready','first_frame','ghostty_load','pi_rpc_get_state','native_capture')):raise ValueError('Incomplete native UI/Ghostty/Pi evidence')
 for key,rel in {'executable':'Contents/MacOS/Orbit','ghostty_path':'Contents/Frameworks/libghostty-vt.dylib','node_path':'Contents/Resources/node/bin/node','bundled_pi_path':'Contents/Resources/bin/pi'}.items():
  # The bundled Pi report may point at its .bin entry rather than shim.
  actual=Path(report.get(key,''))
  if key=='bundled_pi_path':
   if not actual.resolve().is_relative_to((app/'Contents/Resources').resolve()):raise ValueError('Bundled Pi escaped installed app')
  elif actual.resolve()!=(app/rel).resolve():raise ValueError('Dependency not loaded from installed app: '+key)
 expected_home=home/'Library/Application Support/Orbit Go Test/Home'
 if Path(report.get('home','')).resolve()!=expected_home.resolve():raise ValueError('Private Orbit preference home changed')
 if Path(report.get('user_home','')).resolve()!=home.resolve():raise ValueError('Pi HOME does not match synthetic user home')
 if Path(report.get('pi_agent_dir','')).resolve()!=(home/'.pi/agent').resolve():raise ValueError('Pi config does not match synthetic user config')
 if not Path(report.get('workspace','')).resolve().is_relative_to(expected_home.resolve()):raise ValueError('Default workspace escaped private home')
 if not Path(report.get('pi_path','')).resolve().is_relative_to((expected_home/'runtime').resolve()):raise ValueError('Selected Pi must use private materialized runtime')
 if not report.get('started_at') or not report.get('session_id'):raise ValueError('Missing launch/session identity')
 observed=datetime.fromisoformat(report['started_at'].replace('Z','+00:00')).timestamp()
 if observed<started-5 or observed>time.time()+5:raise ValueError('Stale or future launch timestamp')
 if report.get('capture_kind')!='native-content':raise ValueError('Unexpected capture evidence kind')
 capture=Path(str(report_file)+'.png')
 if Path(report.get('capture_path','')).resolve()!=capture.resolve() or not capture.is_file() or capture.is_symlink() or digest(capture)!=report.get('capture_sha256'):raise ValueError('Native capture path/hash mismatch')
 with capture.open('rb') as image:header=image.read(33)
 if len(header)!=33 or header[:8]!=b'\x89PNG\r\n\x1a\n' or header[12:16]!=b'IHDR':raise ValueError('Native capture is not PNG')
 width,height=struct.unpack('>II',header[16:24])
 if not width or not height or width!=report.get('capture_width') or height!=report.get('capture_height'):raise ValueError('Native capture dimensions mismatch')


def smoke(folder,arch):
 if arch not in ARCHITECTURES:raise ValueError('Future macOS app smoke supports arm64 only')
 if platform.system()!='Darwin' or platform.machine()!='arm64':raise ValueError('Wrong native runner')
 package=json.loads((folder/f'package-{arch}.json').read_text());dmg=folder/package['filename']
 if digest(dmg)!=package['sha256']:raise ValueError('DMG changed before smoke')
 expected=json.loads((folder/f'expected-app-{arch}.json').read_text())
 with tempfile.TemporaryDirectory(prefix='Orbit Finder Smoke ') as temp:
  temp=Path(temp);mount=temp/'Mounted Image';installed=temp/'Installed Applications'/APP_NAME;home=temp/'Clean User Home';home.mkdir();mount.mkdir();installed.parent.mkdir()
  run('/usr/bin/hdiutil','attach','-readonly','-nobrowse','-mountpoint',mount,dmg)
  try:
   if not (mount/'Applications').is_symlink() or os.readlink(mount/'Applications')!='/Applications':raise ValueError('DMG Applications drag target missing')
   run('/usr/bin/ditto',mount/APP_NAME,installed)
  finally:run('/usr/bin/hdiutil','detach',mount)
  if tree_hashes(installed)!=expected:raise ValueError('Copied DMG app bytes/modes differ')
  if (folder/'.volume').exists():raise ValueError('Build staging app still exists; smoke could resolve wrong copy')
  require_arch(installed/'Contents/MacOS/Orbit',arch);require_arch(installed/'Contents/Resources/node/bin/node',arch);require_arch(installed/'Contents/Frameworks/libghostty-vt.dylib',arch)
  env={'HOME':str(home),'PATH':SYSTEM_PATH,'TMPDIR':str(temp),'LANG':'en_US.UTF-8'}
  # Validate --env on this actual OS rather than assuming shell env propagates
  # through LaunchServices. No launchctl environment or security setting changes.
  helptext=subprocess.run(['/usr/bin/open','-h'],text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT).stdout
  if '--env' not in helptext:raise ValueError('LaunchServices open --env unavailable; default-home smoke not established')
  reports=[]
  for i in (1,2):
   report=temp/f'launch-{i}.json';started=time.time()
   command=['/usr/bin/open','-n','-W','--env','HOME='+str(home),'--env','PATH='+SYSTEM_PATH,'--env','TMPDIR='+str(temp),'--env','PI_CODING_AGENT_DIR='+str(home/'.pi/agent'),'--env','PI_OFFLINE=1','--env','PI_SKIP_VERSION_CHECK=1',str(installed),'--args','--smoke-report',str(report),'--smoke-exit-after','30s']
   result=subprocess.run(command,cwd='/',env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=65)
   write(folder/f'launch-{arch}-{i}.log',result.stdout)
   if result.returncode or not report.is_file() or report.stat().st_mtime<started:raise ValueError('Finder launch/report failed; no native acceptance claim')
   data=json.loads(report.read_text());verify_report(data,installed,home,arch,started,report);reports.append(data)
   diagnostics=home/'Library/Application Support/Orbit Go Test/Home/.orbit-pi/logs/app.log'
   verify_diagnostics(diagnostics,package['version'],i)
   shutil.copyfile(diagnostics,folder/f'launch-{arch}-{i}-diagnostics.log')
   shutil.copyfile(Path(str(report)+'.png'),folder/f'launch-{arch}-{i}.png')
   if tree_hashes(installed)!=expected:raise ValueError('App bundle mutated during first/second launch')
   write(folder/f'launch-{arch}-{i}.json',json.dumps(data,indent=2)+'\n')
  if reports[0]['pid']==reports[1]['pid']:raise ValueError('Second launch reused stale process evidence')
  if reports[0]['home']!=reports[1]['home'] or reports[0]['workspace']!=reports[1]['workspace'] or reports[0]['pi_path']!=reports[1]['pi_path'] or reports[0]['user_home']!=reports[1]['user_home'] or reports[0]['pi_agent_dir']!=reports[1]['pi_agent_dir']:raise ValueError('Default persistent state changed unexpectedly on second launch')
 package['native_macos_smoke']='PASS';package['smoke_checks']=['final-DMG-integrity','mounted-read-only','copy-installed-with-spaces','DMG-detached-before-open','matching-native-architecture','native-window-visible-first-frame','native-window-PNG','bundled-Ghostty-load','materialized-Pi-RPC-get-state','private-Orbit-state-normal-synthetic-Pi-home','bounded-private-stage-diagnostics','second-launch-persistence','app-tree-unchanged'];package['runner_os']=platform.mac_ver()[0]
 write(folder/f'package-{arch}.json',json.dumps(package,indent=2)+'\n')
 print('PASS native Finder/DMG smoke:',arch)
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--folder',type=Path,required=True);p.add_argument('--arch',choices=ARCHITECTURES,required=True);a=p.parse_args();smoke(a.folder.resolve(),a.arch)
