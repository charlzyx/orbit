#!/usr/bin/env bash
# Run only inside the isolated hosted capture display created by the main harness.
# Coordinates must come from the actual renderer's inspected Simple screenshot.
set -euo pipefail
phase=$1 window=$2 output=$3
mkdir -p "$output"
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
sleep 1
import -window "$window" "$output/usage-${phase}-dark.png"
size=$(identify -format '%w %h' "$output/usage-${phase}-dark.png")
[[ "$size" == '1470 922' ]] || { echo "Invalid client size: $size"; exit 1; }
printf '%s\n' "phase=$phase expected-state-not-yet-visually-verified client=1470x922 x=${4:-n/a} y=${5:-n/a}" >> "$output/usage-actions.log"
