package orbitui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// These two numerical registers follow the frozen Usage formatter: compact
// three significant digits for the board, exact grouped values for readouts.
func compactCount(n uint64) string {
	if n < 1000 {
		return strconv.FormatUint(n, 10)
	}
	v, unit := float64(n)/1e3, "K"
	if n >= 1_000_000_000 {
		v, unit = float64(n)/1e9, "B"
	} else if n >= 1_000_000 {
		v, unit = float64(n)/1e6, "M"
	}
	digits := 0
	if v < 10 {
		digits = 2
	} else if v < 100 {
		digits = 1
	}
	return usageTrim(fmt.Sprintf("%.*f", digits, v)) + unit
}
func usageTrim(v string) string {
	if strings.Contains(v, ".") {
		return strings.TrimRight(strings.TrimRight(v, "0"), ".")
	}
	return v
}
func usageExact(n uint64) string {
	raw := strconv.FormatUint(n, 10)
	var b strings.Builder
	for i, ch := range raw {
		if i > 0 && (len(raw)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	return b.String()
}
func usageCount(n uint64) string {
	if n < 10000 {
		return usageExact(n)
	}
	return compactCount(n)
}
func usageCost(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "—"
	}
	if v <= 0 {
		return "$0.00"
	}
	if v < .01 {
		return "<$0.01"
	}
	text := fmt.Sprintf("%.2f", v)
	parts := strings.Split(text, ".")
	var b strings.Builder
	for i, ch := range parts[0] {
		if i > 0 && (len(parts[0])-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	return "$" + b.String() + "." + parts[1]
}
func usagePercent(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "—"
	}
	m := math.Abs(v)
	if m > 0 && m < .01 {
		return "<0.01%"
	}
	if m < 1 {
		return usageTrim(fmt.Sprintf("%.2f", v)) + "%"
	}
	if v < 100 && fmt.Sprintf("%.1f", v) == "100.0" {
		return "99.9%"
	}
	return usageTrim(fmt.Sprintf("%.1f", v)) + "%"
}
func usageDuration(seconds float64) string {
	ms := seconds * 1000
	if math.IsNaN(ms) || math.IsInf(ms, 0) || ms < 0 {
		return "—"
	}
	switch {
	case ms < 1000:
		return fmt.Sprintf("%.0fms", ms)
	case ms < 10000:
		return fmt.Sprintf("%.1fs", seconds)
	case ms < 60000:
		return fmt.Sprintf("%.0fs", seconds)
	case ms < 3600000:
		return fmt.Sprintf("%dm %ds", int64(seconds)/60, int64(seconds)%60)
	default:
		return fmt.Sprintf("%dh %02dm", int64(seconds)/3600, int64(seconds)/60%60)
	}
}
func usageSpan(span time.Duration) string {
	if span <= 0 {
		return "—"
	}
	s := span.Seconds()
	switch {
	case s < 60:
		return fmt.Sprintf("%.0fs", s)
	case s < 3600:
		return fmt.Sprintf("%dm %.0fs", int64(s)/60, math.Mod(s, 60))
	case s < 86400:
		return fmt.Sprintf("%dh %02dm", int64(s)/3600, int64(s)/60%60)
	default:
		return fmt.Sprintf("%dd %dh", int64(s)/86400, int64(s)/3600%24)
	}
}
