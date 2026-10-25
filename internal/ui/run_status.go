package orbitui

import (
	"fmt"
	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
)

// runStatusStrip follows pinned app/view.rs presentation. Cancel hides the
// current retry immediately, as settings.rs does, without changing the real
// runtime's busy flags; a subsequent retry event owns a new revision.
func (a *App) runStatusStrip(c *ui.Context) {
	retry := a.Model.RetryDetail
	retryActive := retry.Active && !a.retryPresentationCancelled()
	if !retryActive && !a.Model.Compacting {
		return
	}
	t := a.Theme
	text, color := a.Tr("view.preparing_conversation_context"), t.Accent
	if retryActive {
		key := "runtime.attempt"
		if retry.HasMax {
			key = "runtime.attempt_of"
		}
		attempt := orbitassets.Translate(a.catalog, a.english, key, map[string]string{"attempt": fmt.Sprint(retry.Attempt), "max": fmt.Sprint(retry.Max)})
		message := retry.Error
		if !retry.ErrorProvided {
			message = a.Tr("events.transient_error")
		}
		text = orbitassets.Translate(a.catalog, a.english, "runtime.retrying", map[string]string{"attempt": attempt, "error": message})
		color = t.Warning
	}
	ui.Row(c.Key("run-status")).Label("Run status").AlignSelf(ui.Stretch).Margin(0, 0, t.Space(8)).Padding(7, 10).Radius(t.Px(12)).Border(1, t.Border).Background(t.Elevated).AlignItems(ui.Center).Gap(t.Space(8)).Children(func() {
		ui.Box(c).Size(6, 6).Shrink(0).Radius(3).Background(color)
		label := ui.Text(c.Key("run-status-text"), text).Grow(1).MinWidth(0).FontSize(t.Px(12)).TextColor(t.TextSecondary)
		if retryActive {
			label.SingleLine()
			b := ui.ButtonBase(c.Key("cancel-retry")).Label("Cancel retry").Height(t.Px(18)).PaddingX(t.Space(4)).Shrink(0).Radius(t.Px(6)).Disabled(!a.Model.Connected)
			fg := t.TextSecondary
			if b.Hovered() {
				b.Background(t.Hover)
				fg = t.Text
			}
			b.Children(func() { ui.Text(c, a.Tr("view.cancel")).TextColor(fg).FontSize(t.Px(12)) })
			if b.Clicked() {
				a.Dispatch(Action{Kind: "AbortRetry"})
			}
		}
	})
}
