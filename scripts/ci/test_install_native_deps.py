"""Offline command fixtures only: no sudo, apt, networking, or native input."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / 'scripts/ci/install-native-deps.sh'

class SetupPhaseTests(unittest.TestCase):
    def invoke(self, kind, failure='none'):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            fake = root / 'sudo'
            fake.write_text('''#!/bin/bash
printf '%s\\n' "$*" >> "$CALLS"
[[ "$1" == apt-get ]] || exit 99
case "$FAILURE:$2" in
 update:update) exit 42;;
 install:install) exit 43;;
 signal:update) kill -TERM "$$";;
esac
exit 0
''')
            fake.chmod(0o755)
            env = dict(os.environ, PATH=str(root)+os.pathsep+os.environ['PATH'],
                       CALLS=str(root/'calls'), FAILURE=failure)
            result = subprocess.run(['bash', str(SCRIPT), kind], env=env,
                                    text=True, capture_output=True, timeout=5)
            calls = (root/'calls').read_text().splitlines() if (root/'calls').exists() else []
            return result, calls

    def test_both_profiles_keep_exact_original_package_arguments(self):
        expected = {
            'go': 'libgtk-3-0 libpango-1.0-0 libcairo2 libfontconfig1 libgl1-mesa-dri mesa-vulkan-drivers vulkan-tools xvfb openbox xdotool wmctrl x11-utils imagemagick tesseract-ocr',
            'reference': 'clang cmake pkg-config libfontconfig-dev libvulkan1 mesa-vulkan-drivers vulkan-tools libwayland-dev libx11-xcb-dev libxi-dev libxtst-dev libxkbcommon-x11-dev libssl-dev libgtk-3-0 libpango-1.0-0 libcairo2 xvfb openbox xdotool wmctrl x11-utils imagemagick tesseract-ocr',
        }
        for kind, packages in expected.items():
            with self.subTest(kind=kind):
                result, calls = self.invoke(kind)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(calls, ['apt-get update', 'apt-get install -y '+packages])
                self.assertEqual(result.stdout.count('ORBIT_APT_PHASE_START'), 2)
                self.assertEqual(result.stdout.count('ORBIT_APT_PHASE_END'), 2)
                self.assertRegex(result.stdout, r'utc=\d{4}-\d{2}-\d{2}T.*Z')
                self.assertIn('elapsed_s=', result.stdout)

    def test_update_error_is_preserved_and_install_not_started(self):
        result, calls = self.invoke('go', 'update')
        self.assertEqual(result.returncode, 42)
        self.assertEqual(calls, ['apt-get update'])
        self.assertIn('phase=update', result.stdout)
        self.assertIn('exit=42', result.stdout)
        self.assertNotIn('phase=install', result.stdout)

    def test_install_error_is_preserved(self):
        result, calls = self.invoke('reference', 'install')
        self.assertEqual(result.returncode, 43)
        self.assertEqual(len(calls), 2)
        self.assertIn('exit=43', result.stdout)

    def test_signalled_command_is_not_reported_success(self):
        result, calls = self.invoke('go', 'signal')
        self.assertEqual(result.returncode, 143)
        self.assertEqual(calls, ['apt-get update'])
        self.assertIn('exit=143', result.stdout)

    def test_invalid_profile_has_no_command_side_effect(self):
        result, calls = self.invoke('other')
        self.assertEqual(result.returncode, 64)
        self.assertEqual(calls, [])

    def test_workflow_limits_remain_separate_and_failure_not_ignored(self):
        workflow = (ROOT/'.github/workflows/go-native.yml').read_text()
        for job, budget, profile in [('reference-desktop',45,'reference'),('native-desktop',15,'go')]:
            body = workflow.split('  '+job+':\n',1)[1]
            if job == 'reference-desktop': body = body.split('\n  native-desktop:',1)[0]
            self.assertIn('    timeout-minutes: '+str(budget), body)
            self.assertIn('        timeout-minutes: 5\n        run: bash scripts/ci/install-native-deps.sh '+profile, body)
            self.assertNotIn('continue-on-error:', body)
        self.assertEqual(workflow.count('"scripts/ci/**"'),2)

if __name__ == '__main__':
    unittest.main()
