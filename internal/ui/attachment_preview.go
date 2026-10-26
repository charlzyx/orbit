package orbitui

import (
	"bytes"
	"image"
	"io"

	"github.com/charlzyx/orbit/internal/attachments"
	"github.com/egoist/mygo/ui"
)

// Preparation may deliberately preserve an oversized or undecodable raw payload.
// Preview independently admits dimensions before allocating decoded pixels. The
// existing threshold is not a total-memory budget; SVG keeps its separate path.
func attachmentBitmapPreview(data []byte) ui.ImageSource {
	return attachmentBitmapPreviewWith(data, image.DecodeConfig, func(raw []byte) (ui.ImageSource, error) {
		return ui.DecodeBitmap(raw)
	})
}

func attachmentBitmapPreviewWith(data []byte, config func(io.Reader) (image.Config, string, error), decode func([]byte) (ui.ImageSource, error)) ui.ImageSource {
	cfg, _, err := config(bytes.NewReader(data))
	if err != nil || !attachments.DecodeDimensionsAllowed(cfg.Width, cfg.Height) {
		return nil
	}
	preview, err := decode(data)
	if err != nil {
		return nil
	}
	return preview
}
