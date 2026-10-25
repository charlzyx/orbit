# Source only. Never signal a PID/group without matching the launch identity.
chat_identity() {
 local stat rest
 [[ -r /proc/$1/stat ]] || return 1
 stat=$(cat "/proc/$1/stat") || return 1
 rest=${stat##*) }
 # Fields after comm: state(3), ppid(4), pgrp(5), ... starttime(22).
 local fields=(); read -r -a fields <<< "$rest"
 [[ ${fields[2]:-} == "$1" ]] || return 1
 printf '%s:%s\n' "$1" "${fields[19]:-}"
}
chat_cleanup() {
 [[ -n "${app_pid:-}" && -n "${app_identity:-}" ]] || return 0
 [[ $(chat_identity "$app_pid" 2>/dev/null) == "$app_identity" ]] || return 0
 kill -TERM -- -"$app_pid" 2>/dev/null || true
 for attempt in {1..50}; do
  [[ $(chat_identity "$app_pid" 2>/dev/null) == "$app_identity" ]] || return 0
  sleep .1
 done
 [[ $(chat_identity "$app_pid" 2>/dev/null) != "$app_identity" ]] || kill -KILL -- -"$app_pid" 2>/dev/null || true
}
