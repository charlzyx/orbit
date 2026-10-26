"""Offline only: subprocess.run/Popen are mocked, never native execution."""
import copy
import base64
import json
import tempfile
import struct
import zlib
import math
import os
import stat
import subprocess
import time
import unittest
from dataclasses import replace
from types import SimpleNamespace
from unittest.mock import MagicMock, Mock, patch
from contextlib import ExitStack

import window_capture as wc


class WindowCaptureTests(unittest.TestCase):
    def setUp(self):
        self.identity = wc.Identity("a"*64, 42, 100, "/private/tmp/Own.app/Contents/MacOS/Own", 9)
        self.receipt = dict(nonce="a"*64, pid=42, start_us=100, window_number=9,
            visible=True, bounds=dict(x=10,y=20,w=300,h=200), phase="seed", at_unix_ns=200000)
        self.proof = dict(pid=42,start_us=100,executable=self.identity.executable,
            window_number=9,onscreen=True,preflight=True,bounds=dict(x=10,y=900,w=300,h=200))

    def test_coordinate_boundary_no_cross_source_y_comparison(self):
        wc.validate_receipt(self.receipt,self.identity,"seed",300000)
        wc.validate_proof(self.proof,self.identity)
        wc.compare_pair(self.receipt,self.proof)
        wc.validate_continuity(self.receipt,self.receipt,self.proof,self.proof)

    def test_owner_nonce_phase_and_stale_receipts_refused(self):
        for key,value in (("pid",43),("start_us",101),("window_number",10),("nonce","b"*64),
                          ("phase","final"),("visible",False),("pid",True),("at_unix_ns",99999),
                          ("at_unix_ns",300001)):
            r=dict(self.receipt);r[key]=value
            with self.subTest(key=key),self.assertRaises(wc.Blocked): wc.validate_receipt(r,self.identity,"seed",300000)
        with self.assertRaises(wc.Blocked): wc.validate_receipt(self.receipt,self.identity,"seed",6_000_000_000)
        with self.assertRaises(wc.Blocked): wc.validate_receipt(self.receipt,self.identity,"seed",300000,earliest_ns=200000)

    def test_independent_proof_wrong_process_window_permission_refused(self):
        for key,value in (("pid",43),("start_us",101),("executable","/other"),("window_number",10),
                          ("onscreen",False),("preflight",False),("onscreen",1)):
            proof=dict(self.proof);proof[key]=value
            with self.subTest(key=key),self.assertRaises(wc.Blocked): wc.validate_proof(proof,self.identity)

    def test_nonfinite_zero_negative_or_excessive_bounds(self):
        for key,value in (("x",float("nan")),("y",float("inf")),("w",0),("h",-1),("w",4097),("x",True)):
            b=dict(self.proof["bounds"]);b[key]=value
            with self.subTest(key=key),self.assertRaises(wc.Blocked): wc.bounds(b)
        with self.assertRaises(wc.Blocked): wc.bounds(dict(x=0,y=0,w=1))

    def test_changed_each_source_or_size_refused(self):
        after=copy.deepcopy(self.proof);after["bounds"]["y"]+=1
        with self.assertRaises(wc.Blocked): wc.validate_continuity(self.receipt,self.receipt,self.proof,after)
        after=copy.deepcopy(self.receipt);after["bounds"]["x"]+=1
        with self.assertRaises(wc.Blocked): wc.validate_continuity(self.receipt,after,self.proof,self.proof)
        proof=copy.deepcopy(self.proof);proof["bounds"]["w"]+=1
        with self.assertRaises(wc.Blocked): wc.compare_pair(self.receipt,proof)

    def test_output_fixed_private_path_and_existing_refusal(self):
        st=SimpleNamespace(st_mode=stat.S_IFDIR|0o700,st_uid=os.getuid())
        with patch("window_capture.os.path.realpath",side_effect=lambda p:p),patch("window_capture.os.stat",return_value=st),patch("window_capture.os.path.lexists",return_value=False):
            self.assertEqual(wc.private_output(self.identity,"seed"),"/private/tmp/orbit-composer-"+"a"*64+"/seed-os.png")
            for phase in ("../other","screen","", "seed.png"):
                with self.assertRaises(wc.Blocked): wc.private_output(self.identity,phase)
        with patch("window_capture.os.path.realpath",return_value="/elsewhere"):
            with self.assertRaises(wc.Blocked): wc.private_output(self.identity,"seed")
        with patch("window_capture.os.path.realpath",side_effect=lambda p:p),patch("window_capture.os.stat",return_value=st),patch("window_capture.os.path.lexists",return_value=True):
            with self.assertRaises(wc.Blocked): wc.private_output(self.identity,"seed")
        for mode,uid in ((0o755,os.getuid()),(0o700,os.getuid()+1)):
            with patch("window_capture.os.path.realpath",side_effect=lambda p:p),patch("window_capture.os.stat",return_value=SimpleNamespace(st_mode=stat.S_IFDIR|mode,st_uid=uid)):
                with self.assertRaises(wc.Blocked): wc.private_output(self.identity,"seed")

    def test_no_native_calls_without_admission(self):
        with patch("window_capture.subprocess.run") as run,patch("window_capture.subprocess.Popen") as popen:
            with self.assertRaises(wc.Blocked): wc.capture_owned_window(identity=self.identity,phase="seed",verifier="/proof",own_receipt=Mock(),deadline=time.monotonic()+5)
            run.assert_not_called();popen.assert_not_called()

    def mock_capture(self, stack, *, command_status=0, proof_error=None):
        stack.enter_context(patch("window_capture.platform.system",return_value="Darwin"))
        stack.enter_context(patch("window_capture.platform.machine",return_value="arm64"))
        stack.enter_context(patch.dict("os.environ",{"ORBIT_COMPOSER_DISPOSABLE_RUNNER":"1"}))
        stack.enter_context(patch("window_capture._local_executable",return_value="/proof"))
        path="/private/tmp/orbit-composer-"+"a"*64+"/seed-os.png"
        stack.enter_context(patch("window_capture.private_output",return_value=path))
        stack.enter_context(patch("window_capture.time.time_ns",side_effect=[300000,350000,400000,450000,600000]))
        proof=stack.enter_context(patch("window_capture._window_proof",side_effect=proof_error,return_value=self.proof))
        stack.enter_context(patch("window_capture._png",return_value=dict(sha256="f"*64,bytes=100,width=600,height=400)))
        run=stack.enter_context(patch("window_capture.subprocess.run",return_value=SimpleNamespace(returncode=command_status)))
        popen=stack.enter_context(patch("window_capture.subprocess.Popen",side_effect=AssertionError("unmocked native call")))
        receipts=Mock(side_effect=[self.receipt,dict(self.receipt,at_unix_ns=500000)])
        return run,proof,receipts,path

    def test_mocked_exact_single_window_command_before_after_proof(self):
        with ExitStack() as stack:
            run,proof,receipts,path=self.mock_capture(stack)
            result=wc.capture_owned_window(identity=self.identity,phase="seed",verifier="/proof",own_receipt=receipts,deadline=time.monotonic()+5,execute_native=True)
            self.assertEqual(result["status"],"OS_WINDOW_CAPTURE_CANDIDATE")
            self.assertEqual(result["integration"],"phase-handshake-candidate")
            self.assertEqual(run.call_count,1);self.assertEqual(proof.call_count,2)
            self.assertEqual(run.call_args.args[0],["/usr/sbin/screencapture","-x","-l","9",path])
            self.assertEqual([call.args[0] for call in receipts.call_args_list],["before","after"])

    def test_permission_or_proof_failure_never_capture(self):
        with ExitStack() as stack:
            run,proof,receipts,path=self.mock_capture(stack,proof_error=wc.Blocked("preflight false"))
            with self.assertRaises(wc.Blocked): wc.capture_owned_window(identity=self.identity,phase="seed",verifier="/proof",own_receipt=receipts,deadline=time.monotonic()+5,execute_native=True)
            run.assert_not_called()

    def test_command_failure_no_retry_no_fallback(self):
        with ExitStack() as stack:
            run,proof,receipts,path=self.mock_capture(stack,command_status=1)
            with self.assertRaises(wc.Blocked): wc.capture_owned_window(identity=self.identity,phase="seed",verifier="/proof",own_receipt=receipts,deadline=time.monotonic()+5,execute_native=True)
            self.assertEqual(run.call_count,1);self.assertEqual(proof.call_count,1)

    def test_expired_deadline_never_capture(self):
        with ExitStack() as stack:
            run,proof,receipts,path=self.mock_capture(stack)
            with self.assertRaises(wc.Blocked): wc.capture_owned_window(identity=self.identity,phase="seed",verifier="/proof",own_receipt=receipts,deadline=time.monotonic()-1,execute_native=True)
            run.assert_not_called();proof.assert_not_called()

    def test_external_verifier_exact_command_bounded_output(self):
        process=Mock(stdout=Mock());process.stdout.fileno.return_value=73
        process.wait.return_value=0;process.poll.return_value=0
        selector=MagicMock();selector.__enter__.return_value=selector;selector.select.return_value=[True]
        with patch("window_capture.subprocess.Popen",return_value=process) as popen,patch("window_capture.selectors.DefaultSelector",return_value=selector),patch("window_capture.os.read",side_effect=[json.dumps(self.proof).encode(),b""]):
            self.assertEqual(wc._window_proof("/proof",self.identity,time.monotonic()+5),self.proof)
            self.assertEqual(popen.call_args.args[0],["/proof","window-proof","42","100","9"])
        with patch("window_capture.subprocess.Popen",return_value=process),patch("window_capture.selectors.DefaultSelector",return_value=selector),patch("window_capture.os.read",return_value=b"x"*(wc.MAX_JSON+1)):
            with self.assertRaises(wc.Blocked):wc._window_proof("/proof",self.identity,time.monotonic()+5)

    def test_empty_symlink_and_non_png_artifacts_refused(self):
        png=base64.b64decode("iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAEklEQVR42mP4z8DwHwyBNBgAAEnICfcD2WTxAAAAAElFTkSuQmCC")
        with tempfile.TemporaryDirectory(prefix="window-capture-contract-") as directory:
            path=directory+"/seed-os.png"
            for raw in (b"",b"x"*75):
                with open(path,"wb") as stream:stream.write(raw)
                with self.assertRaises(wc.Blocked):wc._png(path)
            with open(path,"wb") as stream:stream.write(png)
            self.assertEqual(wc._png(path)["bytes"],len(png))
            link=directory+"/link";os.symlink(path,link)
            with self.assertRaises(OSError):wc._png(link)

    def complete_png_fixture(self):
        return base64.b64decode("iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAEklEQVR42mP4z8DwHwyBNBgAAEnICfcD2WTxAAAAAElFTkSuQmCC")

    def png_chunk(self, kind, payload=b""):
        return struct.pack(">I",len(payload))+kind+payload+struct.pack(">I",zlib.crc32(kind+payload))

    def assert_png_file(self, raw, *, accepted=False):
        # Real isolated file, real _png: only native process calls remain mocked.
        with tempfile.TemporaryDirectory(prefix="window-png-integrity-") as directory:
            path=directory+"/seed-os.png"
            with open(path,"wb") as stream:stream.write(raw)
            if accepted:
                result=wc._png(path)
                self.assertEqual((result["width"],result["height"],result["bytes"]),(2,2,len(raw)))
            else:
                with self.assertRaises(wc.Blocked):wc._png(path)

    def test_png_complete_real_file_accepted(self):
        self.assert_png_file(self.complete_png_fixture(),accepted=True)

    def test_png_header_only_real_file_refused(self):
        self.assert_png_file(self.complete_png_fixture()[:33])

    def test_png_truncated_real_file_refused(self):
        png=self.complete_png_fixture()
        for raw in (png[:-1],png[:45],png[:33]+b"\0\0\0"):
            self.assert_png_file(raw)

    def test_png_bad_crc_real_file_refused(self):
        png=self.complete_png_fixture()
        for index in (29,41,len(png)-1):
            raw=bytearray(png);raw[index]^=1
            self.assert_png_file(bytes(raw))

    def test_png_missing_or_empty_idat_real_file_refused(self):
        header=self.complete_png_fixture()[:33]
        self.assert_png_file(header+self.png_chunk(b"IEND"))
        self.assert_png_file(header+self.png_chunk(b"IDAT")+self.png_chunk(b"IEND"))

    def test_png_missing_nonterminal_or_nonempty_iend_real_file_refused(self):
        png=self.complete_png_fixture()
        self.assert_png_file(png[:-12])
        self.assert_png_file(png+b"extra")
        self.assert_png_file(png[:-12]+self.png_chunk(b"IEND",b"x"))

    def test_png_chunk_limit_real_file_refused(self):
        png=self.complete_png_fixture()
        raw=png[:-12]+self.png_chunk(b"tEXt",b"x")*8192+png[-12:]
        self.assert_png_file(raw)

    def test_png_good_crc_invalid_compression_real_file_refused(self):
        header=self.complete_png_fixture()[:33]
        self.assert_png_file(header+self.png_chunk(b"IDAT",b"not-zlib")+self.png_chunk(b"IEND"))

    def test_png_good_crc_truncated_pixel_stream_real_file_refused(self):
        header=self.complete_png_fixture()[:33]
        # Complete zlib stream and valid chunk CRC, but insufficient 2x2 RGBA rows.
        self.assert_png_file(header+self.png_chunk(b"IDAT",zlib.compress(b"\0"+b"\xff"*8))+self.png_chunk(b"IEND"))

    def test_png_good_crc_invalid_filter_real_file_refused(self):
        header=self.complete_png_fixture()[:33]
        scanlines=b"\x05"+b"\xff"*8+b"\0"+b"\xff"*8
        self.assert_png_file(header+self.png_chunk(b"IDAT",zlib.compress(scanlines))+self.png_chunk(b"IEND"))

    def test_png_decoder_unavailable_or_truncated_mode_blocked(self):
        import builtins
        original=builtins.__import__
        def no_pillow(name,*args,**kwargs):
            if name=="PIL" or name.startswith("PIL."):raise ImportError("fixture missing Pillow")
            return original(name,*args,**kwargs)
        with patch("builtins.__import__",side_effect=no_pillow):
            self.assert_png_file(self.complete_png_fixture())
        with patch("PIL.ImageFile.LOAD_TRUNCATED_IMAGES",True):
            self.assert_png_file(self.complete_png_fixture())

    def test_decoder_pixel_budget_checked_before_import(self):
        with patch("builtins.__import__",side_effect=AssertionError("must not import decoder")):
            with self.assertRaises(wc.Blocked):wc._verify_png_decode(b"",4097,4096)

    def test_direct_child_terminated_and_reaped_after_failure(self):
        process=Mock(pid=123,stdout=Mock());process.stdout.fileno.return_value=73
        process.poll.return_value=None;process.wait.return_value=-15
        selector=MagicMock();selector.__enter__.return_value=selector;selector.select.return_value=[True]
        with patch("window_capture.subprocess.Popen",return_value=process),patch("window_capture.selectors.DefaultSelector",return_value=selector),patch("window_capture.os.read",return_value=b"x"*(wc.MAX_JSON+1)):
            with self.assertRaises(wc.Blocked):wc._window_proof("/proof",self.identity,time.monotonic()+5)
        process.terminate.assert_called_once();process.wait.assert_called_once_with(timeout=2)
        process.stdout.close.assert_called_once()

    def test_direct_child_unreaped_is_blocked_with_owned_pid_no_sigkill(self):
        process=Mock(pid=123,stdout=Mock());process.stdout.fileno.return_value=73
        process.poll.return_value=None;process.wait.side_effect=subprocess.TimeoutExpired("fixture",2)
        selector=MagicMock();selector.__enter__.return_value=selector;selector.select.return_value=[True]
        with patch("window_capture.subprocess.Popen",return_value=process),patch("window_capture.selectors.DefaultSelector",return_value=selector),patch("window_capture.os.read",return_value=b"x"*(wc.MAX_JSON+1)),patch("window_capture.os.kill") as kill:
            with self.assertRaisesRegex(wc.Blocked,"child-not-reaped pid=123"):wc._window_proof("/proof",self.identity,time.monotonic()+5)
            kill.assert_not_called()
        process.kill.assert_not_called();process.terminate.assert_called_once()
        process.wait.assert_called_once_with(timeout=2);process.stdout.close.assert_called_once()

    def png_fixture_format(self, depth, color, scanlines, *, interlace=0, palette=None, compressed=None):
        header=struct.pack(">IIBBBBB",2,2,depth,color,0,0,interlace)
        raw=b"\x89PNG\r\n\x1a\n"+self.png_chunk(b"IHDR",header)
        if palette is not None:raw+=self.png_chunk(b"PLTE",palette)
        return raw+self.png_chunk(b"IDAT",zlib.compress(scanlines) if compressed is None else compressed)+self.png_chunk(b"IEND")

    def test_valid_palette_lowbit_and_16bit_real_files(self):
        # 1-bit indexed palette, 1/2/4-bit grayscale, and 16-bit grayscale/RGB/RGBA.
        self.assert_png_file(self.png_fixture_format(1,3,b"\0\x40\0\x80",palette=b"\0\0\0\xff\xff\xff"),accepted=True)
        for depth in (1,2,4):
            self.assert_png_file(self.png_fixture_format(depth,0,b"\0\0\0\0"),accepted=True)
        for color,channels in ((0,1),(2,3),(4,2),(6,4)):
            scanlines=(b"\0"+bytes(2*channels*2))*2
            self.assert_png_file(self.png_fixture_format(16,color,scanlines),accepted=True)

    def test_adam7_explicitly_unsupported_real_file(self):
        self.assert_png_file(self.png_fixture_format(8,6,b"\0"+bytes(17),interlace=1))

    def test_illegal_depth_color_real_file_refused(self):
        for depth,color in ((4,2),(16,3),(1,6),(8,1),(8,5)):
            self.assert_png_file(self.png_fixture_format(depth,color,bytes(18)))

    def test_zlib_excess_extra_stream_and_truncated_real_files(self):
        scanlines=(b"\0"+bytes(8))*2
        compressed=zlib.compress(scanlines)
        for payload in (zlib.compress(scanlines+b"x"),compressed+b"extra",compressed+compressed,compressed[:-1]):
            self.assert_png_file(self.png_fixture_format(8,6,scanlines,compressed=payload))

    def test_inflated_output_budget_before_decompression(self):
        with patch("window_capture.zlib.decompressobj",side_effect=AssertionError("no decompression")):
            with self.assertRaises(wc.Blocked):wc._validate_idat_length([b""],4097,4096,bytes((16,6,0,0,0)))
        # Valid CRC and a compact compression bomb targeting a tiny raster.
        self.assert_png_file(self.png_fixture_format(8,6,b"",compressed=zlib.compress(bytes(100000))))

    def test_bounded_inflate_loop_across_output_and_idat_boundaries(self):
        scanlines=(b"\0"+bytes(256))*256  # larger than the 64 KiB transient limit
        compressed=zlib.compress(scanlines)
        wc._validate_idat_length([compressed],256,256,bytes((8,0,0,0,0)))
        wc._validate_idat_length([compressed[i:i+3] for i in range(0,len(compressed),3)],256,256,bytes((8,0,0,0,0)))

    def test_json_empty_duplicate_malformed_nonfinite_refused(self):
        for raw in (b"",b"{}"*(wc.MAX_JSON),b'{"pid":42,"pid":43}',b'{"x":NaN}',b"[]",b"bad"):
            with self.assertRaises(wc.Blocked): wc._json(raw)


if __name__ == "__main__": unittest.main()
