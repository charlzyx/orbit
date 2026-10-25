// Package bugreportui implements the source Report a bug page with native MyGo
// controls. I/O runs on workers; only Tick applies their results to UI state.
package bugreportui

import (
	"context"
	"errors"
	"fmt"
	"image"
	"strconv"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type Options struct {
	OnChange                            func()
	PiBinary, Version, Home, Repository string
	Window                              func() *mygo.Window
	Capture                             func() ([]byte, error)
	OpenURL                             func(string) error
	GitHub                              native.ReportGitHub
	GenerateDraft                       func(context.Context, native.ReportKind, string) (workspace.Draft, error)
	Text                                func(key, fallback string) string
}
type screenshot struct {
	Image      native.ReportScreenshot
	Bitmap     *ui.Bitmap
	Redactions string
}
type result struct {
	generation uint64
	apply      func(*State)
	err        error
}
type State struct {
	Options                                                              Options
	Kind                                                                 native.ReportKind
	Notes, Title, What, Steps                                            string
	IncludeDiagnostics, IncludeIdentity, RedactSensitive, CaptureConsent bool
	Diagnostics                                                          native.Diagnostics
	Busy                                                                 bool
	Activity, Status                                                     string
	Error                                                                error
	shots                                                                []screenshot
	results                                                              chan result
	cancel                                                               context.CancelFunc
	generation                                                           uint64
	closed                                                               bool
	reviewOpen, reviewed                                                 bool
	prepared                                                             native.PreparedReport
	preview                                                              int
}

func New(o Options) *State {
	if o.Repository == "" {
		o.Repository = native.DefaultReportRepository
	}
	if o.GitHub == nil {
		o.GitHub = workspace.GitHub{Dir: o.Home, Binary: native.LocateExecutable("gh")}
	}
	if o.OpenURL == nil {
		o.OpenURL = native.OpenURL
	}
	return &State{Options: o, Kind: native.ReportBug, RedactSensitive: true, results: make(chan result, 8), preview: -1}
}
func (s *State) label(key, fallback string) string {
	if s.Options.Text != nil {
		return s.Options.Text(key, fallback)
	}
	return fallback
}
func (s *State) start(label string, job func(context.Context) (func(*State), error)) {
	if s.Busy || s.closed {
		return
	}
	s.Busy = true
	s.Activity = label
	s.Error = nil
	s.generation++
	generation := s.generation
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	channel, wake := s.results, s.Options.OnChange
	go func() {
		defer cancel()
		apply, e := job(ctx)
		select {
		case channel <- result{generation, apply, e}:
			if wake != nil {
				wake()
			}
		case <-ctx.Done():
		}
	}()
}
func (s *State) Tick() {
	for {
		select {
		case r := <-s.results:
			if s.closed || r.generation != s.generation {
				continue
			}
			s.Busy = false
			s.Activity = ""
			s.cancel = nil
			s.Error = r.err
			if r.apply != nil {
				r.apply(s)
			}
		default:
			return
		}
	}
}
func (s *State) Close() {
	s.closed = true
	s.generation++
	if s.cancel != nil {
		s.cancel()
	}
	s.shots = nil
}
func (s *State) Clear() {
	if s.cancel != nil {
		s.cancel()
	}
	s.generation++
	s.cancel = nil
	s.Busy = false
	s.Activity = ""
	s.Title = ""
	s.What = ""
	s.Steps = ""
	s.Notes = ""
	s.shots = nil
	s.preview = -1
	s.reviewOpen = false
	s.prepared = native.PreparedReport{}
	s.Status = ""
	s.Error = nil
}
func (s *State) RefreshDiagnostics() {
	o := native.DiagnosticOptions{Version: s.Options.Version, PiBinary: s.Options.PiBinary, IncludeGitHubIdentity: s.IncludeIdentity, GitHub: s.Options.GitHub}
	s.start("Collecting diagnostics", func(ctx context.Context) (func(*State), error) {
		d := native.CollectDiagnostics(ctx, o)
		return func(s *State) { s.Diagnostics = d }, nil
	})
}
func (s *State) Generate() {
	notes, kind, generate := s.Notes, s.Kind, s.Options.GenerateDraft
	s.start("Drafting from your notes", func(ctx context.Context) (func(*State), error) {
		d := fallbackDraft(notes)
		var e error
		if generate != nil {
			d, e = generate(ctx, kind, notes)
		}
		if e != nil {
			return nil, e
		}
		return func(s *State) {
			s.Title = d.Title
			s.What = d.Body
			if d.Fallback {
				s.Status = "Local draft used. Review and edit it before sharing."
			}
		}, nil
	})
}
func fallbackDraft(notes string) workspace.Draft {
	notes = strings.TrimSpace(notes)
	title := ""
	for _, l := range strings.Split(notes, "\n") {
		if strings.TrimSpace(l) != "" {
			title = strings.TrimRight(strings.TrimSpace(l), ".:;")
			break
		}
	}
	r := []rune(title)
	if len(r) > 72 {
		title = string(r[:72])
	}
	return workspace.Draft{Title: title, Body: notes, Fallback: true}
}
func (s *State) AddFiles(paths []string) {
	room := native.MaxReportScreenshots - len(s.shots)
	copyPaths := append([]string(nil), paths...)
	s.start("Reading selected screenshots", func(ctx context.Context) (func(*State), error) {
		var shots []screenshot
		for _, path := range copyPaths {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			if len(shots) >= room {
				return nil, errors.New("a report can include at most 6 screenshots")
			}
			img, e := native.LoadReportScreenshot(path)
			if e != nil {
				return nil, e
			}
			bitmap, e := ui.DecodeBitmap(img.PNG)
			if e != nil {
				return nil, e
			}
			shots = append(shots, screenshot{Image: img, Bitmap: bitmap})
		}
		return func(s *State) { s.shots = append(s.shots, shots...) }, nil
	})
}
func (s *State) pickScreenshots() {
	window := s.Options.Window
	room := native.MaxReportScreenshots - len(s.shots)
	s.start("Choose screenshots", func(ctx context.Context) (func(*State), error) {
		var parent *mygo.Window
		if window != nil {
			parent = window()
		}
		paths, e := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: parent, Title: "Add report screenshots", Multiple: true, Filters: []mygo.FileFilter{{Name: "Images", Extensions: []string{"png", "jpg", "jpeg", "webp", "gif", "bmp"}}}})
		if e != nil {
			return nil, e
		}
		if len(paths) > room {
			return nil, errors.New("a report can include at most 6 screenshots")
		}
		var shots []screenshot
		for _, path := range paths {
			if e = ctx.Err(); e != nil {
				return nil, e
			}
			img, e := native.LoadReportScreenshot(path)
			if e != nil {
				return nil, e
			}
			bitmap, e := ui.DecodeBitmap(img.PNG)
			if e != nil {
				return nil, e
			}
			shots = append(shots, screenshot{Image: img, Bitmap: bitmap})
		}
		return func(s *State) { s.shots = append(s.shots, shots...) }, nil
	})
}
func (s *State) Capture() {
	if !s.CaptureConsent || len(s.shots) >= native.MaxReportScreenshots {
		return
	}
	capture := s.Options.Capture
	if capture == nil {
		return
	}
	s.start("Capturing current Orbit window", func(ctx context.Context) (func(*State), error) {
		raw, e := capture()
		if e != nil {
			return nil, e
		}
		if e = ctx.Err(); e != nil {
			return nil, e
		}
		shot, e := native.NewReportScreenshot("orbit-"+time.Now().Format("20060102-150405")+".png", raw)
		if e != nil {
			return nil, e
		}
		bitmap, e := ui.DecodeBitmap(shot.PNG)
		if e != nil {
			return nil, e
		}
		return func(s *State) {
			s.shots = append(s.shots, screenshot{Image: shot, Bitmap: bitmap})
			s.preview = len(s.shots) - 1
			s.CaptureConsent = false
		}, nil
	})
}
func ParseRedactionBoxes(text string) ([]image.Rectangle, error) {
	var boxes []image.Rectangle
	for _, line := range strings.Split(text, ";") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 4 {
			return nil, errors.New("use x,y,width,height; separate boxes with semicolons")
		}
		var n [4]int
		for i, p := range parts {
			v, e := strconv.Atoi(strings.TrimSpace(p))
			if e != nil || v < 0 || v > 100000 {
				return nil, errors.New("redaction coordinates must be nonnegative pixels")
			}
			n[i] = v
		}
		if n[2] == 0 || n[3] == 0 {
			return nil, errors.New("redaction width and height must be positive")
		}
		boxes = append(boxes, image.Rect(n[0], n[1], n[0]+n[2], n[1]+n[3]))
	}
	return boxes, nil
}
func (s *State) redact(index int) {
	shot := s.shots[index]
	boxes, e := ParseRedactionBoxes(shot.Redactions)
	if e != nil {
		s.Error = e
		return
	}
	s.start("Redacting screenshot", func(ctx context.Context) (func(*State), error) {
		redacted, e := native.RedactScreenshot(shot.Image, boxes)
		if e != nil {
			return nil, e
		}
		bitmap, e := ui.DecodeBitmap(redacted.PNG)
		if e != nil {
			return nil, e
		}
		return func(s *State) {
			if index < len(s.shots) {
				s.shots[index] = screenshot{Image: redacted, Bitmap: bitmap}
			}
		}, nil
	})
}
func (s *State) Prepare() {
	in := native.ReportInput{Repository: s.Options.Repository, Kind: s.Kind, Title: s.Title, What: s.What, Steps: s.Steps, IncludeDiagnostics: s.IncludeDiagnostics, Diagnostics: s.Diagnostics, RedactSensitive: s.RedactSensitive, Home: s.Options.Home}
	for _, shot := range s.shots {
		in.Screenshots = append(in.Screenshots, shot.Image)
	}
	s.start("Preparing local preview", func(ctx context.Context) (func(*State), error) {
		p, e := native.PrepareReport(in)
		if e != nil {
			return nil, e
		}
		return func(s *State) { s.prepared = p; s.reviewed = false; s.reviewOpen = true }, nil
	})
}
func (s *State) Submit() {
	if !s.reviewOpen || !s.reviewed {
		return
	}
	p, h := s.prepared, s.Options.GitHub
	s.reviewOpen = false
	s.start("Submitting reviewed report", func(ctx context.Context) (func(*State), error) {
		filed, e := native.FileReport(ctx, h, p, true)
		if e != nil {
			return nil, e
		}
		return func(s *State) {
			if filed.Created {
				s.Clear()
				s.Status = "Issue created: " + filed.URL
			} else {
				s.Status = "GitHub draft prepared. Review and submit it in your browser."
			}
			if filed.ScreenshotDirectory != "" {
				s.Status += " Screenshots: " + filed.ScreenshotDirectory
			}
			if filed.URL != "" {
				if e := s.Options.OpenURL(filed.URL); e != nil {
					s.Error = e
				}
			}
		}, nil
	})
}
func (s *State) ScreenshotCount() int { return len(s.shots) }
func (s *State) Summary() string      { return fmt.Sprintf("%s · %d screenshots", s.Kind, len(s.shots)) }
