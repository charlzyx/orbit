package gitworkbench

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

type avatarResult struct {
	url  string
	data []byte
}

func LoadAvatar(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	request, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return nil, e
	}
	response, e := http.DefaultClient.Do(request)
	if e != nil {
		return nil, e
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("avatar HTTP %d", response.StatusCode)
	}
	raw, e := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if e != nil {
		return nil, e
	}
	if len(raw) > 2*1024*1024 {
		return nil, fmt.Errorf("avatar exceeds image limit")
	}
	cfg, _, e := image.DecodeConfig(bytes.NewReader(raw))
	if e != nil || cfg.Width > 1024 || cfg.Height > 1024 {
		return nil, fmt.Errorf("avatar dimensions unavailable or too large")
	}
	return raw, nil
}
func (s *State) avatar(c *ui.Context, name, email string) {
	url := workspace.AvatarURL(name, email, s.avatarOverrides)
	if bitmap := s.avatars[url]; bitmap != nil {
		ui.Image(c, bitmap).Size(24, 24).Radius(12)
		return
	}
	hue := workspace.AvatarHue(name, email)
	palette := []ui.Color{ui.RGB(57, 91, 129), ui.RGB(110, 72, 132), ui.RGB(50, 109, 101), ui.RGB(129, 91, 51), ui.RGB(111, 78, 91)}
	ui.Box(c).Size(24, 24).Radius(12).Background(palette[hue%uint32(len(palette))]).Center().Children(func() { ui.Text(c, workspace.AvatarInitials(name)).FontSize(10).TextColor(ui.RGB(250, 250, 250)) })
	if url == "" || s.avatarPending[url] || s.Options.AvatarLoader == nil {
		return
	}
	s.avatarPending[url] = true
	load, wake, ctx, ch := s.Options.AvatarLoader, s.Options.OnChange, s.avatarContext, s.avatarResults
	if s.Options.Invalidate != nil {
		wake = s.Options.Invalidate
	}
	go func() {
		data, e := load(ctx, url)
		if e != nil {
			data = nil
		}
		if ctx.Err() != nil {
			return
		}
		select {
		case ch <- avatarResult{url, data}:
			if wake != nil {
				wake()
			}
		case <-ctx.Done():
		}
	}()
}
func (s *State) drainAvatars() {
	for {
		select {
		case r := <-s.avatarResults:
			if len(r.data) > 0 {
				if bitmap, e := ui.DecodeBitmap(r.data); e == nil {
					s.avatars[r.url] = bitmap
				}
			}
		default:
			return
		}
	}
}
