# Existing Pi update status notifications

Pinned Rust `app/pi_update_ui.rs:51–82,266–272` emits informational notices
for an available managed update and for unsupported package-manager/unknown
version installations. Go already emitted the corresponding events from
`native.CheckPiUpdateOnLaunch`, but the owner callback ignored them. This delta
maps only `available`, `unsupported_package_manager`, and `unknown_version` to
the source catalog, including latest-version substitution and English fallback.

No new update check, retry, installer, network operation, navigation, task action,
or installation admission is added. Existing callback owner enqueue is retained.
Updated/failed/migration notifications remain with their prior handlers. Raw
event details are not rendered by the added informational mapper; error handling
and redaction are unchanged. Unknown events remain silent.

**Deferred remains pending:** Rust's next-launch notice says a run is in progress,
but Go's `deferred` also represents absent idle verification and maintenance lease
admission rejection. Mapping all these to a busy-run claim would be incorrect.
This patch deliberately leaves deferred unchanged until a typed reason can
separate those cases; tests verify that no such claim is fabricated.

Mock-event tests first fail against the prior ignored behavior, then verify
source severity/text, latest version, translated/fallback catalogs, mounted toast
delivery, no raw detail leak, and no action/navigation. They do not execute an
actual update check or installer, and do not establish physical OS acceptance.
