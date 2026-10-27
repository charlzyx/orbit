"""Settings glossary correctness and narrow regeneration (no repository writes)."""
import contextlib
import io
from pathlib import Path
import re
import tempfile
import unittest
import gen_locales
from i18n_glossary_settings_completion import GLOSSARY
import update_settings_chinese

class SettingsChinese(unittest.TestCase):
    def test_existing_keys_match_glossary(self):
        en = dict(gen_locales.parse(gen_locales.EN))
        zh = dict(gen_locales.parse(str(Path(gen_locales.LOCALES) / 'zh-CN.yml')))
        expected = {k: gen_locales.GLOSSARY['zh-CN'][v] for k, v in en.items() if v in GLOSSARY['zh-CN']}
        self.assertGreater(len(expected), 200)
        for key, value in expected.items():
            self.assertEqual(zh[key], value, key)

    def test_placeholders_preserved(self):
        pattern = r'%\{[^}]+\}|\$\{[^}]+\}'
        for en, zh in GLOSSARY['zh-CN'].items():
            self.assertEqual(sorted(re.findall(pattern, en)), sorted(re.findall(pattern, zh)), en)

    def test_scoped_generation_preserves_unrelated_bytes_and_is_idempotent(self):
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory) / 'zh-CN.yml'
            target.write_text('# retained comment\nsettings.session_retention: "old"\nunrelated: "DO NOT TOUCH"\n')
            old = gen_locales.LOCALES
            try:
                gen_locales.LOCALES = directory
                with contextlib.redirect_stdout(io.StringIO()):
                    update_settings_chinese.main()
                    first = target.read_bytes()
                    update_settings_chinese.main()
                self.assertIn('settings.session_retention: \"会话保留\"\n'.encode(), first)
                self.assertNotIn(b'\"old\"', first)
                self.assertEqual(first, target.read_bytes())
                self.assertIn(b'# retained comment\n', first)
                self.assertIn(b'unrelated: "DO NOT TOUCH"\n', first)
            finally:
                gen_locales.LOCALES = old

if __name__ == '__main__':
    unittest.main()
