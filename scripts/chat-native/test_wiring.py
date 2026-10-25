import os
from pathlib import Path
import subprocess
import tempfile
import threading
import unittest

HERE=Path(__file__).resolve().parent
ROOT=HERE.parent.parent
class Wiring(unittest.TestCase):
    def test_workflow_is_opt_in_and_same_two_renderers(self):
        s=(ROOT/'.github/workflows/go-native.yml').read_text()
        self.assertIn('capture_chat_stage1:',s);self.assertIn('default: false',s)
        expression="ORBIT_CAPTURE_CHAT_STAGE1: ${{ github.event_name == 'workflow_dispatch' && inputs.capture_chat_stage1 && '1' || '0' }}"
        self.assertEqual(s.count(expression),2)
        self.assertEqual(s.count('"scripts/chat-native/**"'),2)
        self.assertEqual(s.count('"testdata/chat-native/**"'),2)
        self.assertIn("python3 -m unittest discover -s scripts/chat-native",s)
        hook=(ROOT/'scripts/capture-native-reference.sh').read_text()
        self.assertIn('if [[ "${ORBIT_CAPTURE_CHAT_STAGE1:-0}" == 1 ]]',hook)
        self.assertIn('"$output/chat"',hook)
    def test_existing_root_refused_before_preflight_and_launch(self):
        s=(HERE/'capture.sh').read_text()
        check='[[ ! -e "$fixture" && ! -L "$fixture" ]]'
        self.assertLess(s.index(check),s.index('preflight.py'));self.assertLess(s.index(check),s.index('setsid '))
        # Evaluate the exact shell refusal with both directory and dangling link.
        guard=s[s.index(check):s.index('\n',s.index(check))]
        with tempfile.TemporaryDirectory() as d:
            root=Path(d)/'root';root.mkdir()
            for kind in ('directory','dangling'):
                if kind=='dangling':root.rmdir();root.symlink_to(Path(d)/'absent')
                p=subprocess.run(['bash','-c','fixture=$1; '+guard,'test',str(root)],capture_output=True,text=True)
                self.assertNotEqual(p.returncode,0);self.assertIn('Refusing existing Chat fixture',p.stdout)
    def test_cleanup_signals_only_matching_owned_process_group(self):
        child=subprocess.Popen(['sleep','60'],start_new_session=True)
        helper=str(HERE/'owned-process.sh')
        try:
            setup='source "$1"; app_pid=$2; app_identity=$(chat_identity "$app_pid"); '
            mismatch=subprocess.run(['bash','-c',setup+'app_identity+=wrong; chat_cleanup','test',helper,str(child.pid)],capture_output=True,text=True)
            self.assertEqual(mismatch.returncode,0);self.assertIsNone(child.poll())
            # Reap our child promptly so the cleanup loop observes its exit.
            waiter=threading.Thread(target=child.wait);waiter.start()
            matching=subprocess.run(['bash','-c',setup+'chat_cleanup','test',helper,str(child.pid)],capture_output=True,text=True,timeout=8)
            waiter.join(2);self.assertEqual(matching.returncode,0);self.assertIsNotNone(child.poll())
        finally:
            if child.poll() is None:child.terminate();child.wait()
