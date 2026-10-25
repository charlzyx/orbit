import copy
import importlib.util
import json
from pathlib import Path
import tempfile
import time
import unittest
from unittest.mock import patch

HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('chat_capture',HERE/'capture.py');c=importlib.util.module_from_spec(spec);spec.loader.exec_module(c)
SUITE=json.loads((HERE.parent.parent/'testdata/chat-native/cases.json').read_text())
IDENTITY={'pid':123,'startTicks':'1234','exe':'/selected/binary'}

def instance(out):
    x=object.__new__(c.Capture);c.base.Capture.__init__(x,'42',123,out,copy.deepcopy(SUITE));x.identity=IDENTITY;return x

def run(*args):
    if args[1]=='getwindowpid':return '123'
    if args[1]=='search':return '42'
    if args[1]=='getactivewindow':return '42'
    if args[1]=='getwindowgeometry':return 'WIDTH=1470\nHEIGHT=922'
    return ''

class CaptureGuards(unittest.TestCase):
    def test_recipe_allows_only_two_approved_inputs(self):
        c.validate_suite(SUITE)
        for action in ({'key':'Return'},{'key':'ctrl+c'},{'clickExactLabel':'Compact','region':[1,2,3,4]},{'hover':'40%'}):
            suite=copy.deepcopy(SUITE);suite['cases'][0]['actions']=[action]
            with self.assertRaises(ValueError):c.validate_suite(suite)
    def test_wrong_pid_hidden_changed_process_and_inactive_block(self):
        with tempfile.TemporaryDirectory()as d:
            cap=instance(d)
            for case in ('pid','hidden','process','inactive','geometry'):
                def fake(*args):
                    if case=='pid'and args[1]=='getwindowpid':return '999'
                    if case=='hidden'and args[1]=='search':return ''
                    if case=='inactive'and args[1]=='getactivewindow':return '999'
                    if case=='geometry'and args[1]=='getwindowgeometry':return 'WIDTH=1469\nHEIGHT=922'
                    return run(*args)
                identity=dict(IDENTITY,startTicks='changed')if case=='process'else IDENTITY
                with patch.object(cap,'alive'),patch.object(c,'process_identity',return_value=identity),patch.object(c.base,'run',side_effect=fake)as calls:
                    with self.assertRaises(ValueError):cap.guard(active=True)
                    self.assertFalse(any('click'in call.args or 'key'in call.args for call in calls.call_args_list))
    def test_stale_foreign_or_changed_image_refuses_input(self):
        with tempfile.TemporaryDirectory()as d:
            cap=instance(d);image=Path(d)/'fresh.png';image.write_bytes(b'new')
            frame=dict(processIdentity=IDENTITY,capturedMonotonic=time.monotonic(),pngSha256=c.base.sha(image),rpcChildren=[{'identity':{'pid':456}}])
            cap.pi_entry=Path('/pinned/pi')
            with patch.object(cap,'guard'),patch.object(c,'owned_rpc_children',return_value=[{'identity':{'pid':456}}]):
                cap.validate_input_proof(image,frame)
                for bad in (dict(frame,capturedMonotonic=time.monotonic()-6),dict(frame,processIdentity={}),dict(frame,pngSha256='old')):
                    with self.assertRaises(ValueError):cap.validate_input_proof(image,bad)
    def test_missing_or_ambiguous_context_label_has_no_click(self):
        with tempfile.TemporaryDirectory()as d:
            cap=instance(d);row=dict(text='40%',left='1300',top='870',width='25',height='18',conf='97')
            for rows in ([],[row,row]):
                with patch.object(c.base,'run')as calls:
                    with self.assertRaises(ValueError):cap.target(Path(d)/'frame.png',rows,{},'40%',[1000,700,1470,922])
                    calls.assert_not_called()
    def test_key_action_requires_fresh_correct_state_and_logs_before_input(self):
        with tempfile.TemporaryDirectory()as d:
            cap=instance(d);frame=dict(processIdentity=IDENTITY,capturedMonotonic=time.monotonic(),pngSha256='fresh')
            with patch.object(cap,'guard'),patch.object(cap,'capture',return_value=([],frame)),patch.object(cap,'readiness',return_value=(False,{'missing':['tail']},frame)),patch.object(c.base,'run')as calls:
                with self.assertRaises(ValueError):cap.action('CHAT02',1,{'key':'ctrl+Up'})
                self.assertFalse(any('key'in call.args or 'click'in call.args for call in calls.call_args_list))
            with patch.object(cap,'guard'),patch.object(cap,'capture',return_value=([],frame)),patch.object(cap,'readiness',return_value=(True,{},frame)),patch.object(cap,'validate_input_proof'),patch.object(c.base,'run')as calls,patch.object(c.time,'sleep'):
                cap.action('CHAT02',1,{'key':'ctrl+Up'})
                self.assertIn(('xdotool','key','--clearmodifiers','ctrl+Up'),[x.args for x in calls.call_args_list])
                self.assertEqual(cap.actions[0]['outcome'],'INPUT_REQUESTED');self.assertEqual(cap.actions[1]['outcome'],'INPUT_SENT_STATE_NOT_YET_VERIFIED')

class AdditionalGuards(unittest.TestCase):
    def test_missing_runtime_descendant_is_not_readiness(self):
        with tempfile.TemporaryDirectory()as d:
            cap=instance(d)
            good,detail,_=cap.readiness(Path(d)/'frame.png',[],{'rpcChildren':[]},['CHAT TAIL READY'],1)
            self.assertFalse(good);self.assertIn('live pinned Pi RPC descendant',detail['missing'])
    def test_uncertain_key_result_is_recorded_and_not_retried(self):
        with tempfile.TemporaryDirectory()as d:
            cap=instance(d);frame=dict(processIdentity=IDENTITY,capturedMonotonic=time.monotonic(),pngSha256='fresh')
            sent=[]
            def tool(*args):
                if 'key'in args:sent.append(args);raise OSError('uncertain input transport')
                return ''
            with patch.object(cap,'guard'),patch.object(cap,'capture',return_value=([],frame)),patch.object(cap,'readiness',return_value=(True,{},frame)),patch.object(cap,'validate_input_proof'),patch.object(c.base,'run',side_effect=tool):
                with self.assertRaises(OSError):cap.action('CHAT02',1,{'key':'ctrl+Up'})
            self.assertEqual(len(sent),1);self.assertEqual(cap.actions[-1]['outcome'],'INPUT_RESULT_UNCERTAIN_NO_RETRY')
    def test_existing_capture_cannot_be_reused(self):
        with tempfile.TemporaryDirectory()as d:
            Path(d,'capture-result.json').write_text('prior proof')
            with patch.object(c,'process_identity')as identity:
                with self.assertRaises(ValueError):c.Capture('42',123,d,SUITE,'/binary','/pi')
                identity.assert_not_called()
            self.assertEqual(Path(d,'capture-result.json').read_text(),'prior proof')
    def test_result_explicitly_keeps_context_blocked(self):
        with tempfile.TemporaryDirectory()as d:
            cap=instance(d);cap.write_result()
            result=json.loads(Path(d,'capture-result.json').read_text())
            self.assertEqual(result['requestedCases'],2);self.assertEqual(result['capturedCases'],0)
            self.assertEqual(result['scopeStatus'],'PARTIAL_C15_BLOCKED');self.assertEqual(result['blockedCases'][0]['id'],'CHAT03')

if __name__=='__main__':unittest.main()

class ProductionMetadataChain(unittest.TestCase):
    def test_actual_capture_readiness_settle_execute_metadata_shape(self):
        # Do NOT stub Capture.capture/readiness/settle or manufacture metadata.
        # Only the external desktop/OCR boundary is substituted in this unit test.
        with tempfile.TemporaryDirectory() as d:
            cap=instance(d);cap.pi_entry=Path('/pinned/pi');state=['Static Chat fixture']
            def desktop(*args):
                if args[0]=='import':Path(args[-1]).write_bytes(('whole-frame:'+state[0]).encode());return ''
                if args[0]=='identify':return '1470 922'
                if args[0]=='tesseract':
                    return 'level\tleft\ttop\twidth\theight\tconf\ttext\n'+''.join(f'5\t10\t10\t40\t20\t99\t{word}\n'for word in state[0].split())
                if args[0]=='xdotool' and args[1]=='key':
                    state[0]='CHAT TAIL READY' if args[-1]=='ctrl+1' else 'PROMPT ELEVEN RESPONSE NUMBER 11'
                return run(*args)
            with patch.object(cap,'alive'),patch.object(c,'process_identity',return_value=IDENTITY),patch.object(c,'owned_rpc_children',return_value=[{'identity':{'pid':456}}]),patch.object(c.base,'run',side_effect=desktop),patch.object(c.time,'sleep'):
                cap.settle(dict(id='STARTUP',image='evidence/startup.png',require=['Static Chat fixture']))
                cap.execute()
            result=json.loads(Path(d,'capture-result.json').read_text())
            self.assertEqual(result['capturedCases'],2)
            for case in result['cases']:
                proof=case['proof'];self.assertEqual(proof['image'],[1470,922]);self.assertEqual(proof['processIdentity'],IDENTITY)
                self.assertEqual(proof['consecutiveReadySamples'],3);self.assertEqual(proof['selectedReadinessOcrPSM'],11)
                self.assertEqual(proof['pngSha256'],c.base.sha(Path(d)/case['image']))
            self.assertEqual([a['requested']for a in cap.actions if a['outcome']=='INPUT_REQUESTED'],[{'key':'ctrl+1'},{'key':'ctrl+Up'}])


class InputDescendantRefresh(unittest.TestCase):
    def test_exited_or_replaced_child_blocks_input_even_with_fresh_image(self):
        with tempfile.TemporaryDirectory() as d:
            cap=instance(d);cap.pi_entry=Path('/pinned/pi');image=Path(d)/'proof.png';image.write_bytes(b'frame')
            frame=dict(processIdentity=IDENTITY,capturedMonotonic=time.monotonic(),pngSha256=c.base.sha(image),rpcChildren=[{'identity':{'pid':456,'startTicks':'old'}}])
            for children in ([],[{'identity':{'pid':456,'startTicks':'new'}}]):
                with patch.object(cap,'guard'),patch.object(c,'owned_rpc_children',return_value=children):
                    with self.assertRaisesRegex(ValueError,'descendant changed or exited'):cap.validate_input_proof(image,frame)
