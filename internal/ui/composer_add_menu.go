package orbitui

import "github.com/egoist/mygo/ui"

// The native menu resolves a choice against a later build. Key its button by
// the opaque task owner so an old menu cannot insert into a replacement draft.
func (a *App) composerAddMenu(c *ui.Context) {
	owner := a.attachmentOwner()
	a.composerChip(c.Key(struct {
		Kind      string
		Owner     attachmentOwner
		ModeEpoch uint64
	}{"composer-add", owner, a.composerModeEpoch}), "Add attachment", "", "icons/plus.svg", false, 0).Menu(func(m *ui.Menu) {
		// Native menus own OS keyboard dispatch while open. Set the return target
		// when building: MyGo has no public cancellation callback, so this also
		// restores the composer after Escape/outside-click, not only selection.
		a.closeComposerModeMenu(false)
		a.composer.Focus()
		if m.Item(a.text("composer.attach_image", "Attach image…")).Chosen() {
			a.Dispatch(Action{Kind: "BrowseImage"})
		}
		if m.Item("Attach files…").Chosen() {
			a.Dispatch(Action{Kind: "BrowseAttachments"})
		}
		if m.Item(a.text("composer.mention_file", "Mention file")).Chosen() {
			a.queueAttachmentPath(owner, "@")
		}
		if m.Item("Paste image").Chosen() {
			a.Dispatch(Action{Kind: "PasteImage"})
		}
	})
}
