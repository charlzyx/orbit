package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
)

type inlineImageContent struct {
	data, mime string // Immutable references; never logged or copied into keys.
	revision   uint64
}
type inlineImageContentToken struct {
	key      string
	revision uint64
}

// Only first request adopts content on the render path. Existing slots are
// compared at snapshot adoption, not hashed/decoded on every frame.
func (a *App) inlineImagePayload(key string, image agent.Image) inlineImageContentToken {
	if content, ok := a.inlineImages.contents[key]; ok {
		return inlineImageContentToken{key, content.revision}
	}
	return a.setInlineImagePayload(key, image)
}
func (a *App) setInlineImagePayload(key string, image agent.Image) inlineImageContentToken {
	if old, ok := a.inlineImages.contents[key]; ok {
		if old.mime == image.MIMEType && old.data == image.Data {
			return inlineImageContentToken{key, old.revision}
		}
		a.invalidateInlineImagePayload(inlineImageContentToken{key, old.revision})
	}
	if a.inlineImages.contents == nil {
		a.inlineImages.contents = make(map[string]inlineImageContent)
	}
	a.inlineImages.nextContentRevision++
	content := inlineImageContent{data: image.Data, mime: image.MIMEType, revision: a.inlineImages.nextContentRevision}
	a.inlineImages.contents[key] = content
	return inlineImageContentToken{key, content.revision}
}
func (a *App) invalidateInlineImagePayload(token inlineImageContentToken) {
	delete(a.messageImages, token.key)
	delete(a.imageLoading, token.key)
	delete(a.imageErrors, token.key)
	if a.inlineImages.preview == token {
		a.imageOpen = false
		a.imagePreview = nil
		a.inlineImages.preview = inlineImageContentToken{}
	}
}
func (a *App) ownsInlineImagePayload(token inlineImageContentToken) bool {
	content, ok := a.inlineImages.contents[token.key]
	return ok && content.revision == token.revision
}

// Projected full snapshots are the content-change boundary. Invalidate even
// while the row/window is hidden; A→B→A and removal/re-add cannot revive an old
// completion. We only track slots already requested for rendering.
func (a *App) syncInlineImageContents(messages []Message) {
	if len(a.inlineImages.contents) == 0 {
		return
	}
	seen := make(map[string]bool, len(a.inlineImages.contents))
	for _, message := range messages {
		for index, image := range message.Images {
			key := fmt.Sprintf("%s:%d", message.ID, index)
			if _, ok := a.inlineImages.contents[key]; ok {
				a.setInlineImagePayload(key, image)
				seen[key] = true
			}
		}
	}
	for key, content := range a.inlineImages.contents {
		if !seen[key] {
			a.invalidateInlineImagePayload(inlineImageContentToken{key, content.revision})
			delete(a.inlineImages.contents, key)
		}
	}
}
