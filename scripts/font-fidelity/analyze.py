#!/usr/bin/env python3
"""Summarize observed diagnostic facts without certifying visual equivalence."""
from __future__ import annotations
import argparse, collections, hashlib, json, re
from pathlib import Path
SHA=re.compile(r'^[0-9a-f]{64}$')

def read_trace(root:Path):
    files=sorted(root.glob('*.json'))
    if not files:raise ValueError(f'No trace batches in {root}')
    batches=[];events=[];renderers=set();cases=set();problems=[];identities=[]
    for file in files:
        raw=file.read_bytes();b=json.loads(raw)
        if b.get('schema')!=1 or b.get('diagnostic') is not True:raise ValueError(f'Not a version-1 diagnostic trace: {file}')
        if not isinstance(b.get('events'),list) or not b.get('case_id') or not b.get('presentation_semantics'):raise ValueError(f'Incomplete trace batch: {file}')
        if not isinstance(b.get('completed_present_call'),int) or b['completed_present_call']<1:raise ValueError(f'Missing frame sequence: {file}')
        if not isinstance(b.get('dropped_events'),int):raise ValueError(f'Missing dropped-event count: {file}')
        if b['dropped_events']:problems.append(f'{file.name}: {b["dropped_events"]} dropped events')
        renderers.add(b.get('renderer'));cases.add(b['case_id']);batches.append(b);events.extend(b['events'])
        identities.append({'file':file.name,'sha256':hashlib.sha256(raw).hexdigest()})
    if len(renderers)!=1 or next(iter(renderers)) not in ('mygo-pango','gpui-cosmic'):raise ValueError('A trace directory must contain one known renderer/process')
    if len(cases)!=1:raise ValueError('Mixed case IDs must be analyzed separately')
    if not events:raise ValueError('Trace contains no events')
    renderer=next(iter(renderers));faces={};face_rows=[]
    for e in events:
        if e.get('kind')!='face':continue
        fid=e.get('font_id')
        if fid is None:raise ValueError('Face has no font ID')
        if fid in faces and faces[fid]!=e:raise ValueError(f'Ambiguous face ID {fid}: use one process per input directory')
        faces[fid]=e
        loaded=e.get('loaded_bytes_sha256');file=e.get('font_file_sha256')
        if loaded is not None and not SHA.fullmatch(loaded):raise ValueError('Malformed loaded font hash')
        if file is not None and not SHA.fullmatch(file):raise ValueError('Malformed font file hash')
        if loaded is None:problems.append(f'{fid}: loaded font bytes unavailable')
        if e.get('face_index') is None:problems.append(f'{fid}: face index unavailable')
        face_rows.append({'font_id':fid,'postscript_name':e.get('postscript_name',e.get('font_postscriptname')),'face_index':e.get('face_index'),'loaded_sha256':loaded,'file_sha256':file,'file_matches_loaded':None if not loaded or not file else loaded==file,'resolved_weight':e.get('resolved_weight',e.get('resolved_pango_weight')),'fontconfig_weight':e.get('fontconfig_weight'),'fontconfig_embolden':e.get('fontconfig_embolden')})
    shaped=[e for e in events if e.get('kind') in ('shape_run','shape_glyph')]
    if not shaped:problems.append('No final shaped glyphs observed')
    for e in shaped:
        if e.get('font_id') not in faces:problems.append(f'Shape references unknown face {e.get("font_id")}')
    paints=[e for e in events if e.get('kind')=='paint_line']
    texts=sorted({e['text'] for e in events if 'text' in e})
    return {'renderer':renderer,'case_id':next(iter(cases)),'trace_files':identities,'event_counts':dict(collections.Counter(e.get('kind','missing') for e in events)),'faces':face_rows,'texts':texts,'problems':sorted(set(problems)),'paint_observed':bool(paints),'presentation_boundaries':sorted({b['presentation_semantics'] for b in batches}),'events':events}

def compare(left,right):
    def identity(row):return (row['loaded_sha256'],row['face_index']) if row['loaded_sha256'] is not None and row['face_index'] is not None else None
    shared=sorted(set(left['texts'])&set(right['texts']))
    rows=[]
    for text in shared:
        selections=[]
        for side in [left,right]:
            faces={r['font_id']:r for r in side['faces']}
            ids={e.get('font_id') for e in side['events'] if e.get('text')==text and e.get('kind') in ('shape_run','shape_glyph')}
            selections.append({identity(faces[f]) for f in ids if f in faces})
        known=all(s and None not in s for s in selections)
        rows.append({'text':text,'observed_face_set_status':'same' if known and selections[0]==selections[1] else 'different' if known else 'unknown','left_face_set':sorted([list(i) for i in selections[0] if i]),'right_face_set':sorted([list(i) for i in selections[1] if i]),'attribution_limit':'Observed face sets across all requests for this string, not one-to-one run matching. Inspect request, size, weight, span and glyph records before attributing a difference.'})
    return {'shared_texts':rows,'native_pixel_parity':'not evaluated','normalization_recommendation':None,'warnings':['No text geometry, baseline, font synthesis, renderer-setting, native-frame or same-state equivalence is inferred from equal face sets.','Headless specimen results are supplementary and do not establish the fonts or GTK settings used by previously captured applications.']}

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('left',type=Path);p.add_argument('right',type=Path,nargs='?');p.add_argument('--output',type=Path,required=True);a=p.parse_args()
    left=read_trace(a.left);report={'left':left}
    if a.right:report['right']=read_trace(a.right);report['comparison']=compare(left,report['right'])
    for side in ['left','right']:
        if side in report:report[side].pop('events')
    a.output.write_text(json.dumps(report,indent=2,ensure_ascii=False)+'\n');print(a.output)
if __name__=='__main__':main()
