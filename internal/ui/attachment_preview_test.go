package orbitui

import (
	"errors"
	"image"
	"io"
	"testing"

	"github.com/charlzyx/orbit/internal/attachments"
	"github.com/egoist/mygo/ui"
)

func TestAttachmentPreviewRejectsUnsafeConfigBeforeDecode(t *testing.T) {
	for _, tc := range []struct {
		name string
		w, h int
		err  error
	}{
		{"bad header", 1, 1, errors.New("invalid")}, {"zero", 0, 1, nil}, {"negative", -1, 1, nil},
		{"too many pixels", 16384, 16385, nil}, {"overflow", int(^uint(0) >> 1), int(^uint(0) >> 1), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			got := attachmentBitmapPreviewWith([]byte("fixture"), func(io.Reader) (image.Config, string, error) {
				return image.Config{Width: tc.w, Height: tc.h}, "fake", tc.err
			}, func([]byte) (ui.ImageSource, error) { calls++; return nil, nil })
			if calls != 0 || got != nil {
				t.Fatal("unsafe config reached pixel decoder", calls, got)
			}
		})
	}
}
func TestAttachmentPreviewExactThresholdUsesDecoderWithoutAllocatingPixels(t *testing.T) {
	calls := 0
	got := attachmentBitmapPreviewWith(nil, func(io.Reader) (image.Config, string, error) {
		return image.Config{Width: attachments.MaxDecodedPixels, Height: 1}, "fake", nil
	}, func([]byte) (ui.ImageSource, error) { calls++; return nil, nil })
	if calls != 1 || got != nil {
		t.Fatal("unchanged exact threshold was not admitted", calls)
	}
}
