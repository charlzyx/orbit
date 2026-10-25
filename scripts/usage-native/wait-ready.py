#!/usr/bin/env python3
"""Observe completed native Usage body state; sends no input and never edits pixels."""
import argparse
import importlib.util
import json
from pathlib import Path
import time

spec = importlib.util.spec_from_file_location('usage_native_capture', Path(__file__).parents[1] / 'workbench-native' / 'capture.py')
capture = importlib.util.module_from_spec(spec); spec.loader.exec_module(capture)

# These body anchors are not rendered by the loading skeleton or an empty/error
# state. Both renderers use the pinned synthetic, populated October fixture.
REQUIRED = {
    'simple': ['Summary', 'Usage over time', 'requests', 'tokens', 'sessions'],
    'details': ['Breakdown', 'Records', 'Models', 'Sessions', 'Tokens'],
    'models': ['Search', 'Select all', 'All models'],
}


def wait_ready(window, pid, output, phase, timeout=45, poll=1, consecutive=3):
    if phase not in REQUIRED or not 0 < timeout <= 120 or poll <= 0 or consecutive < 2:
        raise ValueError('Invalid bounded Usage readiness configuration')
    out = Path(output)
    evidence = out / 'usage-readiness'; evidence.mkdir(parents=True, exist_ok=True)
    target = out / f'usage-{phase}-dark.png'
    proof_path = evidence / f'{phase}-ready.json'
    if target.exists() or proof_path.exists() or list(evidence.glob(f'{phase}-*.png')):
        raise ValueError('Refusing previous Usage evidence; use a fresh capture output')
    cap = capture.Capture(window, pid, out, {'viewport': [1470, 922]})
    cap.evidence = evidence
    observations, streak = [], 0
    result = dict(phase=phase, status='WAITING', action='NO_INPUT_SENT', required=REQUIRED[phase],
                  timeoutSeconds=timeout, pollSeconds=poll, requiredConsecutiveSamples=consecutive,
                  semanticState='UNREVIEWED', pixelEquality='NOT_COMPARED', observations=observations)
    deadline = time.monotonic() + timeout
    try:
        while time.monotonic() < deadline:
            ordinal = len(observations) + 1
            path = evidence / f'{phase}-{ordinal:03d}.png'
            rows, frame = cap.capture(path)
            good, detail, frame = cap.readiness(path, rows, frame, REQUIRED[phase], ordinal)
            streak = streak + 1 if good else 0
            observations.append(dict(sample=ordinal, imagePath=str(path.relative_to(out)),
                                     ready=good, consecutive=streak, **detail, **frame))
            result.update(samples=ordinal, consecutiveReadySamples=streak)
            proof_path.write_text(json.dumps(result, indent=2) + '\n')
            if streak >= consecutive:
                if time.monotonic() > deadline:
                    raise RuntimeError(f'Usage {phase} did not become ready before the {timeout}s deadline; no subsequent target or click authorized')
                target.write_bytes(path.read_bytes())  # unchanged full native client frame
                result.update(status='OCR_READY_MANUAL_STATE_REVIEW_REQUIRED', image=target.name,
                              pngSha256=capture.sha(target))
                proof_path.write_text(json.dumps(result, indent=2) + '\n')
                return result
            time.sleep(poll)
        raise RuntimeError(f'Usage {phase} did not become ready within {timeout}s; no subsequent target or click authorized')
    except Exception as error:
        result.update(status='BLOCKED', reason=str(error))
        proof_path.write_text(json.dumps(result, indent=2) + '\n')
        raise


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('phase', choices=REQUIRED)
    parser.add_argument('window'); parser.add_argument('pid', type=int); parser.add_argument('output', type=Path)
    args = parser.parse_args()
    result = wait_ready(args.window, args.pid, args.output, args.phase)
    print(json.dumps({key: result[key] for key in ('phase', 'status', 'samples', 'consecutiveReadySamples')}))


if __name__ == '__main__': main()
