package orbitui

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

func usageProviderLabel(id string) string {
	if id == "" {
		return "unknown"
	}
	if id == "chatgpt" {
		return "ChatGPT"
	}
	label := strings.NewReplacer("_", " ", "-", " ").Replace(id)
	bytes := []byte(label)
	if bytes[0] >= 'a' && bytes[0] <= 'z' {
		bytes[0] -= 'a' - 'A'
	}
	return string(bytes)
}
func usageModelLabel(id string) string {
	if _, rest, ok := strings.Cut(id, "/"); ok && rest != "" {
		return rest
	}
	return id
}
func usageModelPair(provider, model string) string {
	return usageModelLabel(model) + " · " + usageProviderLabel(provider)
}
func usageWorkspaceLabel(path string) string {
	if path == "" {
		return "unknown"
	}
	return filepath.Base(path)
}
func usageParentHint(path string) string {
	parent := filepath.Dir(path)
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(parent, home) {
		return "~" + strings.TrimPrefix(parent, home)
	}
	return parent
}
func usageBucketLabel(at time.Time, granularity string, now time.Time) string {
	at = at.Local()
	switch granularity {
	case "hour":
		return at.Format("15:04")
	case "week":
		return "Wk of " + at.Format("Jan 2")
	case "month":
		if at.Year() == now.Year() {
			return at.Format("Jan")
		}
		return at.Format("Jan 2006")
	default:
		if at.Year() == now.Year() {
			return at.Format("Jan 2")
		}
		return at.Format("Jan 2, 2006")
	}
}
func (a *App) usageStamp(at time.Time, granularity string) string {
	at = at.Local()
	switch granularity {
	case "hour":
		return at.Format("Jan 2, 15:04")
	case "week":
		return a.usageText("usage.week_of", "Week of %{date}", map[string]string{"date": at.Format("Jan 2, 2006")})
	case "month":
		return at.Format("January 2006")
	default:
		return at.Format("Jan 2, 2006")
	}
}

func usageSessionModel(r usageAggregate) string {
	if !r.HasModel && r.Model == "" {
		return "—"
	}
	return usageModelPair(r.Provider, r.Model)
}

func (a *App) usageColumnTitle(kind, id, metric string, menu bool) string {
	fallback := titleCase(id)
	if id == "title" {
		fallback = "Session"
	}
	if id == "kind" {
		fallback = "Type"
	}
	if id == "cache-hit" {
		fallback = "Cache hit"
	}
	if !menu && (kind == "breakdown:tools" || id == "share") {
		return fallback
	}
	key := "usage.col_" + id
	switch id {
	case "title", "session":
		key = "usage.col_session"
	case "kind":
		key = "usage.col_type"
	case "cache-hit":
		key = "usage.col_cache_hit"
	case "value":
		key, fallback = "usage.metric_"+metric, titleCase(metric)
		if menu {
			key, fallback = "usage.col_metric", "Metric"
		}
	case "requests", "tokens", "input", "output", "cache", "errors":
		key = "usage.metric_" + id
		if menu && strings.HasPrefix(kind, "breakdown:") && (id == "input" || id == "output") {
			key = "usage.slice_" + id
		}
	}
	return a.text(key, fallback)
}
func (a *App) usageColumns(kind, metric string) []ui.TableColumn {
	cols := usageTableColumns(kind, metric)
	for i := range cols {
		cols[i].Title = a.usageColumnTitle(kind, cols[i].ID, metric, false)
	}
	return cols
}
func (a *App) usageSearchPlaceholder(kind string) string {
	key, fallback := "page.search_sessions", "Search sessions…"
	switch {
	case strings.HasPrefix(kind, "breakdown:"):
		key, fallback = "page.search_breakdown", "Search breakdown…"
	case kind == "series" || kind == "daily":
		key, fallback = "page.search_buckets", "Search buckets…"
	case kind == "failures":
		key, fallback = "page.search_failures", "Search failures…"
	}
	return a.text(key, fallback)
}

func (a *App) usageSessionTitle(title, id string) string {
	if title != "" {
		return title
	}
	runes := []rune(id)
	return a.usageText("usage.session_fallback", "Session %{id}", map[string]string{"id": string(runes[:min(8, len(runes))])})
}
