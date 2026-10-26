# Requested arm64 test release: 2026-10-26.1

User requested a new test release at 04:49 UTC. Runtime is byte-identical to reviewed `a8f42531e3f4bb0c7450dc427a669a971c4545c5`; immediate parent `a9376247f171ac5bc948d6754a1a22829ff0cd0d` changes only the separately guarded Chat capture harness. New unpublished Explorer preflight/version/watcher fixes are deliberately excluded.

The existing macOS app release workflow is reused: exact marker/fingerprint, complete aggregate validation, canonical cross-builds (macOS arm64 only), locked bundled Node/Pi, hosted matching-architecture DMG mount/copy/eject/launch/reopen/Pi/log smoke, then new immutable prerelease. The app name, bundle ID, private Orbit Home, normal Pi configuration inheritance, macOS13 minimum and unsigned/unnotarized status are unchanged. No new signing, credential or security configuration.

The root race invocation now uses the same `-p 1 -timeout 20m` total-suite budget already validated by migration CI. No business timeout, test assertion or test case is removed. The one-shot Chat predicate is retired on this ordinary release-marker commit; the already admitted capture remains attributed to its own immutable SHA.

Tag: `macos-test-v0.2.12-go-test.20261026.1`. Hosted package and download/hash/install/reopen results remain pending until the exact release run completes. Do not infer full parity or user-Mac picker-hang resolution from this package.
