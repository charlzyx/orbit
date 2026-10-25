#!/usr/bin/env python3
"""Drive the real X11 app using public input only; never construct UI state."""
import argparse
import csv
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import time


def run(*args):
    return subprocess.check_output(list(args), text=True, stderr=subprocess.PIPE, timeout=15).strip()


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def words(tsv):
    return [row for row in csv.DictReader(io.StringIO(tsv), delimiter='\t') if row.get('level') == '5' and (row.get('text') or '').strip()]


def plain(rows):
    return re.sub(r'\s+', ' ', ' '.join(row['text'] for row in rows)).casefold()


def exact_target(rows, label, region):
    x1, y1, x2, y2 = region
    matches = []
    for row in rows:
        x, y, w, h = (int(row[k]) for k in ('left', 'top', 'width', 'height'))
        if row['text'] == label and x1 <= x and y1 <= y and x + w <= x2 and y + h <= y2:
            matches.append(dict(label=label, bbox=[x, y, w, h], confidence=float(row['conf']), center=[x + w // 2, y + h // 2]))
    if len(matches) != 1 or not 60 <= matches[0]['confidence'] <= 100:
        raise ValueError(f'Exact OCR target {label!r}: expected one unambiguous match with confidence >=60 in {region}, got {len(matches)}; no click sent')
    return matches[0]


def ready(rows, required):
    text = plain(rows)
    absent = [phrase for phrase in required if phrase.casefold() not in text]
    blocked = [phrase for phrase in ('loading...', 'loading…', 'failed to load', 'could not load', 'runtime unavailable', 'runtime not found', 'workspace unavailable', 'loading local usage', 'unable to load usage data', 'unable to read usage data', 'no usage data yet', 'no usage matches these filters', 'refreshing', 'permission denied', 'no such file or directory', 'not a directory', 'operation not permitted', 'input/output error', 'failed to scan') if phrase in text]
    return not absent and not blocked, dict(missing=absent, blocked=blocked)


class Capture:
    def __init__(self, window, pid, out, suite):
        self.window, self.pid, self.out, self.suite = str(window), pid, Path(out), suite
        self.evidence = self.out / 'evidence'
        self.evidence.mkdir(parents=True, exist_ok=True)
        self.actions = []
        self.cases = []

    def alive(self):
        os.kill(self.pid, 0)

    def capture(self, path):
        self.alive()
        geometry = dict(line.split('=', 1) for line in run('xdotool', 'getwindowgeometry', '--shell', self.window).splitlines() if '=' in line)
        actual = [int(geometry['WIDTH']), int(geometry['HEIGHT'])]
        if actual != self.suite['viewport']:
            raise ValueError(f'Actual X11 client viewport {actual}, expected {self.suite["viewport"]}; no resizing/cropping permitted')
        run('import', '-window', self.window, str(path))
        image_size = [int(v) for v in run('identify', '-format', '%w %h', str(path)).split()]
        if image_size != actual:
            raise ValueError(f'Image/client geometry disagreement: {image_size} vs {actual}')
        tsv = run('tesseract', str(path), 'stdout', '--psm', '11', 'tsv')
        path.with_suffix('.tsv').write_text(tsv + '\n')
        return words(tsv), dict(client=actual, image=image_size, window=self.window,
                               title=run('xdotool', 'getwindowname', self.window), pngSha256=sha(path),
                               ocrSha256=sha(path.with_suffix('.tsv')))

    def readiness(self, image, rows, metadata, required, sample):
        acceptable, detail = ready(rows, required)
        primary = dict(psm=11, **detail, ready=acceptable,
                       pngSha256=metadata.get('pngSha256'), tsvSha256=metadata.get('ocrSha256'))
        passes = [primary]
        selected = 11 if acceptable else None
        # This is a readiness-only fallback. Never merge partial recognitions,
        # alter required phrases, or override a primary loading/error marker.
        # Pointer targets continue to use the original exact-label OCR path.
        if not acceptable and detail['missing'] and not detail['blocked']:
            evidence = self.evidence / f'{image.stem}-fallback-{sample:03d}.png'
            evidence.write_bytes(image.read_bytes())
            evidence.with_suffix('.tsv').write_bytes(image.with_suffix('.tsv').read_bytes())
            before_hash = sha(evidence)
            if before_hash != metadata['pngSha256']:
                raise ValueError('Readiness fallback image differs from the captured whole native window')
            alternate = run('tesseract', str(evidence), 'stdout', '--psm', '3', 'tsv')
            alternate_path = evidence.with_suffix('.psm3.tsv')
            alternate_path.write_text(alternate + '\n')
            if sha(evidence) != before_hash:
                raise ValueError('Readiness OCR modified its input image')
            acceptable, detail = ready(words(alternate), required)
            passes[0].update(image=str(evidence.relative_to(self.out)), tsv=str(evidence.with_suffix('.tsv').relative_to(self.out)))
            passes.append(dict(psm=3, **detail, ready=acceptable, pngSha256=before_hash,
                               image=str(evidence.relative_to(self.out)), tsv=str(alternate_path.relative_to(self.out)),
                               tsvSha256=sha(alternate_path)))
            selected = 3 if acceptable else None
        return acceptable, detail, dict(metadata, readinessOcrPasses=passes, selectedReadinessOcrPSM=selected)

    def settle(self, case):
        config = self.suite['settle']
        deadline = time.monotonic() + config['timeoutSeconds']
        consecutive, samples = 0, []
        image = self.evidence / f'{case["id"]}-settle.png'
        while time.monotonic() < deadline:
            rows, metadata = self.capture(image)
            acceptable, detail, metadata = self.readiness(image, rows, metadata, case['require'], len(samples) + 1)
            consecutive = consecutive + 1 if acceptable else 0
            samples.append(dict(sample=len(samples) + 1, ready=acceptable, consecutive=consecutive, **detail, **metadata))
            (self.evidence / f'{case["id"]}-settle.json').write_text(json.dumps(samples, indent=2) + '\n')
            if consecutive >= config['consecutiveSamples']:
                # Copy the actual full client screenshot; no image transformation.
                target = self.out / case['image']
                target.write_bytes(image.read_bytes())
                return dict(**metadata, samples=len(samples), consecutiveReadySamples=consecutive,
                            expectedPhrases=case['require'], readiness='OCR_READY_MANUAL_STATE_REVIEW_REQUIRED')
            time.sleep(config['pollSeconds'])
        raise RuntimeError(f'{case["id"]} did not settle: {samples[-1] if samples else "no frame"}')

    def target(self, image, rows, metadata, label, region):
        try:
            return exact_target(rows, label, region), dict(metadata, targetOcrPSM=11)
        except ValueError:
            # Only recover an already uniquely located exact literal whose
            # confidence is low. Missing/ambiguous labels are never guessed.
            x1, y1, x2, y2 = region
            candidates = []
            for row in rows:
                if row['text'] != label:
                    continue
                x, y, w, h = (int(row[k]) for k in ('left', 'top', 'width', 'height'))
                if x1 <= x and y1 <= y and x+w <= x2 and y+h <= y2:
                    candidates.append((row, [x, y, w, h]))
            if len(candidates) != 1 or not 0 <= float(candidates[0][0]['conf']) < 60:
                raise
        original, bbox = candidates[0]
        if sha(image) != metadata['pngSha256']:
            raise ValueError('Target fallback image differs from captured whole native window; no click sent')
        alternate = run('tesseract', str(image), 'stdout', '--psm', '3', 'tsv')
        alternate_path = image.with_suffix('.psm3.tsv')
        alternate_path.write_text(alternate + '\n')
        if sha(image) != metadata['pngSha256']:
            raise ValueError('Target OCR modified input image; no click sent')
        target = exact_target(words(alternate), label, region)
        if target['bbox'] != bbox:
            raise ValueError('Second-pass target OCR bounds disagree; no click sent')
        # No threshold reduction, partial-word match, crop, pixel editing or
        # coordinate fallback. Both raw passes and the unchanged frame persist.
        proof = dict(metadata, targetOcrPSM=3, targetOcrPasses=[
            dict(psm=11, label=label, bbox=bbox, confidence=float(original['conf']),
                 tsv=str(image.with_suffix('.tsv').relative_to(self.out)), tsvSha256=metadata['ocrSha256']),
            dict(psm=3, **target, tsv=str(alternate_path.relative_to(self.out)), tsvSha256=sha(alternate_path))])
        return target, proof

    def action(self, case_id, ordinal, action):
        run('xdotool', 'windowactivate', '--sync', self.window)
        event = dict(case=case_id, ordinal=ordinal, requested=action)
        if 'key' in action:
            run('xdotool', 'key', '--clearmodifiers', action['key'])
        else:
            path = self.evidence / f'{case_id}-action-{ordinal}-before.png'
            rows, metadata = self.capture(path)
            target, metadata = self.target(path, rows, metadata, action['clickExactLabel'], action['region'])
            event.update(target=target, beforeCapture=str(path.relative_to(self.out)), before=metadata)
            # Log verified target before sending any pointer input.
            self.actions.append(dict(event, outcome='TARGET_VERIFIED_CLICK_PENDING'))
            self.write_actions()
            run('xdotool', 'mousemove', '--window', self.window, *(str(v) for v in target['center']), 'click', '1')
        event['outcome'] = 'INPUT_SENT_STATE_NOT_YET_VERIFIED'
        self.actions.append(event)
        self.write_actions()
        # Event-order buffer, not the settled-state gate.
        time.sleep(.25)
        if action.get('afterRequire'):
            # Intermediate navigation/focus prerequisites are evidence only,
            # not extra coverage cases. Do not dispatch the next input early.
            run('xdotool', 'mousemove', '--window', self.window, '4', '4')
            ready_id = f'{case_id}-action-{ordinal}-ready'
            event['afterReadiness'] = self.settle(dict(id=ready_id, image=f'evidence/{ready_id}.png', require=action['afterRequire']))
            self.write_actions()

    def write_actions(self):
        (self.out / 'actions.json').write_text(json.dumps(self.actions, indent=2) + '\n')

    def execute(self):
        try:
            for case in self.suite['cases']:
                for index, action in enumerate(case['actions'], 1):
                    self.action(case['id'], index, action)
                run('xdotool', 'mousemove', '--window', self.window, '4', '4')
                proof = self.settle(case)
                self.cases.append(dict(id=case['id'], image=case['image'], expectedState=case['state'],
                                       captureValidity='READY_FOR_MANUAL_REVIEW', semanticState='UNREVIEWED',
                                       pixelEquality='NOT_COMPARED', proof=proof))
                self.write_result()
        except Exception as exc:
            self.write_result(str(exc))
            raise

    def write_result(self, error=None):
        value = dict(schema=1, cases=self.cases, requestedCases=len(self.suite['cases']), capturedCases=len(self.cases),
                     captureValidity='BLOCKED' if error else 'READY_FOR_MANUAL_REVIEW',
                     semanticState='UNREVIEWED', pixelEquality='NOT_COMPARED', error=error)
        (self.out / 'capture-result.json').write_text(json.dumps(value, indent=2) + '\n')


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('window'); p.add_argument('pid', type=int); p.add_argument('out', type=Path); p.add_argument('cases', type=Path)
    args = p.parse_args()
    Capture(args.window, args.pid, args.out, json.loads(args.cases.read_text())).execute()


if __name__ == '__main__':
    main()
