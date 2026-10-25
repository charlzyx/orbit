package orbitui

import (
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

// queueContents retains delivery kinds. Queued is a legacy fixture input only;
// production snapshots project the typed queue, including its empty state.
func (a *App) queueContents() agent.Queue {
	q := a.Model.Queue
	if q.Steering == nil && q.FollowUp == nil && len(a.Model.Queued) > 0 {
		q.FollowUp = a.Model.Queued
	}
	return q
}

func (a *App) queueHeading(q agent.Queue) (string, string) {
	count := len(q.Steering) + len(q.FollowUp)
	label, detail := a.text("view.queue_queued", "Queued"), a.text("view.steering_plus_followups", "Steering now; follow-ups send after the task finishes (%{count})")
	if len(q.FollowUp) == 0 {
		label, detail = a.text("view.queue_steering", "Steering"), a.text("view.steering_turn", "Steering the running turn (%{count})")
	} else if len(q.Steering) == 0 {
		detail = a.text("view.queued_after_task", "Queued — sends after the current task finishes (%{count})")
	}
	return strings.ToUpper(fmt.Sprintf("%s · %d", label, count)), strings.ReplaceAll(detail, "%{count}", fmt.Sprint(count))
}

// queueBar mirrors pinned app/view.rs::queue_bar and helpers.rs::queue_row.
// The header owns the only action. Rows are read-only, ordered steering first.
func (a *App) queueBar(c *ui.Context) {
	q := a.queueContents()
	if len(q.Steering)+len(q.FollowUp) == 0 {
		return
	}
	t := a.Theme
	label, detail := a.queueHeading(q)
	ui.Column(c.Key("queue-bar")).Label("Queue bar").AlignSelf(ui.Stretch).Margin(0, 0, t.Space(8)).Radius(t.Px(12)).Border(1, t.Border).Background(t.Elevated).Clip().Gap(0).Children(func() {
		ui.Row(c.Key("queue-header")).Label("Queue header").AlignSelf(ui.Stretch).Height(30).Shrink(0).PaddingX(t.Space(12)).Gap(t.Space(6)).AlignItems(ui.Center).Children(func() {
			a.icon(c, "icons/clock.svg", t.Px(12), t.TextMuted)
			ui.Text(c.Key("queue-label"), label).Grow(1).MinWidth(0).SingleLine().Tooltip(detail).FontSize(t.Px(10)).FontWeight(500).TextColor(t.TextMuted)
			clear := a.text("view.clear_queue", "Clear queue")
			b := ui.ButtonBase(c.Key("clear-queue")).Label(clear).Tooltip(clear).Size(t.Px(18), t.Px(18)).Shrink(0).Radius(t.Px(6)).Center()
			color := t.TextMuted
			if b.Hovered() {
				b.Background(t.Hover)
				color = t.Text
			}
			b.Children(func() { a.icon(c, "icons/x.svg", t.Px(12), color) })
			if b.Clicked() {
				a.Dispatch(Action{Kind: "ClearQueue"})
			}
		})
		for i, text := range q.Steering {
			a.queueRow(c.Key(fmt.Sprintf("queue-steering-%d", i)), a.text("view.queue_steer", "Steer"), text, false)
		}
		for i, text := range q.FollowUp {
			a.queueRow(c.Key(fmt.Sprintf("queue-follow-up-%d", i)), a.text("view.queue_follow_up", "Follow-up"), text, true)
		}
	})
}

func (a *App) queueRow(c *ui.Context, kind, text string, follow bool) {
	t := a.Theme
	color := t.AccentText
	if follow {
		color = t.TextMuted
	}
	ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Start).Gap(10).Padding(7, t.Space(12)).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Children(func() {
		ui.Text(c, strings.ToUpper(kind)).Shrink(0).SingleLine().Padding(1, 0, 0, 0).FontSize(t.Px(10)).FontWeight(500).TextColor(color)
		ui.Text(c, text).Grow(1).MinWidth(0).MaxLines(2).FontSize(t.Px(12)).FixedLineHeight(t.Px(16)).TextColor(t.Text)
	})
}
