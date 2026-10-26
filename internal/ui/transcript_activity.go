package orbitui

import (
	"bytes"
	"github.com/charlzyx/orbit/internal/transcriptui"
)

// One owner-scoped bit, not another copy/hash of the transcript. Authoritative
// snapshots and projected rows are immutable; Manager.Snapshot deep-copies its
// mutable slices. Only compare on snapshot adoption while an initialized reader
// is away from the tail and has not already been notified.
type transcriptActivity struct {
	owner               inlineImageToken
	initialized, unread bool
}

func (a *App) observeTranscriptActivity(previous []Message) {
	s := &a.transcriptActivity
	owner := a.inlineImageToken()
	if s.owner != owner {
		*s = transcriptActivity{owner: owner}
		return
	}
	if !s.initialized {
		s.unread = false
		return
	}
	atEnd := a.transcript.AtEnd()
	if atEnd && !s.unread {
		return
	}
	rows := a.Model.Messages
	// History replacement/deletion/reordering is a reset, not a new event. Rust
	// resets its scroller for history loads; append/content events call activity.
	if len(rows) < len(previous) {
		s.unread = false
		return
	}
	for i := range previous {
		if previous[i].ID != rows[i].ID || previous[i].Role != rows[i].Role {
			s.unread = false
			return
		}
	}
	if s.unread && !atEnd {
		return
	}
	if len(rows) > len(previous) {
		s.unread = !atEnd
		return
	}
	for i := range rows {
		if a.transcriptMessageChanged(previous[i], rows[i]) {
			s.unread = !atEnd
			return
		}
	}
}
func (a *App) presentTranscriptActivity() bool {
	s := &a.transcriptActivity
	owner := a.inlineImageToken()
	if s.owner != owner {
		*s = transcriptActivity{owner: owner}
	}
	s.initialized = true
	return s.unread
}
func (a *App) transcriptMessageChanged(x, y Message) bool {
	if x.Text != y.Text || x.Thinking != y.Thinking || x.Error != y.Error || x.ToolName != y.ToolName || x.ToolOutput != y.ToolOutput || x.Streaming != y.Streaming {
		return true
	}
	if len(x.Images) != len(y.Images) {
		return true
	}
	for i := range x.Images {
		if x.Images[i] != y.Images[i] {
			return true
		}
	}
	if x.Logical == nil || y.Logical == nil {
		return x.Logical != nil || y.Logical != nil
	}
	arow, brow := x.Logical, y.Logical
	if arow.Aborted != brow.Aborted || arow.Error != brow.Error || len(arow.Steps) != len(brow.Steps) {
		return true
	}
	if arow.Summary == nil || brow.Summary == nil {
		if arow.Summary != nil || brow.Summary != nil {
			return true
		}
	} else if arow.Summary.Text != brow.Summary.Text || arow.Summary.Branch != brow.Summary.Branch {
		return true
	}
	for i := range arow.Steps {
		p, q := arow.Steps[i], brow.Steps[i]
		if p.ID != q.ID || p.Text != q.Text || p.Thinking != q.Thinking || p.Streaming != q.Streaming || len(p.Tools) != len(q.Tools) {
			return true
		}
		for j := range p.Tools {
			if a.transcriptToolChanged(p.Tools[j], q.Tools[j]) {
				return true
			}
		}
	}
	return false
}
func (a *App) transcriptToolChanged(p, q transcriptui.ToolCall) bool {
	if p.ID != q.ID || p.Name != q.Name || p.Path != q.Path || p.Summary != q.Summary || p.Running != q.Running || p.Complete != q.Complete || p.Failed != q.Failed || p.Added != q.Added || p.Removed != q.Removed || !bytes.Equal(p.Arguments, q.Arguments) {
		return true
	}
	outputChanged := !bytes.Equal(p.Output, q.Output) || p.Facts.Truncated != q.Facts.Truncated || !activityCountEqual(p.Facts.OutputLines, q.Facts.OutputLines) || !activityCountEqual(p.Facts.TotalLines, q.Facts.TotalLines)
	if !outputChanged {
		return false
	}
	// Source tool_execution_update only invalidates an expanded detail. A final
	// result is visible activity regardless of expansion. Snapshot classification
	// cannot recreate intermediate RPC events that were never observed by the UI.
	key := "tool:" + q.ID
	if q.ID == "" {
		key += "unknown:" + q.Name
	}
	return q.Complete || !q.Running || a.transcriptUI.Open(key, false)
}
func activityCountEqual(a, b *uint64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
