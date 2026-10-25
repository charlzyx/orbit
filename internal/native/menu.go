package native

import (
	"runtime"
	"strings"

	"github.com/egoist/mygo"
)

// MenuOptions keeps native menu actions on the same command path as the UI.
// Dispatch runs on the native main thread. Rebuild after a locale change.
type MenuOptions struct {
	AppName         string
	Dispatch        func(command string)
	Text            func(key, fallback string) string
	CanCheckUpdates bool
}

func NativeMenu(o MenuOptions) *mygo.Menu { return mygo.NewMenu(NativeMenuTemplate(o, runtime.GOOS)) }

// NativeMenuTemplate mirrors main.rs app_menus; OS editing roles deliberately
// stay native so text fields and terminal selection own cut/copy/paste.
func NativeMenuTemplate(o MenuOptions, goos string) []*mygo.MenuItem {
	name := o.AppName
	if name == "" {
		name = "Orbit"
	}
	tr := func(key, fallback string) string {
		if o.Text != nil {
			fallback = o.Text(key, fallback)
		}
		return strings.ReplaceAll(fallback, "%{app}", name)
	}
	action := func(id, key, fallback, accelerator string) *mygo.MenuItem {
		item := &mygo.MenuItem{ID: id, Label: tr(key, fallback), Accelerator: accelerator, Disabled: o.Dispatch == nil}
		if o.Dispatch != nil {
			item.Click = func(*mygo.MenuItem, *mygo.Window) { o.Dispatch(id) }
		}
		return item
	}
	update := action("CheckForUpdates", "menu.check_for_updates", "Check for updates…", "CmdOrCtrl+Shift+U")
	update.Disabled = update.Disabled || !o.CanCheckUpdates
	appItems := []*mygo.MenuItem{action("OpenAbout", "menu.about", "About "+name, ""), update, mygo.Separator(), action("OpenSettings", "menu.settings", "Settings…", "CmdOrCtrl+,")}
	if goos == "darwin" {
		appItems = append(appItems, mygo.Separator(), &mygo.MenuItem{Role: mygo.RoleServices, Label: tr("menu.services", "Services")})
	}
	appItems = append(appItems, mygo.Separator(), action("Quit", "menu.quit", "Quit "+name, "CmdOrCtrl+Q"))
	app := &mygo.MenuItem{Label: name, Submenu: appItems}
	if goos == "darwin" {
		app.Role = mygo.RoleAppMenu
	}
	edit := []*mygo.MenuItem{{Role: mygo.RoleCut, Label: tr("menu.cut", "Cut")}, {Role: mygo.RoleCopy, Label: tr("menu.copy", "Copy")}, {Role: mygo.RolePaste, Label: tr("menu.paste", "Paste")}, {Role: mygo.RoleSelectAll, Label: tr("menu.select_all", "Select all")}}
	return []*mygo.MenuItem{app, {Label: tr("menu.file", "File"), Submenu: []*mygo.MenuItem{action("NewSession", "menu.new_task", "New task", "CmdOrCtrl+N"), action("RefreshSessions", "menu.refresh_sessions", "Refresh sessions", "CmdOrCtrl+R")}}, {Label: tr("menu.edit", "Edit"), Submenu: edit}, {Label: tr("menu.view", "View"), Submenu: []*mygo.MenuItem{action("ToggleCommandPalette", "menu.command_palette", "Command palette…", "CmdOrCtrl+P"), action("ToggleSearch", "menu.find_in_transcript", "Find in transcript…", "CmdOrCtrl+F"), mygo.Separator(), action("ToggleSidebar", "menu.toggle_sidebar", "Toggle sidebar", "CmdOrCtrl+B"), action("ToggleTerminal", "menu.toggle_terminal", "Toggle terminal", "CmdOrCtrl+J"), action("ToggleProjectPanel", "explorer.toggle", "Toggle project panel", "CmdOrCtrl+Shift+E"), mygo.Separator(), action("ToggleUsage", "menu.usage", "Usage", "CmdOrCtrl+U")}}}
}
