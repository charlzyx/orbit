package orbitui

import (
	"github.com/charlzyx/orbit/internal/mcpui"
	"github.com/egoist/mygo/ui"
	"time"
)

type quotaMCPState struct {
	state *mcpui.State
	tab   string
}

// SetQuotaMCP binds the actual application-owned configuration/probe state.
func (a *App) SetQuotaMCP(state *mcpui.State) {
	a.quotaMCP.state = state
	if state != nil {
		state.PrepareQuotaPopup()
	}
}
func (a *App) selectQuotaTab(tab string) {
	a.quotaMCP.tab = tab
	if tab == "mcp" && a.quotaMCP.state != nil {
		if a.Model.Workspace != "" {
			a.quotaMCP.state.SetWorkspace(a.Model.Workspace)
		}
		a.quotaMCP.state.OpenQuotaPopup(time.Now())
	}
}
func (a *App) quotaTabs(c *ui.Context) {
	for _, tab := range []struct{ id, key, fallback string }{{"providers", "session.tab_providers", "Providers"}, {"mcp", "session.tab_mcp", "MCP"}} {
		if a.button(c, a.text(tab.key, tab.fallback), "", a.quotaMCP.tab == tab.id || a.quotaMCP.tab == "" && tab.id == "providers").Height(28).Clicked() {
			a.selectQuotaTab(tab.id)
		}
	}
}
func (a *App) openQuotaMCPSettings(name string) {
	a.quotaOpen = false
	if s := a.quotaMCP.state; s != nil {
		s.Selected = name
	}
	a.Navigate("settings")
	a.selectSettingsPage("mcp")
}
