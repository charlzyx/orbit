package orbitui

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/attachments"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type attachment struct {
	ID, Name string
	Image    agent.Image
	Preview  ui.ImageSource
}

type attachmentOwner struct {
	runtime, session, path, workspace string
	generation                        uint64
	window                            *mygo.Window
	target                            agent.WorkflowTarget
	managed, live                     bool
}

func (a *App) attachmentOwner() attachmentOwner {
	owner := attachmentOwner{runtime: a.Model.RuntimeID, session: a.Model.SessionID, path: a.Model.SessionPath, workspace: a.Model.Workspace, window: a.Window, live: true}
	if a.Options.AttachmentContext != nil {
		owner.target, owner.generation, owner.live = a.Options.AttachmentContext()
		owner.managed = true
		// The opaque runtime key/task generation survives normal session ID/path
		// hydration but changes for another draft, even while its ID is empty.
		owner.runtime, owner.session, owner.path = "", "", ""
	}
	return owner
}
func (a *App) ownsAttachment(owner attachmentOwner) bool {
	return owner.live && owner == a.attachmentOwner() && (owner.window == nil || !owner.window.IsDestroyed())
}
func (a *App) AddImage(data []byte, mime, name string) {
	a.addImageFor(a.attachmentOwner(), data, mime, name)
}
func (a *App) addImageFor(owner attachmentOwner, data []byte, mime, name string) {
	if !a.ownsAttachment(owner) {
		return
	}
	if len(a.attachments) >= 8 {
		a.Notify("Maximum 8 attachments per message", true)
		return
	}
	a.Async(func() func() {
		im := attachments.Prepare(data, mime)
		var preview ui.ImageSource
		if im.MIME == "image/svg+xml" {
			preview, _ = ui.ParseSVG(im.Bytes)
		} else {
			preview, _ = ui.DecodeBitmap(im.Bytes)
		}
		return func() {
			if !a.ownsAttachment(owner) {
				return
			}
			if len(a.attachments) >= 8 {
				a.Notify("Maximum 8 attachments per message", true)
				return
			}
			a.nextAttachment++
			a.attachments = append(a.attachments, attachment{ID: fmt.Sprint(a.nextAttachment), Name: name, Image: agent.Image{Type: "image", Data: im.Base64(), MIMEType: im.MIME}, Preview: preview})
		}
	})
}
func (a *App) AddPaths(paths []string) { a.addPaths(paths, readAttachmentPath) }
func (a *App) addPaths(paths []string, read func(string) ([]byte, string, error)) {
	owner := a.attachmentOwner()
	if !a.ownsAttachment(owner) {
		return
	}
	for _, p := range paths {
		if !imagePickerPathSupported(p) {
			a.queueAttachmentPath(owner, p+" ")
			continue
		}
		a.Async(func() func() {
			raw, mime, err := read(p)
			return func() {
				// Check the owner captured BEFORE reading, not the task active when the
				// read finishes. Otherwise AddImage would silently adopt a newer task.
				if !a.ownsAttachment(owner) {
					return
				}
				if err != nil {
					a.save(err)
					return
				}
				a.addImageFor(owner, raw, mime, filepath.Base(p))
			}
		})
	}
}
func readAttachmentPath(p string) ([]byte, string, error) {
	st, err := os.Stat(p)
	if err != nil {
		return nil, "", err
	}
	if st.Size() > 64<<20 {
		return nil, "", fmt.Errorf("Image exceeds 64 MiB: %s", filepath.Base(p))
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return nil, "", err
	}
	mime := http.DetectContentType(raw)
	switch strings.ToLower(filepath.Ext(p)) {
	case ".svg":
		mime = "image/svg+xml"
	case ".webp":
		mime = "image/webp"
	case ".tiff", ".tif":
		mime = "image/tiff"
	}
	return raw, mime, nil
}
func (a *App) attachmentRow(c *ui.Context) {
	if len(a.attachments) == 0 {
		return
	}
	ui.Row(c).AlignSelf(ui.Stretch).Gap(8).Children(func() {
		for i, im := range a.attachments {
			i, im := i, im
			ui.Row(c.Key("attachment:"+im.ID)).Padding(6).Gap(6).Background(a.Theme.Elevated).Border(1, a.Theme.Border).Radius(8).AlignItems(ui.Center).Children(func() {
				if im.Preview != nil {
					ui.Image(c, im.Preview).Size(44, 44)
				} else {
					a.icon(c, "icons/image.svg", 24, a.Theme.TextMuted)
				}
				ui.Text(c, im.Name).MaxWidth(120).MaxLines(1).FontSize(12)
				if a.iconButton(c, "Remove "+im.Name, "icons/x.svg").Clicked() {
					a.attachments = append(a.attachments[:i], a.attachments[i+1:]...)
				}
			})
		}
	})
}

func (a *App) messageImageRow(c *ui.Context, m Message) {
	owner := a.inlineImageToken()
	if !a.ownsInlineImage(owner) {
		return
	}
	ui.Row(c).Gap(8).Children(func() {
		for i, im := range m.Images {
			key := fmt.Sprintf("%s:%d", m.ID, i)
			content := a.inlineImagePayload(key, im)
			source := a.messageImages[key]
			if source == nil && !a.imageLoading[key] && a.imageErrors[key] == "" {
				a.imageLoading[key] = true
				a.Async(func() func() {
					var source ui.ImageSource
					var e error
					if len(im.Data) > 96<<20 {
						e = errors.New("Image is too large to preview")
					} else {
						var b []byte
						b, e = base64.StdEncoding.DecodeString(im.Data)
						if e == nil {
							if im.MIMEType == "image/svg+xml" {
								source, e = ui.ParseSVG(b)
							} else {
								source, e = ui.DecodeBitmap(b)
							}
						}
					}
					return func() {
						if !a.ownsInlineImage(owner) || !a.ownsInlineImagePayload(content) {
							return
						}
						a.imageLoading[key] = false
						if e != nil {
							a.imageErrors[key] = e.Error()
						} else {
							a.messageImages[key] = source
						}
					}
				})
			}
			if source != nil {
				b := ui.ButtonBase(c.Key(fmt.Sprintf("inline-image:%d:%d:%s", owner.generation, content.revision, key))).Label("Open image attachment").Padding(4).Radius(8).Border(1, a.Theme.Border).Children(func() { ui.Image(c, source).MaxWidth(240).MaxHeight(180) })
				if b.Clicked() && a.ownsInlineImage(owner) && a.ownsInlineImagePayload(content) {
					a.inlineImages.preview = content
					a.imagePreview = source
					a.imageOpen = true
				}
			} else if err := a.imageErrors[key]; err != "" {
				ui.Text(c, err).TextColor(a.Theme.Error)
			} else {
				ui.Spinner(c).Size(20, 20)
			}
		}
	})
}
func (a *App) imageDialog(c *ui.Context) {
	if !a.imageOpen || a.imagePreview == nil {
		return
	}
	a.modal(c.Key("image-preview"), &a.imageOpen, 1000, func() {
		ui.Row(c).AlignSelf(ui.Stretch).Children(func() {
			ui.Text(c, "Image attachment").Grow(1)
			if a.iconButton(c, "Close image", "icons/x.svg").Clicked() {
				a.imageOpen = false
				a.inlineImages.preview = inlineImageContentToken{}
			}
		})
		width, height := c.Size()
		ui.Scroll(c).MaxHeight(min(600, max(100, height-140))).Children(func() { ui.Image(c, a.imagePreview).MaxWidth(min(960, max(100, width-80))) })
	})
}
