#!/usr/bin/env bash
# Run only inside the isolated hosted capture display created by the main harness.
# Coordinates must come from the actual renderer's inspected Simple screenshot.
set -euo pipefail
phase=$1 window=$2 output=$3
: "${ORBIT_CAPTURE_APP_PID:?Set the actual hosted application PID}"
script_dir=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$output"
if [[ -e "$output/usage-${phase}-dark.png" || -e "$output/usage-readiness/${phase}-ready.json" ]] || compgen -G "$output/usage-readiness/${phase}-*.png" >/dev/null; then
  echo 'Refusing previous Usage evidence before sending input' >&2
  exit 1
fi
xdotool windowactivate --sync "$window"
case "$phase" in
 simple) xdotool key ctrl+u ;;
 details|models)
  [[ $# == 5 ]] || { echo 'Supply a verified target x and y for this renderer'; exit 1; }
  xdotool mousemove --window "$window" "$4" "$5" click 1 ;;
 *) echo 'phase must be simple, details, or models'; exit 1 ;;
esac
# Move away from chart/table to avoid an accidental hover readout. This script
# deliberately does not click a model, prompt, export destination or session.
xdotool mousemove --window "$window" 4 4
# Observe actual completed Usage content, not elapsed time or the filter bar.
# The observer sends no input and persists a BLOCKED report on timeout/error.
python3 "$script_dir/wait-ready.py" "$phase" "$window" "$ORBIT_CAPTURE_APP_PID" "$output"
printf '%s\n' "phase=$phase OCR_READY_MANUAL_STATE_REVIEW_REQUIRED client=1470x922 x=${4:-n/a} y=${5:-n/a}" >> "$output/usage-actions.log"
