package orbitui

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

type quotaVisibilityState struct {
	hidden    map[string]bool
	open      bool
	writeMu   sync.Mutex
	revision  atomic.Uint64
	persisted uint64 // owner-thread observation, not a claim on a failed write
}

func (a *App) loadQuotaVisibility() {
	a.quotaVisibility.hidden = map[string]bool{}
	for _, id := range a.Options.Store.HiddenQuotaProviders() {
		a.quotaVisibility.hidden[id] = true
	}
}
func (a *App) partitionQuotaReports(reports []agent.QuotaReport) (visible, hidden []agent.QuotaReport) {
	for _, report := range reports {
		if a.quotaVisibility.hidden[report.Provider] {
			hidden = append(hidden, report)
		} else {
			visible = append(visible, report)
		}
	}
	return
}
func (a *App) toggleQuotaProvider(id string) {
	if id == "" {
		return
	}
	s := &a.quotaVisibility
	if s.hidden == nil {
		s.hidden = map[string]bool{}
	}
	if s.hidden[id] {
		delete(s.hidden, id)
	} else {
		s.hidden[id] = true
	}
	ids := make([]string, 0, len(s.hidden))
	for id := range s.hidden {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	revision := s.revision.Add(1)
	store := a.Options.Store
	a.Async(func() func() {
		s.writeMu.Lock()
		if revision != s.revision.Load() {
			s.writeMu.Unlock()
			return nil
		}
		err := store.SaveHiddenQuotaProviders(ids)
		s.writeMu.Unlock()
		return func() {
			if revision != s.revision.Load() {
				return
			}
			if err == nil {
				s.persisted = revision
			}
			a.save(err)
		}
	})
}
func (a *App) quotaProviderVisibility(c *ui.Context, id string, hidden bool) {
	label, icon := a.text("session.hide_provider", "Hide provider"), "icons/eye-off.svg"
	if hidden {
		label, icon = a.text("session.show_provider", "Show provider"), "icons/eye.svg"
	}
	if a.iconButton(c, label+" "+id, icon).Size(24, 24).Clicked() {
		a.toggleQuotaProvider(id)
	}
}
func (a *App) quotaPopupReports(c *ui.Context, reports []agent.QuotaReport, now time.Time) {
	visible, hidden := a.partitionQuotaReports(reports)
	for _, report := range visible {
		a.runtimeQuotaCardWithVisibility(c.Key("visible:"+report.Provider), report, now, func() { a.quotaProviderVisibility(c, report.Provider, false) })
	}
	if len(hidden) == 0 {
		return
	}
	icon := "icons/chevron-right.svg"
	if a.quotaVisibility.open {
		icon = "icons/chevron-down.svg"
	}
	label := strings.ReplaceAll(a.text("session.hidden_providers", "Hidden (%{count})"), "%{count}", fmt.Sprint(len(hidden)))
	if a.button(c, label, icon, a.quotaVisibility.open).Label("Hidden providers").Clicked() {
		a.quotaVisibility.open = !a.quotaVisibility.open
	}
	if !a.quotaVisibility.open {
		return
	}
	for _, report := range hidden {
		ui.Column(c.Key("hidden:" + report.Provider)).Opacity(.6).Children(func() {
			a.runtimeQuotaCardWithVisibility(c, report, now, func() { a.quotaProviderVisibility(c, report.Provider, true) })
		})
	}
}

func (a *App) quotaProviderCount(count int) string {
	return strings.ReplaceAll(a.text("session.provider_count", "%{count} providers"), "%{count}", fmt.Sprint(count))
}
