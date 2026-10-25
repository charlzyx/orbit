#!/usr/bin/env bash
# Hosted-only opt-in third launch; static Chat only.
set -euo pipefail
[[ "${GITHUB_ACTIONS:-}" == true && "${RUNNER_OS:-}" == Linux && "${DISPLAY:-}" == :92 ]] || { echo 'Static Chat capture requires the approved hosted Linux Xvfb display' >&2; exit 1; }
binary=$(realpath "$1") output=$(realpath -m "$2") kind=$3 source_root=$(realpath "$4")
source_head=$5 source_dirty=$6 source_sha=$7
[[ "$kind" == go || "$kind" == reference ]] || { echo 'Unknown renderer'; exit 1; }
repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
node_binary=$(realpath "$(command -v node)"); [[ $("$node_binary" --version) == v22.23.4 ]] || { echo 'Expected Node22.23.4'; exit 1; }
pi_binary=/tmp/orbit-pi-runtime/node_modules/.bin/pi
fixture=/tmp/orbit-chat-static-fixture
[[ ! -e "$fixture" && ! -L "$fixture" ]] || { echo 'Refusing existing Chat fixture'; exit 1; }
[[ ! -e "$output" && ! -L "$output" ]] || { echo 'Refusing existing Chat output'; exit 1; }
mkdir -p "$output"
python3 "$repo_root/scripts/chat-native/materialize.py" "$fixture" --manifest "$output/fixture-before.json"
cmp "$output/fixture-before.json" "$repo_root/testdata/chat-native/manifest.json"
# Actual Pi, four read-only commands. Context must remain unknown without auth.
python3 "$repo_root/scripts/chat-native/preflight.py" "$pi_binary" "$fixture" "$output/pi-preflight.json"
install -d -m 700 "$fixture/runtime" "$fixture/config" "$fixture/cache"
app_env=(env -i "HOME=$fixture/home" "XDG_CONFIG_HOME=$fixture/config" "XDG_CACHE_HOME=$fixture/cache" "XDG_RUNTIME_DIR=$fixture/runtime"
 "PATH=/tmp/orbit-pi-runtime/node_modules/.bin:$(dirname "$node_binary"):/usr/local/bin:/usr/bin:/bin" "DISPLAY=$DISPLAY" "GDK_BACKEND=x11" "GDK_SCALE=1" "GDK_DPI_SCALE=1"
 "LIBGL_ALWAYS_SOFTWARE=1" "LANG=C.UTF-8" "LC_ALL=C.UTF-8" "TZ=UTC" "SHELL=/bin/bash" "TERM=xterm-256color"
 "PI_CODING_AGENT_DIR=$fixture/home/.pi/agent" "PI_OFFLINE=1" "PI_SKIP_VERSION_CHECK=1" "GIT_CONFIG_NOSYSTEM=1" "GIT_CONFIG_GLOBAL=/dev/null" "GIT_TERMINAL_PROMPT=0")
for variable in VK_ICD_FILENAMES VK_DRIVER_FILES VK_LOADER_DEBUG;do [[ -z "${!variable:-}" ]] || app_env+=("$variable=${!variable}");done
cd "$fixture/workspace";args=();[[ "$kind" != go ]] || args+=(--offline)
setsid "${app_env[@]}" "$binary" "${args[@]}" > "$output/orbit.log" 2>&1 & app_pid=$!
source "$repo_root/scripts/chat-native/owned-process.sh"
app_identity=
trap chat_cleanup EXIT
# setsid exec may not have run when the shell reports $!; never signal until
# the process is observed as its own group leader and its starttime is saved.
for attempt in {1..100}; do
 app_identity=$(chat_identity "$app_pid" 2>/dev/null) && break
 kill -0 "$app_pid" 2>/dev/null || break
 sleep .01
done
[[ -n "$app_identity" ]] || { echo 'No owned Chat process group established'; exit 1; }
window=''
for attempt in {1..300};do window=$(xdotool search --onlyvisible --pid "$app_pid" 2>/dev/null | head -1 || true);[[ -n "$window" ]] && break;kill -0 "$app_pid" 2>/dev/null || { cat "$output/orbit.log";exit 1; };sleep .1;done
[[ -n "$window" ]] || { echo 'No native Chat window appeared';exit 1; }
wmctrl -ir "$window" -b remove,maximized_vert,maximized_horz;wmctrl -ir "$window" -b remove,fullscreen
xdotool windowactivate --sync "$window";xdotool windowsize --sync "$window" 1470 922;xdotool windowmove "$window" 40 40
status=0
python3 "$repo_root/scripts/chat-native/capture.py" "$window" "$app_pid" "$output" "$repo_root/testdata/chat-native/cases.json" "$binary" "$pi_binary" || status=$?
chat_cleanup;wait "$app_pid" 2>/dev/null || true
app_pid=
python3 "$repo_root/scripts/chat-native/materialize.py" "$fixture" --verify "$output/fixture-before.json" --manifest "$output/fixture-after.json" || status=$?
python3 "$repo_root/scripts/chat-native/manifest.py" "$output" "$kind" "$source_root" "$binary" "$repo_root" "$source_head" "$source_dirty" "$source_sha" "$node_binary" || status=$?
exit "$status"
