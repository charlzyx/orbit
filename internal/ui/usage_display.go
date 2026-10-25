package orbitui

import (
	"os"
	"path/filepath"
	"strings"
	"time"
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
	if model == "" {
		return "—"
	}
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
