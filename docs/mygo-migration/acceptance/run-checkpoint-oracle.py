#!/usr/bin/env python3
"""Run the five unchanged pinned Rust checkpoint tests without GPUI.

Set CARGO/RUSTC/CARGO_HOME as needed. Dependencies must already be cached:
the runner is offline and never installs a toolchain or accesses user repos.
"""
import hashlib
import json
import os
import pathlib
import subprocess
import tempfile

ROOT = pathlib.Path(__file__).resolve().parents[3]
SOURCE = ROOT / "crates/orbit-pi/src/checkpoint.rs"
EXPECTED = "a1df0d342f0444e9ff824bf5abfd1b3fff508758f6f4c14eb389d98ca70c03dd"
assert hashlib.sha256(SOURCE.read_bytes()).hexdigest() == EXPECTED, "pinned source changed"
with tempfile.TemporaryDirectory(prefix="orbit-checkpoint-oracle-") as directory:
    root = pathlib.Path(directory)
    (root / "src").mkdir()
    (root / "Cargo.toml").write_text('''[package]
name = "orbit-checkpoint-oracle"
version = "0.0.0"
edition = "2021"
[dependencies]
anyhow = "=1.0.104"
serde_json = "=1.0.151"
''')
    # Only process/l10n adapter glue is supplied. All tested checkpoint code and
    # the five test bodies come directly from the pinned original source file.
    adapter = r'''
macro_rules! tr { ($key:literal, status = $status:expr) => { format!("git exited with {}", $status) }; }
pub mod git {
    use std::{ffi::OsStr,path::Path,process::Command};
    pub fn command(cwd: &Path) -> Command {
        let mut cmd=Command::new("git");
        cmd.current_dir(cwd).env("GIT_OPTIONAL_LOCKS", "0").env("GIT_TERMINAL_PROMPT", "0");
        cmd
    }
    pub fn run_git_ok<I,S>(cwd: &Path,args:I) -> anyhow::Result<String>
        where I:IntoIterator<Item=S>,S:AsRef<OsStr> {
        let output=command(cwd).args(args).output()?;
        if !output.status.success(){anyhow::bail!("{}",String::from_utf8_lossy(&output.stderr));}
        Ok(String::from_utf8_lossy(&output.stdout).into_owned())
    }
}
'''
    (root / "src/lib.rs").write_text(adapter + "\n#[path=" + json.dumps(str(SOURCE)) + "]\npub mod checkpoint;\n")
    command = [os.environ.get("CARGO", "cargo"), "test", "--offline", "--manifest-path", str(root / "Cargo.toml"), "--lib", "--", "--test-threads=1"]
    raise SystemExit(subprocess.run(command, cwd=root).returncode)
