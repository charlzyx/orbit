#!/usr/bin/env python3
"""Build source-first acceptance shards, preserving the original 139-row inventory."""
import collections,hashlib,json,pathlib,re
ROOT=pathlib.Path(__file__).resolve().parents[3]
DOC=ROOT/'docs/mygo-migration'
OUT=DOC/'acceptance'
BASE='299a32bc87fe3d4851bdb902d9c03dcf9bf99cac'
PUBLISHED='80580ee3cffe8eab19b9399d7813fa6be61ec6d9'
index=json.loads((DOC/'feature-parity.json').read_text())
actions={}
overrides=json.loads((OUT/'row-overrides.json').read_text())
ui_cases=json.loads((DOC/'native-ui-acceptance.json').read_text())['cases']
for line in (OUT/'row-actions.txt').read_text().splitlines():
 key,action,challenge=line.split('|',2);actions[key]=(action,challenge)
evidence=json.loads((OUT/'workspace-evidence.json').read_text())
passed={e.get('Test'):e for e in evidence['results'] if e['Action']=='pass' and e.get('Test')}
ASSERTIONS={
 'W09':[('TestGitDiscardAllPreservesStagedIgnoredAndExternalTargets','Discard All preserves index, ignored files and outside symlink targets; repeating remains safe.'),('TestGitDiscardFailuresDoNotPretendSuccess','Failed/cancelled discard reports error and retains bytes, intentional deviation D02.')],
 'W02':[('TestNativeWatchDetectsPreservedMetadataAndNewNestedTrees','Native filesystem events detect metadata-preserving writes and newly nested directories.'),('TestNativeWatchFollowsLinkedMetadataAndIgnoresCheckpointChurn','Linked-worktree/common refs are watched without refreshing on private checkpoint churn.'),('TestNativeWatchDebouncesBurstAndFallsBackOnOverflow','Debounced bursts and overflow fallback preserve eventual refresh.')],
 'W34':[('TestPinnedSourceOracle','All 1,204 exact pinned-source lexical cases match across 16 languages.'),('TestReviewHighlightCarriesSeparateSidesAndResetsAtHunks','Old/new carries are independent; context uses new side; file/hunk boundaries reset state.'),('TestReviewSourceHeaderPathsMetadataAndSyntheticHunks','File identity follows diff paths, source-hidden metadata stays hidden, synthetic hunks have no invented positions.'),('TestReviewHeaderEscapesAndCompactAsymmetricGaps','Quoted/space paths decode and compact gaps account for both old and new ranges.'),('TestDiffColorSpansPreserveEveryByteAndFontMetrics','Git paint spans concatenate exact text and change color only, including dynamic palette callback.')],
 'W35':[('TestReviewHighlightSurvivesHiddenExpansionAndGrouping','Token spans survive retained context expansion, grouping and split conversion.'),('TestReviewVisibleLimitAndExpansionCap','Visible-line cap remains 50,000 and expansion cannot exceed it.')],
 'W41':[('TestNativeCreateNameTypingDerivesBranchUntilManualEdit','Typing the worktree name derives branch only until the user edits branch explicitly.')],
 'W42':[('TestNativeBranchInUseOpenRequiresExplicitClick','Branch-in-use handling does not open another worktree until explicit Open Existing click.')],
 'W43':[('TestSettingsPersistenceFailureIsTransactionalAndRetryable','Failed worktree settings persistence leaves prior config intact and supports retry.')],
 'W44':[('TestSetupAllowRemembersCanonicalRepositoryBeforeCreation','Explicit consent persists canonical repository trust before setup-enabled creation.')],
 'W45':[('TestSetupConsentFailureAndRetry','Setup consent, failure and retry do not duplicate worktree creation.')],
 'W49':[('TestFailedLockRetryDoesNotRepeatPreviousCreate','Retry of a failed lock cannot replay a previous Create operation.')],
 'W50':[('TestWorkInDismissalRestoresComposerFocus','Actual Escape returns focus to composer across three immediate reopen/dismiss cycles.'),('TestWorkInKeyboardWrapNoopAndDialogFocusGuard','Arrow/Enter wrap, current-worktree no-op and New/existing-dialog focus guard match source.'),('TestWorkInRowsAndNativeMenuSelection','Native Work In menu uses actual main/linked/New rows and selections.')],
 'W51':[('TestThrottledWorktreeWatchEventuallyRefreshesWithoutAnotherEvent','A throttled/busy watch event yields an eventual refresh without a second event.')],
 'W57':[('TestEditorFailedCloseKeepsDraftAndCanRetry','Failed close preserves dirty draft; later successful save closes it, intentional deviation D01.'),('TestEditorInactiveLoadNeverStealsSelection','Background file load registers a document without changing active selection.')],
 'W59':[('TestEditorSelfSaveReloadPreservesUndoHistory','Self-save watcher echoes preserve undo/redo history.'),('TestSlowSaveDoesNotBlockSnapshotsOrLoseConcurrentEdits','A slow save does not block snapshots or erase newer edits.'),('TestEditorDirtyReloadAndFailedSaveKeepBuffer','External dirty conflict and failed save retain user buffer.')],
 'W60':[('TestFileLoadingModesAndGuards','Text/image/binary/loading guards use actual file content and bounds.'),('TestEditorReadonly','Read-only documents reject edits and saves.')],
}
checkpoint_proof=json.loads((OUT/'checkpoint-evidence.json').read_text())
assert checkpoint_proof['exit_code']==0 and checkpoint_proof['revision']['files_unchanged_during_test']
checkpoint_passed={e['Test']:e for e in checkpoint_proof['results'] if e['Action']=='pass'}
checkpoint_assertions=json.loads((OUT/'checkpoint-assertions.json').read_text())
EXTRA={
 'W31':[('crates/orbit-pi/src/sidepane.rs','set_turn_context'),('crates/orbit-pi/src/sidepane.rs','source_menu'),('crates/orbit-pi/src/app.rs','sync_review_context'),('crates/orbit-pi/src/git.rs','resolve_range')],
 'W32':[('crates/orbit-pi/src/app/session.rs','begin_turn'),('crates/orbit-pi/src/app/session.rs','finish_turn'),('crates/orbit-pi/src/checkpoint.rs','capture_turn_start'),('crates/orbit-pi/src/checkpoint.rs','latest_turn')],
 'U01':[('crates/orbit-pi/src/usage/collect.rs','scan_store'),('crates/orbit-pi/src/usage/collect.rs','build_index'),('crates/orbit-pi/src/usage/aggregate.rs','build_calendar')],
 'U08':[('crates/orbit-pi/src/updater.rs','check_for_updates'),('crates/orbit-pi/src/updater.rs','verify_artifact'),('crates/orbit-pi/src/updater.rs','download_to')],
 'U13':[('crates/orbit-pi/src/http.rs','avatar_client'),('crates/orbit-pi/src/bug_report.rs','issue_body')],
 'U14':[('crates/orbit-pi/src/main.rs','main'),('crates/orbit-pi/src/assets.rs','collect_fonts')],
}
all_sources=list((ROOT/'crates').rglob('*.rs'))
def resolve_path(raw):
 for candidate in [ROOT/raw,ROOT/'crates/orbit-pi/src'/raw]:
  if candidate.is_file():return str(candidate.relative_to(ROOT))
 found=[p for p in all_sources if str(p).endswith('/'+raw)]
 return str(found[0].relative_to(ROOT)) if len(found)==1 else None

def anchors(row):
 result=[]
 for loc in row.get('source_locations',[]):
  result.append((loc['path'],loc.get('symbol'),loc.get('line')))
 for source in row['sources']:
  for m in re.finditer(r'([\w./-]+\.rs)(?:::(\{[^}]+\}|[\w:]+))?',source):
   path=resolve_path(m[1]);symbol=m[2]
   if not path:continue
   names=symbol.strip('{}').split(',') if symbol else [None]
   for name in names:result.append((path,name.strip() if name else None,None))
 for path,symbol in EXTRA.get(row['id'],[]):result.insert(0,(path,symbol,None))
 seen=set();out=[]
 for path,symbol,line in result:
  if (path,symbol) in seen or not (ROOT/path).is_file():continue
  seen.add((path,symbol));text=(ROOT/path).read_text(errors='replace')
  if symbol and not line:
   short=symbol.split('::')[-1];m=re.search(r'\b(?:fn|struct|enum|const|static|type|trait)\s+'+re.escape(short)+r'\b',text)
   if m:line=text[:m.start()].count('\n')+1
  if not symbol:continue
  out.append({'path':path,'handler':symbol,'line':line,'url':f'https://github.com/imrj05/orbit/blob/{BASE}/{path}'+(f'#L{line}' if line else ''),'resolution':'definition_located' if line else 'source_reference_needs_manual_disambiguation'})
 # Keep representative handler anchors compact; full inventory remains linked.
 located=[x for x in out if x['line']]
 if not located:raise RuntimeError('No located Rust handler for '+row['id'])
 return located[:4]

FIXTURES={
 'shell':'Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.',
 'chat':'Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.',
 'settings':'Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.',
 'updates-privacy':'Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.',
 'workbench-lifecycle':'Two temporary Git workspaces, fixture Pi stream, open files/tabs and independent watcher/subprocess completion controls.',
 'git-changes-history':'Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.',
 'github':'Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.',
 'review':'Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.',
 'worktrees':'Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.',
 'files':'Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.',
 'terminal-platform':'Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.',
}
FAILURE={
 'shell':'Delay/fail session scan or runtime child used by this action; repeat while changing workspace.',
 'chat':'Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.',
 'settings':'Make this setting\'s file write or provider/package/MCP operation fail; cancel then retry once after recovery.',
 'updates-privacy':'Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.',
 'workbench-lifecycle':'Delay/fail a watcher or refresh callback and switch visible workspace/surface before completion.',
 'git-changes-history':'Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.',
 'github':'Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.',
 'review':'Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.',
 'worktrees':'Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.',
 'files':'Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.',
 'terminal-platform':'Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.',
}

def step(id,kind,action,expected):return {'id':id,'dimension':kind,'action':action,'expected':expected,'status':'UNTESTED','specification':'ROW_SPECIFIC' if kind in ['functional','edge'] else 'DRAFT_NEEDS_SOURCE_EXPANSION','last_verified_sha':None}

def make(row,cat):
 id=row['id'];action,challenge=actions[id];assertions=[]
 for name,expected in ASSERTIONS.get(id,[]):
  e=passed.get(name)
  if not e:raise RuntimeError('Missing executed assertion '+name)
  assertions.append({'id':id+'.assert.'+name,'status':'PASS','expected':expected,'test':name,'package':e['Package'],'command':'go test -race -count=1 '+e['Package'].replace('github.com/charlzyx/orbit','.').replace('..','.')+' -run ^'+name+'$','evidence':'workspace-evidence.json','last_verified_sha':evidence['revision']['sha256'],'sha_kind':'working-tree-content-manifest; not a Git commit','verified_at_utc':e['Time'],'limits':'Only the named assertion. Does not pass the parent acceptance case or visual/native-OS gates.'})
 for name,expected in checkpoint_assertions.get(id,[]):
  e=checkpoint_passed[name]
  assertions.append({'id':id+'.assert.'+name,'status':'PASS','expected':expected,'test':name,'package':e['Package'],'evidence':'checkpoint-evidence.json','last_verified_sha':checkpoint_proof['revision']['sha256'],'sha_kind':'working-tree-content-manifest; not a Git commit','verified_at_utc':e['Time'],'limits':'Named assertion only; no full-scope/visual/native-OS PASS.'})
 functional='PARTIAL' if assertions else 'UNTESTED'
 steps=[
  step(id+'.behavior','functional',action,row['interaction_contract']),
  step(id+'.edge','edge',challenge,row['state_failure_platform']),
  step(id+'.error','error',FAILURE[cat], 'Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: '+row['state_failure_platform']),
  step(id+'.loading','loading','Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.', 'Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.'),
  step(id+'.persistence','persistence','Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.', 'Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.'),
  step(id+'.keyboard','keyboard','Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.', 'Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.'),
  step(id+'.repeat','repeat','Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.', 'No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.'),
 ]
 result={'id':id,'title':row['feature'],'baseline_commit':BASE,'source_handlers':anchors(row),'baseline_behavior':row['interaction_contract'],'baseline_edge_state':row['state_failure_platform'],'fixture':FIXTURES[cat],'steps':steps,'go_mapping':row['go_evidence']['paths'],'existing_test_candidates':row['go_evidence']['tests'],'inventory_evidence':f'../parity/{cat}.json','source_test_inventory':row['tests'],'verdicts':{'functional':{'status':functional,'reason':'Named narrow assertions pass; remaining row steps are unverified.' if assertions else 'Source-derived acceptance procedure is ready; this full row has not been executed end-to-end.','last_verified_sha':None,'last_partial_evidence_sha':assertions[-1]['last_verified_sha'] if assertions else None},'visual':{'status':'UNTESTED','reason':'No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state.','last_verified_sha':None},'native_os':{'status':'UNTESTED','reason':'This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.','last_verified_sha':None}},'assertion_results':assertions,'known_remaining':row['go_evidence']['remaining'],'full_parity_verified':False}

 for item in result['steps']:
  key=item['id'].split('.')[-1]
  definition=overrides.get(id,{}).get(key)
  if id=='W03':definition=overrides.get('S02',{}).get(key)
  if definition:item['action'],item['expected']=definition;item['specification']='ROW_SPECIFIC'
 result['supplemental_cases']=[]
 for case in ui_cases:
  if id not in case['rows']:continue
  observed=case['functional_verdict']
  status='NOT_IMPLEMENTED' if id=='C04' and observed=='failed_missing_implementation' else 'PARTIAL'
  result['supplemental_cases'].append({'id':case['id'],'status':status,'owner_reported_verdict':observed,'steps':case['steps'],'expected':case['expected'],'tests':case['tests'],'remaining':case['remaining'],'evidence':'../native-ui-acceptance.json','last_verified_sha':None,'limits':'Owner-scoped report; no full-row or immutable-revision PASS assigned by this index.'})
  if observed!='integration_in_progress':result['verdicts']['functional']['status']=status;result['verdicts']['functional']['reason']='Owner source challenge identifies scoped tests and/or concrete omissions; see supplemental cases. No full acceptance PASS.'
 result['additional_evidence']=[]
 if id in ['W09','W13','W57','W59','W43','W44','P02','P04','S08']:result['additional_evidence'].append('../intentional-deviations.md')
 if id=='W39':
  proof=json.loads((OUT/'ui-feedback-evidence.json').read_text());result['assertion_results'].append({'id':'W39.assert.TestReviewRefreshKeepsSource650msFeedback','status':proof['status'],'expected':'Native refresh spinner remains active650ms independently of completed content load; repeated clicks extend deadline.','test':'TestReviewRefreshKeepsSource650msFeedback','evidence':'ui-feedback-evidence.json','last_verified_sha':proof['revision']['sha256'],'sha_kind':'dirty-tree source manifest','limits':'Named native CPU feedback regression only.'});result['verdicts']['functional']['status']='PARTIAL';result['additional_evidence']+=['strict-results.json','ui-feedback-evidence.json']
 if id in ['W31','W32','W33']:result['additional_evidence']+=['../checkpoint-acceptance.md','checkpoint-evidence.json','checkpoint-failure-history.json','../intentional-deviations.md']
 if id=='W50':result['additional_evidence'].append('work-in-focus-history.json')
 if id in ['P10','P11','P12','P13']:result['additional_evidence'].append('../resource-ui-implementation.md')
 if id in ['P05','P06','P10','P11','P12','P13','U01','U02','U03','U04']:result['additional_evidence'] += ['../settings-backend-implementation.md','../settings-backend-source-evidence.json']
 if id in ['P14','P15','P16','P17']:result['additional_evidence'].append('../mcp-ui-implementation.md')
 if id in ['P02','P04']:result['additional_evidence'].append('../runtime-acceptance.md')
 if id in ['S07','P18','P19']:result['additional_evidence'].append('../background-acceptance.md')
 return result

summary=[];allrows=[]
for category in index['categories']:
 cat=category['name'];rows=json.loads((DOC/category['json']).read_text())['rows'];items=[make(row,cat) for row in rows];allrows+=items
 # Stable small category groups, maximum four scope rows per shard.
 shards=[]
 for start in range(0,len(items),4):
  group=items[start:start+4];name=f'{cat}-{group[0]["id"]}-{group[-1]["id"]}'
  data={'schema':'orbit-acceptance-shard/v1','category':cat,'baseline_commit':BASE,'rows':group}
  (OUT/(name+'.json')).write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
  md=[f'# {category["title"]}: {group[0]["id"]}–{group[-1]["id"]}', '',f'Pinned baseline: `{BASE}`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.','']
  for r in group:
   v=r['verdicts'];md += [f'## {r["id"]} — {r["title"]}','',f'Functional **{v["functional"]["status"]}** · Visual **{v["visual"]["status"]}** · Native OS **{v["native_os"]["status"]}**','', '**Rust source:** '+ '; '.join(f'[{x["path"]}::{x["handler"]}]({x["url"]})' for x in r['source_handlers']), '', '**Fixture:** '+r['fixture'],'']
   for i,s in enumerate(r['steps'],1):md += [f'{i}. **{s["id"]} [{s["status"]}; {s["specification"]}]** {s["action"]}',f'   - Expected: {s["expected"]}']
   md += ['', '**Go mapping:** '+', '.join('`'+p+'`' for p in r['go_mapping']), '**Existing test candidates, not automatic acceptance:** '+', '.join('`'+p+'`' for p in r['existing_test_candidates']), '', '**Executed narrow assertions:**']
   if r['assertion_results']:
    for a in r['assertion_results']:md += [f'- **PASS** `{a["test"]}`: {a["expected"]} [Evidence]({a["evidence"]}). Tested dirty-tree manifest `{a["last_verified_sha"]}`, not a published commit.']
   else:md += ['- None assigned to this scope yet. Last verified SHA: none.']
   for supplement in r['supplemental_cases']:md += ['',f'**Owner challenge {supplement["id"]} [{supplement["status"]}]:** '+supplement['remaining']+' [Scoped steps/tests](../native-ui-acceptance.json).']
   if r['additional_evidence']:md += ['', '**Additional source-specific procedures:** '+', '.join(f'[{path}]({path})' for path in r['additional_evidence'])]
   md += ['', '**Remaining:** '+r['known_remaining'], '**Visual/OS gates:** '+v['visual']['reason']+' '+v['native_os']['reason'],'']
  (OUT/(name+'.md')).write_text('\n'.join(md)+'\n');shards.append({'ids':[r['id'] for r in group],'json':name+'.json','markdown':name+'.md'})
 counts={dimension:dict(collections.Counter(r['verdicts'][dimension]['status'] for r in items)) for dimension in ['functional','visual','native_os']}
 summary.append({'category':cat,'title':category['title'],'count':len(items),'verdict_totals':counts,'shards':shards})
assert len(allrows)==139 and len({r['id'] for r in allrows})==139
check={'schema':'orbit-acceptance-index/v1','baseline_commit':BASE,'published_implementation_reference':{'sha':PUBLISHED,'url':'https://github.com/charlzyx/orbit/commit/'+PUBLISHED,'scope':'Reference only. Current dirty-tree evidence is NOT attributed to this commit.'},'row_count':139,'step_count':sum(len(r['steps']) for r in allrows),'expanded_row_specific_steps':sum(s['specification']=='ROW_SPECIFIC' for r in allrows for s in r['steps']),'unexpanded_draft_gates':sum(s['specification']=='DRAFT_NEEDS_SOURCE_EXPANSION' for r in allrows for s in r['steps']),'narrow_assertions_passed':sum(a['status']=='PASS' for r in allrows for a in r['assertion_results']),'full_feature_pass_count':0,'status_enum':['NOT_IMPLEMENTED','PARTIAL','UNTESTED','PASS','BLOCKED'],'categories':summary,'manifest_evidence':'workspace-evidence.json','checkpoint_manifest_evidence':'checkpoint-evidence.json','rules':'PASS requires the named complete acceptance scope, exact tested revision and evidence. Visual and native OS verdicts are independent. Pending/blocked source captures never inherit domain-test PASS.'}
(OUT/'index.json').write_text(json.dumps(check,ensure_ascii=False,indent=2)+'\n')
md=['# Orbit source-first acceptance checklist: initial executable inventory','',f'Baseline: [`{BASE}`](https://github.com/imrj05/orbit/tree/{BASE}). Native Go + MyGo target; cloud-only fixtures.','',f'**139 feature scopes · {check["expanded_row_specific_steps"]} row-specific procedures · {check["unexpanded_draft_gates"]} explicitly unexpanded draft gates · {check["narrow_assertions_passed"]} narrow assertions passed · 0 full feature passes.**','','This is an initial checklist to expand and execute, not a completion report. Gates marked DRAFT_NEEDS_SOURCE_EXPANSION still need exact source-specific faults, files, timing and chords; they are not finished behavioral specifications. Every scope links pinned Rust handlers, row-specific behavior/edge actions, Go mapping and evidence. Functional, visual and native-OS verdicts are separate.','', '**Status meanings:** NOT_IMPLEMENTED = no working path; PARTIAL = named subset verified or known gap; UNTESTED = acceptance scope not run; PASS = exact scope executed on identified revision with evidence; BLOCKED = explicit unavailable prerequisite. A failed assertion is recorded as an observed failure under PARTIAL, not hidden in a package-level success.','', '**Safety:** use isolated HOME/session stores, temporary repositories, local bare remotes and fake gh/provider/auth/update endpoints. No user repositories, real credentials, external mutations or user Mac. Native OS tests require an appropriate separately authorized cloud host.','', '| Category | Rows | Functional | Visual | Native OS | Procedures |','|---|---:|---|---|---|---|']
for cat in summary:
 fmt=lambda d:', '.join(f'{k} {v}' for k,v in d.items())
 links=' · '.join(f'[{s["ids"][0]}–{s["ids"][-1]}](acceptance/{s["markdown"]})' for s in cat['shards'])
 md.append(f'| {cat["title"]} | {cat["count"]} | {fmt(cat["verdict_totals"]["functional"])} | {fmt(cat["verdict_totals"]["visual"])} | {fmt(cat["verdict_totals"]["native_os"])} | {links} |')
md+=['','## Revision and evidence rules','',f'Published implementation reference: [`{PUBLISHED}`](https://github.com/charlzyx/orbit/commit/{PUBLISHED}). Current acceptance evidence is a separately hashed dirty-tree snapshot, **not** evidence for that published SHA. [Exact files, command, timestamps and named test outcomes](acceptance/workspace-evidence.json). The manifest was unchanged before/after the executed suite.','', 'Per-row `last_verified_sha` is null until that scope is actually tested. Narrow assertions identify the manifest SHA and test name separately. Existing source/Go tests are candidates, not automatic passes.','', 'Visual procedure for each row: capture pinned Rust and Go at identical OS, scale, installed fonts, locale, viewport and fixture state; compare geometry, ordering, text, focus, clipping and interaction states. Native-OS procedure: repeat the row on every advertised host, including actual PTY/window/menu/clipboard/dialog lifetimes. Cross-compilation and headless CPU rendering are not substitutes.','', '## First independently challenged cases','', '- W31/W32: Last Turn incorrectly targeted a new unfinished prompt and lost the prior diff. Separate completed state, pinned historical identity and safe recovery now pass20 narrow assertions; five original Rust checkpoint tests also pass. [Failure, correction and remaining cases](checkpoint-acceptance.md).', '- W34: fixed and regression-tested reversed numstat/file identity, suppressed raw diff metadata, synthetic unpositioned hunks, escaped paths and asymmetric compact gaps. 1,204 lexical oracle cases plus separate old/new carry and paint-only spans pass; W34 as a whole remains PARTIAL.','- W39: fast native Review refresh failed the650ms feedback floor (27.372ms locally). The ordinary rendered-spinner/repeated-click regression now passes. [Failure and correction evidence](acceptance/strict-results.json).', '- W50: native-menu dismissal failed composer focus restoration. Native popup key handling now passes repeated Escape, wraparound selection and dialog-focus regressions. [History and scoped correction](acceptance/work-in-focus-history.json).', '- W59: self-save watcher reload had reset undo history. Regression now proves undo history survives the save echo; slow-save and dirty-conflict assertions pass. Full editor interaction/OS parity remains open.','- W51: a throttled worktree watch event could be dropped without another event. The pending refresh regression passes. Full Worktrees page parity remains open.','- The independent Usage settings round-trip run initially failed; its owner reports a correction and new scoped evidence in [settings backend verification](settings-backend-implementation.md). This index does not silently convert that whole category to PASS.','', '## Additional owner evidence, separately scoped','', '- [Settings/backend source checklist](settings-backend-implementation.md) and [exact source hashes](settings-backend-source-evidence.json).','- [Native UI source challenges and corrected regressions](native-ui-acceptance.md).', '- [Skills/Plugins:R01–R21 source cases](resource-ui-implementation.md).', '- [Explicit intentional deviations, never counted as unqualified parity](intentional-deviations.md).', '- [Runtime/Agent exact lifecycle acceptance](runtime-acceptance.md).', '- [MCP settings:19 independently scoped cases](mcp-ui-implementation.md).', '- [Backdrop source acceptance, host integration pending](background-acceptance.md).', '- [Original full source inventory and all remaining caveats](feature-parity.md).','- [Paint-only lexical oracle and provenance caveat](highlight-parity.md).','- [Machine-readable checklist index](acceptance/index.json).','']
(DOC/'acceptance-checklist.md').write_text('\n'.join(md))
print('rows',len(allrows),'steps',check['step_count'],'scoped assertions',check['narrow_assertions_passed'])
for cat in summary:print(cat['category'],cat['count'],cat['verdict_totals'])
