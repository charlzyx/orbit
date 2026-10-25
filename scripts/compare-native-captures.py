#!/usr/bin/env python3
"""Compare actual same-state native captures without rescaling or masking.

Pixel equality never substitutes for manual semantic/interaction acceptance.
Requires Pillow; image bytes and metadata are input evidence, never modified.
"""
import argparse
import hashlib
import json
import math
from pathlib import Path

from PIL import Image, ImageChops

MATCH_FIELDS = ("requestedViewport", "palette", "language", "fontSize", "density", "reduceMotion", "state", "display", "dpi", "uiScale", "timezone")


def compare(reference, implementation, output):
    reference, implementation, output = map(Path, (reference, implementation, output))
    output.mkdir(parents=True, exist_ok=True)
    manifests = [json.loads((p / "manifest.json").read_text()) for p in (reference, implementation)]
    mismatches = {key: [m.get(key) for m in manifests] for key in MATCH_FIELDS if manifests[0].get(key) != manifests[1].get(key)}
    names = sorted(set(manifests[0]["actualImageSizes"]) | set(manifests[1]["actualImageSizes"]))
    if not names:
        raise ValueError("No native screenshot cases were declared; zero comparisons is not acceptance")
    missing = {str(i): [key for key in MATCH_FIELDS if m.get(key) is None] for i, m in enumerate(manifests)}
    normalization = []
    if any(missing.values()): normalization.append({"missing_required_metadata": missing})
    for i, m in enumerate(manifests):
        for key in ("resolvedFontFaces", "binarySha256", "sourceRootGitHead"):
            if not m.get(key): normalization.append({"application": i, "unverified": key})
        for key in ("clockFrozen", "animationPhaseFrozen", "focusAndCaretPhaseVerified"):
            if m.get(key) is not True: normalization.append({"application": i, "unverified": key})
        if m.get("sourceRootDirty") is not False: normalization.append({"application": i, "unverified": "clean source root"})
    if manifests[0].get("resolvedFontFaces") != manifests[1].get("resolvedFontFaces"):
        normalization.append({"mismatch": "resolved font faces"})
    report = {"reference_source": manifests[0]["source"], "implementation_source": manifests[1]["source"], "metadata_mismatches": mismatches, "normalized_fixture_gate": "BLOCKED" if normalization or mismatches else "METADATA_READY_MANUAL_REVIEW_REQUIRED", "normalization_gaps": normalization,
              "criteria": "Exact pixel equality: changed_pixels == 0. No tolerance, resizing, crop alignment or masking is applied. Manual semantic review remains independent.", "cases": []}
    for name in names:
        row = {"name": name, "reference": str(reference / name), "implementation": str(implementation / name), "manual_semantic_review": "UNREVIEWED"}
        paths = (reference / name, implementation / name)
        if not all(p.is_file() for p in paths):
            row.update(status="MISSING", reason="One or both native screenshots are absent")
        else:
            row["sha256"] = [hashlib.sha256(p.read_bytes()).hexdigest() for p in paths]
            with Image.open(paths[0]) as aa, Image.open(paths[1]) as bb:
                a, b = aa.convert("RGBA"), bb.convert("RGBA")
            row["dimensions"] = [list(a.size), list(b.size)]
            if a.size != b.size or list(a.size) != manifests[0]["requestedViewport"] or list(b.size) != manifests[1]["requestedViewport"]:
                row.update(status="RESHOOT", reason="Native client dimensions differ from each other or required viewport; comparison refused")
            elif mismatches:
                row.update(status="RESHOOT", reason="State/render metadata differ; comparison refused")
            else:
                diff = ImageChops.difference(a, b)
                values = list(diff.get_flattened_data() if hasattr(diff, "get_flattened_data") else diff.getdata())
                changed = sum(any(pixel) for pixel in values)
                channels = len(values) * 4
                absolute = sum(sum(pixel) for pixel in values)
                square = sum(sum(v * v for v in pixel) for pixel in values)
                # RGB max projected to L keeps alpha differences visible in bbox.
                channel_max = ImageChops.lighter(ImageChops.lighter(diff.getchannel(0), diff.getchannel(1)), ImageChops.lighter(diff.getchannel(2), diff.getchannel(3)))
                row.update(status="EXACT_PIXELS" if changed == 0 else "DIFFERENT", changed_pixels=changed,
                           changed_percent=100 * changed / len(values), mae=absolute / channels,
                           normalized_rmse=math.sqrt(square / channels) / 255,
                           max_channel_delta=max((max(p) for p in values), default=0), difference_bbox=channel_max.getbbox())
                diff.save(output / (Path(name).stem + "-difference.png"))
        report["cases"].append(row)
    report["counts"] = {status: sum(c["status"] == status for c in report["cases"]) for status in ("MISSING", "RESHOOT", "DIFFERENT", "EXACT_PIXELS")}
    (output / "comparison.json").write_text(json.dumps(report, indent=2) + "\n")
    return report


if __name__ == "__main__":
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("reference"); p.add_argument("implementation"); p.add_argument("output")
    args = p.parse_args()
    print(json.dumps(compare(args.reference, args.implementation, args.output)["counts"]))
