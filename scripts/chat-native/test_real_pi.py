import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import materialize
import preflight

class ActualPiMetadata(unittest.TestCase):
    @unittest.skipUnless(os.getenv('ORBIT_PI_INTEGRATION_BIN'),'explicit pinned real Pi opt-in required')
    def test_persisted_session_stats_without_prompt(self):
        with tempfile.TemporaryDirectory()as d:
            p=Path(d)
            with patch.object(materialize,'WORK',p/'fixture/workspace'):
                m=materialize.materialize(p/'fixture',p/'fixture.json')
                result=preflight.probe(os.environ['ORBIT_PI_INTEGRATION_BIN'],p/'fixture',p/'preflight.json')
            self.assertEqual([x['command']['type']for x in result['commands']],['switch_session','get_state','get_messages','get_session_stats'])
            with patch.object(materialize,'WORK',p/'fixture/workspace'):
                self.assertTrue(materialize.verify(p/'fixture',m)['integrityAccepted'])
if __name__=='__main__':unittest.main()
