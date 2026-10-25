#!/usr/bin/env bash
# Second app launch on the harness's isolated hosted X11 display. Never a user display.
set -euo pipefail
[[ "${GITHUB_ACTIONS:-}" == true && "${RUNNER_OS:-}" == Linux && "${DISPLAY:-}" == :92 ]] || {
  echo 'Workbench native capture is restricted to the hosted Linux Xvfb harness' >&2; exit 1;
}
binary=$(realpath "$1") output=$(realpath -m "$2") kind=$3 source_root=$(realpath "$4")
source_head=$5 source_dirty=$6 source_sha=$7
repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
node_binary=$(realpath "$(command -v node)")
node_version=$("$node_binary" --version)
[[ "$node_version" == v22.23.4 ]] || { echo "Expected pinned Node v22.23.4, got $node_version"; exit 1; }
node_directory=$(dirname "$node_binary")
fixture=/tmp/orbit-workbench-fixture
[[ ! -e "$output" && ! -L "$output" ]] || { echo 'Refusing preexisting workbench output'; exit 1; }
for command in tesseract xdotool import identify setsid; do command -v "$command" >/dev/null; done
mkdir -p "$output"
python3 "$repo_root/scripts/workbench-native/materialize.py" "$fixture" --manifest "$output/fixture-before.json"
# This compares the generated content, commits, status and diff hashes against
# the checked-in oracle before any real app reads it.
cmp "$output/fixture-before.json" "$repo_root/testdata/workbench-native/manifest.json"
python3 "$repo_root/scripts/workbench-native/materialize.py" "$fixture" --verify "$output/fixture-before.json" --manifest "$output/fixture-preflight.json"
install -d -m 700 "$fixture/runtime" "$fixture/config" "$fixture/cache"
app_env=(env -i "HOME=$fixture/home" "XDG_CONFIG_HOME=$fixture/config" "XDG_CACHE_HOME=$fixture/cache" "XDG_RUNTIME_DIR=$fixture/runtime"
  "PATH=/tmp/orbit-pi-runtime/node_modules/.bin:$node_directory:/usr/local/bin:/usr/bin:/bin" "DISPLAY=$DISPLAY" "GDK_BACKEND=x11" "GDK_SCALE=1" "GDK_DPI_SCALE=1"
  "LIBGL_ALWAYS_SOFTWARE=1" "LANG=C.UTF-8" "LC_ALL=C.UTF-8" "TZ=UTC" "SHELL=/bin/bash" "TERM=xterm-256color"
  "PI_CODING_AGENT_DIR=$fixture/home/.pi/agent" "PI_OFFLINE=1" "PI_SKIP_VERSION_CHECK=1" "GIT_CONFIG_NOSYSTEM=1" "GIT_CONFIG_GLOBAL=/dev/null" "GIT_TERMINAL_PROMPT=0")
for variable in VK_ICD_FILENAMES VK_DRIVER_FILES VK_LOADER_DEBUG; do
  [[ -z "${!variable:-}" ]] || app_env+=("$variable=${!variable}")
done
cd "$fixture/workspace"
args=()
[[ "$kind" != go ]] || args+=(--offline)
setsid "${app_env[@]}" "$binary" "${args[@]}" > "$output/orbit.log" 2>&1 & app_pid=$!
trap 'kill -- -"${app_pid:-0}" 2>/dev/null || true' EXIT
window=''
for attempt in {1..300}; do
  window=$(xdotool search --onlyvisible --pid "$app_pid" 2>/dev/null | head -1 || true)
  [[ -n "$window" ]] && break
  kill -0 "$app_pid" 2>/dev/null || { cat "$output/orbit.log"; exit 1; }
  sleep .1
done
[[ -n "$window" ]] || { echo 'No workbench native window appeared'; exit 1; }
wmctrl -ir "$window" -b remove,maximized_vert,maximized_horz
wmctrl -ir "$window" -b remove,fullscreen
xdotool windowactivate --sync "$window"
xdotool windowsize --sync "$window" 1470 922
xdotool windowmove "$window" 40 40
# Slot 1 must exist before opening it. Repeated sampling proves the sidebar
# label is ready; a fixed startup delay cannot establish that prerequisite.
capture_status=0
python3 - "$repo_root" "$window" "$app_pid" "$output" <<'PY' || capture_status=$?
import importlib.util, json, pathlib, sys
root, window, pid, out = sys.argv[1:]
spec = importlib.util.spec_from_file_location('capture', pathlib.Path(root) / 'scripts/workbench-native/capture.py')
module = importlib.util.module_from_spec(spec); spec.loader.exec_module(module)
suite = json.loads((pathlib.Path(root) / 'testdata/workbench-native/cases.json').read_text())
module.Capture(window, int(pid), out, suite).settle(dict(id='STARTUP', image='startup.png', require=['Workbench fixture']))
# Diagnostic startup is not a ninth workbench case.
(pathlib.Path(out) / 'startup.png').rename(pathlib.Path(out) / 'evidence/startup.png')
PY
if [[ "$capture_status" == 0 ]]; then
  python3 "$repo_root/scripts/workbench-native/capture.py" "$window" "$app_pid" "$output" "$repo_root/testdata/workbench-native/cases.json" || capture_status=$?
else
  printf '%s\n' '{"schema":1,"cases":[],"requestedCases":8,"capturedCases":0,"captureValidity":"BLOCKED","semanticState":"UNREVIEWED","pixelEquality":"NOT_COMPARED","error":"STARTUP readiness failed; inspect evidence/STARTUP-settle.json"}' > "$output/capture-result.json"
fi
# Stop the isolated app and its Pi/shell children before inspecting fixture writes.
kill -- -"$app_pid" 2>/dev/null || true
wait "$app_pid" 2>/dev/null || true
verify_status=0
python3 "$repo_root/scripts/workbench-native/materialize.py" "$fixture" --verify "$output/fixture-before.json" --manifest "$output/fixture-after.json" --session-evidence || verify_status=$?
python3 "$repo_root/scripts/workbench-native/manifest.py" "$output" "$kind" "$source_root" "$binary" "$repo_root" "$source_head" "$source_dirty" "$source_sha" "$node_binary"
[[ "$capture_status" == 0 && "$verify_status" == 0 ]]
