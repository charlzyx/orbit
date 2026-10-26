"""Require the frozen complete offline hitsets, not merely a zero exit status."""
import collections,json,sys
from pathlib import Path

def check(python_log,go_log,manifest):
    v=json.loads(manifest.read_text());log=python_log.read_text()
    for name in v['python_tests']:
        mod,cls,fn=name.rsplit('.',2)
        if f'{fn} ({mod}.{cls}.{fn}) ... ok' not in log:raise ValueError('missing Python pass '+name)
    if f"Ran {len(v['python_tests'])} tests" not in log or not log.endswith('OK\n'):raise ValueError('Python gate summary')
    counts=collections.defaultdict(collections.Counter)
    for line in go_log.read_text().splitlines():
        event=json.loads(line)
        if event.get('Test'):counts[event['Test']][event['Action']]+=1
        if event['Action'] in ('skip','fail'):raise ValueError('Go skipped/failed')
    if set(counts)!=set(v['go_policy_counts']):raise ValueError('Go hitset mismatch')
    if any(c['run']!=3 or c['pass']!=3 for c in counts.values()):raise ValueError('Go repeat count')
if __name__=='__main__':check(Path(sys.argv[1]),Path(sys.argv[2]),Path(__file__).with_name('validation.json'))
