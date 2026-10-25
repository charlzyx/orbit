package orbitui

import "github.com/charlzyx/orbit/internal/agent"

// Login effects are once per session and URL, not once per repaint. Bounded
// memory holds no credentials and prevents old parked sessions reopening tabs.
// The normal native OpenURL path validates http(s) before invoking the browser.
type authBrowserEffects struct{ seen []authBrowserEffect }
type authBrowserEffect struct{ session, url string }

func (s *authBrowserEffects) take(a agent.AuthState) string {
	if a.SessionID == "" || a.URL == "" || a.Status != "waiting" {
		return ""
	}
	key := authBrowserEffect{a.SessionID, a.URL}
	for _, old := range s.seen {
		if old == key {
			return ""
		}
	}
	s.seen = append(s.seen, key)
	if len(s.seen) > 32 {
		copy(s.seen, s.seen[len(s.seen)-32:])
		s.seen = s.seen[:32]
	}
	return a.URL
}
func (x *Controller) openAuthBrowser(a agent.AuthState, open func(string) error) {
	if url := x.authBrowsers.take(a); url != "" {
		x.UI.save(open(url))
	}
}
