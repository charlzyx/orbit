package orbitui

// Source command_palette::command_title uses the logical panel-open state,
// not whether a retained panel is currently painted behind another page.
func (a *App) paletteCommandTitle(cmd CommandSpec) string {
	alternate := false
	switch cmd.ID {
	case "ToggleSidebar":
		alternate = a.SidebarVisible
	case "ToggleSidePanel":
		alternate = a.ReviewVisible
	case "ToggleProjectPanel":
		alternate = a.ExplorerVisible
	case "ToggleTerminal":
		alternate = a.Options.Terminal != nil && a.Options.Terminal.IsOpen()
	case "PinSession":
		if a.Model.SessionPath != "" {
			if a.Options.SessionPinned != nil {
				alternate = a.Options.SessionPinned(a.Model.SessionPath)
			} else {
				for _, row := range a.Model.Sessions {
					if row.Path == a.Model.SessionPath {
						alternate = row.Pinned
						break
					}
				}
			}
		}
	}
	return a.Tr(paletteCommandTitleKey(cmd, alternate))
}
func paletteCommandTitleKey(cmd CommandSpec, alternate bool) string {
	if alternate {
		switch cmd.ID {
		case "ToggleSidebar":
			return "command_palette.hide_sidebar"
		case "ToggleSidePanel":
			return "command_palette.hide_side_panel"
		case "ToggleTerminal":
			return "command_palette.hide_terminal"
		case "ToggleProjectPanel":
			return "explorer.hide"
		case "PinSession":
			return "sidebar.unpin_session"
		}
	}
	return cmd.TitleKey
}
