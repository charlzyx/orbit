package orbitui

import (
	"encoding/json"
	"fmt"
	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/transcriptui"
	"github.com/egoist/mygo/ui"
	"github.com/go-text/typesetting/segmenter"
	"strings"
	"time"
)

// Source: transcript_view.rs 80–95, 1396–1734. Keep this outside the
// retained transcript selection list: navigation never replaces its rows.
// Only this explicitly designated navigation shell preserves the selection.
type navigationState struct {
	turns   []int
	count   int
	hovered int // user row + 1, zero means no hover
	active  int // user row + 1; detects active changes for rail autoscroll
	scroll  ui.ScrollState
}

func (a *App) userTurns() []int {
	if a.navigation.turns == nil || a.navigation.count != len(a.Model.Messages) {
		a.navigation.turns = make([]int, 0)
		a.navigation.count = len(a.Model.Messages)
		for i, m := range a.Model.Messages {
			if m.Role == "user" {
				a.navigation.turns = append(a.navigation.turns, i)
			}
		}
	}
	return a.navigation.turns
}
func (a *App) activeUserTurn() int {
	end := len(a.Model.Messages) - 1
	if !a.transcript.AtEnd() {
		end, _ = a.transcript.Visible()
	} else {
		for i := len(a.Model.Messages) - 1; i >= 0; i-- {
			if a.Model.Messages[i].Streaming {
				end = i
				break
			}
		}
	}
	turns := a.userTurns()
	for i := len(turns) - 1; i >= 0; i-- {
		if turns[i] <= end {
			return turns[i]
		}
	}
	return -1
}
func (a *App) jumpTurn(delta int) {
	turns := a.userTurns()
	if len(turns) == 0 {
		return
	}
	current, _ := a.transcript.Visible()
	if a.transcript.AtEnd() {
		current = len(a.Model.Messages)
	}
	pos := 0
	for i, row := range turns {
		if row <= current {
			pos = i
		}
	}
	next := max(0, min(len(turns)-1, pos+delta))
	if next != pos {
		a.transcript.ScrollTo(turns[next], ui.Start)
	}
}
func (a *App) dismissRailHint() {
	if a.railHintDismissed {
		return
	}
	a.railHintDismissed = true
	a.railHintAt = time.Time{}
	if store := a.Options.Store; store != nil {
		a.railHintSaving = true
		a.Async(func() func() {
			err := store.Update("hints.json", func(m map[string]json.RawMessage) error { m["rail_hint_seen"] = json.RawMessage("true"); return nil })
			return func() { a.railHintSaving = false; a.save(err) }
		})
	}
}
func railSnippet(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	var seg segmenter.Segmenter
	if err := seg.InitWithString(s); err != nil {
		r := []rune(s)
		if len(r) > n {
			return string(r[:n]) + "…"
		}
		return s
	}
	it := seg.GraphemeIterator()
	end, count := 0, 0
	for it.Next() {
		g := it.Grapheme()
		continuation := end > 0 && strings.HasSuffix(s[:end], "\u200d")
		if count == n && !continuation {
			return s[:end] + "…"
		}
		end = g.OffsetInBytes + g.LengthInBytes
		if !continuation {
			count++
		}
	}
	return s
}
func (a *App) railPreview(row int) (string, string) {
	if row < 0 || row >= len(a.Model.Messages) || a.Model.Messages[row].Role != "user" {
		return "", ""
	}
	prompt := railSnippet(a.Model.Messages[row].Text, 100)
	for i := row + 1; i < len(a.Model.Messages) && a.Model.Messages[i].Role != "user"; i++ {
		if s := strings.TrimSpace(a.Model.Messages[i].Text); s != "" {
			return prompt, railSnippet(s, 240)
		}
	}
	return prompt, ""
}
func (a *App) navigationRail(c *ui.Context, bounds ui.Rect, scope ui.Element) {
	turns := a.userTurns()
	if len(turns) < 2 || bounds.W < 1040 || a.transcriptScroll.MaxY <= 0 {
		a.navigation.hovered = 0
		return
	}
	t := a.Theme
	contactAlpha := float32(.06)
	if !t.Dark {
		contactAlpha = .03
	}
	height := min(float32(len(turns))*12, bounds.H*.8)
	active := a.activeUserTurn()
	pos := -1
	for i, row := range turns {
		if row == active {
			pos = i
			break
		}
	}
	if a.navigation.active != active+1 && pos >= 0 {
		top := float32(pos) * 12
		if top < a.navigation.scroll.Y || top+12 > a.navigation.scroll.Y+height {
			a.navigation.scroll.Y = max(0, min(float32(len(turns))*12-height, top+6-height/2))
		}
		a.navigation.active = active + 1
	}
	if !a.railHintDismissed {
		if a.railHintAt.IsZero() {
			a.railHintAt = c.Now()
		}
		remaining := 10*time.Second - c.Now().Sub(a.railHintAt)
		if remaining <= 0 {
			a.dismissRailHint()
		} else {
			c.After(remaining)
		}
	}
	shell := ui.Column(c.Key("navigation-rail")).Label("Conversation navigation rail").Absolute().Left(16).Top((bounds.H - height) / 2).Width(44).Height(height)
	transcriptui.RetainSelectionInteraction(shell, scope)
	oldHover := a.navigation.hovered
	emphasized := -1
	for i, row := range turns {
		if row+1 == oldHover {
			emphasized = i
			break
		}
	}
	shell.Children(func() {
		ui.Scroll(c.Key("navigation-ticks")).Width(44).Height(height).TrackScroll(&a.navigation.scroll).Children(func() {
			for i, row := range turns {
				tick := ui.Column(c.Key(fmt.Sprintf("navigation-tick-%d", row))).Label(fmt.Sprintf("Go to turn %d", i+1)).Cursor(ui.CursorPointer).Width(44).Height(12).Shrink(0).AlignItems(ui.Start).Justify(ui.Center)
				if tick.Hovered() {
					a.navigation.hovered = row + 1
				} else if a.navigation.hovered == row+1 {
					a.navigation.hovered = 0
				}
				if tick.Clicked() {
					a.dismissRailHint()
					a.transcript.ScrollTo(row, ui.Start)
				}
				scale := float32(.25)
				if emphasized >= 0 {
					d := i - emphasized
					if d < 0 {
						d = -d
					}
					if d < 4 {
						scale = []float32{1, .68, .44, .25}[d]
					}
				}
				color := t.TextMuted.Alpha(.45)
				if row == active || row+1 == a.navigation.hovered {
					color = t.Text
				}
				tick.Children(func() {
					ui.Box(c.Key(fmt.Sprintf("tick-bar-%d", row))).Label(fmt.Sprintf("Turn %d tick", i+1)).Width(32 * scale).Height(2).Radius(1).Background(color)
				})
			}
		})
		if a.navigation.scroll.Y > .5 {
			ui.Box(c).Label("Rail top fade").Absolute().Left(0).Top(0).Size(44, 20).Gradient(t.Background, t.Background.Alpha(0), 90).PassThrough()
		}
		if a.navigation.scroll.Y < float32(len(turns))*12-height-.5 {
			ui.Box(c).Label("Rail bottom fade").Absolute().Left(0).Bottom(0).Size(44, 20).Gradient(t.Background.Alpha(0), t.Background, 90).PassThrough()
		}
		hp := -1
		for i, row := range turns {
			if row+1 == a.navigation.hovered {
				hp = i
				break
			}
		}
		if hp >= 0 {
			prompt, response := a.railPreview(turns[hp])
			top := max(float32(0), min(height-126, float32(hp)*12+6-a.navigation.scroll.Y-63))
			ui.Column(c.Key("turn-preview")).Label("Turn preview").Absolute().Left(56).Top(top).Width(320).MaxHeight(126).Clip().Padding(12, 15).Gap(7).Background(t.Elevated).Border(1, t.BorderStrong()).Radius(12).Shadow(0, 2, 3, 0, ui.RGBA(0, 0, 0, .12)).Shadow(0, 1, 0, 0, ui.RGBA(0, 0, 0, contactAlpha)).Children(func() {
				ui.Text(c, orbitassets.Translate(a.catalog, a.english, "transcript_view.turn_of_total_hint", map[string]string{"current": fmt.Sprint(hp + 1), "total": fmt.Sprint(len(turns))})).FontSize(t.Px(10)).FixedLineHeight(t.Px(14)).TextColor(t.TextMuted)
				ui.Text(c, prompt).SingleLine().FontSize(t.Px(14)).FixedLineHeight(t.Px(20)).FontWeight(600).TextColor(t.Text)
				if response != "" {
					ui.Text(c, response).MaxHeight(60).Clip().FontSize(t.Px(14)).FixedLineHeight(t.Px(20)).TextColor(t.TextMuted)
				}
			})
		} else if !a.railHintDismissed {
			hint := ui.Column(c.Key("rail-hint")).Label("Navigation rail hint").Cursor(ui.CursorPointer).Absolute().Left(56).Top(0).Width(216).Padding(10, 13).Gap(3).Background(t.Elevated).Border(1, t.BorderStrong()).Radius(12).Shadow(0, 2, 3, 0, ui.RGBA(0, 0, 0, .12)).Shadow(0, 1, 0, 0, ui.RGBA(0, 0, 0, contactAlpha))
			hint.Children(func() {
				ui.Text(c, a.Tr("transcript_view.jump_between_turns")).FontSize(t.Px(12)).FixedLineHeight(t.Px(16))
				ui.Text(c, a.Tr("transcript_view.click_a_line_or_press_to_revisit_any_prompt")).FontSize(t.Px(12)).FixedLineHeight(t.Px(16)).TextColor(t.TextMuted)
			})
			if hint.Clicked() {
				a.dismissRailHint()
			}
		}
	})
	if !shell.Hovered() {
		a.navigation.hovered = 0
	}
	if oldHover != a.navigation.hovered {
		c.Invalidate()
	}
}
