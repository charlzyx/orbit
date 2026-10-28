package native

import (
	"github.com/egoist/mygo"
	"testing"
)

func TestNativeMenuCommandParity(t *testing.T) {
	for _, goos := range []string{"darwin", "linux", "windows"} {
		var commands []string
		template := NativeMenuTemplate(MenuOptions{CanCheckUpdates: true, Dispatch: func(id string) { commands = append(commands, id) }, Text: func(key, fallback string) string {
			if key == "menu.about" {
				return "About %{app}"
			}
			return fallback
		}}, goos)
		if len(template) != 4 {
			t.Fatal("menu count differs from source app_menus", len(template))
		}
		seen := map[string]bool{}
		var walk func([]*mygo.MenuItem)
		walk = func(items []*mygo.MenuItem) {
			for _, i := range items {
				if i.ID != "" {
					seen[i.ID] = true
					if i.Click == nil || i.Disabled {
						t.Fatalf("inert %s", i.ID)
					}
					i.Click(i, nil)
				}
				walk(i.Submenu)
			}
		}
		walk(template)
		for _, want := range []string{"OpenAbout", "CheckForUpdates", "OpenSettings", "Quit", "NewSession", "RefreshSessions", "ToggleCommandPalette", "ToggleSearch", "ToggleSidebar", "ToggleTerminal", "ToggleProjectPanel", "ToggleUsage"} {
			if !seen[want] {
				t.Fatal(goos, want)
			}
		}
		if len(commands) != 12 || template[0].Submenu[0].Label != "About Orbit" {
			t.Fatal(commands, template[0].Submenu[0])
		}
		if template[2].Submenu[1].Role != mygo.RoleCopy {
			t.Fatal("copy not native")
		}
	}
	disabled := NativeMenuTemplate(MenuOptions{}, "linux")
	if !disabled[0].Submenu[1].Disabled {
		t.Fatal("keyless updater enabled")
	}
}

func TestNativeEditMenuUsesContentFirstRoleDispatch(t *testing.T) {
	for _, goos := range []string{"darwin", "linux", "windows"} {
		menu := NativeMenuTemplate(MenuOptions{}, goos)[2].Submenu
		for i, role := range []mygo.MenuRole{mygo.RoleCut, mygo.RoleCopy, mygo.RolePaste, mygo.RoleSelectAll} {
			if menu[i].Role != role || menu[i].Click != nil || menu[i].ID != "" {
				t.Fatalf("%s edit %d replaced native role dispatch: %+v", goos, i, menu[i])
			}
		}
	}
}

func TestGoPiMenuDisplayNamePreservesNativeEditRoles(t *testing.T) {
	menu := NativeMenuTemplate(MenuOptions{AppName: "GoPi"}, "darwin")
	if menu[0].Label != "GoPi" || menu[0].Submenu[0].Label != "About GoPi" {
		t.Fatal("display name", menu[0])
	}
	if menu[2].Submenu[1].Role != mygo.RoleCopy {
		t.Fatal("copy role changed")
	}
}
