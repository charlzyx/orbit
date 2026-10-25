#!/usr/bin/env bash
# Read-only source validation/build. No publication or signing.
set -euo pipefail
ROOT="$(cd -- "$(dirname "$0")/../.." && pwd -P)"
cd "$ROOT"
: "${RUNNER_TEMP:?}" "${GITHUB_SHA:?}" "${GITHUB_REPOSITORY:?}" "${GITHUB_RUN_ID:?}"
WORK="$RUNNER_TEMP/orbit-macos-app-$GITHUB_SHA"
[ ! -e "$WORK" ] || { echo 'Build staging already exists; refusing overwrite' >&2; exit 1; }
mkdir -p "$WORK/binaries" "$WORK/logs"
export ORBIT_PI_INTEGRATION_BIN="$RUNNER_TEMP/orbit-pi-runtime/node_modules/.bin/pi"
export ORBIT_TERMINAL_TEST=1 ORBIT_REQUIRE_COMPOSER_PATCH=1 ORBIT_REQUIRE_SELECTION_PATCH=1
[ -x "$ORBIT_PI_INTEGRATION_BIN" ]
cp "$RUNNER_TEMP/orbit-pi-runtime/package-lock.json" "$WORK/logs/pi-package-lock.json"
VERSION="$(python3 -c 'import json; print(json.load(open(".github/macos-app-prerelease.json"))["version"])')"
python3 scripts/macos-test/test_package.py 2>&1 | tee "$WORK/logs/legacy-packaging-tests.log"
python3 scripts/macos-app/test_app_package.py 2>&1 | tee "$WORK/logs/packaging-tests.log"
scripts/go-native.sh vet ./... 2>&1 | tee "$WORK/logs/vet.log"
scripts/go-native.sh test -race -json ./... > "$WORK/logs/root-race.jsonl"
scripts/go-native.sh test -race github.com/egoist/mygo/ui -run '^TestOrbit(Composer|Selection)' -count=1 2>&1 | tee "$WORK/logs/toolkit.log"
scripts/go-native.sh -C third_party/mygo-terminal test -race ./... 2>&1 | tee "$WORK/logs/terminal.log"
python3 scripts/test-go-native-overlay.py 2>&1 | tee "$WORK/logs/wrapper.log"
for target in linux/amd64 linux/arm64 darwin/arm64 windows/amd64 windows/arm64; do
  GOOS=${target%/*} GOARCH=${target#*/} CGO_ENABLED=0 scripts/go-native.sh build -ldflags "-X main.version=$VERSION" -o "$WORK/binaries/orbit-${target%/*}-${target#*/}" ./cmd/orbit
 done
export ORBIT_BUILD_WORK="$WORK"
python3 - <<'PY'
from pathlib import Path
import hashlib,json,os
w=Path(os.environ['ORBIT_BUILD_WORK']);counts={'pass':0,'skip':0,'fail':0}
for line in (w/'logs/root-race.jsonl').read_text().splitlines():
 x=json.loads(line)
 if x.get('Test') and x.get('Action') in counts:counts[x['Action']]+=1
if counts['fail'] or not counts['pass']:raise SystemExit('No successful aggregate evidence')
summary={'source_commit':os.environ['GITHUB_SHA'],'status':'PASS','passed_stages':['root-race','vet','toolkit','terminal','wrapper','five-cross-builds','packaging-tests'],'root_tests':counts,'pi_top_level_pin':'@earendil-works/pi-coding-agent@1.4.0','pi_transitives':'Exact repository-locked production graph; resolved package lock preserved in validation logs','pi_package_lock_sha256':hashlib.sha256((w/'logs/pi-package-lock.json').read_bytes()).hexdigest(),'native_macOS_execution':'NOT_RUN','successful_model_inference':'NOT_ESTABLISHED','ci_run':f'https://github.com/{os.environ["GITHUB_REPOSITORY"]}/actions/runs/{os.environ["GITHUB_RUN_ID"]}','binary_sha256':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((w/'binaries').iterdir())}}
(w/'validation-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
PY
echo "binaries=$WORK/binaries" >> "$GITHUB_OUTPUT"
echo "validation=$WORK/validation-summary.json" >> "$GITHUB_OUTPUT"
echo "logs=$WORK/logs" >> "$GITHUB_OUTPUT"
