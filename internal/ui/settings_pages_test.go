package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppearanceSourceGroupsMappedChoicesAndPersistence(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.selectSettingsPage("appearance")
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	text := strings.Join(tt.Texts(), "\n")
	for _, label := range []string{"THEME & BACKGROUND", "TYPE & DENSITY", "INTERFACE PREVIEW", "TERMINAL PREVIEW", "Orbit Azure", "IBM Plex Sans", "Lilex", "English", "14 px", "100 %"} {
		if !strings.Contains(text, label) {
			t.Fatalf("missing source appearance content %q\n%s", label, text)
		}
	}
	for _, label := range []string{"Dark theme", "Light theme", "TYPOGRAPHY", "LAYOUT AND MOTION", ".ZedSans"} {
		if strings.Contains(text, label) {
			t.Fatal("obsolete generic appearance row", label)
		}
	}
	if err := tt.Click("Light"); err != nil {
		t.Fatal(err)
	}
	if a.Options.Store.Appearance().Mode != "light" {
		t.Fatal("segmented mode did not persist")
	}
	tt.Key(0, ui.KeyRight)
	if a.Options.Store.Appearance().Mode != "dark" {
		t.Fatal("arrow key did not choose adjacent appearance")
	}
	if err := tt.Click("Theme selector"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Orbit Cobalt"); err != nil {
		t.Fatal(err)
	}
	if a.Options.Store.Appearance().DarkTheme != "orbit-cobalt" {
		t.Fatal("mapped theme did not persist wire value")
	}
	if a.Options.Store.Appearance().LightTheme != "orbit-azure-light" {
		t.Fatal("editing active palette overwrote inactive palette")
	}
	if err := tt.Click("Interface font selector"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Interface font filter"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Inter")
	tt.Key(0, ui.KeyEnter)
	if a.Options.Store.Fonts().UIFamily != "Inter" {
		t.Fatal("mapped font did not persist")
	}
}
func TestShortcutsSourceAdvertisedRowsAndComposerModes(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.selectSettingsPage("shortcuts")
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	text := strings.Join(tt.Texts(), "\n")
	for _, label := range []string{sourceShortcutLabel("secondary-1") + "–" + sourceShortcutLabel("secondary-9"), "COMPOSER", "Queue follow-up (while running)", "Steer current task (while running)", sourceShortcutLabel("alt-enter"), sourceShortcutLabel("secondary-shift-enter")} {
		if !strings.Contains(text, label) {
			t.Fatalf("source help missing %q\n%s", label, text)
		}
	}
	if strings.Contains(text, "Command palette") || strings.Contains(text, "Rename Session") {
		t.Fatal("palette-only actions leaked into source shortcut help")
	}
	if got := sourceAdvertisedChord("AbortRun"); got != "Esc" {
		t.Fatal("alternate binding displayed rather than primary", got)
	}
	a.Prefs.ComposerSendMode = "steer"
	tt.Frame()
	// The two current send conventions are generated from the preference, not
	// separate app command bindings, and remain present after the preference changes.
	if !strings.Contains(strings.Join(tt.Texts(), "\n"), "Queue follow-up (while running)") {
		t.Fatal("composer preference lost help row")
	}
}

func TestSourceWorktreeInlinePreferencesNavigationAndFailure(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.selectSettingsPage("worktrees")
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Worktree directory"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type(" .branches ")
	tt.Frame()
	if a.worktreePrefs.Directory != ".branches" {
		t.Fatal("inline directory did not adopt trimmed source input", a.worktreePrefs.Directory)
	}
	path := filepath.Join(a.Options.Store.RootDir, "worktrees.json")
	if got := workspace.LoadWorktreeConfig(path); got.Directory != ".branches" {
		t.Fatal("inline edit not persisted", got)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	tt.Frame()
	if a.worktreePrefs.Directory != ".branches" {
		t.Fatal("empty transient input overwrote valid path")
	}
	if err := tt.Click("Open Worktrees"); err != nil {
		t.Fatal(err)
	}
	if a.page != "worktrees" {
		t.Fatal("source manage action did not navigate")
	}
	a.Navigate("settings")
	a.settingsPage = "worktrees"
	tt.Frame()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	before := a.worktreePrefs.RunSetupAuto
	if err := tt.Click("Run setup automatically toggle"); err != nil {
		t.Fatal(err)
	}
	if a.worktreePrefs.RunSetupAuto != before {
		t.Fatal("failed save adopted an unpersisted setup policy")
	}
}
func TestSourceRuntimePatchAndPrivacyControls(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "runtime"
	a.Model.Extensions.Patches.Enabled = true
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("RPC patches toggle"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "SetRPCPatches" || actions[0].Data["enabled"] != false {
		t.Fatal(actions)
	}
	if !a.Model.Extensions.Patches.Enabled {
		t.Fatal("patch state changed before actual persisted controller result")
	}
	text := strings.Join(tt.Texts(), "\n")
	for _, extra := range []string{"DEPENDENCIES", "Pi installation", "Choose pi executable…"} {
		if strings.Contains(text, extra) {
			t.Fatal("migration-only runtime row", extra)
		}
	}
	a.rpcPatchSaving = true
	tt.Frame()
	if err := tt.Click("RPC patches toggle"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatal("pending patch save admitted another operation")
	}
	a.settingsPage = "privacy"
	a.nativeSettings.privacyLoaded = true
	a.nativeSettings.temp = native.TempFiles{Directory: "/fixture/orbit-pi", Entries: []native.TempEntry{{Name: "orbit-fixture", Bytes: 2097152, IsDir: true, Modified: time.Now()}}, AIEntries: []native.TempEntry{{Name: "pi-output-fixture", Bytes: 512, Modified: time.Now()}}}
	tt.Frame()
	text = strings.Join(tt.Texts(), "\n")
	for _, label := range []string{"Archive", "30 days", "orbit-fixture", "Folder · 2.0 MB · now", "AI SCRATCH FILES"} {
		if !strings.Contains(text, label) {
			t.Fatal("missing source privacy content", label, text)
		}
	}
	if strings.Contains(text, "LOCAL DATA") {
		t.Fatal("migration-only storage card displaced source temp lists")
	}
	if err := tt.Click("Automatic session cleanup toggle"); err != nil {
		t.Fatal(err)
	}
	if !a.Options.Store.Retention().Enabled {
		t.Fatal("retention switch did not persist")
	}
	if err := tt.Click("Retention action selector"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Delete")
	tt.Key(0, ui.KeyEnter)
	if a.Options.Store.Retention().Mode != "delete" {
		t.Fatal("retention picker did not store mapped value")
	}
}
func TestSourceProviderGridCapabilitiesFilteringAndAboutIdentity(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "providers"
	a.providerState.catalog = workbench.ProviderCatalog{Providers: []workbench.ProviderMetadata{{ID: "a", Name: "Alpha", APIKey: true, CatalogCount: 2, HasCatalogCount: true}, {ID: "b", Name: "Beta", OAuth: true}, {ID: "c", Name: "Gamma", APIKey: true}, {ID: "d", Name: "Delta"}}}
	a.Model.AuthProviders = []agent.AuthProvider{{ID: "b", Name: "Beta", Methods: []agent.AuthMethod{{ID: "device", Label: "Use a device code"}}}}
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	alpha, ok := tt.Find("Provider card a")
	if !ok {
		t.Fatal("alpha card absent")
	}
	beta, _ := tt.Find("Provider card b")
	delta, _ := tt.Find("Provider card d")
	gamma, _ := tt.Find("Provider card c")
	if alpha.Y != beta.Y || alpha.Y != delta.Y || gamma.Y <= alpha.Y || math.Abs(float64(alpha.W-296)) > 1 || math.Abs(float64(beta.X-alpha.X-308)) > 1 {
		t.Fatal("source three-column geometry", alpha, beta, delta, gamma)
	}
	if _, ok := tt.Find("Configure a"); ok {
		t.Fatal("unconfigured builtin advertised custom Configure")
	}
	if err := tt.Click("Sign in b device"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "Login" || actions[0].Value != "b" || actions[0].Data["method"] != "device" {
		t.Fatal(actions)
	}
	if err := tt.Click("API key a"); err != nil {
		t.Fatal(err)
	}
	if !a.secretOpen || a.secretProvider != "a" {
		t.Fatal("API key did not open real secret flow")
	}
	a.secretOpen = false
	tt.Frame()
	if err := tt.Click("Search providers"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Gamma")
	tt.Frame()
	if _, ok := tt.Find("Provider card a"); ok {
		t.Fatal("provider search did not filter")
	}
	if _, ok := tt.Find("Provider card c"); !ok {
		t.Fatal("provider search lost match")
	}
	a.settingsPage = "about"
	tt.Frame()
	text := strings.Join(tt.Texts(), "\n")
	if !strings.Contains(text, "MyGo") || strings.Contains(text, "GPUI") || strings.Contains(text, "Signed updates are not configured") {
		t.Fatal("About misidentified implementation or inserted extra updater card", text)
	}
	title, _ := tt.Find("Settings page title")
	if title.Y != 72 {
		t.Fatal("About source app icon/header footprint absent", title)
	}
}

func TestSettingsReconstructionMeasuredSourceCardGeometry(t *testing.T) {
	for _, scenario := range []struct {
		page   string
		groups []struct {
			key       string
			y, bottom float32
		}
	}{
		{"appearance", []struct {
			key       string
			y, bottom float32
		}{{"settings.theme_and_background", 121, 385}, {"settings.type_and_density", 429, 906}}},
		{"shortcuts", []struct {
			key       string
			y, bottom float32
		}{{"shortcut.group_session", 121, 408}, {"shortcut.group_agent", 452, 596}, {"shortcut.group_composer", 640, 922}}},
		{"worktrees", []struct {
			key       string
			y, bottom float32
		}{{"worktree.settings.manage_section", 121, 191}, {"worktree.settings.location_section", 236, 377}, {"worktree.settings.creation_section", 422, 563}}},
	} {
		t.Run(scenario.page, func(t *testing.T) {
			a := fixture(t)
			a.Navigate("settings")
			a.settingsPage = scenario.page
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			for _, group := range scenario.groups {
				r, ok := tt.Find("Settings group " + a.Tr(group.key))
				if !ok || math.Abs(float64(r.Y-group.y)) > 2 || math.Abs(float64(r.Y+r.H-group.bottom)) > 2 {
					t.Errorf("actual source native card %s expected y %.0f bottom %.0f, mounted %+v", group.key, group.y, group.bottom, r)
				}
			}
		})
	}
}
func TestSourceSettingsPickerEmptyEscapeAndReopen(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "appearance"
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	initial := a.Appearance.DarkTheme
	if err := tt.Click("Theme selector"); err != nil {
		t.Fatal(err)
	}
	tt.Type("no-such-palette")
	tt.Key(0, ui.KeyEnter)
	if a.Appearance.DarkTheme != initial || a.nativeSettings.selectMenu.id != "Theme" {
		t.Fatal("empty picker applied invalid index or dismissed")
	}
	if !tt.HasText("No matches") {
		t.Fatal("empty picker feedback missing")
	}
	tt.Key(0, ui.KeyEscape)
	if a.nativeSettings.selectMenu.id != "" {
		t.Fatal("Escape did not dismiss native picker")
	}
	if err := tt.Click("Theme selector"); err != nil {
		t.Fatal(err)
	}
	if a.nativeSettings.selectMenu.query != "" {
		t.Fatal("reopened picker retained stale filter")
	}
	tt.Type("Cobalt")
	tt.Key(0, ui.KeyEnter)
	if a.Options.Store.Appearance().DarkTheme != "orbit-cobalt" {
		t.Fatal("filtered Enter did not select original option mapping")
	}
}

func TestSourceProviderAuthenticationCardLifecycle(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "providers"
	a.Model.Connected = true
	a.providerState.catalog = workbench.ProviderCatalog{Providers: []workbench.ProviderMetadata{{ID: "p", Name: "Provider", OAuth: true}, {ID: "other", Name: "Other", APIKey: true}}}
	a.Model.AuthProviders = []agent.AuthProvider{{ID: "p", Methods: []agent.AuthMethod{{ID: "device", Label: "Sign in"}}}}
	a.Model.Auth = agent.AuthState{SessionID: "one", Provider: "p", Method: "device", Status: "waiting", UserCode: "FIXTURE-CODE", VerificationURI: "https://example.invalid/device"}
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if _, ok := tt.Find("Sign in p device"); ok {
		t.Fatal("pending authentication left start action visible")
	}
	card, _ := tt.Find("Provider card p")
	status, ok := tt.Find("Authentication status p")
	if !ok || status.X < card.X || status.Y < card.Y || status.Y+status.H > card.Y+card.H {
		t.Fatal("auth lifecycle escaped its provider card", card, status)
	}
	for _, label := range []string{"Open authentication page p", "Copy authentication code p", "Cancel sign in p"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	if len(actions) != 3 || actions[0].Kind != "OpenLink" || actions[0].Value != a.Model.Auth.VerificationURI || actions[1].Kind != "CopyText" || actions[1].Value != "FIXTURE-CODE" || actions[2].Kind != "CancelLogin" {
		t.Fatal(actions)
	}
	if a.Model.Auth.Status != "waiting" {
		t.Fatal("cancel manufactured a terminal state")
	}
	a.Model.Auth.Status, a.Model.Auth.Error = "failed", "Provider rejected this request"
	tt.Frame()
	if _, ok := tt.Find("Cancel sign in p"); ok {
		t.Fatal("terminal failure offered pending cancellation")
	}
	if !strings.Contains(strings.Join(tt.Texts(), "\n"), a.Model.Auth.Error) {
		t.Fatal("provider error absent")
	}
	if err := tt.Click("Retry sign in p"); err != nil {
		t.Fatal(err)
	}
	last := actions[len(actions)-1]
	if last.Kind != "Login" || last.Value != "p" || last.Data["method"] != "device" {
		t.Fatal(last)
	}
	if err := tt.Click("Dismiss authentication p"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("Authentication status p"); ok {
		t.Fatal("dismissal retained terminal card")
	}
	if a.Model.Auth.Status != "failed" {
		t.Fatal("presentation dismissal changed observed state")
	}
	a.Model.Auth.SessionID, a.Model.Auth.Status = "two", "succeeded"
	tt.Frame()
	if _, ok := tt.Find("Authentication status p"); !ok {
		t.Fatal("old dismissal suppressed new session")
	}
	if _, ok := tt.Find("Retry sign in p"); ok {
		t.Fatal("success offered retry")
	}
	if err := tt.Click("Dismiss authentication p"); err != nil {
		t.Fatal(err)
	}
}

func TestSourceRetryCancelPresentationRetainsAuthoritativeBusyState(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "agent"
	a.Model.Connected, a.Model.Retrying = true, true
	var action Action
	a.Options.OnAction = func(v Action) { action = v }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click(a.Tr("settings.abort_retry")); err != nil {
		t.Fatal(err)
	}
	if action.Kind != "AbortRetry" || !a.Model.Retrying {
		t.Fatal("retry cancellation must use RPC, not optimistic snapshot mutation", action)
	}
	// Rust settings.rs5828 clears retry presentation immediately; retain the
	// manager's authoritative busy fact instead of falsifying it locally.
	if _, ok := tt.Find(a.Tr("settings.abort_retry")); ok {
		t.Fatal("optimistic source cancellation did not hide settings action")
	}
	a.Model.Retrying = false
	tt.Frame()
	if _, ok := tt.Find(a.Tr("settings.abort_retry")); ok {
		t.Fatal("idle runtime offered abort retry")
	}
}

func TestSourceProviderUnknownLiveCapabilitiesStayUnknown(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "providers"
	a.Model.ModelCatalog = []agent.ModelChoice{{Provider: "future-provider", ID: "future-model"}}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if _, ok := tt.Find("Provider card future-provider"); !ok {
		t.Fatal("valid live provider was silently omitted")
	}
	if _, ok := tt.Find("API key future-provider"); ok {
		t.Fatal("unknown provider capability invented an API-key action")
	}
}
