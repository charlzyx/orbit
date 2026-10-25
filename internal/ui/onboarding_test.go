package orbitui

import (
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"testing"
)

func setupDeps(pi, node, git bool) []native.Dependency {
	return []native.Dependency{{Name: "pi", Bin: "pi", Required: true, Installed: pi, InstallHint: "fixture install pi", DetailKey: "onboarding.pi_detail"}, {Name: "Node.js", Bin: "node", Required: true, Installed: node, InstallHint: "fixture install node", DetailKey: "onboarding.node_detail"}, {Name: "git", Bin: "git", Installed: git, InstallHint: "fixture install git", DetailKey: "onboarding.git_detail"}}
}
func TestOnboardingMissingRequiredReplacesComposer(t *testing.T) {
	for _, missing := range []string{"pi", "node"} {
		t.Run(missing, func(t *testing.T) {
			a := fixture(t)
			a.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(missing != "pi", missing != "node", false), facts: []native.HostFact{{Label: "Platform", Value: "fixture OS · aarch64"}, {Label: "Orbit config", Value: "/fixture/.orbit-pi"}}}
			tt := ui.NewTester(a.View, 1200, 800)
			if _, ok := tt.Find(a.Tr("view.set_up_orbit")); !ok {
				t.Fatal("missing first-run setup page")
			}
			if _, ok := tt.Find("Message"); ok {
				t.Fatal("missing required runtime exposed composer")
			}
			if _, ok := tt.Find("fixture OS · aarch64"); !ok {
				t.Fatal("missing host facts")
			}
			if _, ok := tt.Find("/fixture/.orbit-pi"); !ok {
				t.Fatal("missing config path")
			}
		})
	}
}
func TestOnboardingOptionalGitDoesNotBlockChat(t *testing.T) {
	a := fixture(t)
	a.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(true, true, false)}
	tt := ui.NewTester(a.View, 1200, 800)
	if _, ok := tt.Find("Message"); !ok {
		t.Fatal("optional git blocked composer")
	}
}
func TestOnboardingCopyHintDoesNotExecuteAndSettingsAccessible(t *testing.T) {
	a := fixture(t)
	a.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Copy install command pi"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "fixture install pi" || len(actions) != 0 {
		t.Fatal("copy command executed operation", tt.Clipboard(), actions)
	}
	a.command("OpenSettings", "")
	tt.Frame()
	if a.page != "settings" {
		t.Fatal("setup trapped settings")
	}
	a.Navigate("chat")
	tt.Frame()
	if _, ok := tt.Find(a.Tr("view.set_up_orbit")); !ok {
		t.Fatal("missing runtime bypassed setup after Settings")
	}
}

func TestOnboardingInitialCheckingAndReadyPreservesDraft(t *testing.T) {
	a := fixture(t)
	a.Draft = "kept before setup"
	a.setup = onboardingState{enabled: true, refreshing: true}
	tt := ui.NewTester(a.View, 1200, 800)
	if _, ok := tt.Find(a.Tr("common.checking")); !ok {
		t.Fatal("initial check has no loading feedback")
	}
	if _, ok := tt.Find("Message"); ok {
		t.Fatal("initial check exposed composer")
	}
	a.setup.loaded = true
	a.setup.refreshing = false
	a.setup.dependencies = setupDeps(true, true, true)
	tt.Frame()
	if _, ok := tt.Find("Message"); !ok || a.Draft != "kept before setup" {
		t.Fatal("ready transition lost draft or did not return chat")
	}
}

func TestOnboardingCapture(t *testing.T) {
	out := os.Getenv("ORBIT_SETUP_CAPTURE")
	if out == "" {
		t.Skip("optional native CPU setup capture")
	}
	a := fixture(t)
	deps := setupDeps(false, true, false)
	deps[0].InstallHint = native.DependencyInstallHint("darwin", "pi")
	deps[1].Version = "v24.0.0"
	deps[2].InstallHint = native.DependencyInstallHint("darwin", "git")
	a.setup = onboardingState{enabled: true, loaded: true, dependencies: deps, facts: []native.HostFact{{Label: "Platform", Value: "macOS fixture · aarch64"}, {Label: "pi sessions", Value: "~/.pi/agent/sessions", Note: "not created yet"}, {Label: "Orbit config", Value: "~/.orbit-pi"}, {Label: "Pi install", Value: "missing", Alert: true}, {Label: "Pi path", Value: "—"}, {Label: "Pi version", Value: "—"}}}
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

// Rust's root accepts setup drops and global task shortcuts, but settings
// rejects drops and a hidden composer does not receive Enter/steer/text input.
func TestOnboardingInputRetainsDropWithoutSubmittingHiddenDraft(t *testing.T) {
	for _, page := range []string{"chat", "settings"} {
		t.Run(page, func(t *testing.T) {
			a := fixture(t)
			a.Draft = "preserved "
			var actions []Action
			a.Options.OnAction = func(x Action) { actions = append(actions, x) }
			tt := ui.NewTester(a.View, 1200, 800)
			if err := tt.Click("Message"); err != nil {
				t.Fatal(err)
			}
			a.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
			a.Navigate(page)
			tt.Frame()
			tt.Type("unfocused typing")
			for _, mods := range []ui.Modifiers{0, ui.Alt, ui.Cmd | ui.Shift} {
				tt.Key(mods, ui.KeyEnter)
			}
			tt.DropFiles(600, 400, "/fixture/retained.txt")
			tt.Frame()
			if len(actions) != 0 {
				t.Fatal("setup input submitted or invoked runtime", actions)
			}
			a.setup.dependencies = setupDeps(true, true, true)
			a.Navigate("chat")
			tt.Frame()
			want := "preserved "
			if page == "chat" {
				want += "/fixture/retained.txt "
			}
			if a.Draft != want {
				t.Fatalf("drop/hidden key behavior: got %q want %q", a.Draft, want)
			}
			if len(actions) != 0 {
				t.Fatal("ready transition submitted retained drop", actions)
			}
		})
	}
}
func TestOnboardingGlobalSettingsAndTaskKeysRemainUserDriven(t *testing.T) {
	a := fixture(t)
	a.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt := ui.NewTester(a.View, 1200, 800)
	for _, command := range []string{"OpenSettings", "NewSession"} {
		found := false
		for _, b := range Bindings {
			if b.Command == command && b.Context == "" {
				mods, key, ok := keyChord(b.Chord)
				if !ok {
					continue
				}
				tt.Key(mods, key)
				tt.Frame()
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("no global binding for %s", command)
		}
		if command == "OpenSettings" && a.page != "settings" {
			t.Fatal("setup trapped Settings shortcut")
		}
	}
	if len(actions) != 1 || actions[0].Kind != "NewSession" {
		t.Fatal("global task key not source-aligned", actions)
	}
}
