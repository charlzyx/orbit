import types
import unittest
from unittest.mock import patch
import runner
import window_capture


class PhaseHandshakeTests(unittest.TestCase):
    def run_case(self, mode="png-text", mutate=None, fail=None):
        owner = types.SimpleNamespace(pid=42, start_us=123, executable="/tmp/test.app/Contents/MacOS/Orbit")
        writes=[]; phases=[]
        def wait(path, deadline, launcher):
            phase=path.rsplit("/",1)[1].split(".os-")[0]
            value=dict(nonce="a"*64,pid=42,start_us=123,window_number=88,phase=phase)
            if mutate: mutate(value)
            return value,b"",0
        def capture(**kw):
            phase=kw['phase'];phases.append(phase)
            if fail==phase: raise window_capture.Blocked('mock failure')
            kw['own_receipt']('before',phase,100)
            kw['own_receipt']('after',phase,100)
            return dict(status="OS_WINDOW_CAPTURE_CANDIDATE",phase=phase)
        with patch.object(runner,'_wait_artifact',side_effect=wait),patch.object(runner,'_publish_json',side_effect=lambda path,value:writes.append((path,value))),patch.object(window_capture,'capture_owned_window',side_effect=capture),patch.object(runner.time,'monotonic',return_value=0):
            try: result=runner._capture_os_phases(mode,"/private/tmp/orbit-composer-"+"a"*64,"a"*64,owner,"/tmp/verifier",100,None)
            except Exception as e: return e,writes,phases
        return result,writes,phases

    def test_exact_phases_requests_proofs_ack(self):
        for mode in runner.CASES:
            with self.subTest(mode=mode):
                result,writes,phases=self.run_case(mode)
                self.assertEqual(result,runner._os_phases(mode));self.assertEqual(phases,result)
                for i,p in enumerate(result):
                    self.assertEqual([x[0].split(p+'.os-')[-1] for x in writes[i*4:i*4+4]],["before-request.json","after-request.json","proof.json","ack.json"])
                    self.assertEqual(writes[i*4+3][1]['status'],'PASS')
    def test_foreign_or_wrong_phase_no_capture_or_ack(self):
        for field,value in [('pid',43),('start_us',124),('nonce','b'*64),('phase','wrong'),('window_number',0),('window_number',2**32)]:
            with self.subTest(field=field,value=value):
                result,writes,phases=self.run_case(mutate=lambda d:d.update({field:value}))
                self.assertIsInstance(result,runner.Blocked);self.assertEqual(writes,[]);self.assertEqual(phases,[])
    def test_capture_failure_no_ack_no_retry_or_later_phase(self):
        result,writes,phases=self.run_case(fail='seed')
        self.assertIsInstance(result,runner.Blocked);self.assertEqual(phases,['before','seed']);self.assertEqual(len(writes),4)
    def test_unknown_mode_refused(self):
        with self.assertRaises(runner.Blocked):runner._os_phases('other')
    def test_missing_ready_stops_without_input_or_ack(self):
        with patch.object(runner,'_wait_artifact',side_effect=runner.Blocked('missing')),patch.object(runner,'_publish_json') as publish:
            with self.assertRaises(runner.Blocked):runner._capture_os_phases('png-text','/tmp/x','a'*64,None,'/tmp/v',100,None)
            publish.assert_not_called()

    def test_window_number_changed_between_phases_no_later_capture(self):
        calls=[0]
        def change(v):
            if v['phase']=='seed':v['window_number']=89
        result,writes,phases=self.run_case(mutate=change)
        self.assertIsInstance(result,runner.Blocked);self.assertEqual(phases,['before']);self.assertEqual(len(writes),4)
