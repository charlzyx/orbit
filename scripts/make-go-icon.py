#!/usr/bin/env python3
"""Rebuild Go-only icon assets from vector source. Requires installed librsvg/cairo.
No downloads, platform iconutil, fonts or ImageMagick delegates are needed.
Generated PNG-in-ICNS uses Apple's standard size/retina chunk identifiers.
"""
from pathlib import Path
import ctypes as C
import ctypes.util
import hashlib
import json
import struct
import tempfile
import sys

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / 'assets/icons/go/orbit-go.svg'
ICNS = ROOT / 'assets/icons/go/orbit-go.icns'
APP = ROOT / 'crates/orbit-pi/assets/go-app-icon.png'
MANIFEST = ROOT / 'assets/icons/go/manifest.json'
SLOTS = {'icp4':16,'icp5':32,'icp6':64,'ic07':128,'ic08':256,'ic09':512,'ic10':1024,'ic11':32,'ic12':64,'ic13':256,'ic14':512}

def digest(data): return hashlib.sha256(data).hexdigest()
def png_size(data):
    if data[:8] != b'\x89PNG\r\n\x1a\n' or data[12:16] != b'IHDR': raise ValueError('Expected PNG IHDR')
    return struct.unpack('>II', data[16:24])

def render(source, destination, size):
    class Rect(C.Structure):
        _fields_ = [('x',C.c_double),('y',C.c_double),('width',C.c_double),('height',C.c_double)]
    names = [C.util.find_library(x) for x in ('rsvg-2','cairo','gobject-2.0')]
    if not all(names): raise RuntimeError('Install librsvg/cairo development tooling to regenerate; checked-in assets need no generator at runtime')
    r,c,g = [C.CDLL(n) for n in names]
    r.rsvg_handle_new_from_file.argtypes=[C.c_char_p,C.c_void_p];r.rsvg_handle_new_from_file.restype=C.c_void_p
    r.rsvg_handle_render_document.argtypes=[C.c_void_p,C.c_void_p,C.c_void_p,C.c_void_p];r.rsvg_handle_render_document.restype=C.c_int
    c.cairo_image_surface_create.argtypes=[C.c_int,C.c_int,C.c_int];c.cairo_image_surface_create.restype=C.c_void_p
    c.cairo_create.argtypes=[C.c_void_p];c.cairo_create.restype=C.c_void_p
    c.cairo_surface_write_to_png.argtypes=[C.c_void_p,C.c_char_p];c.cairo_surface_write_to_png.restype=C.c_int
    c.cairo_destroy.argtypes=[C.c_void_p];c.cairo_surface_destroy.argtypes=[C.c_void_p];g.g_object_unref.argtypes=[C.c_void_p]
    handle=r.rsvg_handle_new_from_file(str(source).encode(),None)
    if not handle: raise ValueError('Unable to load SVG')
    surface=c.cairo_image_surface_create(0,size,size);ctx=c.cairo_create(surface)
    try:
        if not r.rsvg_handle_render_document(handle,ctx,C.byref(Rect(0,0,size,size)),None): raise ValueError('SVG render failed')
        if c.cairo_surface_write_to_png(surface,str(destination).encode()): raise ValueError('PNG write failed')
    finally:
        c.cairo_destroy(ctx);c.cairo_surface_destroy(surface);g.g_object_unref(handle)

def rebuild():
    images={}
    with tempfile.TemporaryDirectory(prefix='orbit-go-icon-') as temp:
        for size in sorted(set(SLOTS.values())):
            path=Path(temp)/f'{size}.png';render(SOURCE,path,size);images[size]=path.read_bytes()
            if png_size(images[size])!=(size,size): raise ValueError('Incorrect render dimensions')
    chunks=b''.join(kind.encode()+struct.pack('>I',8+len(images[size]))+images[size] for kind,size in SLOTS.items())
    ICNS.write_bytes(b'icns'+struct.pack('>I',8+len(chunks))+chunks)
    APP.write_bytes(images[512])
    manifest={'source':str(SOURCE.relative_to(ROOT)),'source_sha256':digest(SOURCE.read_bytes()),'design':'Go-only pixel PI; stepped charcoal tile with silver square path glyphs shared with GOPI wordmark; no font or badge', 'icns_slots':SLOTS,'outputs':{str(p.relative_to(ROOT)):{'sha256':digest(p.read_bytes()),'bytes':p.stat().st_size} for p in (ICNS,APP)}}
    MANIFEST.write_text(json.dumps(manifest,indent=2,sort_keys=True)+'\n')
    return manifest

if __name__=='__main__':
    if sys.argv[1:] == ['--check']:
        sys.path.insert(0,str(ROOT/'scripts/macos-app'))
        from bundle import verify_go_icon
        print(json.dumps(verify_go_icon(ROOT),indent=2,sort_keys=True))
    elif sys.argv[1:]:
        raise SystemExit('usage: make-go-icon.py [--check]')
    else:
        print(json.dumps(rebuild(),indent=2,sort_keys=True))
