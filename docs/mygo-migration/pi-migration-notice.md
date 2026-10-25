# Pi migration notification source mapping

The live startup `migration_required` event previously directed users to
Settings → Runtime, but the source-matched Runtime page has no migration control.
This narrow correction uses the existing source `pi_update.migration_required`
locale key, substitutes `event.Latest`, and uses error severity as in pinned
`crates/orbit-pi/src/app/pi_update_ui.rs:60–70`. Existing owner-thread enqueue,
other update/error notifications, redaction and RuntimeChanged dispatch are
unchanged. There is no new navigation, runtime operation or installer action.

This fixes notification copy, **not the underlying source entry contradiction**:
Rust `setup_open` has no reachable true assignment; setup is hidden when all
required dependencies are ready. Its notification nevertheless says to migrate
on the setup page. The separate npm migration card only exists when setup is
already visible. This patch does not invent an About/settings button or claim an
all-ready user can reach that card. That semantic entry gap remains explicit.

The fixture first failed against the legacy message (wrong text, missing latest
version and info severity), then checks catalog/fallback substitution, error
toast, mounted delivery and absence of action dispatch/navigation. Tests never
run the network update check or any installer; physical native acceptance and a
reachable all-ready migration route are not established by these fixtures.
