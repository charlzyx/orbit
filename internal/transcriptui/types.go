// Package transcriptui projects Pi messages into the pinned Orbit transcript
// contract and renders them using native MyGo controls. No model work or I/O
// occurs here; missing provider facts remain absent.
package transcriptui

import (
	"encoding/json"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/highlight"
	"github.com/egoist/mygo/ui"
)

const SourcePin = "299a32bc87fe3d4851bdb902d9c03dcf9bf99cac"

type ToolFacts struct {
	Truncated               bool
	OutputLines, TotalLines *uint64
}
type ToolCall struct {
	ID, Name, Path, Summary   string
	Arguments, Output         json.RawMessage
	Running, Failed, Complete bool
	Facts                     ToolFacts
	Added, Removed            uint64
}
type Usage struct {
	Input, Output, CacheRead, CacheWrite, Total uint64
	Cost                                        *float64
}
type Step struct {
	ID, Text, Thinking               string
	Tools                            []ToolCall
	Usage                            *Usage
	Timestamp                        time.Time
	Duration                         *time.Duration // Pi durationMs only; no inferred reload clock.
	ThinkingDuration, GenerationTime *time.Duration
	RequestStartedAt                 time.Time
	Streaming                        bool
}
type SummaryRow struct {
	Branch       bool
	Text         string
	TokensBefore *uint64
}
type Row struct {
	ID, Role, Error    string
	Steps              []Step
	Images             []agent.Image
	Summary            *SummaryRow
	Streaming, Aborted bool
	FinishedAt         time.Time
	// Elapsed is a measured live turn clock supplied by a host. Project does
	// not derive it from timestamps, which include tools and arbitrary gaps.
	Elapsed *time.Duration
}
type ChangedFile struct {
	Path           string
	Added, Removed uint64
}

type Colors struct {
	Text, Muted, Accent, Border, Surface, Code, Error, Warning, Success, Selection ui.Color
	Syntax                                                                         map[highlight.Class]ui.Color
}

// Options contains host services, not fake presentation data. Text uses the
// source locale keys and named %{variables}. A nil callback uses English.
type Options struct {
	Colors             Colors
	Font, CodeFont     string
	FontSize, CodeSize float32
	Text               func(key, fallback string) string
	Icon               func(c *ui.Context, path string, size float32, color ui.Color)
	Markdown           func(c *ui.Context, text string, streaming bool)
	Copy               func(string)
	Image              func(c *ui.Context, image agent.Image)
	Review             func()
	Workspace          string
	SuppressFooter     bool
	TimestampLabel     string
	Now                func() time.Time
}

// State is per-session view state. Reset it when switching transcript identity,
// not on every streamed snapshot. Keys include stable row and tool identities.
type State struct {
	session        string
	open           map[string]bool
	chosen         map[string]bool
	copied         map[string]time.Time
	clocks         map[string]turnClock
	thoughtScrolls map[string]*ui.ScrollState
}

func (s *State) Reset(session string) {
	s.session = session
	s.open = make(map[string]bool)
	s.chosen = make(map[string]bool)
	s.copied = make(map[string]time.Time)
	s.clocks = make(map[string]turnClock)
	s.thoughtScrolls = make(map[string]*ui.ScrollState)
}
func (s *State) Session(session string) {
	if s.open == nil || s.session != session {
		s.Reset(session)
	}
}
func (s *State) Open(key string, fallback bool) bool {
	if s.chosen[key] {
		return s.open[key]
	}
	return fallback
}
func (s *State) Toggle(key string, fallback bool) {
	if s.open == nil {
		s.Reset(s.session)
	}
	s.open[key] = !s.Open(key, fallback)
	s.chosen[key] = true
}
