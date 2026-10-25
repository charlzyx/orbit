import os
from pathlib import Path
import subprocess
import unittest

HERE=Path(__file__).resolve().parent
class HostedScope(unittest.TestCase):
    def test_non_hosted_environment_refuses_before_any_argument_or_launch(self):
        result=subprocess.run(['bash',str(HERE/'capture.sh')],env={'PATH':os.environ['PATH'],'DISPLAY':':92','RUNNER_OS':'Linux'},capture_output=True,text=True)
        self.assertNotEqual(result.returncode,0);self.assertIn('approved hosted Linux',result.stderr)
    def test_shell_has_no_disallowed_action_or_workflow_wiring(self):
        source=(HERE/'capture.sh').read_text();self.assertNotIn('click',source);self.assertNotIn('--session',source)
        self.assertNotIn('--model',source);self.assertNotIn('ORBIT_NO_RPC_PATCHES',source)
        self.assertIn('v22.23.4',source);self.assertIn('env -i',source)
if __name__=='__main__':unittest.main()
