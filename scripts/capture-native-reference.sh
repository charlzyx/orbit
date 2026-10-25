#!/usr/bin/env bash
# Hosted-runner-only native reference capture. Never uses a user's display/home.
set -euo pipefail
binary=$(realpath "$1")
output=$(realpath -m "$2")
kind=${3:-reference}
source_root=${4:-}
repo_root=$(git rev-parse --show-toplevel)
node_binary=$(realpath "$(command -v node)")
node_version=$("$node_binary" --version)
[[ "$node_version" == v22.23.4 ]] || { echo "Expected pinned Node v22.23.4, got $node_version"; exit 1; }
source_sha=$(git rev-parse HEAD)
if [[ -z "$source_root" ]]; then
  [[ "$kind" == go ]] || { echo "Reference capture requires its actual pinned source root as argument4"; exit 1; }
  source_root="$repo_root"
fi
source_root=$(realpath "$source_root")
[[ "$kind" == reference ]] && source_sha=299a32bc87fe3d4851bdb902d9c03dcf9bf99cac
# Record source cleanliness before creating capture outputs inside the checkout.
# The later screenshots/logs are evidence, not application source modifications.
capture_source_head=$(git -C "$source_root" rev-parse HEAD)
capture_source_dirty=false
[[ -z "$(git -C "$source_root" status --porcelain)" ]] || capture_source_dirty=true
mkdir -p "$output"
fixture=/tmp/orbit-visual-fixture
[[ ! -e "$fixture" ]] || { echo "Refusing to reuse a preexisting capture fixture"; exit 1; }
mkdir -m 700 "$fixture"
export HOME="$fixture/home" XDG_CONFIG_HOME="$fixture/config" XDG_CACHE_HOME="$fixture/cache" XDG_RUNTIME_DIR="$fixture/runtime"
install -d -m 700 "$XDG_RUNTIME_DIR"
export PI_CODING_AGENT_DIR="$HOME/.pi/agent" PI_OFFLINE=1 PI_SKIP_VERSION_CHECK=1
export PATH="/tmp/orbit-pi-runtime/node_modules/.bin:$PATH"
export DISPLAY=:92 GDK_BACKEND=x11 GDK_SCALE=1 GDK_DPI_SCALE=1 LIBGL_ALWAYS_SOFTWARE=1
export LANG=C.UTF-8 LC_ALL=C.UTF-8 TZ=UTC
icd=$(find /usr/share/vulkan/icd.d -maxdepth 1 -name 'lvp_icd*.json' -print -quit)
if [[ -n "$icd" ]]; then
  export VK_ICD_FILENAMES="$icd" VK_DRIVER_FILES="$icd"
fi
export VK_LOADER_DEBUG=error,warn
mkdir -p "$HOME/.orbit-pi" "$PI_CODING_AGENT_DIR" "$XDG_CACHE_HOME" "$fixture/workspace"
printf '%s\n' '{"language":"en","ui_font_size":14,"terminal_font_size":13,"editor_font_size":13,"spacing_density":100,"reduce_motion":true}' > "$HOME/.orbit-pi/ui.json"
printf '%s\n' '{"mode":"dark","dark_theme":"orbit-azure","light_theme":"orbit-azure-light"}' > "$HOME/.orbit-pi/theme"
# Optional synthetic input is identical for both apps; never seed a real HOME.
if [[ "${ORBIT_CAPTURE_USAGE:-0}" == 1 ]]; then
  python3 "$repo_root/scripts/usage-native/materialize.py" "$HOME" --manifest "$output/usage-fixture.json"
  mkdir -p "$fixture/workspace/alpha" "$fixture/workspace/beta"
fi
Xvfb "$DISPLAY" -screen 0 1920x1200x24 -dpi 96 -nolisten tcp > "$output/xvfb.log" 2>&1 & display_pid=$!
trap 'kill ${app_pid:-} ${wm_pid:-} "$display_pid" 2>/dev/null || true' EXIT
for attempt in {1..100}; do xdpyinfo >/dev/null 2>&1 && break; sleep .1; done
openbox > "$output/window-manager.log" 2>&1 & wm_pid=$!
{
  printf 'Selected software Vulkan ICD: %s\n' "${icd:-default discovery}"
  ls -l /usr/share/vulkan/icd.d
  xdpyinfo | head -30
  vulkaninfo --summary
} > "$output/graphics.log" 2>&1 || true
cd "$fixture/workspace"
if [[ "$kind" == go ]]; then
  "$binary" --offline > "$output/orbit.log" 2>&1 & app_pid=$!
else
  "$binary" > "$output/orbit.log" 2>&1 & app_pid=$!
fi
window=""
for attempt in {1..300}; do
  window=$(xdotool search --onlyvisible --pid "$app_pid" 2>/dev/null | head -1 || true)
  [[ -n "$window" ]] && break
  kill -0 "$app_pid" 2>/dev/null || { cat "$output/orbit.log"; exit 1; }
  sleep .1
done
[[ -n "$window" ]] || { echo 'No native window appeared'; exit 1; }
wmctrl -ir "$window" -b remove,maximized_vert,maximized_horz
wmctrl -ir "$window" -b remove,fullscreen
sleep .3
xdotool windowactivate --sync "$window"
xdotool windowsize --sync "$window" 1470 922
xdotool windowmove "$window" 40 40
sleep 3
import -window "$window" "$output/new-task-dark.png"
xdotool key ctrl+comma
sleep 1
import -window "$window" "$output/settings-general-dark.png"
# These are the source's actual14 navigation rows at the fixed client size.
# The first viewport is not evidence for controls below its fold; those retain
# separate unexecuted register cases. No network/provider login is performed.
while read -r section y; do
  xdotool mousemove --window "$window" 80 "$y" click 1
  sleep 1
  import -window "$window" "$output/settings-${section}-dark.png"
done <<'SECTIONS'
appearance 202
shortcuts 232
privacy 262
runtime 307
agent 337
worktrees 367
providers 397
models 427
skills 472
plugins 502
mcp 532
about 577
report-bug 622
SECTIONS
if [[ "${ORBIT_CAPTURE_USAGE:-0}" == 1 ]]; then
  # Escape leaves Settings before the source global Usage shortcut. No clicks
  # on guessed chart/table coordinates, no session opening or model inference.
  xdotool key Escape
  export ORBIT_CAPTURE_APP_PID="$app_pid"
  bash "$repo_root/scripts/usage-native/capture-state.sh" simple "$window" "$output"
  # Resolve Details from THIS run's full native client image. An older
  # screenshot's coordinates are not authority for a new pointer action.
  # Verification sends no input; the existing helper sends exactly one click.
  details_target=$(python3 "$repo_root/scripts/workbench-native/verify-target.py" "$window" "$app_pid" "$output/usage-evidence" Details 1000 70 1470 160 --require-ready Summary "Usage over time" requests tokens sessions)
  read -r details_x details_y <<< "$details_target"
  bash "$repo_root/scripts/usage-native/capture-state.sh" details "$window" "$output" "$details_x" "$details_y"
  printf '%s\n' 'Details target evidence: usage-evidence/details-target.json; selected state remains unreviewed' >> "$output/usage-actions.log"
fi
SOURCE_SHA="$source_sha" CAPTURE_KIND="$kind" CAPTURE_OUTPUT="$output" CAPTURE_REPO="$source_root" CAPTURE_BINARY="$binary" CAPTURE_HARNESS="$repo_root" CAPTURE_SOURCE_HEAD="$capture_source_head" CAPTURE_SOURCE_DIRTY="$capture_source_dirty" CAPTURE_NODE_BINARY="$node_binary" python3 - <<'PY'
import json, os, subprocess, pathlib, hashlib
out = os.environ['CAPTURE_OUTPUT']
images = {}
for name in sorted(p.name for p in pathlib.Path(out).glob('*.png')):
    dimensions = subprocess.check_output(['identify', '-format', '%w %h', out + '/' + name], text=True)
    images[name] = [int(v) for v in dimensions.split()]
    if images[name] != [1470, 922]:
        raise SystemExit(f"Invalid native client viewport for {name}: {images[name]}; no image resizing allowed")
manifest = dict(source=os.environ['SOURCE_SHA'], application=os.environ['CAPTURE_KIND'],
    display='Xvfb/Openbox Linux software graphics', requestedViewport=[1470, 922],
    actualImageSizes=images, dpi=96, uiScale=1, palette='orbit-azure', language='en', fontSize=14,
    density=100, reduceMotion=True, state=('isolated synthetic Usage fixture, no credentials or inference' if os.environ.get('ORBIT_CAPTURE_USAGE') == '1' else 'isolated empty home, no credentials or inference'),
    timezone='UTC', clockFrozen=False, animationPhaseFrozen=False,
    fontEvidence='Bundled source inputs only; per-renderer resolved glyph faces remain unverified',
    volatileCases={'settings-runtime-dark.png': ['process PID', 'uptime'], 'settings-about-dark.png': ['development version/build identity']},
    note='Verify selected states and client image sizes before pixel comparison; no resizing is allowed to mask layout errors.')
usage_fixture = pathlib.Path(out) / 'usage-fixture.json'
if usage_fixture.exists():
    manifest['usageFixtureSha256'] = hashlib.sha256(usage_fixture.read_bytes()).hexdigest()
    manifest['usageFixture'] = json.loads(usage_fixture.read_text())
    manifest['usageCaptureState'] = 'Simple and Details each passed three consecutive complete-body OCR readiness observations without recognized loading/error. Details also required completed Simple body anchors and a unique exact-label OCR target on one fresh frame. No guessed target. Semantic states and pixels still require manual review'
    manifest['usageReadiness'] = {phase: json.loads((pathlib.Path(out) / 'usage-readiness' / (phase + '-ready.json')).read_text()) for phase in ('simple', 'details')}
    manifest['usageDetailsTargetEvidence'] = json.loads((pathlib.Path(out) / 'usage-evidence/details-target.json').read_text())
    manifest['volatileCases']['usage-simple-dark.png'] = ['scan-relative header age', 'activity calendar trailing-year anchor']
    manifest['volatileCases']['usage-details-dark.png'] = ['scan-relative header age']
source_root = pathlib.Path(os.environ['CAPTURE_REPO'])
manifest['runtimeNode'] = dict(path=os.environ['CAPTURE_NODE_BINARY'], version=subprocess.check_output([os.environ['CAPTURE_NODE_BINARY'], '--version'], text=True).strip(), sha256=hashlib.sha256(pathlib.Path(os.environ['CAPTURE_NODE_BINARY']).read_bytes()).hexdigest())
manifest['binarySha256'] = hashlib.sha256(pathlib.Path(os.environ['CAPTURE_BINARY']).read_bytes()).hexdigest()
manifest['sourceRootGitHead'] = os.environ['CAPTURE_SOURCE_HEAD']
manifest['sourceRootDirty'] = os.environ['CAPTURE_SOURCE_DIRTY'] == 'true'
manifest['sourceCleanlinessRecorded'] = 'Before starting app or creating capture output; includes untracked source files'
manifest['harnessGitHead'] = subprocess.check_output(['git', '-C', os.environ['CAPTURE_HARNESS'], 'rev-parse', 'HEAD'], text=True).strip()
manifest['resolvedFontFaces'] = None
manifest['focusAndCaretPhaseVerified'] = False
manifest['compatibilityManifests'] = {str(p.relative_to(source_root)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((source_root / 'third_party').glob('mygo-*/baseline.json'))} if os.environ['CAPTURE_KIND'] == 'go' else {}
fonts = source_root / 'crates/orbit-pi/assets/fonts'
manifest['bundledFontInputs'] = {str(p.relative_to(fonts)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(fonts.rglob('*')) if p.suffix.lower() in ('.ttf', '.otf')}
with open(out + '/manifest.json', 'w') as f: json.dump(manifest, f, indent=2)
PY
# Workbench is a separate real app launch/home so the existing Settings/Usage
# fixture and its image count are unchanged. Child output has its own manifest.
if [[ "${ORBIT_CAPTURE_WORKBENCH:-0}" == 1 ]]; then
  kill "$app_pid" 2>/dev/null || true
  wait "$app_pid" 2>/dev/null || true
  bash "$repo_root/scripts/workbench-native/capture.sh" "$binary" "$output/workbench" "$kind" "$source_root" "$capture_source_head" "$capture_source_dirty" "$source_sha"
fi
