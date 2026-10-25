import base64
import hashlib
import io
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest.mock import patch
import vendor_reference as v

class VendorReference(unittest.TestCase):
    def test_corrupt_archive_rejected_before_unpack_or_execution(self):
        with tempfile.TemporaryDirectory() as d:
            p=Path(d)/'bad.tgz';p.write_bytes(b'not pinned vendor')
            with patch.object(v.tarfile,'open') as unpack,patch.object(v.subprocess,'run') as run:
                with self.assertRaisesRegex(ValueError,'integrity'):v.build(p,p,Path(d))
                unpack.assert_not_called();run.assert_not_called()
    def test_escape_and_link_members_refused_even_after_integrity_gate(self):
        # Test-only fixed integrity for synthetic tar; production constant is not
        # configurable via CLI/environment or any child receipt.
        for name,kind in [('../escape',tarfile.REGTYPE),('/absolute',tarfile.REGTYPE),('package/link',tarfile.SYMTYPE),('package/hard',tarfile.LNKTYPE)]:
            with self.subTest(name=name),tempfile.TemporaryDirectory() as d:
                buf=io.BytesIO()
                with tarfile.open(fileobj=buf,mode='w:gz') as tar:
                    member=tarfile.TarInfo(name);member.type=kind;member.linkname='/outside';tar.addfile(member,io.BytesIO(b''))
                raw=buf.getvalue();p=Path(d)/'fixture.tgz';p.write_bytes(raw)
                integrity='sha512-'+base64.b64encode(hashlib.sha512(raw).digest()).decode()
                with patch.object(v,'VENDOR_INTEGRITY',integrity),patch.object(v.subprocess,'check_output',return_value='v22.23.4'),patch.object(v.subprocess,'run') as run:
                    with self.assertRaises(ValueError):v.build(p,p,Path(d))
                    run.assert_not_called()
