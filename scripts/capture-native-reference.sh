#!/usr/bin/env bash
# Hosted-runner-only native reference capture. Never uses a user's display/home.
set -euo pipefail
binary=$(realpath "$1")
output=$(realpath -m "$2")
kind=${3:-reference}
source_root=${4:-}
repo_root=$(git rev-parse --show-toplevel)
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
SOURCE_SHA="$source_sha" CAPTURE_KIND="$kind" CAPTURE_OUTPUT="$output" CAPTURE_REPO="$source_root" CAPTURE_BINARY="$binary" CAPTURE_HARNESS="$repo_root" CAPTURE_SOURCE_HEAD="$capture_source_head" CAPTURE_SOURCE_DIRTY="$capture_source_dirty" python3 - <<'PY'
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
    density=100, reduceMotion=True, state='isolated empty home, no credentials or inference',
    timezone='UTC', clockFrozen=False, animationPhaseFrozen=False,
    fontEvidence='Bundled source inputs only; per-renderer resolved glyph faces remain unverified',
    volatileCases={'settings-runtime-dark.png': ['process PID', 'uptime'], 'settings-about-dark.png': ['development version/build identity']},
    note='Verify selected states and client image sizes before pixel comparison; no resizing is allowed to mask layout errors.')
source_root = pathlib.Path(os.environ['CAPTURE_REPO'])
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
