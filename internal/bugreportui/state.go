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

// Theme supplies source palette and density tokens without coupling this package
// to the application shell. Omitted values use the active MyGo theme.
type Theme struct {
	SurfaceSecondary, Elevated, TextSecondary, Action, ActionText ui.Color
	ActionHover, ActionPressed                                    ui.Color
	Px                                                            func(float32) float32
	Space                                                         func(int) float32
}

type Options struct {
	OnChange                            func()
	PiBinary, Version, Home, Repository string
	Window                              func() *mygo.Window
	Capture                             func() ([]byte, error)
	OpenURL                             func(string) error
	GitHub                              native.ReportGitHub
	GenerateDraft                       func(context.Context, native.ReportKind, string) (workspace.Draft, error)
	Text                                func(key, fallback string) string
	Theme                               func() Theme
	ModelLabel                          func() string
	// PickScreenshots permits a host picker or an offline test fixture.
	PickScreenshots func(context.Context) ([]string, error)
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
	reviewOpen, reviewed, optionsOpen                                    bool
	cancelling                                                           bool
	operation                                                            string
	done                                                                 chan struct{}
	scroll                                                               ui.ScrollState
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
	return &State{Options: o, Kind: native.ReportBug, RedactSensitive: true, results: make(chan result, 8), done: make(chan struct{}), preview: -1}
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
	s.operation = label
	s.cancelling = false
	s.Error = nil
	s.Status = ""
	s.generation++
	generation := s.generation
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	channel, wake, done := s.results, s.Options.OnChange, s.done
	go func() {
		defer cancel()
		apply, e := job(ctx)
		// A cancellation must still arrive at Tick so the duplicate-operation
		// gate stays closed until the worker has actually returned.
		if ctx.Err() != nil {
			apply, e = nil, ctx.Err()
		}
		select {
		case channel <- result{generation, apply, e}:
			if wake != nil {
				wake()
			}
		case <-done:
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
			if s.cancelling && errors.Is(r.err, context.Canceled) {
				s.Error = nil
				s.Status = "Cancelled. Your report has been kept."
				if s.operation == "Submitting reviewed report" {
					s.Status = "Submission stopped. GitHub may already have created the issue; check the repository before retrying."
				}
			}
			s.cancelling = false
			if r.apply != nil {
				r.apply(s)
			}
		default:
			return
		}
	}
}
func (s *State) Close() {
	if s.closed {
		return
	}
	s.closed = true
	close(s.done)
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
	s.optionsOpen = false
	s.reviewed = false
	s.cancelling = false
	s.CaptureConsent = false
	s.prepared = native.PreparedReport{}
	s.Status = ""
	s.Error = nil
}

// Cancel stops the active operation without dropping the user's report. Busy
// remains true until its worker exits, preventing an overlapping submission.
func (s *State) Cancel() {
	if s.Busy && s.cancel != nil && !s.cancelling {
		s.cancelling = true
		s.Activity = "Cancelling"
		s.cancel()
	}
}

// Escape dismisses only the foremost report surface, then cancels active work.
func (s *State) Escape() bool {
	if s.preview >= 0 {
		s.preview = -1
		return true
	}
	if s.optionsOpen {
		s.optionsOpen = false
		s.CaptureConsent = false
		return true
	}
	if s.reviewOpen {
		s.reviewOpen = false
		s.reviewed = false
		return true
	}
	if s.Busy {
		s.Cancel()
		return true
	}
	return false
}

func (s *State) SelectKind(kind native.ReportKind) {
	if s.Busy || s.closed || (kind != native.ReportBug && kind != native.ReportFeature && kind != native.ReportOther) {
		return
	}
	s.Kind = kind
	s.Error = nil
}

func (s *State) RefreshDiagnostics() {
	if !s.IncludeDiagnostics {
		return
	}
	o := native.DiagnosticOptions{Version: s.Options.Version, PiBinary: s.Options.PiBinary, IncludeGitHubIdentity: s.IncludeIdentity, GitHub: s.Options.GitHub}
	s.start("Collecting diagnostics", func(ctx context.Context) (func(*State), error) {
		d := native.CollectDiagnostics(ctx, o)
		return func(s *State) { s.Diagnostics = d }, nil
	})
}
func (s *State) Generate() {
	if s.Busy || s.closed {
		return
	}
	if strings.TrimSpace(s.Notes) == "" {
		s.Error = errors.New(s.label("bug_report.context_required", "Add some context before generating a draft."))
		return
	}
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
				if d.Warning != "" {
					s.Status += " " + d.Warning
				}
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

// readScreenshots mirrors the source's six-image cap and name deduplication.
// A bad file does not discard other successfully selected images.
func readScreenshots(ctx context.Context, paths []string, existing []screenshot) ([]screenshot, error) {
	room := native.MaxReportScreenshots - len(existing)
	names := make(map[string]bool, len(existing))
	for _, shot := range existing {
		names[shot.Image.Name] = true
	}
	var shots []screenshot
	var lastError error
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(shots) >= room {
			break
		}
		img, err := native.LoadReportScreenshot(path)
		if err != nil {
			lastError = err
			continue
		}
		if names[img.Name] {
			continue
		}
		bitmap, err := ui.DecodeBitmap(img.PNG)
		if err != nil {
			lastError = err
			continue
		}
		names[img.Name] = true
		shots = append(shots, screenshot{Image: img, Bitmap: bitmap})
	}
	return shots, lastError
}

func (s *State) AddFiles(paths []string) {
	if s.Busy || s.closed {
		return
	}
	if len(s.shots) >= native.MaxReportScreenshots {
		s.Error = localValidationError("Up to 6 screenshots.")
		return
	}
	copyPaths := append([]string(nil), paths...)
	existing := append([]screenshot(nil), s.shots...)
	s.start("Reading selected screenshots", func(ctx context.Context) (func(*State), error) {
		shots, err := readScreenshots(ctx, copyPaths, existing)
		return func(s *State) { s.shots = append(s.shots, shots...) }, err
	})
}
func (s *State) pickScreenshots() {
	if s.Busy || s.closed {
		return
	}
	if len(s.shots) >= native.MaxReportScreenshots {
		s.Error = localValidationError("Up to 6 screenshots.")
		return
	}
	window, pick := s.Options.Window, s.Options.PickScreenshots
	existing := append([]screenshot(nil), s.shots...)
	title := s.label("bug_report.add_screenshot", "Add screenshot")
	s.start("Choose screenshots", func(ctx context.Context) (func(*State), error) {
		var paths []string
		var err error
		if pick != nil {
			paths, err = pick(ctx)
		} else {
			var parent *mygo.Window
			if window != nil {
				parent = window()
			}
			paths, err = mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: parent, Title: title, Multiple: true, Filters: []mygo.FileFilter{{Name: "Images", Extensions: []string{"png", "jpg", "jpeg", "webp", "gif", "bmp"}}}})
		}
		if err != nil {
			return nil, err
		}
		shots, err := readScreenshots(ctx, paths, existing)
		return func(s *State) { s.shots = append(s.shots, shots...) }, err
	})
}
func (s *State) RemoveScreenshot(index int) {
	if s.Busy || index < 0 || index >= len(s.shots) {
		return
	}
	s.shots = append(s.shots[:index], s.shots[index+1:]...)
	s.preview = -1
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
			return nil, localValidationError("use x,y,width,height; separate boxes with semicolons")
		}
		var n [4]int
		for i, p := range parts {
			v, e := strconv.Atoi(strings.TrimSpace(p))
			if e != nil || v < 0 || v > 100000 {
				return nil, localValidationError("redaction coordinates must be nonnegative pixels")
			}
			n[i] = v
		}
		if n[2] == 0 || n[3] == 0 {
			return nil, localValidationError("redaction width and height must be positive")
		}
		boxes = append(boxes, image.Rect(n[0], n[1], n[0]+n[2], n[1]+n[3]))
	}
	return boxes, nil
}
func (s *State) redact(index int) {
	if s.Busy || index < 0 || index >= len(s.shots) {
		return
	}
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
	if s.Busy || s.closed {
		return
	}
	if strings.TrimSpace(s.Title) == "" {
		s.Error = errors.New(s.label("bug_report.title_required", "Add a title for the issue."))
		return
	}
	if strings.TrimSpace(s.What) == "" {
		s.Error = errors.New(s.label("bug_report.body_required", "Describe what happened."))
		return
	}
	diagnostics := s.Diagnostics
	// Revoking identity inclusion must take effect even if an earlier diagnostic
	// snapshot already contains it or the refresh fails.
	if !s.IncludeIdentity {
		diagnostics.GHUser = ""
	}
	in := native.ReportInput{Repository: s.Options.Repository, Kind: s.Kind, Title: s.Title, What: s.What, Steps: s.Steps, IncludeDiagnostics: s.IncludeDiagnostics, Diagnostics: diagnostics, RedactSensitive: s.RedactSensitive, Home: s.Options.Home}
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
	if s.Busy || s.closed || !s.reviewOpen || !s.reviewed {
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

// HasModal lets native application-menu routing respect report review and
// screenshot preview. Merely entering this state never transmits a report.
func (s *State) HasModal() bool {
	return s != nil && (s.reviewOpen || s.optionsOpen || (s.preview >= 0 && s.preview < len(s.shots)))
}
