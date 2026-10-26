# User-approved distinct Go test icon release candidate

Version 0.2.12-go-test.20261026.3, based on exact published ea43b176368a56609afcfc99ce48f84f5f57cf71 plus the reviewed dedicated cyan/GO icon. Existing Rust/Alpha assets, BundleID and Orbit private Home are unchanged. User approved the actual preview; source vector and generated ICNS/PNG/manifest are included in the final file freeze.

This includes already published and fully CI-tested editor buffer/identity protection batches, but NOT the separate unreviewed filesystem transaction candidate. Package .2 remains unchanged. A local temporary .app resource copy proves the icon file/hash wiring, not native Dock rendering. The release must still pass the exact-source full aggregate, hosted ARM64 Mac DMG installation/reopen/Pi/log checks, then actual public four-asset download/hash verification.

No new model/auth calls, user Mac operation, signing/notarization or security bypass. Full function/feel/pixel parity remains open.

Local integrated gate completed: all 1823 frozen non-document hashes match; 100 existing critical identities plus all five icon test/subtest names actually passed. Full race suite has 3372 PASS events, 19 conditional skips; UI 375.328 seconds. Toolkit 19 PASS/zero skips; terminal 115 PASS/4 conditional skips; vet, Darwin arm64 build, packaging36+nested9 tests passed. Hosted exact-source release gates and post-publication download checks remain required.
