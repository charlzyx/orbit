#!/usr/bin/env python3
import importlib.util
import itertools
from pathlib import Path
import unittest

ROOT = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("candidate_specimens", ROOT / "run-specimens.py")
M = importlib.util.module_from_spec(spec)
spec.loader.exec_module(M)


class SpecimenRosterTests(unittest.TestCase):
    def setUp(self):
        self.rows = [{"style": {"Family": f, "Italic": i, "Weight": w, "Size": 14}, "text": t}
                     for f, i, w, t in itertools.product(
                         (".ZedSans", ".ZedMono"), (False, True), (400, 500, 600, 700),
                         ("General", "This machine", "AV office fi 0123456789 → 中文 🙂"))]

    def test_exact_roster(self):
        M.check_roster(self.rows)

    def test_missing_case(self):
        with self.assertRaises(ValueError):
            M.check_roster(self.rows[:-1])

    def test_duplicate_substitution(self):
        self.rows[0] = self.rows[1]
        with self.assertRaises(ValueError):
            M.check_roster(self.rows)

    def test_wrong_size(self):
        self.rows[0]["style"]["Size"] = 16
        with self.assertRaises(ValueError):
            M.check_roster(self.rows)


if __name__ == "__main__":
    unittest.main(verbosity=2)
