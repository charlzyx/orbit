#!/usr/bin/env python3
"""Validate the modular preservation contract and assemble optional full downloads."""
from pathlib import Path
import argparse
import json

ROOT = Path(__file__).resolve().parent

def assemble():
    index = json.loads((ROOT / 'feature-parity.json').read_text())
    assert index['schema'] == 'orbit-parity-index/v1'
    structural = {'schema', 'row_count', 'full_parity_verified', 'categories',
                  'source_inventory_files', 'source_inventory_order', 'evidence_file',
                  'legacy_workbench_source_files', 'markdown_sections'}
    result = {k: v for k, v in index.items() if k not in structural}
    rows = []
    for category in index['categories']:
        records = json.loads((ROOT / category['json']).read_text())['rows']
        assert [r['id'] for r in records] == category['row_ids'], category['name']
        rows.extend(records)
    ids = [r['id'] for r in rows]
    assert len(ids) == len(set(ids)) == index['row_count'], 'missing/duplicate row'
    assert all(isinstance(r['full_parity_verified'], bool) for r in rows)
    result['rows'] = rows
    inventory = {}
    for file in index['source_inventory_files']:
        for record in json.loads((ROOT / file).read_text()):
            assert record['path'] not in inventory, record['path']
            inventory[record['path']] = record
    assert set(inventory) == set(index['source_inventory_order'])
    result['source_inventory'] = [inventory[p] for p in index['source_inventory_order']]
    result['workbench_matrix_markdown'] = ''.join((ROOT / path).read_text() for path in index['legacy_workbench_source_files'])
    result.update(json.loads((ROOT / index['evidence_file']).read_text()))
    markdown = (ROOT / 'feature-parity.md').read_text()
    for section in index['markdown_sections']:
        markdown += '\n\n' + (ROOT / section).read_text()
    return result, markdown

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true', help='validate without generating files')
    parser.add_argument('--output', type=Path, default=ROOT / 'generated')
    args = parser.parse_args()
    result, markdown = assemble()
    if not args.check:
        args.output.mkdir(parents=True, exist_ok=True)
        (args.output / 'feature-parity.full.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
        (args.output / 'feature-parity.full.md').write_text(markdown)
    verified = sum(r['full_parity_verified'] for r in result['rows'])
    print(f"Validated {len(result['rows'])} unique rows, {len(result['source_inventory'])} source records; {verified} full-parity gates verified.")

if __name__ == '__main__':
    main()
