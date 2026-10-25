package bugreportui

import (
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
)

func button(c *ui.Context, label string, disabled bool) ui.Element {
	return ui.Button(c, label).Disabled(disabled)
}
func View(c *ui.Context, s *State) ui.Element {
	s.Tick()
	return ui.Scroll(c).MinWidth(0).Padding(24, 32).Children(func() {
		ui.Column(c).WidthPercent(100).MaxWidth(760).AlignSelf(ui.Center).Gap(12).Children(func() {
			ui.Text(c, s.label("bug_report.heading", "Report a bug")).FontSize(20).FontWeight(500)
			ui.Text(c, s.label("bug_report.review_intro", "Describe the problem, review exactly what will be shared, then submit it to GitHub.")).TextColor(c.Theme().TextMuted)
			ui.Row(c).Gap(8).Children(func() {
				for _, kind := range []native.ReportKind{native.ReportBug, native.ReportFeature, native.ReportOther} {
					label := map[native.ReportKind]string{native.ReportBug: "Bug", native.ReportFeature: "Feature", native.ReportOther: "Other"}[kind]
					if s.Kind == kind {
						label = "✓ " + label
					}
					if button(c, label, s.Busy).Clicked() {
						s.Kind = kind
					}
				}
			})
			ui.Text(c, s.label("bug_report.context", "Context")).FontWeight(500)
			ui.TextArea(c, &s.Notes).Label(s.label("bug_report.context", "Context")).Placeholder("Only these notes are sent to the selected AI model when you choose Generate draft.").Height(90).AlignSelf(ui.Stretch).Disabled(s.Busy)
			if button(c, s.label("bug_report.generate", "Generate draft"), s.Busy || strings.TrimSpace(s.Notes) == "").AlignSelf(ui.Start).Clicked() {
				s.Generate()
			}
			ui.Text(c, s.label("bug_report.title", "Title")).FontWeight(500)
			ui.TextInput(c, &s.Title).Placeholder(s.label("bug_report.title_placeholder", "A one-line summary of the problem")).Label(s.label("bug_report.title", "Title")).AlignSelf(ui.Stretch).Disabled(s.Busy)
			what, steps := s.Kind.Headings()
			ui.Text(c, what).FontWeight(500)
			ui.TextArea(c, &s.What).Placeholder(s.label("bug_report.what_placeholder", "Describe the issue and what you expected instead.")).Label(what).Height(120).AlignSelf(ui.Stretch).Disabled(s.Busy)
			ui.Text(c, steps).FontWeight(500)
			ui.TextArea(c, &s.Steps).Placeholder(s.label("bug_report.steps_placeholder", "One step per line, if you have them.")).Label(steps).Height(85).AlignSelf(ui.Stretch).Disabled(s.Busy)
			ui.Checkbox(c, &s.RedactSensitive, "Redact recognizable secrets and my home path in report text").Disabled(s.Busy)
			ui.Text(c, "Redaction cannot detect every secret. Review text and screenshots yourself. No logs, credentials or session files are collected.").FontSize(12).TextColor(c.Theme().TextMuted)
			ui.Row(c).Gap(8).Children(func() {
				if ui.Checkbox(c, &s.IncludeDiagnostics, "Include build and runtime diagnostics").Disabled(s.Busy).Changed() && s.IncludeDiagnostics {
					s.RefreshDiagnostics()
				}
				if button(c, "Refresh diagnostics", s.Busy || !s.IncludeDiagnostics).Clicked() {
					s.RefreshDiagnostics()
				}
			})
			if s.IncludeDiagnostics {
				if ui.Checkbox(c, &s.IncludeIdentity, "Include my GitHub username (contacts GitHub)").Disabled(s.Busy).Changed() {
					s.RefreshDiagnostics()
				}
				ui.Text(c, s.Diagnostics.Markdown()).Selectable().FontSize(12)
			}
			ui.Text(c, fmt.Sprintf("Screenshots (%d/%d)", len(s.shots), native.MaxReportScreenshots)).Bold()
			ui.Text(c, "Only images you select or explicitly capture are included. Image metadata is removed; use Preview to redact pixels.").FontSize(12).TextColor(c.Theme().TextMuted)
			ui.Row(c).Gap(8).Children(func() {
				if button(c, s.label("bug_report.add_screenshot", "Add screenshots…"), s.Busy || len(s.shots) >= native.MaxReportScreenshots).Clicked() {
					s.pickScreenshots()
				}
				if s.Options.Capture != nil {
					if button(c, "Capture this window", s.Busy || !s.CaptureConsent || len(s.shots) >= native.MaxReportScreenshots).Clicked() {
						s.Capture()
					}
				}
			})
			if s.Options.Capture != nil {
				ui.Checkbox(c, &s.CaptureConsent, "I want to capture the current Orbit window, including any visible private content").Disabled(s.Busy)
			}
			for index := 0; index < len(s.shots); index++ {
				shot := s.shots[index]
				ui.Row(c).Gap(8).Children(func() {
					ui.Image(c, shot.Bitmap).Width(100).Height(65)
					ui.Text(c, fmt.Sprintf("%s · %d × %d", shot.Image.Name, shot.Image.Width, shot.Image.Height)).Grow(1)
					if button(c, "Preview / redact", s.Busy).Clicked() {
						s.preview = index
					}
					if button(c, "Remove", s.Busy).Clicked() {
						s.shots = append(s.shots[:index], s.shots[index+1:]...)
						s.preview = -1
					}
				})
				if len(s.shots) <= index {
					break
				}
			}
			if s.Activity != "" {
				ui.Text(c, s.Activity+"…").TextColor(c.Theme().TextMuted)
			}
			if s.Error != nil {
				ui.Text(c, s.Error.Error()).TextColor(ui.RGB(0xe0, 0x6b, 0x70)).Selectable()
			}
			if s.Status != "" {
				ui.Text(c, s.Status).Selectable()
			}
			ui.Row(c).Gap(8).Children(func() {
				if button(c, s.label("bug_report.clear", "Clear"), s.Busy).Clicked() {
					s.Clear()
				}
				if button(c, "Review report…", s.Busy || strings.TrimSpace(s.Title) == "" || strings.TrimSpace(s.What) == "").Clicked() {
					s.Prepare()
				}
			})
			if s.reviewOpen {
				s.review(c)
			}
			if s.preview >= 0 && s.preview < len(s.shots) {
				s.previewView(c)
			}
		})
	})
}

func dialog(c *ui.Context, open *bool, fn func()) {
	ui.DialogBase(c, open, func(back, panel ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, .45))
		panel.Width(720).MaxWidthPercent(96).MaxHeight(760).MaxHeightPercent(95).Padding(18).Gap(12).Radius(10).Background(c.Theme().Surface).Border(1, c.Theme().Border).Children(fn)
	})
}
func (s *State) review(c *ui.Context) {
	dialog(c, &s.reviewOpen, func() {
		ui.Text(c, "Review public GitHub report").FontSize(18).Bold()
		ui.Text(c, "Destination: https://github.com/"+s.prepared.Repository+"/issues").Selectable()
		ui.Text(c, s.prepared.Title).Bold()
		ui.Scroll(c).Height(300).AlignSelf(ui.Stretch).Children(func() { ui.Text(c, s.prepared.BrowserBody).Selectable() })
		ui.Text(c, fmt.Sprintf("%d explicitly selected screenshots will be shared. Nothing is sent until you confirm.", len(s.prepared.Screenshots)))
		ui.Checkbox(c, &s.reviewed, "I reviewed the report and screenshots and want to submit them publicly")
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			if button(c, "Back to editing", false).Clicked() {
				s.reviewOpen = false
			}
			if button(c, "Copy reviewed report", false).Clicked() {
				c.Services().WriteClipboard(s.prepared.Title + "\n\n" + s.prepared.BrowserBody)
			}
			if button(c, "Submit reviewed report", !s.reviewed || s.Busy).Clicked() {
				s.Submit()
			}
		})
	})
}
func (s *State) previewView(c *ui.Context) {
	open := true
	index := s.preview
	shot := &s.shots[index]
	dialog(c, &open, func() {
		ui.Text(c, shot.Image.Name).Bold()
		ui.Image(c, shot.Bitmap).Height(320).MaxWidth(680)
		ui.Text(c, fmt.Sprintf("Original pixels: %d × %d. Opaque black boxes remove selected pixels from the attachment only.", shot.Image.Width, shot.Image.Height)).FontSize(12)
		ui.Text(c, "Redaction boxes: x,y,width,height; separate boxes with semicolons").FontSize(12)
		ui.TextInput(c, &shot.Redactions).Label("Redaction boxes: x,y,width,height; x,y,width,height").Placeholder("20,40,200,50").AlignSelf(ui.Stretch)
		ui.Row(c).Gap(8).Children(func() {
			if button(c, "Apply redaction", s.Busy || strings.TrimSpace(shot.Redactions) == "").Clicked() {
				s.redact(index)
			}
			if button(c, "Done", s.Busy).Clicked() {
				s.preview = -1
			}
		})
	})
	if !open {
		s.preview = -1
	}
}
