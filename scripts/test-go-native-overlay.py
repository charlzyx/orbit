#!/usr/bin/env python3
"""Read-only unit checks for canonical-wrapper flag preservation."""
import importlib.util
from pathlib import Path
import unittest

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

if __name__ == '__main__': unittest.main()
