#!/usr/bin/env python3
"""Resolve one exact label from a fresh whole native window; send no input."""
import argparse
import importlib.util
import json
from pathlib import Path
import re

spec = importlib.util.spec_from_file_location('native_capture', Path(__file__).with_name('capture.py'))
capture = importlib.util.module_from_spec(spec); spec.loader.exec_module(capture)


def verify(window, pid, output, label, region, required=()):
    output = Path(output)
    name = re.sub(r'[^a-z0-9]+', '-', label.casefold()).strip('-')
    path = output / f'{name}-before.png'
    proof_path = output / f'{name}-target.json'
    if any(p.exists() for p in (path, path.with_suffix('.tsv'), proof_path)):
        raise ValueError('Refusing to reuse target evidence from a previous action')
    cap = capture.Capture(window, pid, output, {'viewport': [1470, 922]})
    rows, frame = cap.capture(path)
    try:
        if required:
            acceptable, detail = capture.ready(rows, required)
            frame = dict(frame, preTargetReadiness=detail, requiredPhrases=list(required))
            if not acceptable:
                raise ValueError(f'Fresh target frame is not ready: {detail}; no input sent')
        target = capture.exact_target(rows, label, region)
    except ValueError as error:
        proof_path.write_text(json.dumps(dict(status='BLOCKED', action='NO_INPUT_SENT', frame=frame, region=region, label=label, reason=str(error)), indent=2) + '\n')
        raise
    proof_path.write_text(json.dumps(dict(status='TARGET_VERIFIED', action='NO_INPUT_SENT', frame=frame, region=region, target=target,
                                         image=path.name, semanticState='UNREVIEWED'), indent=2) + '\n')
    return target['center']


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('window'); p.add_argument('pid', type=int); p.add_argument('output', type=Path); p.add_argument('label')
    p.add_argument('region', type=int, nargs=4)
    p.add_argument('--require-ready', nargs='+', default=[])
    args = p.parse_args()
    print(*verify(args.window, args.pid, args.output, args.label, args.region, args.require_ready))


if __name__ == '__main__': main()
