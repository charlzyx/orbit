# Existing Pi configuration inheritance

The Go macOS app no longer substitutes a private HOME for ordinary Pi launches.
The source reference chooses `PI_CODING_AGENT_DIR` when set, otherwise
`~/.pi/agent`; the corrected Go implementation uses the same user/config roots.

## Path ownership

- `Store.Home` / `RunOptions.Home`: Orbit preferences and app-managed extension,
  patch-support, backup and versioned public-runtime files. The bundled default
  remains `~/Library/Application Support/Orbit Go Test/Home`.
- `Store.UserHome` / `RunOptions.UserHome`: actual Pi HOME (also USERPROFILE for
  children). Normal launches inherit it; it is not the private Orbit home.
- `Store.PiDir` / `RunOptions.PiAgentDir`: `PI_CODING_AGENT_DIR` or
  `UserHome/.pi/agent`, resolved once before switching workspaces. `~` and `~/...`
  expand against UserHome. Relative overrides are intentionally anchored at
  Orbit startup, rather than moving when a session changes its working directory.
- `ORBIT_PREFERENCES_HOME`: only Orbit's guard/title/workflow extensions and
  preference helpers use this. It cannot redirect stock Pi configuration.

The main RPC child, saved-session metadata child, model refresh, generators,
provider CRUD, MCP/global resource commands and quota auth fallback all use the
resolved Pi configuration. External-terminal OAuth fallback carries explicit
non-secret home/config/PATH values across LaunchServices, never copied ambient
credentials. User-scoped skills use UserHome/.agents/skills.

Normal launch preserves ordinary provider/cloud/proxy/SSH and XDG/npm environment
settings while prepending bundled runtime paths. Environment data is not logged.
Helpers redact inherited credential values from errors and user-facing text;
structured stdout is parsed before redaction, so JSON booleans/numbers are not corrupted. A separate
allowlist applies only to opt-in native smoke instrumentation, with offline mode
forced and all homes supplied by its disposable harness.

## Writes and scope

Orbit does not copy auth, models, settings, sessions or credentials into its
private test store and adds no automatic config migration. This is not a read-only
SDK bridge: unmodified stock Pi can perform normal startup migrations, auth/token
refresh, model-cache/settings writes and session persistence. Ordinary user Pi
configuration symlinks remain supported; private app/runtime integrity checks
remain intact. Bundled production updating is still disabled.

Explicit `--home` overrides both Orbit and Pi configuration roots. It does not
strip inherited API-key environment variables or create a filesystem/network
sandbox; tests supply a minimal synthetic environment separately.

This change does not expand the existing session-directory contract: Orbit still
passes its configured `<PiDir>/sessions` explicitly, so stock `settings.sessionDir`
and `PI_CODING_AGENT_SESSION_DIR` are not newly supported by this correction.

## Verification

- Disposable bootstrap tests cover default/tilde config, explicit --home,
  repeated launches, private runtime persistence and no bootstrap config copying.
- Preferences/manager/helper tests cover separate private/user/Pi roots,
  complete child environments, config-path precedence and inherited-secret redaction.
- `TestLiveBundledPiInheritsExistingModelsOffline` launches actual stock Pi 1.4.0
  RPC with a synthetic loopback-only provider/model/default, for both default
  and tilde override paths, twice each. It sends no prompt, login or provider
  request and uses no real credential/configuration environment.
- Quota/guard/title/workflow Node regressions use synthetic auth and mocked fetch;
  changing PiDir invalidates auth and quota report caches.
- Native smoke report verification separately requires private Orbit state,
  the disposable user Pi HOME/config, installed runtime paths and actual
  native first-frame/Ghostty/Pi RPC evidence. These checks were not relaxed.

No native macOS execution, user-Mac access, live provider inference, release,
publication or production-data migration is established by these local tests.
