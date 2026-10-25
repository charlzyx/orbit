package orbitui

import (
	"context"
	"fmt"
	"image"
	"time"

	"github.com/charlzyx/orbit/internal/background"
	"github.com/egoist/mygo/ui"
)

type backdropState struct {
	cancel     context.CancelFunc
	closed     bool
	store      *background.PhotoStore
	field      *background.Field
	clock      background.Clock
	config     background.PhotoConfig
	pixels     *image.NRGBA
	bitmap     *ui.Bitmap
	loading    bool
	err        string
	generation uint64
	pending    *background.Tuning
}

func (a *App) initBackground() {
	s := &a.backdrop
	s.store = background.NewPhotoStore(a.Options.Store.RootDir)
	s.field, _ = background.NewField()
	s.config.Tuning = background.DefaultTuning()
	a.backgroundJob(nil)
}

// All PhotoStore calls, including configuration reads and cached-image lookup,
// run in this serialized worker. Rapid slider edits coalesce to the latest step.
func (a *App) backgroundJob(op func(context.Context, *background.PhotoStore) error) {
	s := &a.backdrop
	if s.loading || s.closed {
		return
	}
	s.loading = true
	s.generation++
	generation, store := s.generation, s.store
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	a.Async(func() func() {
		defer cancel()
		var err error
		if op != nil {
			err = op(ctx, store)
		}
		cfg := store.Config()
		pixels, imageErr := store.Image(ctx)
		if err == nil {
			err = imageErr
		}
		return func() {
			if generation != s.generation {
				return
			}
			s.loading = false
			s.err = ""
			if err != nil {
				s.err = err.Error()
			}
			s.config = cfg
			if imageErr == nil && pixels != s.pixels {
				s.pixels = pixels
				s.bitmap = nil
				if pixels != nil {
					s.bitmap = ui.NewBitmap(pixels)
				}
			}
			if s.pending != nil {
				next := *s.pending
				s.pending = nil
				if next != cfg.Tuning {
					a.setBackgroundTuning(next)
				}
			}
		}
	})
}
func (a *App) chooseBackground(path string) {
	if path == "" || a.backdrop.loading {
		return
	}
	a.backgroundJob(func(ctx context.Context, s *background.PhotoStore) error { _, err := s.Choose(ctx, path); return err })
}
func (a *App) clearBackground() {
	if a.backdrop.loading {
		return
	}
	a.backgroundJob(func(_ context.Context, s *background.PhotoStore) error { return s.Clear() })
}
func (a *App) setBackgroundTuning(t background.Tuning) {
	t = t.Normalize()
	s := &a.backdrop
	if s.loading {
		s.pending = &t
		return
	}
	if t == s.config.Tuning {
		return
	}
	a.backgroundJob(func(_ context.Context, s *background.PhotoStore) error { _, err := s.SetTuning(t); return err })
}

// The paint box is clipped so Cover never bleeds over the sidebar. It is only
// mounted below the main top bar on an empty task, never over a feature page.
func (a *App) paintBackdrop(p *ui.Painter, r ui.Rect, animate bool) {
	s, t := &a.backdrop, a.Theme
	if s.bitmap != nil {
		p.Image(s.bitmap, r, ui.Cover)
		if fade := s.config.Fade; fade > 0 {
			h := r.H * fade
			p.FillGradient(ui.Rect{X: r.X, Y: r.Y + r.H - h, W: r.W, H: h}, ui.LinearGradient{From: t.Background.Alpha(0), To: t.Background.Alpha(s.config.FadePeak()), Angle: 180}, 0)
		}
		return
	}
	if s.field == nil {
		return
	}
	phase := float32(0)
	if animate {
		phase = s.clock.Phase(p.Now(), a.Prefs.ReduceMotion)
	}
	s.field.Paint(p, r, phase, t.Px(1), t.TextMuted, t.Dark)
	if animate && !a.Prefs.ReduceMotion {
		p.After(90 * time.Millisecond)
	}
}
func (a *App) sourceBackgroundRow(c *ui.Context) {
	s := &a.backdrop
	label := a.Tr("settings.choose_image")
	current := ""
	if s.config.Source != nil {
		current = *s.config.Source
		label = a.Tr("settings.replace_ellipsis")
	}
	a.sourceSettingsRow(c, a.Tr("settings.background_image"), a.Tr("settings.a_dithered_image_behind_the_new_task_page"), current, func() {
		ui.Row(c).Gap(a.Theme.Space(8)).Children(func() {
			if a.sourceSettingsButton(c, label).Disabled(s.loading).Clicked() {
				a.Dispatch(Action{Kind: "BrowseBackground"})
			}
			if s.config.Source != nil && a.sourceSettingsButton(c, "Reset").Disabled(s.loading).Clicked() {
				a.clearBackground()
			}
		})
	})
	if s.loading {
		ui.Row(c).Padding(8, 16).Gap(8).Children(func() {
			ui.Spinner(c).Size(12, 12)
			ui.Text(c, "Processing background…").TextColor(a.Theme.TextMuted)
		})
	}
	if s.err != "" {
		ui.Text(c, s.err).Padding(8, 16).TextColor(a.Theme.Error).Selectable()
	}
}
func (a *App) backgroundSettings(c *ui.Context) {
	s := &a.backdrop
	if s.config.Source == nil {
		return
	}
	a.sourceSettingsSection(c, a.Tr("settings.background_tuning"), "", func() {
		tuning := s.config.Tuning
		if s.pending != nil {
			tuning = *s.pending
		}
		a.sourceSettingsRow(c, a.Tr("settings.blur"), a.Tr("settings.softens_the_picture_before_the_dither_pass_so_a_"), "", func() {
			blur := float64(tuning.Blur)
			if ui.StepSlider(c, &blur, 0, float64(background.BlurMax), float64(background.BlurStep)).Width(160).Label("Background blur").Changed() {
				tuning.Blur = float32(blur)
				a.setBackgroundTuning(tuning)
			}
			ui.Text(c, fmt.Sprintf("%.1f", blur)).Width(28)
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.pixel_size"), a.Tr("settings.the_dither_cell_edge_fine_cells_read_as_halftone"), "", func() {
			choices := []settingsChoice{}
			for _, v := range background.CellLabels {
				choices = append(choices, settingsChoice{v, v})
			}
			a.sourceSettingsSelect(c, "Background pixel size", background.CellLabels[tuning.CellIndex()], choices, func(v string) {
				for i, label := range background.CellLabels {
					if v == label {
						tuning.Cell = background.CellSizes[i]
						a.setBackgroundTuning(tuning)
						break
					}
				}
			})
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.bottom_fade"), a.Tr("settings.how_far_the_picture_fades_into_the_page_behind_t"), "", func() {
			choices := []settingsChoice{}
			for _, v := range background.FadeLabels {
				choices = append(choices, settingsChoice{v, v})
			}
			a.sourceSettingsSelect(c, "Background fade", background.FadeLabels[tuning.FadeIndex()], choices, func(v string) {
				for i, label := range background.FadeLabels {
					if v == label {
						tuning.Fade = background.FadeHeights[i]
						a.setBackgroundTuning(tuning)
						break
					}
				}
			})
		})
	})
}

func (a *App) closeBackground() {
	s := &a.backdrop
	s.closed = true
	s.generation++
	if s.cancel != nil {
		s.cancel()
	}
	s.pending = nil
}
