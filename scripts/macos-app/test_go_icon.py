import ctypes.util, hashlib, importlib.util, json, shutil, struct, tempfile, unittest
from pathlib import Path
import bundle
ROOT=Path(__file__).resolve().parents[2]
class GoIconTests(unittest.TestCase):
 def test_generated_assets_source_and_all_icns_sizes(self):
  m=bundle.verify_go_icon(ROOT)
  self.assertEqual(len(m['icns_slots']),11)
  self.assertNotEqual(bundle.digest(ROOT/'assets/icons/go/orbit-go.icns'),bundle.digest(ROOT/'assets/icons/alpha-logo.icns'))
 @unittest.skipUnless(all(ctypes.util.find_library(x) for x in ("rsvg-2","cairo","gobject-2.0")), "Optional regeneration requires librsvg/cairo; structural/hash verification remains mandatory")
 def test_rebuild_all_sizes_is_byte_reproducible(self):
  spec=importlib.util.spec_from_file_location('goicon',ROOT/'scripts/make-go-icon.py');mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
  with tempfile.TemporaryDirectory() as d:
   r=Path(d);mod.ICNS=r/'icon.icns';mod.APP=r/'app.png';mod.MANIFEST=r/'manifest.json'
   # Render all slots without altering checked-in files.
   images={}
   for size in sorted(set(mod.SLOTS.values())):
    out=r/f'{size}.png';mod.render(mod.SOURCE,out,size);images[size]=out.read_bytes()
   self.assertEqual(images[512],(ROOT/'crates/orbit-pi/assets/go-app-icon.png').read_bytes())
   chunks=b''.join(kind.encode()+struct.pack('>I',8+len(images[size]))+images[size] for kind,size in mod.SLOTS.items())
   rebuilt=b'icns'+struct.pack('>I',8+len(chunks))+chunks
   self.assertEqual(rebuilt,(ROOT/'assets/icons/go/orbit-go.icns').read_bytes())
 def test_tamper_or_missing_slot_rejected(self):
  with tempfile.TemporaryDirectory() as d:
   r=Path(d)
   for rel in ['assets/icons/go/orbit-go.svg','assets/icons/go/orbit-go.icns','assets/icons/go/manifest.json','crates/orbit-pi/assets/go-app-icon.png']:
    p=r/rel;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(ROOT/rel,p)
   target=r/'assets/icons/go/orbit-go.icns';data=target.read_bytes();target.write_bytes(data[:-1])
   with self.assertRaises(ValueError):bundle.verify_go_icon(r)
   target.write_bytes(data);m=json.loads((r/'assets/icons/go/manifest.json').read_text());m['icns_slots'].pop('icp4');(r/'assets/icons/go/manifest.json').write_text(json.dumps(m))
   with self.assertRaises(ValueError):bundle.verify_go_icon(r)
 def test_packaged_icon_bytes_and_provenance(self):
  with tempfile.TemporaryDirectory() as d:
   res=Path(d)/'GoPi.app/Contents/Resources'
   bundle.install_go_icon(ROOT,res)
   self.assertEqual((res/'OrbitTest.icns').read_bytes(),(ROOT/'assets/icons/go/orbit-go.icns').read_bytes())
   self.assertEqual((res/'Provenance/go-icon-manifest.json').read_bytes(),(ROOT/'assets/icons/go/manifest.json').read_bytes())
 def test_bundle_uses_go_asset_without_changing_identity(self):
  source=(ROOT/'scripts/macos-app/bundle.py').read_text()
  self.assertIn("copy(root/'assets/icons/go/orbit-go.icns',res/'OrbitTest.icns')",source)
  self.assertIn("'CFBundleIdentifier':'dev.orbit.go.experimental'",source)
  self.assertNotIn("copy(root/'assets/icons/alpha-logo.icns'",source)
  self.assertIn("res/'Provenance/go-icon-manifest.json'",source)
if __name__=='__main__':unittest.main()
