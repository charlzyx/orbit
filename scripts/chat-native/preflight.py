#!/usr/bin/env python3
"""Read metadata through real Pi1.4 JSONL only; never prompt/compact/model/auth."""
import argparse
import json
import os
from pathlib import Path
import selectors
import signal
import subprocess
import time
import materialize

ALLOWED = frozenset(['switch_session','get_state','get_messages','get_session_stats'])

def probe(binary, root, output):
    root,output=Path(root),Path(output)
    stderr_path=output.with_suffix('.stderr.log')
    if any(os.path.lexists(p) for p in (output,stderr_path)): raise ValueError('Refusing reused preflight evidence')
    _, session=materialize.home_files();path=root/'home'/session
    env=dict(PATH=os.environ['PATH'],HOME=str(root/'home'),PI_CODING_AGENT_DIR=str(root/'home/.pi/agent'),PI_OFFLINE='1',PI_SKIP_VERSION_CHECK='1',LANG='C.UTF-8',LC_ALL='C.UTF-8',TZ='UTC')
    version=subprocess.check_output([str(binary),'--version'],env=env,cwd=root/'workspace',timeout=10,text=True).strip()
    if version != '1.4.0':raise ValueError('Expected pinned actual Pi1.4.0, got '+version)
    stderr_file=stderr_path.open('xb')
    process=subprocess.Popen([str(binary),'--mode','rpc'],cwd=root/'workspace',env=env,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=stderr_file,start_new_session=True)
    selector=selectors.DefaultSelector();selector.register(process.stdout,selectors.EVENT_READ)
    records=[];buffer=b'';received=0
    def request(kind,**fields):
        nonlocal buffer,received
        if kind not in ALLOWED: raise ValueError('Forbidden capture preflight command')
        command=dict(id=f'chat-preflight-{len(records)+1}',type=kind,**fields)
        process.stdin.write((json.dumps(command)+'\n').encode());process.stdin.flush();deadline=time.monotonic()+20
        while time.monotonic()<deadline:
            if b'\n' not in buffer:
                if not selector.select(max(0,deadline-time.monotonic())): break
                data=os.read(process.stdout.fileno(),65536)
                if not data: raise RuntimeError('Pi exited before metadata reply')
                buffer+=data;received+=len(data)
                if received>2*1024*1024:raise ValueError('Pi metadata reply exceeded 2 MiB bound')
                continue
            line,buffer=buffer.split(b'\n',1);reply=json.loads(line)
            if reply.get('id')!=command['id']: continue
            records.append(dict(command=command,response=reply))
            if reply.get('type')!='response' or reply.get('success') is not True:raise ValueError('Pi metadata request rejected: '+str(reply))
            return reply.get('data',{})
        raise TimeoutError('Pi metadata request timed out: '+kind)
    try:
        request('switch_session',sessionPath=str(path));state=request('get_state');messages=request('get_messages');stats=request('get_session_stats')
        if state.get('sessionId')!=materialize.SESSION_ID:raise ValueError('Pi selected unexpected session identity')
        if len(messages.get('messages',[]))!=24:raise ValueError('Pi transcript message count differs')
        if state.get('model',{}).get('id')!='unknown' or stats.get('contextUsage') is not None:
            raise ValueError('Credential-free fixture unexpectedly acquired a known context/model')
        if stats.get('tokens',{}).get('total')!=411000:raise ValueError('Pi session accounting differs')
        value=dict(schema=1,status='READ_ONLY_PI_METADATA_VERIFIED',piVersion=version,nodeVersion=subprocess.check_output(['node','--version'],env=env,text=True,timeout=5).strip(),binary=str(Path(binary).resolve()),binarySha256=materialize.digest(Path(binary).resolve().read_bytes()),commands=records,contextDetails='BLOCKED_NO_VERIFIED_TARGET_WITHOUT_PROVIDER_AUTH',uiState='NOT_CAPTURED',pixelEquality='NOT_COMPARED')
        output.write_text(materialize.pretty(value));return value
    except Exception as error:
        output.write_text(materialize.pretty(dict(schema=1,status='BLOCKED',error=str(error),commands=records,contextDetails='BLOCKED_NO_VERIFIED_TARGET_WITHOUT_PROVIDER_AUTH',uiState='NOT_CAPTURED',pixelEquality='NOT_COMPARED')))
        raise
    finally:
        selector.close()
        if process.poll() is None:os.killpg(process.pid,signal.SIGTERM)
        try:process.communicate(timeout=5)
        except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);process.communicate()
        stderr_file.close()

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary');p.add_argument('root',type=Path);p.add_argument('output',type=Path);a=p.parse_args();probe(a.binary,a.root,a.output)
if __name__=='__main__':main()
