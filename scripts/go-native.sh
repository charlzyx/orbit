#!/usr/bin/env bash
# Canonical native build/test/run/vet: preserve the upstream pin and never edit
# the Go module cache. ORBIT_GO may select the installed Go 1.27.3 toolchain.
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
exec python3 "$root/scripts/go-native-overlay.py" "$@"
