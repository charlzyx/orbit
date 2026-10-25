package orbitui

import (
	"errors"
	"strings"
)

// The pinned source has a legacy pi /login escape hatch only when auth RPC is
// unsupported. This does not start sign-in during page load or infer support
// from a missing response. Native OpenTerminalProgram passes literal arguments.
func (x *Controller) loginProviderTerminal(id string, open func(string, ...string) error) error {
	if x.UI.Model.Auth.Support != "unsupported" {
		return errors.New("terminal sign-in is available only when Pi authentication RPC is unsupported")
	}
	if id == "" {
		return errors.New("provider id is required")
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
			return errors.New("invalid provider id for terminal sign-in")
		}
	}
	name := ""
	for _, p := range x.UI.providerState.catalog.Providers {
		if p.ID == id && p.OAuth {
			name = p.Name
			if name == "" {
				name = id
			}
			break
		}
	}
	if name == "" {
		return errors.New("provider did not advertise OAuth support")
	}
	binary := x.Agent.PiBinary()
	if binary == "" {
		binary = "pi"
	}
	if err := open(binary, "/login", id); err != nil {
		return err
	}
	x.UI.Notify(strings.ReplaceAll(x.UI.Tr("settings.finish_sign_in_terminal"), "%{name}", name), false)
	return nil
}
