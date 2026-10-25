package orbitui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
)

func (a *App) setupMigrationCard(c *ui.Context) {
	ui.Column(c.Key("setup-npm-migration")).AlignSelf(ui.Stretch).Padding(12, 14).Gap(8).Background(a.Theme.Elevated).Border(1, a.Theme.Border).Radius(10).Children(func() {
		ui.Text(c, a.Tr("pi_update.migrate_title")).FontWeight(600)
		ui.Text(c, a.Tr("pi_update.migrate_note")).TextColor(a.Theme.TextMuted)
		label := a.Tr("pi_update.migrate")
		if a.setup.migrating {
			label = a.Tr("pi_update.migrating")
		}
		if a.button(c, label, "", false).Label("Migrate npm Pi installation").Disabled(a.setup.migrating || a.setup.refreshing).Clicked() {
			a.Dispatch(Action{Kind: "RequestSetupMigration"})
		}
	})
}
func (x *Controller) requestSetupMigration() {
	u := x.UI
	if x.hostStopped() || u.setup.refreshing || !u.setupBlocked() || u.setup.installation.Kind != native.PiNPM || x.setupInstall.pending != nil {
		return
	}
	intent := x.captureSetupInstallIntent(true)
	usingNative := x.setupInstall.backend == nil
	if usingNative && !x.nativeSetupTargetMatches() {
		u.Notify("The selected Pi home differs from the native installer environment; use the selected installation’s package manager instead", true)
		return
	}
	backend := nativeSetupInstallBackend()
	if x.setupInstall.backend != nil {
		backend = *x.setupInstall.backend
	}
	u.Confirm(u.Tr("pi_update.migrate_title"), u.Tr("pi_update.migrate_note"), func() {
		if x.hostStopped() {
			return
		}
		if usingNative && !x.nativeSetupTargetMatches() {
			u.Notify("The native installer target changed; refresh setup before retrying", true)
			return
		}
		if u.setup.refreshing || !u.setupBlocked() || u.setup.installation.Kind != native.PiNPM || !x.ownsSetupInstallIntent(intent) {
			u.Notify("Setup confirmation is stale; refresh and try again", true)
			return
		}
		x.setupInstall.onReceipt = x.applySetupMigrationReceipt
		if err := x.beginSetupInstall(intent, backend); err != nil {
			u.Notify(err.Error(), true)
			return
		}
		u.setup.migrating = true
	})
}
func (x *Controller) applySetupMigrationReceipt(receipt setupInstallReceipt) {
	u := x.UI
	s := &u.setup
	s.migrating = false
	// Early admission/preflight failures have no new dependency snapshot. Keep
	// the current report rather than inventing that dependencies disappeared.
	if len(receipt.dependencies) > 0 {
		s.dependencies = append([]native.Dependency(nil), receipt.dependencies...)
		s.installation = receipt.installation
		s.loaded = true
		s.error = ""
		facts := make([]native.HostFact, 0, len(s.facts)+3)
		for _, f := range s.facts {
			switch f.LabelKey {
			case "onboarding.pi_install", "onboarding.pi_path", "onboarding.pi_version":
				continue
			}
			facts = append(facts, f)
		}
		s.facts = append(facts, native.PiInstallFacts(receipt.installation)...)
	}
	if receipt.err != nil {
		u.Notify(strings.ReplaceAll(u.Tr("pi_update.migrate_failed"), "%{detail}", receipt.err.Error()), true)
		return
	}
	if !receipt.verified {
		u.Notify(u.Tr("pi_update.verify_failed"), true)
		return
	}
	u.Notify(strings.ReplaceAll(u.Tr("pi_update.migrated"), "%{version}", receipt.installation.Version), false)
	if receipt.warning != "" {
		u.Notify(u.Tr("pi_update.installed_not_on_path"), true)
	}
	// Source migrate_pi_install updates dependency/installation facts but does
	// not spawn/restart a task. Required Node readiness is still a separate gate.
}

// Native installation currently uses the process environment. Do not silently
// migrate a different user/agent home when Orbit has an explicitly isolated one.
// Conservative textual comparison may reject aliases; it never rewrites PATH/HOME.
func (x *Controller) nativeSetupTargetMatches() bool {
	home, err := os.UserHomeDir()
	return err == nil && home != "" && filepath.Clean(x.Store.UserHome) == filepath.Clean(home) && filepath.Clean(x.Store.PiDir) == filepath.Clean(native.PiAgentDirectory())
}
