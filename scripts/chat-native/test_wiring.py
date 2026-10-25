import os
from pathlib import Path
import subprocess
import tempfile
import threading
import unittest

HERE=Path(__file__).resolve().parent
ROOT=HERE.parent.parent
MANUAL_EXPRESSION="github.event_name == 'workflow_dispatch' && inputs.capture_chat_stage1 && '1' || '0'"
ONE_SHOT_EXPRESSION="((github.event_name == 'workflow_dispatch' && inputs.capture_chat_stage1) || (github.event_name == 'push' && github.event.before == 'a5c40cc2f85b57b94c1e2027d5cfc690b8692a49' && github.event.head_commit.message == 'Capture static Chat stage 1 after verified threaded ownership')) && '1' || '0'"
class Wiring(unittest.TestCase):
    def test_workflow_is_opt_in_and_same_two_renderers(self):
        s=(ROOT/'.github/workflows/go-native.yml').read_text()
        self.assertIn('capture_chat_stage1:',s);self.assertIn('default: false',s)
        expressions=[line.split('${{',1)[1].rsplit('}}',1)[0].strip()
                     for line in s.splitlines() if 'ORBIT_CAPTURE_CHAT_STAGE1:' in line]
        self.assertEqual(len(expressions),2)
        self.assertEqual(expressions[0],expressions[1])
        self.assertIn(expressions[0],(MANUAL_EXPRESSION,ONE_SHOT_EXPRESSION))
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


class OneShotEventMatrix(unittest.TestCase):
    def test_exact_workflow_expression_event_matrix(self):
        source=(ROOT/'.github/workflows/go-native.yml').read_text()
        expressions=[line.split('${{',1)[1].rsplit('}}',1)[0].strip()
                     for line in source.splitlines() if 'ORBIT_CAPTURE_CHAT_STAGE1:' in line]
        self.assertEqual(len(expressions),2)
        self.assertEqual(expressions[0],expressions[1])
        self.assertIn(expressions[0],(MANUAL_EXPRESSION,ONE_SHOT_EXPRESSION))
        parent='a5c40cc2f85b57b94c1e2027d5cfc690b8692a49'
        title='Capture static Chat stage 1 after verified threaded ownership'
        cases=[
            ('planned push','push',False,parent,title,'1' if expressions[0]==ONE_SHOT_EXPRESSION else '0'),
            ('normal push','push',False,'another-parent','Ordinary change','0'),
            ('expired first marker','push',False,'00e685d1c0b13eebba068ab5119be0ef258ca59c','Capture static Chat stage 1 once (2026-10-25 verified wave4)','0'),
            ('wrong parent','push',False,'another-parent',title,'0'),
            ('wrong title','push',False,parent,'Ordinary change','0'),
            ('title suffix','push',False,parent,title+' extra','0'),
            ('manual false','workflow_dispatch',False,parent,title,'0'),
            ('manual true','workflow_dispatch',True,'','', '1'),
            ('pull request','pull_request',True,parent,title,'0'),
            ('push ignores manual input','push',True,parent,'Ordinary change','0'),
        ]
        for label,event,manual,before,message,want in cases:
            with self.subTest(label=label):
                # Evaluate the actual YAML predicate, not a copied Python rule.
                # Vocabulary is fixed above by the exact-expression assertion.
                expression=expressions[0]
                values={'github.event_name':event,'inputs.capture_chat_stage1':manual,
                        'github.event.before':before,'github.event.head_commit.message':message}
                for name,value in values.items(): expression=expression.replace(name,repr(value))
                expression=expression.replace('&&',' and ').replace('||',' or ')
                self.assertEqual(eval(expression,{'__builtins__':{}},{}),want)
