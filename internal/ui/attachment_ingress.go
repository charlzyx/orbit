package orbitui

import (
	"fmt"
	"path/filepath"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/attachments"
	"github.com/egoist/mygo/ui"
)

type attachmentOrigin uint8

const (
	attachmentDrop attachmentOrigin = iota
	attachmentFilePicker
	attachmentImagePicker
	attachmentClipboard
)
const (
	attachmentMaxPaths      = 256
	attachmentMaxPathBytes  = 256 << 10
	attachmentMaxIngress    = 8
	attachmentMaxIO         = 2
	attachmentMaxClipboard  = 64 << 20
	attachmentMaxRawPending = 128 << 20
	attachmentIngressBusy   = "Attachment input is busy or exceeds the input limit; try again after pending images finish"
)

// All accounting belongs to the UI owner. Retired queues retain their active
// slot and clipboard bytes until the actual worker result is consumed.
type attachmentIngressState struct {
	current                   *attachmentIngressQueue
	requests, slots, rawBytes int
}
type attachmentIngressQueue struct {
	owner           attachmentOwner
	requests        []*attachmentIngressRequest
	active          *attachmentIngressRequest
	cancelled, slot bool
}
type attachmentIngressRequest struct {
	origin     attachmentOrigin
	paths      []string
	index      int
	raw        []byte
	mime, name string
	read       func(string) ([]byte, string, error)
	token      *attachmentImageToken
	released   bool
}

func attachmentPathsFit(paths []string) bool {
	if len(paths) > attachmentMaxPaths {
		return false
	}
	remaining := attachmentMaxPathBytes
	for _, p := range paths {
		if len(p) > remaining {
			return false
		}
		remaining -= len(p)
	}
	return true
}
func attachmentRawFits(size, pending int) bool {
	return size >= 0 && size <= attachmentMaxClipboard && pending >= 0 && pending <= attachmentMaxRawPending && size <= attachmentMaxRawPending-pending
}
func (a *App) AddPaths(paths []string) { a.addPaths(paths, readAttachmentPath) }
func (a *App) addPaths(paths []string, read func(string) ([]byte, string, error)) {
	a.addPathsOrigin(paths, attachmentDrop, read)
}
func (a *App) addPathsOrigin(paths []string, origin attachmentOrigin, read func(string) ([]byte, string, error)) {
	if len(paths) == 0 {
		return
	}
	if !attachmentPathsFit(paths) {
		a.Notify(attachmentIngressBusy, true)
		return
	}
	needsIO := false
	for _, p := range paths {
		if imagePickerPathSupported(p) {
			needsIO = true
			break
		}
	}
	a.admitAttachmentIngress(a.attachmentOwner(), &attachmentIngressRequest{origin: origin, paths: paths, read: read}, needsIO)
}
func (a *App) AddImage(data []byte, mime, name string) {
	a.addImageFor(a.attachmentOwner(), data, mime, name)
}
func (a *App) addImageFor(owner attachmentOwner, data []byte, mime, name string) {
	a.admitAttachmentIngress(owner, &attachmentIngressRequest{origin: attachmentClipboard, raw: data, mime: mime, name: name}, true)
}
func (a *App) admitAttachmentIngress(owner attachmentOwner, r *attachmentIngressRequest, needsIO bool) {
	if !a.ownsAttachment(owner) {
		return
	}
	a.pruneAttachmentImages()
	s := &a.attachmentIngress
	q := s.current
	if q != nil && q.owner != owner {
		a.cancelAttachmentImages()
		q = nil
	}
	if s.requests >= attachmentMaxIngress || !attachmentRawFits(len(r.raw), s.rawBytes) || (needsIO && (q == nil || !q.slot) && s.slots >= attachmentMaxIO) {
		a.Notify(attachmentIngressBusy, true)
		return
	}
	if q == nil {
		q = &attachmentIngressQueue{owner: owner}
		s.current = q
	}
	if needsIO && !q.slot {
		q.slot = true
		s.slots++
	}
	// Nothing is copied, reserved or inserted before whole-request admission.
	r.paths = append([]string(nil), r.paths...)
	if r.raw != nil {
		r.raw = append([]byte(nil), r.raw...)
	}
	s.rawBytes += len(r.raw)
	s.requests++
	r.token = a.beginAttachmentImage(owner)
	q.requests = append(q.requests, r)
	a.pumpAttachmentIngress(q)
}
func (a *App) releaseAttachmentRequest(r *attachmentIngressRequest) {
	if r.released {
		return
	}
	r.released = true
	a.attachmentIngress.requests--
	a.attachmentIngress.rawBytes -= len(r.raw)
	r.raw = nil
	r.paths = nil
	a.finishAttachmentImage(r.token)
}
func (a *App) releaseAttachmentSlot(q *attachmentIngressQueue) {
	if q.slot {
		q.slot = false
		a.attachmentIngress.slots--
	}
}
func (a *App) cancelAttachmentIngress() {
	q := a.attachmentIngress.current
	if q == nil {
		return
	}
	a.attachmentIngress.current = nil
	q.cancelled = true
	for _, r := range q.requests {
		if r != q.active {
			a.releaseAttachmentRequest(r)
		}
	}
	q.requests = nil
	if q.active == nil {
		a.releaseAttachmentSlot(q)
	}
}
func (a *App) attachmentIngressAlive(q *attachmentIngressQueue, r *attachmentIngressRequest) bool {
	if !q.cancelled && a.attachmentIngress.current == q && a.ownsAttachment(q.owner) {
		return true
	}
	if a.attachmentIngress.current == q {
		a.cancelAttachmentImages()
	}
	q.active = nil
	a.releaseAttachmentRequest(r)
	a.releaseAttachmentSlot(q)
	return false
}
func (a *App) finishAttachmentIngress(q *attachmentIngressQueue, r *attachmentIngressRequest) {
	a.releaseAttachmentRequest(r)
	q.requests = q.requests[1:]
}
func (a *App) pumpAttachmentIngress(q *attachmentIngressQueue) {
	if q.active != nil || q.cancelled {
		return
	}
	for len(q.requests) > 0 {
		r := q.requests[0]
		if r.origin == attachmentClipboard {
			if len(a.attachments) >= 8 {
				a.Notify("Maximum 8 attachments per message", true)
				a.finishAttachmentIngress(q, r)
				continue
			}
			a.decodeAttachmentIngress(q, r, r.raw, r.mime, r.name)
			return
		}
		for r.index < len(r.paths) {
			if r.origin == attachmentDrop && len(a.attachments) >= 8 {
				a.Notify("Maximum 8 attachments per message", true)
				r.index = len(r.paths)
				break
			}
			p := r.paths[r.index]
			r.index++
			if !imagePickerPathSupported(p) {
				if r.origin != attachmentImagePicker {
					a.queueAttachmentPath(q.owner, p+" ")
				}
				continue
			}
			q.active = r
			read := r.read
			consumed := false
			a.Async(func() func() {
				raw, mime, err := read(p)
				return func() {
					if consumed {
						return
					}
					consumed = true
					if !a.attachmentIngressAlive(q, r) {
						return
					}
					q.active = nil
					if err != nil {
						if r.origin == attachmentImagePicker {
							a.save(err)
						} else {
							a.queueAttachmentPath(q.owner, p+" ")
						}
						a.pumpAttachmentIngress(q)
						return
					}
					a.decodeAttachmentIngress(q, r, raw, mime, filepath.Base(p))
				}
			})
			return
		}
		a.finishAttachmentIngress(q, r)
	}
	a.releaseAttachmentSlot(q)
	if a.attachmentIngress.current == q {
		a.attachmentIngress.current = nil
	}
}
func (a *App) decodeAttachmentIngress(q *attachmentIngressQueue, r *attachmentIngressRequest, data []byte, mime, name string) {
	q.active = r
	consumed := false
	a.Async(func() func() {
		im := attachments.Prepare(data, mime)
		var preview ui.ImageSource
		if im.MIME == "image/svg+xml" {
			preview, _ = ui.ParseSVG(im.Bytes)
		} else {
			preview, _ = ui.DecodeBitmap(im.Bytes)
		}
		result := attachment{Name: name, Image: agent.Image{Type: "image", Data: im.Base64(), MIMEType: im.MIME}, Preview: preview}
		return func() {
			if consumed {
				return
			}
			consumed = true
			if !a.attachmentIngressAlive(q, r) {
				return
			}
			q.active = nil
			if len(a.attachments) >= 8 {
				a.Notify("Maximum 8 attachments per message", true)
			} else {
				a.nextAttachment++
				result.ID = fmt.Sprint(a.nextAttachment)
				a.attachments = append(a.attachments, result)
			}
			if r.origin == attachmentClipboard {
				a.finishAttachmentIngress(q, r)
			}
			a.pumpAttachmentIngress(q)
		}
	})
}
