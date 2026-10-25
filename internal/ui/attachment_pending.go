package orbitui

// One current-owner batch, not a global counter: completions from a discarded
// task can never unlock another draft's pending image work.
type attachmentImageBatch struct {
	owner     attachmentOwner
	count     int
	cancelled bool
}
type attachmentImageToken struct {
	batch *attachmentImageBatch
	done  bool
}

func (a *App) beginAttachmentImage(owner attachmentOwner) *attachmentImageToken {
	batch := a.pendingAttachmentImages
	if batch == nil || batch.owner != owner {
		a.cancelAttachmentImages()
		batch = &attachmentImageBatch{owner: owner}
		a.pendingAttachmentImages = batch
	}
	batch.count++
	return &attachmentImageToken{batch: batch}
}
func (a *App) finishAttachmentImage(token *attachmentImageToken) {
	if token.done {
		return
	}
	token.done = true
	batch := token.batch
	if !batch.cancelled {
		batch.count--
	}
	if a.pendingAttachmentImages == batch && batch.count == 0 {
		a.pendingAttachmentImages = nil
	}
}
func (a *App) cancelAttachmentImages() {
	if batch := a.pendingAttachmentImages; batch != nil {
		batch.cancelled = true
		batch.count = 0
		a.pendingAttachmentImages = nil
	}
}
func (a *App) pruneAttachmentImages() {
	if batch := a.pendingAttachmentImages; batch != nil && !a.ownsAttachment(batch.owner) {
		a.cancelAttachmentImages()
	}
}
func (a *App) hasPendingAttachmentImages() bool {
	a.pruneAttachmentImages()
	return a.pendingAttachmentImages != nil && a.pendingAttachmentImages.count > 0
}
