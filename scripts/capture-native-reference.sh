#!/usr/bin/env bash
# Hosted-runner-only native reference capture. Never uses a user's display/home.
set -euo pipefail
binary=$(realpath "$1")
output=$(realpath -m "$2")
kind=${3:-reference}
source_sha=$(git rev-parse HEAD)
[[ "$kind" == reference ]] && source_sha=299a32bc87fe3d4851bdb902d9c03dcf9bf99cac
mkdir -p "$output"
fixture=$(mktemp -d)
export HOME="$fixture/home" XDG_CONFIG_HOME="$fixture/config" XDG_CACHE_HOME="$fixture/cache"
export PI_CODING_AGENT_DIR="$HOME/.pi/agent" PI_OFFLINE=1 PI_SKIP_VERSION_CHECK=1
export PATH="/tmp/orbit-pi-runtime/node_modules/.bin:$PATH"
export DISPLAY=:92 GDK_BACKEND=x11 LIBGL_ALWAYS_SOFTWARE=1
export LANG=C.UTF-8 LC_ALL=C.UTF-8
export VK_ICD_FILENAMES=/usr/share/vulkan/icd.d/lvp_icd.x86_64.json
mkdir -p "$HOME/.orbit-pi" "$PI_CODING_AGENT_DIR" "$XDG_CACHE_HOME" "$fixture/workspace"
printf '%s\n' '{"language":"en","ui_font_size":14,"terminal_font_size":13,"editor_font_size":13,"spacing_density":100,"reduce_motion":true}' > "$HOME/.orbit-pi/ui.json"
printf '%s\n' '{"mode":"dark","dark_theme":"orbit-azure","light_theme":"orbit-azure-light"}' > "$HOME/.orbit-pi/theme"
Xvfb "$DISPLAY" -screen 0 1470x922x24 -nolisten tcp > "$output/xvfb.log" 2>&1 & display_pid=$!
trap 'kill ${app_pid:-} ${wm_pid:-} "$display_pid" 2>/dev/null || true' EXIT
for attempt in {1..100}; do xdpyinfo >/dev/null 2>&1 && break; sleep .1; done
openbox > "$output/window-manager.log" 2>&1 & wm_pid=$!
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
xdotool windowactivate --sync "$window"
xdotool windowsize "$window" 1470 922
xdotool windowmove "$window" 0 0
sleep 3
import -window "$window" "$output/new-task-dark.png"
xdotool key ctrl+comma
sleep 1
import -window "$window" "$output/settings-general-dark.png"
SOURCE_SHA="$source_sha" CAPTURE_KIND="$kind" CAPTURE_OUTPUT="$output" python3 - <<'PY'
import json, os, subprocess
out = os.environ['CAPTURE_OUTPUT']
images = {}
for name in ('new-task-dark.png', 'settings-general-dark.png'):
    dimensions = subprocess.check_output(['identify', '-format', '%w %h', out + '/' + name], text=True)
    images[name] = [int(v) for v in dimensions.split()]
manifest = dict(source=os.environ['SOURCE_SHA'], application=os.environ['CAPTURE_KIND'],
    display='Xvfb/Openbox Linux software graphics', requestedViewport=[1470, 922],
    actualImageSizes=images, dpi=1, palette='orbit-azure', language='en', fontSize=14,
    density=100, reduceMotion=True, state='isolated empty home, no credentials or inference',
    note='Verify selected states and client image sizes before pixel comparison; no resizing is allowed to mask layout errors.')
with open(out + '/manifest.json', 'w') as f: json.dump(manifest, f, indent=2)
PY
