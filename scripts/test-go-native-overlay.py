#!/usr/bin/env python3
"""Read-only unit checks for canonical-wrapper flag preservation."""
import importlib.util
from pathlib import Path
import unittest
import tempfile
import hashlib
import json
import copy
from unittest import mock
import shutil
import subprocess

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

class MenuHitManifest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.module, self.compat = self.root / "module", self.root / "compat"
        (self.module / "ui").mkdir(parents=True)
        self.compat.mkdir()
        self.identity = {k: "fixture-" + k for k in overlay.MENU_HIT_IDENTITIES}
        self.manifest = dict(self.identity, baseline_sha256={}, overlays={})
        for target in overlay.MENU_HIT_BASELINE | overlay.MENU_HIT_ADDED:
            source = Path(target).name + ".txt"
            added = target in overlay.MENU_HIT_ADDED
            (self.compat / source).write_text("reviewed fixture overlay " + target)
            if not added:
                (self.module / target).write_text("upstream fixture " + target)
                self.manifest["baseline_sha256"][target] = overlay.sha(self.module / target)
            self.manifest["overlays"][target] = {"source":source, "sha256":overlay.sha(self.compat/source), "added":added}
        (self.compat / "menu-hit.patch").write_text("fixture reviewed patch")
        self.manifest["patch_sha256"] = overlay.sha(self.compat / "menu-hit.patch")
        self.save()
    def save(self):
        (self.compat / "baseline.json").write_text(json.dumps(self.manifest))
    def verify(self):
        return overlay.verify_menu_hit(self.module, self.compat, self.identity)
    def test_exact_explicit_files(self):
        self.assertEqual(set(self.verify()), overlay.MENU_HIT_BASELINE | overlay.MENU_HIT_ADDED)
        (self.compat / "unlisted.go.txt").write_text("must never be discovered")
        self.assertEqual(len(self.verify()), 4)
    def test_rejects_identity_hash_and_patch_mismatch(self):
        original = copy.deepcopy(self.manifest)
        for mutate in [
            lambda m: m.update(version="unreviewed"),
            lambda m: m["baseline_sha256"].update({"ui/menu.go":"bad"}),
            lambda m: m["overlays"]["ui/menu.go"].update(sha256="bad"),
            lambda m: m.update(patch_sha256="bad"),
        ]:
            self.manifest=copy.deepcopy(original); mutate(self.manifest); self.save()
            with self.assertRaises(RuntimeError): self.verify()
    def test_rejects_unknown_missing_traversal_and_kind(self):
        original=copy.deepcopy(self.manifest)
        for mutate in [
            lambda m: m["overlays"].update({"ui/not_reviewed.go":{}}),
            lambda m: m["baseline_sha256"].pop("ui/menu.go"),
            lambda m: m["overlays"]["ui/menu.go"].update(source="../../outside.go"),
            lambda m: m["overlays"]["ui/menu.go"].update(added=True),
            lambda m: m["overlays"]["ui/menu.go"].update(added=0),
            lambda m: m.update(unreviewed_key=True),
        ]:
            self.manifest=copy.deepcopy(original); mutate(self.manifest); self.save()
            with self.assertRaises(RuntimeError): self.verify()
    def test_rejects_existing_added_file_and_symlink_escape(self):
        target=self.module / "ui/orbit_menu_hit.go"
        target.write_text("unexpected cache mutation")
        with self.assertRaises(RuntimeError): self.verify()
        target.unlink()
        source=self.compat / "menu.go.txt"
        outside=self.root / "outside.txt"
        outside.write_bytes(source.read_bytes())
        source.unlink(); source.symlink_to(outside)
        with self.assertRaises(RuntimeError): self.verify()

class DiagnosticModes(unittest.TestCase):
    def test_normal_does_not_import_diagnostic_assets(self):
        with mock.patch.object(overlay, "import_verifier", side_effect=AssertionError("must not import")):
            self.assertEqual(overlay.verify_diagnostic(Path("/absent"), None), {})
        for flags, args in [([], ["build", "./cmd/orbit"]), (["-tags=user"], ["test", "-tags=last", "./..."])]:
            self.assertEqual(overlay.diagnostic_flags(flags, args, None), (flags, args))
        self.assertEqual(overlay.diagnostic_mode(["-C", "third_party/mygo-terminal", "test"]), (None, ["-C", "third_party/mygo-terminal", "test"]))

    def test_explicit_modes_and_unknowns(self):
        for mode in overlay.DIAGNOSTIC_MODES:
            self.assertEqual(overlay.diagnostic_mode([mode, "build"]), (mode, ["build"]))
        for args in [["--diagnostic-unreviewed", "build"], ["build", "--diagnostic-text-metrics"], ["--diagnostic-text-metrics", "--diagnostic-text-metrics-observed", "build"], ["--diagnostic-text-metrics"]]:
            with self.assertRaises(RuntimeError): overlay.diagnostic_mode(args)
        with self.assertRaises(RuntimeError): overlay.verify_diagnostic(Path("/absent"), "unknown")

    def test_tag_preserves_effective_caller_choice(self):
        mode = "--diagnostic-text-metrics"
        tag = overlay.DIAGNOSTIC_TAG
        self.assertEqual(overlay.diagnostic_flags([], ["build"], mode), (["-tags=" + tag], ["build"]))
        self.assertEqual(overlay.diagnostic_flags(["-tags=inherited"], ["test", "-tags", "caller,other"], mode),
                         (["-tags=inherited"], ["test", "-tags", "caller,other," + tag]))
        self.assertEqual(overlay.diagnostic_flags(["-tags=inherited"], ["test"], mode),
                         (["-tags=inherited," + tag], ["test"]))
        for mode in [None, "--diagnostic-text-metrics"]:
            with self.assertRaises(RuntimeError): overlay.diagnostic_flags(["-tags=other," + tag], ["test"], mode)
        with self.assertRaises(RuntimeError): overlay.diagnostic_flags([], ["build", "-tags"], mode)

    def test_startup_is_explicit_and_pre_asset_not_init(self):
        root = overlay.ROOT / "internal/ui"
        production = (root / "diagnostic_text_metrics_default.go").read_text()
        enabled = (root / "diagnostic_text_metrics_linux.go").read_text()
        unsupported = (root / "diagnostic_text_metrics_unsupported.go").read_text()
        self.assertIn("!" + overlay.DIAGNOSTIC_TAG, production)
        self.assertNotIn("EnableOrbitFractionalTextMetrics", production)
        self.assertIn("ui.EnableOrbitFractionalTextMetrics()", enabled)
        self.assertIn("no fallback", unsupported)
        for text in (production, enabled, unsupported):
            self.assertNotIn("func init(", text)
            self.assertNotIn("os.Getenv", text)
        run = (root / "run.go").read_text().split("func Run(o RunOptions) error {", 1)[1]
        self.assertLess(run.index("configureDiagnosticTextMetrics()"), run.index("LoadAssets()"))
        tests = (root / "snapshot_test.go").read_text().split("func TestMain(m *testing.M) {", 1)[1]
        self.assertLess(tests.index("configureDiagnosticTextMetrics()"), tests.index("m.Run()"))

    def test_any_unreviewed_collision_or_escape_fails(self):
        self.assertEqual(overlay.merge_reviewed([{"ui/a.go":Path("a")}, {"ui/b.go":Path("b")}]),
                         {"ui/a.go":Path("a"), "ui/b.go":Path("b")})
        for maps in [[{"ui/a.go":Path("a")}, {"ui/a.go":Path("a")}], [{"../escape.go":Path("a")}], [{"/absolute.go":Path("a")}]]:
            with self.assertRaises(RuntimeError): overlay.merge_reviewed(maps)


class CombinedDiagnosticVerification(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.source = self.root / "source"
        self.assets = self.root / "assets"
        root = overlay.ROOT / "scripts/font-fidelity"
        shutil.copytree(root, self.assets, ignore=shutil.ignore_patterns("__pycache__"))
        self.verifier = overlay.import_verifier(root / "combined-verify.py")
        self.candidate_root = overlay.ROOT / "scripts/mygo-compat/text-metrics"
        self.candidate = overlay.import_verifier(self.candidate_root / "prepare.py")
        # Reconstruct only the reviewed pristine files, without module cache or
        # a particular checkout. Reverse patch also verifies exact derivation.
        for target, filename in self.verifier.FILES.items():
            path = self.source / target
            path.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(self.assets / filename, path)
        subprocess.run(["git", "apply", "--reverse", str(self.assets / "combined-observer.patch")], cwd=self.source, check=True, capture_output=True)
        _, self.mapping = self.candidate.verify(self.source)

    def verify(self):
        return self.verifier.verify(self.source, self.mapping, self.candidate_root / "baseline.json", self.assets)

    def mutate(self, fn):
        p = self.assets / "combined-baseline.json"
        m = json.loads(p.read_text()); fn(m); p.write_text(json.dumps(m))

    def test_exact_composed_paths_preserve_both_components(self):
        before = {str(p.relative_to(self.source)): overlay.sha(p) for p in self.source.rglob("*") if p.is_file()}
        result = self.verify()
        self.assertEqual(len(result), 11)
        self.assertEqual(set(result), set(self.mapping) | set(self.verifier.FILES))
        for target, source in self.mapping.items():
            if target != "internal/text/pango_linux.go": self.assertEqual(result[target], source)
        pango = result["internal/text/pango_linux.go"].read_text()
        self.assertIn("orbitFractionalMetrics", pango)
        self.assertIn("e.orbitShape(", pango)
        self.assertIn("e.orbitScaled(", pango)
        observer = result["internal/text/orbit_mygo_pango_trace.go"].read_text()
        self.assertIn("pango_context_get_round_glyph_positions", observer)
        self.assertIn("pango_cairo_context_get_font_options", observer)
        self.assertNotIn('"round_positions": pangoLib.contextSetRoundPositions != nil', observer)
        self.assertNotIn('"context_hint_metrics": cairoHintMetricsOn', observer)
        after = {str(p.relative_to(self.source)): overlay.sha(p) for p in self.source.rglob("*") if p.is_file()}
        self.assertEqual(before, after)

    def test_reject_unknown_overlay_paths(self):
        self.mutate(lambda m: m["overlays"].update({"ui/unreviewed.go": {}}))
        with self.assertRaises(ValueError): self.verify()

    def test_reject_collision_manifest_expansion(self):
        self.mutate(lambda m: m["intentional_collisions"].append("ui/paint.go"))
        with self.assertRaises(ValueError): self.verify()

    def test_reject_input_and_candidate_drift(self):
        p = self.assets / "mygo_pango_trace.go.txt"
        p.write_text(p.read_text() + "\n")
        with self.assertRaises(ValueError): self.verify()
        self.mapping["ui/unreviewed.go"] = p
        with self.assertRaises(ValueError): self.verify()

    def test_reject_candidate_source_substitution(self):
        p = self.root / "different-candidate.go"
        target = "ui/orbit_text_metrics_linux.go"
        p.write_bytes(self.mapping[target].read_bytes())
        self.mapping[target] = p
        with self.assertRaises(ValueError): self.verify()

    def test_reject_combined_source_drift(self):
        p = self.assets / "combined_pango_linux.go.txt"
        p.write_text(p.read_text() + "\n")
        with self.assertRaises(ValueError): self.verify()

    def test_reject_additive_collision(self):
        p = self.source / "internal/text/orbit_mygo_trace.go"
        p.write_text("already here")
        with self.assertRaises(ValueError): self.verify()

    def test_reject_symlink_escape(self):
        p = self.assets / "combined_pango_linux.go.txt"
        outside = self.root / "outside.txt"; outside.write_bytes(p.read_bytes())
        p.unlink(); p.symlink_to(outside)
        with self.assertRaises(ValueError): self.verify()

if __name__ == '__main__': unittest.main()
