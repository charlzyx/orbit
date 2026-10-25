#!/usr/bin/env python3
"""Read-only unit checks for canonical-wrapper flag preservation."""
import importlib.util
from pathlib import Path
import unittest
import tempfile
import hashlib
import json
import copy

spec = importlib.util.spec_from_file_location("overlay", Path(__file__).with_name("go-native-overlay.py"))
overlay = importlib.util.module_from_spec(spec)
spec.loader.exec_module(overlay)

class WrapperFlags(unittest.TestCase):
    def test_go_flags_quote_grammar(self):
        self.assertEqual(overlay.split_flags('-trimpath "-ldflags=-s -w"'), ['-trimpath', '-ldflags=-s -w'])
        self.assertEqual(overlay.split_flags(r'"-overlay=C:\some path\x.json" -race'), [r'-overlay=C:\some path\x.json', '-race'])
        self.assertEqual(overlay.split_flags("-tags='one,two'"), ["-tags='one,two'"])
        self.assertEqual(overlay.split_flags('"x"y'), ['x', 'y'])
        self.assertEqual(overlay.split_flags("'$(no-execution); x'"), ['$(no-execution); x'])
        with self.assertRaises(RuntimeError): overlay.split_flags('"missing')
    def test_roundtrip(self):
        for words in [['-race','-trimpath'], ['-ldflags=-X main.version=two words'], [r'-overlay=C:\some path\x.json'], ["literal'quote"], ['"first'], ['']]:
            self.assertEqual(overlay.split_flags(' '.join(overlay.quote_flag(w) for w in words)), words)
    def test_overlay_detach(self):
        self.assertEqual(overlay.detach_overlay(['-race','-overlay=x','-tags=test']), (['-race','-tags=test'],'x'))
        self.assertEqual(overlay.detach_overlay(['-overlay','x','-overlay=y']), ([], 'y'))
        with self.assertRaises(RuntimeError): overlay.detach_overlay(['-overlay'])
    def test_default_trimpath_preserves_explicit_debug_choice(self):
        original = ['-race']
        self.assertEqual(overlay.default_trimpath(original, ['test', './...']), ['-race', '-trimpath'])
        self.assertEqual(original, ['-race'])
        for choice in ['-trimpath', '-trimpath=true', '-trimpath=false']:
            self.assertEqual(overlay.default_trimpath([choice], ['build']), [choice])
            self.assertEqual(overlay.default_trimpath(['-race'], ['build', choice]), ['-race'])

class SelectionManifest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.module, self.compat = self.root / "module", self.root / "compat"
        (self.module / "ui").mkdir(parents=True)
        self.compat.mkdir()
        self.identity = {k: "fixture-" + k for k in overlay.SELECTION_IDENTITIES}
        self.manifest = dict(self.identity, baseline_sha256={}, overlays={})
        for target in overlay.SELECTION_BASELINE | overlay.SELECTION_ADDED:
            source = Path(target).name + ".txt"
            added = target in overlay.SELECTION_ADDED
            (self.compat / source).write_text("reviewed fixture overlay " + target)
            if not added:
                (self.module / target).write_text("upstream fixture " + target)
                self.manifest["baseline_sha256"][target] = overlay.sha(self.module / target)
            self.manifest["overlays"][target] = {"source":source, "sha256":overlay.sha(self.compat/source), "added":added}
        (self.compat / "selection-retention.patch").write_text("fixture reviewed patch")
        self.manifest["patch_sha256"] = overlay.sha(self.compat / "selection-retention.patch")
        self.save()
    def save(self):
        (self.compat / "baseline.json").write_text(json.dumps(self.manifest))
    def verify(self):
        return overlay.verify_selection(self.module, self.compat, self.identity)
    def test_exact_explicit_files(self):
        self.assertEqual(set(self.verify()), overlay.SELECTION_BASELINE | overlay.SELECTION_ADDED)
        (self.compat / "unlisted.go.txt").write_text("must never be discovered")
        self.assertEqual(len(self.verify()), 5)
    def test_rejects_identity_hash_and_patch_mismatch(self):
        original = copy.deepcopy(self.manifest)
        for mutate in [
            lambda m: m.update(version="unreviewed"),
            lambda m: m["baseline_sha256"].update({"ui/input.go":"bad"}),
            lambda m: m["overlays"]["ui/input.go"].update(sha256="bad"),
            lambda m: m.update(patch_sha256="bad"),
        ]:
            self.manifest=copy.deepcopy(original); mutate(self.manifest); self.save()
            with self.assertRaises(RuntimeError): self.verify()
    def test_rejects_unknown_missing_traversal_and_kind(self):
        original=copy.deepcopy(self.manifest)
        for mutate in [
            lambda m: m["overlays"].update({"ui/not_reviewed.go":{}}),
            lambda m: m["baseline_sha256"].pop("ui/input.go"),
            lambda m: m["overlays"]["ui/input.go"].update(source="../../outside.go"),
            lambda m: m["overlays"]["ui/input.go"].update(added=True),
            lambda m: m["overlays"]["ui/input.go"].update(added=0),
            lambda m: m.update(unreviewed_key=True),
        ]:
            self.manifest=copy.deepcopy(original); mutate(self.manifest); self.save()
            with self.assertRaises(RuntimeError): self.verify()
    def test_rejects_existing_added_file_and_symlink_escape(self):
        target=self.module / "ui/orbit_selection.go"
        target.write_text("unexpected cache mutation")
        with self.assertRaises(RuntimeError): self.verify()
        target.unlink()
        source=self.compat / "input.go.txt"
        outside=self.root / "outside.txt"
        outside.write_bytes(source.read_bytes())
        source.unlink(); source.symlink_to(outside)
        with self.assertRaises(RuntimeError): self.verify()

if __name__ == '__main__': unittest.main()
