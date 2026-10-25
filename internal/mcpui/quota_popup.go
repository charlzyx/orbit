package mcpui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

// OpenQuotaPopup follows Rust's 20-second stale-probe rule. Config/probe work
// remains in the existing State workers; selecting a tab never invents status.
// PrepareQuotaPopup loads configured servers asynchronously, without probing
// endpoints merely to decide whether the MCP-only entry point should exist.
func (s *State) PrepareQuotaPopup() {
	if !s.closed && !s.loaded && !s.Busy {
		s.refresh(false, false)
	}
}
func (s *State) OpenQuotaPopup(now time.Time) {
	if s.closed || s.Probing {
		return
	}
	stale := !s.loaded || s.checkedAt.IsZero() || now.Sub(s.checkedAt) >= 20*time.Second
	if !stale {
		return
	}
	if s.Busy {
		s.quotaProbeRequested = true
		return
	}
	s.Refresh()
}
func (s *State) quotaCounts(enabledOnly bool) (connected, total, unknown int) {
	for _, server := range s.Servers {
		if enabledOnly && !enabled(server.Definition) {
			continue
		}
		total++
		if !enabled(server.Definition) {
			continue
		}
		switch s.Runtime[server.Name].State {
		case "connected":
			connected++
		case "", "unknown":
			unknown++
		}
	}
	return
}

// QuotaChip only exposes the MCP-only entry point when an enabled server exists,
// like Rust. Unprobed servers are explicitly unknown, never disconnected zeros.
func (s *State) QuotaChip() (string, bool) {
	connected, total, unknown := s.quotaCounts(true)
	if total == 0 {
		return "", false
	}
	if unknown == total {
		return "MCP · " + s.label("mcp.status_unknown", "Not checked"), true
	}
	label := fmt.Sprintf("MCP %d/%d", connected, total)
	if unknown > 0 {
		label += " · " + s.label("mcp.status_unknown", "Not checked")
	}
	return label, true
}
func (s *State) QuotaSummary() string {
	connected, total, unknown := s.quotaCounts(false)
	if total == 0 {
		return ""
	}
	if connected == 0 && unknown > 0 {
		return s.label("mcp.not_checked", "Not checked yet")
	}
	text := s.label("mcp.popup_connected", "%{connected}/%{total} connected", "connected", fmt.Sprint(connected), "total", fmt.Sprint(total))
	if unknown > 0 {
		text += " · " + s.label("mcp.status_unknown", "Not checked")
	}
	return text
}
func (s *State) QuotaHeader(c *ui.Context) {
	a := s.appearance(c)
	if label := s.QuotaSummary(); label != "" {
		ui.Text(c, label).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
	}
	if s.iconButton(c, "quota-refresh", "↻", s.label("mcp.refresh", "Refresh MCP")).Disabled(s.Busy).Clicked() {
		s.Refresh()
	}
}
func (s *State) QuotaPanel(c *ui.Context, openSettings func(string)) {
	a := s.appearance(c)
	ui.Column(c.Key("quota-mcp-body")).Gap(a.Space(6)).Children(func() {
		if s.Error != "" {
			ui.Text(c, s.Error).FontSize(a.Px(11)).TextColor(a.native.Danger)
		}
		for _, message := range append(append([]string{}, s.Errors...), s.probeErrors...) {
			ui.Text(c, message).FontSize(a.Px(11)).TextColor(a.native.Danger)
		}
		if s.probeNote != "" {
			ui.Text(c, s.probeNote).FontSize(a.Px(10)).TextColor(a.native.TextMuted)
		}
		if len(s.Servers) == 0 {
			if s.Busy && !s.loaded {
				ui.Text(c, s.label("mcp.checking", "Checking connections…"))
			} else if len(s.Errors) == 0 {
				ui.Text(c, s.label("mcp.empty_title", "No MCP servers")).FontWeight(500)
				ui.Text(c, s.label("mcp.empty_hint", "Add a server in Settings to get started.")).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
			}
		}
		for _, server := range s.Servers {
			s.quotaRow(c, server, openSettings)
		}
		ui.Row(c.Key("footer")).Gap(a.Space(6)).AlignItems(ui.Center).Children(func() {
			if s.button(c, "quota-open-settings", s.label("mcp.popup_open_settings", "Open MCP settings"), false).Clicked() && openSettings != nil {
				openSettings("")
			}
			status := s.label("mcp.not_checked", "Not checked yet")
			if s.Probing {
				status = s.label("mcp.checking", "Checking connections…")
			} else if !s.checkedAt.IsZero() {
				status = s.label("mcp.checked_ago", "Checked %{seconds}s ago", "seconds", fmt.Sprint(max(1, int(time.Since(s.checkedAt).Seconds()))))
			}
			ui.Text(c, status).FontSize(a.Px(10)).TextColor(a.native.TextMuted)
		})
	})
}
func (s *State) quotaRow(c *ui.Context, server workbench.MCPServer, openSettings func(string)) {
	a := s.appearance(c)
	runtime := s.Runtime[server.Name]
	if !enabled(server.Definition) {
		runtime = Runtime{State: "disabled"}
	}
	ui.Row(c.Key("quota-mcp-row:" + server.Name)).Padding(a.Space(6)).Gap(a.Space(6)).AlignItems(ui.Center).Children(func() {
		facts := ui.ButtonBase(c.Key("open")).Label("MCP server " + server.Name).Grow(1).MinWidth(0)
		facts.Children(func() {
			ui.Column(c).Gap(a.Space(2)).Children(func() {
				ui.Text(c, server.Name).FontWeight(500).SingleLine()
				parts := []string{s.label("mcp.exposure_"+strings.ReplaceAll(exposure(server.Definition), "-", "_"), exposure(server.Definition)), s.label("mcp.scope_"+server.Scope, server.Scope)}
				if oauthCandidate(server) {
					parts = append(parts, s.label("mcp.oauth", "OAuth"))
				}
				ui.Text(c, strings.Join(parts, " · ")).FontSize(a.Px(10)).TextColor(a.native.TextMuted)
				if runtime.Error != "" {
					ui.Text(c, runtime.Error).FontSize(a.Px(10)).TextColor(a.native.Danger).MaxLines(2)
				}
			})
		})
		if facts.Clicked() && openSettings != nil {
			openSettings(server.Name)
		}
		if s.authTarget == server.Name {
			ui.Text(c, s.label("mcp.waiting_for_browser", "Waiting for browser…")).FontSize(a.Px(10))
			if s.button(c, "cancel-auth", s.label("mcp.cancel", "Cancel"), false).Label("Cancel MCP sign-in "+server.Name).Clicked() && s.cancel != nil {
				s.cancel()
			}
		} else if runtime.State == "connected" {
			ui.Text(c, s.label("mcp.popup_tools", "%{count} tools", "count", fmt.Sprint(len(runtime.Tools)))).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
		} else if runtime.State == "needs-auth" && oauthCandidate(server) {
			if s.button(c, "sign-in", s.label("mcp.sign_in", "Sign in"), true).Label("Sign in to MCP " + server.Name).Disabled(s.Busy && !s.Probing).Clicked() {
				s.auth(server.Name, true)
			}
		} else {
			label, color := s.runtimeLabel(runtime.State, a)
			ui.Text(c, label).FontSize(a.Px(11)).TextColor(color)
		}
	})
}
