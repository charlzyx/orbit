#!/usr/bin/env bash
# CI-only phase logging. Actions owns the five-minute step deadline and cleanup.
# No custom process supervisor, mirror overrides, retries, or permission changes.
set -euo pipefail
case "${1:-}" in
  go) packages=(libgtk-3-0 libpango-1.0-0 libcairo2 libfontconfig1 libgl1-mesa-dri mesa-vulkan-drivers vulkan-tools xvfb openbox xdotool wmctrl x11-utils imagemagick tesseract-ocr) ;;
  reference) packages=(clang cmake pkg-config libfontconfig-dev libvulkan1 mesa-vulkan-drivers vulkan-tools libwayland-dev libx11-xcb-dev libxi-dev libxtst-dev libxkbcommon-x11-dev libssl-dev libgtk-3-0 libpango-1.0-0 libcairo2 xvfb openbox xdotool wmctrl x11-utils imagemagick tesseract-ocr) ;;
  *) echo 'usage: install-native-deps.sh go|reference' >&2; exit 64 ;;
esac
phase() (
  local name="$1"; shift
  local started=$SECONDS
  printf 'ORBIT_APT_PHASE_START phase=%s utc=%s\n' "$name" "$(date -u +%FT%TZ)"
  trap 'rc=$?; printf "ORBIT_APT_PHASE_END phase=%s utc=%s elapsed_s=%s exit=%s\n" "$name" "$(date -u +%FT%TZ)" "$((SECONDS-started))" "$rc"; exit "$rc"' EXIT
  sudo apt-get "$@"
)
phase update update
phase install install -y "${packages[@]}"
