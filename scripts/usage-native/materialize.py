#!/usr/bin/env python3
"""Write the bounded, credential-free Usage fixture into a NEW capture home.
No current user files, credentials, network, APIs, or model calls are used.
"""
import argparse, datetime as dt, hashlib, json, pathlib
UTC=dt.timezone.utc
BASE=dt.datetime(2026,10,1,12,tzinfo=UTC)
WORK='/tmp/orbit-visual-fixture/workspace'

def iso(t): return t.isoformat(timespec='milliseconds').replace('+00:00','Z')
def millis(t): return int(t.timestamp()*1000)
def fixture():
    files={}; totals=dict(requests=0,input=0,output=0,cache_read=0,cache_write=0,total_tokens=0,cost_usd=0.0,prompts=0,tool_runs=4,tool_errors=1,request_errors=1,aborted=1,unbilled_provider_errors=1)
    for session in range(4):
        provider=('fixture-alpha','fixture-beta')[session%2]
        model=('shared-model','shared-model','large-model','large-model')[session]
        workspace=WORK+('/alpha' if session<2 else '/beta')
        sid=f'usage-native-{session+1:02d}'
        start=BASE+dt.timedelta(days=session,hours=session)
        lines=[dict(type='session',version=3,id=sid,timestamp=iso(start-dt.timedelta(seconds=5)),cwd=workspace)]
        parent=None
        def entry(t,message):
            nonlocal parent
            eid=f'{sid}-{len(lines):03d}'
            lines.append(dict(type='message',id=eid,parentId=parent,timestamp=iso(t),message=message));parent=eid
        for request in range(8):
            t=BASE+dt.timedelta(days=(request*2+session)%14,hours=session,minutes=request)
            input_=1000+session*200+request*25; output=120+request*10
            read=400 if request%3==0 else 0; write=100 if request%3==1 else 0
            tokens=input_+output+read+write; cost=round((session+1)*.01+request*.001,3)
            entry(t,dict(role='user',content=[dict(type='text',text=f'Usage fixture task {session+1}' if request==0 else f'Fixture turn {request+1}')],timestamp=millis(t)))
            content=[dict(type='text',text='Recorded fixture response. No model call was made.')]
            if request==1: content.append(dict(type='toolCall',id=f'tool-{sid}',name='bash' if session%2==0 else 'read',arguments={}))
            usage=dict(input=input_,output=output,cacheRead=read,cacheWrite=write,totalTokens=tokens,cost={'total':cost})
            if request==0: usage['reasoning']=20
            end=t+dt.timedelta(seconds=2+request)
            entry(end,dict(role='assistant',api='fixture-messages',provider=provider,model=model,content=content,usage=usage,stopReason='aborted' if session==3 and request==7 else 'stop',timestamp=millis(t)))
            if request==1:
                entry(end+dt.timedelta(seconds=1),dict(role='toolResult',toolCallId=f'tool-{sid}',toolName='bash' if session%2==0 else 'read',isError=session==2,content=[dict(type='text',text='Fixture tool failure' if session==2 else 'Fixture tool result')],timestamp=millis(end+dt.timedelta(seconds=1))))
            totals['requests']+=1;totals['prompts']+=1
            for key,value in [('input',input_),('output',output),('cache_read',read),('cache_write',write),('total_tokens',tokens),('cost_usd',cost)]:totals[key]+=value
        if session==0:
            t=BASE+dt.timedelta(days=13,hours=4)
            entry(t,dict(role='user',content='Fixture failed attempt',timestamp=millis(t)));totals['prompts']+=1
            entry(t+dt.timedelta(seconds=1),dict(role='assistant',provider=provider,model=model,content=[],usage=dict(input=0,output=0,cost={'total':0}),stopReason='error',errorMessage='Fixture rate limit',timestamp=millis(t)));totals['requests']+=1
            entry(t+dt.timedelta(seconds=2),dict(role='assistant',provider=provider,model=model,content=[],stopReason='error',errorMessage='Fixture unbilled provider failure',timestamp=millis(t+dt.timedelta(seconds=1))))
        # Actual stores are chronological. This makes the previous-message latency
        # oracle independent of the generator's session/day distribution loop.
        lines[1:]=sorted(lines[1:],key=lambda item:item['timestamp'])
        earliest=dt.datetime.fromisoformat(lines[1]['timestamp'].replace('Z','+00:00'))
        lines[0]['timestamp']=iso(earliest-dt.timedelta(seconds=5))
        next(item for item in lines[1:] if item['message']['role']=='user')['message']['content']=[dict(type='text',text=f'Usage fixture task {session+1}')]
        parent=None
        for entry_ in lines[1:]:entry_['parentId']=parent;parent=entry_['id']
        files[f'.pi/agent/sessions/--tmp--orbit-visual-fixture--workspace--{ "alpha" if session<2 else "beta" }--/{sid}.jsonl']=''.join(json.dumps(line,separators=(',',':'))+'\n' for line in lines)
    totals['cost_usd']=round(totals['cost_usd'],6)
    prefs=dict(preset='custom',custom_start=millis(BASE.replace(hour=0)),custom_end=millis(BASE.replace(hour=0)+dt.timedelta(days=13)),metric='tokens',usage_mode='simple',breakdown_tab='models',detail_tab='sessions',session_sort='tokens',session_sort_desc=True,page_size=25,latency_metric='average',hidden_columns=[],breakdown_sort='tokens',breakdown_sort_desc=True)
    files['.orbit-pi/usage.json']=json.dumps(prefs,indent=2)+'\n'
    manifest=dict(schema=1,name='usage-native-fixed-october',credentials=False,inference=False,range_utc=['2026-10-01T00:00:00Z','2026-10-15T00:00:00Z'],sessions=4,providers=2,provider_model_pairs=4,totals=totals,files={p:hashlib.sha256(v.encode()).hexdigest() for p,v in sorted(files.items())},volatile=['scan-relative header age','activity calendar trailing-year anchor if executions occur on different UTC dates'],source_clock_frozen=False)
    return files,manifest

def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('home',type=pathlib.Path);parser.add_argument('--manifest',type=pathlib.Path,required=True);args=parser.parse_args()
    files,manifest=fixture()
    # Preflight the entire owned set before any write. Never replace prior data.
    targets=[args.home/p for p in files]+[args.manifest]
    existing=[str(p) for p in targets if p.exists()]
    if existing:raise SystemExit('Refusing to overwrite fixture targets: '+', '.join(existing))
    for relative,contents in files.items():
        path=args.home/relative;path.parent.mkdir(parents=True,exist_ok=True);path.write_text(contents)
    args.manifest.parent.mkdir(parents=True,exist_ok=True);args.manifest.write_text(json.dumps(manifest,indent=2)+'\n')
    print(json.dumps(manifest['totals'],sort_keys=True))
if __name__=='__main__':main()
