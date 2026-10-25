package gitworkbench

import (
	"context"
	"fmt"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func (s *State) stashes(c *ui.Context) {
	ui.Row(c).PaddingX(14).Gap(8).Children(func() {
		ui.Text(c, fmt.Sprintf("Stashes (%d)", len(s.Data.Stashes))).Bold()
		ui.Spacer(c)
		if button(c, "Stash changes…", s.Busy).Clicked() {
			s.openDialog("stash")
		}
	})
	if len(s.Data.Stashes) == 0 && !s.Busy {
		ui.Column(c).Padding(30).Gap(8).Children(func() {
			ui.Text(c, "No stashes").FontSize(18).Bold()
			muted(c, "Save work in progress to return to a clean workspace, then apply or pop it later.")
		})
	}
	s.list.Key = func(i int) any { return s.Data.Stashes[i].Index }
	ui.List(c.Key("stashes-list"), &s.list, len(s.Data.Stashes), func(i int) {
		stash := s.Data.Stashes[i]
		ui.Row(c.Key(stash.Index)).Padding(10, 14).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Grow(1).Gap(4).Children(func() {
				ui.Text(c, stash.Message).MaxLines(2)
				muted(c, fmt.Sprintf("stash@{%d}", stash.Index)).Font("monospace")
			})
			if button(c, "Apply", s.Busy).Clicked() {
				s.RunGit("Apply stash", func(ctx context.Context, g workspace.Git) (string, error) {
					return g.ApplyStash(ctx, stash.Index, false)
				})
			}
			if button(c, "Pop", s.Busy).Clicked() {
				s.RunGit("Pop stash", func(ctx context.Context, g workspace.Git) (string, error) {
					return g.ApplyStash(ctx, stash.Index, true)
				})
			}
			if button(c, "Drop…", s.Busy).Clicked() {
				s.Confirm("Drop stash?", stash.Message+"\nThis removes the saved stash from the stack.", func() {
					s.RunGit("Drop stash", func(ctx context.Context, g workspace.Git) (string, error) { return g.DropStash(ctx, stash.Index) })
				})
			}
		})
	}).Grow(1).MinHeight(0)
}
