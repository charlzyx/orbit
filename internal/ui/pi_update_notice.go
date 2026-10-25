package orbitui

import (
	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/native"
)

// notifyPiUpdateStatus maps only source-equivalent informational events. A Go
// deferred event implies a busy run only with the explicit busy-timeout reason.
func (a *App) notifyPiUpdateStatus(event native.PiUpdateEvent) {
	var key string
	switch event.State {
	case "deferred":
		if event.Deferral != native.PiUpdateDeferralBusyTimeout {
			return
		}
		key = "pi_update.available_next_launch"
	case "available":
		key = "pi_update.available_background"
	case string(native.PiUnsupportedManager):
		key = "pi_update.unsupported_package_manager"
	case string(native.PiUnknownVersion):
		key = "pi_update.unsupported_unknown"
	default:
		return
	}
	a.Notify(orbitassets.Translate(a.catalog, a.english, key, map[string]string{"version": event.Latest}), false)
}
