#!/usr/bin/env python3
"""Only static Chat tail and previous turn; Context is explicitly blocked."""
import argparse
import importlib.util
import json
import os
from pathlib import Path
import time

HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('static_chat_base_capture',HERE.parent/'workbench-native/capture.py')
base=importlib.util.module_from_spec(spec);spec.loader.exec_module(base)
EXPECTED_ACTIONS=[[{'key':'ctrl+1'}],[{'key':'ctrl+Up'}]]
BEFORE={'CHAT01':['Static Chat fixture'],'CHAT02':['CHAT TAIL READY']}

def process_identity(pid):
    path=Path('/proc')/str(pid)
    # PID reuse must not turn an old screenshot into authority for another app.
    fields=(path/'stat').read_text().rsplit(')',1)[1].split()
    return {'pid':pid,'startTicks':fields[19],'exe':str((path/'exe').resolve(strict=True))}

def owned_rpc_children(pid,pi_entry):
    found=[];pending=[pid];seen=set()
    while pending:
        current=pending.pop()
        if current in seen:continue
        seen.add(current)
        if len(seen)>64:raise ValueError('Unexpected capture child process fanout')
        try:
            pending.extend(int(p) for p in Path(f'/proc/{current}/task/{current}/children').read_text().split())
            argv=Path(f'/proc/{current}/cmdline').read_bytes().split(b'\0')
            args=[s.decode('utf-8','replace')for s in argv if s]
            if any(args[i:i+2]==['--mode','rpc'] for i in range(len(args)-1)) and any(s.startswith('/') and Path(s).resolve()==pi_entry for s in args):
                found.append(dict(identity=process_identity(current),argv=args))
        except (FileNotFoundError,ProcessLookupError):continue
    return found

def validate_suite(suite):
    if suite.get('viewport')!=[1470,922] or suite.get('suite')!='static-native-chat-stage1':raise ValueError('Unexpected static Chat suite')
    cases=suite.get('cases',[])
    if [c.get('image')for c in cases]!=['chat-tail.png','chat-previous-turn.png']:raise ValueError('Unapproved capture output paths')
    if suite.get('settle')!={'consecutiveSamples':3,'pollSeconds':1,'timeoutSeconds':45}:raise ValueError('Unapproved readiness bounds')
    if [c.get('id')for c in cases]!=['CHAT01','CHAT02'] or [c.get('actions')for c in cases]!=EXPECTED_ACTIONS:
        raise ValueError('Only two approved static Chat inputs are supported; no prompt/compact/hover/drag/scroll/model/auth')

class Capture(base.Capture):
    def __init__(self,window,pid,out,suite,binary,pi_entry):
        validate_suite(suite)
        if any((Path(out)/p).exists() for p in ('actions.json','capture-result.json','evidence')):raise ValueError('Refusing prior capture evidence reuse')
        if pid<=0 or not str(window).isdigit():raise ValueError('Invalid native process/window identity')
        super().__init__(window,pid,out,suite)
        self.identity=process_identity(pid)
        self.pi_entry=Path(pi_entry).resolve(strict=True)
        if self.identity['exe']!=str(Path(binary).resolve(strict=True)):raise ValueError('Window process is not selected capture binary')

    def guard(self,active=False):
        self.alive()
        if process_identity(self.pid)!=self.identity:raise ValueError('Native process identity changed; no input')
        if base.run('xdotool','getwindowpid',self.window)!=str(self.pid):raise ValueError('Window PID does not match capture process; no input')
        visible=base.run('xdotool','search','--onlyvisible','--pid',str(self.pid)).splitlines()
        if self.window not in visible:raise ValueError('Selected native window is not visible; no input')
        if active and base.run('xdotool','getactivewindow')!=self.window:raise ValueError('Selected native window is not active; no input')
        geometry=dict(line.split('=',1)for line in base.run('xdotool','getwindowgeometry','--shell',self.window).splitlines()if '='in line)
        if [int(geometry['WIDTH']),int(geometry['HEIGHT'])]!=self.suite['viewport']:raise ValueError('Native client geometry changed; no input')

    def capture(self,path):
        self.guard();captured=time.monotonic()
        rows,metadata=super().capture(path)
        self.guard()
        return rows,dict(metadata,processIdentity=self.identity,capturedMonotonic=captured,rpcChildren=owned_rpc_children(self.pid,self.pi_entry))

    def readiness(self,image,rows,metadata,required,sample):
        if not metadata.get('rpcChildren'):
            return False,dict(missing=['live pinned Pi RPC descendant'],blocked=[]),metadata
        return super().readiness(image,rows,metadata,required,sample)

    def validate_input_proof(self,image,frame):
        self.guard(active=True)
        if frame.get('processIdentity')!=self.identity:raise ValueError('Foreign process screenshot; no input')
        age=time.monotonic()-frame['capturedMonotonic']
        if not 0<=age<=5:raise ValueError('Screenshot input proof expired; no input')
        if base.sha(image)!=frame['pngSha256']:raise ValueError('Screenshot changed after target verification; no input')

    def action(self,case_id,ordinal,action):
        expected=EXPECTED_ACTIONS[['CHAT01','CHAT02'].index(case_id)]
        if ordinal!=1 or action!=expected[0]:raise ValueError('Unapproved static Chat input')
        self.guard()
        base.run('xdotool','windowactivate','--sync',self.window)
        path=self.evidence/f'{case_id}-action-{ordinal}-before.png'
        if path.exists():raise ValueError('Refusing old action screenshot reuse')
        rows,frame=self.capture(path)
        ready,detail,frame=self.readiness(path,rows,frame,BEFORE[case_id],1)
        if not ready:raise ValueError('Static Chat input prerequisite failed: '+str(detail))
        event=dict(case=case_id,ordinal=ordinal,requested=action,beforeCapture=str(path.relative_to(self.out)),before=frame)
        command=['xdotool','key','--clearmodifiers',action['key']]
        self.validate_input_proof(path,frame)
        self.actions.append(dict(event,outcome='INPUT_REQUESTED'));self.write_actions()
        try:base.run(*command)
        except Exception:
            self.actions.append(dict(event,outcome='INPUT_RESULT_UNCERTAIN_NO_RETRY'));self.write_actions();raise
        self.actions.append(dict(event,outcome='INPUT_SENT_STATE_NOT_YET_VERIFIED'));self.write_actions()
        time.sleep(.25)

    def write_result(self,error=None):
        super().write_result(error)
        path=self.out/'capture-result.json';result=json.loads(path.read_text())
        result['scopeStatus']='PARTIAL_C15_BLOCKED'
        result['blockedCases']=self.suite['blockedCases']
        result['selectionHoverAndSyntheticRPC']='NOT_AUTHORIZED_IN_THIS_INCREMENT'
        path.write_text(json.dumps(result,indent=2)+'\n')

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('window');p.add_argument('pid',type=int);p.add_argument('out',type=Path);p.add_argument('cases',type=Path);p.add_argument('binary',type=Path);p.add_argument('pi_entry',type=Path)
    a=p.parse_args();suite=json.loads(a.cases.read_text())
    if any((a.out/p).exists()for p in ('actions.json','capture-result.json','evidence')):raise ValueError('Refusing prior capture evidence reuse')
    capture=None
    try:
        capture=Capture(a.window,a.pid,a.out,suite,a.binary,a.pi_entry)
        capture.settle(dict(id='STARTUP',image='evidence/startup.png',require=['Static Chat fixture']))
        capture.execute()
    except Exception as error:
        if capture is not None:capture.write_result(str(error))
        else:
            (a.out/'capture-result.json').write_text(json.dumps(dict(schema=1,cases=[],requestedCases=2,capturedCases=0,captureValidity='BLOCKED',semanticState='UNREVIEWED',pixelEquality='NOT_COMPARED',error=str(error),blockedCases=suite.get('blockedCases',[])),indent=2)+'\n')
        raise
if __name__=='__main__':main()
