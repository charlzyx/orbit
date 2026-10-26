import io
from pathlib import Path
import re
import unittest
from contextlib import redirect_stderr,redirect_stdout
import run_receipt_regression as runner
from test_wiring import MANUAL_EXPRESSION, ONE_SHOT_EXPRESSION

ROOT=Path(__file__).resolve().parent.parent.parent

def job(source,name):
    return re.search(r'^  '+re.escape(name)+r':\n(.*?)(?=^  [a-z][\w-]*:\n|\Z)',source,re.M|re.S).group(1)

class RequiredReceiptWiring(unittest.TestCase):
    def test_required_metadata_job_has_no_chat_input_or_optional_event_condition(self):
        source=(ROOT/'.github/workflows/go-native.yml').read_text();receipt=job(source,'receipt-guard-regression')
        for expected in ('needs: proc-guard-regression','runs-on: ubuntu-24.04',"node-version: '22.23.4'",'299a32bc87fe3d4851bdb902d9c03dcf9bf99cac','@earendil-works/pi-coding-agent@1.4.0','ORBIT_REQUIRE_PROC_CHILDREN: "1"','run_receipt_regression.py','set -o pipefail','if: always()'):
            self.assertIn(expected,receipt)
        for name in runner.REQUIRED:self.assertIn(name,receipt)
        for forbidden in ('capture.sh','xdotool','ORBIT_CAPTURE_CHAT_STAGE1','continue-on-error','workflow_dispatch','pull_request_target'):
            self.assertNotIn(forbidden,receipt)
        self.assertNotRegex(receipt,r'^    if:', 'metadata gate cannot be opt-in')
        for name in ('native-desktop','reference-desktop'):
            self.assertIn('needs: [proc-guard-regression, receipt-guard-regression]',job(source,name))
        expressions=[line.split('${{',1)[1].rsplit('}}',1)[0].strip() for line in source.splitlines()if 'ORBIT_CAPTURE_CHAT_STAGE1:' in line]
        self.assertIn(expressions, [ [MANUAL_EXPRESSION]*2, [ONE_SHOT_EXPRESSION]*2 ])
        self.assertIn('default: false',source)
    def test_strict_runner_rejects_skips_and_missing_test_count(self):
        class Skipped(unittest.TestCase):
            @unittest.skip('deliberate negative fixture')
            def test_skip(self):pass
        class Pass(unittest.TestCase):
            def test_pass(self):pass
        with redirect_stdout(io.StringIO()),redirect_stderr(io.StringIO()):
            self.assertEqual(runner.run_strict(unittest.defaultTestLoader.loadTestsFromTestCase(Skipped)),1)
            self.assertEqual(runner.run_strict(unittest.defaultTestLoader.loadTestsFromTestCase(Pass)),1)
            self.assertEqual(runner.run_strict(unittest.TestSuite([Pass('test_pass'),Pass('test_pass')])),0)
