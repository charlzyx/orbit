package orbitui

// Match the actual root/content paint branches, not retained dock-open flags.
// Reuse each model's existing asynchronous admission/coalescing/owner guards;
// No new agent RPC, runtime Apply/restart, install, or hidden-page refresh.
// Existing MCP refresh may perform its source-defined read-side probe.
func (x *Controller) refreshVisibleModels() {
	a := x.UI
	if a.page == "settings" {
		switch a.settingsPage {
		case "mcp":
			a.loadSettings("mcp") // original read + probe; never Apply
		case "plugins":
			// Rust refresh_plugins only discovers installed packages. The
			// page's RefreshPlugins also checks updates, so do not call it.
			x.Resources.Refresh()
		}
		return
	}
	if a.setupBlocked() {
		return
	}
	if a.page == "usage" {
		a.refreshUsage()
		return
	}
	if a.ReviewVisible {
		a.RefreshReview()
	}
}
