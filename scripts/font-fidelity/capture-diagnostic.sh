#!/usr/bin/env bash
# Explicitly diagnostic, hosted-runner-only, one static native app scene.
# This script is never run during local/display-free validation.
set -euo pipefail
[[ "${GITHUB_ACTIONS:-}" == true && "${CI:-}" == true ]] || { echo 'Diagnostic native capture requires an explicitly scheduled hosted GitHub Actions job' >&2; exit 2; }
[[ $# == 4 ]] || { echo 'Usage: capture-diagnostic.sh BINARY BUILD_PROVENANCE NEW_OUTPUT trace-off|trace-on' >&2; exit 2; }
binary=$(realpath "$1"); provenance=$(realpath "$2"); output=$(realpath -m "$3"); trace_mode=$4
[[ "$trace_mode" == trace-off || "$trace_mode" == trace-on ]] || exit 2
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
[[ ! -e "$output" ]] || { echo 'Refusing to overwrite capture evidence' >&2; exit 2; }
python3 - "$binary" "$provenance" "$trace_mode" <<'PY'
import hashlib,json,pathlib,sys
binary,provenance,trace=sys.argv[1:]
m=json.loads(pathlib.Path(provenance).read_text())
assert m['diagnostic'] is True and m['production'] is False
assert m['mode'] in ('fractional','fractional-observed')
assert m['binary_sha256']==hashlib.sha256(pathlib.Path(binary).read_bytes()).hexdigest()
assert trace!='trace-on' or m['mode']=='fractional-observed'
assert m['startup_build_tag']=='orbit_diagnostic_fractional_text_linux'
PY
mkdir -m 700 "$output"
cp "$provenance" "$output/build-provenance.json"
printf '{"diagnostic":true,"captureStatus":"starting","scene":"settings-general-dark","traceMode":"%s"}\n' "$trace_mode" > "$output/diagnostic-capture.json"
fixture=/tmp/orbit-visual-fixture
[[ ! -e "$fixture" ]] || { echo 'Refusing to reuse a capture fixture' >&2; exit 2; }
mkdir -m 700 "$fixture"
export HOME="$fixture/home" XDG_CONFIG_HOME="$fixture/config" XDG_CACHE_HOME="$fixture/cache" XDG_RUNTIME_DIR="$fixture/runtime"
install -d -m 700 "$XDG_RUNTIME_DIR"
export PI_CODING_AGENT_DIR="$HOME/.pi/agent" PI_OFFLINE=1 PI_SKIP_VERSION_CHECK=1
export DISPLAY=:92 GDK_BACKEND=x11 GDK_SCALE=1 GDK_DPI_SCALE=1 LIBGL_ALWAYS_SOFTWARE=1 LANG=C.UTF-8 LC_ALL=C.UTF-8 TZ=UTC
unset WAYLAND_DISPLAY ORBIT_METRIC_EXPERIMENT ORBIT_COMPAT_SPECIMEN_ENABLE ORBIT_FONT_TRACE_DIR ORBIT_FONT_TRACE_TEXTS ORBIT_FONT_TRACE_CASE
icd=$(find /usr/share/vulkan/icd.d -maxdepth 1 -name 'lvp_icd*.json' -print -quit)
[[ -z "$icd" ]] || export VK_ICD_FILENAMES="$icd" VK_DRIVER_FILES="$icd"
mkdir -p "$HOME/.orbit-pi" "$PI_CODING_AGENT_DIR" "$XDG_CACHE_HOME" "$fixture/workspace" "$output/trace"
printf '%s\n' '{"language":"en","ui_font_size":14,"terminal_font_size":13,"editor_font_size":13,"spacing_density":100,"reduce_motion":true}' > "$HOME/.orbit-pi/ui.json"
printf '%s\n' '{"mode":"dark","dark_theme":"orbit-azure","light_theme":"orbit-azure-light"}' > "$HOME/.orbit-pi/theme"
if [[ "$trace_mode" == trace-on ]]; then
  export ORBIT_FONT_TRACE_DIR="$output/trace" ORBIT_FONT_TRACE_TEXTS="$root/scripts/font-fidelity/fixture-strings.json" ORBIT_FONT_TRACE_CASE=settings-general-dark-diagnostic
fi
python3 - "$output" <<'PYFIXTURE'
import hashlib,json,os,pathlib,sys
home=pathlib.Path(os.environ['HOME'])
files={str(p.relative_to(home)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(home.rglob('*')) if p.is_file()}
(pathlib.Path(sys.argv[1])/'fixture-inputs.json').write_text(json.dumps({'home':str(home),'workspace':'/tmp/orbit-visual-fixture/workspace','initial_files_sha256':files},indent=2)+'\n')
PYFIXTURE
Xvfb "$DISPLAY" -screen 0 1920x1200x24 -dpi 96 -nolisten tcp > "$output/xvfb.log" 2>&1 & display_pid=$!
trap 'kill ${app_pid:-} ${wm_pid:-} "$display_pid" 2>/dev/null || true' EXIT
for attempt in {1..100}; do xdpyinfo >/dev/null 2>&1 && break; sleep .1; done
xdpyinfo > "$output/xdpyinfo.txt"
openbox > "$output/window-manager.log" 2>&1 & wm_pid=$!
{ vulkaninfo --summary; pkg-config --modversion pango pangocairo cairo harfbuzz fontconfig freetype2; } > "$output/native-versions.txt" 2>&1 || true
cd "$fixture/workspace"
"$binary" --offline > "$output/orbit.log" 2>&1 & app_pid=$!
window=''
for attempt in {1..300}; do
  window=$(xdotool search --onlyvisible --pid "$app_pid" 2>/dev/null | head -1 || true)
  [[ -z "$window" ]] || break
  kill -0 "$app_pid" 2>/dev/null || { cat "$output/orbit.log"; exit 1; }
  sleep .1
done
[[ -n "$window" ]] || { echo 'No native window appeared'; exit 1; }
wmctrl -ir "$window" -b remove,maximized_vert,maximized_horz
wmctrl -ir "$window" -b remove,fullscreen
xdotool windowactivate --sync "$window"
xdotool windowsize --sync "$window" 1470 922
xdotool windowmove "$window" 40 40
sleep 3
xdotool key ctrl+comma
sleep 2
import -window "$window" "$output/settings-general-dark.png"
# A second uncropped image permits stability checking; neither is altered.
sleep 1
import -window "$window" "$output/settings-general-dark-repeat.png"
cp "/proc/$app_pid/maps" "$output/process-maps.txt"
xdotool getwindowfocus > "$output/focused-window.txt"
xdotool getwindowgeometry --shell "$window" > "$output/window-geometry.txt"
python3 - "$output" "$trace_mode" "$window" <<'PY'
import hashlib,json,os,pathlib,subprocess,sys
out=pathlib.Path(sys.argv[1]);mode=sys.argv[2]
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
images={p.name:{'sha256':sha(p),'size':[int(x) for x in subprocess.check_output(['identify','-format','%w %h',str(p)],text=True).split()]} for p in out.glob('*.png')}
assert len(images)==2 and all(i['size']==[1470,922] for i in images.values()),'Wrong viewport; no resizing allowed'
libraries={}
for line in (out/'process-maps.txt').read_text().splitlines():
    fields=line.split(maxsplit=5)
    if len(fields)==6 and fields[5].startswith('/') and '.so' in fields[5]:
        p=pathlib.Path(fields[5])
        if p.is_file(): libraries[str(p)]={'sha256':sha(p),'bytes':p.stat().st_size}
traces=list((out/'trace').glob('mygo-*.json'))
observations=[]
if mode=='trace-on':
    assert traces,'Missing native observation batches'
    for path in traces:
        b=json.loads(path.read_text());assert b['dropped_events']==0 and 'host.present returned' in b['presentation_semantics']
        observations.extend(b['events'])
    shapes=[e for e in observations if e['kind']=='shape_run' and e.get('text')=='General']
    assert shapes and all(e.get('round_positions') is False and e.get('context_hint_metrics')==1 for e in shapes),'Actual native metrics not established'
    assert any(e['kind']=='paint_line' and e.get('text')=='General' for e in observations),'No General app paint observed'
    assert any(e['kind']=='scaled_font' and e.get('hint_metrics')==1 for e in observations),'No fractional Cairo scaled font observed'
else: assert not traces
m={'schema':1,'diagnostic':True,'application':'go-fractional-text-diagnostic','captureStatus':'captured-not-accepted','traceMode':mode,
   'scene':'settings-general-dark','window':sys.argv[3],'images':images,'build_provenance_sha256':sha(out/'build-provenance.json'),
   'trace_files':{p.name:sha(p) for p in traces},'loadedNativeLibraries':libraries,
   'processMapsSha256':sha(out/'process-maps.txt'),'focusedWindow':(out/'focused-window.txt').read_text().strip(),
   'geometryEvidence':(out/'window-geometry.txt').read_text(),
   'fixtureInputSha256':sha(out/'fixture-inputs.json'),
   'traceAllowlistSha256':sha(pathlib.Path(os.environ['ORBIT_FONT_TRACE_TEXTS'])) if mode=='trace-on' else None,
   'nativeEnvironment':{k:os.environ.get(k) for k in ('DISPLAY','GDK_BACKEND','GDK_SCALE','GDK_DPI_SCALE','LIBGL_ALWAYS_SOFTWARE','VK_ICD_FILENAMES','VK_DRIVER_FILES','LANG','LC_ALL','TZ')},
   'requestedViewport':[1470,922],'dpi':96,'uiScale':1,'palette':'orbit-azure',
   'language':'en','fontSize':14,'density':100,'reduceMotion':True,'state':'fresh isolated empty HOME and workspace; offline; no credentials',
   'actualNativePolicyObserved':mode=='trace-on','focusAndCaretPhaseVerified':False,'sceneAndFrameCorrelationVerified':False,
   'native_pixel_parity':'not evaluated','observer_invariance':'requires whole-image zero-changed-pixel comparison of matched trace-off/on captures',
   'limitations':['Startup events can precede the requested scene; trace case ID is not frame correlation.',
                  'Presentation return is not a GPU fence or compositor/scanout proof.',
                  'Wrapping, clipping, caret/selection, decorations and GTK setting transitions need additional matched native cases.']}
(out/'manifest.json').write_text(json.dumps(m,indent=2)+'\n')
(out/'diagnostic-capture.json').write_text(json.dumps(m,indent=2)+'\n')
PY
