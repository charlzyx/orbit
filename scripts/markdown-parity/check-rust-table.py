#!/usr/bin/env python3
"""Run verbatim pinned Rust parser fixture with the existing CI Rust compiler.
No downloads, Cargo dependencies, GUI actions, or repository writes.
"""
import hashlib
from pathlib import Path
import subprocess
import sys
import tempfile

source = Path(sys.argv[1]) / 'crates/orbit-pi/src/transcript_view.rs'
raw = source.read_bytes()
assert hashlib.sha256(raw).hexdigest() == 'cf06a57001b3fddcdf4d6eaca5f13814b4262a070c2b093fa0469e56a10f7abc', 'pinned Rust source changed'
text = raw.decode()
fixture = Path(__file__).with_name('rust-table-parser.rs').read_text()
for start, end in [
 ('#[derive(Clone)]\nenum Block {', 'impl TableAlign {'),
 ('#[derive(Clone, Copy, Debug, PartialEq, Eq)]\nenum AlertKind {', 'impl AlertKind {'),
 ('#[derive(Clone)]\nstruct ListItem {', 'thread_local! {'),
]:
    begin = text.index(start)
    piece = text[begin:text.index(end, begin)]
    assert piece in fixture, 'fixture differs from actual pinned parser'
with tempfile.TemporaryDirectory(prefix='orbit-rust-table-') as directory:
    binary = str(Path(directory) / 'parser-fixture')
    subprocess.run(['rustc', '--edition=2021', '--crate-name', 'orbit_table_parser_fixture', str(Path(__file__).with_name('rust-table-parser.rs')), '-o', binary], check=True)
    subprocess.run([binary], check=True)
