#!/usr/bin/env python3
"""Validate the visual manifest without making captures or claiming visual parity."""
import hashlib
import json
import re
import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
index = json.loads((HERE / 'index.json').read_text())
rows = []
for shard in index['shards']:
    doc = json.loads((HERE / shard['json']).read_text())
    assert (HERE / shard['markdown']).is_file()
    assert len(doc['cases']) == shard['count']
    rows.extend(doc['cases'])
assert len(rows) == index['enumerated_case_count']
assert len({r['id'] for r in rows}) == len(rows)
assert len({r['slug'] for r in rows}) == len(rows)
assert sorted(r['id'] for r in rows) == [f'V{i:03d}' for i in range(1, len(rows) + 1)]
assert sum(r['scope'] == 'canonical' for r in rows) == index['canonical_in_app_cases']
assert sum(r['scope'] == 'native-os' for r in rows) == index['native_os_blocked_cases']
assert sum(r['comparison']['valid_matching_pair'] for r in rows) == index['valid_matching_pairs']
source = {}
for path, expected in index['source_hashes'].items():
    blob = subprocess.check_output(['git', 'show', index['source_baseline'] + ':' + path], cwd=ROOT)
    assert hashlib.sha256(blob).hexdigest() == expected, path
    source[path] = blob.decode()
for r in rows:
    assert r['source']
    assert r['source_baseline'] == index['source_baseline']
    for a in r['source']:
        line = source[a['path']].splitlines()[a['line'] - 1]
        assert re.search(r'\bfn\s+' + re.escape(a['handler']) + r'\b', line), (r['id'], a)
    if r['status'] == 'PASS':
        assert r['comparison']['valid_matching_pair']
        assert r['manual_review']['status'] == 'PASS'
        for key in ('rust_image', 'go_image'):
            assert r[key]['actual_path'] and r[key]['sha256'] and r[key]['capture_commit']
    else:
        assert r['status'] in {'UNTESTED', 'BLOCKED', 'PARTIAL', 'NOT_IMPLEMENTED'}
cross = json.loads((HERE / index['crosswalk']).read_text())
assert len(cross) == index['feature_scope_count'] == 139
assert len({x['feature_id'] for x in cross}) == 139
known = {r['id'] for r in rows}
for x in cross:
    assert set(x['visual_cases']) <= known
    assert x['visual_cases'] or x['backend_only_reason']
    assert x['visual_cases'] == [r['id'] for r in sorted(rows, key=lambda r: r['id']) if x['feature_id'] in r['feature_ids']]
expected_sections = {'general','appearance','shortcuts','privacy','runtime','agent','worktrees','providers','models','skills','plugins','mcp','about','report-a-bug'}
actual = {r['slug'].removeprefix('settings-').removesuffix('-baseline') for r in rows if r['slug'].startswith('settings-') and r['slug'].endswith('-baseline')}
assert actual == expected_sections, actual
print(f"VALID: {len(rows)} numbered cases, {len(cross)} scopes, {len(actual)} Settings sections; {index['valid_matching_pairs']} valid matched pairs. Inventory validation is not visual PASS.")
