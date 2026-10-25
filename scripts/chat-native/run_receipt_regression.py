#!/usr/bin/env python3
"""Hosted metadata-only gate: a skipped real proof is a failure, never success."""
import os
from pathlib import Path
import sys
import unittest

REQUIRED=('ORBIT_RECEIPT_NODE','ORBIT_RECEIPT_PI_RUNTIME','ORBIT_RECEIPT_VENDOR_ARCHIVE','ORBIT_RECEIPT_RUST_SOURCE')

def run_strict(suite):
    result=unittest.TextTestRunner(verbosity=2).run(suite)
    if result.skipped or not result.wasSuccessful() or result.testsRun!=2:
        print('RECEIPT_PROOF_BLOCKED: exactly two real tests must pass with zero skips',file=sys.stderr)
        return 1
    print('RECEIPT_METADATA_AND_OWNED_ANCESTRY_VERIFIED; NATIVE_UI_NOT_RUN')
    return 0

def main():
    if os.environ.get('ORBIT_REQUIRE_PROC_CHILDREN')!='1':raise SystemExit('Required ancestry mode must be enabled')
    for name in REQUIRED:
        if not os.environ.get(name) or not Path(os.environ[name]).is_absolute() or not Path(os.environ[name]).exists():raise SystemExit('Missing explicit real fixture path: '+name)
    suite=unittest.defaultTestLoader.discover(str(Path(__file__).resolve().parent),pattern='test_receipt_real_pi.py')
    raise SystemExit(run_strict(suite))
if __name__=='__main__':main()
