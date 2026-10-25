#!/usr/bin/env python3
"""Create isolated, explicitly diagnostic dependency copies. Never patch inputs."""
from __future__ import annotations
import argparse, hashlib, importlib.util, json, os, shutil, subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parent
MYGO_PIN='9c69cf829ece196c5b3a3601f1c7bbe70526376f'
RUST_PIN='299a32bc87fe3d4851bdb902d9c03dcf9bf99cac'
EXPECTED={
 'mygo':{
  'internal/text/pango_linux.go':'2f260729c9321875c121f05b42a15116696d1cddc8a10dd0f79efd4ce8d57275',
  'ui/paint.go':'b864926e69be1c2c08403c09af7601f942b0ecbdf5381b7e81d5a3b00352b29b',
  'ui/runtime.go':'6631fdb20daf4a31a9abb579c150a414bc18888ad677b05ea931a4c788834463'},
 'gpui':{
  'Cargo.toml':'f5686c4367d640e41e57a53679b89417ddf82b11ef4a484b8d1e6134763f74fc',
  'src/gpui.rs':'31c6ee8d2f6572a38287bb98f5523d0e1aee670dfdeff89a25d5e344252a373a',
  'src/window.rs':'40d58470971ccd658ead4d4f788219658fb68b9486aff73e143835c09237ab18',
  'src/text_system.rs':'684330bdba94a43750a8ff28872f218017380019d1f313db61d338c0fbb75fc3',
  'src/text_system/line.rs':'4aec81b5d8492fdb4c4d68245dafaedb0a3e00779da32050bc0dabcea9fa0e81',
  'src/platform/linux/text_system.rs':'a3415cb019a36f6a0ddd70649a34c591e6c79974466715b5a4ac4ac63c8b1585'}
}
EXPECTED_TREE={'mygo':'96dae6d68d612b648267f72fdf90f2211698612d29f3c89ca923a314c23af110','gpui':'e19dd725f79a9e348c07388f12bae1f2c38e85660012a0e29c2d7a96e0bdcb51'}
def tree_sha(root):
 rows={str(p.relative_to(root)):sha(p) for p in root.rglob('*') if p.is_file() and '.git' not in p.relative_to(root).parts and p.name!='.cargo-ok'}
 return hashlib.sha256(json.dumps(rows,sort_keys=True,separators=(',',':')).encode()).hexdigest()
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def replace(root,path,old,new,count=1):
 p=root/path;s=p.read_text()
 if s.count(old)!=count:raise ValueError(f'anchor drift {path}: expected {count}, got {s.count(old)}')
 p.write_text(s.replace(old,new))
def verify(root,kind):
 if tree_sha(root)!=EXPECTED_TREE[kind]:raise ValueError(f'Pinned {kind} full source tree mismatch')
 for rel,digest in EXPECTED[kind].items():
  if sha(root/rel)!=digest:raise ValueError(f'Pinned {kind} source hash mismatch: {rel}')
def prepare_mygo(src,out,repo):
 verify(src,'mygo');shutil.copytree(src,out,ignore=shutil.ignore_patterns('.git'))
 # Apply the SAME reviewed composer/selection compatibility, independently of
 # the production wrapper. Its checks are imported, not weakened or modified.
 spec=importlib.util.spec_from_file_location('native_overlay',repo/'scripts/go-native-overlay.py');mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
 compat=repo/'third_party/mygo-composer';manifest=json.loads((compat/'baseline.json').read_text())
 if manifest['upstream_commit']!=MYGO_PIN:raise ValueError('wrong MyGo pin')
 for rel,digest in manifest['baseline_sha256'].items():
  if sha(src/rel)!=digest:raise ValueError(f'composer baseline hash mismatch: {rel}')
 source=compat/'orbit_composer.go.txt'
 if sha(source)!=manifest['added_sha256']:raise ValueError('composer source hash mismatch')
 shutil.copyfile(source,out/manifest['added_file'])
 for rel,source in mod.verify_selection(src,repo/'third_party/mygo-selection',manifest).items():shutil.copyfile(source,out/rel)
 for name in ('mygo_trace.go','mygo_pango_trace.go','mygo_specimen_test.go'):shutil.copyfile(ROOT/(name+'.txt'),out/'internal/text'/('orbit_'+name))
 replace(out,'internal/text/pango_linux.go','\te.applyOptions()\n\tfor _, sf := range e.scaled {','\te.applyOptions()\n\torbitRecord(orbitEvent{"kind":"font_rendering_settings","source":"MyGo setFontRendering caller inputs","antialias":antialias,"hinting":hinting,"subpixels":subpixels})\n\tfor _, sf := range e.scaled {')
 replace(out,'internal/text/pango_linux.go','\treturn lines\n}', '\te.orbitShape(text, style, spans, width, rtl, wholeWords, lines)\n\treturn lines\n}')
 replace(out,'internal/text/pango_linux.go','\te.scaled[key] = sf\n', '\te.scaled[key] = sf\n\te.orbitScaled(f, sf, scale, subpixel)\n')
 replace(out,'ui/paint.go','\t\tbaseline := sys.Baseline((y + line.Baseline) * s)\n','\t\tbaseline := sys.Baseline((y + line.Baseline) * s)\n\t\ttext.OrbitTracePaint(l, li, x, y, s, baseline, [4]float32{p.clip.X,p.clip.Y,p.clip.W,p.clip.H}, p.transform.Set || moves)\n')
 replace(out,'ui/paint.go','\t\t\tif run != nil {\n\t\t\t\trun.add(p, g, gi, pen, sg, glyphShade, baseline)', '\t\t\ttext.OrbitTraceGlyph(l, li, g, pen, baseline, [4]float32{sg.X,sg.Y,sg.W,sg.H})\n\t\t\tif run != nil {\n\t\t\t\trun.add(p, g, gi, pen, sg, glyphShade, baseline)')
 replace(out,'ui/runtime.go','\trt.host.present(&rt.scene)\n','\trt.host.present(&rt.scene)\n\ttext.OrbitTracePresent()\n',count=2)
 verify(src,'mygo')
def prepare_gpui(src,out):
 verify(src,'gpui');shutil.copytree(src,out)
 shutil.copyfile(ROOT/'gpui_trace.rs.txt',out/'src/orbit_font_trace.rs')
 shutil.copyfile(ROOT/'gpui_specimen.rs.txt',out/'src/platform/linux/orbit_font_specimen.rs')
 with (out/'src/platform/linux/text_system.rs').open('a') as f:f.write('\ninclude!("orbit_font_specimen.rs");\n')
 replace(out,'Cargo.toml','[dependencies.serde_json]\n','[dependencies.sha2]\nversion = "=0.10.9"\n\n[dependencies.serde_json]\n')
 replace(out,'src/gpui.rs','mod text_system;','#[cfg(target_os = "linux")]\nmod orbit_font_trace;\n#[cfg(target_os = "linux")]\n#[doc(hidden)]\npub fn orbit_font_specimen() { crate::platform::orbit_font_specimen_impl(); }\nmod text_system;')
 replace(out,'src/text_system.rs','        let mut runs = runs.iter().filter(|run| run.len > 0).cloned().peekable();','        #[cfg(target_os = "linux")]\n        crate::orbit_font_trace::request(&text, font_size, runs);\n        let mut runs = runs.iter().filter(|run| run.len > 0).cloned().peekable();')
 replace(out,'src/text_system.rs','        let mut last_run = None::<&TextRun>;','        #[cfg(target_os = "linux")]\n        crate::orbit_font_trace::request(text, font_size, runs);\n        let mut last_run = None::<&TextRun>;')
 # Observe AFTER CosmicText fallback correction and emoji skip, before the
 # unchanged ShapedGlyph construction. All original expressions remain intact.
 replace(out,'src/platform/linux/text_system.rs','            let shaped_glyph = ShapedGlyph {','''            if crate::orbit_font_trace::allowed(text) {
                let face = self.font_system.db().face(loaded_font.font.id()).unwrap();
                crate::orbit_font_trace::face(font_id, &loaded_font.font, face);
                crate::orbit_font_trace::record(serde_json::json!({
                    "kind":"shape_glyph", "text":text, "font_id":font_id.0,
                    "requested_font_id":glyph.metadata, "fallback":font_id.0 != glyph.metadata,
                    "id":glyph.glyph_id, "start_byte":glyph.start,"end_byte":glyph.end,
                    "x":glyph.x,"y":glyph.y,"advance":glyph.w,"x_offset":glyph.x_offset,"y_offset":glyph.y_offset,
                    "size":font_size.0,"line_width":layout.w,"ascent":layout.max_ascent,"descent":layout.max_descent,
                    "locale":self.font_system.locale(),"cluster_units":"UTF-8 byte index"
                }));
            }
            let shaped_glyph = ShapedGlyph {''')
 replace(out,'src/text_system/line.rs','        paint_line(\n            origin,','        paint_line(\n            &self.text,\n            origin,',count=2)
 replace(out,'src/text_system/line.rs','fn paint_line(\n    origin:', 'fn paint_line(\n    trace_text: &str,\n    origin:')
 replace(out,'src/text_system/line.rs','    window.paint_layer(line_bounds, |window| {\n        let padding_top', '    #[cfg(target_os = "linux")]\n    crate::orbit_font_trace::paint(trace_text, layout, origin, line_height, window.scale_factor());\n    window.paint_layer(line_bounds, |window| {\n        let padding_top')
 replace(out,'src/text_system/line.rs','                if max_glyph_bounds.intersects(&content_mask.bounds) {','                if max_glyph_bounds.intersects(&content_mask.bounds) {\n                    #[cfg(target_os = "linux")]\n                    crate::orbit_font_trace::glyph(trace_text, run.font_id, glyph, glyph_origin + baseline_offset, window.scale_factor());')
 replace(out,'src/window.rs','        self.platform_window.draw(&self.rendered_frame.scene);','        self.platform_window.draw(&self.rendered_frame.scene);\n        #[cfg(target_os = "linux")]\n        crate::orbit_font_trace::present();')
 verify(src,'gpui')
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--diagnostic',action='store_true',required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--mygo-source',type=Path);p.add_argument('--gpui-source',type=Path);p.add_argument('--repo',type=Path,default=ROOT.parent.parent);a=p.parse_args()
 if not a.mygo_source and not a.gpui_source:p.error('supply at least one pinned source')
 if a.output.exists():p.error('output must not exist (no in-place or repeated patching)')
 # Explicitly reject output beneath source or repository, and source beneath output.
 output=a.output.resolve()
 for source in [a.repo,a.mygo_source,a.gpui_source]:
  if source and (output.is_relative_to(source.resolve()) or source.resolve().is_relative_to(output)):p.error('output and input roots must be disjoint')
 a.output.mkdir(parents=True)
 if a.mygo_source:prepare_mygo(a.mygo_source,a.output/'mygo',a.repo)
 if a.gpui_source:prepare_gpui(a.gpui_source,a.output/'gpui')
 manifest={'schema':1,'diagnostic':True,'rust_baseline':RUST_PIN,'mygo_pin':MYGO_PIN,'gpui_version':'0.2.2','inputs':EXPECTED,'input_tree_sha256':EXPECTED_TREE,'scripts':{p.name:sha(p) for p in ROOT.iterdir() if p.is_file()},'outputs':{str(p.relative_to(a.output)):sha(p) for p in a.output.rglob('*') if p.is_file() and (p.suffix in ('.go','.rs') or p.name=='Cargo.toml')}}
 (a.output/'diagnostic-provenance.json').write_text(json.dumps(manifest,indent=2)+'\n')
 print(a.output/'diagnostic-provenance.json')
if __name__=='__main__':main()
