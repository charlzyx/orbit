#!/usr/bin/env python3
"""Regenerate only this audited settings glossary's existing Chinese catalog values.

The historical all-locale generator has unrelated baseline drift. Reuse its
parser, merged glossary and quoting; leave every other key/locale byte intact.
"""
from pathlib import Path
import gen_locales
from i18n_glossary_settings_completion import GLOSSARY


def main():
    target = Path(gen_locales.LOCALES) / 'zh-CN.yml'
    english = dict(gen_locales.parse(gen_locales.EN))
    selected = GLOSSARY['zh-CN']
    lines = target.read_text().splitlines(keepends=True)
    count = 0
    for i, line in enumerate(lines):
        key = line.partition(': ')[0]
        source = english.get(key)
        if source in selected:
            lines[i] = f'{key}: {gen_locales.quote(gen_locales.GLOSSARY["zh-CN"][source])}\n'
            count += 1
    target.write_text(''.join(lines))
    print(f'Generated {count} existing settings-related Chinese keys; other locales unchanged.')


if __name__ == '__main__':
    main()
