package orbitui

import "github.com/egoist/mygo/ui"

// Inline transcript images have their own lifetime. Markdown's bounded cache
// does not own these maps. The attachment context carries the opaque live task
// generation even when a draft has no ID or a runtime reuses an ID.
type inlineImageOwner struct {
	runtime, session, path, workspace string
	task                              attachmentOwner
}
type inlineImageState struct {
	contents            map[string]inlineImageContent
	nextContentRevision uint64
	preview             inlineImageContentToken
	owner               inlineImageOwner
	generation          uint64
	known, closed       bool
}
type inlineImageToken struct {
	owner      inlineImageOwner
	generation uint64
}

func (a *App) currentInlineImageOwner() inlineImageOwner {
	return inlineImageOwner{a.Model.RuntimeID, a.Model.SessionID, a.Model.SessionPath, a.Model.Workspace, a.attachmentOwner()}
}
func (a *App) resetInlineImages() {
	a.inlineImages.generation++
	a.inlineImages.contents = nil
	a.inlineImages.preview = inlineImageContentToken{}
	a.messageImages = map[string]ui.ImageSource{}
	a.imageLoading = map[string]bool{}
	a.imageErrors = map[string]string{}
	a.imageOpen = false
	a.imagePreview = nil
}
func (a *App) syncInlineImageOwner() {
	owner := a.currentInlineImageOwner()
	if a.inlineImages.known && owner != a.inlineImages.owner {
		a.resetInlineImages()
	}
	a.inlineImages.owner = owner
	a.inlineImages.known = true
}
func (a *App) inlineImageToken() inlineImageToken {
	a.syncInlineImageOwner()
	return inlineImageToken{a.inlineImages.owner, a.inlineImages.generation}
}
func (a *App) ownsInlineImage(token inlineImageToken) bool {
	return !a.inlineImages.closed && token.generation == a.inlineImages.generation && token.owner == a.currentInlineImageOwner() && a.ownsAttachment(token.owner.task)
}
func (a *App) closeInlineImages() {
	a.inlineImages.closed = true
	a.resetInlineImages()
}
